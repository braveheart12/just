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

package serialization

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"testing"

	"github.com/insolar/insolar/network/consensus/adapters"
	"github.com/insolar/insolar/network/consensus/common"
	"github.com/insolar/insolar/network/consensus/gcpv2/packets"
	"github.com/insolar/insolar/platformpolicy"
	"github.com/stretchr/testify/require"
)

var digester = func() common.DataDigester {
	scheme := platformpolicy.NewPlatformCryptographyScheme()
	digester := adapters.NewSha3512Digester(scheme)
	return digester
}()

var signer = func() common.DigestSigner {
	processor := platformpolicy.NewKeyProcessor()
	key, _ := processor.GeneratePrivateKey()
	scheme := platformpolicy.NewPlatformCryptographyScheme()
	signer := adapters.NewECDSADigestSigner(key.(*ecdsa.PrivateKey), scheme)
	return signer
}()

func TestHeader_IsRelayRestricted(t *testing.T) {
	h := Header{}

	require.False(t, h.IsRelayRestricted())

	h.PacketFlags = 1 // 0b00000001
	require.True(t, h.IsRelayRestricted())
}

func TestHeader_setIsRelayRestricted(t *testing.T) {
	h := Header{}

	require.False(t, h.IsRelayRestricted())

	h.setIsRelayRestricted(true)
	require.True(t, h.IsRelayRestricted())

	h.setIsRelayRestricted(false)
	require.False(t, h.IsRelayRestricted())
}

func TestHeader_IsBodyEncrypted(t *testing.T) {
	h := Header{
		PacketFlags: 0,
	}

	require.False(t, h.IsBodyEncrypted())

	h.PacketFlags = 2 // 0b00000010
	require.True(t, h.IsBodyEncrypted())

	h.PacketFlags = 3 // 0b00000011
	require.True(t, h.IsBodyEncrypted())
}

func TestHeader_setIsBodyEncrypted(t *testing.T) {
	h := Header{}

	require.False(t, h.IsBodyEncrypted())

	h.setIsBodyEncrypted(true)
	require.True(t, h.IsBodyEncrypted())

	h.setIsBodyEncrypted(false)
	require.False(t, h.IsBodyEncrypted())
}

func TestHeader_HasFlag(t *testing.T) {
	h := Header{}

	require.False(t, h.HasFlag(0))

	h.PacketFlags = 4 // 0b00000100

	require.True(t, h.HasFlag(0))
}

func TestHeader_HasFlag_Panics(t *testing.T) {
	h := Header{}

	require.Panics(t, func() { h.HasFlag(maxFlagIndex + 1) })
}

func TestHeader_SetFlag(t *testing.T) {
	h := Header{}

	require.False(t, h.HasFlag(0))

	h.SetFlag(0)
	require.True(t, h.HasFlag(0))
}

func TestHeader_SetFlag_Panics(t *testing.T) {
	h := Header{}

	require.Panics(t, func() { h.SetFlag(maxFlagIndex + 1) })
}

func TestHeader_ClearFlag(t *testing.T) {
	h := Header{}

	require.False(t, h.HasFlag(0))

	h.SetFlag(0)
	require.True(t, h.HasFlag(0))

	h.ClearFlag(0)
	require.False(t, h.HasFlag(0))
}

func TestHeader_ClearFlag_Panics(t *testing.T) {
	h := Header{}

	require.Panics(t, func() { h.ClearFlag(maxFlagIndex + 1) })
}

func TestHeader_GetFlagRangeInt(t *testing.T) {
	h := Header{}

	require.Panics(t, func() { h.GetFlagRangeInt(2, 1) })
	require.Panics(t, func() { h.GetFlagRangeInt(1, maxFlagIndex+1) })
}

func TestHeader_GetFlagRangeInt_Panic(t *testing.T) {
	h := Header{}

	require.EqualValues(t, 0, h.GetFlagRangeInt(0, 2))

	h.PacketFlags = 20 // 0b00010100

	require.EqualValues(t, 5, h.GetFlagRangeInt(0, 2))
	require.EqualValues(t, 1, h.GetFlagRangeInt(0, 0))
	require.EqualValues(t, 1, h.GetFlagRangeInt(0, 1))
}

func TestHeader_GetSourceID(t *testing.T) {
	h := Header{
		SourceID: 123,
	}

	require.Equal(t, common.ShortNodeID(123), h.GetSourceID())
}

func TestHeader_GetProtocolType(t *testing.T) {
	h := Header{}

	require.Equal(t, ProtocolTypePulsar, h.GetProtocolType())

	h.ProtocolAndPacketType = 16 // 0b00010000
	require.Equal(t, ProtocolTypeGlobulaConsensus, h.GetProtocolType())
}

func TestHeader_setProtocolType(t *testing.T) {
	h := Header{}

	require.Equal(t, ProtocolTypePulsar, h.GetProtocolType())

	h.setProtocolType(ProtocolTypeGlobulaConsensus)
	require.Equal(t, ProtocolTypeGlobulaConsensus, h.GetProtocolType())
}

func TestHeader_setProtocolType_Panic(t *testing.T) {
	h := Header{}

	require.Panics(t, func() { h.setProtocolType(protocolTypeMax + 1) })
}

func TestHeader_GetPacketType(t *testing.T) {
	h := Header{}

	require.Equal(t, packets.PacketPhase0, h.GetPacketType())

	h.ProtocolAndPacketType = 1 // 0b00000001
	require.Equal(t, packets.PacketPhase1, h.GetPacketType())

	h.ProtocolAndPacketType = 2 // 0b00000010
	require.Equal(t, packets.PacketPhase2, h.GetPacketType())
}

func TestHeader_setPacketType(t *testing.T) {
	h := Header{}

	require.Equal(t, packets.PacketPhase0, h.GetPacketType())

	h.setPacketType(packets.PacketPhase3)
	require.Equal(t, packets.PacketPhase3, h.GetPacketType())
}

func TestHeader_setPacketType_Panic(t *testing.T) {
	h := Header{}

	require.Panics(t, func() { h.setPacketType(packetTypeMax + 1) })
}

func TestHeader_getPayloadLength(t *testing.T) {
	h := Header{}

	require.EqualValues(t, 0, h.getPayloadLength())

	h.HeaderAndPayloadLength = 123
	require.EqualValues(t, 123, h.getPayloadLength())
}

func TestHeader_setPayloadLength(t *testing.T) {
	h := Header{}

	require.EqualValues(t, 0, h.getPayloadLength())

	h.setPayloadLength(1000)
	require.EqualValues(t, 1000, h.getPayloadLength())
}

func TestHeader_setPayloadLength_Panic(t *testing.T) {
	h := Header{}

	require.Panics(t, func() { h.setPayloadLength(payloadLengthMax + 1) })
}

func TestHeader_SerializeTo(t *testing.T) {
	h := Header{
		ReceiverID: 123,
		SourceID:   456,
		TargetID:   789,
	}
	h.setIsBodyEncrypted(true)
	h.setIsRelayRestricted(true)
	h.setProtocolType(ProtocolTypeGlobulaConsensus)
	h.setPacketType(packets.PacketPhase3)

	buf := bytes.NewBuffer(make([]byte, 0, packetMaxSize))

	err := h.SerializeTo(nil, buf)
	require.NoError(t, err)
	require.Equal(t, 16, buf.Len())
}

func TestHeader_DeserializeFrom(t *testing.T) {
	h1 := Header{
		ReceiverID: 123,
		SourceID:   456,
		TargetID:   789,
	}
	h1.setIsBodyEncrypted(true)
	h1.setIsRelayRestricted(true)
	h1.setProtocolType(ProtocolTypeGlobulaConsensus)
	h1.setPacketType(packets.PacketPhase3)

	buf := bytes.NewBuffer(make([]byte, 0, packetMaxSize))
	err := h1.SerializeTo(nil, buf)
	require.NoError(t, err)

	h2 := Header{}
	err = h2.DeserializeFrom(nil, buf)
	require.NoError(t, err)

	require.Equal(t, h1, h2)
}

func TestPacket_getPulseNumber(t *testing.T) {
	p := Packet{}

	require.EqualValues(t, 0, p.getPulseNumber())

	p.PulseNumber = 123
	require.EqualValues(t, 123, p.getPulseNumber())
}

func TestPacket_setPulseNumber(t *testing.T) {
	p := Packet{}

	require.EqualValues(t, 0, p.getPulseNumber())

	p.setPulseNumber(1000)
	require.EqualValues(t, 1000, p.getPulseNumber())
}

func TestPacket_setPulseNumber_Panic(t *testing.T) {
	p := Packet{}

	require.Panics(t, func() { p.setPulseNumber(pulseNumberMax + 1) })
}

func TestPacket_SerializeTo_NilBody(t *testing.T) {
	p := Packet{}

	n, err := p.SerializeTo(context.Background(), bytes.NewBuffer(nil), digester, signer)

	require.Error(t, err)
	require.Contains(t, err.Error(), ErrNilBody.Error())
	require.EqualValues(t, 0, n)
}

func TestPacket_DeserializeFrom_NilBody(t *testing.T) {
	p := Packet{
		EncryptableBody: &GlobulaConsensusPacketBody{},
	}
	p.Header.setProtocolType(3) // Unknown protocol

	buf := bytes.NewBuffer(nil)
	_, err := p.SerializeTo(context.Background(), buf, digester, signer)
	require.NoError(t, err)

	n, err := p.DeserializeFrom(context.Background(), buf)
	require.Error(t, err)
	require.Contains(t, err.Error(), ErrNilBody.Error())
	require.EqualValues(t, 0, n)
}
