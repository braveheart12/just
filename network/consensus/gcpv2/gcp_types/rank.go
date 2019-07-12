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

package gcp_types

import (
	"fmt"
)

/*
	Power      common2.MemberPower // serialized to [00-07]
	Index      uint16              // serialized to [08-17]
	TotalCount uint16              // serialized to [18-27]
	Condition  MemberCondition     //serialized to [28-31]
*/type MembershipRank uint32

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
	r |= mode.AsUnit32() << 28
	return MembershipRank(r)
}

func ensureNodeIndex(v uint16) uint32 {
	if v > 0x03FF {
		panic("out of bounds")
	}
	return uint32(v & 0x03FF)
}