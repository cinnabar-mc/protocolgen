package mininggen

import (
	"testing"

	"github.com/bedrock-mc/protocolgen/data/internal/nativeblock"
)

// TestNoToolRequirementIsKnown preserves hand mining even without an effective-tool tag.
func TestNoToolRequirementIsKnown(t *testing.T) {
	required := false
	fact, err := classify(nativeblock.Properties{RequiresCorrectToolForDrops: &required}, nil, 0)
	if err != nil || fact.level != -1 || fact.harvest != 0 || fact.effective != 0 {
		t.Fatalf("hand-harvestable block became unknown or tool-gated: %+v, %v", fact, err)
	}
}

// TestNativeTagsOwnHarvestRequirements rejects old material-based tier inference.
func TestNativeTagsOwnHarvestRequirements(t *testing.T) {
	required := true
	fact, err := classify(nativeblock.Properties{RequiresCorrectToolForDrops: &required}, []string{"minecraft:is_pickaxe_item_destructible", "minecraft:diamond_tier_destructible"}, shears)
	if err != nil || fact.level != 3 || fact.harvest != pickaxe || fact.effective != pickaxe|shears {
		t.Fatalf("native harvest and supplemental speed facts were conflated: %+v, %v", fact, err)
	}
	if _, err := classify(nativeblock.Properties{}, nil, 0); err == nil {
		t.Fatal("unknown tool requirement accepted as no requirement")
	}
}

// TestEffectiveToolDoesNotRequireThatToolToHarvest covers axe-effective hand-harvestable blocks.
func TestEffectiveToolDoesNotRequireThatToolToHarvest(t *testing.T) {
	required := false
	fact, err := classify(nativeblock.Properties{RequiresCorrectToolForDrops: &required}, []string{"minecraft:is_axe_item_destructible"}, 0)
	if err != nil || fact.effective != axe || fact.harvest != 0 || fact.level != -1 {
		t.Fatalf("effectiveness incorrectly restricted drops: %+v, %v", fact, err)
	}
}
