//
// Modified BSD 3-Clause Clear License
//
// Copyright (config) 2019 Insolar Technologies GmbH
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
//  * Neither the addr of Insolar Technologies GmbH nor the names of its contributors
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
//    (config) distribute this software (including without limitation in source code, binary or
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
	"crypto"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/insolar/insolar/certificate"
	"github.com/insolar/insolar/configuration"
	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/insolar/insolar/keystore"
	"github.com/insolar/insolar/network/consensusadapters"
	"github.com/insolar/insolar/network/node"
	"github.com/insolar/insolar/network/nodenetwork"
	transport2 "github.com/insolar/insolar/network/transport"
	"github.com/insolar/insolar/platformpolicy"
	"github.com/insolar/insolar/testutils"
)

func TestConsensusMain(t *testing.T) {
	startedAt := time.Now()

	ctx := initLogger()
	network := initNetwork(ctx)

	nodeIdents := generateNodeIdentities(0, 1, 3, 5)
	nodeInfos := generateNodeInfos(nodeIdents)
	nodes, discoveryNodes := nodesFromInfo(nodeInfos)

	for i, n := range nodes {
		nodeKeeper := nodenetwork.NewNodeKeeper(n)
		nodeKeeper.SetInitialSnapshot(nodes)
		certificateManager := initCrypto(n, discoveryNodes)
		handler := consensusadapters.NewDatagramHandler()
		transportFactory := transport2.NewFactory(configuration.NewHostNetwork().Transport)
		transport, _ := transportFactory.CreateDatagramTransport(handler)

		consensusAdapter := NewEmuHostConsensusAdapter(n.Address())

		bundle := consensusadapters.NewConsensusBundle(ctx, consensusadapters.ConsensusDep{
			PrimingCloudStateHash: [64]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0},
			Scheme:                platformpolicy.NewPlatformCryptographyScheme(),
			CertificateManager:    certificateManager,
			KeyStore:              keystore.NewInplaceKeyStore(nodeInfos[i].privateKey),
			NodeKeeper:            nodeKeeper,
			Stater:                &nshGen{nshDelay: defaultNshGenerationDelay},
			PulseChanger:          &pulseChanger{},
			PacketBuilder:         NewEmuPacketBuilder,
			DatagramTransport:     transport,
			// TODO: remove
			PacketSender: consensusAdapter,
		})

		controller := bundle.Controller()
		handler.SetConsensusController(controller)
		consensusAdapter.SetConsensusController(controller)

		_ = transport.Start(ctx)
		consensusAdapter.ConnectTo(network)
	}

	fmt.Println("===", len(nodes), "=================================================")

	network.Start(ctx)

	go CreateGenerator(2, 10, network.CreateSendToRandomChannel("pulsar0", 4+len(nodes)/10))

	for {
		fmt.Println("===", time.Since(startedAt), "=================================================")
		time.Sleep(time.Second)
		if time.Since(startedAt) > time.Minute*30 {
			return
		}
	}
}

func initLogger() context.Context {
	ctx := context.Background()
	logger := inslogger.FromContext(ctx).WithCaller(false)
	logger, _ = logger.WithLevelNumber(insolar.DebugLevel)
	logger, _ = logger.WithFormat(insolar.Text)
	ctx = inslogger.SetLogger(ctx, logger)
	return ctx
}

func initNetwork(ctx context.Context) *EmuNetwork {
	strategy := NewDelayNetStrategy(DelayStrategyConf{
		MinDelay:         100 * time.Millisecond,
		MaxDelay:         300 * time.Millisecond,
		Variance:         0.2,
		SpikeProbability: 0.1,
	})
	network := NewEmuNetwork(strategy, ctx)
	return network
}

func generateNodeIdentities(countNeutral, countHeavy, countLight, countVirtual int) []nodeIdentity {
	r := make([]nodeIdentity, 0, countNeutral+countHeavy+countLight+countVirtual)

	r = _generateNodeIdentity(r, countNeutral, insolar.StaticRoleUnknown)
	r = _generateNodeIdentity(r, countHeavy, insolar.StaticRoleHeavyMaterial)
	r = _generateNodeIdentity(r, countLight, insolar.StaticRoleLightMaterial)
	r = _generateNodeIdentity(r, countVirtual, insolar.StaticRoleVirtual)

	return r
}

const portOffset = 10000

func _generateNodeIdentity(r []nodeIdentity, count int, role insolar.StaticRole) []nodeIdentity {
	for i := 0; i < count; i++ {
		port := portOffset + len(r)
		r = append(r, nodeIdentity{
			role: role,
			addr: fmt.Sprintf("127.0.0.1:%d", port),
		})
	}
	return r
}

func generateNodeInfos(nodeIdents []nodeIdentity) []*nodeInfo {
	keyProcessor := platformpolicy.NewKeyProcessor()

	nodeInfos := make([]*nodeInfo, 0, len(nodeIdents))
	for _, ni := range nodeIdents {
		privateKey, _ := keyProcessor.GeneratePrivateKey()
		publicKey := keyProcessor.ExtractPublicKey(privateKey)

		nodeInfos = append(nodeInfos, &nodeInfo{
			nodeIdentity: ni,
			publicKey:    publicKey,
			privateKey:   privateKey,
		})
	}
	return nodeInfos
}

type nodeIdentity struct {
	role insolar.StaticRole
	addr string
}

type nodeInfo struct {
	nodeIdentity
	privateKey crypto.PrivateKey
	publicKey  crypto.PublicKey
}

func nodesFromInfo(nodeInfos []*nodeInfo) ([]insolar.NetworkNode, []insolar.NetworkNode) {
	nodes := make([]insolar.NetworkNode, len(nodeInfos))
	discoveryNodes := make([]insolar.NetworkNode, 0)

	for i, info := range nodeInfos {
		var isDiscovery bool
		if info.role == insolar.StaticRoleHeavyMaterial || info.role == insolar.StaticRoleUnknown {
			isDiscovery = true
		}

		nn := newNetworkNode(i, info.addr, info.role, info.publicKey)
		nodes[i] = nn
		if isDiscovery {
			discoveryNodes = append(discoveryNodes, nn)
		}
	}

	return nodes, discoveryNodes
}

const shortNodeIdOffset = 1000

func newNetworkNode(id int, addr string, role insolar.StaticRole, pk crypto.PublicKey) node.MutableNode {
	n := node.NewNode(
		testutils.RandomRef(),
		role,
		pk,
		addr,
		"",
	)
	mn := n.(node.MutableNode)
	mn.SetShortID(insolar.ShortNodeID(shortNodeIdOffset + id))
	return mn
}

func initCrypto(node insolar.NetworkNode, discoveryNodes []insolar.NetworkNode) *certificate.CertificateManager {
	pubKey := node.PublicKey()

	proc := platformpolicy.NewKeyProcessor()
	publicKey, _ := proc.ExportPublicKeyPEM(pubKey)

	bootstrapNodes := make([]certificate.BootstrapNode, 0, len(discoveryNodes))
	for _, dn := range discoveryNodes {
		pubKey := dn.PublicKey()
		pubKeyBuf, _ := proc.ExportPublicKeyPEM(pubKey)

		bootstrapNode := certificate.NewBootstrapNode(
			pubKey,
			string(pubKeyBuf[:]),
			dn.Address(),
			dn.ID().String(),
		)
		bootstrapNodes = append(bootstrapNodes, *bootstrapNode)
	}

	cert := &certificate.Certificate{
		AuthorizationCertificate: certificate.AuthorizationCertificate{
			PublicKey: string(publicKey[:]),
			Reference: node.ID().String(),
			Role:      node.Role().String(),
		},
		BootstrapNodes: bootstrapNodes,
	}

	// dump cert and read it again from json for correct private files initialization
	jsonCert, _ := cert.Dump()
	cert, _ = certificate.ReadCertificateFromReader(pubKey, proc, strings.NewReader(jsonCert))
	return certificate.NewCertificateManager(cert)
}

const defaultNshGenerationDelay = time.Millisecond * 0

type nshGen struct {
	nshDelay time.Duration
}

func (ng *nshGen) State() ([]byte, error) {
	delay := ng.nshDelay
	if delay != 0 {
		time.Sleep(delay)
	}

	nshBytes := make([]byte, 64)
	rand.Read(nshBytes)

	return nshBytes, nil
}

type pulseChanger struct{}

func (pc *pulseChanger) ChangePulse(ctx context.Context, pulse insolar.Pulse) {
	inslogger.FromContext(ctx).Info(">>>>>> Change pulse called")
}