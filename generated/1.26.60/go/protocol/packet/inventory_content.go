// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// InventoryContent is sent by the server to update the full content of a particular inventory. It is usually
// sent for the main inventory of the player, but also works for other inventories that are currently opened
// by the player.
type InventoryContent struct {
	// ContainerID is the ID that identifies one of the windows that the client currently has opened, or one of
	// the consistent windows such as the main inventory.
	ContainerID uint32
	// Slots is the new content of the inventory. The length of this slice must be equal to the full size of the
	// inventory window updated.
	Slots []protocol.NetworkItemStackDescriptorSerializedData
	// FullContainerName is the protocol.FullContainerName that describes the container that the content is for.
	FullContainerName protocol.FullContainerName
	// StorageItem is the item that is acting as the storage container for the inventory. If the inventory is not
	// a dynamic container then this field should be left empty. When set, only the item type is used by the
	// client and none of the other stack info.
	StorageItem protocol.NetworkItemStackDescriptorSerializedData
}

// ID ...
func (*InventoryContent) ID() uint32 {
	return IDInventoryContent
}

func (pk *InventoryContent) Marshal(io protocol.IO) {
	io.Varuint32(&pk.ContainerID)
	protocol.Slice(io, &pk.Slots)
	pk.FullContainerName.Marshal(io)
	pk.StorageItem.Marshal(io)
}
