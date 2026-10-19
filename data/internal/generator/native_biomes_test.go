package generator

import (
	"encoding/json"
	"maps"
	"reflect"
	"strings"
	"testing"
)

// nativeBiomeFields creates a complete synthetic native record from its owning schema.
func nativeBiomeFields(t *testing.T) map[string]json.RawMessage {
	t.Helper()
	value := rawBiome{
		Temperature: 0.75, Downfall: 0.5, RedSporeDensity: 0.125,
		BlueSporeDensity: 0.25, AshDensity: 0.375, WhiteAshDensity: 0.5,
		FoliageSnow: 0.625, Depth: -1.5, Scale: 0.25, Rain: true,
	}
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}
	delete(fields, "id")
	delete(fields, "tags")
	fields["mapWaterColor"] = json.RawMessage(`"#11223344"`)
	return fields
}

// nativeBiomeJSON wraps one synthetic record in the namespaced native format.
func nativeBiomeJSON(t *testing.T, fields map[string]json.RawMessage) []byte {
	t.Helper()
	data, err := json.Marshal(map[string]any{"minecraft:plains": fields})
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// TestNativeBiomesOwnScalarsAndKeepIndependentTags checks source ownership and colour order.
func TestNativeBiomesOwnScalarsAndKeepIndependentTags(t *testing.T) {
	fields := nativeBiomeFields(t)
	fields["tags"] = json.RawMessage(`["unverified_native_membership"]`)
	fields["id"] = json.RawMessage(`99`)
	native, err := DecodeNativeBiomes(nativeBiomeJSON(t, fields))
	if err != nil {
		t.Fatal(err)
	}
	value := native["minecraft:plains"]
	if color := value.MapWaterColor; color.A != 0x11 || color.R != 0x22 || color.G != 0x33 || color.B != 0x44 {
		t.Fatalf("native #AARRGGBB decoded as %#v", color)
	}
	if value.ID != nil || len(value.Tags) != 0 {
		t.Fatalf("scalar input supplied unsupported identity or memberships: %#v", value)
	}
	id := int32(123)
	original := map[string]rawBiome{"minecraft:plains": {
		ID: &id, Temperature: -10, Depth: 10, Rain: false, Tags: []string{"overworld", "plains"},
	}}
	combined, err := applyNativeBiomes(original, native)
	if err != nil {
		t.Fatal(err)
	}
	want := value
	want.Tags = []string{"overworld", "plains"}
	if got := combined["minecraft:plains"]; !reflect.DeepEqual(got, want) {
		t.Fatalf("combined facts = %#v, want native scalars and original tags %#v", got, want)
	}
	combined["minecraft:plains"].Tags[0] = "changed"
	if original["minecraft:plains"].Tags[0] != "overworld" {
		t.Fatal("combined tags alias the original source")
	}
}

// TestNativeBiomeRequiredFieldsFollowOwner checks absence and null for every used schema field.
func TestNativeBiomeRequiredFieldsFollowOwner(t *testing.T) {
	owner := reflect.TypeOf(rawBiome{})
	for i := range owner.NumField() {
		field := owner.Field(i)
		if field.Name == "ID" || field.Name == "Tags" {
			continue
		}
		key := strings.Split(field.Tag.Get("json"), ",")[0]
		for _, null := range []bool{false, true} {
			t.Run(key+map[bool]string{false: "/missing", true: "/null"}[null], func(t *testing.T) {
				fields := nativeBiomeFields(t)
				delete(fields, key)
				if null {
					fields[key] = json.RawMessage(`null`)
				}
				if _, err := DecodeNativeBiomes(nativeBiomeJSON(t, fields)); err == nil || !strings.Contains(err.Error(), key) {
					t.Fatalf("missing %s accepted or poorly diagnosed: %v", key, err)
				}
			})
		}
	}
}

// TestNativeBiomesRejectInvalidValues checks typed scalar values and complete colour channels.
func TestNativeBiomesRejectInvalidValues(t *testing.T) {
	for _, scenario := range []struct {
		field string
		value json.RawMessage
	}{
		{"temperature", json.RawMessage(`1e1000`)},
		{"temperature", json.RawMessage(`"warm"`)},
		{"rain", json.RawMessage(`1`)},
		{"mapWaterColor", json.RawMessage(`"#123456"`)},
		{"mapWaterColor", json.RawMessage(`"11223344"`)},
		{"mapWaterColor", json.RawMessage(`"#1122334455"`)},
		{"mapWaterColor", json.RawMessage(`"#112233gg"`)},
		{"mapWaterColor", json.RawMessage(`{"a":17,"r":34,"g":51,"b":68}`)},
	} {
		t.Run(scenario.field+string(scenario.value), func(t *testing.T) {
			fields := nativeBiomeFields(t)
			fields[scenario.field] = scenario.value
			if _, err := DecodeNativeBiomes(nativeBiomeJSON(t, fields)); err == nil {
				t.Fatalf("accepted invalid %s: %s", scenario.field, scenario.value)
			}
		})
	}
	for _, input := range []string{
		`{}`, `null`, `[]`, `{"plains":{}}`, `{"minecraft:":{}}`, `{"minecraft:Bad Name":{}}`,
		`{"minecraft:plains":null}`, `{"minecraft:plains":{},"minecraft:plains":{}}`, `{} {}`,
	} {
		if _, err := DecodeNativeBiomes([]byte(input)); err == nil {
			t.Fatalf("accepted invalid native definitions %s", input)
		}
	}
}

// TestNativeBiomeIDsRejectIncompleteAndAmbiguousIdentities checks ID bounds and uniqueness.
func TestNativeBiomeIDsRejectIncompleteAndAmbiguousIdentities(t *testing.T) {
	got, err := DecodeNativeBiomeIDs([]byte(`{"minecraft:first":0,"minecraft:last":65535}`))
	if err != nil || !reflect.DeepEqual(got, map[string]int32{"minecraft:first": 0, "minecraft:last": 65535}) {
		t.Fatalf("IDs = %#v, error = %v", got, err)
	}
	for _, input := range []string{
		`{}`, `null`, `[]`, `{"plains":1}`, `{"minecraft:":1}`, `{"minecraft:Bad Name":1}`,
		`{"minecraft:plains":null}`, `{"minecraft:plains":-1}`, `{"minecraft:plains":65536}`,
		`{"minecraft:plains":1.5}`, `{"minecraft:plains":"1"}`,
		`{"minecraft:plains":1,"minecraft:other":1}`,
		`{"minecraft:plains":1,"minecraft:plains":2}`, `{"minecraft:plains":1} {}`,
	} {
		if _, err := DecodeNativeBiomeIDs([]byte(input)); err == nil {
			t.Fatalf("accepted invalid native identities %s", input)
		}
	}
}

// TestNativeBiomeReplacementRejectsDomainDrift checks missing and unexpected biome names.
func TestNativeBiomeReplacementRejectsDomainDrift(t *testing.T) {
	original := map[string]rawBiome{"minecraft:plains": {}, "minecraft:ocean": {}}
	for _, native := range []map[string]rawBiome{
		nil,
		{"minecraft:plains": {}},
		{"minecraft:plains": {}, "minecraft:unexpected": {}},
		{"minecraft:plains": {}, "minecraft:ocean": {}, "minecraft:unexpected": {}},
	} {
		if _, err := applyNativeBiomes(original, native); err == nil {
			t.Fatalf("accepted different native biome domain %#v", native)
		}
	}
	if _, err := applyNativeBiomes(nil, nil); err == nil {
		t.Fatal("accepted empty biome domains")
	}
	if _, err := applyNativeBiomes(original, maps.Clone(original)); err != nil {
		t.Fatalf("rejected matching complete domains: %v", err)
	}
}
