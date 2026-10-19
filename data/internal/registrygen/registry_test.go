package registrygen

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"github.com/bedrock-mc/protocolgen/data/source"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/nbt"
)

// TestGeneratePreservesRegistryContracts checks state order, complete compounds,
// numeric types, empty components, and deterministic output bytes.
func TestGeneratePreservesRegistryContracts(t *testing.T) {
	input := registryFixture(t)
	files, err := Generate(input, "fixture-lock")
	if err != nil {
		t.Fatal(err)
	}
	again, err := Generate(input, "fixture-lock")
	if err != nil || !reflect.DeepEqual(files, again) {
		t.Fatalf("generation is not deterministic: %v", err)
	}
	reader := bytes.NewReader(files["registry/block_states.nbt"])
	decoder := nbt.NewDecoder(reader)
	for _, expected := range []string{"minecraft:z", "minecraft:a"} {
		var state blockState
		if err := decoder.Decode(&state); err != nil {
			t.Fatal(err)
		}
		if state.Name != expected || state.Version != 1 || state.States["age"] != int32(3) || state.States["open"] != uint8(1) {
			t.Fatalf("state order or types changed: %+v", state)
		}
	}
	if reader.Len() != 0 {
		t.Fatalf("state stream has %d trailing bytes", reader.Len())
	}
	var driven struct {
		Blocks []dataDrivenBlock `nbt:"blocks"`
	}
	if err := nbt.UnmarshalEncoding(files["registry/data_driven_blocks.nbt"], &driven, nbt.LittleEndian); err != nil {
		t.Fatal(err)
	}
	if len(driven.Blocks) != 1 || driven.Blocks[0].Name != "minecraft:z" || !reflect.DeepEqual(driven.Blocks[0].Components, fixtureComponents()) {
		t.Fatalf("data-driven component payload changed: %#v", driven)
	}
	var items map[string]itemEntry
	if err := nbt.Unmarshal(files["registry/items.nbt"], &items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || items["minecraft:z"].RuntimeID != -7 || !items["minecraft:z"].ComponentBased || items["minecraft:z"].Version != 2 || !reflect.DeepEqual(items["minecraft:z"].Data, fixtureComponents()) {
		t.Fatalf("item identity or component payload changed: %#v", items)
	}
	if items["minecraft:a"].Data == nil || len(items["minecraft:a"].Data) != 0 {
		t.Fatal("empty component compound was not retained")
	}
}

// TestGenerateRejectsRegistryDrift checks incomplete joins and invalid source data.
func TestGenerateRejectsRegistryDrift(t *testing.T) {
	for _, test := range []struct {
		name   string
		change func(*Inputs)
		want   string
	}{
		{"truncated_palette", func(in *Inputs) { in.BlockPalette = in.BlockPalette[:len(in.BlockPalette)-4] }, "block palette"},
		{"missing_components", func(in *Inputs) { in.ItemComponents = sourceNBT(t, map[string]any{"minecraft:a": map[string]any{}}) }, "component count"},
		{"unknown_components", func(in *Inputs) {
			in.ItemComponents = sourceNBT(t, map[string]any{"minecraft:a": map[string]any{}, "minecraft:extra": map[string]any{}})
		}, "no component compound"},
		{"missing_registry_field", func(in *Inputs) { in.ItemRegistry = []byte(`[{"name":"minecraft:a","id":1,"version":2}]`) }, "incomplete registry fields"},
		{"duplicate_runtime_id", func(in *Inputs) {
			in.ItemRegistry = []byte(`[{"name":"minecraft:a","id":1,"version":2,"componentBased":false},{"name":"minecraft:z","id":1,"version":2,"componentBased":true}]`)
		}, "share runtime ID"},
		{"oversized_runtime_id", func(in *Inputs) { in.ItemRegistry = bytes.Replace(in.ItemRegistry, []byte(`-7`), []byte(`40000`), 1) }, "invalid runtime ID"},
		{"unknown_block_components", func(in *Inputs) {
			in.DataDrivenBlocks = sourceNBT(t, map[string]any{"minecraft:unknown": map[string]any{}})
		}, "no palette state"},
		{"empty_palette", func(in *Inputs) {
			in.BlockPalette = nil
		}, "palette is empty"},
		{"duplicate_block_state", func(in *Inputs) {
			states := fixtureStates()
			states[1].Name = states[0].Name
			in.BlockPalette = paletteNBT(t, states)
		}, "duplicate block state"},
		{"invalid_property_type", func(in *Inputs) {
			states := fixtureStates()
			states[0].States["invalid"] = float32(1)
			in.BlockPalette = paletteNBT(t, states)
		}, "unsupported type"},
	} {
		t.Run(test.name, func(t *testing.T) {
			input := registryFixture(t)
			test.change(&input)
			if _, err := Generate(input, "fixture-lock"); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
		})
	}
}

// TestLockedRegistryOutput checks the committed projection when authenticated
// Cloudburst inputs are available locally for the release regeneration check.
func TestLockedRegistryOutput(t *testing.T) {
	dir := os.Getenv("PROTOCOLGEN_REGISTRY_DIR")
	if dir == "" {
		t.Skip("set PROTOCOLGEN_REGISTRY_DIR to the locked Cloudburst input directory")
	}
	var input Inputs
	inputs, err := source.Open("")
	if err != nil {
		t.Fatal(err)
	}
	input.BlockPalette, err = inputs.Read(context.Background(), "block_palette", "")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range []struct {
		name string
		out  *[]byte
	}{{"data_driven_blocks.nbt", &input.DataDrivenBlocks}, {"runtime_item_states.json", &input.ItemRegistry}, {"item_components.nbt", &input.ItemComponents}} {
		data, err := os.ReadFile(filepath.Join(dir, file.name))
		if err != nil {
			t.Fatal(err)
		}
		*file.out = data
	}
	files, err := Generate(input, "ignored-metadata")
	if err != nil {
		t.Fatal(err)
	}
	for name, data := range files {
		if !strings.HasSuffix(name, ".nbt") {
			continue
		}
		committed, err := os.ReadFile(filepath.Join("..", "..", "..", "generated", "data", filepath.FromSlash(name)))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(committed, data) {
			t.Errorf("%s is stale", name)
		}
	}
}

// registryFixture builds small public-schema inputs covering the consumer formats.
func registryFixture(t *testing.T) Inputs {
	t.Helper()
	registry, err := json.Marshal([]map[string]any{
		{"name": "minecraft:z", "id": -7, "componentBased": true, "version": 2},
		{"name": "minecraft:a", "id": 1, "componentBased": false, "version": 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	return Inputs{
		BlockPalette:     paletteNBT(t, fixtureStates()),
		DataDrivenBlocks: sourceNBT(t, map[string]any{"minecraft:z": fixtureComponents()}),
		ItemRegistry:     registry,
		ItemComponents:   sourceNBT(t, map[string]any{"minecraft:z": fixtureComponents(), "minecraft:a": map[string]any{}}),
	}
}

// fixtureStates returns a non-lexical source palette with multiple NBT property types.
func fixtureStates() []blockState {
	return []blockState{
		{Name: "minecraft:z", States: map[string]any{"age": int32(3), "open": uint8(1)}, Version: 1},
		{Name: "minecraft:a", States: map[string]any{"age": int32(3), "open": uint8(1)}, Version: 1},
	}
}

// fixtureComponents covers nested maps, ordered lists, arrays, and exact numeric tag widths.
func fixtureComponents() map[string]any {
	return map[string]any{"nested": map[string]any{"byte": uint8(1), "short": int16(2), "int": int32(3), "long": int64(4), "float": float32(1.25), "double": float64(2.5)}, "ordered": []int32{9, 2, 7}, "bytes": [3]byte{1, 2, 3}}
}

// sourceNBT encodes a synthetic source using Cloudburst's gzip and big-endian envelope.
func sourceNBT(t *testing.T, value any) []byte {
	t.Helper()
	encoded, err := encodeNBT(value, nbt.BigEndian)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	writer := gzip.NewWriter(&output)
	if _, err := writer.Write(encoded); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}

// paletteNBT emits native palette compounds in their sequential network order.
func paletteNBT(t *testing.T, states []blockState) []byte {
	t.Helper()
	var output bytes.Buffer
	encoder := nbt.NewEncoder(&output)
	encoder.SortMaps = true
	for _, state := range states {
		if err := encoder.Encode(state); err != nil {
			t.Fatal(err)
		}
	}
	return output.Bytes()
}
