package entity

import (
	"sort"
	"testing"
)

func TestBreeding_AcceptsOnlyCanonicalFood(t *testing.T) {
	for _, test := range []struct {
		name string
		want bool
	}{
		{"minecraft:wheat", true}, {"wheat", false}, {"minecraft:Wheat", false},
		{"minecraft:carrot", false}, {"minecraft:wheat:0", false}, {"", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := Cow.Breeding.AcceptsFood(test.name); got != test.want {
				t.Fatalf("AcceptsFood(%q) = %t, want %t", test.name, got, test.want)
			}
		})
	}
	var unavailable *Breeding
	if unavailable.AcceptsFood("minecraft:wheat") {
		t.Fatal("nil descriptor accepts food")
	}
}

func TestEntity_BreedingUsesTheExistingRegistry(t *testing.T) {
	farmNames := []string{"minecraft:cow", "minecraft:mooshroom", "minecraft:sheep", "minecraft:pig", "minecraft:chicken", "minecraft:rabbit"}
	for _, name := range farmNames {
		value, ok := Lookup(name)
		if !ok || value.Breeding == nil || !value.Breeding.RequiresAdult || value.Breeding.RequiresTamed || value.Breeding.RequiresFullHealth {
			t.Fatalf("missing simple farm breeding prerequisites for %s: %+v", name, value.Breeding)
		}
	}
	for _, name := range []string{"minecraft:fox", "minecraft:panda", "minecraft:turtle", "minecraft:horse", "minecraft:donkey", "minecraft:ocelot"} {
		value, ok := Lookup(name)
		if !ok || value.Breeding != nil {
			t.Fatalf("unsupported eligibility exposed for %s: %+v", name, value.Breeding)
		}
	}
	supported := 0
	for _, value := range All() {
		resolved, ok := Lookup(value.Name)
		if !ok || resolved.Breeding != value.Breeding {
			t.Fatalf("lookup and enumeration disagree for %s", value.Name)
		}
		if value.Breeding == nil {
			continue
		}
		supported++
		if !sort.StringsAreSorted(value.Breeding.Foods) || len(value.Breeding.Foods) == 0 {
			t.Fatalf("invalid breeding foods for %s", value.Name)
		}
	}
	if supported != len(farmNames) {
		t.Fatalf("supported breeding descriptors = %d, want %d", supported, len(farmNames))
	}
}
