package emitrust

import (
	"strings"
	"testing"

	"protocolgen/internal/manifest"
)

func TestRustAnonymousFieldHintsPreserveDistinctLayouts(t *testing.T) {
	field := func(ordinal int, name, code string) manifest.Field {
		return manifest.Field{Ordinal: ordinal, Name: name, Encode: manifest.Primitive(code), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}
	}
	first := manifest.Struct(field(0, "First", "u8"), field(1, "Tail", "u8"))
	second := manifest.Struct(field(0, "Second", "u16le"), field(1, "Tail", "u8"))
	for _, contract := range []int{0, CodecContractVersion} {
		m := contractFixture()
		m.Packets[0].Fields = m.Packets[0].Fields[:2]
		m.Packets[0].Fields[0].Name, m.Packets[0].Fields[0].Encode = "foo-bar", first
		m.Packets[0].Fields[1].Name, m.Packets[0].Fields[1].Encode = "foo_bar", second
		files, err := GenerateFilesWithOptions(m, Options{CodecContract: contract})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(files["src/types.rs"], "pub second: wire::U16LE") {
			t.Errorf("profile %d lost the second anonymous layout", contract)
		}
	}
}
