package runtimegen

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bedrock-mc/protocolgen/data/source"
)

// runtimeFixture supplies independent registry, property and supplemental-tag inputs.
func runtimeFixture() (map[string]RegistryEntry, map[string]Properties, map[string][]string, Corrections) {
	registry := map[string]RegistryEntry{
		"minecraft:apple": {RuntimeID: 1, Version: 2},
		"minecraft:new":   {RuntimeID: 2, ComponentBased: true},
	}
	properties := map[string]Properties{
		"minecraft:apple": {MaxStackSize: 64},
		"minecraft:new":   {Tags: []string{"minecraft:harness"}},
	}
	tags := map[string][]string{
		"minecraft:allow_offhand": {"minecraft:apple"},
		"minecraft:head":          {"minecraft:apple"},
	}
	return registry, properties, tags, Corrections{StackSizes: map[string]int{"minecraft:new": 16}, Offhand: []string{"minecraft:new"}, Evidence: map[string]string{"fixture": "synthetic"}}
}

func TestRuntimeCatalogCombinesSourcesDeterministically(t *testing.T) {
	registry, properties, tags, corrections := runtimeFixture()
	first, err := Generate(registry, properties, tags, corrections)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Generate(registry, properties, tags, corrections)
	if err != nil || !bytes.Equal(first, second) {
		t.Fatalf("generation is not deterministic: %v", err)
	}
	if !bytes.Contains(first, []byte(`Name: "minecraft:new", RuntimeID: 2, ComponentBased: true, Version: 0, MaxCount: 16, BodyWearable: true, AllowOffHand: true, ArmourSlot: 0, ArmourWearable: false`)) {
		t.Fatalf("registry, stack and capability inputs were not combined:\n%s", first)
	}
}

// TestNativePropertiesKeepInventoryFacts checks stack sizes and optional empty tags.
func TestNativePropertiesKeepInventoryFacts(t *testing.T) {
	registry := map[string]RegistryEntry{
		"minecraft:air": {RuntimeID: -158},
		"minecraft:new": {RuntimeID: 2},
	}
	properties, err := DecodeNativeProperties([]byte(`{
		"minecraft:air":{"id":-158,"maxStackSize":64,"maxDamage":0},
		"minecraft:new":{"id":2,"maxStackSize":1,"maxDamage":59,"tags":["minecraft:harness"]}
	}`), registry)
	if err != nil {
		t.Fatal(err)
	}
	if got := properties["minecraft:air"]; got.MaxStackSize != 64 || len(got.Tags) != 0 {
		t.Fatalf("empty-tag item facts = %#v", got)
	}
	if got := properties["minecraft:new"]; got.MaxStackSize != 1 || len(got.Tags) != 1 || got.Tags[0] != "minecraft:harness" {
		t.Fatalf("tagged item facts = %#v", got)
	}
}

// TestNativePropertiesRejectIncompleteOrMismatchedDumps checks complete registry coverage.
func TestNativePropertiesRejectIncompleteOrMismatchedDumps(t *testing.T) {
	registry := map[string]RegistryEntry{"minecraft:air": {RuntimeID: -158}}
	for _, scenario := range []struct {
		name, input, want string
	}{
		{"missing item", `{}`, "missing item minecraft:air"},
		{"unknown item", `{"minecraft:new":{"id":2,"maxStackSize":1,"maxDamage":0}}`, "unknown item minecraft:new"},
		{"null item", `{"minecraft:air":null}`, "incomplete properties"},
		{"missing ID", `{"minecraft:air":{"maxStackSize":64,"maxDamage":0}}`, "incomplete properties"},
		{"wrong ID", `{"minecraft:air":{"id":0,"maxStackSize":64,"maxDamage":0}}`, "registry has -158"},
		{"missing size", `{"minecraft:air":{"id":-158,"maxDamage":0}}`, "incomplete properties"},
		{"zero size", `{"minecraft:air":{"id":-158,"maxStackSize":0,"maxDamage":0}}`, "invalid stack size"},
		{"oversized stack", `{"minecraft:air":{"id":-158,"maxStackSize":65,"maxDamage":0}}`, "invalid stack size"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			if _, err := DecodeNativeProperties([]byte(scenario.input), registry); err == nil || !strings.Contains(err.Error(), scenario.want) {
				t.Fatalf("error = %v, want %q", err, scenario.want)
			}
		})
	}
}

// TestNativeTagsDoNotInventInventoryCapabilities keeps unsupported slots on supplemental facts.
func TestNativeTagsDoNotInventInventoryCapabilities(t *testing.T) {
	registry, _, tags, _ := runtimeFixture()
	for _, scenario := range []struct {
		tag  string
		body bool
	}{
		{"minecraft:horse_armor", true},
		{"minecraft:nautilus_armor", true},
		{"minecraft:harness", true},
		{"minecraft:is_armor", false},
		{"minecraft:head", false},
		{"minecraft:allow_offhand", false},
	} {
		t.Run(scenario.tag, func(t *testing.T) {
			input := fmt.Sprintf(`{
				"minecraft:apple":{"id":1,"maxStackSize":64,"maxDamage":0},
				"minecraft:new":{"id":2,"maxStackSize":1,"maxDamage":59,"tags":[%q],"allowOffHand":true,"armourSlot":0}
			}`, scenario.tag)
			properties, err := DecodeNativeProperties([]byte(input), registry)
			if err != nil {
				t.Fatal(err)
			}
			output, err := Generate(registry, properties, tags, Corrections{})
			if err != nil {
				t.Fatal(err)
			}
			want := fmt.Sprintf(`Name: "minecraft:new", RuntimeID: 2, ComponentBased: true, Version: 0, MaxCount: 1, BodyWearable: %t, AllowOffHand: false, ArmourSlot: 0, ArmourWearable: false`, scenario.body)
			if !bytes.Contains(output, []byte(want)) {
				t.Fatalf("native facts fabricated or lost inventory capabilities:\n%s", output)
			}
		})
	}
}

// TestRuntimeSourcesUseNativeProperties exercises locked native input with supplemental capability tags.
func TestRuntimeSourcesUseNativeProperties(t *testing.T) {
	cache := t.TempDir()
	inputs := &source.Sources{Lock: source.Lock{Release: "1.26.51", Target: source.Release{ProtocolVersion: 2193}, Inputs: map[string]source.Input{}}}
	for name, data := range map[string]string{
		"native_version": `{"protocol":2193,"version":"1.26.51"}`,
		"item_registry":  `[{"name":"minecraft:apple","id":1,"componentBased":false,"version":2},{"name":"minecraft:new","id":2,"componentBased":true,"version":0}]`,
		"native_item_properties": `{
			"minecraft:apple":{"id":1,"maxStackSize":64,"maxDamage":0},
			"minecraft:new":{"id":2,"maxStackSize":64,"maxDamage":59,"tags":["minecraft:harness"]}
		}`,
		"native_item_tags": `{"minecraft:harness":["minecraft:new"]}`,
		"item_tags":        `{"minecraft:allow_offhand":["minecraft:apple"],"minecraft:head":["minecraft:apple"]}`,
		"item_corrections": `{"stack_limits":{"minecraft:new":{"expected_source_count":64,"effective_count":1}},"offhand":["minecraft:new"],"evidence":{"fixture":"synthetic"}}`,
	} {
		digest := fmt.Sprintf("%x", sha256.Sum256([]byte(data)))
		if err := os.WriteFile(filepath.Join(cache, digest), []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
		inputs.Lock.Inputs[name] = source.Input{SHA256: digest}
	}
	output, err := FromSources(context.Background(), inputs, cache)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(output, []byte(`MaxCount: 1, BodyWearable: true, AllowOffHand: true`)) {
		t.Fatalf("native facts or supplemental capability correction were lost:\n%s", output)
	}
}

// TestStackLimitOverridesRequireUnchangedRawFacts guards measured virtual-getter corrections.
func TestStackLimitOverridesRequireUnchangedRawFacts(t *testing.T) {
	registry, properties, tags, _ := runtimeFixture()
	properties["minecraft:new"] = Properties{MaxStackSize: 64}
	for _, scenario := range []struct {
		name  string
		limit StackLimit
		edit  func(map[string]RegistryEntry, map[string]Properties, *Corrections)
		want  string
	}{
		{"measured limit", StackLimit{64, 1}, nil, ""},
		{"changed raw field", StackLimit{16, 1}, nil, "is stale"},
		{"no longer needed", StackLimit{64, 64}, nil, "is stale"},
		{"invalid expected", StackLimit{0, 1}, nil, "invalid counts"},
		{"invalid effective", StackLimit{64, 0}, nil, "invalid counts"},
		{"oversized effective", StackLimit{64, 65}, nil, "invalid counts"},
		{"missing evidence", StackLimit{64, 1}, func(_ map[string]RegistryEntry, _ map[string]Properties, c *Corrections) {
			c.Evidence = nil
		}, "need evidence"},
		{"removed item", StackLimit{64, 1}, func(r map[string]RegistryEntry, _ map[string]Properties, _ *Corrections) {
			delete(r, "minecraft:new")
		}, "unknown item"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			corrections := Corrections{
				StackLimits: map[string]StackLimit{"minecraft:new": scenario.limit},
				Evidence:    map[string]string{"fixture": "synthetic measured effective capacity"},
			}
			r, p := maps.Clone(registry), maps.Clone(properties)
			if scenario.edit != nil {
				scenario.edit(r, p, &corrections)
			}
			output, err := Generate(r, p, tags, corrections)
			if scenario.want != "" {
				if err == nil || !strings.Contains(err.Error(), scenario.want) {
					t.Fatalf("error = %v, want %q", err, scenario.want)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Contains(output, []byte(`Name: "minecraft:new", RuntimeID: 2, ComponentBased: true, Version: 0, MaxCount: 1`)) {
				t.Fatalf("effective capacity was not preserved:\n%s", output)
			}
		})
	}
}

func TestRuntimeGeneratorHandlesEverySchemaField(t *testing.T) {
	r, p, tags, c := runtimeFixture()
	output, err := Generate(r, p, tags, c)
	if err != nil {
		t.Fatal(err)
	}
	file, err := parser.ParseFile(token.NewFileSet(), "runtime_generated.go", output, 0)
	if err != nil {
		t.Fatal(err)
	}
	// Read the sibling catalog schema without making either module import the other.
	definition, err := parser.ParseFile(token.NewFileSet(), "../../../generated/data/item/runtime.go", nil, 0)
	if err != nil {
		t.Fatalf("read Runtime schema from the repository checkout: %v", err)
	}
	schema := make(map[string]bool)
	ast.Inspect(definition, func(node ast.Node) bool {
		spec, ok := node.(*ast.TypeSpec)
		if !ok || spec.Name.Name != "Runtime" {
			return true
		}
		structure, ok := spec.Type.(*ast.StructType)
		if !ok {
			t.Fatal("Runtime must be a struct")
		}
		for _, field := range structure.Fields.List {
			if len(field.Names) == 0 {
				t.Fatal("Runtime has an embedded field that the generator cannot handle")
			}
			for _, name := range field.Names {
				schema[name.Name] = true
			}
		}
		return false
	})
	if len(schema) == 0 {
		t.Fatal("Runtime schema was not found or has no fields")
	}
	rows := 0
	ast.Inspect(file, func(node ast.Node) bool {
		literal, ok := node.(*ast.CompositeLit)
		if !ok || literal.Type != nil {
			return true
		}
		rows++
		fields := make(map[string]bool)
		for _, element := range literal.Elts {
			pair, ok := element.(*ast.KeyValueExpr)
			if !ok {
				t.Fatal("Runtime rows must name every field")
			}
			name := pair.Key.(*ast.Ident).Name
			if !schema[name] {
				t.Errorf("generator emits unknown Runtime.%s", name)
			}
			fields[name] = true
		}
		for name := range schema {
			if !fields[name] {
				t.Errorf("generator silently omitted Runtime.%s", name)
			}
		}
		return false
	})
	if rows != len(r) {
		t.Fatalf("checked %d Runtime rows, want %d", rows, len(r))
	}
}

func TestRuntimeCatalogRejectsMissingAndStaleFacts(t *testing.T) {
	for _, scenario := range []struct {
		name string
		edit func(map[string]RegistryEntry, map[string]Properties, map[string][]string, *Corrections)
		want string
	}{
		{"missing size", func(_ map[string]RegistryEntry, _ map[string]Properties, _ map[string][]string, c *Corrections) {
			delete(c.StackSizes, "minecraft:new")
		}, "missing or invalid stack size"},
		{"stale size", func(_ map[string]RegistryEntry, p map[string]Properties, _ map[string][]string, _ *Corrections) {
			p["minecraft:new"] = Properties{MaxStackSize: 16}
		}, "is stale"},
		{"removed item", func(r map[string]RegistryEntry, _ map[string]Properties, _ map[string][]string, _ *Corrections) {
			delete(r, "minecraft:new")
		}, "unknown item"},
		{"stale tag", func(_ map[string]RegistryEntry, _ map[string]Properties, tags map[string][]string, _ *Corrections) {
			tags["minecraft:allow_offhand"] = append(tags["minecraft:allow_offhand"], "minecraft:new")
		}, "stale or duplicated"},
		{"missing tag", func(_ map[string]RegistryEntry, _ map[string]Properties, tags map[string][]string, _ *Corrections) {
			delete(tags, "minecraft:head")
		}, "missing or empty"},
		{"duplicate ID", func(r map[string]RegistryEntry, _ map[string]Properties, _ map[string][]string, _ *Corrections) {
			r["minecraft:apple"] = r["minecraft:new"]
		}, "share runtime ID"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			r, p, tags, c := runtimeFixture()
			scenario.edit(r, p, tags, &c)
			if _, err := Generate(r, p, tags, c); err == nil || !strings.Contains(err.Error(), scenario.want) {
				t.Fatalf("error = %v, want %q", err, scenario.want)
			}
		})
	}
}

// TestDecodeRegistry preserves zero values and rejects incomplete or duplicate identities.
func TestDecodeRegistry(t *testing.T) {
	const valid = `{"name":"minecraft:air","id":-158,"version":0,"componentBased":false}`
	rows, err := DecodeRegistry([]byte(`[` + valid + `]`))
	if err != nil || len(rows) != 1 || rows["minecraft:air"] != (RegistryEntry{RuntimeID: -158}) {
		t.Fatalf("registry = %v, error = %v", rows, err)
	}
	for _, input := range []string{
		`[` + valid + `,` + valid + `]`,
		`[{"name":"minecraft:air","id":-158,"version":0}]`,
		`[{"name":"minecraft:air","componentBased":false,"version":0}]`,
		`[{"name":"minecraft:air","id":-158,"componentBased":false}]`,
	} {
		if _, err := DecodeRegistry([]byte(input)); err == nil {
			t.Fatalf("accepted invalid registry %s", input)
		}
	}
}

// TestNativeTagMembershipRejectsDrift prevents either view of the dump from silently losing members.
func TestNativeTagMembershipRejectsDrift(t *testing.T) {
	properties := map[string]Properties{"minecraft:apple": {Tags: []string{"minecraft:food"}}}
	for _, data := range []string{`{}`, `{"minecraft:food":["minecraft:new"]}`, `{"minecraft:food":["minecraft:apple","minecraft:apple"]}`} {
		if err := validateNativeTags(properties, []byte(data)); err == nil {
			t.Fatalf("accepted inconsistent native tags %s", data)
		}
	}
	if err := validateNativeTags(properties, []byte(`{"minecraft:food":["minecraft:apple"]}`)); err != nil {
		t.Fatal(err)
	}
}
