// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// SetActorData is sent by the server to update the entity metadata of an entity. It includes flags such as if
// the entity is on fire, but also properties such as the air it has left until it starts drowning.
type SetActorData struct {
	// EntityRuntimeID is the runtime ID of the entity. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	TargetRuntimeID uint64
	// EntityMetadata is a map of entity metadata, which includes flags and data properties that alter in
	// particular the way the entity looks. Flags include ones such as 'on fire' and 'sprinting'. The metadata
	// values are indexed by their property key.
	ActorData protocol.SynchedActorDataCopyableDataList
	// EntityProperties is a list of properties that the entity inhibits. These properties define and alter
	// specific attributes of the entity.
	SynchedProperties protocol.PropertySyncData
	// Tick is the server tick at which the packet was sent. It is used in relation to
	// CorrectPlayerMovePrediction.
	Tick uint64
}

// ID ...
func (*SetActorData) ID() uint32 {
	return IDSetActorData
}

func (pk *SetActorData) Marshal(io protocol.IO) {
	io.ActorRuntimeID(&pk.TargetRuntimeID)
	pk.ActorData.Marshal(io)
	pk.SynchedProperties.Marshal(io)
	io.PlayerInputTick(&pk.Tick)
}
