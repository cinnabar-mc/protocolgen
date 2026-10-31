package changelog

import (
	"strings"
	"testing"
)

func TestGenerateTracksInlinePacketDependencies(t *testing.T) {
	from, to := t.TempDir(), t.TempDir()
	for dir, underlying := range map[string]string{from: "int8", to: "int64"} {
		writeSchema(t, dir, "InlinePacket", `{"title":"InlinePacket","x-minecraft-version":"1.0","x-protocol-version":1,"$metaProperties":{"[cereal:packet]":7},"type":"object","properties":{"Data":{"$ref":"./Shared.json","x-ordinal-index":0}},"required":["Data"]}`)
		writeSchema(t, dir, "Shared", objectSchema("1.0", 1, `"Value":{"type":"integer","x-underlying-type":"`+underlying+`","x-ordinal-index":0}`, `"Value"`))
	}
	got, err := Generate(Config{FromDir: from, ToDir: to})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "**Also affects:** `InlinePacket`") || !strings.Contains(string(got), "**1** packet(s) touched") {
		t.Fatalf("inline dependency was lost: %s", got)
	}
}
