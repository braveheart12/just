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

package claimhandler

import (
	"testing"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/network/consensusv1/packets"
	"github.com/insolar/insolar/testutils"
	"github.com/stretchr/testify/assert"
)

func TestApprovedJoinersCount(t *testing.T) {
	assert.Equal(t, 1, ApprovedJoinersCount(1, 1))
	assert.Equal(t, 1, ApprovedJoinersCount(2, 3))
	assert.Equal(t, 3, ApprovedJoinersCount(5, 10))
	assert.Equal(t, 2, ApprovedJoinersCount(2, 10))
}

func TestClaimHandler_FilterClaims(t *testing.T) {
	// announcers references do not affect joiner claims filter logic, so choose random
	ref1 := testutils.RandomRef()
	ref2 := testutils.RandomRef()
	ref3 := testutils.RandomRef()

	claims := make(map[insolar.Reference][]packets.ReferendumClaim)
	claims[ref1] = []packets.ReferendumClaim{getJoinClaim(t, insolar.Reference{152})}
	claims[ref2] = []packets.ReferendumClaim{getJoinClaim(t, insolar.Reference{0}), getJoinClaim(t, insolar.Reference{154})}
	claims[ref3] = []packets.ReferendumClaim{getJoinClaim(t, insolar.Reference{1}), getJoinClaim(t, insolar.Reference{153})}

	containsJoinClaim := func(claims []packets.ReferendumClaim, ref insolar.Reference) bool {
		for _, claim := range claims {
			joinClaim, ok := claim.(*packets.NodeJoinClaim)
			if !ok {
				continue
			}
			if joinClaim.NodeRef.Equal(ref) {
				return true
			}
		}
		return false
	}

	handler := NewClaimHandler(6, claims)
	result := handler.FilterClaims([]insolar.Reference{ref1, ref2, ref3}, insolar.Entropy{0})
	// 2 JoinClaims
	assert.Equal(t, 2, len(result.ApprovedClaims))
	assert.True(t, containsJoinClaim(result.ApprovedClaims, insolar.Reference{154}))
	assert.True(t, containsJoinClaim(result.ApprovedClaims, insolar.Reference{153}))
	assert.False(t, containsJoinClaim(result.ApprovedClaims, insolar.Reference{0}))
	assert.False(t, containsJoinClaim(result.ApprovedClaims, insolar.Reference{1}))

	// 2 JoinClaims
	result = handler.FilterClaims([]insolar.Reference{ref1, ref2}, insolar.Entropy{0})
	assert.Equal(t, 2, len(result.ApprovedClaims))
	assert.True(t, containsJoinClaim(result.ApprovedClaims, insolar.Reference{154}))
	assert.True(t, containsJoinClaim(result.ApprovedClaims, insolar.Reference{152}))
	assert.False(t, containsJoinClaim(result.ApprovedClaims, insolar.Reference{0}))
	assert.False(t, containsJoinClaim(result.ApprovedClaims, insolar.Reference{1}))

	// only 2 JoinClaims
	result = handler.FilterClaims([]insolar.Reference{ref2, ref3}, insolar.Entropy{0})
	assert.Equal(t, 2, len(result.ApprovedClaims))
	assert.True(t, containsJoinClaim(result.ApprovedClaims, insolar.Reference{154}))
	assert.True(t, containsJoinClaim(result.ApprovedClaims, insolar.Reference{153}))
	assert.False(t, containsJoinClaim(result.ApprovedClaims, insolar.Reference{0}))
	assert.False(t, containsJoinClaim(result.ApprovedClaims, insolar.Reference{1}))
}

func TestClaimHandler_GetClaims(t *testing.T) {
	// announcers references do not affect joiner claims filter logic, so choose random
	ref1 := testutils.RandomRef()
	ref2 := testutils.RandomRef()
	ref3 := testutils.RandomRef()

	claims := make(map[insolar.Reference][]packets.ReferendumClaim)
	claims[ref1] = []packets.ReferendumClaim{getJoinClaim(t, insolar.Reference{0})}
	claims[ref2] = []packets.ReferendumClaim{getJoinClaim(t, insolar.Reference{1}), getJoinClaim(t, insolar.Reference{2})}

	handler := NewClaimHandler(6, claims)
	assert.Equal(t, 3, len(handler.GetClaims()))

	handler.SetClaimsFromNode(ref3, []packets.ReferendumClaim{})
	assert.Equal(t, 3, len(handler.GetClaims()))

	handler.SetClaimsFromNode(ref3, []packets.ReferendumClaim{getJoinClaim(t, insolar.Reference{3})})
	assert.Equal(t, 4, len(handler.GetClaims()))
}
