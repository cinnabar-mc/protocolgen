// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// BlockActorData is sent by the server to update data of a block entity client-side, for example the data of
// a chest.
type BlockActorData struct {
	// Position is the position of the block that holds the block entity. If no block entity is at this position,
	// the packet is ignored by the client.
	BlockPosition protocol.BlockPos
	ActorDataTags []byte
}

// ID ...
func (*BlockActorData) ID() uint32 {
	return IDBlockActorData
}

func (pk *BlockActorData) Marshal(io protocol.IO) {
	pk.BlockPosition.Marshal(io)
	io.NBT(&pk.ActorDataTags, protocol.NBTNetwork)
}
