package wire

import (
	"math"
	"protocolgen/internal/manifest"
	"testing"
)

func TestVarint32RejectsOverflow(t *testing.T) {
	for _, code := range []string{"var_u32", "var_i32", "zigzag_i32"} {
		for _, data := range [][]byte{{0x80, 0x80, 0x80, 0x80, 0x10}, {0x80, 0x80, 0x80, 0x80, 0x80, 0}} {
			if _, err := DecodeAll(manifest.Primitive(code), data); err == nil {
				t.Errorf("%s accepted %x", code, data)
			}
			if _, err := DecodeAll(manifest.Bytes(manifest.Primitive(code)), data); err == nil {
				t.Errorf("%s length accepted %x", code, data)
			}
		}
	}
	for _, tc := range []struct {
		code  string
		value any
	}{{"var_u32", uint32(math.MaxUint32)}, {"var_i32", int32(math.MinInt32)}, {"zigzag_i32", int32(math.MinInt32)}} {
		data, err := Encode(manifest.Primitive(tc.code), tc.value)
		if err != nil {
			t.Fatal(err)
		}
		got, err := DecodeAll(manifest.Primitive(tc.code), data)
		if err != nil || got != tc.value {
			t.Fatalf("%s boundary = %v, %v", tc.code, got, err)
		}
	}
}
