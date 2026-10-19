package block

import "testing"

// TestMiningCoverageRequiresExplicitNativeFacts prevents untagged no-tool blocks from becoming unknown.
func TestMiningCoverageRequiresExplicitNativeFacts(t *testing.T) {
	for i := range StateCount() {
		state, _ := StateAt(i)
		block, _ := BlockAt(state.Block)
		facts, ok := MiningByName(block.Name)
		if !ok || facts.Evidence != MiningNative || facts.SpecialEvidence != MiningProvisional {
			t.Fatalf("missing or mislabeled mining evidence for %s", block.Name)
		}
	}
	for _, name := range []string{"minecraft:bed", "minecraft:glass"} {
		facts, ok := MiningByName(name)
		if !ok || facts.HarvestLevel != -1 || facts.HarvestTools != 0 || facts.EffectiveTools != 0 {
			t.Fatalf("no-tool block %s = %+v, found=%t", name, facts, ok)
		}
	}
	if _, ok := MiningByName("custom:unmeasured"); ok {
		t.Fatal("unknown block gained native evidence")
	}
}
