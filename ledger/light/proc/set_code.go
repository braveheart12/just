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

	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/pkg/errors"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/bus"
	"github.com/insolar/insolar/insolar/flow"
	"github.com/insolar/insolar/insolar/payload"
	"github.com/insolar/insolar/insolar/record"
	"github.com/insolar/insolar/ledger/light/hot"
	"github.com/insolar/insolar/ledger/object"
)

type SetCode struct {
	message  payload.Meta
	record   record.Virtual
	code     []byte
	recordID insolar.ID
	jetID    insolar.JetID

	dep struct {
		writer  hot.WriteAccessor
		records object.RecordModifier
		pcs     insolar.PlatformCryptographyScheme
		sender  bus.Sender
	}
}

func NewSetCode(msg payload.Meta, rec record.Virtual, recID insolar.ID, jetID insolar.JetID) *SetCode {
	return &SetCode{
		message:  msg,
		record:   rec,
		recordID: recID,
		jetID:    jetID,
	}
}

func (p *SetCode) Dep(
	w hot.WriteAccessor,
	r object.RecordModifier,
	pcs insolar.PlatformCryptographyScheme,
	s bus.Sender,
) {
	p.dep.writer = w
	p.dep.records = r
	p.dep.pcs = pcs
	p.dep.sender = s
}

func (p *SetCode) Proceed(ctx context.Context) error {
	done, err := p.dep.writer.Begin(ctx, flow.Pulse(ctx))
	if err != nil {
		if err == hot.ErrWriteClosed {
			return flow.ErrCancelled
		}
		return err
	}
	defer done()

	material := record.Material{
		Virtual: &p.record,
		JetID:   p.jetID,
	}

	err = p.dep.records.Set(ctx, p.recordID, material)
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

	msg, err := payload.NewMessage(&payload.ID{ID: p.recordID})
	if err != nil {
		return errors.Wrap(err, "failed to create reply")
	}

	go p.dep.sender.Reply(ctx, p.message, msg)

	return nil
}
