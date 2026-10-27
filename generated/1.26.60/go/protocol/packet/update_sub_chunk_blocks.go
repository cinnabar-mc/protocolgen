// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// UpdateSubChunkBlocks is essentially just UpdateBlock packet, however for a set of blocks in a sub-chunk.
type UpdateSubChunkBlocks struct {
	// Position is the block position of the sub-chunk being referred to.
	SubChunkBlockPosition protocol.BlockPos
	BlocksChanged         protocol.UpdateSubChunkBlocksChangedInfo
}

// ID ...
func (*UpdateSubChunkBlocks) ID() uint32 {
	return IDUpdateSubChunkBlocks
}

func (pk *UpdateSubChunkBlocks) Marshal(io protocol.IO) {
	pk.SubChunkBlockPosition.Marshal(io)
	pk.BlocksChanged.Marshal(io)
}
