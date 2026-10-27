// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// ContainerOpen is sent by the server to open a container client-side. This container must be physically
// present in the world, for the packet to have any effect. Unlike Java Edition, Bedrock Edition requires that
// chests for example must be present and in range to open its inventory.
type ContainerOpen struct {
	// WindowID is the ID representing the window that is being opened. It may be used later to close the
	// container using a ContainerClose packet.
	ContainerID uint8
	// ContainerType is the type ID of the container that is being opened when opening the container at the
	// position of the packet. It depends on the block/entity, and could, for example, be the window type of a
	// chest or a hopper, but also a horse inventory.
	ContainerType uint8
	// ContainerPosition is the position of the container opened. The position must point to a block entity that
	// actually has a container. If that is not the case, the window will not be opened and the packet will be
	// ignored, if a valid ContainerEntityUniqueID has not also been provided.
	Position protocol.BlockPos
	// ContainerEntityUniqueID is the unique ID of the entity container that was opened. It is only used if the
	// ContainerType is one that points to an entity, for example a horse.
	TargetActorID int64
}

// ID ...
func (*ContainerOpen) ID() uint32 {
	return IDContainerOpen
}

func (pk *ContainerOpen) Marshal(io protocol.IO) {
	io.Uint8(&pk.ContainerID)
	io.Uint8(&pk.ContainerType)
	pk.Position.Marshal(io)
	io.ActorUniqueID(&pk.TargetActorID)
}
