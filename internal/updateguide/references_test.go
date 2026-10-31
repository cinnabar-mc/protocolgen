package updateguide

import (
	"strings"
	"testing"
)

func TestGenerateQualifiesPacketObjectReferences(t *testing.T) {
	directory := t.TempDir()
	writeFixture(t, directory, "Color.json", `{"title":"Color","x-protocol-version":2,"type":"object","properties":{}}`)
	writeFixture(t, directory, "PaintPacket.json", `{"title":"PaintPacket","x-protocol-version":2,"type":"object","properties":{
		"Required":{"$ref":"Color.json","x-ordinal-index":0},
		"Optional":{"$ref":"Color.json","x-ordinal-index":1},
		"Colors":{"type":"array","items":{"$ref":"Color.json"},"x-ordinal-index":2},
		"MaybeColors":{"type":"array","items":{"$ref":"Color.json"},"x-ordinal-index":3}
	},"required":["Required","Colors"]}`)
	got, err := Generate([]byte("# Bedrock protocol changes — 1 to 2\n\n## Modified Packets\n\n### PaintPacket\n"), directory)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"protocol.Color", "protocol.Optional[protocol.Color]", "[]protocol.Color", "protocol.Optional[[]protocol.Color]", "protocol.Single(io, &pk.Required)", "protocol.OptionalMarshaler(io, &pk.Optional)", "protocol.Slice(io, &pk.Colors)"} {
		if !strings.Contains(string(got), want) {
			t.Errorf("missing %q: %s", want, got)
		}
	}
}

func TestGenerateHonorsEnumReferenceOptions(t *testing.T) {
	for _, option := range []string{"Compression", "Unknown"} {
		t.Run(option, func(t *testing.T) {
			directory := t.TempDir()
			writeFixture(t, directory, "Mode.json", `{"title":"Mode","x-protocol-version":2,"type":"string","enum":["A"],"x-underlying-type":"uint32"}`)
			writeFixture(t, directory, "Settings.json", `{"title":"Settings","x-protocol-version":2,"type":"object","properties":{"Mode":{"$ref":"Mode.json","x-serialization-options":["`+option+`"],"x-ordinal-index":0}},"required":["Mode"]}`)
			got, err := Generate([]byte("# Bedrock protocol changes — 1 to 2\n\n## Modified Types\n\n### Settings\n"), directory)
			if option == "Unknown" {
				if err == nil || !strings.Contains(err.Error(), "unsupported serialization option") {
					t.Fatalf("unknown enum option accepted: %s, %v", got, err)
				}
			} else if err != nil || !strings.Contains(string(got), "io.Varuint32(&pk.Mode)") {
				t.Fatalf("enum reference options lost: %s, %v", got, err)
			}
		})
	}
}
