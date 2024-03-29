//
// Modified BSD 3-Clause Clear License
//
// Copyright (c) 2019 Insolar Technologies GmbH
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without modification,
// are permitted (subject to the limitations in the disclaimer below) provided that
// the following conditions are met:
//  * Redistributions of source code must retain the above copyright notice, this list
//    of conditions and the following disclaimer.
//  * Redistributions in binary form must reproduce the above copyright notice, this list
//    of conditions and the following disclaimer in the documentation and/or other materials
//    provided with the distribution.
//  * Neither the name of Insolar Technologies GmbH nor the names of its contributors
//    may be used to endorse or promote products derived from this software without
//    specific prior written permission.
//
// NO EXPRESS OR IMPLIED LICENSES TO ANY PARTY'S PATENT RIGHTS ARE GRANTED
// BY THIS LICENSE. THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS
// AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES,
// INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL
// THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT,
// INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING,
// BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS
// OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
// ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
//
// Notwithstanding any other provisions of this license, it is prohibited to:
//    (a) use this software,
//
//    (b) prepare modifications and derivative works of this software,
//
//    (c) distribute this software (including without limitation in source code, binary or
//        object code form), and
//
//    (d) reproduce copies of this software
//
//    for any commercial purposes, and/or
//
//    for the purposes of making available this software to third parties as a service,
//    including, without limitation, any software-as-a-service, platform-as-a-service,
//    infrastructure-as-a-service or other similar online service, irrespective of
//    whether it competes with the products or services of Insolar Technologies GmbH.
//

// +build never_run

package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/insolar/insolar/network/consensus/gcpv2/core"
)

func TestConsensusMain(t *testing.T) {
	startedAt := time.Now()

	ctx := context.Background()
	logger := inslogger.FromContext(ctx).WithCaller(false)
	logger, _ = logger.WithLevelNumber(insolar.DebugLevel)
	logger, _ = logger.WithFormat(insolar.TextFormat)

	ctx = inslogger.SetLogger(ctx, logger)

	strategy := NewDelayNetStrategy(DelayStrategyConf{
		MinDelay:         10 * time.Millisecond,
		MaxDelay:         30 * time.Millisecond,
		Variance:         0.2,
		SpikeProbability: 0.1,
	})
	network := NewEmuNetwork(strategy, ctx)
	config := NewEmuLocalConfig(ctx)
	primingCloudStateHash := NewEmuNodeStateHash(1234567890)
	nodes := NewEmuNodeIntros(generateNameList(0, 1, 3, 5)...)

	strategyFactory := &EmuRoundStrategyFactory{}
	candidateFeeder := &core.SequencialCandidateFeeder{}

	for i, n := range nodes {
		chronicles := NewEmuChronicles(nodes, i, &primingCloudStateHash)
		node := NewConsensusHost(n.GetDefaultEndpoint().GetNameAddress())
		controlFeeder := &EmuControlFeeder{}
		//if i % 5 == 2 {
		//	controlFeeder.leaveReason = uint32(i) //simulate leave
		//}
		node.ConnectTo(chronicles, network, strategyFactory, candidateFeeder, controlFeeder, config)
	}

	network.Start(ctx)

	go CreateGenerator(10, 2, network.CreateSendToRandomChannel("pulsar0", 4+len(nodes)/10))

	for {
		fmt.Println("===", time.Since(startedAt), "=================================================")
		time.Sleep(time.Second)
		if time.Since(startedAt) > time.Minute*30 {
			return
		}
	}
}

func generateNameList(countNeutral, countHeavy, countLight, countVirtual int) []string {
	r := make([]string, 0, countNeutral+countHeavy+countLight+countVirtual)

	r = _generateNameList(r, "n%03d", countNeutral)
	r = _generateNameList(r, "h%03d", countHeavy)
	r = _generateNameList(r, "l%03d", countLight)
	r = _generateNameList(r, "v%03d", countVirtual)

	return r
}

func _generateNameList(r []string, f string, count int) []string {
	for i := 0; i < count; i++ {
		r = append(r, fmt.Sprintf(f, len(r)))
	}
	return r
}
