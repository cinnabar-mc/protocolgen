// Package block provides immutable generated Bedrock block-state properties.
package block

import "sort"

// Box is an axis-aligned box in block-local coordinates.
type Box [6]float32

// Block identifies a vanilla block type shared by one or more states.
type Block struct {
	Name string
	// TranslationKey is empty when the supplemental source omits this block.
	TranslationKey string
}

// Properties contains the static properties of a block state.
//
// Shape indexes are resolved with Shape. An index of zero means that the
// shape is unavailable because the source omitted it or supplied invalid data.
// An available empty shape has a non-zero
// index and resolves to an empty slice.
type Properties struct {
	Hardness                    float32
	ExplosionResistance         float32
	Friction                    float32
	Thickness                   float32
	Translucency                float32
	MapColor                    uint32
	TintMethod                  string
	LiquidReactionOnTouch       string
	LightEmission               uint8
	LightDampening              uint8
	BurnOdds                    uint8
	FlameOdds                   uint8
	CollisionShape              uint16
	OutlineShape                uint16
	VisualShape                 uint16
	UIShape                     uint16
	LiquidClipShape             uint16
	Solid                       bool
	RequiresCorrectToolForDrops bool
	CanContainLiquidSource      bool
	// SupplementalPropertiesAvailable covers TintMethod, LiquidReactionOnTouch
	// and CanContainLiquidSource. When false, these fields have no measured value.
	// Shapes report availability separately through Shape.
	SupplementalPropertiesAvailable bool
}

// State identifies one state in the source Bedrock block palette.
type State struct {
	Hash       uint32
	Block      uint16
	Properties uint16
}

// StateCount returns the number of generated block states.
func StateCount() int {
	return len(states)
}

// StateAt returns the block state at its source palette position.
func StateAt(index int) (State, bool) {
	if index < 0 || index >= len(states) {
		return State{}, false
	}
	return states[index], true
}

// StateByHash finds a block state by its Bedrock state hash.
func StateByHash(hash uint32) (State, bool) {
	index := sort.Search(len(stateHashIndex), func(i int) bool {
		return states[stateHashIndex[i]].Hash >= hash
	})
	if index == len(stateHashIndex) {
		return State{}, false
	}
	state := states[stateHashIndex[index]]
	return state, state.Hash == hash
}

// BlockAt resolves a State.Block index.
func BlockAt(index uint16) (Block, bool) {
	if int(index) >= len(blocks) {
		return Block{}, false
	}
	return blocks[index], true
}

// PropertiesAt resolves a State.Properties index.
func PropertiesAt(index uint16) (Properties, bool) {
	if int(index) >= len(properties) {
		return Properties{}, false
	}
	return properties[index], true
}

// Shape resolves a shape index. The returned slice must be treated as
// immutable. The boolean is false for an unavailable shape or invalid index.
func Shape(index uint16) ([]Box, bool) {
	if index == 0 || int(index) >= len(shapes) {
		return nil, false
	}
	return shapes[index], true
}
