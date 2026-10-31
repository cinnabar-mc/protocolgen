package semantics

import (
	"protocolgen/internal/manifest"
	"testing"
)

func TestSemanticsIndexesDecodeShapes(t *testing.T) {
	for _, location := range []string{"packet", "field", "variant"} {
		t.Run(location, func(t *testing.T) {
			m := fixture()
			decoded := manifest.Struct(manifest.Field{Name: "Actor ID", Encode: manifest.Primitive("i64le")})
			decoded.TypeID = "DecodedActor"
			root := manifest.Void()
			switch location {
			case "packet":
				m.Packets[0].Fields[0].Decode = &decoded
			case "field":
				root = manifest.Struct(manifest.Field{Name: "Inner", Encode: manifest.Void(), Decode: &decoded})
			case "variant":
				root = manifest.Union(manifest.Primitive("u8"), manifest.Variant{Value: 0, Name: "Value", Encode: manifest.Void(), Decode: &decoded})
			}
			m.Packets[0].Fields[0].Encode = root
			document := Document{SchemaVersion: 1, Target: m.Target, Entries: []Entry{{TypeID: "DecodedActor", Field: "Actor ID", Semantic: ActorUniqueID}}}
			if err := ValidateOverlay(m, document); err != nil {
				t.Fatal(err)
			}
		})
	}
}
