package generator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

// DecodeNativeBiomeIDs reads complete namespaced numeric identities without
// accepting duplicate names, shared IDs or implicit null values.
func DecodeNativeBiomeIDs(data []byte) (map[string]int32, error) {
	rows, err := decodeNativeBiomeObject(data)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("native biome ID table is empty")
	}
	ids := make(map[string]int32, len(rows))
	seen := make(map[int32]string, len(rows))
	for name, message := range rows {
		if !validNativeBiomeName(name) {
			return nil, fmt.Errorf("native biome has invalid name %q", name)
		}
		var id *int32
		if err := json.Unmarshal(message, &id); err != nil {
			return nil, fmt.Errorf("native biome %s ID: %w", name, err)
		}
		if id == nil || *id < 0 || *id > 65535 {
			return nil, fmt.Errorf("native biome %s has missing or invalid numeric ID", name)
		}
		if previous, exists := seen[*id]; exists {
			return nil, fmt.Errorf("native biomes %s and %s share numeric ID %d", previous, name, *id)
		}
		ids[name], seen[*id] = *id, name
	}
	return ids, nil
}

// DecodeNativeBiomes reads every scalar owned by rawBiome from the native dump.
// IDs come from the separate identity input and tag memberships stay with the
// source that provides them. Native water colours use #AARRGGBB ordering.
func DecodeNativeBiomes(data []byte) (map[string]rawBiome, error) {
	rows, err := decodeNativeBiomeObject(data)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("native biome definitions are empty")
	}
	biomes := make(map[string]rawBiome, len(rows))
	for name, message := range rows {
		if !validNativeBiomeName(name) {
			return nil, fmt.Errorf("native biome has invalid name %q", name)
		}
		fields, err := decodeNativeBiomeObject(message)
		if err != nil {
			return nil, fmt.Errorf("native biome %s: %w", name, err)
		}
		var biome rawBiome
		value := reflect.ValueOf(&biome).Elem()
		owner := value.Type()
		for i := range owner.NumField() {
			field := owner.Field(i)
			if field.Name == "ID" || field.Name == "Tags" {
				continue
			}
			key := strings.Split(field.Tag.Get("json"), ",")[0]
			raw, exists := fields[key]
			if !exists || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
				return nil, fmt.Errorf("native biome %s is missing %s", name, key)
			}
			if field.Name == "MapWaterColor" {
				var color string
				if err := json.Unmarshal(raw, &color); err != nil || len(color) != 9 || color[0] != '#' {
					return nil, fmt.Errorf("native biome %s has invalid #AARRGGBB water colour", name)
				}
				channels, err := strconv.ParseUint(color[1:], 16, 32)
				if err != nil {
					return nil, fmt.Errorf("native biome %s has invalid #AARRGGBB water colour", name)
				}
				biome.MapWaterColor.A = uint8(channels >> 24)
				biome.MapWaterColor.R = uint8(channels >> 16)
				biome.MapWaterColor.G = uint8(channels >> 8)
				biome.MapWaterColor.B = uint8(channels)
				continue
			}
			member := value.Field(i)
			if err := json.Unmarshal(raw, member.Addr().Interface()); err != nil {
				return nil, fmt.Errorf("native biome %s %s: %w", name, key, err)
			}
			if member.Kind() == reflect.Float32 || member.Kind() == reflect.Float64 {
				if f := member.Float(); math.IsNaN(f) || math.IsInf(f, 0) {
					return nil, fmt.Errorf("native biome %s has non-finite %s", name, key)
				}
			}
		}
		biomes[name] = biome
	}
	return biomes, nil
}

// applyNativeBiomes replaces scalar facts only for an identical biome domain,
// retaining independent tag memberships without sharing mutable tag slices.
func applyNativeBiomes(original, native map[string]rawBiome) (map[string]rawBiome, error) {
	if len(original) == 0 || len(original) != len(native) {
		return nil, fmt.Errorf("native biome definitions cover %d names, original source has %d", len(native), len(original))
	}
	combined := make(map[string]rawBiome, len(native))
	for name, value := range native {
		previous, exists := original[name]
		if !exists {
			return nil, fmt.Errorf("native biome %s is absent from the original source", name)
		}
		value.Tags = slices.Clone(previous.Tags)
		combined[name] = value
	}
	return combined, nil
}

// decodeNativeBiomeObject rejects duplicate JSON members and trailing documents.
func decodeNativeBiomeObject(data []byte) (map[string]json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	opening, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	if opening != json.Delim('{') {
		return nil, fmt.Errorf("native biome data must be a JSON object")
	}
	members := make(map[string]json.RawMessage)
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		name := token.(string)
		if _, exists := members[name]; exists {
			return nil, fmt.Errorf("duplicate native biome member %q", name)
		}
		var message json.RawMessage
		if err := decoder.Decode(&message); err != nil {
			return nil, err
		}
		members[name] = message
	}
	if _, err := decoder.Token(); err != nil {
		return nil, err
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		return nil, fmt.Errorf("native biome data has trailing JSON")
	}
	return members, nil
}

// validNativeBiomeName accepts the namespaced vanilla identifiers used by the dump.
func validNativeBiomeName(name string) bool {
	const prefix = "minecraft:"
	if !strings.HasPrefix(name, prefix) || len(name) == len(prefix) {
		return false
	}
	return !strings.ContainsFunc(name[len(prefix):], func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || strings.ContainsRune("_.-/", r))
	})
}
