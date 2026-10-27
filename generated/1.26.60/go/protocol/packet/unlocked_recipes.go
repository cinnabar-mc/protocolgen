// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// UnlockedRecipes gives the client a list of recipes that have been unlocked, restricting the recipes that
// appear in the recipe book.
type UnlockedRecipes struct {
	// UnlockType is the type of unlock that the packet represents, and can either be adding or removing a list of
	// recipes. It is one of the constants listed above.
	PacketType protocol.PacketType
	// Recipes is a list of recipe names that have been unlocked.
	UnlockedRecipesList []string
}

// ID ...
func (*UnlockedRecipes) ID() uint32 {
	return IDUnlockedRecipes
}

func (pk *UnlockedRecipes) Marshal(io protocol.IO) {
	pk.PacketType.Marshal(io)
	protocol.FuncSlice(io, &pk.UnlockedRecipesList, io.Varuint32, io.String)
}
