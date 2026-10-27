// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// SetSpawnPosition is sent by the server to update the spawn position of a player, for example when sleeping
// in a bed.
type SetSpawnPosition struct {
	SpawnPositionType protocol.SpawnPositionType
	// Position is the new position of the spawn that was set. If SpawnType is SpawnTypeWorld, compasses will
	// point to this position. As of 1.16, Position is always the position of the player.
	BlockPosition protocol.BlockPos
	DimensionType protocol.DimensionType
	SpawnBlockPos protocol.BlockPos
}

// ID ...
func (*SetSpawnPosition) ID() uint32 {
	return IDSetSpawnPosition
}

func (pk *SetSpawnPosition) Marshal(io protocol.IO) {
	pk.SpawnPositionType.Marshal(io)
	pk.BlockPosition.Marshal(io)
	pk.DimensionType.Marshal(io)
	pk.SpawnBlockPos.Marshal(io)
}
