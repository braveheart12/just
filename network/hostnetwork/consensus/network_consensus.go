/*
 *    Copyright 2018 Insolar
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package consensus

import (
	"context"

	"github.com/insolar/insolar/core"
	"github.com/insolar/insolar/log"
	"github.com/insolar/insolar/network/consensus"
	"github.com/insolar/insolar/network/hostnetwork/hosthandler"
)

type participantWrapper struct {
	node *core.ActiveNode
}

// GetActiveNode implements Participant interface for ActiveNode wrapper.
func (an *participantWrapper) GetActiveNode() *core.ActiveNode {
	return an.node
}

type selfWrapper struct {
	keeper consensus.NodeKeeper
}

// GetActiveNode implements Participant interface for NodeKeeper wrapper.
func (s *selfWrapper) GetActiveNode() *core.ActiveNode {
	return s.keeper.GetSelf()
}

// NetworkConsensus binds all functionality related to consensus with the network layer
type NetworkConsensus struct {
	consensus       consensus.Consensus
	communicatorSnd *communicatorSender
	communicatorRcv *communicatorReceiver
	keeper          consensus.NodeKeeper
	self            *selfWrapper
}

// ProcessPulse is called when we get new pulse from pulsar. Should be called in goroutine
func (ic *NetworkConsensus) ProcessPulse(ctx context.Context, pulse core.Pulse) {
	activeNodes := ic.keeper.GetActiveNodes()
	if len(activeNodes) == 0 {
		return
	}
	participants := make([]consensus.Participant, 0)
	var parts string
	for _, activeNode := range activeNodes {
		if activeNode.NodeID == ic.keeper.GetID() {
			continue
		}
		participants = append(participants, &participantWrapper{activeNode})
		parts += activeNode.NodeID.String() + ", "
	}
	log.Warn("consensus participants: %s", parts)
	success, unsyncList := ic.keeper.SetPulse(pulse.PulseNumber)
	if !success {
		log.Error("ConsensusProcessor: could not set new pulse to NodeKeeper, aborting")
		return
	}
	unsyncCandidates, err := ic.consensus.DoConsensus(ctx, unsyncList, ic.self, participants)
	if err != nil {
		log.Errorf("ConsensusProcessor: error performing consensus steps: %s", err.Error())
	}
	// We have to keep in mind a scenario when DoConsensus takes too long time and a new ProcessPulse is called
	// simultaneously with the current call. It will happen if DoConsensus takes more time than the delay between two
	// consecutive pulses.
	// In this scenario ic.keeper.SetPulse(pulse + 1) will happen earlier than ic.keeper.Sync(syncCandidates, pulse).
	// ic.keeper.SetPulse(pulse + 1) will internally call Sync(nil) to update NodeKeeper's unsync, sync and active lists.
	// That's why we have to pass PulseNumber to ic.keeper.Sync to check relevance of the pulse and to ignore the call
	// if we detect this kind of race condition.
	ic.keeper.Sync(unsyncCandidates, pulse.PulseNumber)
}

// IsPartOfConsensus returns whether we should perform all consensus interactions or not
func (ic *NetworkConsensus) IsPartOfConsensus() bool {
	return ic.keeper.GetSelf() != nil
}

// ReceiverHandler return handler that is responsible to handle consensus network requests
func (ic *NetworkConsensus) ReceiverHandler() consensus.CommunicatorReceiver {
	return ic.communicatorRcv
}

// SetNodeKeeper set NodeKeeper for the processor to integrate Processor with unsync -> sync -> active pipeline
func (ic *NetworkConsensus) SetNodeKeeper(keeper consensus.NodeKeeper) {
	ic.keeper = keeper
	ic.self = &selfWrapper{keeper}
	ic.communicatorSnd.keeper = keeper
	ic.communicatorRcv.keeper = keeper
}

// NewInsolarConsensus creates new object to handle all consensus events
func NewInsolarConsensus(handler hosthandler.HostHandler) consensus.Processor {
	communicatorSnd := &communicatorSender{handler: handler}
	communicatorRcv := &communicatorReceiver{handler: handler}
	return &NetworkConsensus{
		consensus:       consensus.NewConsensus(communicatorSnd),
		communicatorSnd: communicatorSnd,
		communicatorRcv: communicatorRcv,
	}
}
