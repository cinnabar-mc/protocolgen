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

func TestGenerateTracksSharedPacketPayloadDependencies(t *testing.T) {
	from, to := t.TempDir(), t.TempDir()
	for dir, underlying := range map[string]string{from: "int8", to: "int64"} {
		writeSchema(t, dir, "A", packetSchema("1.0", 1, 1, "APayload"))
		writeSchema(t, dir, "B", packetSchema("1.0", 1, 2, "BPayload"))
		writeSchema(t, dir, "APayload", objectSchema("1.0", 1, `"Value":{"type":"integer","x-underlying-type":"`+underlying+`","x-ordinal-index":0}`, `"Value"`))
		writeSchema(t, dir, "BPayload", objectSchema("1.0", 1, `"Data":{"$ref":"APayload.json","x-ordinal-index":0}`, `"Data"`))
	}
	got, err := Generate(Config{FromDir: from, ToDir: to})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "**Also affects:** `B`") || !strings.Contains(string(got), "**2** packet(s) touched") {
		t.Fatalf("payload dependency lost: %s", got)
	}
}

func TestPacketIDChangesDoNotAffectPayloadConsumers(t *testing.T) {
	from, to := t.TempDir(), t.TempDir()
	for dir, id := range map[string]int{from: 1, to: 3} {
		writeSchema(t, dir, "A", packetSchema("1.0", 1, id, "APayload"))
		writeSchema(t, dir, "B", packetSchema("1.0", 1, 2, "BPayload"))
		writeSchema(t, dir, "APayload", objectSchema("1.0", 1, `"Value":{"type":"boolean","x-ordinal-index":0}`, `"Value"`))
		writeSchema(t, dir, "BPayload", objectSchema("1.0", 1, `"Data":{"$ref":"APayload.json","x-ordinal-index":0}`, `"Data"`))
	}
	got, err := Generate(Config{FromDir: from, ToDir: to})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(got), "**Also affects:** `B`") || !strings.Contains(string(got), "**1** packet(s) touched") {
		t.Fatalf("packet ID propagated into payload: %s", got)
	}
}

func TestRenamedPacketAffectedBySharedTypeCountsOnce(t *testing.T) {
	from, to := t.TempDir(), t.TempDir()
	for dir, packet := range map[string]string{from: "OldPacket", to: "NewPacket"} {
		writeSchema(t, dir, packet, packetSchema("1.0", 1, 1, "Payload"))
		writeSchema(t, dir, "Payload", objectSchema("1.0", 1, `"Data":{"$ref":"Shared.json","x-ordinal-index":0}`, `"Data"`))
		underlying := "int8"
		if dir == to {
			underlying = "int64"
		}
		writeSchema(t, dir, "Shared", objectSchema("1.0", 1, `"Value":{"type":"integer","x-underlying-type":"`+underlying+`","x-ordinal-index":0}`, `"Value"`))
	}
	got, err := Generate(Config{FromDir: from, ToDir: to})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "**1** packet(s) touched") {
		t.Fatalf("renamed dependent counted twice: %s", got)
	}
}
