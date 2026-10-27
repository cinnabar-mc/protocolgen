// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// StructureBlockUpdate is sent by the client when it updates a structure block using the in-game UI. The data
// it contains depends on the type of structure block that it is. In Minecraft Bedrock Edition v1.11, there is
// only the Export structure block type, but in v1.13 the ones present in Java Edition will, according to the
// wiki, be added too.
type StructureBlockUpdate struct {
	// Position is the position of the structure block that is updated.
	BlockPosition protocol.BlockPos
	StructureData protocol.StructureEditorData
	// ShouldTrigger specifies if the structure block should be triggered immediately after this packet reaches
	// the server.
	Trigger bool
	// Waterlogged specifies if non-air blocks replace water or combine with water.
	IsWaterlogged bool
}

// ID ...
func (*StructureBlockUpdate) ID() uint32 {
	return IDStructureBlockUpdate
}

func (pk *StructureBlockUpdate) Marshal(io protocol.IO) {
	pk.BlockPosition.Marshal(io)
	pk.StructureData.Marshal(io)
	io.Bool(&pk.Trigger)
	io.Bool(&pk.IsWaterlogged)
}
