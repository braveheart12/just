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

package common

import (
	"fmt"

	"github.com/insolar/insolar/network/consensus/common"
)

/*
	Power      common2.MemberPower // serialized to [00-07]
	Index      uint16              // serialized to [08-17]
	TotalCount uint16              // serialized to [18-27]
	Condition  MemberCondition     //serialized to [28-31]
*/
type MembershipRank uint32

const JoinerMembershipRank MembershipRank = 0

func (v MembershipRank) GetPower() MemberPower {
	return MemberPower(v)
}

func (v MembershipRank) GetIndex() uint16 {
	return uint16(v>>8) & 0x03FF
}

func (v MembershipRank) GetTotalCount() uint16 {
	return uint16(v>>18) & 0x03FF
}

func (v MembershipRank) GetMode() MemberOpMode {
	return MemberOpMode(v >> 28)
}

func (v MembershipRank) IsJoiner() bool {
	return v == JoinerMembershipRank
}

func (v MembershipRank) String() string {
	if v.IsJoiner() {
		return "{joiner}"
	}
	return fmt.Sprintf("{%v %d/%d pw:%v}", v.GetMode(), v.GetIndex(), v.GetTotalCount(), v.GetPower())
}

func NewMembershipRank(mode MemberOpMode, pw MemberPower, idx, count uint16) MembershipRank {
	if idx >= count {
		panic("illegal value")
	}

	r := uint32(pw)
	r |= ensureNodeIndex(idx) << 8
	r |= ensureNodeIndex(count) << 18
	r |= mode.asUnit32() << 28
	return MembershipRank(r)
}

func ensureNodeIndex(v uint16) uint32 {
	if v > 0x03FF {
		panic("out of bounds")
	}
	return uint32(v & 0x03FF)
}

type MemberOpMode uint8 //4-bit value
const (
	MemberModeBits                                = 4
	MemberModeFlagRestrictedBehavior MemberOpMode = 4
	MemberModeFlagValidationWarning  MemberOpMode = 2
	MemberModeFlagSuspendedOps       MemberOpMode = 1

	MemberModeNormal                    = 0
	MemberModeSuspected                 = /* 0x01 */ MemberModeFlagSuspendedOps
	MemberModePossibleFraud             = /* 0x02 */ MemberModeFlagValidationWarning
	MemberModePossibleFraudAndSuspected = /* 0x03 */ MemberModeFlagSuspendedOps | MemberModeFlagValidationWarning
	MemberModeRestrictedAnnouncement    = /* 0x04 */ MemberModeFlagRestrictedBehavior
	MemberModeEvictedGracefully         = /* 0x05 */ MemberModeFlagRestrictedBehavior | MemberModeFlagSuspendedOps
	MemberModeEvictedAsFraud            = /* 0x06 */ MemberModeFlagRestrictedBehavior | MemberModeFlagValidationWarning
	MemberModeEvictedAsSuspected        = /* 0x07 */ MemberModeFlagRestrictedBehavior | MemberModeFlagValidationWarning | MemberModeFlagSuspendedOps
)

func (v MemberOpMode) IsEvicted() bool {
	return v >= MemberModeEvictedGracefully
}

func (v MemberOpMode) IsRestricted() bool {
	return v&MemberModeFlagRestrictedBehavior != 0
}

func (v MemberOpMode) IsMistrustful() bool {
	return v&MemberModeFlagValidationWarning != 0
}

func (v MemberOpMode) IsSuspended() bool {
	return v&MemberModeFlagSuspendedOps != 0
}

/* Is allowed to take some work, but needs power >0 to be a working node (to be assigned for some work) */
func (v MemberOpMode) IsPowerful() bool {
	return !(v.IsSuspended() || v.IsEvicted())
}

func (v MemberOpMode) asUnit32() uint32 {
	if v >= 1<<MemberModeBits {
		panic("illegal value")
	}
	return uint32(v)
}

func (v MemberOpMode) String() string {
	switch v {
	case MemberModeNormal:
		return "mode:norm"
	case MemberModeSuspected:
		return "mode:susp"
	case MemberModePossibleFraud:
		return "mode:warn"
	case MemberModePossibleFraudAndSuspected:
		return "mode:warn+susp"
	case MemberModeRestrictedAnnouncement:
		return "mode:joiner"
	case MemberModeEvictedGracefully:
		return "evict:norm"
	case MemberModeEvictedAsFraud:
		return "evict:fraud"
	case MemberModeEvictedAsSuspected:
		return "evict:susp"
	default:
		return fmt.Sprintf("?%d?", v)
	}
}

type GlobulaNodeState struct {
	NodeStateHash      common.Bits512
	PulseDataHash      common.Bits256
	LastCloudStateHash common.Bits224 // CSH is 512 and is folded down then high 32 bits are discarded
	NodeRank           MembershipRank
}

type SignedGlobulaNodeState struct {
	GlobulaNodeState GlobulaNodeState
	Signature        common.Bits512
}

//go:generate minimock -i github.com/insolar/insolar/network/consensus/gcpv2/common.NodeStateHashEvidence -o . -s _mock.go

type NodeStateHashEvidence interface {
	GetNodeStateHash() NodeStateHash
	GetGlobulaNodeStateSignature() common.SignatureHolder
}
