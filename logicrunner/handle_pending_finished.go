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

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/bus"
	"github.com/insolar/insolar/insolar/flow"
	"github.com/insolar/insolar/insolar/message"
	"github.com/insolar/insolar/insolar/payload"
	"github.com/insolar/insolar/insolar/reply"
	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/pkg/errors"
)

type HandlePendingFinished struct {
	dep *Dependencies

	Message payload.Meta
	Parcel  insolar.Parcel
}

func (h *HandlePendingFinished) Present(ctx context.Context, f flow.Flow) error {
	ctx = loggerWithTargetID(ctx, h.Parcel)
	lr := h.dep.lr
	inslogger.FromContext(ctx).Debug("HandlePendingFinished.Present starts ...")
	replyOk := bus.ReplyAsMessage(ctx, &reply.OK{})

	msg := h.Parcel.Message().(*message.PendingFinished)
	ref := msg.DefaultTarget()

	broker := lr.StateStorage.UpsertExecutionState(*ref)

	broker.executionState.Lock()
	broker.executionState.pending = message.NotPending
	if !broker.currentList.Empty() {
		broker.executionState.Unlock()
		return errors.New("[ HandlePendingFinished ] received PendingFinished when we are already executing")
	}
	broker.executionState.Unlock()

	broker.StartProcessorIfNeeded(ctx)

	h.dep.Sender.Reply(ctx, h.Message, replyOk)
	return nil
}
