package emitgo

import (
	"protocolgen/internal/manifest"
	"protocolgen/internal/naming"
	"protocolgen/internal/semantics"
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

func TestBitsetReservationsCannotAliasOpaqueTypeIDs(t *testing.T) {
	nodes := []manifest.Node{
		manifest.Struct(manifest.Field{Name: "Value", Encode: manifest.Primitive("u8"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}),
		manifest.Union(manifest.Primitive("u8"), manifest.Variant{Value: 0, Name: "Entry", Encode: manifest.Void()}),
		manifest.Enum("u8", manifest.EnumValue{Value: 0, Name: "Entry"}),
	}
	for _, node := range nodes {
		node.TypeID = "reserved:bitset:Bitset8"
		for _, ordered := range [][]manifest.Node{{node, manifest.Bitset(8)}, {manifest.Bitset(8), node}} {
			_, err := GenerateWithOptions(identityFixture(ordered...), Options{ProtocolImportPath: "wiregen", Naming: naming.Overlay{Names: map[string]string{node.TypeID: "Bitset8"}}})
			if err == nil {
				t.Errorf("reused an incompatible %s definition through an opaque type ID", node.Kind)
			}
		}
	}
}

func TestNonNativeIdentifierLayoutsCannotAlias(t *testing.T) {
	m := identityFixture(manifest.Primitive("var_u32"), manifest.Primitive("zigzag_i64"))
	overlay := semantics.Overlay{Fields: map[string]string{}}
	for _, field := range m.Packets[0].Fields {
		overlay.Fields[semantics.FieldKey("FixturePacket", field.Name)] = semantics.ActorRuntimeID
	}
	if _, err := GenerateWithOptions(m, Options{ProtocolImportPath: "wiregen", Semantics: overlay}); err == nil {
		t.Fatal("reused an identifier struct with a different wire layout")
	}
}

func TestSynthesizedUnionNamesCannotAliasManifestTypes(t *testing.T) {
	message := manifest.Struct(manifest.Field{Name: "Text", Encode: manifest.Primitive("u8"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}})
	message.TypeID = "Message"
	choice := manifest.Union(manifest.Primitive("u8"), manifest.Variant{Value: 0, Name: "First", Encode: message}, manifest.Variant{Value: 1, Name: "Second", Encode: message})
	choice.TypeID = "Choice"
	other := manifest.Struct(manifest.Field{Name: "Other", Encode: manifest.Primitive("u16le"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}})
	other.TypeID = "ChoiceSecond"
	for _, nodes := range [][]manifest.Node{{choice, other}, {other, choice}} {
		files, err := Generate(identityFixture(nodes...), "wiregen")
		if err != nil {
			continue
		}
		source := ""
		for _, text := range files {
			source += text
		}
		if !strings.Contains(source, "Other uint16") {
			t.Fatal("discarded manifest fields after a synthesized-name collision")
		}
	}
}

func TestRuntimeNamesCannotAliasManifestTypes(t *testing.T) {
	for _, name := range []string{"Optional", "Reader", "Writer", "OrderedEntry", "NewReader"} {
		node := manifest.Struct(manifest.Field{Name: "Value", Encode: manifest.Primitive("u8"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}})
		node.TypeID = name
		if _, err := Generate(identityFixture(node, manifest.Primitive("u8")), "wiregen"); err == nil {
			t.Errorf("accepted runtime collision %s", name)
		}
	}
}
