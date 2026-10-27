// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// PartyChanged is sent by the client to the server to indicate that the player's party ID has changed.
type PartyChanged struct {
	PartyInfo protocol.Optional[protocol.PlayerPartyInfo]
}

// ID ...
func (*PartyChanged) ID() uint32 {
	return IDPartyChanged
}

func (pk *PartyChanged) Marshal(io protocol.IO) {
	protocol.OptionalMarshaler(io, &pk.PartyInfo)
}
