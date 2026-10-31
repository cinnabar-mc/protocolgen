package changelog

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"protocolgen/internal/updateguide"
)

func TestGenerateReportsWireConstraintAndEnumChanges(t *testing.T) {
	from := t.TempDir()
	to := t.TempDir()

	writeSchema(t, from, "RequestNetworkSettingsPacket", `{
  "title":"RequestNetworkSettingsPacket","x-minecraft-version":"1.26.40-beta.0","x-protocol-version":2168,
  "$ref":"./RequestNetworkSettingsPacketPayload.json","$metaProperties":{"[cereal:packet]":193}}
`)
	writeSchema(t, to, "RequestNetworkSettingsPacket", `{
  "title":"RequestNetworkSettingsPacket","x-minecraft-version":"1.26.50","x-protocol-version":2169,
  "$ref":"./RequestNetworkSettingsPacketPayload.json","$metaProperties":{"[cereal:packet]":193}}
`)
	writeSchema(t, from, "RequestNetworkSettingsPacketPayload", objectSchema("1.26.40-beta.0", 2168, `
    "ClientNetworkVersion":{"type":"integer","x-underlying-type":"int32","x-serialization-options":["Big Endian"],"x-ordinal-index":0,"minimum":2168,"maximum":2168}`, `"ClientNetworkVersion"`))
	writeSchema(t, to, "RequestNetworkSettingsPacketPayload", objectSchema("1.26.50", 2169, `
    "ClientNetworkVersion":{"type":"integer","x-underlying-type":"int32","x-serialization-options":["Big Endian"],"x-ordinal-index":0,"minimum":2169,"maximum":2169}`, `"ClientNetworkVersion"`))

	writeSchema(t, from, "TextDataPayload", objectSchema("1.26.40-beta.0", 2168, `
    "Text":{"type":"string","x-ordinal-index":0},
    "DepthTest":{"type":"boolean","x-underlying-type":"boolean","x-ordinal-index":1}`, `"Text","DepthTest"`))
	writeSchema(t, to, "TextDataPayload", objectSchema("1.26.50", 2169, `
    "Text":{"type":"string","x-ordinal-index":0},
    "LineGapHeight":{"type":"number","x-underlying-type":"float","x-ordinal-index":1},
    "DepthTest":{"type":"boolean","x-underlying-type":"boolean","x-ordinal-index":2}`, `"Text","DepthTest"`))
	writeSchema(t, from, "PrimitiveShapesPacket", packetSchema("1.26.40-beta.0", 2168, 328, "PrimitiveShapesPacketPayload"))
	writeSchema(t, to, "PrimitiveShapesPacket", packetSchema("1.26.50", 2169, 328, "PrimitiveShapesPacketPayload"))
	writeSchema(t, from, "PrimitiveShapesPacketPayload", objectWithRef("1.26.40-beta.0", 2168, "Shape", "TextDataPayload"))
	writeSchema(t, to, "PrimitiveShapesPacketPayload", objectWithRef("1.26.50", 2169, "Shape", "TextDataPayload"))

	writeSchema(t, from, "persona__AnimatedTextureType", enumSchema("1.26.40-beta.0", 2168, `"Face","Body32x32"`))
	writeSchema(t, to, "persona__AnimatedTextureType", enumSchema("1.26.50", 2169, `"None","Face","Body32x32"`))
	writeSchema(t, from, "PlayerSkinPacket", packetSchema("1.26.40-beta.0", 2168, 93, "PlayerSkinPacketPayload"))
	writeSchema(t, to, "PlayerSkinPacket", packetSchema("1.26.50", 2169, 93, "PlayerSkinPacketPayload"))
	writeSchema(t, from, "PlayerSkinPacketPayload", objectWithRef("1.26.40-beta.0", 2168, "AnimationType", "persona__AnimatedTextureType"))
	writeSchema(t, to, "PlayerSkinPacketPayload", objectWithRef("1.26.50", 2169, "AnimationType", "persona__AnimatedTextureType"))

	got, err := Generate(Config{
		FromDir: from, ToDir: to,
		FromBranch: "automated/1.26.40", ToBranch: "r/26_u4",
		FromUpstream: "0e00fe80f4", ToUpstream: "0f6a0bff19",
		FromFixer: "9fa8eb75f9", ToFixer: "9fa8eb75f9",
	})
	if err != nil {
		t.Fatal(err)
	}
	text := string(got)
	for _, want := range []string{
		"# Bedrock protocol changes — 2168 to 2169",
		"1.26.40-beta.0 → 1.26.50",
		"**0** new, **0** removed, **0** renamed, **3** modified",
		"**2** type(s) change the bytes on the wire",
		"**3** packet(s) touched, counting those that only embed a changed type",
		"constraints changed from maximum=2168, minimum=2168 to maximum=2169, minimum=2169",
		"### TextDataPayload\nstruct · **wire break**",
		"added `LineGapHeight` at ordinal 1 (float, optional)",
		"`DepthTest` moved from ordinal 1 to 2",
		"**Also affects:** `PrimitiveShapesPacket`",
		"### persona__AnimatedTextureType\nenum",
		"added value `None` = 0",
		"**Also affects:** `PlayerSkinPacket`",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("output missing %q\n%s", want, text)
		}
	}
	if _, err := updateguide.Generate(got, to); err != nil {
		t.Fatalf("generated changelog is not accepted by update-guide: %v", err)
	}
}

func TestGenerateRejectsInconsistentMetadata(t *testing.T) {
	from := t.TempDir()
	to := t.TempDir()
	writeSchema(t, from, "One", objectSchema("1.0", 1, `"A":{"type":"string","x-ordinal-index":0}`, `"A"`))
	writeSchema(t, from, "Two", objectSchema("1.0", 2, `"A":{"type":"string","x-ordinal-index":0}`, `"A"`))
	writeSchema(t, to, "One", objectSchema("2.0", 3, `"A":{"type":"string","x-ordinal-index":0}`, `"A"`))
	if _, err := Generate(Config{FromDir: from, ToDir: to}); err == nil || !strings.Contains(err.Error(), "inconsistent protocol version") {
		t.Fatalf("expected inconsistent protocol version error, got %v", err)
	}
}

func TestGenerateIgnoresDocumentationOnlyChanges(t *testing.T) {
	from := t.TempDir()
	to := t.TempDir()
	writeSchema(t, from, "Shared", `{"title":"Shared","description":"old docs","x-minecraft-version":"1.0","x-protocol-version":1,"type":"object","properties":{"Value":{"description":"old field docs","type":"string","x-ordinal-index":0}},"required":["Value"]}`)
	writeSchema(t, to, "Shared", `{"title":"Shared","description":"new docs","x-minecraft-version":"2.0","x-protocol-version":2,"type":"object","properties":{"Value":{"description":"new field docs","type":"string","x-ordinal-index":0}},"required":["Value"]}`)

	got, err := Generate(Config{FromDir: from, ToDir: to})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(got), "## Modified Types") || !strings.Contains(string(got), "**0** new, **0** removed, **0** renamed, **0** modified") {
		t.Fatalf("documentation-only edits were reported as protocol changes:\n%s", got)
	}
}

func TestGenerateRendersAddedRemovedAndRenamedDefinitions(t *testing.T) {
	from := t.TempDir()
	to := t.TempDir()
	writeSchema(t, from, "OldType", objectSchema("1.0", 1, `"Value":{"type":"string","x-ordinal-index":0}`, `"Value"`))
	writeSchema(t, to, "NewType", objectSchema("2.0", 2, `"Value":{"type":"string","x-ordinal-index":0}`, `"Value"`))
	writeSchema(t, from, "OldPacket", packetSchema("1.0", 1, 7, "OldPacketPayload"))
	writeSchema(t, from, "OldPacketPayload", objectSchema("1.0", 1, `"Value":{"type":"string","x-ordinal-index":0}`, `"Value"`))
	writeSchema(t, to, "NewPacket", packetSchema("2.0", 2, 7, "NewPacketPayload"))
	writeSchema(t, to, "NewPacketPayload", objectSchema("2.0", 2, `"Value":{"type":"string","x-ordinal-index":0},"Added":{"type":"boolean","x-ordinal-index":1}`, `"Value","Added"`))

	got, err := Generate(Config{FromDir: from, ToDir: to})
	if err != nil {
		t.Fatal(err)
	}
	text := string(got)
	for _, want := range []string{
		"**1** new, **1** removed, **1** renamed, **1** modified",
		"## Added Packets\n\n### NewPacket",
		"## Added Types\n\n### NewType",
		"## Removed Packets\n\n### OldPacket",
		"## Removed Types\n\n### OldType",
		"## Renames\n\n- `OldPacket` → `NewPacket` (packet id `7`)",
		"added `Added` at ordinal 1 (boolean)",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("output missing %q\n%s", want, text)
		}
	}
	guide, err := updateguide.Generate(got, to)
	if err != nil {
		t.Fatalf("rename changelog is not accepted by update-guide: %v", err)
	}
	if count := strings.Count(string(guide), "type NewPacket struct"); count != 1 {
		t.Fatalf("renamed target definition rendered %d times, want once:\n%s", count, guide)
	}
}

func TestGenerateReportsPacketIDArrayAndEnumWireChanges(t *testing.T) {
	from := t.TempDir()
	to := t.TempDir()
	writeSchema(t, from, "ChangedPacket", packetSchema("1.0", 1, 7, "ChangedPacketPayload"))
	writeSchema(t, to, "ChangedPacket", packetSchema("2.0", 2, 99, "ChangedPacketPayload"))
	writeSchema(t, from, "ChangedPacketPayload", objectSchema("1.0", 1, `"Values":{"type":"array","items":{"type":"integer","x-underlying-type":"int8"},"x-ordinal-index":0}`, `"Values"`))
	writeSchema(t, to, "ChangedPacketPayload", objectSchema("2.0", 2, `"Values":{"type":"array","items":{"type":"integer","x-underlying-type":"int64"},"x-ordinal-index":0}`, `"Values"`))
	writeSchema(t, from, "Mode", `{"title":"Mode","x-minecraft-version":"1.0","x-protocol-version":1,"type":"string","enum":["A","B"],"x-enum-binary-value":[10,20],"x-underlying-type":"uint8"}`)
	writeSchema(t, to, "Mode", `{"title":"Mode","x-minecraft-version":"2.0","x-protocol-version":2,"type":"string","enum":["B","A"],"x-enum-binary-value":[20,10],"x-underlying-type":"uint32"}`)

	got, err := Generate(Config{FromDir: from, ToDir: to})
	if err != nil {
		t.Fatal(err)
	}
	text := string(got)
	for _, want := range []string{
		"packet id changed from `7` to `99`",
		"type changed from array of int8 to array of int64",
		"packet id `99` · struct · **wire break**",
		"value order changed from `A`, `B` to `B`, `A`",
		"underlying type changed from uint8 to uint32",
		"enum · **wire break**",
		"**2** type(s) change the bytes on the wire",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("output missing %q\n%s", want, text)
		}
	}
}

func TestGenerateRejectsMissingMetadataAndDuplicatePacketIDs(t *testing.T) {
	from := t.TempDir()
	to := t.TempDir()
	writeSchema(t, from, "Missing", `{"title":"Missing","type":"object"}`)
	writeSchema(t, to, "Valid", objectSchema("2.0", 2, `"A":{"type":"string","x-ordinal-index":0}`, `"A"`))
	if _, err := Generate(Config{FromDir: from, ToDir: to}); err == nil || !strings.Contains(err.Error(), "missing x-minecraft-version") {
		t.Fatalf("expected missing metadata error, got %v", err)
	}

	from = t.TempDir()
	to = t.TempDir()
	for _, dir := range []string{from, to} {
		writeSchema(t, dir, "OnePacket", packetSchema("1.0", 1, 7, "OnePayload"))
		writeSchema(t, dir, "TwoPacket", packetSchema("1.0", 1, 7, "TwoPayload"))
		writeSchema(t, dir, "OnePayload", objectSchema("1.0", 1, `"A":{"type":"string","x-ordinal-index":0}`, `"A"`))
		writeSchema(t, dir, "TwoPayload", objectSchema("1.0", 1, `"A":{"type":"string","x-ordinal-index":0}`, `"A"`))
	}
	if _, err := Generate(Config{FromDir: from, ToDir: to}); err == nil || !strings.Contains(err.Error(), "duplicate packet id 7") {
		t.Fatalf("expected duplicate packet id error, got %v", err)
	}
}

func TestGenerateDoesNotTreatPropertyNamesAsSchemaKeywords(t *testing.T) {
	from := t.TempDir()
	to := t.TempDir()
	writeSchema(t, from, "Shared", objectSchema("1.0", 1, `"description":{"type":"string","x-ordinal-index":0}`, `"description"`))
	writeSchema(t, to, "Shared", objectSchema("2.0", 2, `"description":{"type":"integer","x-underlying-type":"int32","x-ordinal-index":0}`, `"description"`))
	got, err := Generate(Config{FromDir: from, ToDir: to})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "`description` type changed from string to int32") {
		t.Fatalf("property change was hidden:\n%s", got)
	}
}

func writeSchema(t *testing.T, dir, name, body string) {
	t.Helper()
	body = strings.Replace(body, `"title":"Payload"`, `"title":"`+name+`"`, 1)
	body = strings.Replace(body, `"title":"Packet"`, `"title":"`+name+`"`, 1)
	if err := os.WriteFile(filepath.Join(dir, name+".json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func objectSchema(version string, protocol int, properties, required string) string {
	return `{"title":"Payload","x-minecraft-version":"` + version + `","x-protocol-version":` + strconv.Itoa(protocol) + `,"type":"object","properties":{` + properties + `},"required":[` + required + `]}`
}

func packetSchema(version string, protocol, id int, payload string) string {
	return `{"title":"Packet","x-minecraft-version":"` + version + `","x-protocol-version":` + strconv.Itoa(protocol) + `,"$ref":"./` + payload + `.json","$metaProperties":{"[cereal:packet]":` + strconv.Itoa(id) + `}}`
}

func objectWithRef(version string, protocol int, field, ref string) string {
	return objectSchema(version, protocol, `"`+field+`":{"$ref":"./`+ref+`.json","x-ordinal-index":0}`, `"`+field+`"`)
}

func enumSchema(version string, protocol int, values string) string {
	return `{"title":"persona::AnimatedTextureType","x-minecraft-version":"` + version + `","x-protocol-version":` + strconv.Itoa(protocol) + `,"type":"string","enum":[` + values + `],"x-underlying-type":"uint32"}`
}

func TestGenerateReportsOfficialEnumValues(t *testing.T) {
	from, to := t.TempDir(), t.TempDir()
	writeSchema(t, from, "Mode", `{"title":"Mode","x-minecraft-version":"1.0","x-protocol-version":1,"type":"string","enum":["A","B"],"x-enum-binary-value":[-1,7],"x-underlying-type":"int32"}`)
	writeSchema(t, to, "Mode", `{"title":"Mode","x-minecraft-version":"2.0","x-protocol-version":2,"type":"string","enum":["A","B","Alias"],"x-enum-binary-value":[-1,9,9],"x-underlying-type":"int32"}`)
	got, err := Generate(Config{FromDir: from, ToDir: to})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"added value `Alias` = 9", "value `B` changed from 7 to 9"} {
		if !strings.Contains(string(got), want) {
			t.Fatalf("missing %q in %s", want, got)
		}
	}
}

func TestEnumMetadataAdditionDoesNotInventWireChanges(t *testing.T) {
	before := &schema{Enum: []string{"A", "B"}, Underlying: "int32"}
	after := &schema{Enum: []string{"A", "B"}, Underlying: "int32", EnumValues: []interface{}{float64(0), float64(1)}}
	got := compareEnums("Mode", before, after)
	if got.Wire || strings.Contains(strings.Join(got.Bullets, " "), "changed from") {
		t.Fatalf("metadata addition invented a wire change: %+v", got)
	}
	if len(got.Bullets) != 1 || got.Bullets[0] != "explicit binary enum values added" {
		t.Fatalf("metadata addition missing: %+v", got)
	}
}

func TestGenerateReportsEffectiveEnumValueChanges(t *testing.T) {
	for _, tc := range []struct{ name, before, after, want string }{
		{"implicit insertion", `"enum":["A","B"]`, `"enum":["X","A","B"]`, "value `B` changed from 1 to 2"},
		{"explicit added", `"enum":["A","B"]`, `"enum":["A","B"],"x-enum-binary-value":[-1,7]`, "value `B` changed from 1 to 7"},
		{"explicit removed", `"enum":["A","B"],"x-enum-binary-value":[-1,7]`, `"enum":["A","B"]`, "value `B` changed from 7 to 1"},
		{"large integer", `"enum":["A"],"x-enum-binary-value":[9007199254740992]`, `"enum":["A"],"x-enum-binary-value":[9007199254740993]`, "value `A` changed from 9007199254740992 to 9007199254740993"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			from, to := t.TempDir(), t.TempDir()
			for dir, values := range map[string]string{from: tc.before, to: tc.after} {
				writeSchema(t, dir, "Mode", `{"title":"Mode","x-minecraft-version":"1.0","x-protocol-version":1,"type":"string","x-underlying-type":"int64",`+values+`}`)
			}
			got, err := Generate(Config{FromDir: from, ToDir: to})
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(got), tc.want) || !strings.Contains(string(got), "enum · **wire break**") {
				t.Fatalf("missing effective enum change %q: %s", tc.want, got)
			}
		})
	}
}

func TestGenerateRejectsInvalidEnumMappingLength(t *testing.T) {
	for _, mapping := range []string{"[]", "null", "[1]"} {
		t.Run(mapping, func(t *testing.T) {
			from, to := t.TempDir(), t.TempDir()
			writeSchema(t, from, "Mode", `{"title":"Mode","x-minecraft-version":"1.0","x-protocol-version":1,"type":"string","enum":["A","B"],"x-enum-binary-value":`+mapping+`}`)
			if _, err := Generate(Config{FromDir: from, ToDir: to}); err == nil || !strings.Contains(err.Error(), "x-enum-binary-value") {
				t.Fatalf("expected mapping error, got %v", err)
			}
		})
	}
}

func TestGenerateTracksDefaultPresenceInOptionality(t *testing.T) {
	for _, defaultValue := range []string{"false", "0", "null"} {
		t.Run(defaultValue, func(t *testing.T) {
			from, to := t.TempDir(), t.TempDir()
			writeSchema(t, from, "Settings", objectSchema("1.0", 1, `"Value":{"type":"boolean","x-ordinal-index":0}`, ``))
			writeSchema(t, to, "Settings", objectSchema("2.0", 2, `"Value":{"type":"boolean","x-ordinal-index":0,"default":`+defaultValue+`}`, ``))
			got, err := Generate(Config{FromDir: from, ToDir: to})
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(got), "`Value` changed from optional to required") {
				t.Fatalf("default presence was ignored: %s", got)
			}
			reverse, err := Generate(Config{FromDir: to, ToDir: from})
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(reverse), "`Value` changed from required to optional") {
				t.Fatalf("default removal was ignored: %s", reverse)
			}
		})
	}
}

func TestDefaultValueChangesDoNotChangeWireShape(t *testing.T) {
	before := map[string]interface{}{"type": "boolean", "default": false}
	after := map[string]interface{}{"type": "boolean", "default": true}
	if !equalWire(before, after) {
		t.Fatal("default value change changed the wire shape")
	}
	for _, defaultValue := range []interface{}{false, 0, nil} {
		s := &schema{Properties: map[string]property{"Value": {Type: "boolean", Raw: map[string]interface{}{"default": defaultValue}}}}
		if got := fields(s); len(got) != 1 || got[0].Optional {
			t.Fatalf("defaulted field is optional: %+v", got)
		}
	}
}
