// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// SetPlayerInventoryOptions is a bidirectional packet that can be used to update the inventory options of a
// player.
type SetPlayerInventoryOptions struct {
	// LeftInventoryTab is the tab that is selected on the left side of the inventory. This is usually for the
	// creative inventory. It is one of the InventoryLeftTab constants above.
	LeftInventoryTab protocol.InventoryLeftTabIndex
	// RightInventoryTab is the tab that is selected on the right side of the inventory. This is usually for the
	// player's own inventory. It is one of the InventoryRightTab constants above.
	RightInventoryTab protocol.InventoryRightTabIndex
	// Filtering is whether the player has enabled the filtering between recipes they have unlocked or not.
	Filtering bool
	// InventoryLayout is the layout of the inventory. It is one of the InventoryLayout constants above.
	LayoutInv protocol.InventoryLayout
	// CraftingLayout is the layout of the crafting inventory. It is one of the InventoryLayout constants above.
	LayoutCraft protocol.InventoryLayout
}

// ID ...
func (*SetPlayerInventoryOptions) ID() uint32 {
	return IDSetPlayerInventoryOptions
}

func (pk *SetPlayerInventoryOptions) Marshal(io protocol.IO) {
	pk.LeftInventoryTab.Marshal(io)
	pk.RightInventoryTab.Marshal(io)
	io.Bool(&pk.Filtering)
	pk.LayoutInv.Marshal(io)
	pk.LayoutCraft.Marshal(io)
}
