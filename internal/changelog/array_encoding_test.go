package changelog

import (
	"strings"
	"testing"
)

func TestGenerateReportsNestedArrayEncodingChanges(t *testing.T) {
	for _, tc := range []struct{ name, before, after, want string }{
		{"endianness", `{"type":"integer","x-underlying-type":"int32","x-serialization-options":["Big Endian"]}`, `{"type":"integer","x-underlying-type":"int32"}`, "array of int32 (Big Endian) to array of int32"},
		{"nested width", `{"type":"array","items":{"type":"integer","x-underlying-type":"int8"}}`, `{"type":"array","items":{"type":"integer","x-underlying-type":"int64"}}`, "array of array of int8 to array of array of int64"},
		{"inline object", `{"type":"object","properties":{"x":{"type":"boolean","x-ordinal-index":0}},"required":["x"]}`, `{"type":"object","properties":{"x":{"type":"string","x-ordinal-index":0}},"required":["x"]}`, "`Values` item encoding changed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			from, to := t.TempDir(), t.TempDir()
			for dir, items := range map[string]string{from: tc.before, to: tc.after} {
				writeSchema(t, dir, "Values", objectSchema("1.0", 1, `"Values":{"type":"array","items":`+items+`,"x-ordinal-index":0}`, `"Values"`))
			}
			got, err := Generate(Config{FromDir: from, ToDir: to})
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(got), tc.want) || !strings.Contains(string(got), "struct · **wire break**") {
				t.Fatalf("missing array encoding change: %s", got)
			}
		})
	}
}
