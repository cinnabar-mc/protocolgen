// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// PlayerHotBar is sent by the server to the client. It used to be used to link hot bar slots of the player to
// actual slots in the inventory, but as of 1.2, this was changed and hot bar slots are no longer a free
// floating part of the inventory. Since 1.2, the packet has been re-purposed, but its new functionality is
// not clear.
type PlayerHotbar struct {
	// SelectedHotBarSlot ...
	SelectedSlot uint32
	// WindowID ...
	ContainerID uint8
	// SelectHotBarSlot ...
	ShouldSelectSlot bool
}

// ID ...
func (*PlayerHotbar) ID() uint32 {
	return IDPlayerHotbar
}

func (pk *PlayerHotbar) Marshal(io protocol.IO) {
	io.Varuint32(&pk.SelectedSlot)
	io.Uint8(&pk.ContainerID)
	io.Bool(&pk.ShouldSelectSlot)
}
