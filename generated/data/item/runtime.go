package item

import "sort"

// Runtime contains the vanilla registry entry and its inventory capabilities.
// Runtime IDs describe this data release; a connected server may provide a
// different palette, which must still be resolved for that session.
type Runtime struct {
	Name           string
	RuntimeID      int32
	ComponentBased bool
	Version        int32
	MaxCount       int
	BodyWearable   bool
	AllowOffHand   bool
	ArmourSlot     uint8
	ArmourWearable bool
}

// RuntimeItems returns all vanilla registry entries in name order. The returned
// slice is shared and must be treated as immutable.
func RuntimeItems() []Runtime { return runtimeItems }

// LookupRuntime finds an item by its namespaced identifier.
func LookupRuntime(name string) (Runtime, bool) {
	i := sort.Search(len(runtimeItems), func(i int) bool { return runtimeItems[i].Name >= name })
	if i == len(runtimeItems) || runtimeItems[i].Name != name {
		return Runtime{}, false
	}
	return runtimeItems[i], true
}
