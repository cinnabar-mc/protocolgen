package updateguide

import (
	"fmt"
	"strings"
	"testing"
)

func TestGeneratedMethodsReserveFieldNames(t *testing.T) {
	for _, tc := range []struct {
		kind, name string
		reject     bool
	}{{"Types", "Marshal", true}, {"Packets", "Marshal", true}, {"Packets", "ID", true}, {"Types", "ID", false}} {
		t.Run(tc.kind+tc.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFixture(t, dir, "Example.json", fmt.Sprintf(`{"title":"Example","x-protocol-version":2,"type":"object","properties":{%q:{"type":"boolean","x-ordinal-index":0}},"required":[%q]}`, tc.name, tc.name))
			_, err := Generate([]byte("# Bedrock protocol changes — 1 to 2\n\n## Modified "+tc.kind+"\n\n### Example\n"), dir)
			if (err != nil) != tc.reject {
				t.Fatalf("collision error = %v", err)
			}
		})
	}
}

func TestStringAndBooleanOptionsFailClosed(t *testing.T) {
	for _, typ := range []string{"string", "boolean"} {
		for _, array := range []bool{false, true} {
			for _, required := range []bool{false, true} {
				t.Run(fmt.Sprint(typ, array, required), func(t *testing.T) {
					field := fmt.Sprintf(`{"type":%q,"x-serialization-options":["Mystery codec"]}`, typ)
					if array {
						field = `{"type":"array","items":` + field + `}`
					}
					schema := fmt.Sprintf(`{"title":"Example","x-protocol-version":2,"type":"object","properties":{"Value":%s}}`, strings.TrimSuffix(field, "}")+`,"x-ordinal-index":0}`)
					if required {
						schema = strings.TrimSuffix(schema, "}") + `,"required":["Value"]}`
					}
					dir := t.TempDir()
					writeFixture(t, dir, "Example.json", schema)
					_, err := Generate([]byte("# Bedrock protocol changes — 1 to 2\n\n## Modified Types\n\n### Example\n"), dir)
					if err == nil || !strings.Contains(err.Error(), "mystery codec") {
						t.Fatalf("unsupported option: %v", err)
					}
				})
			}
		}
	}
}

func TestInlineNumericEnumsUseUnderlyingCodec(t *testing.T) {
	schema := map[string]any{"type": "string", "enum": []any{"None", "Active"}, "x-underlying-type": "uint32"}
	typ, codec, _, err := renderFieldSchema(schema, schemaSet{}, "&pk.Mode", "")
	if err != nil || typ != "uint32" || codec != "io.Uint32(&pk.Mode)" {
		t.Fatalf("inline enum = %s %s %v", typ, codec, err)
	}
	typ, codec, _, err = renderFieldSchema(map[string]any{"type": "array", "items": schema}, schemaSet{}, "&pk.Modes", "")
	if err != nil || typ != "[]uint32" || !strings.Contains(codec, "io.Uint32") {
		t.Fatalf("inline enum array = %s %s %v", typ, codec, err)
	}
}
