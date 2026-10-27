// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// MobEquipment is sent by the client to the server and the server to the client to make the other side aware
// of the new item that an entity is holding. It is used to show the item in the hand of entities such as
// zombies too.
type MobEquipment struct {
	// EntityRuntimeID is the runtime ID of the entity. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	TargetRuntimeID uint64
	// NewItem is the new item held after sending the MobEquipment packet. The entity will be shown holding that
	// item to the player it was sent to.
	Item protocol.NetworkItemStackDescriptorSerializedData
	// InventorySlot is the slot in the inventory that was held. This is the same as HotBarSlot, and only remains
	// for backwards compatibility.
	Slot uint8
	// HotBarSlot is the slot in the hot bar that was held. It is the same as InventorySlot, which is only there
	// for backwards compatibility purposes.
	SelectedSlot uint8
	// WindowID is the window ID of the window that had its equipped item changed. This is usually the window ID
	// of the normal inventory, but may also be something else, for example with the off hand.
	ContainerID uint8
}

// ID ...
func (*MobEquipment) ID() uint32 {
	return IDMobEquipment
}

func (pk *MobEquipment) Marshal(io protocol.IO) {
	io.ActorRuntimeID(&pk.TargetRuntimeID)
	pk.Item.Marshal(io)
	io.Uint8(&pk.Slot)
	io.Uint8(&pk.SelectedSlot)
	io.Uint8(&pk.ContainerID)
}
