package generator

import (
	"fmt"
	"math"

	"github.com/bedrock-mc/protocolgen/data/internal/nativeblock"
)

// applyNativeBlocks takes measured scalar properties from the native snapshot;
// Cloudburst remains the owner of the state-dependent shape and rendering fields.
func applyNativeBlocks(raw []rawBlockState, native nativeblock.Catalog) error {
	if len(raw) != len(native.States) {
		return fmt.Errorf("native state domain has %d entries, palette has %d", len(native.States), len(raw))
	}
	seen := make(map[string]bool, len(native.Properties))
	for i := range raw {
		value, measured := &raw[i], native.States[i]
		properties, exists := native.Properties[value.Name]
		if !exists {
			return fmt.Errorf("native properties omit palette block %s", value.Name)
		}
		// The catalog reserves the unknown block's wire identity explicitly.
		if value.Hash != *measured.Hash && !(value.Name == "minecraft:unknown" && value.Hash == math.MaxUint32-1) {
			return fmt.Errorf("native state %d hash %#x differs from palette %s/%#x", i, *measured.Hash, value.Name, value.Hash)
		}
		value.Hardness = *measured.Hardness
		value.ExplosionResistance = *measured.ExplosionResistance
		value.Friction = *measured.Friction
		value.LightEmission = *measured.LightEmission
		value.LightDampening = *measured.LightDamping
		value.FlameOdds = *measured.FlameOdds
		value.BurnOdds = *measured.BurnOdds
		value.Solid = *measured.Solid
		value.RequiresCorrectToolForDrops = *properties.RequiresCorrectToolForDrops
		seen[value.Name] = true
	}
	if len(seen) != len(native.Properties) {
		return fmt.Errorf("native properties contain blocks absent from the palette")
	}
	return nil
}
