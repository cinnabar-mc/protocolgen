package nbtencoding

import (
	"protocolgen/internal/manifest"
	"testing"
)

// decodeFixture builds one asymmetric packet field for overlay path tests.
func decodeFixture(root manifest.Node, decode *manifest.Node) manifest.Manifest {
	return manifest.Manifest{SchemaVersion: 2, Target: manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 1}, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "1", Digest: "fixture"}}, Packets: []manifest.Packet{{ID: 1, Name: "Packet", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{{Name: "Value", Encode: root, Decode: decode, Symmetry: manifest.Asymmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}}}
}

func TestNBTOverlayCoversExplicitDecodePaths(t *testing.T) {
	nbt := manifest.Primitive("nbt_le")
	field := manifest.Field{Name: "Inner", Encode: manifest.Void(), Decode: &nbt, Symmetry: manifest.Asymmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}
	for _, tc := range []struct {
		path   string
		root   manifest.Node
		decode *manifest.Node
	}{
		{"decode", manifest.Void(), &nbt},
		{"encode.fields[0].decode", manifest.Struct(field), nil},
		{"encode.variants[0].decode", manifest.Union(manifest.Primitive("u8"), manifest.Variant{Value: 0, Name: "Value", Encode: manifest.Void(), Decode: &nbt}), nil},
		{"encode.cases[0].decode[0]", manifest.Node{Kind: manifest.KindConditional, CompareTo: "selector", Cases: []manifest.Case{{Value: "0", Encode: []manifest.Node{manifest.Void()}, Decode: []manifest.Node{nbt}}}}, nil},
	} {
		t.Run(tc.path, func(t *testing.T) {
			m := decodeFixture(tc.root, tc.decode)
			if tc.decode == nil {
				m.Packets[0].Fields[0].Symmetry = manifest.Symmetric
			}
			table := Table{SchemaVersion: 1, Target: m.Target, Source: validSource()}
			if err := table.Validate(m.Target, m.Packets); err == nil {
				t.Error("accepted uncovered decode NBT")
			}
			table.Fields = []Entry{{PacketID: 1, PacketName: "Packet", FieldName: "Value", Path: tc.path, Encoding: manifest.NBTNetwork, Evidence: validEvidence()}}
			if err := table.Apply(&m); err != nil {
				t.Fatal(err)
			}
			if err := manifest.Validate(m); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestNBTOverlayRejectsPathAliasesBeforeMutation(t *testing.T) {
	for _, path := range []string{"encode", "encode.fields[0].encode"} {
		root := manifest.Primitive("nbt_le")
		if path != "encode" {
			root = manifest.Struct(manifest.Field{Name: "Inner", Encode: root, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}})
		}
		m := decodeFixture(root, nil)
		table := Table{SchemaVersion: 1, Target: m.Target, Source: validSource(), Fields: []Entry{{PacketID: 1, PacketName: "Packet", FieldName: "Value", Path: path, Encoding: manifest.NBTNetwork, Evidence: validEvidence()}, {PacketID: 1, PacketName: "Packet", FieldName: "Value", Path: path + ".", Encoding: manifest.NBTPersistent, Evidence: validEvidence()}}}
		if err := table.Apply(&m); err == nil {
			t.Error("accepted alias for duplicate node")
		}
		node := m.Packets[0].Fields[0].Encode
		if path != "encode" {
			node = node.Fields[0].Encode
		}
		if node.Encoding != "" {
			t.Fatal("failed validation mutated manifest")
		}
	}
}
