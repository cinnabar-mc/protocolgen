// Package nativeblock validates the pinned native block measurement domain.
package nativeblock

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	"github.com/bedrock-mc/protocolgen/data/source"
)

// Properties are the per-type fields needed by the mining and property joins.
type Properties struct {
	RequiresCorrectToolForDrops *bool `json:"requiresCorrectToolForDrops"`
}

// State retains measured state fields without conflating absent and zero values.
type State struct {
	Hash                *uint32  `json:"hashedNetworkId"`
	Hardness            *float32 `json:"destroySpeed"`
	ExplosionResistance *float32 `json:"explosionResistance"`
	Friction            *float32 `json:"friction"`
	LightEmission       *uint8   `json:"lightEmission"`
	LightDamping        *uint8   `json:"lightDamping"`
	FlameOdds           *uint8   `json:"flameOdds"`
	BurnOdds            *uint8   `json:"burnOdds"`
	Solid               *bool    `json:"solid"`
}

// Catalog joins complete type and state domains with the sparse native tag map.
type Catalog struct {
	Properties map[string]Properties
	States     []State
	Tags       map[string][]string
}

// Read authenticates the snapshot and rejects a different protocol before decoding it.
func Read(ctx context.Context, sources *source.Sources, cache string) (Catalog, error) {
	if err := sources.ValidateNative(ctx, cache); err != nil {
		return Catalog{}, err
	}
	var raw struct {
		Properties map[string]Properties
		States     map[string]State
		Tags       map[string][]string
	}
	for _, input := range []struct {
		name string
		out  any
	}{
		{"native_block_properties", &raw.Properties},
		{"native_block_states", &raw.States},
		{"native_block_tags", &raw.Tags},
	} {
		data, err := sources.Read(ctx, input.name, cache)
		if err != nil {
			return Catalog{}, err
		}
		if err := json.Unmarshal(data, input.out); err != nil {
			return Catalog{}, fmt.Errorf("decode %s: %w", input.name, err)
		}
	}
	return Decode(raw.Properties, raw.States, raw.Tags)
}

// Decode validates measured fields, contiguous ordinals and tag ownership.
func Decode(properties map[string]Properties, states map[string]State, tags map[string][]string) (Catalog, error) {
	if len(properties) == 0 || len(states) == 0 {
		return Catalog{}, fmt.Errorf("native block domains are empty")
	}
	for name, p := range properties {
		if name == "" || p.RequiresCorrectToolForDrops == nil {
			return Catalog{}, fmt.Errorf("native block %q has no tool requirement", name)
		}
	}
	for name, members := range tags {
		if _, ok := properties[name]; !ok {
			return Catalog{}, fmt.Errorf("native tags reference unknown block %q", name)
		}
		seen := make(map[string]bool)
		for _, tag := range members {
			if tag == "" || seen[tag] {
				return Catalog{}, fmt.Errorf("native block %s has an empty or duplicate tag", name)
			}
			seen[tag] = true
		}
	}
	result := Catalog{Properties: properties, Tags: tags, States: make([]State, len(states))}
	hashes := make(map[uint32]bool, len(states))
	for i := range result.States {
		state, ok := states[strconv.Itoa(i)]
		if !ok || state.Hash == nil || state.Hardness == nil || state.ExplosionResistance == nil || state.Friction == nil || state.LightEmission == nil || state.LightDamping == nil || state.FlameOdds == nil || state.BurnOdds == nil || state.Solid == nil {
			return Catalog{}, fmt.Errorf("native state %d has missing measured fields", i)
		}
		if hashes[*state.Hash] || *state.LightEmission > 15 || *state.LightDamping > 15 {
			return Catalog{}, fmt.Errorf("native state %d has a duplicate identity or invalid light", i)
		}
		for _, v := range []float32{*state.Hardness, *state.ExplosionResistance, *state.Friction} {
			if math.IsInf(float64(v), 0) || math.IsNaN(float64(v)) {
				return Catalog{}, fmt.Errorf("native state %d has a non-finite property", i)
			}
		}
		hashes[*state.Hash] = true
		result.States[i] = state
	}
	return result, nil
}
