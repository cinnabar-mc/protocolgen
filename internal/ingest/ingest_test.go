package ingest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"protocolgen/internal/manifest"
)

func TestMojangIngestionRetainsWireVocabulary(t *testing.T) {
	root := t.TempDir()
	writeJSON(t, filepath.Join(root, "Vocabulary.json"), map[string]any{
		"title": "VocabularyPacket", "$metaProperties": map[string]any{"[cereal:packet]": 1},
		"x-minecraft-version": "1.26.40", "x-protocol-version": 2168,
		"properties": map[string]any{
			"Optional": map[string]any{"type": "integer", "x-underlying-type": "int32", "x-ordinal-index": 0},
			"Double":   map[string]any{"type": "integer", "x-underlying-type": "int32", "x-serialization-options": []string{"+double-optional"}, "x-ordinal-index": 1},
			"Bytes":    map[string]any{"type": "string", "x-underlying-type": "bytearray", "x-ordinal-index": 2},
			"Fixed":    map[string]any{"type": "array", "minItems": 2, "maxItems": 2, "items": map[string]any{"type": "integer", "x-underlying-type": "uint16"}, "x-ordinal-index": 3},
			"Choice": map[string]any{"oneOf": []any{
				map[string]any{"title": "None", "type": "null", "x-ordinal-index": 0},
				map[string]any{"title": "Payload", "type": "string", "x-ordinal-index": 7},
			}, "x-control-value-type": "uint32", "x-ordinal-index": 4},
			"Mode": map[string]any{"type": "integer", "x-underlying-type": "uint8", "enum": []string{"Ready", "Later"}, "x-enum-binary-value": []int{4, 9}, "x-serialization-options": []string{"Enum-as-Value"}, "x-ordinal-index": 5},
		},
		"required": []string{"Bytes", "Fixed", "Choice", "Mode"},
	})
	result, err := ParseMojang(root, fixturePin("mojang"), "")
	if err != nil {
		t.Fatalf("ParseMojang: %v", err)
	}
	if len(result.Claims) != 6 {
		t.Fatalf("claims = %d, want 6", len(result.Claims))
	}
	if got := result.Claims[0].Encode.Kind; got != manifest.KindOptional {
		t.Errorf("optional kind = %q, want optional", got)
	}
	if got := result.Claims[1].Encode.Kind; got != manifest.KindOptional || result.Claims[1].Encode.Value.Kind != manifest.KindOptional {
		t.Errorf("double optional = %+v", result.Claims[1].Encode)
	}
	if got := result.Claims[2].Encode.Kind; got != manifest.KindBytes {
		t.Errorf("bytes kind = %q, want bytes", got)
	}
	if got := result.Claims[3].Encode.Kind; got != manifest.KindFixedArray || result.Claims[3].Encode.Length != 2 {
		t.Errorf("fixed array = %+v", result.Claims[3].Encode)
	}
	if got := result.Claims[4].Encode.Kind; got != manifest.KindUnion || result.Claims[4].Encode.Variants[1].Value != 1 {
		t.Errorf("union = %+v", result.Claims[4].Encode)
	}
	if got := result.Claims[5].Encode.Kind; got != manifest.KindEnum || result.Claims[5].Encode.Variants[0].Value != 4 {
		t.Errorf("enum = %+v", result.Claims[5].Encode)
	}
}

func TestMojangIngestionRetainsValidationConstraints(t *testing.T) {
	lowerer := &mojangLowerer{documents: map[string]any{}, active: map[string]bool{}}
	tests := []struct {
		name   string
		schema map[string]any
		want   manifest.Constraints
	}{
		{"string", map[string]any{"type": "string", "minLength": 1, "maxLength": 65536, "pattern": "^[a-z]+$"}, manifest.Constraints{MinLength: pointerTo(uint64(1)), MaxLength: pointerTo(uint64(65536)), Pattern: "^[a-z]+$"}},
		{"array", map[string]any{"type": "array", "items": map[string]any{"type": "integer", "x-underlying-type": "uint8"}, "minItems": 1, "maxItems": 4}, manifest.Constraints{MinItems: pointerTo(uint64(1)), MaxItems: pointerTo(uint64(4))}},
		{"number", map[string]any{"type": "integer", "x-underlying-type": "int32", "minimum": -1.0, "maximum": 64.0}, manifest.Constraints{Minimum: pointerTo(-1.0), Maximum: pointerTo(64.0)}},
		{"map", map[string]any{"type": "object", "additionalProperties": map[string]any{"type": "integer", "x-underlying-type": "uint8"}, "minProperties": 2, "maxProperties": 8}, manifest.Constraints{MinProperties: pointerTo(uint64(2)), MaxProperties: pointerTo(uint64(8))}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			node := lowerer.lowerSchema(test.schema, "Packet.json", "Value")
			if node.Constraints == nil || !reflect.DeepEqual(*node.Constraints, test.want) {
				t.Fatalf("constraints = %#v, want %#v", node.Constraints, test.want)
			}
		})
	}
}

// Both sources put prose in "pattern"; it must not become a regex constraint.
func TestIngestionDropsProsePatterns(t *testing.T) {
	lowerer := &mojangLowerer{documents: map[string]any{}, active: map[string]bool{}}
	node := lowerer.lowerSchema(map[string]any{"type": "string", "pattern": "A string in the format of <uuid>_<semver>"}, "Packet.json", "Value")
	if node.Constraints != nil {
		t.Fatalf("prose pattern became a constraint: %#v", node.Constraints)
	}
	node = withEndstoneConstraintMap(manifest.String(manifest.Primitive("var_u32")), map[string]any{"pattern": "A string in the format of <uuid>_<semver>"})
	if node.Constraints != nil {
		t.Fatalf("Endstone prose pattern became a constraint: %#v", node.Constraints)
	}
}

func TestMojangArrayIgnoresObjectPropertyBounds(t *testing.T) {
	lowerer := &mojangLowerer{documents: map[string]any{}, active: map[string]bool{}}
	node := lowerer.lowerSchema(map[string]any{
		"type": "array", "items": map[string]any{"type": "string"},
		"maxProperties": 65535, "maxItems": 100,
	}, "Resource_Pack_Client_Response_-_Downloading.json", "Downloading Packs")
	if node.Kind != manifest.KindArray || node.Constraints == nil || node.Constraints.MaxProperties != nil || node.Constraints.MaxItems == nil || *node.Constraints.MaxItems != 100 {
		t.Fatalf("array = %#v, want only the valid item bound", node)
	}
}

func pointerTo[T any](value T) *T { return &value }

func TestMojangPacketRootReferenceUsesPayloadFields(t *testing.T) {
	root := t.TempDir()
	writeJSON(t, filepath.Join(root, "LoginPacket.json"), map[string]any{
		"title": "LoginPacket", "$metaProperties": map[string]any{"[cereal:packet]": 1},
		"x-minecraft-version": "1.26.40", "x-protocol-version": 2168,
		"$ref": "./LoginPacketPayload.json",
	})
	writeJSON(t, filepath.Join(root, "LoginPacketPayload.json"), map[string]any{
		"title": "LoginPacketPayload", "type": "object",
		"x-minecraft-version": "1.26.40", "x-protocol-version": 2168,
		"properties": map[string]any{
			"Client Network Version": map[string]any{"type": "integer", "x-underlying-type": "int32", "x-serialization-options": []string{"Big Endian"}, "x-ordinal-index": 0},
			"Connection Request":     map[string]any{"type": "string", "x-ordinal-index": 1},
		},
		"required": []string{"Client Network Version", "Connection Request"},
	})

	result, err := ParseMojang(root, fixturePin("mojang"), "")
	if err != nil {
		t.Fatalf("ParseMojang: %v", err)
	}
	if len(result.Claims) != 2 {
		t.Fatalf("claims = %d, want 2", len(result.Claims))
	}
	if got := result.Claims[0].Encode.Primitive.Code; got != "i32be" {
		t.Fatalf("client network version codec = %q, want i32be", got)
	}
}

func TestMojangPacketRootReferenceAllowsEmptyObject(t *testing.T) {
	root := t.TempDir()
	writeJSON(t, filepath.Join(root, "HandshakePacket.json"), map[string]any{
		"title": "HandshakePacket", "$metaProperties": map[string]any{"[cereal:packet]": 4},
		"x-minecraft-version": "1.26.40", "x-protocol-version": 2168,
		"$ref": "./HandshakePacketPayload.json",
	})
	writeJSON(t, filepath.Join(root, "HandshakePacketPayload.json"), map[string]any{
		"title": "HandshakePacketPayload", "type": "object",
		"x-minecraft-version": "1.26.40", "x-protocol-version": 2168,
	})

	result, err := ParseMojang(root, fixturePin("mojang"), "")
	if err != nil {
		t.Fatalf("ParseMojang: %v", err)
	}
	if len(result.Claims) != 0 {
		t.Fatalf("claims = %d, want 0", len(result.Claims))
	}
}

func TestMojangBareSelfReferenceIsUnresolved(t *testing.T) {
	root := t.TempDir()
	writeJSON(t, filepath.Join(root, "HandshakePacket.json"), map[string]any{
		"title": "HandshakePacket", "$metaProperties": map[string]any{"[cereal:packet]": 3},
		"x-minecraft-version": "1.26.40", "x-protocol-version": 2168,
		"properties": map[string]any{
			"Token": map[string]any{"$ref": "./WebToken.json", "x-ordinal-index": 0},
		},
		"required": []string{"Token"},
	})
	writeJSON(t, filepath.Join(root, "WebToken.json"), map[string]any{
		"title": "WebToken", "$ref": "./WebToken.json",
		"x-minecraft-version": "1.26.40", "x-protocol-version": 2168,
	})

	result, err := ParseMojang(root, fixturePin("mojang"), "")
	if err != nil {
		t.Fatalf("ParseMojang: %v", err)
	}
	if got := result.Claims[0].Encode.Kind; got != manifest.KindUnresolved {
		t.Fatalf("self reference kind = %q, want unresolved", got)
	}
}

func TestMojangUnionUsesPositionalVaruint32WithoutCompressionOrOrdinals(t *testing.T) {
	lowerer := &mojangLowerer{documents: map[string]any{}, active: map[string]bool{}}
	node := lowerer.lowerSchema(map[string]any{
		"x-control-value-type": "uint32",
		"oneOf": []any{
			map[string]any{"title": "Payload", "type": "string"},
			map[string]any{"title": "None", "type": "null"},
		},
	}, "Packet.json", "PacketChoice")
	if node.Kind != manifest.KindUnion || len(node.Variants) != 2 || node.Variants[0].Value != 0 || node.Variants[1].Value != 1 {
		t.Fatalf("union = %#v, want positional selectors 0 and 1", node)
	}
	if node.Control.Primitive == nil || node.Control.Primitive.Code != "var_u32" {
		t.Fatalf("control = %#v, want varuint32", node.Control)
	}
}

func TestMojangUnionKeepsInnerTagSeparateFromSelector(t *testing.T) {
	lowerer := &mojangLowerer{documents: map[string]any{}, active: map[string]bool{}}
	branches := []any{}
	for _, value := range []int{1, 0} {
		branches = append(branches, map[string]any{
			"type": "object",
			"properties": map[string]any{
				"Action": map[string]any{
					"type": "string", "enum": []any{"Action"},
					"x-enum-binary-value": []any{value}, "x-underlying-type": "uint8",
					"x-serialization-options": []any{"Enum-as-Value"}, "x-ordinal-index": 0,
				},
			},
			"required": []any{"Action"},
		})
	}
	node := lowerer.lowerSchema(map[string]any{
		"oneOf": branches, "x-control-value-type": "uint32",
	}, "Packet.json", "Entries")
	if node.Kind != manifest.KindUnion || node.Control.Primitive == nil || node.Control.Primitive.Code != "var_u32" || len(node.Variants) != 2 {
		t.Fatalf("union = %#v, want two alternatives with a varuint32 selector", node)
	}
	for i, variant := range node.Variants {
		if variant.Value != int64(i) || variant.Encode.Kind != manifest.KindStruct || len(variant.Encode.Fields) != 1 {
			t.Fatalf("variant %d = %#v, want its index followed by the Action field", i, variant)
		}
		action := variant.Encode.Fields[0].Encode
		if action.Kind != manifest.KindEnum || action.Primitive.Code != "u8" || len(action.Variants) != 1 || action.Variants[0].Value != int64(1-i) {
			t.Fatalf("Action %d = %#v, want uint8 value %d", i, action, 1-i)
		}
	}
}

func TestMojangUnionRejectsMissingOrMalformedAlternatives(t *testing.T) {
	for _, branches := range [][]any{{}, {"invalid"}} {
		lowerer := &mojangLowerer{documents: map[string]any{}, active: map[string]bool{}}
		node := lowerer.lowerSchema(map[string]any{"oneOf": branches, "x-control-value-type": "uint32"}, "Packet.json", "Choice")
		if node.Kind != manifest.KindUnresolved {
			t.Fatalf("union = %#v, want unresolved malformed alternatives", node)
		}
	}
}

// TestMojangJSONColourAlternativesRemainUnresolved uses the published Color255RGB shape.
func TestMojangJSONColourAlternativesRemainUnresolved(t *testing.T) {
	lowerer := &mojangLowerer{documents: map[string]any{}, active: map[string]bool{}}
	node := lowerer.lowerSchema(map[string]any{
		"title": "Color255RGB",
		"oneOf": []any{
			map[string]any{"type": "string", "pattern": "^#[a-fA-F0-9]{6}$"},
			map[string]any{"type": "array", "items": map[string]any{"type": "integer", "x-underlying-type": "int32"}, "minItems": 3, "maxItems": 3},
		},
	}, "Color255RGB.json", "Color255RGB")
	if node.Kind != manifest.KindUnresolved {
		t.Fatalf("colour = %#v, want unresolved JSON alternatives", node)
	}
}

func TestMojangUUIDReferenceUsesCanonicalPrimitive(t *testing.T) {
	lowerer := &mojangLowerer{documents: map[string]any{
		"mce__UUID.json": map[string]any{
			"title": "mce::UUID", "type": "object",
			"properties": map[string]any{
				"Most Significant Bits":  map[string]any{"type": "integer", "x-underlying-type": "uint64", "x-ordinal-index": 0},
				"Least Significant Bits": map[string]any{"type": "integer", "x-underlying-type": "uint64", "x-ordinal-index": 1},
			},
		},
	}, active: map[string]bool{}}
	node := lowerer.lowerSchema(map[string]any{"$ref": "./mce__UUID.json"}, "Packet.json", "PacketUUID")
	if node.Kind != manifest.KindPrimitive || node.Primitive == nil || node.Primitive.Code != "uuid" {
		t.Fatalf("UUID reference = %#v, want canonical uuid primitive", node)
	}
}

func TestEndstoneIngestionRetainsMapsSwitchesAndFixedArrays(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "packets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "types"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "enums"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("**Minecraft Version:** `1.26.40.1`\n**Network Version:** `2168`\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeJSON(t, filepath.Join(root, "enums", "Mode.json"), map[string]any{"name": "Mode", "values": []any{map[string]any{"name": "None", "value": 0}, map[string]any{"name": "Payload", "value": 7}}})
	writeJSON(t, filepath.Join(root, "types", "Entry.json"), map[string]any{"name": "Entry", "fields": []any{map[string]any{"name": "Value", "type": "uint8"}}})
	writeJSON(t, filepath.Join(root, "packets", "Vocabulary.json"), map[string]any{"name": "Vocabulary", "id": 1, "fields": []any{
		map[string]any{"name": "Optional", "type": "int32", "optional": true},
		map[string]any{"name": "Fixed", "type": map[string]any{"type": "uint16", "repeat": map[string]any{"count": 2}}},
		map[string]any{"name": "Choice", "type": map[string]any{"switch": map[string]any{"type": "uint8", "enum": "Mode"}, "cases": []any{nil, "Entry"}}},
		map[string]any{"name": "Text", "type": map[string]any{"type": "string", "encoding": "utf8"}},
		map[string]any{"name": "Bytes", "type": "restBuffer"},
		map[string]any{"name": "Entries", "type": map[string]any{"key": "string", "value": "uint8"}},
	}})
	result, err := ParseEndstone(root, fixturePin("endstone"), "")
	if err != nil {
		t.Fatalf("ParseEndstone: %v", err)
	}
	if len(result.Claims) != 6 {
		t.Fatalf("claims = %d, want 6", len(result.Claims))
	}
	if result.Claims[1].Encode.Kind != manifest.KindFixedArray || result.Claims[2].Encode.Kind != manifest.KindUnion || result.Claims[4].Encode.Kind != manifest.KindBytes || result.Claims[5].Encode.Kind != manifest.KindMap {
		t.Fatalf("unexpected Endstone shapes: %#v", result.Claims)
	}
}

func TestEndstonePacketFieldAppliesRepeatPrefix(t *testing.T) {
	root := t.TempDir()
	for _, directory := range []string{"packets", "types", "enums"} {
		if err := os.Mkdir(filepath.Join(root, directory), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("**Minecraft Version:** `1.26.40.1`\n**Network Version:** `2168`\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeJSON(t, filepath.Join(root, "packets", "ListPacket.json"), map[string]any{
		"name": "ListPacket", "id": 1,
		"fields": []any{map[string]any{"name": "Items", "type": "uint16", "repeat": map[string]any{"prefix": "uvarint32"}}},
	})

	result, err := ParseEndstone(root, fixturePin("endstone"), "")
	if err != nil {
		t.Fatalf("ParseEndstone: %v", err)
	}
	node := result.Claims[0].Encode
	if node.Kind != manifest.KindArray || node.Prefix == nil || node.Prefix.Primitive == nil || node.Prefix.Primitive.Code != "var_u32" {
		t.Fatalf("field = %#v, want var_u32-prefixed array", node)
	}
}

func TestEndstoneFieldEnumUsesItsCerealValueConstraints(t *testing.T) {
	lowerer := &endstoneLowerer{
		types: map[string]any{},
		enums: map[string]any{
			"GameType.json": map[string]any{"name": "GameType", "values": []any{
				map[string]any{"name": "Survival", "value": 0},
				map[string]any{"name": "Creative", "value": 1},
				map[string]any{"name": "WorldDefault", "value": 0},
			}},
		},
		active: map[string]bool{},
	}
	field := map[string]any{
		"name": "Game Type", "type": "varint32", "enum": "GameType",
		"constraints": map[string]any{"enum_values": []any{0, 1}},
	}
	node := lowerer.applyFieldWrappers(endstoneScalar("varint32"), field, "StartGamePacket.Game Type")
	if node.Kind != manifest.KindEnum || len(node.Variants) != 2 {
		t.Fatalf("enum = %#v, want two constrained values", node)
	}
	if node.Variants[0].Name != "Survival" || node.Variants[1].Name != "Creative" {
		t.Fatalf("variants = %#v, want first canonical names for values 0 and 1", node.Variants)
	}
}

func TestEndstoneFieldRetainsValidationConstraints(t *testing.T) {
	lowerer := &endstoneLowerer{types: map[string]any{}, enums: map[string]any{}, active: map[string]bool{}}
	node := lowerer.applyFieldWrappers(endstoneScalar("string"), map[string]any{
		"name": "Message", "type": "string",
		"constraints": map[string]any{"min_length": 1, "max_length": 65536, "pattern": "^[a-z]+$"},
	}, "TextPacket.Message")
	if node.Constraints == nil || node.Constraints.MinLength == nil || *node.Constraints.MinLength != 1 || node.Constraints.MaxLength == nil || *node.Constraints.MaxLength != 65536 || node.Constraints.Pattern != "^[a-z]+$" {
		t.Fatalf("constraints = %#v, want string bounds", node.Constraints)
	}
}

func TestEndstoneFieldRetainsNestedValidationConstraints(t *testing.T) {
	array := withEndstoneConstraints(manifest.Array(manifest.Primitive("var_u32"), manifest.String(manifest.Primitive("var_u32"))), map[string]any{
		"constraints": map[string]any{"max_properties": 65535, "items": map[string]any{"max_length": 128}},
	})
	if array.Constraints == nil || array.Constraints.MaxItems == nil || *array.Constraints.MaxItems != 65535 || array.Element == nil || array.Element.Constraints == nil || array.Element.Constraints.MaxLength == nil || *array.Element.Constraints.MaxLength != 128 {
		t.Fatalf("array constraints = %#v, element = %#v", array.Constraints, array.Element)
	}

	mapNode := withEndstoneConstraints(manifest.Map(manifest.Primitive("var_u32"), manifest.String(manifest.Primitive("var_u32")), manifest.Primitive("i32le")), map[string]any{
		"constraints": map[string]any{
			"max_properties":        255,
			"property_names":        map[string]any{"max_length": 64},
			"additional_properties": map[string]any{"minimum": 0, "maximum": 100},
		},
	})
	if mapNode.Constraints == nil || mapNode.Constraints.MaxProperties == nil || *mapNode.Constraints.MaxProperties != 255 || mapNode.Key == nil || mapNode.Key.Constraints == nil || mapNode.Key.Constraints.MaxLength == nil || *mapNode.Key.Constraints.MaxLength != 64 || mapNode.Value == nil || mapNode.Value.Constraints == nil || mapNode.Value.Constraints.Minimum == nil || *mapNode.Value.Constraints.Minimum != 0 || mapNode.Value.Constraints.Maximum == nil || *mapNode.Value.Constraints.Maximum != 100 {
		t.Fatalf("map constraints = %#v, key = %#v, value = %#v", mapNode.Constraints, mapNode.Key, mapNode.Value)
	}
}

func TestEndstoneSwitchUsesVariantCerealSelectorConstraints(t *testing.T) {
	lowerer := &endstoneLowerer{
		types: map[string]any{
			"MessageOnly.json": map[string]any{"name": "MessageOnly", "fields": []any{
				map[string]any{"name": "Message Type", "type": "uint8", "enum": "MessageType", "constraints": map[string]any{"enum_values": []any{0, 2}}},
				map[string]any{"name": "Message", "type": "string"},
			}},
		},
		enums: map[string]any{
			"MessageType.json": map[string]any{"name": "MessageType", "values": []any{
				map[string]any{"name": "Raw", "value": 0}, map[string]any{"name": "Chat", "value": 1}, map[string]any{"name": "System", "value": 2},
			}},
		},
		active: map[string]bool{},
	}
	node := lowerer.lowerTypeValue(map[string]any{
		"switch": map[string]any{"name": "Message Type", "type": "uint8", "enum": "MessageType"},
		"cases":  []any{"MessageOnly"},
	}, "PacketBody", "Packet.Body")
	if node.Kind != manifest.KindUnion || len(node.Variants) != 2 || node.Variants[0].Value != 0 || node.Variants[1].Value != 2 {
		t.Fatalf("union = %#v, want selectors 0 and 2", node)
	}
	for _, variant := range node.Variants {
		if variant.Encode.Kind != manifest.KindStruct || len(variant.Encode.Fields) != 1 || variant.Encode.Fields[0].Name != "Message" {
			t.Fatalf("variant payload = %#v, want discriminator removed", variant.Encode)
		}
	}
}

func TestEndstoneSwitchRetainsDifferentlyEncodedCompatibilityDiscriminator(t *testing.T) {
	lowerer := &endstoneLowerer{
		types: map[string]any{"Coordinates.json": map[string]any{"name": "Coordinates", "fields": []any{
			map[string]any{"name": "Packet Type", "type": "varint32", "constraints": map[string]any{"enum_values": []any{0}}},
			map[string]any{"name": "X", "type": "float32"},
		}}},
		enums:  map[string]any{},
		active: map[string]bool{},
	}
	node := lowerer.lowerTypeValue(map[string]any{
		"switch": map[string]any{"name": "Packet Type", "type": "uvarint32"},
		"cases":  []any{"Coordinates"},
	}, "Location", "PlayerLocationPacket.Location")
	if node.Kind != manifest.KindUnion || len(node.Variants) != 1 || len(node.Variants[0].Encode.Fields) != 2 {
		t.Fatalf("union = %#v, want compatibility discriminator retained before payload", node)
	}
}

func TestEndstoneSwitchUsesCaseOrderWithoutNamedEnum(t *testing.T) {
	lowerer := &endstoneLowerer{
		types: map[string]any{
			"Stop.json":      map[string]any{"name": "Stop", "fields": []any{}},
			"SetVolume.json": map[string]any{"name": "SetVolume", "fields": []any{map[string]any{"name": "Volume", "type": "float32"}}},
		},
		enums:  map[string]any{},
		active: map[string]bool{},
	}
	node := lowerer.lowerTypeValue(map[string]any{
		"switch": map[string]any{"type": "uvarint32"},
		"cases":  []any{"Stop", "SetVolume"},
	}, "SoundEvent", "Packet.Event")
	if node.Kind != manifest.KindUnion || len(node.Variants) != 2 {
		t.Fatalf("union = %#v, want two variants", node)
	}
	if node.Variants[0].Value != 0 || node.Variants[1].Value != 1 {
		t.Fatalf("selectors = [%d, %d], want [0, 1]", node.Variants[0].Value, node.Variants[1].Value)
	}
}

func TestEndstoneRetainsRepeatedCerealUnionFields(t *testing.T) {
	root := t.TempDir()
	for _, directory := range []string{"packets", "types", "enums"} {
		if err := os.Mkdir(filepath.Join(root, directory), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("**Minecraft Version:** `1.26.40.1`\n**Network Version:** `2168`\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	unionType := map[string]any{
		"switch": map[string]any{"type": "uvarint32"},
		"cases":  []any{"SoundDataEvent::Stop", "SoundDataEvent::SetVolume"},
	}
	writeJSON(t, filepath.Join(root, "packets", "SoundPacket.json"), map[string]any{
		"name": "SoundPacket", "id": 7,
		"fields": []any{
			map[string]any{"name": "Handle", "type": "uint64"},
			map[string]any{"name": "Stop", "type": unionType},
			map[string]any{"name": "SetVolume", "type": unionType},
			map[string]any{"name": "Tail", "type": "bool"},
		},
	})
	writeJSON(t, filepath.Join(root, "types", "Stop.json"), map[string]any{"name": "SoundDataEvent::Stop", "fields": []any{}})
	writeJSON(t, filepath.Join(root, "types", "SetVolume.json"), map[string]any{"name": "SoundDataEvent::SetVolume", "fields": []any{map[string]any{"name": "Volume", "type": "float32"}}})

	result, err := ParseEndstone(root, fixturePin("endstone"), "")
	if err != nil {
		t.Fatalf("ParseEndstone: %v", err)
	}
	if len(result.Claims) != 4 {
		t.Fatalf("claims = %d, want handle, both union slots, and tail", len(result.Claims))
	}
	if result.Claims[1].Encode.Kind != manifest.KindUnion {
		t.Fatalf("first union field = %#v, want union", result.Claims[1].Encode)
	}
	if result.Claims[1].Name != "Stop" || result.Claims[2].Name != "SetVolume" {
		t.Fatalf("union field names = %q, %q, want Stop and SetVolume", result.Claims[1].Name, result.Claims[2].Name)
	}
	if result.Claims[3].Name != "Tail" || result.Claims[3].Ordinal != 3 {
		t.Fatalf("tail = %#v, want source ordinal 3", result.Claims[3])
	}
}

func TestEndstoneRecordsEmptyPackets(t *testing.T) {
	root := t.TempDir()
	for _, directory := range []string{"packets", "types", "enums"} {
		if err := os.Mkdir(filepath.Join(root, directory), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("**Minecraft Version:** `1.26.40.1`\n**Network Version:** `2168`\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeJSON(t, filepath.Join(root, "packets", "EmptyPacket.json"), map[string]any{"name": "EmptyPacket", "id": 4, "fields": []any{}})

	result, err := ParseEndstone(root, fixturePin("endstone"), "")
	if err != nil {
		t.Fatalf("ParseEndstone: %v", err)
	}
	if len(result.Packets) != 1 || result.Packets[0].ID != 4 || result.Packets[0].Name != "EmptyPacket" {
		t.Fatalf("packet claims = %#v, want empty packet metadata", result.Packets)
	}
}

func TestCorrectionRequiresPrePatchFingerprintAndEvidence(t *testing.T) {
	documents := map[string]any{"Packet.json": map[string]any{"value": "string"}}
	before, err := canonicalDigest(documents["Packet.json"].(map[string]any)["value"])
	if err != nil {
		t.Fatal(err)
	}
	corrections := t.TempDir()
	writeJSON(t, filepath.Join(corrections, "bytes.json"), map[string]any{"operations": []any{map[string]any{
		"id": "bytes-field", "file": "Packet.json", "pointer": "/value", "pre_patch_sha256": before, "replace": "bytearray", "why": "fixture byte evidence", "evidence": []any{map[string]any{"locator": "fixture/runtime"}},
	}}})
	proofs, err := applyCorrections(documents, corrections, fixturePin("mojang"))
	if err != nil {
		t.Fatalf("applyCorrections: %v", err)
	}
	if len(proofs) != 1 || proofs[0].PostPatchNodeSHA256 == before {
		t.Fatalf("proofs = %#v", proofs)
	}
	if proofs[0].PrePatchContextSHA256 == "" {
		t.Fatal("correction proof has no complete pre-patch context fingerprint")
	}
	documents["Packet.json"].(map[string]any)["value"] = "mutated"
	if _, err := applyCorrections(documents, corrections, fixturePin("mojang")); err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("stale correction error = %v", err)
	}
}

func TestCorrectionCanAddMissingObjectMetadataWithoutOverwriting(t *testing.T) {
	documents := map[string]any{"Packet.json": map[string]any{"field": map[string]any{"name": "Value", "type": "uint8"}}}
	before, err := canonicalDigest(documents["Packet.json"].(map[string]any)["field"])
	if err != nil {
		t.Fatal(err)
	}
	corrections := t.TempDir()
	writeJSON(t, filepath.Join(corrections, "optional.json"), map[string]any{"operations": []any{map[string]any{
		"id": "optional-field", "file": "Packet.json", "pointer": "/field", "pre_patch_sha256": before,
		"merge": map[string]any{"optional": true}, "why": "fixture optional evidence", "evidence": []any{map[string]any{"locator": "fixture/runtime"}},
	}}})
	if _, err := applyCorrections(documents, corrections, fixturePin("endstone")); err != nil {
		t.Fatalf("applyCorrections: %v", err)
	}
	field := documents["Packet.json"].(map[string]any)["field"].(map[string]any)
	if field["optional"] != true {
		t.Fatalf("field = %#v, want optional metadata merged", field)
	}
}

func TestMojangExplicitWireKindNBTLowersToNetworkTag(t *testing.T) {
	lowerer := &mojangLowerer{documents: map[string]any{}, active: map[string]bool{}, source: fixturePin("mojang")}
	node := lowerer.lowerSchema(map[string]any{"type": "object", "x-wire-kind": "nbt"}, "fixture.json", "Payload")
	if node.Kind != manifest.KindPrimitive || node.Primitive == nil || node.Primitive.Code != "nbt_le" {
		t.Fatalf("node = %#v, want nbt_le primitive", node)
	}
}

func TestMojangRejectsMixedProtocolDocuments(t *testing.T) {
	root := t.TempDir()
	writeJSON(t, filepath.Join(root, "A.json"), map[string]any{
		"title": "A", "$metaProperties": map[string]any{"[cereal:packet]": 1},
		"x-minecraft-version": "1.26.40", "x-protocol-version": 2168,
		"properties": map[string]any{"Value": map[string]any{"type": "integer", "x-underlying-type": "uint8", "x-ordinal-index": 0}},
		"required":   []string{"Value"},
	})
	writeJSON(t, filepath.Join(root, "B.json"), map[string]any{
		"title": "B", "$metaProperties": map[string]any{"[cereal:packet]": 2},
		"x-minecraft-version": "1.26.50", "x-protocol-version": 2169,
		"properties": map[string]any{"Value": map[string]any{"type": "integer", "x-underlying-type": "uint8", "x-ordinal-index": 0}},
		"required":   []string{"Value"},
	})
	if _, err := ParseMojang(root, fixturePin("mojang"), ""); err == nil || !strings.Contains(err.Error(), "mixes") {
		t.Fatalf("ParseMojang error = %v, want mixed-version rejection", err)
	}
}

func TestMojangMissingOrdinalDiagnosticIsDeterministic(t *testing.T) {
	properties := map[string]any{
		"Value": map[string]any{"type": "integer"},
		"Key":   map[string]any{"type": "string"},
	}
	_, err := lowerMojangFields(nil, 0, "MapEntry", manifest.DirectionUnknown, "MapEntry.json", properties, nil, "")
	if err == nil || !strings.Contains(err.Error(), "property Key has no explicit ordinal") {
		t.Fatalf("lowerMojangFields error = %v, want the first lexical property", err)
	}
}

func TestMojangMapEntrySchemasLowerKeyAndValue(t *testing.T) {
	lowerer := &mojangLowerer{documents: map[string]any{}, active: map[string]bool{}}
	node := lowerer.lowerSchema(map[string]any{
		"type": "object",
		"additionalProperties": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"key":   map[string]any{"type": "integer", "x-underlying-type": "uint16"},
				"value": map[string]any{"type": "number", "x-underlying-type": "float"},
			},
		},
	}, "Map.json", "Fixture")
	if node.Kind != manifest.KindMap || node.Key == nil || node.Value == nil {
		t.Fatalf("node = %#v, want map with key and value", node)
	}
	if node.Key.Primitive == nil || node.Key.Primitive.Code != "u16le" || node.Value.Primitive == nil || node.Value.Primitive.Code != "f32le" {
		t.Fatalf("map key/value = %#v/%#v, want u16le/f32le", node.Key, node.Value)
	}
}

func TestEndstoneSpecialUnknownsStayReachableUnresolved(t *testing.T) {
	lowerer := &endstoneLowerer{types: map[string]any{}, enums: map[string]any{}, active: map[string]bool{}}
	for _, name := range []string{"cereal::UnknownBuiltin"} {
		node := lowerer.lowerNamed(name, "fixture")
		if node.Kind != manifest.KindUnresolved || !node.Reachable {
			t.Fatalf("%s lowered to %#v, want reachable unresolved", name, node)
		}
		if node.Element != nil || node.Prefix != nil {
			t.Fatalf("%s invented a byte shape: %#v", name, node)
		}
	}
}

func TestUUIDUsesCanonicalPrimitive(t *testing.T) {
	node := primitive("uuid", nil, "")
	if node.Kind != manifest.KindPrimitive || node.Primitive == nil || node.Primitive.Code != "uuid" {
		t.Fatalf("uuid node = %#v, want canonical uuid primitive", node)
	}
}

func TestEndstoneScalarSpellingsUseCanonicalPrimitives(t *testing.T) {
	tests := map[string]string{
		"int32_be":    "i32be",
		"varint32":    "zigzag_i32",
		"varint64":    "zigzag_i64",
		"uvarint32":   "var_u32",
		"uvarint64":   "var_u64",
		"mce::UUID":   "uuid",
		"CompoundTag": "nbt_le",
	}
	for input, want := range tests {
		t.Run(input, func(t *testing.T) {
			node := endstoneScalar(input)
			if node.Kind != manifest.KindPrimitive || node.Primitive == nil || node.Primitive.Code != want {
				t.Fatalf("endstoneScalar(%q) = %#v, want %s", input, node, want)
			}
		})
	}
}

func TestEndstoneFixedWidthBitsetRetainsItsCerealBitCount(t *testing.T) {
	node := endstoneScalar("brstd::bitset<131>")
	if node.Kind != manifest.NodeKind("bitset") || node.Length != 131 {
		t.Fatalf("bitset = %#v, want 131-bit wire shape", node)
	}
}

func TestEndstoneAlwaysPresentMarkersFoldIntoOptionals(t *testing.T) {
	fields, err := canonicalEndstoneFields([]any{
		map[string]any{"type": "bool", "value": true},
		map[string]any{"name": "Value", "type": "uint32", "optional": true},
	}, "Fixture")
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 1 || !fields[0].outerOptional || fields[0].sourceIndex != 1 {
		t.Fatalf("canonical fields = %#v, want one outer optional sourced from field 1", fields)
	}
	lowerer := &endstoneLowerer{}
	node := lowerer.lowerTypeValue(fields[0].object["type"], "Value", "Fixture.Value")
	node = lowerer.applyFieldWrappers(node, fields[0].object, "Fixture.Value")
	if fields[0].outerOptional {
		node = manifest.Optional(node)
	}
	if node.Kind != manifest.KindOptional || node.Value == nil || node.Value.Kind != manifest.KindOptional {
		t.Fatalf("node = %#v, want double optional", node)
	}
}

func TestEndstoneCerealDynamicValueLowersToRecursiveTaggedValue(t *testing.T) {
	lowerer := &endstoneLowerer{types: map[string]any{}, enums: map[string]any{}, active: map[string]bool{}}
	node := lowerer.lowerNamed("cereal::DynamicValue", "fixture")
	if node.Kind != manifest.KindUnion || node.TypeID != "cereal::DynamicValue" || len(node.Variants) != 7 {
		t.Fatalf("dynamic value = %#v, want seven-way recursive union", node)
	}
	if node.Control == nil || node.Control.Primitive == nil || node.Control.Primitive.Code != "i32le" {
		t.Fatalf("dynamic value control = %#v, want fixed i32", node.Control)
	}
	list := node.Variants[5].Encode
	if list.Kind != manifest.KindArray || list.Element == nil || list.Element.Kind != manifest.KindRecursive || list.Element.Target != "cereal::DynamicValue" {
		t.Fatalf("dynamic list = %#v, want recursive value array", list)
	}
	mapping := node.Variants[6].Encode
	if mapping.Kind != manifest.KindMap || mapping.Value == nil || mapping.Value.Kind != manifest.KindRecursive || mapping.Value.Target != "cereal::DynamicValue" {
		t.Fatalf("dynamic map = %#v, want recursive string/value map", mapping)
	}
}

func TestEndstoneNamedDocumentUsesDeclaredName(t *testing.T) {
	document := map[string]any{"name": "Connection::DisconnectFailReason", "values": []any{}}
	got, ok := namedDocument(map[string]any{"Connection__DisconnectFailReason.json": document}, "Connection::DisconnectFailReason")
	if !ok || got == nil {
		t.Fatal("namedDocument did not match the document's declared C++ name")
	}
}

func fixturePin(id string) manifest.SourcePin {
	return manifest.SourcePin{ID: id, Kind: id, Revision: "fixture-2168", Digest: "fixture:" + id, MinecraftVersion: "1.26.40", ProtocolVersion: 2168}
}

func writeJSON(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestMojangDefaultedPropertyIsNotOptional(t *testing.T) {
	root := t.TempDir()
	writeJSON(t, filepath.Join(root, "TextData.json"), map[string]any{
		"title": "TextDataPacket", "$metaProperties": map[string]any{"[cereal:packet]": 1},
		"x-minecraft-version": "1.26.40", "x-protocol-version": 2168,
		"properties": map[string]any{
			"Text":            map[string]any{"type": "string", "x-ordinal-index": 0},
			"UseRotation":     map[string]any{"type": "boolean", "default": false, "x-ordinal-index": 1},
			"BackgroundColor": map[string]any{"type": "integer", "x-underlying-type": "uint32", "x-ordinal-index": 2},
			"LineGapHeight":   map[string]any{"type": "number", "x-underlying-type": "float", "default": 0.0, "x-ordinal-index": 3},
		},
		"required": []string{"Text"},
	})
	result, err := ParseMojang(root, fixturePin("mojang"), "")
	if err != nil {
		t.Fatalf("ParseMojang: %v", err)
	}
	for _, claim := range result.Claims {
		optional := claim.Encode.Kind == manifest.KindOptional
		want := claim.Name == "BackgroundColor"
		if optional != want {
			t.Errorf("%s optional = %v, want %v", claim.Name, optional, want)
		}
	}
}

// TestMojangBinaryEnumUses preserves the U6 mapping for numeric uses and text encoding otherwise.
func TestMojangBinaryEnumUses(t *testing.T) {
	root := t.TempDir()
	writeJSON(t, filepath.Join(root, "Mode.json"), map[string]any{
		"title": "Mode", "type": "string", "x-underlying-type": "int32",
		"enum":                []string{"Undefined", "Survival", "Creative", "Default", "WorldDefault"},
		"x-enum-binary-value": []int{-1, 0, 1, 5, 0},
	})
	writeJSON(t, filepath.Join(root, "Packet.json"), map[string]any{
		"title": "ModesPacket", "$metaProperties": map[string]any{"[cereal:packet]": 1},
		"x-minecraft-version": "fixture", "x-protocol-version": 2208,
		"properties": map[string]any{
			"Numeric": map[string]any{"$ref": "./Mode.json", "x-underlying-type": "int32", "x-serialization-options": []string{"Enum-as-Value", "Compression"}, "x-ordinal-index": 0},
			"Text":    map[string]any{"$ref": "./Mode.json", "x-ordinal-index": 1},
		},
		"required": []string{"Numeric", "Text"},
	})
	result, err := ParseMojang(root, manifest.SourcePin{ID: "fixture", MinecraftVersion: "fixture", ProtocolVersion: 2208}, "")
	if err != nil {
		t.Fatal(err)
	}
	numeric := result.Claims[0].Encode
	if numeric.Kind != manifest.KindEnum || numeric.Primitive.Code != "zigzag_i32" {
		t.Fatalf("numeric enum = %+v", numeric)
	}
	for i, want := range []int64{-1, 0, 1, 5, 0} {
		if numeric.Variants[i].Value != want {
			t.Fatalf("variant %d = %+v, want %d", i, numeric.Variants[i], want)
		}
	}
	if text := result.Claims[1].Encode; text.Kind != manifest.KindString || text.Prefix.Primitive.Code != "var_u32" {
		t.Fatalf("text enum = %+v", text)
	}
}

// TestMojangBinaryEnumRejectsInvalidMappings ensures malformed metadata never becomes guessed ordinals.
func TestMojangBinaryEnumRejectsInvalidMappings(t *testing.T) {
	for name, mapping := range map[string]any{
		"missing": nil, "short": []any{float64(1)}, "long": []any{float64(1), float64(2), float64(3)},
		"not array": "1,2", "fraction": []any{1.5, float64(2)}, "numeric string": []any{"1junk", float64(2)},
		"rounded": []any{float64(1 << 54), float64(2)},
	} {
		t.Run(name, func(t *testing.T) {
			schema := map[string]any{"type": "string", "enum": []any{"One", "Two"}, "x-underlying-type": "int64", "x-serialization-options": []any{"Enum-as-Value"}}
			if mapping != nil {
				schema["x-enum-binary-value"] = mapping
			}
			lowerer := &mojangLowerer{}
			if node := lowerer.lowerSchema(schema, "Mode.json", "Mode"); node.Kind != manifest.KindUnresolved {
				t.Fatalf("invalid mapping lowered to %+v", node)
			}
		})
	}
}

// TestMojangEnumReferenceCombinesOptions keeps definition flags when a use selects numeric encoding.
func TestMojangEnumReferenceCombinesOptions(t *testing.T) {
	lowerer := &mojangLowerer{active: map[string]bool{}, documents: map[string]any{
		"Mode.json": map[string]any{"type": "string", "enum": []any{"Value"}, "x-enum-binary-value": []any{float64(7)}, "x-underlying-type": "int32", "x-serialization-options": []any{"Compression"}},
	}}
	node := lowerer.lowerSchema(map[string]any{"$ref": "./Mode.json", "x-serialization-options": []any{"Enum-as-Value"}}, "Packet.json", "Mode")
	if node.Kind != manifest.KindEnum || node.Primitive.Code != "zigzag_i32" {
		t.Fatalf("combined enum flags = %+v", node)
	}
}

// TestMojangScalarReferencePreservesUseSiteEncoding checks one shared integer in fixed and compressed fields.
func TestMojangScalarReferencePreservesUseSiteEncoding(t *testing.T) {
	definition := map[string]any{"type": "integer", "x-underlying-type": "int32"}
	lowerer := &mojangLowerer{documents: map[string]any{"StackID.json": definition}, active: map[string]bool{}}
	for _, test := range []struct {
		options []any
		code    string
	}{
		{[]any{"Compression"}, "zigzag_i32"},
		{nil, "i32le"},
		{[]any{"Big Endian"}, "i32be"},
	} {
		node := lowerer.lowerSchema(map[string]any{"$ref": "./StackID.json", "x-serialization-options": test.options}, "Packet.json", "ID")
		if node.Kind != manifest.KindPrimitive || node.Primitive.Code != test.code {
			t.Fatalf("reference with options %v = %#v, want %s", test.options, node, test.code)
		}
	}
	if _, exists := definition["x-serialization-options"]; exists {
		t.Fatal("use-site encoding changed the shared definition")
	}
}

// TestMojangCompressionPreservesSmallIntegerWidths checks the documented 32- and 64-bit compression boundary.
func TestMojangCompressionPreservesSmallIntegerWidths(t *testing.T) {
	for _, test := range []struct{ underlying, code string }{
		{"int8", "i8"}, {"uint8", "u8"}, {"int16", "i16le"}, {"uint16", "u16le"},
		{"int32", "zigzag_i32"}, {"uint32", "var_u32"}, {"int64", "zigzag_i64"}, {"uint64", "var_u64"},
	} {
		node := primitive(test.underlying, []string{"Compression"}, "integer")
		if node.Kind != manifest.KindPrimitive || node.Primitive.Code != test.code {
			t.Fatalf("compressed %s = %#v, want %s", test.underlying, node, test.code)
		}
	}
}

// TestEndstoneU6StringDiscriminatorPreservesItsWidthAndOrdinalCases checks the compact U6 schema form.
func TestEndstoneU6StringDiscriminatorPreservesItsWidthAndOrdinalCases(t *testing.T) {
	lowerer := &endstoneLowerer{types: map[string]any{}, enums: map[string]any{}, active: map[string]bool{}}
	for _, control := range []string{"uvarint32", "uint8"} {
		compact := lowerer.lowerTypeValue(map[string]any{"switch": control, "cases": []any{nil, nil}}, "Choice", "Packet.Choice")
		previous := lowerer.lowerTypeValue(map[string]any{"switch": map[string]any{"type": control}, "cases": []any{nil, nil}}, "Choice", "Packet.Choice")
		if !reflect.DeepEqual(compact, previous) || compact.Kind != manifest.KindUnion || len(compact.Variants) != 2 || compact.Variants[1].Value != 1 {
			t.Fatalf("compact switch differs: %#v vs %#v", compact, previous)
		}
	}
}

// TestEndstoneCompactRepeatPreservesAllFraming checks variable and fixed U6 arrays and rejects malformed metadata.
func TestEndstoneCompactRepeatPreservesAllFraming(t *testing.T) {
	inner := manifest.Primitive("u8")
	for _, pair := range []struct {
		compact any
		legacy  map[string]any
	}{
		{"uvarint32", map[string]any{"prefix": "uvarint32"}},
		{"uint32", map[string]any{"prefix": "uint32"}},
		{float64(4), map[string]any{"count": float64(4)}},
	} {
		a, b := lowerEndstoneRepeat(pair.compact, inner, "fixture"), lowerEndstoneRepeat(pair.legacy, inner, "fixture")
		if !reflect.DeepEqual(a, b) || a.Kind == manifest.KindUnresolved {
			t.Fatalf("compact repeat differs: %#v / %#v", a, b)
		}
	}
	if node := lowerEndstoneRepeat(true, inner, "fixture"); node.Kind != manifest.KindUnresolved {
		t.Fatal("malformed repeat silently disappeared")
	}
	lowerer := &endstoneLowerer{types: map[string]any{}, enums: map[string]any{}, active: map[string]bool{}}
	node := lowerer.lowerTypeValue(map[string]any{"switch": "uvarint32", "cases": []any{map[string]any{"type": "string", "repeat": "uvarint32"}}}, "Choice", "Packet.Choice")
	if node.Kind != manifest.KindUnion || node.Variants[0].Encode.Kind != manifest.KindArray {
		t.Fatalf("inline repeated union case: %#v", node)
	}
}

// TestEndstoneCompactSwitchKeepsInnerTagsSeparate checks reversed and multi-value legacy tags.
func TestEndstoneCompactSwitchKeepsInnerTagsSeparate(t *testing.T) {
	for _, control := range []string{"uint8", "uvarint32"} {
		for _, firstValues := range [][]any{{1}, {1, 2}} {
			lowerer := &endstoneLowerer{
				types: map[string]any{},
				enums: map[string]any{"Action.json": map[string]any{"values": []any{
					map[string]any{"name": "Add", "value": 0},
					map[string]any{"name": "Remove", "value": 1},
					map[string]any{"name": "Other", "value": 2},
				}}},
				active: map[string]bool{},
			}
			for i, name := range []string{"First", "Second"} {
				values := []any{0}
				if i == 0 {
					values = firstValues
				}
				lowerer.types[name+".json"] = map[string]any{"name": name, "fields": []any{
					map[string]any{"name": "Action", "type": "uint8", "enum": "Action", "constraints": map[string]any{"enum_values": values}},
				}}
			}
			node := lowerer.lowerTypeValue(map[string]any{"switch": control, "cases": []any{"First", "Second"}}, "Choice", "Packet.Choice")
			if node.Kind != manifest.KindUnion || len(node.Variants) != 2 {
				t.Fatalf("compact switch = %#v, want two positional alternatives", node)
			}
			for i, variant := range node.Variants {
				if variant.Value != int64(i) || variant.Encode.Kind != manifest.KindStruct || len(variant.Encode.Fields) != 1 {
					t.Fatalf("variant %d = %#v, want its index and retained Action field", i, variant)
				}
				action := variant.Encode.Fields[0].Encode
				want := []any{0}
				if i == 0 {
					want = firstValues
				}
				if action.Kind != manifest.KindEnum || action.Primitive.Code != "u8" || len(action.Variants) != len(want) {
					t.Fatalf("inner tag = %#v, want its own constrained uint8 enum", action)
				}
				for j, value := range want {
					if action.Variants[j].Value != int64(value.(int)) {
						t.Fatalf("inner tag value = %d, want %v", action.Variants[j].Value, value)
					}
				}
			}
		}
	}
}

func TestSingletonAllOfRetainsOuterAndInnerConstraints(t *testing.T) {
	lowerer := &mojangLowerer{documents: map[string]any{"String.json": map[string]any{"type": "string", "minLength": 2}}, active: map[string]bool{}}
	for _, branch := range []map[string]any{{"type": "string", "minLength": 2}, {"$ref": "String.json"}} {
		node := lowerer.lowerSchema(map[string]any{"allOf": []any{branch}, "maxLength": 4}, "Packet.json", "Value")
		if node.Constraints == nil || node.Constraints.MaxLength == nil || *node.Constraints.MaxLength != 4 || node.Constraints.MinLength == nil || *node.Constraints.MinLength != 2 {
			t.Fatalf("allOf constraints = %#v", node.Constraints)
		}
	}
}

func TestEndstoneUsesDiscoveredFileAndExactEnumIntegers(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"packets", "types", "enums"} {
		if err := os.Mkdir(filepath.Join(root, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	writeJSON(t, filepath.Join(root, "enums", "Large.json"), map[string]any{"name": "Large", "values": []any{map[string]any{"name": "Exact", "value": int64(9007199254740993)}}})
	writeJSON(t, filepath.Join(root, "packets", "A.json"), map[string]any{"name": "B", "id": 1, "fields": []any{map[string]any{"name": "Value", "type": "int64", "enum": "Large"}}})
	writeJSON(t, filepath.Join(root, "packets", "B.json"), map[string]any{"name": "C", "id": 2, "fields": []any{map[string]any{"name": "Text", "type": "string"}}})
	result, err := ParseEndstone(root, fixturePin("endstone"), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Claims) != 2 {
		t.Fatalf("claims = %#v", result.Claims)
	}
	first := result.Claims[0]
	if first.PacketID != 1 || first.PacketName != "B" || first.Name != "Value" || first.Locator != "packets/A.json#/fields/0" {
		t.Fatalf("first claim = %#v", first)
	}
	if first.Encode.Kind != manifest.KindEnum || first.Encode.Variants[0].Value != 9007199254740993 {
		t.Fatalf("enum = %#v", first.Encode)
	}
	if result.Packets[0].Locator != "packets/A.json" || result.Claims[1].Locator != "packets/B.json#/fields/0" {
		t.Fatal("incorrect source locators")
	}
}

func TestCorrectionsRetainPrePatchContextAndExactNumbers(t *testing.T) {
	for _, operation := range []string{"merge", "replace"} {
		t.Run(operation, func(t *testing.T) {
			node := map[string]any{"name": "Value"}
			documents := map[string]any{"Packet.json": map[string]any{"field": node}}
			pin := fixturePin("endstone")
			before, err := canonicalDigest(node)
			if err != nil {
				t.Fatal(err)
			}
			context, err := correctionContextDigest(pin, "Packet.json#/field", node)
			if err != nil {
				t.Fatal(err)
			}
			patch := map[string]any{"id": "value", "file": "Packet.json", "pointer": "/field", "pre_patch_sha256": before, "why": "fixture", "evidence": []any{map[string]any{"locator": "fixture/runtime"}}}
			value := map[string]any{"value": int64(9007199254740993)}
			if operation == "replace" {
				value["name"] = "Value"
			}
			patch[operation] = value
			corrections := t.TempDir()
			writeJSON(t, filepath.Join(corrections, "value.json"), map[string]any{"operations": []any{patch}})
			proofs, err := applyCorrections(documents, corrections, pin)
			if err != nil {
				t.Fatal(err)
			}
			if proofs[0].PrePatchContextSHA256 != context {
				t.Error("context does not fingerprint original node")
			}
			updated, _ := valueAt(documents["Packet.json"], "/field/value")
			if got, ok := asInt(updated); !ok || got != 9007199254740993 {
				t.Fatalf("corrected integer = %v", updated)
			}
		})
	}
}
