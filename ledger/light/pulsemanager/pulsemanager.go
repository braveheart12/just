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

package pulsemanager

import (
	"context"
	"sync"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/jet"
	"github.com/insolar/insolar/insolar/node"
	"github.com/insolar/insolar/insolar/pulse"
	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/insolar/insolar/instrumentation/instracer"
	"github.com/insolar/insolar/ledger/light/artifactmanager"
	"github.com/insolar/insolar/ledger/light/executor"
	"github.com/insolar/insolar/ledger/light/hot"
	"github.com/insolar/insolar/ledger/light/replication"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

var (
	errZeroNodes = errors.New("zero nodes from network")
	errNoPulse   = errors.New("no previous pulses")
)

//go:generate minimock -i github.com/insolar/insolar/ledger/light/pulsemanager.ActiveListSwapper -o ../../../testutils -s _mock.go

type ActiveListSwapper interface {
	MoveSyncToActive(ctx context.Context, number insolar.PulseNumber) error
}

// PulseManager implements insolar.PulseManager.
type PulseManager struct {
	Bus               insolar.MessageBus        `inject:""`
	NodeNet           insolar.NodeNetwork       `inject:""`
	GIL               insolar.GlobalInsolarLock `inject:""`
	ActiveListSwapper ActiveListSwapper         `inject:""`
	MessageHandler    *artifactmanager.MessageHandler

	JetReleaser hot.JetReleaser `inject:""`

	JetModifier jet.Modifier `inject:""`
	JetSplitter executor.JetSplitter

	NodeSetter node.Modifier `inject:""`
	Nodes      node.Accessor `inject:""`

	PulseAccessor   pulse.Accessor   `inject:""`
	PulseCalculator pulse.Calculator `inject:""`
	PulseAppender   pulse.Appender   `inject:""`

	LightReplicator replication.LightReplicator
	HotSender       executor.HotSender

	WriteManager hot.WriteManager

	// setLock locks Set method call.
	setLock sync.RWMutex
}

// NewPulseManager creates PulseManager instance.
func NewPulseManager(
	jetSplitter executor.JetSplitter,
	lightToHeavySyncer replication.LightReplicator,
	writeManager hot.WriteManager,
	hotSender executor.HotSender,
) *PulseManager {
	pm := &PulseManager{
		JetSplitter:     jetSplitter,
		LightReplicator: lightToHeavySyncer,
		WriteManager:    writeManager,
		HotSender:       hotSender,
	}
	return pm
}

// Set set's new pulse and closes current jet drop.
func (m *PulseManager) Set(ctx context.Context, newPulse insolar.Pulse) error {
	logger := inslogger.FromContext(ctx)

	m.setLock.Lock()
	defer m.setLock.Unlock()

	defer func() {
		err := m.Bus.OnPulse(ctx, newPulse)
		if err != nil {
			inslogger.FromContext(ctx).Error(errors.Wrap(err, "MessageBus OnPulse() returns error"))
		}
	}()

	ctx, span := instracer.StartSpan(
		ctx, "pulse.process", trace.WithSampler(trace.AlwaysSample()),
	)
	span.AddAttributes(
		trace.Int64Attribute("pulse.PulseNumber", int64(newPulse.PulseNumber)),
	)
	defer span.End()

	jets, endedPulse, err := m.setUnderGilSection(ctx, newPulse)
	if err != nil {
		if err == errZeroNodes || err == errNoPulse {
			return nil
		}
		panic(errors.Wrap(err, "under gil error"))
	}

	err = m.WriteManager.CloseAndWait(ctx, endedPulse.PulseNumber)
	if err != nil {
		panic(errors.Wrap(err, "can't close pulse for writing"))
	}
	err = m.HotSender.SendHot(ctx, endedPulse.PulseNumber, newPulse.PulseNumber, jets)
	if err != nil {
		logger.Error("send Hot failed: ", err)
	}
	go m.LightReplicator.NotifyAboutPulse(ctx, newPulse.PulseNumber)

	err = m.WriteManager.Open(ctx, newPulse.PulseNumber)
	if err != nil {
		panic(errors.Wrap(err, "failed to open pulse for writing"))
	}

	m.MessageHandler.OnPulse(ctx, newPulse)
	return nil
}

func (m *PulseManager) setUnderGilSection(ctx context.Context, newPulse insolar.Pulse) (
	[]insolar.JetID, insolar.Pulse, error,
) {
	m.GIL.Acquire(ctx)
	ctx, span := instracer.StartSpan(ctx, "pulse.gil_locked")
	defer span.End()
	defer m.GIL.Release(ctx)

	logger := inslogger.FromContext(ctx)
	logger.WithFields(map[string]interface{}{
		"new_pulse": newPulse.PulseNumber,
	}).Debugf("received pulse")

	// Dealing with node lists.
	{
		err := m.ActiveListSwapper.MoveSyncToActive(ctx, newPulse.PulseNumber)
		if err != nil {
			panic(errors.Wrap(err, "failed to apply new active node list"))
		}
		fromNetwork := m.NodeNet.GetWorkingNodes()
		if len(fromNetwork) == 0 {
			logger.Warnf("received zero nodes for pulse %d", newPulse.PulseNumber)
			return nil, insolar.Pulse{}, errZeroNodes
		}
		toSet := make([]insolar.Node, 0, len(fromNetwork))
		for _, n := range fromNetwork {
			toSet = append(toSet, insolar.Node{ID: n.ID(), Role: n.Role()})
		}
		err = m.NodeSetter.Set(newPulse.PulseNumber, toSet)
		if err != nil {
			panic(errors.Wrap(err, "call of SetActiveNodes failed"))
		}
	}

	if err := m.PulseAppender.Append(ctx, newPulse); err != nil {
		panic(errors.Wrap(err, "failed to add pulse"))
	}

	// Updating jet tree if its network start.
	{
		_, err := m.PulseCalculator.Backwards(ctx, newPulse.PulseNumber, 2)
		if err != nil {
			if err == pulse.ErrNotFound {
				err := m.JetModifier.Update(ctx, newPulse.PulseNumber, true, insolar.ZeroJetID)
				if err != nil {
					panic(errors.Wrap(err, "failed tp update jets"))
				}
			} else {
				panic(errors.Wrap(err, "failed to calculate previous pulse"))
			}
		}
	}

	endedPulse, err := m.PulseCalculator.Backwards(ctx, newPulse.PulseNumber, 1)
	if err != nil {
		if err == pulse.ErrNotFound {
			return nil, insolar.Pulse{}, errNoPulse
		}
		panic(errors.Wrap(err, "failed to calculate ended pulse"))
	}

	jets, err := m.JetSplitter.Do(ctx, endedPulse.PulseNumber, newPulse.PulseNumber)
	if err != nil {
		panic(errors.Wrap(err, "failed to split jets"))
	}

	m.JetReleaser.ThrowTimeout(ctx, newPulse.PulseNumber)
	return jets, endedPulse, nil
}
