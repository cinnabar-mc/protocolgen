// Package nativeblock validates the pinned native block measurement domain.
package nativeblock

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	"github.com/bedrock-mc/protocolgen/data/source"
	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

// Properties are the per-type fields needed by the mining and property joins.
type Properties struct {
	RequiresCorrectToolForDrops *bool           `json:"requiresCorrectToolForDrops"`
	Thickness                   *float32        `json:"thickness"`
	Translucency                *float32        `json:"translucency"`
	MapColor                    *string         `json:"mapColor"`
	VisualShape                 json.RawMessage `json:"visualShape"`
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
	Names      []string
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
	if raw.Tags == nil {
		return Catalog{}, fmt.Errorf("native block tags must be a non-null object")
	}
	catalog, err := Decode(raw.Properties, raw.States, raw.Tags)
	if err != nil {
		return Catalog{}, err
	}
	if _, ok := sources.Lock.Inputs["block_palette"]; ok {
		palette, err := sources.Read(ctx, "block_palette", cache)
		if err != nil {
			return Catalog{}, err
		}
		if err := catalog.readNames(palette); err != nil {
			return Catalog{}, err
		}
	}
	return catalog, nil
}

// readNames checks the palette domain used to join supplemental data by hash.
func (c *Catalog) readNames(palette []byte) error {
	reader := bytes.NewReader(palette)
	decoder := nbt.NewDecoder(reader)
	seen := make(map[string]bool)
	for reader.Len() != 0 {
		if len(c.Names) >= len(c.States) {
			return fmt.Errorf("native palette has more names than measured states")
		}
		var state struct {
			Name    string         `nbt:"name"`
			States  map[string]any `nbt:"states"`
			Version int32          `nbt:"version"`
		}
		if err := decoder.Decode(&state); err != nil {
			return fmt.Errorf("native palette: %w", err)
		}
		if state.States == nil || state.Version <= 0 {
			return fmt.Errorf("native palette block %q lacks state properties or version", state.Name)
		}
		if _, ok := c.Properties[state.Name]; !ok {
			return fmt.Errorf("native palette names unknown block %q", state.Name)
		}
		c.Names = append(c.Names, state.Name)
		seen[state.Name] = true
	}
	if len(c.Names) != len(c.States) || len(seen) != len(c.Properties) {
		return fmt.Errorf("native palette does not cover the measured state and block domains")
	}
	return nil
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
