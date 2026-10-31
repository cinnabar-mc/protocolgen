package layout

import (
	"protocolgen/internal/manifest"
	"testing"
)

func TestLayoutIndexesDecodeShapes(t *testing.T) {
	for _, location := range []string{"field", "variant"} {
		t.Run(location, func(t *testing.T) {
			m := fixture()
			enum := m.Packets[0].Fields[0].Encode
			decoded := manifest.Struct(manifest.Field{Name: "Mode", Encode: enum})
			decoded.TypeID = "Decoded"
			field := &m.Packets[0].Fields[0]
			field.Encode = manifest.Void()
			field.Decode = &decoded
			if location == "variant" {
				field.Decode = nil
				field.Encode = manifest.Union(manifest.Primitive("u8"), manifest.Variant{Value: 0, Name: "Value", Encode: manifest.Void(), Decode: &decoded})
			}
			document := Document{SchemaVersion: 1, Target: m.Target, Constants: []ConstantEntry{{TypeID: "enums/GameType", Package: "protocol", File: "mode", Names: map[string]string{"Creative": "Creative"}}}, Fields: []FieldEntry{{TypeID: "Decoded", Field: "Mode", Name: "GameMode"}}}
			if err := ValidateOverlay(m, document); err != nil {
				t.Fatal(err)
			}
			document.Fields[0].Field = "Missing"
			if err := ValidateOverlay(m, document); err == nil {
				t.Fatal("accepted unknown decode field")
			}
			document.Fields = nil
			document.Constants[0].Names = map[string]string{"Missing": "Missing"}
			if err := ValidateOverlay(m, document); err == nil {
				t.Fatal("accepted unknown decode variant")
			}
		})
	}
}
