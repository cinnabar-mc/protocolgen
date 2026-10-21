// Package entity provides generated values extracted from vanilla Bedrock entity
// definitions. Breeding is a conservative projection of reviewed lifecycle
// transitions, rather than a general model of events, filters, or goals.
package entity

import "sort"

// Float is an optional floating-point value.
type Float struct {
	Value   float32
	Present bool
}

// Bool is an optional boolean value.
type Bool struct {
	Value   bool
	Present bool
}

// Number is either a fixed value or a random range from a vanilla component.
type Number struct {
	Value   float32
	Min     float32
	Max     float32
	Ranged  bool
	Present bool
}

// CollisionBox contains the base collision dimensions of an entity.
type CollisionBox struct {
	Width   float32
	Height  float32
	Present bool
}

// Health contains the initial health expression and an optional explicit
// maximum. A ranged initial value is represented by Initial.Min/Max.
type Health struct {
	Initial Number
	Maximum Float
	Present bool
}

// Breeding describes supported vanilla breeding foods and observable prerequisites.
// Foods contains sorted, namespaced item identifiers. The descriptor and its slice
// must be treated as immutable. A descriptor does not establish readiness: the
// server owns breeding cooldowns and may reject a feeding attempt. Consumers must
// also exclude dead animals and animals already in love using current server facts.
type Breeding struct {
	Foods              []string
	RequiresAdult      bool
	RequiresTamed      bool
	RequiresFullHealth bool
}

// AcceptsFood reports whether name is an exact, namespaced breeding food.
// A nil descriptor accepts nothing; unqualified names and item aliases are not matched.
func (b *Breeding) AcceptsFood(name string) bool {
	if b == nil {
		return false
	}
	index := sort.SearchStrings(b.Foods, name)
	return index < len(b.Foods) && b.Foods[index] == name
}

// Entity contains stable base-component values and optional breeding metadata.
type Entity struct {
	Name              string
	SpawnCategory     string
	Experimental      bool
	Summonable        bool
	Spawnable         bool
	CollisionBox      CollisionBox
	Health            Health
	Movement          Number
	Scale             Number
	ProjectileGravity Float
	ProjectilePower   Float
	ProjectileBounce  Bool
	Pushable          Bool
	PushableByPiston  Bool
	HasPhysics        bool
	Families          []string
	// Breeding is nil when applicability or additional conditions are unsupported.
	Breeding *Breeding
}

// All returns every entity specification in name order. The returned data must
// be treated as immutable.
func All() []Entity {
	return all
}

// Lookup finds an entity specification by its namespaced identifier.
func Lookup(name string) (Entity, bool) {
	index := sort.Search(len(all), func(i int) bool { return all[i].Name >= name })
	if index == len(all) || all[index].Name != name {
		return Entity{}, false
	}
	return all[index], true
}
