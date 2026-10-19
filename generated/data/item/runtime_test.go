package item

import "testing"

func TestRuntimeCatalogCoverageAndInventoryLimits(t *testing.T) {
	ids := make(map[int32]string)
	for _, item := range RuntimeItems() {
		if resolved, ok := LookupRuntime(item.Name); !ok || resolved != item {
			t.Fatalf("runtime entry cannot be resolved: %+v", item)
		}
		if previous, ok := ids[item.RuntimeID]; ok {
			t.Fatalf("%s and %s share ID %d", previous, item.Name, item.RuntimeID)
		}
		ids[item.RuntimeID] = item.Name
		if item.MaxCount < 1 || item.MaxCount > 64 {
			t.Fatalf("%s has invalid maximum stack count %d", item.Name, item.MaxCount)
		}
	}
	for _, test := range []struct {
		name    string
		count   int
		offhand bool
	}{{"photo_item", 1, true}, {"portfolio", 1, false}, {"straw_bed", 16, false}, {"item.straw_bed", 64, false}, {"black_cushion", 16, false}, {"red_concrete_slab", 64, false}, {"poplar_boat", 1, false}, {"poplar_sign", 16, false}} {
		got, ok := LookupRuntime("minecraft:" + test.name)
		if !ok || got.MaxCount != test.count || got.AllowOffHand != test.offhand {
			t.Errorf("%s = %+v, present=%t", test.name, got, ok)
		}
	}
}

// TestRuntimeTagsPreserveNativeMemberships checks known tags without changing registry comparability.
func TestRuntimeTagsPreserveNativeMemberships(t *testing.T) {
	if tags := RuntimeTags("custom:unknown"); tags != nil {
		t.Fatalf("unknown item tags = %v", tags)
	}
	found := false
	for _, tag := range RuntimeTags("minecraft:apple") {
		found = found || tag == "minecraft:is_food"
	}
	if !found {
		t.Fatalf("apple lost native food membership: %v", RuntimeTags("minecraft:apple"))
	}
	for name := range runtimeTags {
		if _, ok := LookupRuntime(name); !ok {
			t.Fatalf("tags reference unknown item %s", name)
		}
	}
}
