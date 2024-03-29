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

package dispatcher

import (
	"context"
	"strconv"
	"sync/atomic"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/insolar/insolar/insolar/bus"
	"github.com/insolar/insolar/instrumentation/instracer"
	"github.com/pkg/errors"

	"github.com/insolar/insolar/instrumentation/inslogger"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/flow"
	"github.com/insolar/insolar/insolar/flow/internal/pulse"
	"github.com/insolar/insolar/insolar/flow/internal/thread"
)

type Dispatcher struct {
	handles struct {
		present flow.MakeHandle
		future  flow.MakeHandle
		past    flow.MakeHandle
	}
	controller         *thread.Controller
	currentPulseNumber uint32
}

func NewDispatcher(present flow.MakeHandle, future flow.MakeHandle, past flow.MakeHandle) *Dispatcher {
	d := &Dispatcher{
		controller: thread.NewController(),
	}
	d.handles.present = present
	d.handles.future = future
	d.handles.past = past
	d.currentPulseNumber = insolar.FirstPulseNumber
	return d
}

// ChangePulse is a handle for pulse change vent.
func (d *Dispatcher) ChangePulse(ctx context.Context, pulse insolar.Pulse) {
	d.controller.Pulse()
	atomic.StoreUint32(&d.currentPulseNumber, uint32(pulse.PulseNumber))
}

func (d *Dispatcher) getHandleByPulse(msgPulseNumber insolar.PulseNumber) flow.MakeHandle {
	currentPulse := atomic.LoadUint32(&d.currentPulseNumber)
	if uint32(msgPulseNumber) > currentPulse {
		return d.handles.future
	}
	if uint32(msgPulseNumber) < currentPulse {
		return d.handles.past
	}
	return d.handles.present
}

func (d *Dispatcher) InnerSubscriber(msg *message.Message) ([]*message.Message, error) {
	ctx := context.Background()
	ctx = inslogger.ContextWithTrace(ctx, msg.Metadata.Get(bus.MetaTraceID))
	parentSpan, err := instracer.Deserialize([]byte(msg.Metadata.Get(bus.MetaSpanData)))
	if err == nil {
		ctx = instracer.WithParentSpan(ctx, parentSpan)
	} else {
		inslogger.FromContext(ctx).Error(err)
	}
	logger := inslogger.FromContext(ctx)
	go func() {
		f := thread.NewThread(msg, d.controller)
		err := f.Run(ctx, d.handles.present(msg))
		if err != nil {
			logger.Error("Handling failed: ", err)
		}
	}()
	return nil, nil
}

// Process handles incoming message.
func (d *Dispatcher) Process(msg *message.Message) ([]*message.Message, error) {
	ctx := context.Background()

	for k, v := range msg.Metadata {
		ctx, _ = inslogger.WithField(ctx, k, v)
	}
	logger := inslogger.FromContext(ctx)

	pn, err := insolar.NewPulseNumberFromStr(msg.Metadata.Get(bus.MetaPulse))
	if err != nil {
		logger.Error("failed to handle message: ", err)
		return nil, nil
	}
	ctx = pulse.ContextWith(ctx, pn)
	ctx = inslogger.ContextWithTrace(ctx, msg.Metadata.Get(bus.MetaTraceID))
	parentSpan := instracer.MustDeserialize([]byte(msg.Metadata.Get(bus.MetaSpanData)))
	ctx = instracer.WithParentSpan(ctx, parentSpan)
	go func() {
		f := thread.NewThread(msg, d.controller)
		handle := d.getHandleByPulse(pn)
		err := f.Run(ctx, handle(msg))
		if err != nil {
			logger.Error(errors.Wrap(err, "Handling failed: "))
		}
	}()
	return nil, nil
}

func pulseFromString(p string) (insolar.PulseNumber, error) {
	u64, err := strconv.ParseUint(p, 10, 32)
	if err != nil {
		return insolar.PulseNumber(0), errors.Wrap(err, "can't convert string value to pulse")
	}
	pInt := uint32(u64)
	return insolar.PulseNumber(pInt), nil
}
