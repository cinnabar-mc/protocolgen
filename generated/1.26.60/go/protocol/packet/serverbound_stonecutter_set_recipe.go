// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

type ServerboundStonecutterSetRecipe struct {
	ContainerID uint8
	RecipeIndex int32
}

// ID ...
func (*ServerboundStonecutterSetRecipe) ID() uint32 {
	return IDServerboundStonecutterSetRecipe
}

func (pk *ServerboundStonecutterSetRecipe) Marshal(io protocol.IO) {
	io.Uint8(&pk.ContainerID)
	io.Varint32(&pk.RecipeIndex)
}
