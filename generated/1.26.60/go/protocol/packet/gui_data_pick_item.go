// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// GUIDataPickItem is sent by the server to make the client 'select' a hot bar slot. It currently appears to
// be broken however, and does not actually set the selected slot to the hot bar slot set in the packet.
type GuiDataPickItem struct {
	// ItemName is the name of the item that shows up in the top part of the popup that shows up when selecting an
	// item. It is shown as if an item was selected by the player itself.
	ItemName string
	// ItemEffects is the line under the ItemName, where the effects of the item are usually situated.
	ItemEffectName string
	// HotBarSlot is the hot bar slot to be selected/picked. This does not currently work, so it does not matter
	// what number this is.
	Slot int32
}

// ID ...
func (*GuiDataPickItem) ID() uint32 {
	return IDGuiDataPickItem
}

func (pk *GuiDataPickItem) Marshal(io protocol.IO) {
	io.String(&pk.ItemName)
	io.String(&pk.ItemEffectName)
	io.Int32(&pk.Slot)
}
