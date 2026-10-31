package emitgo

import (
	"protocolgen/internal/manifest"
	"strings"
	"testing"
)

// identityFixture gives emitter regressions valid packet and source metadata.
func identityFixture(nodes ...manifest.Node) manifest.Manifest {
	m := manifest.Manifest{SchemaVersion: 2, Target: manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 1}, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "1", Digest: "fixture"}}, Packets: []manifest.Packet{{ID: 1, Name: "FixturePacket", Direction: manifest.DirectionClientbound}}}
	for i, node := range nodes {
		m.Packets[0].Fields = append(m.Packets[0].Fields, manifest.Field{Ordinal: i, Name: []string{"foo-bar", "foo_bar"}[i], Encode: node, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}})
	}
	return m
}

func TestPacketAnonymousHintsAreDistinct(t *testing.T) {
	first := manifest.Struct(manifest.Field{Name: "First", Encode: manifest.Primitive("u8"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}})
	second := manifest.Struct(manifest.Field{Name: "Second", Encode: manifest.Primitive("u16le"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}})
	files, err := Generate(identityFixture(first, second), "wiregen")
	if err != nil {
		t.Fatal(err)
	}
	source := ""
	for _, s := range files {
		source += s
	}
	for _, want := range []string{"First uint8", "Second uint16", "Uint16(&x.Second)"} {
		if !strings.Contains(source, want) {
			t.Errorf("missing distinct shape %q", want)
		}
	}
}

func TestRecursivePacketPayloadRemainsNamed(t *testing.T) {
	tree := manifest.Struct(manifest.Field{Name: "Children", Encode: manifest.Array(manifest.Primitive("var_u32"), manifest.Recursive("Tree")), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}})
	tree.TypeID = "Tree"
	files, err := Generate(identityFixture(tree), "wiregen")
	if err != nil {
		t.Fatal(err)
	}
	source := ""
	for _, s := range files {
		source += s
	}
	if !strings.Contains(source, "type Tree struct") || !strings.Contains(source, "Children []Tree") {
		t.Fatal("recursive definition missing")
	}
}

func TestBitsetNamesCannotReuseStructs(t *testing.T) {
	node := manifest.Struct(manifest.Field{Name: "Value", Encode: manifest.Primitive("u8"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}})
	node.TypeID = "Bitset8"
	for _, nodes := range [][]manifest.Node{{node, manifest.Bitset(8)}, {manifest.Bitset(8), node}} {
		if _, err := Generate(identityFixture(nodes...), "wiregen"); err == nil {
			t.Error("silently reused incompatible bitset definition")
		}
	}
}
