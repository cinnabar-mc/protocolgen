// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// UpdateBlockSynced is sent by the server to synchronise the falling of a falling block entity with the
// transitioning back and forth from and to a solid block. It is used to prevent the entity from flickering,
// and is used in places such as the pushing of blocks with pistons.
type UpdateBlockSynced struct {
	// BlockPosition is the block position at which a block is updated.
	BlockPosition protocol.BlockPos
	// BlockRuntimeID is the runtime ID of the block that is placed at Position after sending the packet to the
	// client.
	BlockRuntimeID uint32
	// Flags is a combination of flags that specify the way the block is updated client-side. It is a combination
	// of the flags above, but typically sending only the BlockUpdateNetwork flag is sufficient.
	Flags uint32
	// Layer is the world layer on which the block is updated. For most blocks, this is the first layer, as that
	// layer is the default layer to place blocks on, but for blocks inside of each other, this differs.
	Layer uint32
	// UniqueActorID is the unique ID of the falling block entity that the block transitions to or that the entity
	// transitions from. Note that for both possible values for TransitionType, the EntityUniqueID should point to
	// the falling block entity involved.
	UniqueActorID uint64
	// ActorSyncMessage is the type of the transition that happened. It is either BlockToEntityTransition, when a
	// block placed becomes a falling entity, or EntityToBlockTransition, when a falling entity hits the ground
	// and becomes a solid block again.
	ActorSyncMessage uint64
}

// ID ...
func (*UpdateBlockSynced) ID() uint32 {
	return IDUpdateBlockSynced
}

func (pk *UpdateBlockSynced) Marshal(io protocol.IO) {
	pk.BlockPosition.Marshal(io)
	io.Varuint32(&pk.BlockRuntimeID)
	io.Varuint32(&pk.Flags)
	io.Varuint32(&pk.Layer)
	io.Varuint64(&pk.UniqueActorID)
	io.Varuint64(&pk.ActorSyncMessage)
}
