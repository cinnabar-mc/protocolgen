package generator

import (
	"fmt"
	"math"

	"github.com/bedrock-mc/protocolgen/data/internal/nativeblock"
)

// alignNativeBlocks keeps the native palette order and joins supplemental
// fields only when both the block name and state hash match. Missing shapes stay unavailable.
func alignNativeBlocks(raw []rawBlockState, native nativeblock.Catalog) ([]rawBlockState, error) {
	if len(native.Names) != len(native.States) {
		return nil, fmt.Errorf("native palette names do not match its state domain")
	}
	byHash := make(map[uint32]rawBlockState, len(raw))
	for _, state := range raw {
		if _, exists := byHash[state.Hash]; exists {
			return nil, fmt.Errorf("duplicate supplemental block hash %d", state.Hash)
		}
		byHash[state.Hash] = state
	}
	states := make([]rawBlockState, len(native.States))
	for i, measured := range native.States {
		name, hash := native.Names[i], *measured.Hash
		if name == "minecraft:unknown" {
			hash = math.MaxUint32 - 1
		}
		state, found := byHash[hash]
		if found && state.Name != name {
			return nil, fmt.Errorf("supplemental hash %d names %s, native palette names %s", hash, state.Name, name)
		}
		if !found {
			properties := native.Properties[name]
			if properties.Thickness == nil || properties.Translucency == nil || properties.MapColor == nil {
				return nil, fmt.Errorf("native block %s lacks rendering scalars", name)
			}
			if !finiteFloat32(*properties.Thickness) || !finiteFloat32(*properties.Translucency) {
				return nil, fmt.Errorf("native block %s has non-finite rendering scalars", name)
			}
			// Block map colours already use the generator's #RRGGBBAA order.
			color := *properties.MapColor
			if len(color) != 9 || color[0] != '#' {
				return nil, fmt.Errorf("native block %s has invalid map colour", name)
			}
			state = rawBlockState{Name: name, Hash: hash, Thickness: *properties.Thickness,
				Translucency: *properties.Translucency, MapColor: color,
				VisualShape: properties.VisualShape, supplementalUnavailable: true}
		}
		states[i] = state
	}
	if err := applyNativeBlocks(states, native); err != nil {
		return nil, err
	}
	return states, nil
}

// finiteFloat32 rejects non-finite measured scalar values before Go emission.
func finiteFloat32(value float32) bool {
	return !math.IsNaN(float64(value)) && !math.IsInf(float64(value), 0)
}

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
