// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// TakeItemActor is sent by the server when a player picks up an item entity. It makes the item entity
// disappear to viewers and shows the pick-up animation.
type TakeItemActor struct {
	// ItemEntityRuntimeID is the entity runtime ID of the item that is being taken by another entity. It will
	// disappear to viewers after showing the pick-up animation.
	ItemRuntimeID uint64
	// TakerEntityRuntimeID is the runtime ID of the entity that took the item, which is usually a player, but
	// could be another entity like a zombie too.
	ActorRuntimeID uint64
}

// ID ...
func (*TakeItemActor) ID() uint32 {
	return IDTakeItemActor
}

func (pk *TakeItemActor) Marshal(io protocol.IO) {
	io.ActorRuntimeID(&pk.ItemRuntimeID)
	io.ActorRuntimeID(&pk.ActorRuntimeID)
}
