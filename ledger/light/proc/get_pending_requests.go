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

	watermillMsg "github.com/ThreeDotsLabs/watermill/message"
	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/bus"
	"github.com/insolar/insolar/insolar/message"
	"github.com/insolar/insolar/insolar/reply"
	"github.com/insolar/insolar/ledger/light/recentstorage"
)

type GetPendingRequests struct {
	message  *watermillMsg.Message
	msg      *message.GetPendingRequests
	jet      insolar.JetID
	reqPulse insolar.PulseNumber

	Dep struct {
		RecentStorageProvider recentstorage.Provider
		Sender                bus.Sender
	}
}

func NewGetPendingRequests(jetID insolar.JetID, message *watermillMsg.Message, msg *message.GetPendingRequests, reqPulse insolar.PulseNumber) *GetPendingRequests {
	return &GetPendingRequests{
		msg:      msg,
		message:  message,
		jet:      jetID,
		reqPulse: reqPulse,
	}
}

func (p *GetPendingRequests) Proceed(ctx context.Context) error {
	msg := p.msg
	jetID := insolar.ID(p.jet)

	hasPendingRequests := false
	pendingStorage := p.Dep.RecentStorageProvider.GetPendingStorage(ctx, jetID)
	for _, reqID := range pendingStorage.GetRequestsForObject(*msg.Object.Record()) {
		if reqID.Pulse() < p.reqPulse {
			hasPendingRequests = true
			break
		}
	}
	rep := bus.ReplyAsMessage(ctx, &reply.HasPendingRequests{Has: hasPendingRequests})
	p.Dep.Sender.Reply(ctx, p.message, rep)
	return nil
}
