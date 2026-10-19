package generator

import (
	"encoding/json"
	"testing"

	"github.com/bedrock-mc/protocolgen/data/internal/nativeblock"
)

// nativeFixture builds the measured scalar fields while keeping the shape source independent.
func nativeFixture(t *testing.T) nativeblock.Catalog {
	t.Helper()
	var properties map[string]nativeblock.Properties
	var states map[string]nativeblock.State
	if err := json.Unmarshal([]byte(`{"minecraft:bed":{"requiresCorrectToolForDrops":false}}`), &properties); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(`{"0":{"hashedNetworkId":7,"destroySpeed":0.2,"explosionResistance":0.2,"friction":0.6,"lightEmission":0,"lightDamping":0,"flameOdds":0,"burnOdds":0,"solid":false}}`), &states); err != nil {
		t.Fatal(err)
	}
	catalog, err := nativeblock.Decode(properties, states, nil)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

// TestNativePropertiesOwnTheirFields preserves measured values without replacing shapes.
func TestNativePropertiesOwnTheirFields(t *testing.T) {
	catalog := nativeFixture(t)
	shape := json.RawMessage(`[0,0,0,1,0.5625,1]`)
	states := []rawBlockState{{Name: "minecraft:bed", Hash: 7, Hardness: 12, RequiresCorrectToolForDrops: true, CollisionShape: shape}}
	if err := applyNativeBlocks(states, catalog); err != nil {
		t.Fatal(err)
	}
	if states[0].Hardness != 0.2 || states[0].RequiresCorrectToolForDrops || string(states[0].CollisionShape) != string(shape) {
		t.Fatal("native scalar ownership or independent collision shape changed")
	}
}

// TestNativePropertyJoinRejectsIdentityDrift prevents joining by an ordinal alone.
func TestNativePropertyJoinRejectsIdentityDrift(t *testing.T) {
	catalog := nativeFixture(t)
	for _, states := range [][]rawBlockState{
		{{Name: "minecraft:bed", Hash: 8}},
		{{Name: "minecraft:missing", Hash: 7}},
		nil,
	} {
		if err := applyNativeBlocks(states, catalog); err == nil {
			t.Fatal("mismatched native property domain accepted")
		}
	}
}
