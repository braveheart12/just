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

package tests

import (
	"context"
	"errors"
	"runtime/debug"

	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/insolar/insolar/network/consensus/adapters"
	common2 "github.com/insolar/insolar/network/consensus/gcpv2/common"
	"github.com/insolar/insolar/network/consensus/gcpv2/core"
	"github.com/insolar/insolar/network/consensus/gcpv2/packets"

	"github.com/insolar/insolar/network/consensus/common"
)

type EmuHostConsensusAdapter struct {
	packetProcessor adapters.PacketProcessor

	hostAddr common.HostAddress
	inbound  <-chan Packet
	outbound chan<- Packet
}

func NewEmuHostConsensusAdapter(hostAddr string) *EmuHostConsensusAdapter {
	return &EmuHostConsensusAdapter{hostAddr: common.HostAddress(hostAddr)}
}

func (h *EmuHostConsensusAdapter) SetPacketProcessor(packetProcessor adapters.PacketProcessor) {
	h.packetProcessor = packetProcessor
}

func (h *EmuHostConsensusAdapter) ConnectTo(network *EmuNetwork) {
	ctx := network.ctx
	h.inbound, h.outbound = network.AddHost(ctx, h.hostAddr)
	go h.run(ctx)
}

func (h *EmuHostConsensusAdapter) run(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
			inslogger.FromContext(ctx).Errorf("host has died: %v, %v", h.hostAddr, r)
			close(h.outbound)
		}
	}()

	for {
		var err error
		payload, from, err := h.receive(ctx)
		if err == nil {
			var packet packets.PacketParser

			packet, err = h.parsePayload(payload)
			if err == nil {
				if packet != nil {
					hostFrom := common.HostIdentity{Addr: *from}
					err = h.packetProcessor.ProcessPacket(ctx, packet, &hostFrom)
				}
			}
		}

		if err != nil {
			inslogger.FromContext(ctx).Error(err)
		}
	}
}

func (h *EmuHostConsensusAdapter) SendPacketToTransport(ctx context.Context, t common2.NodeProfile, sendOptions core.PacketSendOptions, payload interface{}) {
	h.send(t.GetDefaultEndpoint().GetNameAddress(), payload)
}

func (h *EmuHostConsensusAdapter) receive(ctx context.Context) (payload interface{}, from *common.HostAddress, err error) {
	packet, ok := <-h.inbound
	if !ok {
		panic(errors.New("connection closed"))
	}
	inslogger.FromContext(ctx).Infof("receivedBy: %s - %+v", h.hostAddr, packet)
	if packet.Payload == nil {
		return nil, &packet.Host, errors.New("missing payload")
	}
	err, ok = packet.Payload.(error)
	if ok {
		return nil, &packet.Host, err
	}
	return packet.Payload, &packet.Host, nil
}

func (h *EmuHostConsensusAdapter) send(target common.HostAddress, payload interface{}) {
	parser := payload.(packets.PacketParser)
	pkt := Packet{Host: target, Payload: WrapPacketParser(parser)}
	h.outbound <- pkt
}

func (h *EmuHostConsensusAdapter) parsePayload(payload interface{}) (packets.PacketParser, error) {
	return UnwrapPacketParser(payload), nil
}
