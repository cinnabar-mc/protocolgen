// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// StructureTemplateDataRequest is sent by the client to request data of a structure.
type StructureTemplateDataRequest struct {
	// StructureName is the name of the structure that was set in the structure block's UI. This is the name used
	// to export the structure to a file.
	StructureName string
	// StructurePosition is the position of the structure block that has its template data requested.
	StructurePosition protocol.BlockPos
	// StructureSettings is a struct of settings that should be used for exporting the structure. These settings
	// are identical to the last sent in the StructureBlockUpdate packet by the client.
	StructureSettings protocol.StructurePlacementSettings
	// RequestedOperation specifies the type of template data request that the player sent. It is one of the
	// constants found above.
	RequestedOperation protocol.StructureTemplateRequestOperation
}

// ID ...
func (*StructureTemplateDataRequest) ID() uint32 {
	return IDStructureTemplateDataRequest
}

func (pk *StructureTemplateDataRequest) Marshal(io protocol.IO) {
	io.StringLimits(&pk.StructureName, 0, 256)
	pk.StructurePosition.Marshal(io)
	pk.StructureSettings.Marshal(io)
	pk.RequestedOperation.Marshal(io)
}
