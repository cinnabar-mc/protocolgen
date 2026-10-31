package updateguide

import (
	"strings"
	"testing"
)

func TestGenerateRejectsNormalizedNameCollisions(t *testing.T) {
	for _, tc := range []struct{ name, schema string }{
		{"fields", `"type":"object","properties":{"Foo Bar":{"type":"boolean","x-ordinal-index":0},"FooBar":{"type":"boolean","x-ordinal-index":1}},"required":["Foo Bar","FooBar"]`},
		{"enum", `"type":"string","enum":["Foo Bar","FooBar"],"x-underlying-type":"uint32"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			directory := t.TempDir()
			writeFixture(t, directory, "Example.json", `{"title":"Example","x-protocol-version":2,`+tc.schema+`}`)
			_, err := Generate([]byte("# Bedrock protocol changes — 1 to 2\n\n## Modified Types\n\n### Example\n"), directory)
			if err == nil || !strings.Contains(err.Error(), "collision") || !strings.Contains(err.Error(), "Foo Bar") || !strings.Contains(err.Error(), "FooBar") {
				t.Fatalf("expected names in collision error, got %v", err)
			}
		})
	}
}
