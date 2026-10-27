// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// AnvilDamage is sent by the client to request the dealing damage to an anvil. This packet is completely
// pointless and the server should never listen to it.
type AnvilDamage struct {
	// BlockPosition is the position in the world that the anvil can be found at.
	BlockPosition protocol.BlockPos
}

// ID ...
func (*AnvilDamage) ID() uint32 {
	return IDAnvilDamage
}

func (pk *AnvilDamage) Marshal(io protocol.IO) {
	pk.BlockPosition.Marshal(io)
}
