// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// MobArmourEquipment is sent by the server to the client to update the armour an entity is wearing. It is
// sent for both players and other entities, such as zombies.
type MobArmorEquipment struct {
	// EntityRuntimeID is the runtime ID of the entity. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	TargetRuntimeID uint64
	// Helmet is the equipped helmet of the entity. Items that are not wearable on the head will not be rendered
	// by the client. Unlike in Java Edition, blocks cannot be worn.
	Head protocol.NetworkItemStackDescriptorSerializedData
	// Chestplate is the chestplate of the entity. Items that are not wearable as chestplate will not be rendered.
	Torso protocol.NetworkItemStackDescriptorSerializedData
	// Leggings is the item worn as leggings by the entity. Items not wearable as leggings will not be rendered
	// client-side.
	Legs protocol.NetworkItemStackDescriptorSerializedData
	// Boots is the item worn as boots by the entity. Items not wearable as boots will not be rendered.
	Feet protocol.NetworkItemStackDescriptorSerializedData
	// Body is the item worn on the body of the entity. Items not wearable on the body will not be rendered.
	Body protocol.NetworkItemStackDescriptorSerializedData
}

// ID ...
func (*MobArmorEquipment) ID() uint32 {
	return IDMobArmorEquipment
}

func (pk *MobArmorEquipment) Marshal(io protocol.IO) {
	io.ActorRuntimeID(&pk.TargetRuntimeID)
	pk.Head.Marshal(io)
	pk.Torso.Marshal(io)
	pk.Legs.Marshal(io)
	pk.Feet.Marshal(io)
	pk.Body.Marshal(io)
}
