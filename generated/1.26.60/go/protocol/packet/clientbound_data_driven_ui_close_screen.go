// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// ClientBoundDataDrivenUICloseScreen is sent by the server to close a data-driven UI screen on the client. If
// FormID is not set, all data-driven UI screens are closed.
type ClientboundDataDrivenUICloseScreen struct {
	// FormID is the optional unique instance ID of the form to close. If not set, all forms are closed.
	FormID protocol.Optional[uint32]
}

// ID ...
func (*ClientboundDataDrivenUICloseScreen) ID() uint32 {
	return IDClientboundDataDrivenUICloseScreen
}

func (pk *ClientboundDataDrivenUICloseScreen) Marshal(io protocol.IO) {
	protocol.OptionalFunc(io, &pk.FormID, io.Uint32)
}
