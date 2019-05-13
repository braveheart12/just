//
// Copyright 2019 Insolar Technologies GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package artifactmanager

import (
	"context"
	"fmt"
	"time"

	"github.com/insolar/insolar/ledger/light/handle"
	"github.com/pkg/errors"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"

	"github.com/insolar/insolar/configuration"
	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/flow"
	"github.com/insolar/insolar/insolar/flow/bus"
	"github.com/insolar/insolar/insolar/flow/handler"
	"github.com/insolar/insolar/insolar/jet"
	"github.com/insolar/insolar/insolar/message"
	"github.com/insolar/insolar/insolar/node"
	"github.com/insolar/insolar/insolar/pulse"
	"github.com/insolar/insolar/insolar/record"
	"github.com/insolar/insolar/insolar/reply"
	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/insolar/insolar/instrumentation/insmetrics"
	"github.com/insolar/insolar/ledger/blob"
	"github.com/insolar/insolar/ledger/drop"
	"github.com/insolar/insolar/ledger/light/hot"
	"github.com/insolar/insolar/ledger/light/proc"
	"github.com/insolar/insolar/ledger/light/recentstorage"
	"github.com/insolar/insolar/ledger/object"
)

// MessageHandler processes messages for local storage interaction.
type MessageHandler struct {
	RecentStorageProvider      recentstorage.Provider             `inject:""`
	Bus                        insolar.MessageBus                 `inject:""`
	PlatformCryptographyScheme insolar.PlatformCryptographyScheme `inject:""`
	JetCoordinator             jet.Coordinator                    `inject:""`
	CryptographyService        insolar.CryptographyService        `inject:""`
	DelegationTokenFactory     insolar.DelegationTokenFactory     `inject:""`
	JetStorage                 jet.Storage                        `inject:""`

	DropModifier drop.Modifier `inject:""`

	BlobModifier blob.Modifier `inject:""`
	BlobAccessor blob.Accessor `inject:""`
	Blobs        blob.Storage  `inject:""`

	IDLocker object.IDLocker `inject:""`

	RecordModifier object.RecordModifier `inject:""`
	RecordAccessor object.RecordAccessor `inject:""`
	Nodes          node.Accessor         `inject:""`

	HotDataWaiter hot.JetWaiter   `inject:""`
	JetReleaser   hot.JetReleaser `inject:""`

	Index      object.Index
	IndexState object.IndexStateModifier

	conf           *configuration.Ledger
	middleware     *middleware
	jetTreeUpdater jet.Fetcher

	FlowHandler *handler.Handler
	handlers    map[insolar.MessageType]insolar.MessageHandler
}

// NewMessageHandler creates new handler.
func NewMessageHandler(
	index object.Index,
	indexState object.IndexStateModifier,
	conf *configuration.Ledger,
) *MessageHandler {

	h := &MessageHandler{
		handlers:   map[insolar.MessageType]insolar.MessageHandler{},
		conf:       conf,
		Index:      index,
		IndexState: indexState,
	}

	dep := &proc.Dependencies{
		FetchJet: func(p *proc.FetchJet) {
			p.Dep.JetAccessor = h.JetStorage
			p.Dep.Coordinator = h.JetCoordinator
			p.Dep.JetUpdater = h.jetTreeUpdater
			p.Dep.CheckJet = proc.NewCheckJet(h.jetTreeUpdater, h.JetCoordinator)
		},
		WaitHot: func(p *proc.WaitHot) {
			p.Dep.Waiter = h.HotDataWaiter
		},
		GetIndex: func(p *proc.GetIndex) {
			p.Dep.Index = h.Index
			p.Dep.IndexState = h.IndexState
			p.Dep.Locker = h.IDLocker
			p.Dep.Coordinator = h.JetCoordinator
			p.Dep.Bus = h.Bus
		},
		SendObject: func(p *proc.SendObject) {
			p.Dep.Jets = h.JetStorage
			p.Dep.Blobs = h.Blobs
			p.Dep.Coordinator = h.JetCoordinator
			p.Dep.JetUpdater = h.jetTreeUpdater
			p.Dep.Bus = h.Bus
			p.Dep.RecordAccessor = h.RecordAccessor
		},
		GetCode: func(p *proc.GetCode) {
			p.Dep.Bus = h.Bus
			p.Dep.RecordAccessor = h.RecordAccessor
			p.Dep.Coordinator = h.JetCoordinator
			p.Dep.CheckJet = proc.NewCheckJet(h.jetTreeUpdater, h.JetCoordinator)
			p.Dep.BlobAccessor = h.BlobAccessor
		},
	}

	h.FlowHandler = handler.NewHandler(func(msg bus.Message) flow.Handle {
		return (&handle.Init{
			Dep:     dep,
			Message: msg,
		}).Present
	})
	return h
}

func instrumentHandler(name string) Handler {
	return func(handler insolar.MessageHandler) insolar.MessageHandler {
		return func(ctx context.Context, p insolar.Parcel) (insolar.Reply, error) {
			inslog := inslogger.FromContext(ctx)
			start := time.Now()
			code := "2xx"
			ctx = insmetrics.InsertTag(ctx, tagMethod, name)

			repl, err := handler(ctx, p)

			latency := time.Since(start)
			if err != nil {
				code = "5xx"
				inslog.Errorf("AM's handler %v returns error: %v", name, err)
			}
			inslog.Debugf("measured time of AM method %v is %v", name, latency)

			ctx = insmetrics.ChangeTags(
				ctx,
				tag.Insert(tagMethod, name),
				tag.Insert(tagResult, code),
			)
			stats.Record(ctx, statCalls.M(1), statLatency.M(latency.Nanoseconds()/1e6))

			return repl, err
		}
	}
}

// Init initializes handlers and middleware.
func (h *MessageHandler) Init(ctx context.Context) error {
	m := newMiddleware(h)
	h.middleware = m

	h.jetTreeUpdater = jet.NewFetcher(h.Nodes, h.JetStorage, h.Bus, h.JetCoordinator)

	h.setHandlersForLight(m)

	return nil
}

func (h *MessageHandler) OnPulse(ctx context.Context, pn insolar.Pulse) {
	h.FlowHandler.ChangePulse(ctx, pn)
}

func (h *MessageHandler) setHandlersForLight(m *middleware) {
	// Generic.

	h.Bus.MustRegister(insolar.TypeGetCode, h.FlowHandler.WrapBusHandle)
	h.Bus.MustRegister(insolar.TypeGetObject, h.FlowHandler.WrapBusHandle)

	h.Bus.MustRegister(insolar.TypeGetDelegate,
		BuildMiddleware(h.handleGetDelegate,
			instrumentHandler("handleGetDelegate"),
			m.addFieldsToLogger,
			m.checkJet,
			m.waitForHotData))

	h.Bus.MustRegister(insolar.TypeGetChildren,
		BuildMiddleware(h.handleGetChildren,
			instrumentHandler("handleGetChildren"),
			m.addFieldsToLogger,
			m.checkJet,
			m.waitForHotData))

	h.Bus.MustRegister(insolar.TypeSetRecord,
		BuildMiddleware(h.handleSetRecord,
			instrumentHandler("handleSetRecord"),
			m.addFieldsToLogger,
			m.checkJet,
			m.waitForHotData))

	h.Bus.MustRegister(insolar.TypeUpdateObject,
		BuildMiddleware(h.handleUpdateObject,
			instrumentHandler("handleUpdateObject"),
			m.addFieldsToLogger,
			m.checkJet,
			m.waitForHotData))

	h.Bus.MustRegister(insolar.TypeRegisterChild,
		BuildMiddleware(h.handleRegisterChild,
			instrumentHandler("handleRegisterChild"),
			m.addFieldsToLogger,
			m.checkJet,
			m.waitForHotData))

	h.Bus.MustRegister(insolar.TypeSetBlob,
		BuildMiddleware(h.handleSetBlob,
			instrumentHandler("handleSetBlob"),
			m.addFieldsToLogger,
			m.checkJet,
			m.waitForHotData))

	h.Bus.MustRegister(insolar.TypeGetObjectIndex,
		BuildMiddleware(h.handleGetObjectIndex,
			instrumentHandler("handleGetObjectIndex"),
			m.addFieldsToLogger,
			m.checkJet,
			m.waitForHotData))

	h.Bus.MustRegister(insolar.TypeGetPendingRequests,
		BuildMiddleware(h.handleHasPendingRequests,
			instrumentHandler("handleHasPendingRequests"),
			m.addFieldsToLogger,
			m.checkJet,
			m.waitForHotData))

	h.Bus.MustRegister(insolar.TypeGetJet,
		BuildMiddleware(h.handleGetJet,
			instrumentHandler("handleGetJet")))

	h.Bus.MustRegister(insolar.TypeHotRecords,
		BuildMiddleware(h.handleHotRecords,
			instrumentHandler("handleHotRecords"),
			m.releaseHotDataWaiters))

	h.Bus.MustRegister(
		insolar.TypeGetRequest,
		BuildMiddleware(
			h.handleGetRequest,
			instrumentHandler("handleGetRequest"),
			m.checkJet,
		),
	)

	h.Bus.MustRegister(
		insolar.TypeGetPendingRequestID,
		BuildMiddleware(
			h.handleGetPendingRequestID,
			instrumentHandler("handleGetPendingRequestID"),
			m.checkJet,
		),
	)

	h.Bus.MustRegister(insolar.TypeValidateRecord, h.handleValidateRecord)
}

func (h *MessageHandler) handleSetRecord(ctx context.Context, parcel insolar.Parcel) (insolar.Reply, error) {
	msg := parcel.Message().(*message.SetRecord)
	virtRec, err := object.DecodeVirtual(msg.Record)
	if err != nil {
		return nil, errors.Wrap(err, "can't deserialize record")
	}
	jetID := jetFromContext(ctx)

	calculatedID := object.NewRecordIDFromRecord(h.PlatformCryptographyScheme, parcel.Pulse(), virtRec)
	rec := record.MaterialRecord{
		Record: virtRec,
		JetID:  insolar.JetID(jetID),
	}

	err = h.RecordModifier.Set(ctx, *calculatedID, rec)

	if err == object.ErrOverride {
		inslogger.FromContext(ctx).WithField("type", fmt.Sprintf("%T", virtRec)).Warn("set record override")
	} else if err != nil {
		return nil, errors.Wrap(err, "can't save record into storage")
	}

	switch r := virtRec.(type) {
	case object.Request:
		if h.RecentStorageProvider.Count() > h.conf.PendingRequestsLimit {
			return &reply.Error{ErrType: reply.ErrTooManyPendingRequests}, nil
		}
		recentStorage := h.RecentStorageProvider.GetPendingStorage(ctx, jetID)
		recentStorage.AddPendingRequest(ctx, r.GetObject(), *calculatedID)

		err := h.Index.SetRequest(ctx, parcel.Pulse(), r.GetObject(), *calculatedID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to set a request")
		}
		err = h.IndexState.SetLifelineUsage(ctx, parcel.Pulse(), r.GetObject())
		if err != nil {
			return nil, errors.Wrapf(err, "can't update lifeline usage")
		}
	case *object.ResultRecord:
		recentStorage := h.RecentStorageProvider.GetPendingStorage(ctx, jetID)
		recentStorage.RemovePendingRequest(ctx, r.Object, *r.Request.Record())

		err := h.Index.SetResultRecord(ctx, parcel.Pulse(), *r.Request.Record(), *calculatedID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to set a result record")
		}
	}

	return &reply.ID{ID: *calculatedID}, nil
}

func (h *MessageHandler) handleSetBlob(ctx context.Context, parcel insolar.Parcel) (insolar.Reply, error) {
	msg := parcel.Message().(*message.SetBlob)
	jetID := jetFromContext(ctx)
	calculatedID := object.CalculateIDForBlob(h.PlatformCryptographyScheme, parcel.Pulse(), msg.Memory)

	_, err := h.BlobAccessor.ForID(ctx, *calculatedID)
	if err == nil {
		return &reply.ID{ID: *calculatedID}, nil
	}
	if err != nil && err != blob.ErrNotFound {
		return nil, err
	}

	err = h.BlobModifier.Set(ctx, *calculatedID, blob.Blob{Value: msg.Memory, JetID: insolar.JetID(jetID)})
	if err == nil {
		return &reply.ID{ID: *calculatedID}, nil
	}
	if err == blob.ErrOverride {
		return &reply.ID{ID: *calculatedID}, nil
	}
	return nil, err
}

func (h *MessageHandler) handleHasPendingRequests(ctx context.Context, parcel insolar.Parcel) (insolar.Reply, error) {
	msg := parcel.Message().(*message.GetPendingRequests)
	jetID := jetFromContext(ctx)

	for _, reqID := range h.RecentStorageProvider.GetPendingStorage(ctx, jetID).GetRequestsForObject(*msg.Object.Record()) {
		if reqID.Pulse() < parcel.Pulse() {
			return &reply.HasPendingRequests{Has: true}, nil
		}
	}

	return &reply.HasPendingRequests{Has: false}, nil
}

func (h *MessageHandler) handleGetJet(ctx context.Context, parcel insolar.Parcel) (insolar.Reply, error) {
	msg := parcel.Message().(*message.GetJet)

	jetID, actual := h.JetStorage.ForID(ctx, msg.Pulse, msg.Object)

	return &reply.Jet{ID: insolar.ID(jetID), Actual: actual}, nil
}

func (h *MessageHandler) handleGetDelegate(ctx context.Context, parcel insolar.Parcel) (insolar.Reply, error) {
	msg := parcel.Message().(*message.GetDelegate)
	jetID := jetFromContext(ctx)

	h.IDLocker.Lock(msg.Head.Record())
	defer h.IDLocker.Unlock(msg.Head.Record())

	idx, err := h.Index.LifelineForID(ctx, parcel.Pulse(), *msg.Head.Record())
	if err == object.ErrLifelineNotFound {
		heavy, err := h.JetCoordinator.Heavy(ctx, parcel.Pulse())
		if err != nil {
			return nil, err
		}
		idx, err = h.saveIndexFromHeavy(ctx, jetID, msg.Head, heavy, parcel.Pulse())
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch index from heavy")
		}
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to fetch object index")
	}
	err = h.IndexState.SetLifelineUsage(ctx, parcel.Pulse(), *msg.Head.Record())
	if err != nil {
		return nil, errors.Wrapf(err, "can't update lifeline")
	}

	delegateRef, ok := idx.DelegateByKey(msg.AsType)
	if !ok {
		return nil, errors.New("the object has no delegate for this type")
	}

	rep := reply.Delegate{
		Head: delegateRef,
	}

	return &rep, nil
}

func (h *MessageHandler) handleGetChildren(
	ctx context.Context, parcel insolar.Parcel,
) (insolar.Reply, error) {
	msg := parcel.Message().(*message.GetChildren)
	jetID := jetFromContext(ctx)

	h.IDLocker.Lock(msg.Parent.Record())
	defer h.IDLocker.Unlock(msg.Parent.Record())

	idx, err := h.Index.LifelineForID(ctx, parcel.Pulse(), *msg.Parent.Record())
	if err == object.ErrLifelineNotFound {
		heavy, err := h.JetCoordinator.Heavy(ctx, parcel.Pulse())
		if err != nil {
			return nil, err
		}
		idx, err = h.saveIndexFromHeavy(ctx, jetID, msg.Parent, heavy, parcel.Pulse())
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch index from heavy")
		}
		if idx.ChildPointer == nil {
			return &reply.Children{Refs: nil, NextFrom: nil}, nil
		}
	} else if err != nil {
		return nil, errors.Wrap(err, "failed to fetch object index")
	}
	err = h.IndexState.SetLifelineUsage(ctx, parcel.Pulse(), *msg.Parent.Record())
	if err != nil {
		return nil, errors.Wrapf(err, "can't update lifeline")
	}

	var (
		refs         []insolar.Reference
		currentChild *insolar.ID
	)

	// Counting from specified child or the latest.
	if msg.FromChild != nil {
		currentChild = msg.FromChild
	} else {
		currentChild = idx.ChildPointer
	}

	// The object has no children.
	if currentChild == nil {
		return &reply.Children{Refs: nil, NextFrom: nil}, nil
	}

	var childJet *insolar.ID
	onHeavy, err := h.JetCoordinator.IsBeyondLimit(ctx, parcel.Pulse(), currentChild.Pulse())
	if err != nil && err != pulse.ErrNotFound {
		return nil, err
	}
	if onHeavy {
		node, err := h.JetCoordinator.Heavy(ctx, parcel.Pulse())
		if err != nil {
			return nil, err
		}
		return reply.NewGetChildrenRedirect(h.DelegationTokenFactory, parcel, node, *currentChild)
	}

	childJetID, actual := h.JetStorage.ForID(ctx, currentChild.Pulse(), *msg.Parent.Record())
	childJet = (*insolar.ID)(&childJetID)

	if !actual {
		actualJet, err := h.jetTreeUpdater.Fetch(ctx, *msg.Parent.Record(), currentChild.Pulse())
		if err != nil {
			return nil, err
		}
		childJet = actualJet
	}

	// Try to fetch the first child.
	_, err = h.RecordAccessor.ForID(ctx, *currentChild)

	if err == object.ErrNotFound {
		node, err := h.JetCoordinator.NodeForJet(ctx, *childJet, parcel.Pulse(), currentChild.Pulse())
		if err != nil {
			return nil, err
		}
		return reply.NewGetChildrenRedirect(h.DelegationTokenFactory, parcel, node, *currentChild)
	}

	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch child")
	}

	counter := 0
	for currentChild != nil {
		// We have enough results.
		if counter >= msg.Amount {
			return &reply.Children{Refs: refs, NextFrom: currentChild}, nil
		}
		counter++

		rec, err := h.RecordAccessor.ForID(ctx, *currentChild)

		// We don't have this child reference. Return what was collected.
		if err == object.ErrNotFound {
			return &reply.Children{Refs: refs, NextFrom: currentChild}, nil
		}
		if err != nil {
			return nil, errors.New("failed to retrieve children")
		}

		virtRec := rec.Record
		childRec, ok := virtRec.(*object.ChildRecord)
		if !ok {
			return nil, errors.New("failed to retrieve children")
		}
		currentChild = childRec.PrevChild

		// Skip records later than specified pulse.
		recPulse := childRec.Ref.Record().Pulse()
		if msg.FromPulse != nil && recPulse > *msg.FromPulse {
			continue
		}
		refs = append(refs, childRec.Ref)
	}

	return &reply.Children{Refs: refs, NextFrom: nil}, nil
}

func (h *MessageHandler) handleGetRequest(ctx context.Context, parcel insolar.Parcel) (insolar.Reply, error) {
	msg := parcel.Message().(*message.GetRequest)

	rec, err := h.RecordAccessor.ForID(ctx, msg.Request)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch request")
	}

	virtRec := rec.Record
	req, ok := virtRec.(*object.RequestRecord)
	if !ok {
		return nil, errors.New("failed to decode request")
	}

	rep := reply.Request{
		ID:     msg.Request,
		Record: object.EncodeVirtual(req),
	}

	return &rep, nil
}

func (h *MessageHandler) handleGetPendingRequestID(ctx context.Context, parcel insolar.Parcel) (insolar.Reply, error) {
	jetID := jetFromContext(ctx)
	msg := parcel.Message().(*message.GetPendingRequestID)

	requests := h.RecentStorageProvider.GetPendingStorage(ctx, jetID).GetRequestsForObject(msg.ObjectID)
	if len(requests) == 0 {
		return &reply.Error{ErrType: reply.ErrNoPendingRequests}, nil
	}

	rep := reply.ID{
		ID: requests[0],
	}

	return &rep, nil
}

func (h *MessageHandler) handleUpdateObject(ctx context.Context, parcel insolar.Parcel) (insolar.Reply, error) {
	msg := parcel.Message().(*message.UpdateObject)
	jetID := jetFromContext(ctx)
	logger := inslogger.FromContext(ctx).WithFields(map[string]interface{}{
		"object": msg.Object.Record().DebugString(),
		"pulse":  parcel.Pulse(),
	})

	virtRec, err := object.DecodeVirtual(msg.Record)
	if err != nil {
		return nil, errors.Wrap(err, "can't deserialize record")
	}
	state, ok := virtRec.(object.State)
	if !ok {
		return nil, errors.New("wrong object state record")
	}

	err = h.IndexState.SetLifelineUsage(ctx, parcel.Pulse(), *msg.Object.Record())
	if err != nil {
		return nil, errors.Wrapf(err, "can't update lifeline")
	}

	calculatedID := object.CalculateIDForBlob(h.PlatformCryptographyScheme, parcel.Pulse(), msg.Memory)
	// FIXME: temporary fix. If we calculate blob id on the client, pulse can change before message sending and this
	//  id will not match the one calculated on the server.
	err = h.BlobModifier.Set(ctx, *calculatedID, blob.Blob{JetID: insolar.JetID(jetID), Value: msg.Memory})
	if err != nil && err != blob.ErrOverride {
		return nil, errors.Wrap(err, "failed to set blob")
	}

	switch s := state.(type) {
	case *object.ActivateRecord:
		s.Memory = calculatedID
	case *object.AmendRecord:
		s.Memory = calculatedID
	}

	h.IDLocker.Lock(msg.Object.Record())
	defer h.IDLocker.Unlock(msg.Object.Record())

	idx, err := h.Index.LifelineForID(ctx, parcel.Pulse(), *msg.Object.Record())
	// No index on our node.
	if err == object.ErrLifelineNotFound {
		if state.ID() == object.StateActivation {
			// We are activating the object. There is no index for it anywhere.
			idx = object.Lifeline{State: object.StateUndefined}
		} else {
			logger.Debug("failed to fetch index (fetching from heavy)")
			// We are updating object. Index should be on the heavy executor.
			heavy, err := h.JetCoordinator.Heavy(ctx, parcel.Pulse())
			if err != nil {
				return nil, err
			}
			idx, err = h.saveIndexFromHeavy(ctx, jetID, msg.Object, heavy, parcel.Pulse())
			if err != nil {
				return nil, errors.Wrap(err, "failed to fetch index from heavy")
			}
		}
	} else if err != nil {
		return nil, err
	}
	err = h.IndexState.SetLifelineUsage(ctx, parcel.Pulse(), *msg.Object.Record())
	if err != nil {
		return nil, errors.Wrapf(err, "can't update lifeline")
	}

	if err = validateState(idx.State, state.ID()); err != nil {
		return &reply.Error{ErrType: reply.ErrDeactivated}, nil
	}

	recID := object.NewRecordIDFromRecord(h.PlatformCryptographyScheme, parcel.Pulse(), virtRec)

	// Index exists and latest record id does not match (preserving chain consistency).
	// For the case when vm can't save or send result to another vm and it tries to update the same record again
	if idx.LatestState != nil && !state.PrevStateID().Equal(*idx.LatestState) && idx.LatestState != recID {
		return nil, errors.New("invalid state record")
	}

	id := object.NewRecordIDFromRecord(h.PlatformCryptographyScheme, parcel.Pulse(), virtRec)
	rec := record.MaterialRecord{
		Record: virtRec,
		JetID:  insolar.JetID(jetID),
	}

	err = h.RecordModifier.Set(ctx, *id, rec)

	if err == object.ErrOverride {
		logger.WithField("type", fmt.Sprintf("%T", virtRec)).Warn("set record override (#1)")
		id = recID
	} else if err != nil {
		return nil, errors.Wrap(err, "can't save record into storage")
	}
	idx.LatestState = id
	idx.State = state.ID()
	if state.ID() == object.StateActivation {
		idx.Parent = state.(*object.ActivateRecord).Parent
	}

	idx.LatestUpdate = parcel.Pulse()
	idx.JetID = insolar.JetID(jetID)
	err = h.Index.SetLifeline(ctx, parcel.Pulse(), *msg.Object.Record(), idx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to save lifeline")
	}

	logger.WithField("state", idx.LatestState.DebugString()).Debug("saved object")

	rep := reply.Object{
		Head:         msg.Object,
		State:        *idx.LatestState,
		Prototype:    state.GetImage(),
		IsPrototype:  state.GetIsPrototype(),
		ChildPointer: idx.ChildPointer,
		Parent:       idx.Parent,
	}
	return &rep, nil
}

func (h *MessageHandler) handleRegisterChild(ctx context.Context, parcel insolar.Parcel) (insolar.Reply, error) {
	logger := inslogger.FromContext(ctx)

	msg := parcel.Message().(*message.RegisterChild)
	jetID := jetFromContext(ctx)
	r, err := object.DecodeVirtual(msg.Record)
	if err != nil {
		return nil, errors.Wrap(err, "can't deserialize record")
	}
	childRec, ok := r.(*object.ChildRecord)
	if !ok {
		return nil, errors.New("wrong child record")
	}

	h.IDLocker.Lock(msg.Parent.Record())
	defer h.IDLocker.Unlock(msg.Parent.Record())

	var child *insolar.ID
	idx, err := h.Index.LifelineForID(ctx, parcel.Pulse(), *msg.Parent.Record())
	if err == object.ErrLifelineNotFound {
		heavy, err := h.JetCoordinator.Heavy(ctx, parcel.Pulse())
		if err != nil {
			return nil, err
		}
		idx, err = h.saveIndexFromHeavy(ctx, jetID, msg.Parent, heavy, parcel.Pulse())
		if err != nil {
			return nil, errors.Wrap(err, "failed to fetch index from heavy")
		}
	} else if err != nil {
		return nil, err
	}
	err = h.IndexState.SetLifelineUsage(ctx, parcel.Pulse(), *msg.Parent.Record())
	if err != nil {
		return nil, errors.Wrapf(err, "can't update lifeline")
	}

	recID := object.NewRecordIDFromRecord(h.PlatformCryptographyScheme, parcel.Pulse(), childRec)

	// Children exist and pointer does not match (preserving chain consistency).
	// For the case when vm can't save or send result to another vm and it tries to update the same record again
	if idx.ChildPointer != nil && !childRec.PrevChild.Equal(*idx.ChildPointer) && idx.ChildPointer != recID {
		return nil, errors.New("invalid child record")
	}

	child = object.NewRecordIDFromRecord(h.PlatformCryptographyScheme, parcel.Pulse(), childRec)
	rec := record.MaterialRecord{
		Record: childRec,
		JetID:  insolar.JetID(jetID),
	}

	err = h.RecordModifier.Set(ctx, *child, rec)

	if err == object.ErrOverride {
		logger.WithField("type", fmt.Sprintf("%T", r)).Warn("set record override (#2)")
		child = recID
	} else if err != nil {
		return nil, errors.Wrap(err, "can't save record into storage")
	}

	idx.ChildPointer = child
	if msg.AsType != nil {
		idx.SetDelegate(*msg.AsType, msg.Child)
	}
	idx.LatestUpdate = parcel.Pulse()
	idx.JetID = insolar.JetID(jetID)
	err = h.Index.SetLifeline(ctx, parcel.Pulse(), *msg.Parent.Record(), idx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to set lifeline")
	}

	return &reply.ID{ID: *child}, nil
}

func (h *MessageHandler) handleValidateRecord(ctx context.Context, parcel insolar.Parcel) (insolar.Reply, error) {
	return &reply.OK{}, nil
}

func (h *MessageHandler) handleGetObjectIndex(ctx context.Context, parcel insolar.Parcel) (insolar.Reply, error) {
	msg := parcel.Message().(*message.GetObjectIndex)

	h.IDLocker.Lock(msg.Object.Record())
	defer h.IDLocker.Unlock(msg.Object.Record())

	idx, err := h.Index.LifelineForID(ctx, parcel.Pulse(), *msg.Object.Record())
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch object index")
	}
	err = h.IndexState.SetLifelineUsage(ctx, parcel.Pulse(), *msg.Object.Record())
	if err != nil {
		return nil, errors.Wrapf(err, "can't update object")
	}

	buf := object.EncodeIndex(idx)

	return &reply.ObjectIndex{Index: buf}, nil
}

func validateState(old object.StateID, new object.StateID) error {
	if old == object.StateDeactivation {
		return ErrObjectDeactivated
	}
	if old == object.StateUndefined && new != object.StateActivation {
		return errors.New("object is not activated")
	}
	if old != object.StateUndefined && new == object.StateActivation {
		return errors.New("object is already activated")
	}
	return nil
}

func (h *MessageHandler) saveIndexFromHeavy(
	ctx context.Context, jetID insolar.ID, obj insolar.Reference, heavy *insolar.Reference, parcelPN insolar.PulseNumber,
) (object.Lifeline, error) {
	genericReply, err := h.Bus.Send(ctx, &message.GetObjectIndex{
		Object: obj,
	}, &insolar.MessageSendOptions{
		Receiver: heavy,
	})
	if err != nil {
		return object.Lifeline{}, errors.Wrap(err, "failed to send")
	}
	rep, ok := genericReply.(*reply.ObjectIndex)
	if !ok {
		return object.Lifeline{}, fmt.Errorf("failed to fetch object index: unexpected reply type %T (reply=%+v)", genericReply, genericReply)
	}
	idx, err := object.DecodeIndex(rep.Index)
	if err != nil {
		return object.Lifeline{}, errors.Wrap(err, "failed to decode")
	}

	idx.JetID = insolar.JetID(jetID)
	err = h.Index.SetLifeline(ctx, parcelPN, *obj.Record(), idx)
	if err != nil {
		return idx, errors.Wrap(err, "failed to save lifeline")
	}

	return idx, nil
}

func (h *MessageHandler) handleHotRecords(ctx context.Context, parcel insolar.Parcel) (insolar.Reply, error) {
	logger := inslogger.FromContext(ctx)

	msg := parcel.Message().(*message.HotData)
	jetID := insolar.JetID(*msg.Jet.Record())

	logger.WithFields(map[string]interface{}{
		"jet": jetID.DebugString(),
	}).Info("received hot data")

	err := h.DropModifier.Set(ctx, msg.Drop)
	if err == drop.ErrOverride {
		err = nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "[jet]: drop error (pulse: %v)", msg.Drop.Pulse)
	}

	pendingStorage := h.RecentStorageProvider.GetPendingStorage(ctx, insolar.ID(jetID))
	logger.Debugf("received %d pending requests", len(msg.PendingRequests))

	var notificationList []insolar.ID
	for objID, objContext := range msg.PendingRequests {
		if !objContext.Active {
			notificationList = append(notificationList, objID)
		}

		objContext.Active = false
		pendingStorage.SetContextToObject(ctx, objID, objContext)
	}

	go func() {
		for _, objID := range notificationList {
			go func(objID insolar.ID) {
				rep, err := h.Bus.Send(ctx, &message.AbandonedRequestsNotification{
					Object: objID,
				}, nil)

				if err != nil {
					logger.Error("failed to notify about pending requests")
					return
				}
				if _, ok := rep.(*reply.OK); !ok {
					logger.Error("received unexpected reply on pending notification")
				}
			}(objID)
		}
	}()

	for _, meta := range msg.HotIndexes {
		decodedIndex, err := object.DecodeIndex(meta.Index)
		if err != nil {
			logger.Error(err)
			continue
		}

		err = h.Index.SetLifeline(ctx, parcel.Pulse(), meta.ObjID, decodedIndex)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to set lifeline")
		}
	}

	h.JetStorage.Update(
		ctx, msg.PulseNumber, true, insolar.JetID(jetID),
	)

	h.jetTreeUpdater.Release(ctx, jetID, msg.PulseNumber)

	return &reply.OK{}, nil
}
