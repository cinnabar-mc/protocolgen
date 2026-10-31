package wire

import (
	"protocolgen/internal/manifest"
	"testing"
)

func TestLengthPrefixesRejectOutOfRangeCounts(t *testing.T) {
	for _, tc := range []struct {
		code  string
		limit int
	}{{"u8", 255}, {"i8", 127}} {
		prefix := manifest.Primitive(tc.code)
		for _, length := range []int{tc.limit, tc.limit + 1} {
			entries, values := make([]MapEntry, length), make([]any, length)
			for i := range values {
				values[i] = uint8(0)
				entries[i] = MapEntry{Key: uint8(0), Value: uint8(0)}
			}
			for _, item := range []struct {
				node  manifest.Node
				value any
			}{{manifest.Bytes(prefix), make([]byte, length)}, {manifest.String(prefix), string(make([]byte, length))}, {manifest.Array(prefix, manifest.Primitive("u8")), values}, {manifest.Map(prefix, manifest.Primitive("u8"), manifest.Primitive("u8")), entries}} {
				_, err := Encode(item.node, item.value)
				if (err != nil) != (length > tc.limit) {
					t.Errorf("%s/%s length %d: %v", tc.code, item.node.Kind, length, err)
				}
			}
		}
	}
}
