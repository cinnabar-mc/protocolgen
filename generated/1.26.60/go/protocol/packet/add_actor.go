// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"

	"github.com/go-gl/mathgl/mgl32"
)

// AddActor is sent by the server to the client to spawn an entity to the player. It is used for every entity
// except other players, for which the AddPlayer packet is used.
type AddActor struct {
	// EntityUniqueID is the unique ID of the entity. The unique ID is a value that remains consistent across
	// different sessions of the same world, but most servers simply fill the runtime ID of the entity out for
	// this field.
	TargetActorID int64
	// EntityRuntimeID is the runtime ID of the entity. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	TargetRuntimeID uint64
	// EntityType is the string entity type of the entity, for example 'minecraft:skeleton'. A list of these
	// entities may be found online.
	ActorType string
	// Position is the position to spawn the entity on. If the entity is on a distance that the player cannot see
	// it, the entity will still show up if the player moves closer.
	Position mgl32.Vec3
	// Velocity is the initial velocity the entity spawns with. This velocity will initiate client side movement
	// of the entity.
	Velocity      mgl32.Vec3
	Rotation      mgl32.Vec2
	YHeadRotation float32
	YBodyRotation float32
	// Attributes is a slice of attributes that the entity has. It includes attributes such as its health,
	// movement speed, etc.
	AttributesList    []protocol.SyncedAttribute
	ActorData         protocol.SynchedActorDataCopyableDataList
	SynchedProperties protocol.PropertySyncData
	// EntityLinks is a list of entity links that are currently active on the entity. These links alter the way
	// the entity shows up when first spawned in terms of it shown as riding an entity. Setting these links is
	// important for new viewers to see the entity is riding another entity.
	ActorLinks         []protocol.EntityLink
	PassengerBlockData protocol.Optional[protocol.PassengerOfBlockArguments]
}

// ID ...
func (*AddActor) ID() uint32 {
	return IDAddActor
}

func (pk *AddActor) Marshal(io protocol.IO) {
	io.ActorUniqueID(&pk.TargetActorID)
	io.ActorRuntimeID(&pk.TargetRuntimeID)
	io.String(&pk.ActorType)
	io.Vec3(&pk.Position)
	io.Vec3(&pk.Velocity)
	io.Vec2(&pk.Rotation)
	io.Float32(&pk.YHeadRotation)
	io.Float32(&pk.YBodyRotation)
	protocol.Slice(io, &pk.AttributesList)
	pk.ActorData.Marshal(io)
	pk.SynchedProperties.Marshal(io)
	protocol.Slice(io, &pk.ActorLinks)
	protocol.OptionalMarshaler(io, &pk.PassengerBlockData)
}
