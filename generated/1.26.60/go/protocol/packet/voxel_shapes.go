// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// VoxelShapes is sent by the server to send voxel shape data to the client.
type VoxelShapes struct {
	// Shapes is a list of voxel shapes.
	Shapes []protocol.VoxelShapesSerializableVoxelShape
	// NameMap is a map of shape names to IDs.
	NameMap []protocol.OrderedEntry[string, protocol.VoxelShapesRegistryHandle]
	// CustomShapeCount is the number of custom shapes.
	CustomShapeCount uint16
}

// ID ...
func (*VoxelShapes) ID() uint32 {
	return IDVoxelShapes
}

func (pk *VoxelShapes) Marshal(io protocol.IO) {
	protocol.Slice(io, &pk.Shapes)
	protocol.OrderedMap(io, &pk.NameMap, io.Varuint32, io.String, func(value *protocol.VoxelShapesRegistryHandle) {
		value.Marshal(io)
	})
	io.Uint16(&pk.CustomShapeCount)
}
