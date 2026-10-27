// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// PrimitiveShapes is a packet sent by the server to instruct the client to render one or more shapes in the
// world. Shapes can be added, removed or updated based on the data provided individually.
type PrimitiveShapes struct {
	// ArrayOfPrimitiveShapesCanBeAMixOfNewUpdatedOrRemoved is a list of shapes to draw on the client-side.
	ArrayOfPrimitiveShapesCanBeAMixOfNewUpdatedOrRemoved []protocol.PrimitiveShape
}

// ID ...
func (*PrimitiveShapes) ID() uint32 {
	return IDPrimitiveShapes
}

func (pk *PrimitiveShapes) Marshal(io protocol.IO) {
	protocol.SliceLimits(io, &pk.ArrayOfPrimitiveShapesCanBeAMixOfNewUpdatedOrRemoved, 0, 1048576)
}
