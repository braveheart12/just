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

package handle

import (
	"context"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/flow"
	"github.com/insolar/insolar/insolar/message"
	"github.com/insolar/insolar/insolar/payload"
	"github.com/insolar/insolar/ledger/light/proc"
)

type GetChildren struct {
	dep    *proc.Dependencies
	meta   payload.Meta
	parcel insolar.Parcel
}

func NewGetChildren(dep *proc.Dependencies, meta payload.Meta, parcel insolar.Parcel) *GetChildren {
	return &GetChildren{
		dep:    dep,
		meta:   meta,
		parcel: parcel,
	}
}

func (s *GetChildren) Present(ctx context.Context, f flow.Flow) error {
	msg := s.parcel.Message().(*message.GetChildren)

	var jetID insolar.JetID
	if s.parcel.DelegationToken() == nil {
		jet := proc.NewFetchJet(*msg.DefaultTarget().Record(), flow.Pulse(ctx), s.meta)
		s.dep.FetchJet(jet)
		if err := f.Procedure(ctx, jet, false); err != nil {
			return err
		}
		hot := proc.NewWaitHot(jet.Result.Jet, flow.Pulse(ctx), s.meta)
		s.dep.WaitHot(hot)
		if err := f.Procedure(ctx, hot, false); err != nil {
			return err
		}

		jetID = jet.Result.Jet
	} else {
		// Workaround to fetch object states.
		jet := proc.NewFetchJet(*msg.DefaultTarget().Record(), msg.FromChild.Pulse(), s.meta)
		s.dep.FetchJet(jet)
		if err := f.Procedure(ctx, jet, false); err != nil {
			return err
		}
		jetID = jet.Result.Jet
	}

	getIndex := proc.NewEnsureIndex(msg.Parent, jetID, s.meta, flow.Pulse(ctx))
	s.dep.GetIndex(getIndex)
	if err := f.Procedure(ctx, getIndex, false); err != nil {
		return err
	}

	getChildren := proc.NewGetChildren(msg, s.parcel, s.meta)
	s.dep.GetChildren(getChildren)
	return f.Procedure(ctx, getChildren, false)
}
