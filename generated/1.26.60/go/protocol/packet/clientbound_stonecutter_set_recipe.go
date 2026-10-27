// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

type ClientboundStonecutterSetRecipe struct {
	PlayerID    int64
	ContainerID uint8
	RecipeIndex int32
}

// ID ...
func (*ClientboundStonecutterSetRecipe) ID() uint32 {
	return IDClientboundStonecutterSetRecipe
}

func (pk *ClientboundStonecutterSetRecipe) Marshal(io protocol.IO) {
	io.ActorUniqueID(&pk.PlayerID)
	io.Uint8(&pk.ContainerID)
	io.Varint32(&pk.RecipeIndex)
}
