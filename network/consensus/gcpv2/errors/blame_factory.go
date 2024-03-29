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

package errors

import (
	"fmt"

	"github.com/insolar/insolar/network/consensus/common"
	common2 "github.com/insolar/insolar/network/consensus/gcpv2/common"
)

var _ MisbehaviorReport = &BlameError{}

type BlameError struct {
	blameType    int
	msg          string
	violatorHost common.HostIdentity
	violatorNode common2.NodeProfile
	details      []interface{}
	captureMark  interface{}
}

func (p *BlameError) IsUnknown() bool {
	return p.blameType != 0
}

func (p *BlameError) MisbehaviorType() MisbehaviorType {
	return Blame.Of(p.blameType)
}

func (p *BlameError) CaptureMark() interface{} {
	return p.captureMark
}

func (p *BlameError) Details() []interface{} {
	return p.details
}

func (p *BlameError) ViolatorNode() common2.NodeProfile {
	return p.violatorNode
}

func (p *BlameError) ViolatorHost() common.HostIdentity {
	return p.violatorHost
}

func (p BlameError) BlameType() int {
	return p.blameType
}

func (p BlameError) Error() string {
	var v interface{} = p.violatorNode
	if v == nil {
		v = p.violatorHost
	}
	c := p.captureMark
	if c == nil {
		c = ""
	}
	return fmt.Sprintf("blame: type=%v(%v)%v, violator=%v, details=%+v", p.blameType, p.msg, c, v, p.details)
}

const (
	_ = iota
	BlameExcessiveIntro
	MismatchedPulsarPacket
)

func NewBlameFactory(capture MisbehaviorReportFunc) BlameFactory {
	return BlameFactory{capture: capture}
}

type BlameFactory struct {
	capture MisbehaviorReportFunc
}

func (p BlameFactory) NewBlame(fraudType int, msg string, violatorHost common.HostIdentityHolder, violatorNode common2.NodeProfile, details ...interface{}) BlameError {
	err := BlameError{
		blameType:    fraudType,
		msg:          msg,
		violatorNode: violatorNode,
		details:      details}
	if violatorHost != nil {
		err.violatorHost = common.NewHostIdentityFromHolder(violatorHost)
	}
	if p.capture != nil {
		err.captureMark = p.capture(&err)
	}
	return err
}

func (p BlameFactory) NewNodeBlame(fraudType int, msg string, violatorNode common2.NodeProfile, details ...interface{}) BlameError {
	return p.NewBlame(fraudType, msg, nil, violatorNode, details...)
}

func (p BlameFactory) NewHostBlame(fraudType int, msg string, violatorHost common.HostIdentityHolder, details ...interface{}) BlameError {
	return p.NewBlame(fraudType, msg, violatorHost, nil, details...)
}

func (p BlameFactory) ExcessiveIntro(violator common2.NodeProfile, evidence1 common.SignedEvidenceHolder, evidence2 common.SignedEvidenceHolder) BlameError {
	return p.NewNodeBlame(BlameExcessiveIntro, "excessive intro", violator, evidence1, evidence2)
}

func (p BlameFactory) NewMismatchedPulsarPacket(from common.HostIdentityHolder, expected common2.OriginalPulsarPacket, received common2.OriginalPulsarPacket) error {
	return p.NewHostBlame(MismatchedPulsarPacket, "mixed pulsar pulses", from, expected, received)
}

func (p BlameFactory) NewMismatchedPulsePacket(from common2.NodeProfile, expected common2.OriginalPulsarPacket, received common2.OriginalPulsarPacket) error {
	return p.NewNodeBlame(MismatchedPulsarPacket, "mixed pulsar pulses", from, expected, received)
}
