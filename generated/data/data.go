// Package data reports the target release, source versions and catalog counts.
package data

// ModulePath identifies this runtime catalog module.
const ModulePath = "github.com/cinnabar-mc/protocolgen/generated/data"

// Counts reports the number of records in each generated dataset.
type Counts struct {
	BlockStates uint32 `json:"block_states"`
	Biomes      uint32 `json:"biomes"`
	VoxelShapes uint32 `json:"voxel_shapes"`
	Entities    uint32 `json:"entities"`
	Foods       uint32 `json:"foods"`
	// UnavailableLiquidClipShapes counts reviewed invalid source shapes omitted from the catalog.
	UnavailableLiquidClipShapes uint32 `json:"unavailable_liquid_clip_shapes"`
}
