// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// UpdateClientOptions is sent by the client when some of the client's options are updated, such as the
// graphics mode.
type UpdateClientOptions struct {
	// GraphicsModeChange is the graphics mode that the client is using. It is one of the constants above.
	GraphicsModeChange protocol.Optional[protocol.GraphicsMode]
	// FilterProfanityChange is if the client only uses filtered messages or not.
	FilterProfanityChange protocol.Optional[bool]
}

// ID ...
func (*UpdateClientOptions) ID() uint32 {
	return IDUpdateClientOptions
}

func (pk *UpdateClientOptions) Marshal(io protocol.IO) {
	protocol.OptionalMarshaler(io, &pk.GraphicsModeChange)
	protocol.OptionalFunc(io, &pk.FilterProfanityChange, io.Bool)
}
