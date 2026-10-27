// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// InventorySlot is sent by the server to update a single slot in one of the inventory windows that the client
// currently has opened. Usually this is the main inventory, but it may also be the off hand or, for example,
// a chest inventory.
type InventorySlot struct {
	// ContainerID is the ID of the window that the packet modifies. It must point to one of the windows that the
	// client currently has opened.
	ContainerID uint8
	// Slot is the index of the slot that the packet modifies. The new item will be set to the slot at this index.
	Slot uint32
	// FullContainerName is the protocol.FullContainerName that describes the container that the content is for.
	FullContainerName protocol.Optional[protocol.FullContainerName]
	// StorageItem is the item that is acting as the storage container for the inventory. If the inventory is not
	// a dynamic container then this field should be left empty. When set, only the item type is used by the
	// client and none of the other stack info.
	StorageItem protocol.Optional[protocol.NetworkItemStackDescriptorSerializedData]
	// Item is the item to be put in the slot at Slot. It will overwrite any item that may currently be present in
	// that slot.
	Item protocol.NetworkItemStackDescriptorSerializedData
}

// ID ...
func (*InventorySlot) ID() uint32 {
	return IDInventorySlot
}

func (pk *InventorySlot) Marshal(io protocol.IO) {
	io.Uint8(&pk.ContainerID)
	io.Varuint32(&pk.Slot)
	protocol.OptionalMarshaler(io, &pk.FullContainerName)
	protocol.OptionalMarshaler(io, &pk.StorageItem)
	pk.Item.Marshal(io)
}
