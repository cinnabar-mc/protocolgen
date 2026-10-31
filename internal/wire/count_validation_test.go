package wire

import (
	"protocolgen/internal/manifest"
	"testing"
)

func TestMapCountsCannotPanicOrAllocateBeforeReading(t *testing.T) {
	for _, tc := range []struct {
		code string
		data []byte
	}{{"var_u64", []byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 1}}, {"i8", []byte{255}}, {"u32le", []byte{255, 255, 255, 127}}} {
		t.Run(tc.code, func(t *testing.T) {
			defer func() {
				if p := recover(); p != nil {
					t.Fatalf("decode panicked: %v", p)
				}
			}()
			node := manifest.Map(manifest.Primitive(tc.code), manifest.Primitive("u8"), manifest.Primitive("u8"))
			if _, err := DecodeAll(node, tc.data); err == nil {
				t.Fatal("accepted truncated map")
			}
		})
	}
}

func TestEmptyMapKeepsAnEmptyEntrySlice(t *testing.T) {
	node := manifest.Map(manifest.Primitive("var_u32"), manifest.Primitive("u8"), manifest.Primitive("u8"))
	value, err := DecodeAll(node, []byte{0})
	if err != nil {
		t.Fatal(err)
	}
	if entries, ok := value.([]MapEntry); !ok || entries == nil {
		t.Fatalf("empty map = %#v", value)
	}
}
