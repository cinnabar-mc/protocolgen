package parity

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"protocolgen/internal/manifest"
)

func TestAxolotlV1NormalizationMatchesCanonicalByteShape(t *testing.T) {
	m := manifest.Manifest{
		SchemaVersion: 2,
		Target:        manifest.Target{MinecraftVersion: "1.26.40", ProtocolVersion: 2168},
		Sources:       []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:parity", MinecraftVersion: "1.26.40", ProtocolVersion: 2168}},
		Packets: []manifest.Packet{{ID: 7, Name: "Fixture", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{
			{Ordinal: 0, Name: "Value", Encode: manifest.Primitive("u8"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
			{Ordinal: 1, Name: "Maybe", Encode: manifest.Optional(manifest.String(manifest.Primitive("var_u32"))), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		}}},
	}
	axolotl := `{"schema_version":1,"source":"fixture","minecraft_version":"1.26.40","protocol_version":2168,"packets":[{"id":7,"name":"Fixture","operations":[{"kind":"primitive","field":"Value","op":"U8"},{"kind":"option","field":"Maybe","presence":"Bool","value":[{"kind":"string","field":"Maybe","prefix":"VarInt","encoding":"utf8"}]}]}]}`
	path := filepath.Join(t.TempDir(), "axolotl.json")
	if err := os.WriteFile(path, []byte(axolotl), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := CompareFile(m, path); err != nil {
		t.Fatalf("CompareFile: %v", err)
	}
}

func TestAxolotlParityRejectsProtocolMix(t *testing.T) {
	m := manifest.Manifest{SchemaVersion: 2, Target: manifest.Target{MinecraftVersion: "1.26.40", ProtocolVersion: 2168}, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:parity", MinecraftVersion: "1.26.40", ProtocolVersion: 2168}}, Packets: []manifest.Packet{{ID: 1, Name: "P", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{{Ordinal: 0, Name: "Value", Encode: manifest.Primitive("u8"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}}}
	path := filepath.Join(t.TempDir(), "axolotl.json")
	if err := os.WriteFile(path, []byte(`{"schema_version":1,"source":"fixture","minecraft_version":"1.26.50","protocol_version":2169,"packets":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := CompareFile(m, path); err == nil || !strings.Contains(err.Error(), "mix") {
		t.Fatalf("CompareFile error = %v, want protocol mix failure", err)
	}
}

func TestNBTNormalizationRetainsEncoding(t *testing.T) {
	got, err := normalizeAxolotlOps([]axolotlOp{{Kind: "primitive", Field: "Data", Op: "Nbt"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, encoding := range []manifest.NBTEncoding{manifest.NBTNetwork, manifest.NBTPersistent, manifest.NBTNetworkCompound} {
		want, err := normalizeNode("Data", manifest.NBT(encoding))
		if err != nil {
			t.Fatal(err)
		}
		if reflect.DeepEqual(want, got) != (encoding == manifest.NBTNetwork) {
			t.Errorf("NBT %s comparison = %v", encoding, reflect.DeepEqual(want, got))
		}
	}
}

func TestNormalizationRejectsNestedDecodeOverrides(t *testing.T) {
	decode := manifest.Primitive("u16le")
	nested := manifest.Struct(manifest.Field{Name: "Inner", Encode: manifest.Primitive("u8"), Decode: &decode, Symmetry: manifest.Asymmetric})
	control := manifest.Primitive("u8")
	for name, node := range map[string]manifest.Node{
		"struct":      nested,
		"array":       manifest.Array(control, nested),
		"union":       {Kind: manifest.KindUnion, Control: &control, Variants: []manifest.Variant{{Value: 1, Name: "One", Encode: control, Decode: &decode}}},
		"conditional": {Kind: manifest.KindConditional, CompareTo: "Mode", Cases: []manifest.Case{{Value: "1", Encode: []manifest.Node{control}, Decode: []manifest.Node{decode}}}},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := normalizeNode("Data", node); err == nil {
				t.Fatal("accepted decode override")
			}
		})
	}
}

func TestNormalizationMatchesEmptyAndCountedBytes(t *testing.T) {
	control := manifest.Primitive("u8")
	for _, tc := range []struct {
		name string
		node manifest.Node
		op   axolotlOp
	}{
		{"conditional", manifest.Node{Kind: manifest.KindConditional, CompareTo: "Mode", Cases: []manifest.Case{{Value: "1", Encode: []manifest.Node{control}}}}, axolotlOp{Kind: "conditional", Field: "Data", CompareTo: "Mode", Cases: []axolotlCase{{Value: "1", Operations: []axolotlOp{{Kind: "primitive", Field: "Data", Op: "U8"}}}}}},
		{"void union", manifest.Node{Kind: manifest.KindUnion, Control: &control, Variants: []manifest.Variant{{Value: 1, Name: "Empty", Encode: manifest.Node{Kind: manifest.KindVoid}}}}, axolotlOp{Kind: "union", Field: "Data", Control: "U8", Variants: []axolotlVariant{{Value: 1, Name: "Empty"}}}},
		{"bytes", manifest.Bytes(manifest.Primitive("var_u32")), axolotlOp{Kind: "array", Field: "Data", Prefix: "VarInt", Element: []axolotlOp{{Kind: "primitive", Field: "Data[]", Op: "U8"}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			want, err := normalizeNode("Data", tc.node)
			if err != nil {
				t.Fatal(err)
			}
			got, err := normalizeAxolotlOps([]axolotlOp{tc.op})
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(want, got) {
				t.Fatalf("canonical %#v != source %#v", want, got)
			}
		})
	}
	want, err := normalizeCanonical(manifest.Manifest{Packets: []manifest.Packet{{ID: 1, Name: "Empty"}}})
	if err != nil {
		t.Fatal(err)
	}
	got, err := normalizeAxolotl(axolotlManifest{Packets: []axolotlPacket{{ID: 1, Name: "Empty"}}})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(want, got) {
		t.Fatal("empty packets differ")
	}
	bytes, _ := normalizeNode("Data", manifest.Bytes(manifest.Primitive("var_u32")))
	for _, op := range []axolotlOp{
		{Kind: "array", Field: "Data", Prefix: "U16LE", Element: []axolotlOp{{Kind: "primitive", Field: "Data[]", Op: "U8"}}},
		{Kind: "array", Field: "Data", Prefix: "VarInt", Element: []axolotlOp{{Kind: "primitive", Field: "Data[]", Op: "U16LE"}}},
	} {
		other, err := normalizeAxolotlOps([]axolotlOp{op})
		if err != nil {
			t.Fatal(err)
		}
		if reflect.DeepEqual(bytes, other) {
			t.Fatal("different byte shape compared equal")
		}
	}
}
