package ingest

import (
	"encoding/json"
	"strings"
	"testing"

	"protocolgen/internal/emitgo"
	"protocolgen/internal/emitrust"
	"protocolgen/internal/manifest"
)

func TestIntegerConstraintsRemainExactThroughRustGeneration(t *testing.T) {
	for _, bounds := range []struct{ code, min, max string }{
		{"uint64", "9007199254740993", "18446744073709551614"},
		{"int64", "-9007199254740993", "9007199254740993"},
	} {
		t.Run(bounds.code, func(t *testing.T) {
			lowerer := &mojangLowerer{documents: map[string]any{}, active: map[string]bool{}}
			node := lowerer.lowerSchema(map[string]any{"type": "integer", "x-underlying-type": bounds.code, "minimum": json.Number(bounds.min), "maximum": json.Number(bounds.max)}, "Packet.json", "Value")
			data, err := json.Marshal(node)
			if err != nil {
				t.Fatal(err)
			}
			for _, value := range []string{bounds.min, bounds.max} {
				if !strings.Contains(string(data), value) {
					t.Fatalf("manifest lost exact bound %s: %s", value, data)
				}
			}
			if err := json.Unmarshal(data, &node); err != nil {
				t.Fatal(err)
			}
			m := manifest.Manifest{SchemaVersion: 2, Target: manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 1}, Sources: []manifest.SourcePin{fixturePin("mojang")}, Packets: []manifest.Packet{{ID: 1, Name: "LimitsPacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{{Ordinal: 0, Name: "Value", Encode: node, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"mojang"}}}}}}}
			m.Sources[0].MinecraftVersion = "fixture"
			m.Sources[0].ProtocolVersion = 1
			goFiles, err := emitgo.Generate(m, "fixture/protocol")
			if err != nil {
				t.Fatal(err)
			}
			var goSource strings.Builder
			for _, contents := range goFiles {
				goSource.WriteString(contents)
			}
			for _, value := range []string{bounds.min, bounds.max} {
				if !strings.Contains(goSource.String(), ", "+value+")") {
					t.Fatalf("Go lost exact bound %s", value)
				}
			}
			files, err := emitrust.GenerateFiles(m)
			if err != nil {
				t.Fatal(err)
			}
			var source strings.Builder
			for name, contents := range files {
				if strings.HasSuffix(name, ".rs") {
					source.WriteString(contents)
				}
			}
			for _, value := range []string{bounds.min, bounds.max} {
				if !strings.Contains(source.String(), "Some("+value+")") {
					t.Fatalf("Rust lost exact bound %s", value)
				}
			}
		})
	}
}

func TestNullNumericBoundsRemainAbsent(t *testing.T) {
	for _, node := range []manifest.Node{manifest.String(manifest.Primitive("var_u32")), manifest.Primitive("u32le")} {
		got := withMojangConstraints(node, map[string]any{"minimum": nil, "maximum": nil})
		if got.Constraints != nil {
			t.Fatalf("null bounds became numeric constraints: %+v", got.Constraints)
		}
	}
}
