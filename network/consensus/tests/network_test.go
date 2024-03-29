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
	"fmt"
	"math/rand"
	"sync"

	"github.com/insolar/insolar/network/consensus/common"
)

type NetStrategy interface {
	GetLinkStrategy(hostAddress common.HostAddress) LinkStrategy
}

type PacketFunc func(packet *Packet)

type LinkStrategy interface {
	BeforeSend(packet *Packet, out PacketFunc)
	BeforeReceive(packet *Packet, out PacketFunc)
}

type Packet struct {
	Payload interface{}
	Host    common.HostAddress
}

type EmuRoute struct {
	host     common.HostAddress
	network  *EmuNetwork
	strategy LinkStrategy
	toHost   chan<- Packet
	fromHost <-chan Packet
}

type EmuNetwork struct {
	hostsSync sync.RWMutex
	ctx       context.Context
	hosts     map[common.HostAddress]*EmuRoute
	strategy  NetStrategy
	running   bool
	bufSize   int
}

type errEmuNetwork struct {
	errType string
	details interface{}
}

func (e errEmuNetwork) Error() string {
	return fmt.Sprintf("emu-net error - %s: %v", e.errType, e.details)
}

func ErrUnknownEmuHost(host common.HostAddress) error {
	return errEmuNetwork{errType: "Unknown host", details: host}
}

func IsEmuError(err *error) bool {
	_, ok := (*err).(errEmuNetwork)
	return ok
}

func NewEmuNetwork(nwStrategy NetStrategy, ctx context.Context) *EmuNetwork {
	return &EmuNetwork{strategy: nwStrategy, ctx: ctx}
}

func (emuNet *EmuNetwork) AddHost(ctx context.Context, host common.HostAddress) (toHost <-chan Packet, fromHost chan<- Packet) {
	emuNet.hostsSync.Lock()
	defer emuNet.hostsSync.Unlock()

	_, isPresent := emuNet.hosts[host]
	if isPresent {
		panic(fmt.Sprintf("Duplicate host: %v", host))
	}

	var routeStrategy LinkStrategy
	if emuNet.strategy != nil {
		routeStrategy = emuNet.strategy.GetLinkStrategy(host)
	}
	if routeStrategy == nil {
		routeStrategy = stubLinkStrategyValue
	}

	chanBufSize := emuNet.bufSize
	if chanBufSize <= 0 {
		chanBufSize = 10
	}

	fromHostC := make(chan Packet, chanBufSize)
	toHostC := make(chan Packet, chanBufSize)

	if emuNet.hosts == nil {
		emuNet.hosts = make(map[common.HostAddress]*EmuRoute)
	}

	route := EmuRoute{host: host, strategy: routeStrategy, toHost: toHostC, fromHost: fromHostC, network: emuNet}
	emuNet.hosts[host] = &route

	if emuNet.running {
		go route.run(ctx)
	}

	return toHostC, fromHostC
}

func (emuNet *EmuNetwork) DropHost(host common.HostAddress) bool {
	route := emuNet.getHostRoute(host)
	if route == nil {
		return false
	}

	route.closeRoute()
	return true
}

func (emuNet *EmuNetwork) SendToHost(host common.HostAddress, payload interface{}, fromHost common.HostAddress) bool {
	route := emuNet.getHostRoute(host)
	if route == nil {
		return false
	}

	targetPacket := Packet{Payload: payload, Host: fromHost}
	route.pushPacket(targetPacket)
	return true
}

func (emuNet *EmuNetwork) SendToAll(payload interface{}, fromHost common.HostAddress) {
	for _, route := range emuNet.getRoutes() {
		targetPacket := Packet{Payload: payload, Host: fromHost}
		route.pushPacket(targetPacket)
	}
}

func (emuNet *EmuNetwork) SendRandom(payload interface{}, fromHost common.HostAddress) {
	targetPacket := Packet{Payload: payload, Host: fromHost}
	routes := emuNet.getRoutes()
	routes[rand.Intn(len(routes))].pushPacket(targetPacket)
}

func (emuNet *EmuNetwork) CreateSendToAllChannel() chan<- Packet {
	inbound := make(chan Packet)
	go func() {
		for {
			inboundPacket, ok := <-inbound
			if !ok {
				return
			}
			emuNet.SendToAll(inboundPacket.Payload, inboundPacket.Host)
		}
	}()
	return inbound
}

func (emuNet *EmuNetwork) CreateSendToAllFromOneChannel(sender common.HostAddress) chan<- interface{} {
	inbound := make(chan interface{})
	go func() {
		for {
			payload, ok := <-inbound
			if !ok {
				return
			}
			emuNet.SendToAll(payload, sender)
		}
	}()
	return inbound
}

func (emuNet *EmuNetwork) CreateSendToRandomChannel(sender common.HostAddress, attempts int) chan<- interface{} {
	inbound := make(chan interface{})
	go func() {
		for {
			payload, ok := <-inbound
			if !ok {
				return
			}
			for i := 0; i < attempts; i++ {
				emuNet.SendRandom(payload, sender)
			}
		}
	}()
	return inbound
}

func (emuNet *EmuNetwork) GetHosts() []*common.HostAddress {
	emuNet.hostsSync.RLock()
	defer emuNet.hostsSync.RUnlock()

	keys := make([]*common.HostAddress, 0, len(emuNet.hosts))
	for k := range emuNet.hosts {
		keys = append(keys, &k)
	}

	return keys
}

func (emuNet *EmuNetwork) getRoutes() []*EmuRoute {
	emuNet.hostsSync.RLock()
	defer emuNet.hostsSync.RUnlock()

	routes := make([]*EmuRoute, 0, len(emuNet.hosts))
	for _, v := range emuNet.hosts {
		routes = append(routes, v)
	}

	return routes
}

func (emuNet *EmuNetwork) Start(ctx context.Context) {
	emuNet.hostsSync.Lock()
	defer emuNet.hostsSync.Unlock()

	if emuNet.running {
		return
	}
	emuNet.running = true

	for _, route := range emuNet.hosts {
		go route.run(ctx)
	}
}

func (emuNet *EmuNetwork) getHostRoute(host common.HostAddress) *EmuRoute {
	emuNet.hostsSync.RLock()
	defer emuNet.hostsSync.RUnlock()

	return emuNet.hosts[host]
}

func (emuNet *EmuNetwork) internalRemoveHost(route *EmuRoute) {
	emuNet.hostsSync.Lock()
	defer emuNet.hostsSync.Unlock()
	delete(emuNet.hosts, route.host)
}

func (emuRt *EmuRoute) run(ctx context.Context) {
	defer emuRt.closeRoute()

	for {
		select {
		case <-ctx.Done():
			return
		case originPacket, ok := <-emuRt.fromHost:
			if !ok {
				return
			}
			// strategy can modify target and payload of a packet before delivery
			emuRt.strategy.BeforeSend(&originPacket, emuRt._sendPacket)
		}
	}
}

func (emuRt *EmuRoute) pushPacket(packet Packet) {
	emuRt.strategy.BeforeReceive(&packet, emuRt._recvPacket)
}

func (emuRt *EmuRoute) _sendPacket(originPacket *Packet) {

	targetPacket := Packet{Payload: originPacket.Payload, Host: emuRt.host}

	var outRoute *EmuRoute
	if originPacket.Host.IsLocalHost() || originPacket.Host == emuRt.host {
		outRoute = emuRt
	} else {
		outRoute = emuRt.network.getHostRoute(originPacket.Host)
		if outRoute == nil {
			targetPacket.Payload = ErrUnknownEmuHost(originPacket.Host)
			// inbound strategy MUST NOT be applied to error replies
			emuRt.toHost <- targetPacket
			return
		}
	}

	outRoute.pushPacket(targetPacket)
}

func (emuRt *EmuRoute) _recvPacket(originPacket *Packet) {
	defer func() {
		if recover() == nil {
			return
		}
		outRoute := emuRt.network.getHostRoute(originPacket.Host)
		if outRoute == nil {
			// the sender receiver is not available anymore
			return
		}
		targetPacket := Packet{Payload: ErrUnknownEmuHost(emuRt.host), Host: outRoute.host}
		outRoute.toHost <- targetPacket
	}()
	emuRt.toHost <- *originPacket
}

func (emuRt *EmuRoute) closeRoute() {
	defer func() {
		recover()
		if emuRt.network != nil {
			emuRt.network.internalRemoveHost(emuRt)
		}
	}()
	close(emuRt.toHost)
}

var stubLinkStrategyValue LinkStrategy = &stubLinkStrategy{}

type stubLinkStrategy struct{}

func (stubLinkStrategy) BeforeSend(packet *Packet, out PacketFunc) {
	out(packet)
}

func (stubLinkStrategy) BeforeReceive(packet *Packet, out PacketFunc) {
	out(packet)
}
