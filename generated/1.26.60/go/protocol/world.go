// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
)

// DimensionDefinition contains information specifying dimension-specific properties, used for data-driven
// dimensions. These include the range (the height min/max), generator variant, and more.
type DimensionDefinition struct {
	// MinimumY is the lowest Y coordinate that exists in the dimension.
	MinimumY int32
	// HeightRange is the number of blocks above MinimumY that exist in the dimension, so that the highest Y
	// coordinate in the dimension is MinimumY + HeightRange.
	HeightRange int32
	// Generator is the variant of generator that exists in the provided dimension. These can be one of the
	// constants defined above. If this is set to GeneratorLegacy, the legacy horizontal world limits will be
	// enforced.
	GeneratorType GeneratorType
	// DimensionType is the numeric identifier of the dimension. This cannot override a vanilla dimension (0-2),
	// but custom dimensions should start from 1000 like vanilla.
	DimensionType DimensionType
	// PackID is the UUID of the behaviour pack which has added the dimension.
	PackID uuid.UUID
	// DefaultBiome is the identifier of the biome that the dimension defaults to.
	DefaultBiome string
	CloudHeight  int32
	RenderClouds bool
}

// Marshal reads or writes DimensionDefinition using its canonical wire layout.
func (x *DimensionDefinition) Marshal(io IO) {
	io.Varint32(&x.MinimumY)
	io.Varint32(&x.HeightRange)
	x.GeneratorType.Marshal(io)
	x.DimensionType.Marshal(io)
	io.UUID(&x.PackID)
	io.StringLimits(&x.DefaultBiome, 0, 256)
	io.Varint32(&x.CloudHeight)
	io.Bool(&x.RenderClouds)
}

type WorldPosition struct {
	Position      mgl32.Vec3
	DimensionType DimensionType
}

// Marshal reads or writes WorldPosition using its canonical wire layout.
func (x *WorldPosition) Marshal(io IO) {
	io.Vec3(&x.Position)
	x.DimensionType.Marshal(io)
}
