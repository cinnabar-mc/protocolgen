package emitrust

import (
	"fmt"
	"strings"
	"testing"

	"protocolgen/internal/manifest"
)

// collectionFixture places one collection at a stable packet field site.
func collectionFixture(node manifest.Node) manifest.Manifest {
	m := contractFixture()
	m.Packets[0].Fields = m.Packets[0].Fields[:1]
	m.Packets[0].Fields[0].Encode = node
	return m
}

func TestCollectionMinimumUsesUnionPayload(t *testing.T) {
	payload := manifest.Node{Kind: manifest.KindStruct, Fields: []manifest.Field{
		{Name: "Type", Encode: manifest.Primitive("u8")},
		{Name: "Value", Encode: manifest.Primitive("u8")},
	}}
	union := manifest.Union(manifest.Primitive("u8"), manifest.Variant{Value: 0, Name: "Byte", Encode: payload})
	if got := (&generator{}).minSizeOf(union); got != 2 {
		t.Fatalf("union minimum = %d, want 2", got)
	}
	for _, node := range []manifest.Node{{Kind: manifest.KindStruct}, manifest.Void()} {
		if got := (&generator{}).minSizeOf(node); got != 0 {
			t.Errorf("%s minimum = %d, want 0", node.Kind, got)
		}
	}
}

func TestCollectionConstraintsReachElementCodecs(t *testing.T) {
	number := manifest.Primitive("u16le")
	number.Constraints = &manifest.Constraints{Maximum: pointerTo(4.0)}
	text := manifest.String(manifest.Primitive("var_u32"))
	text.Constraints = &manifest.Constraints{MaxLength: pointerTo(uint64(1))}
	fixed := manifest.Array(manifest.Primitive("u32le"), manifest.Primitive("u16le"))
	fixed.Constraints = &manifest.Constraints{MaxItems: pointerTo(uint64(1))}
	for _, contract := range []int{0, CodecContractVersion} {
		for _, tc := range []struct {
			name  string
			node  manifest.Node
			check string
		}{
			{"array", manifest.Array(manifest.Primitive("var_u32"), number), "validate_number_limits"},
			{"map-key", manifest.Map(manifest.Primitive("var_u32"), number, manifest.Primitive("u8")), "validate_number_limits"},
			{"map-value", manifest.Map(manifest.Primitive("var_u32"), manifest.Primitive("u8"), number), "validate_number_limits"},
			{"string", manifest.Array(manifest.Primitive("var_u32"), text), "decode_string_limits"},
			{"fixed-prefix", fixed, "decode_collection_u32le_limits"},
		} {
			t.Run(fmt.Sprintf("%d/%s", contract, tc.name), func(t *testing.T) {
				files, err := GenerateFilesWithOptions(collectionFixture(tc.node), Options{CodecContract: contract})
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(files["src/packets.rs"], tc.check) {
					t.Fatalf("element/bound check %q missing:\n%s", tc.check, files["src/packets.rs"])
				}
			})
		}
	}
}

func TestCollectionPrefixesArePreserved(t *testing.T) {
	for _, contract := range []int{0, CodecContractVersion} {
		for _, node := range []manifest.Node{
			manifest.String(manifest.Primitive("u32le")),
			manifest.Bytes(manifest.Primitive("u32le")),
			manifest.Map(manifest.Primitive("u32le"), manifest.Primitive("u8"), manifest.Primitive("u8")),
			manifest.Map(manifest.Primitive("u32le"), manifest.Optional(manifest.Primitive("u8")), manifest.Primitive("u8")),
		} {
			files, err := GenerateFilesWithOptions(collectionFixture(node), Options{CodecContract: contract})
			if err != nil {
				t.Fatal(err)
			}
			source := files["src/packets.rs"]
			if !strings.Contains(source, "u32le") && !strings.Contains(source, "from_le_bytes") {
				t.Errorf("%d/%s discards prefix:\n%s", contract, node.Kind, source)
			}
		}
	}
}

func TestUnsupportedCollectionPrefixesFailGeneration(t *testing.T) {
	for _, contract := range []int{0, CodecContractVersion} {
		prefix := manifest.Primitive("u8")
		for _, node := range []manifest.Node{manifest.String(prefix), manifest.Bytes(prefix), manifest.Map(prefix, manifest.Primitive("u8"), manifest.Primitive("u8"))} {
			if _, err := GenerateFilesWithOptions(collectionFixture(node), Options{CodecContract: contract}); err == nil {
				t.Errorf("%d/%s silently substitutes a prefix", contract, node.Kind)
			}
		}
	}
}
