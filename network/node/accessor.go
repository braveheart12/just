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

package node

import (
	"math/rand"
	"sort"

	"github.com/insolar/insolar/insolar"
)

type Accessor struct {
	snapshot  *Snapshot
	refIndex  map[insolar.Reference]insolar.NetworkNode
	sidIndex  map[insolar.ShortNodeID]insolar.NetworkNode
	addrIndex map[string]insolar.NetworkNode
	roleIndex map[insolar.StaticRole]*refSet
	// should be removed in future
	active []insolar.NetworkNode
}

func (a *Accessor) GetActiveNodeByShortID(shortID insolar.ShortNodeID) insolar.NetworkNode {
	return a.sidIndex[shortID]
}

func (a *Accessor) GetActiveNodeByAddr(address string) insolar.NetworkNode {
	return a.addrIndex[address]
}

func (a *Accessor) GetActiveNodes() []insolar.NetworkNode {
	result := make([]insolar.NetworkNode, len(a.active))
	copy(result, a.active)
	return result
}

func (a *Accessor) GetActiveNode(ref insolar.Reference) insolar.NetworkNode {
	return a.refIndex[ref]
}

func (a *Accessor) GetWorkingNode(ref insolar.Reference) insolar.NetworkNode {
	node := a.GetActiveNode(ref)
	if node == nil || node.GetState() != insolar.NodeReady {
		return nil
	}
	return node
}

func (a *Accessor) GetWorkingNodes() []insolar.NetworkNode {
	workingList := a.snapshot.nodeList[ListWorking]
	result := make([]insolar.NetworkNode, len(workingList))
	copy(result, workingList)
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID().Compare(result[j].ID()) < 0
	})
	return result
}

func (a *Accessor) GetRandomWorkingNode() insolar.NetworkNode {
	workingList := a.snapshot.nodeList[ListWorking]
	if l := len(workingList); l == 0 {
		return nil
	} else {
		return workingList[rand.Intn(l)]
	}
}

func (a *Accessor) GetWorkingNodesByRole(role insolar.DynamicRole) []insolar.Reference {
	staticRole := dynamicToStaticRole(role)
	nodes := a.roleIndex[staticRole]
	if nodes == nil {
		return []insolar.Reference{}
	}
	return nodes.Collect()
}

func GetSnapshotActiveNodes(snapshot *Snapshot) []insolar.NetworkNode {
	joining := snapshot.nodeList[ListJoiner]
	working := snapshot.nodeList[ListWorking]
	leaving := snapshot.nodeList[ListLeaving]

	result := make([]insolar.NetworkNode, len(joining)+len(working)+len(leaving))
	copy(result[:len(joining)], joining)
	copy(result[len(joining):len(joining)+len(working)], working)
	copy(result[len(joining)+len(working):], leaving)

	return result
}

func (a *Accessor) addToIndex(node insolar.NetworkNode) {
	a.refIndex[node.ID()] = node
	a.sidIndex[node.ShortID()] = node
	a.addrIndex[node.Address()] = node

	if node.GetState() != insolar.NodeReady {
		return
	}

	list, ok := a.roleIndex[node.Role()]
	if !ok {
		list = newRefSet()
	}

	list.Add(node.ID())
	a.roleIndex[node.Role()] = list
}

func NewAccessor(snapshot *Snapshot) *Accessor {
	result := &Accessor{
		snapshot:  snapshot,
		refIndex:  make(map[insolar.Reference]insolar.NetworkNode),
		sidIndex:  make(map[insolar.ShortNodeID]insolar.NetworkNode),
		roleIndex: make(map[insolar.StaticRole]*refSet),
		addrIndex: make(map[string]insolar.NetworkNode),
	}
	result.active = GetSnapshotActiveNodes(snapshot)
	for _, node := range result.active {
		result.addToIndex(node)
	}
	return result
}

func dynamicToStaticRole(role insolar.DynamicRole) insolar.StaticRole {
	switch role {
	case insolar.DynamicRoleVirtualExecutor:
		return insolar.StaticRoleVirtual
	case insolar.DynamicRoleVirtualValidator:
		return insolar.StaticRoleVirtual
	case insolar.DynamicRoleLightExecutor:
		return insolar.StaticRoleLightMaterial
	case insolar.DynamicRoleLightValidator:
		return insolar.StaticRoleLightMaterial
	case insolar.DynamicRoleHeavyExecutor:
		return insolar.StaticRoleHeavyMaterial
	default:
		return insolar.StaticRoleUnknown
	}
}
