// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// UpdateBlock is sent by the server to update a block client-side, without resending the entire chunk that
// the block is located in. It is particularly useful for small modifications like block breaking/placing.
type UpdateBlock struct {
	// BlockPosition is the block position at which a block is updated.
	BlockPosition protocol.BlockPos
	// BlockRuntimeID is the runtime ID of the block that is placed at Position after sending the packet to the
	// client.
	BlockRuntimeID uint32
	// Flags is a combination of flags that specify the way the block is updated client-side. It is a combination
	// of the flags above, but typically sending only the BlockUpdateNetwork flag is sufficient.
	Flags uint32
	// Layer is the world layer on which the block is updated. For most blocks, this is the first layer, as that
	// layer is the default layer to place blocks on, but for blocks inside of each other, this differs.
	Layer uint32
}

// ID ...
func (*UpdateBlock) ID() uint32 {
	return IDUpdateBlock
}

func (pk *UpdateBlock) Marshal(io protocol.IO) {
	pk.BlockPosition.Marshal(io)
	io.Varuint32(&pk.BlockRuntimeID)
	io.Varuint32(&pk.Flags)
	io.Varuint32(&pk.Layer)
}
