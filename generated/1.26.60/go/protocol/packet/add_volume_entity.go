// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// AddVolumeEntity sends a volume entity's definition and metadata from server to client.
type AddVolumeEntity struct {
	EntityNetworkID protocol.EntityNetID
	Components      []byte
	JSONIdentifier  string
	InstanceName    string
	MinBounds       protocol.BlockPos
	MaxBounds       protocol.BlockPos
	DimensionType   protocol.DimensionType
	// EngineVersion is the engine version the entity is using, for example, '1.17.0'.
	EngineVersion string
}

// ID ...
func (*AddVolumeEntity) ID() uint32 {
	return IDAddVolumeEntity
}

func (pk *AddVolumeEntity) Marshal(io protocol.IO) {
	pk.EntityNetworkID.Marshal(io)
	io.NBT(&pk.Components, protocol.NBTNetwork)
	io.StringLimits(&pk.JSONIdentifier, 1, 18446744073709551615)
	io.StringLimits(&pk.InstanceName, 1, 18446744073709551615)
	pk.MinBounds.Marshal(io)
	pk.MaxBounds.Marshal(io)
	pk.DimensionType.Marshal(io)
	io.String(&pk.EngineVersion)
}
