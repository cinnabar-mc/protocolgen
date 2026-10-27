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

// TestNativePaletteJoinPreservesUnknownGeometry covers reordered and new states.
func TestNativePaletteJoinPreservesUnknownGeometry(t *testing.T) {
	native := nativeFixture(t)
	thickness, translucency, color := float32(0), float32(0.5), "#80112233"
	properties := native.Properties["minecraft:bed"]
	properties.Thickness, properties.Translucency, properties.MapColor = &thickness, &translucency, &color
	native.Properties["minecraft:bed"] = properties
	native.Names = []string{"minecraft:bed", "minecraft:bed"}
	newState := native.States[0]
	hash := uint32(8)
	newState.Hash = &hash
	native.States = append(native.States, newState)
	shape := json.RawMessage(`[[0,0,0,1,0.5625,1]]`)
	rows := []rawBlockState{{Name: "minecraft:bed", Hash: 7, MapColor: "#ffffffff", CollisionShape: shape}}
	got, err := alignNativeBlocks(rows, native)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Hash != 7 || got[1].Hash != 8 || string(got[0].CollisionShape) != string(shape) {
		t.Fatal("palette identities or measured supplemental shape changed")
	}
	if !got[1].supplementalUnavailable || got[1].CollisionShape != nil || got[1].OutlineShape != nil || got[1].MapColor != "#11223380" || got[1].Hardness != 0.2 {
		t.Fatal("new state lost native scalars or guessed supplemental properties")
	}
	native.Names[0] = "minecraft:other"
	if _, err := alignNativeBlocks(rows, native); err == nil {
		t.Fatal("same-hash different-name join accepted")
	}
}
