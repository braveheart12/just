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

package proc

import (
	"context"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/bus"
	"github.com/insolar/insolar/insolar/flow"
	"github.com/insolar/insolar/insolar/payload"
	"github.com/insolar/insolar/insolar/record"
	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/insolar/insolar/ledger/light/executor"
	"github.com/insolar/insolar/ledger/light/hot"
	"github.com/insolar/insolar/ledger/object"
	"github.com/pkg/errors"
)

type SetResult struct {
	message  payload.Meta
	result   record.Result
	resultID insolar.ID
	jetID    insolar.JetID

	dep struct {
		writer   hot.WriteAccessor
		filament executor.FilamentModifier
		sender   bus.Sender
		locker   object.IndexLocker
	}
}

func NewSetResult(
	msg payload.Meta,
	res record.Result,
	resID insolar.ID,
	jetID insolar.JetID,
) *SetResult {
	return &SetResult{
		message:  msg,
		result:   res,
		resultID: resID,
		jetID:    jetID,
	}
}

func (p *SetResult) Dep(
	w hot.WriteAccessor,
	s bus.Sender,
	l object.IndexLocker,
	f executor.FilamentModifier,
) {
	p.dep.writer = w
	p.dep.sender = s
	p.dep.locker = l
	p.dep.filament = f
}

func (p *SetResult) Proceed(ctx context.Context) error {
	done, err := p.dep.writer.Begin(ctx, flow.Pulse(ctx))
	if err != nil {
		if err == hot.ErrWriteClosed {
			return flow.ErrCancelled
		}
		return err
	}
	defer done()

	p.dep.locker.Lock(&p.result.Object)
	defer p.dep.locker.Unlock(&p.result.Object)

	err = p.dep.filament.SetResult(ctx, p.resultID, p.jetID, p.result)
	if err == object.ErrOverride {
		inslogger.FromContext(ctx).Errorf("can't save record into storage: %s", err)
		// Since there is no deduplication yet it's quite possible that there will be
		// two writes by the same key. For this reason currently instead of reporting
		// an error we return OK (nil error). When deduplication will be implemented
		// we should change `nil` to `ErrOverride` here.
		return nil
	} else if err != nil {
		return errors.Wrap(err, "failed to store record")
	}

	msg, err := payload.NewMessage(&payload.ID{ID: p.resultID})
	if err != nil {
		return errors.Wrap(err, "failed to create reply")
	}

	go p.dep.sender.Reply(ctx, p.message, msg)

	return nil
}
