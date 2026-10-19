package nativeblock

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/bedrock-mc/protocolgen/data/source"
)

// fixture decodes one complete measurement with legitimate zero and false values.
func fixture(t *testing.T) (map[string]Properties, map[string]State) {
	t.Helper()
	var properties map[string]Properties
	var states map[string]State
	if err := json.Unmarshal([]byte(`{"minecraft:bed":{"requiresCorrectToolForDrops":false}}`), &properties); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(`{"0":{"hashedNetworkId":1,"destroySpeed":0.2,"explosionResistance":0.2,"friction":0.6,"lightEmission":0,"lightDamping":0,"flameOdds":0,"burnOdds":0,"solid":false}}`), &states); err != nil {
		t.Fatal(err)
	}
	return properties, states
}

// TestDecodeSeparatesAbsentValuesFromKnownZero prevents a no-tool block becoming unknown.
func TestDecodeSeparatesAbsentValuesFromKnownZero(t *testing.T) {
	properties, states := fixture(t)
	got, err := Decode(properties, states, nil)
	if err != nil {
		t.Fatal(err)
	}
	if *got.Properties["minecraft:bed"].RequiresCorrectToolForDrops || *got.States[0].Solid || *got.States[0].LightEmission != 0 {
		t.Fatal("known false and zero fields changed")
	}
	state := states["0"]
	state.Solid = nil
	states["0"] = state
	if _, err := Decode(properties, states, nil); err == nil {
		t.Fatal("missing solid measurement accepted as false")
	}
}

// TestDecodeRejectsDomainDrift guards sparse tags and contiguous state ordinals.
func TestDecodeRejectsDomainDrift(t *testing.T) {
	properties, states := fixture(t)
	if _, err := Decode(properties, states, map[string][]string{"minecraft:missing": {"minecraft:is_pickaxe_item_destructible"}}); err == nil {
		t.Fatal("tag outside the property domain accepted")
	}
	if _, err := Decode(properties, states, map[string][]string{"minecraft:bed": {"wood", "wood"}}); err == nil {
		t.Fatal("duplicate tag accepted")
	}
	states["1"] = states["0"]
	if _, err := Decode(properties, states, nil); err == nil {
		t.Fatal("duplicate state identity accepted")
	}
	delete(states, "0")
	if _, err := Decode(properties, states, nil); err == nil {
		t.Fatal("noncontiguous state ordinals accepted")
	}
}

// TestBundledNativeSnapshot verifies the authenticated measurements offline.
func TestBundledNativeSnapshot(t *testing.T) {
	sources, err := source.Open("")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := Read(context.Background(), sources, "")
	if err != nil {
		t.Fatal(err)
	}
	if *catalog.Properties["minecraft:bed"].RequiresCorrectToolForDrops {
		t.Fatal("bed requires a correct tool")
	}
	if !*catalog.Properties["minecraft:obsidian"].RequiresCorrectToolForDrops {
		t.Fatal("obsidian lost its tool requirement")
	}
}

// TestNativeSnapshotRejectsDifferentRelease prevents a protocol-compatible dump from being relabeled silently.
func TestNativeSnapshotRejectsDifferentRelease(t *testing.T) {
	for _, change := range []func(*source.Lock){
		func(lock *source.Lock) { lock.Release = "other" },
		func(lock *source.Lock) { lock.Target.ProtocolVersion++ },
	} {
		inputs, err := source.Open("")
		if err != nil {
			t.Fatal(err)
		}
		change(&inputs.Lock)
		if err := inputs.ValidateNative(context.Background(), ""); err == nil {
			t.Fatal("different native snapshot accepted")
		}
	}
}
