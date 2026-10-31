package wire

import (
	"bytes"
	"protocolgen/internal/manifest"
	"reflect"
	"testing"
)

func TestAsymmetricDecodeUsesDeclaredShape(t *testing.T) {
	decoded := manifest.Primitive("u16le")
	for _, node := range []manifest.Node{
		manifest.Struct(manifest.Field{Name: "Value", Encode: manifest.Primitive("u8"), Decode: &decoded, Symmetry: manifest.Asymmetric}),
		manifest.Union(manifest.Primitive("u8"), manifest.Variant{Value: 0, Name: "Value", Encode: manifest.Primitive("u8"), Decode: &decoded}),
	} {
		input := []byte{1, 2}
		var want, value any = map[string]any{"Value": uint16(513)}, map[string]any{"Value": uint8(7)}
		expected := []byte{7}
		if node.Kind == manifest.KindUnion {
			input = append([]byte{0}, input...)
			want = UnionValue{Tag: 0, Value: uint16(513)}
			value = UnionValue{Tag: 0, Value: uint8(7)}
			expected = []byte{0, 7}
		}
		got, err := DecodeAll(node, input)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Errorf("decode = %#v, %v", got, err)
		}
		encoded, err := Encode(node, value)
		if err != nil || !bytes.Equal(encoded, expected) {
			t.Errorf("encode = %x, %v", encoded, err)
		}
	}
}
