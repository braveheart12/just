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
	"encoding/gob"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/message"
)

type CaseRequest struct {
	Parcel     insolar.Parcel
	Request    insolar.Reference
	MessageBus insolar.MessageBus
	Reply      insolar.Reply
	Error      string
}

// CaseBinder is a whole result of executor efforts on every object it seen on this pulse
type CaseBind struct {
	Requests []CaseRequest
}

func NewCaseBind() *CaseBind {
	return &CaseBind{Requests: make([]CaseRequest, 0)}
}

func NewCaseBindFromValidateMessage(ctx context.Context, mb insolar.MessageBus, msg *message.ValidateCaseBind) *CaseBind {
	res := &CaseBind{
		Requests: make([]CaseRequest, len(msg.Requests)),
	}
	for i, req := range msg.Requests {
		// TODO: here we used message bus player
		res.Requests[i] = CaseRequest{
			Parcel:  req.Parcel,
			Request: req.Request,
			Reply:   req.Reply,
			Error:   req.Error,
		}
	}
	return res
}

func NewCaseBindFromExecutorResultsMessage(msg *message.ExecutorResults) *CaseBind {
	panic("not implemented")
}

func (cb *CaseBind) getCaseBindForMessage(_ context.Context) []message.CaseBindRequest {
	return make([]message.CaseBindRequest, 0)
	// TODO: we don't validate at the moment, just send empty case bind
	//
	//if cb == nil {
	//	return make([]message.CaseBindRequest, 0)
	//}
	//
	//requests := make([]message.CaseBindRequest, len(cb.Requests))
	//
	//for i, req := range cb.Requests {
	//	var buf bytes.Buffer
	//	err := req.MessageBus.(insolar.TapeWriter).WriteTape(ctx, &buf)
	//	if err != nil {
	//		panic("couldn't write tape: " + err.Error())
	//	}
	//	requests[i] = message.CaseBindRequest{
	//		Parcel:         req.Parcel,
	//		Request:        req.Request,
	//		MessageBusTape: buf.Bytes(),
	//		Reply:          req.Reply,
	//		Error:          req.Error,
	//	}
	//}
	//
	//return requests
}

func (cb *CaseBind) ToValidateMessage(ctx context.Context, ref Ref, pulse insolar.Pulse) *message.ValidateCaseBind {
	res := &message.ValidateCaseBind{
		RecordRef: ref,
		Requests:  cb.getCaseBindForMessage(ctx),
		Pulse:     pulse,
	}
	return res
}

func (cb *CaseBind) NewRequest(p insolar.Parcel, request Ref, mb insolar.MessageBus) *CaseRequest {
	res := CaseRequest{
		Parcel:     p,
		Request:    request,
		MessageBus: mb,
	}
	cb.Requests = append(cb.Requests, res)
	return &cb.Requests[len(cb.Requests)-1]
}

type CaseBindReplay struct {
	Pulse    insolar.Pulse
	CaseBind CaseBind
	Request  int
	Record   int
	Steps    int
	Fail     int
}

func NewCaseBindReplay(cb CaseBind) *CaseBindReplay {
	return &CaseBindReplay{
		CaseBind: cb,
		Request:  -1,
		Record:   -1,
	}
}

func (r *CaseBindReplay) NextRequest() *CaseRequest {
	if r.Request+1 >= len(r.CaseBind.Requests) {
		return nil
	}
	r.Request++
	return &r.CaseBind.Requests[r.Request]
}

func (lr *LogicRunner) Validate(ctx context.Context, ref Ref, p insolar.Pulse, cb CaseBind) (int, error) {
	//os := LogicRunner.UpsertObjectState(ref)
	//vs := os.StartValidation(ref)
	//
	//vs.Lock()
	//defer vs.Unlock()
	//
	//checker := &ValidationChecker{
	//	LogicRunner: LogicRunner,
	//	cb: NewCaseBindReplay(cb),
	//}
	//vs.Behaviour = checker
	//
	//for {
	//	request := checker.NextRequest()
	//	if request == nil {
	//		break
	//	}
	//
	//	traceID := "TODO" // FIXME
	//
	//	ctx = inslogger.ContextWithTrace(ctx, traceID)
	//
	//	// TODO: here we were injecting message bus into context
	//
	//	sender := request.Parcel.GetSender()
	//	vs.Current = &CurrentExecution{
	//		Context:       ctx,
	//		Request:       &request.Request,
	//		RequesterNode: &sender,
	//	}
	//
	//	rep, err := func() (insolar.Reply, error) {
	//		vs.Unlock()
	//		defer vs.Lock()
	//		return LogicRunner.executeAndReply(ctx, vs, request.Parcel)
	//	}()
	//
	//	err = vs.Behaviour.Result(rep, err)
	//	if err != nil {
	//		return 0, errors.Wrap(err, "validation step failed")
	//	}
	//}
	return 1, nil
}

func init() {
	gob.Register(&CaseRequest{})
	gob.Register(&CaseBind{})
}
