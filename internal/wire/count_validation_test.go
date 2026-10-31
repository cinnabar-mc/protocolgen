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

func TestZeroByteMapRespectsMaterializationBudget(t *testing.T) {
	for _, child := range []manifest.Node{manifest.Void(), manifest.Struct()} {
		node := manifest.Map(manifest.Primitive("var_u64"), child, child)
		data, err := Encode(manifest.Primitive("var_u64"), uint64(1<<20+1))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := DecodeAll(node, data); err == nil {
			t.Error("materialized a map beyond the default item budget")
		}
	}
}

func TestLargeZeroByteArraysRoundTrip(t *testing.T) {
	for _, node := range []manifest.Node{manifest.Array(manifest.Primitive("var_u32"), manifest.Void()), manifest.FixedArray(1<<20+1, manifest.Void())} {
		value := make([]any, 1<<20+1)
		data, err := Encode(node, value)
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := DecodeAllWithBudget(node, data, uint64(len(value)))
		if err != nil {
			t.Fatal(err)
		}
		if len(decoded.([]any)) != len(value) {
			t.Fatal("lost zero-byte elements")
		}
	}
}

func TestDuplicateStructNamesCannotOverwriteWireValues(t *testing.T) {
	node := manifest.Struct(manifest.Field{Ordinal: 0, Name: "X", Encode: manifest.Primitive("u8")}, manifest.Field{Ordinal: 1, Name: "X", Encode: manifest.Primitive("u16le")})
	if _, err := DecodeAll(node, []byte{7, 1, 2}); err == nil {
		t.Error("decoded duplicate field names")
	}
	if _, err := Encode(node, map[string]any{"X": uint16(513)}); err == nil {
		t.Error("encoded duplicate field names")
	}
}

func BenchmarkZeroByteMapCount(b *testing.B) {
	node := manifest.Map(manifest.Primitive("var_u64"), manifest.Void(), manifest.Void())
	data, err := Encode(manifest.Primitive("var_u64"), uint64(1<<20+1))
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		_, _ = DecodeAll(node, data)
	}
}

func TestMaximumZeroByteMapCountFailsWithoutAllocation(t *testing.T) {
	for _, child := range []manifest.Node{manifest.Void(), manifest.Struct()} {
		node := manifest.Map(manifest.Primitive("var_u64"), child, child)
		if _, err := DecodeAllWithBudget(node, []byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 1}, ^uint64(0)); err == nil {
			t.Fatal("accepted an unrepresentable map count")
		}
	}
}

func TestCollectionBudgetAppliesAcrossNestedCollections(t *testing.T) {
	node := manifest.Array(manifest.Primitive("u8"), manifest.Array(manifest.Primitive("u8"), manifest.Void()))
	if _, err := DecodeAllWithBudget(node, []byte{2, 1, 1}, 3); err == nil {
		t.Fatal("nested items escaped the shared budget")
	}
	if _, err := DecodeAllWithBudget(node, []byte{2, 1, 1}, 4); err != nil {
		t.Fatal(err)
	}
}
