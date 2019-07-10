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

package logicrunner

import (
	"context"

	watermillMsg "github.com/ThreeDotsLabs/watermill/message"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/flow"
	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/insolar/insolar/instrumentation/instracer"
)

type GetLedgerPendingRequest struct {
	dep *Dependencies

	Message *watermillMsg.Message
}

func (p *GetLedgerPendingRequest) Present(ctx context.Context, f flow.Flow) error {
	ctx, span := instracer.StartSpan(ctx, "LogicRunner.getLedgerPendingRequest")
	defer span.End()

	lr := p.dep.lr
	es := lr.StateStorage.GetExecutionState(Ref{}.FromSlice(p.Message.Payload))
	if es == nil {
		return nil
	}

	es.getLedgerPendingMutex.Lock()
	defer es.getLedgerPendingMutex.Unlock()

	proc := &UnsafeGetLedgerPendingRequest{
		es:         es,
		dep:        p.dep,
		hasPending: false,
	}

	err := f.Procedure(ctx, proc, true)
	if err != nil {
		inslogger.FromContext(ctx).Debug("GetLedgerPendingRequest.Present err: ", err)
		return nil
	}

	if !proc.hasPending {
		return nil
	}

	es.Broker.StartProcessorIfNeeded(ctx)
	return nil
}

type UnsafeGetLedgerPendingRequest struct {
	dep        *Dependencies
	es         *ExecutionState
	hasPending bool
}

func (u *UnsafeGetLedgerPendingRequest) Proceed(ctx context.Context) error {
	es := u.es
	lr := u.dep.lr

	es.Lock()
	if es.Broker.HasLedgerRequest(ctx) != nil || !es.LedgerHasMoreRequests {
		es.Unlock()
		return nil
	}
	es.Unlock()

	id := *es.Ref.Record()

	requestRef, request, err := lr.ArtifactManager.GetPendingRequest(ctx, id)
	if err != nil {
		if err != insolar.ErrNoPendingRequest {
			inslogger.FromContext(ctx).Debug("GetPendingRequest failed with error")
			return nil
		}
		es.Lock()
		defer es.Unlock()

		select {
		case <-ctx.Done():
			inslogger.FromContext(ctx).Debug("UnsafeGetLedgerPendingRequest: pulse changed. Do nothing")
			return nil
		default:
		}

		es.LedgerHasMoreRequests = false
		return nil
	}
	es.Lock()
	defer es.Unlock()

	// <<<<<<< HEAD
	// 	msg := parcel.Message().(*message.CallMethod)
	//
	// 	pulse, err := u.dep.lr.PulseAccessor.Latest(ctx)
	// 	if err != nil {
	// 		return errors.Wrap(err, "[ UnsafeGetLedgerPendingRequest::Proceed ] Couldn't get current pulse")
	// 	}
	// 	caller := msg.IncomingRequest.Caller
	// 	if caller.IsEmpty() {
	// 		parcel.SetSender(msg.IncomingRequest.APISender)
	// 	} else {
	// 		sender, err := u.dep.lr.JetCoordinator.VirtualExecutorForObject(ctx, *caller.Record(), pulse.PulseNumber)
	// 		if err != nil {
	// 			return errors.Wrap(err, "[ UnsafeGetLedgerPendingRequest::Proceed ] Couldn't get current VE for object")
	// 		}
	//
	// 		parcel.SetSender(*sender)
	// 	}

	// =======
	pulse := lr.pulse(ctx).PulseNumber
	// >>>>>>> 4acd6872dcb731a128b469fe5e2edc0e00faf979
	authorized, err := lr.JetCoordinator.IsAuthorized(
		ctx, insolar.DynamicRoleVirtualExecutor, id, pulse, lr.JetCoordinator.Me(),
	)
	if err != nil {
		inslogger.FromContext(ctx).Debug("Authorization failed with error in getLedgerPendingRequest")
		return nil
	}

	if !authorized {
		inslogger.FromContext(ctx).Debug("pulse changed, can't process abandoned messages for this object")
		return nil
	}

	select {
	case <-ctx.Done():
		inslogger.FromContext(ctx).Debug("UnsafeGetLedgerPendingRequest: pulse changed. Do nothing")
		return nil
	default:
	}

	u.hasPending = true
	es.LedgerHasMoreRequests = true

	t := NewTranscript(ctx, requestRef, *request)
	t.FromLedger = true
	es.Broker.Prepend(ctx, true, t)

	return nil
}
