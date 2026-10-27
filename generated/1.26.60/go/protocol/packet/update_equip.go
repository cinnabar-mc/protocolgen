// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// UpdateEquip is sent by the server to the client upon opening a horse inventory. It is used to set the
// content of the inventory and specify additional properties, such as the items that are allowed to be put in
// slots of the inventory.
type UpdateEquip struct {
	ContainerID uint8
	// WindowType is the type of the window that was opened. Generally, this is the type of a horse inventory, as
	// the packet is specifically made for that.
	Type uint8
	// Size is the size of the horse inventory that should be opened. A bigger size does, in fact, change the
	// amount of slots displayed.
	Size int32
	// EntityUniqueID is the unique ID of the entity whose equipment was 'updated' to the player. It is typically
	// the horse entity that had its inventory opened.
	EntityUniqueID int64
	Data           []byte
}

// ID ...
func (*UpdateEquip) ID() uint32 {
	return IDUpdateEquip
}

func (pk *UpdateEquip) Marshal(io protocol.IO) {
	io.Uint8(&pk.ContainerID)
	io.Uint8(&pk.Type)
	io.Varint32(&pk.Size)
	io.ActorUniqueID(&pk.EntityUniqueID)
	io.NBT(&pk.Data, protocol.NBTNetwork)
}
