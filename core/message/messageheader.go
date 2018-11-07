package message

import (
	"github.com/insolar/insolar/core"
)

// SignedMessageHeader is a struct with meta for the signed message
type SignedMessageHeader struct {
	Sender core.RecordRef
	Target core.RecordRef
	Role   core.JetRole
}

// NewSignedMessageHeader creates header from the message-body
func NewSignedMessageHeader(sender core.RecordRef, msg core.Message) SignedMessageHeader {
	return SignedMessageHeader{
		Sender: sender,
		Target: ExtractTarget(msg),
		Role:   ExtractRole(msg),
	}
}
