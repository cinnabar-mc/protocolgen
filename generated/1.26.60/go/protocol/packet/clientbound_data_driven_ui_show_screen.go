// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// ClientBoundDataDrivenUIShowScreen is sent by the server to show a data-driven UI screen on the client.
type ClientboundDataDrivenUIShowScreen struct {
	// ScreenID is the identifier of the screen to show.
	ScreenID string
	// FormID is a unique instance ID for the form, used for scripting to identify specific screen instances.
	FormID uint32
	// DataInstanceID is an optional data ID associated with the screen.
	DataInstanceID protocol.Optional[uint32]
}

// ID ...
func (*ClientboundDataDrivenUIShowScreen) ID() uint32 {
	return IDClientboundDataDrivenUIShowScreen
}

func (pk *ClientboundDataDrivenUIShowScreen) Marshal(io protocol.IO) {
	io.StringLimits(&pk.ScreenID, 0, 500)
	io.Uint32(&pk.FormID)
	protocol.OptionalFunc(io, &pk.DataInstanceID, io.Uint32)
}
