// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// ClientBoundDebugRenderer is sent by the server to spawn an outlined cube on client-side.
type ClientboundDebugRenderer struct {
	// Type is the type of action. It is one of the constants above.
	Type            string
	DebugMarkerData protocol.Optional[protocol.DebugMarkerData]
}

// ID ...
func (*ClientboundDebugRenderer) ID() uint32 {
	return IDClientboundDebugRenderer
}

func (pk *ClientboundDebugRenderer) Marshal(io protocol.IO) {
	io.String(&pk.Type)
	protocol.OptionalMarshaler(io, &pk.DebugMarkerData)
}
