package generator

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/bedrock-mc/protocolgen/data/internal/runtimegen"
)

// breedingFixture is an authored lifecycle, not a copy of a vanilla source asset.
func breedingFixture(t *testing.T) rawEntity {
	t.Helper()
	var raw rawEntity
	if err := json.Unmarshal([]byte(`{"minecraft:entity": {
		"description": {"identifier": "minecraft:test_animal"},
		"component_groups": {
			"baby": {"minecraft:is_baby": {}, "minecraft:ageable": {"grow_up": {"event": "grow", "target": "self"}}},
			"adult": {"minecraft:breedable": {"require_tame": false, "breed_items": ["wheat", "minecraft:carrot"], "breeds_with": {"minecraft:test_animal": {}}}},
			"coat": {"minecraft:variant": {"value": 1}}
		},
		"events": {
			"minecraft:entity_spawned": {"sequence": [
				{"randomize": [{"add": {"component_groups": ["baby"]}}, {"trigger": "adult_spawn"}]},
				{"filters": {"test": "has_biome_tag", "value": "cold"}, "add": {"component_groups": ["coat"]}}
			]},
			"adult_spawn": {"add": {"component_groups": ["adult"]}},
			"minecraft:entity_born": {"add": {"component_groups": ["baby"]}},
			"grow": {"remove": {"component_groups": ["baby"]}, "add": {"component_groups": ["adult"]}}
		}
	}}`), &raw); err != nil {
		t.Fatal(err)
	}
	return raw
}

// breedingItems supplies independently stated canonical identifiers for fixtures.
func breedingItems() map[string]runtimegen.RegistryEntry {
	return map[string]runtimegen.RegistryEntry{"minecraft:wheat": {}, "minecraft:carrot": {}}
}

// editBreedable edits a fixture component without duplicating its other facts.
func editBreedable(t *testing.T, raw *rawEntity, edit func(map[string]any)) {
	t.Helper()
	var fields map[string]any
	if err := json.Unmarshal(raw.Entity.Groups["adult"]["minecraft:breedable"], &fields); err != nil {
		t.Fatal(err)
	}
	edit(fields)
	encoded, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	raw.Entity.Groups["adult"]["minecraft:breedable"] = encoded
}

func TestEntityBreeding_ProvesAdultFoodApplicability(t *testing.T) {
	value, owner, reason := projectBreeding(breedingFixture(t), breedingItems())
	want := &projectedBreeding{Foods: []string{"minecraft:carrot", "minecraft:wheat"}, RequiresAdult: true}
	if reason != "" || owner != "adult" || !reflect.DeepEqual(value, want) {
		t.Fatalf("projection = %+v, %q, %q; want %+v", value, owner, reason, want)
	}
}

func TestEntityBreeding_UntamedSittingRequiresExplicitPermission(t *testing.T) {
	for _, component := range []string{
		"minecraft:sittable",
		"minecraft:behavior.stay_while_sitting",
		"minecraft:behavior.random_sitting",
		"minecraft:behavior.random_look_around_and_sit",
		"minecraft:behavior.ocelot_sit_on_block",
	} {
		for _, location := range []string{"base", "adult", "coat"} {
			for _, permission := range []struct {
				name    string
				allowed *bool
			}{
				{"default", nil},
				{"false", new(false)},
				{"true", new(true)},
			} {
				t.Run(component+"/"+location+"/"+permission.name, func(t *testing.T) {
					raw := breedingFixture(t) // require_tame is explicitly false.
					if location == "base" {
						raw.Entity.Components = map[string]json.RawMessage{component: json.RawMessage(`{}`)}
					} else {
						raw.Entity.Groups[location][component] = json.RawMessage(`{}`)
					}
					if permission.allowed != nil {
						editBreedable(t, &raw, func(m map[string]any) { m["allow_sitting"] = *permission.allowed })
					}
					value, owner, reason := projectBreeding(raw, breedingItems())
					if permission.allowed == nil || !*permission.allowed {
						if value != nil || owner != "adult" || reason != "unrepresented sitting prerequisite" {
							t.Fatalf("unsafe sitting projection = %+v, %q, %q", value, owner, reason)
						}
						return
					}
					want := &projectedBreeding{Foods: []string{"minecraft:carrot", "minecraft:wheat"}, RequiresAdult: true}
					if reason != "" || owner != "adult" || !reflect.DeepEqual(value, want) {
						t.Fatalf("sitting permitted projection = %+v, %q, %q; want %+v", value, owner, reason, want)
					}
				})
			}
		}
	}
}

func TestEntityBreeding_NonSittingAllowsStandingPrerequisite(t *testing.T) {
	raw := breedingFixture(t)
	editBreedable(t, &raw, func(m map[string]any) { m["allow_sitting"] = false })
	value, owner, reason := projectBreeding(raw, breedingItems())
	want := &projectedBreeding{Foods: []string{"minecraft:carrot", "minecraft:wheat"}, RequiresAdult: true}
	if reason != "" || owner != "adult" || !reflect.DeepEqual(value, want) {
		t.Fatalf("non-sitting projection = %+v, %q, %q; want %+v", value, owner, reason, want)
	}
}

func TestEntityBreeding_RejectsUnobservableConditions(t *testing.T) {
	for _, test := range []struct {
		name string
		edit func(map[string]any)
	}{
		{"environment", func(m map[string]any) {
			m["environment_requirements"] = map[string]any{"blocks": "bamboo", "count": 8, "radius": 5}
		}},
		{"love filter", func(m map[string]any) {
			m["love_filters"] = map[string]any{"test": "enum_property", "value": "unrolled"}
		}},
		{"pregnancy", func(m map[string]any) { m["causes_pregnancy"] = true }},
		{"unknown future constraint", func(m map[string]any) { m["new_condition"] = true }},
		{"item result transformation", func(m map[string]any) {
			m["breed_items"] = []any{map[string]any{"item": "wheat", "result_item": "bowl"}}
		}},
		{"legacy item alias", func(m map[string]any) { m["breed_items"] = []string{"muttonRaw"} }},
		{"duplicate food", func(m map[string]any) { m["breed_items"] = []string{"wheat", "minecraft:wheat"} }},
		{"missing tame defaults true", func(m map[string]any) { delete(m, "require_tame") }},
		{"tame with unrepresented sitting", func(m map[string]any) { m["require_tame"] = true }},
		{"invalid full-health prerequisite", func(m map[string]any) { m["require_full_health"] = nil }},
		{"partner event", func(m map[string]any) {
			m["breeds_with"] = map[string]any{"minecraft:test_animal": map[string]any{"event": "pregnant"}}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			raw := breedingFixture(t)
			editBreedable(t, &raw, test.edit)
			if value, _, reason := projectBreeding(raw, breedingItems()); value != nil || reason == "" {
				t.Fatalf("unsafe projection = %+v, reason=%q", value, reason)
			}
		})
	}
}

func TestEntityBreeding_ProjectsExplicitTameAndHealthPrerequisites(t *testing.T) {
	raw := breedingFixture(t)
	editBreedable(t, &raw, func(m map[string]any) {
		delete(m, "require_tame") // The documented default must remain true.
		m["require_full_health"] = true
		m["allow_sitting"] = true
	})
	value, _, reason := projectBreeding(raw, breedingItems())
	if reason != "" || value == nil || !value.RequiresAdult || !value.RequiresTamed || !value.RequiresFullHealth {
		t.Fatalf("lost explicit prerequisites: %+v, reason=%q", value, reason)
	}
}

func TestEntityBreeding_RejectsUnprovenComponentOwners(t *testing.T) {
	for _, test := range []struct {
		name string
		edit func(*rawEntity)
	}{
		{"conditional growth", func(raw *rawEntity) {
			raw.Entity.Events["grow"] = json.RawMessage(`{"filters":{"test":"is_tamed","value":true},"remove":{"component_groups":["baby"]},"add":{"component_groups":["adult"]}}`)
		}},
		{"trusting owner outside growth", func(raw *rawEntity) {
			raw.Entity.Groups["trusting"] = raw.Entity.Groups["adult"]
			raw.Entity.Groups["adult"] = map[string]json.RawMessage{}
		}},
		{"multiple food owners", func(raw *rawEntity) { raw.Entity.Groups["other"] = raw.Entity.Groups["adult"] }},
		{"adult owner removed conditionally", func(raw *rawEntity) {
			raw.Entity.Events["conditional"] = json.RawMessage(`{"filters":{"test":"bool_property","value":true},"remove":{"component_groups":["adult"]}}`)
		}},
		{"unproven spawn path", func(raw *rawEntity) {
			raw.Entity.Events["minecraft:entity_spawned"] = json.RawMessage(`{"randomize":[{"add":{"component_groups":["adult"]}},{}]}`)
		}},
		{"wrong born path", func(raw *rawEntity) {
			raw.Entity.Events["minecraft:entity_born"] = json.RawMessage(`{"add":{"component_groups":["adult"]}}`)
		}},
		{"recursive event", func(raw *rawEntity) { raw.Entity.Events["loop"] = json.RawMessage(`{"trigger":"loop"}`) }},
		{"unknown action", func(raw *rawEntity) {
			raw.Entity.Events["commands"] = json.RawMessage(`{"queue_command":{"command":"event entity @s change"}}`)
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			raw := breedingFixture(t)
			test.edit(&raw)
			if value, _, reason := projectBreeding(raw, breedingItems()); value != nil || reason == "" {
				t.Fatalf("unsafe projection = %+v, reason=%q", value, reason)
			}
		})
	}
}

func TestEntityBreeding_ProjectionPreservesBaseAndClearsStaleFacts(t *testing.T) {
	raw := breedingFixture(t)
	document, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	inputs := EntityBreedingInputs{
		Definitions: map[string][]byte{"behavior_pack/entities/test_animal.json": document},
		Items:       breedingItems(), SourceLockSHA256: strings.Repeat("1", 64),
		Repository: "Mojang/bedrock-samples", Revision: strings.Repeat("2", 40), ArchiveSHA256: strings.Repeat("3", 64),
	}
	files := map[string][]byte{
		"entity/test_animal_generated.go": []byte(generatedSource("entity", `var TestAnimal = Entity{Name: "minecraft:test_animal", HasPhysics: true}`)),
		"entity/index_generated.go":       []byte(generatedSource("entity", `var all = []Entity{TestAnimal}`)),
	}
	first, count, err := GenerateEntityBreeding(files, inputs)
	if err != nil || count != 1 {
		t.Fatalf("GenerateEntityBreeding = %d, %v", count, err)
	}
	assertContains(t, first, "entity/test_animal_generated.go", "HasPhysics: true")
	assertContains(t, first, "entity/test_animal_generated.go", "RequiresAdult: true")
	if !bytes.Equal(files["entity/index_generated.go"], first["entity/index_generated.go"]) {
		t.Fatal("projection changed entity registry")
	}
	second, _, err := GenerateEntityBreeding(first, inputs)
	if err != nil || !reflect.DeepEqual(first, second) {
		t.Fatalf("regeneration is not deterministic: %v", err)
	}
	editBreedable(t, &raw, func(m map[string]any) { m["love_filters"] = map[string]any{} })
	inputs.Definitions["behavior_pack/entities/test_animal.json"], err = json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	cleared, count, err := GenerateEntityBreeding(first, inputs)
	if err != nil || count != 0 {
		t.Fatalf("clear stale projection = %d, %v", count, err)
	}
	assertNotContains(t, cleared, "entity/test_animal_generated.go", "Breeding:")
	assertContains(t, cleared, "entity/test_animal_generated.go", "HasPhysics: true")
}

func TestGenerate_AppliesBreedingProjectionToBaseEntities(t *testing.T) {
	raw := breedingFixture(t)
	raw.Entity.Description.Identifier = "minecraft:donkey"
	editBreedable(t, &raw, func(m map[string]any) {
		m["breeds_with"] = map[string]any{"minecraft:donkey": map[string]any{}}
	})
	document, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	inputs := EntityBreedingInputs{
		Definitions: map[string][]byte{"behavior_pack/entities/donkey.json": document},
		Items:       breedingItems(), SourceLockSHA256: strings.Repeat("1", 64),
		Repository: "Mojang/bedrock-samples", Revision: strings.Repeat("2", 40), ArchiveSHA256: strings.Repeat("3", 64),
	}
	files, stats, err := Generate(Config{
		CloudburstDir: filepath.Join("testdata", "cloudburst"), BDSDir: filepath.Join("testdata", "bds"),
		BiomeIDs: map[string]int32{"minecraft:plains": 1}, EntityBreeding: &inputs,
	})
	if err != nil {
		t.Fatal(err)
	}
	if stats.Entities != 2 {
		t.Fatalf("projection changed the base entity count: %d", stats.Entities)
	}
	assertContains(t, files, "entity/donkey_generated.go", "Min: 15")
	assertContains(t, files, "entity/donkey_generated.go", "RequiresAdult: true")
	assertNotContains(t, files, "entity/arrow_generated.go", "Breeding:")
}
