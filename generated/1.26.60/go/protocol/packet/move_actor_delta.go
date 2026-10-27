// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// MoveActorDelta is sent by the server to move an entity. The packet is specifically optimised to save as
// much space as possible, by only writing non-zero fields. As of 1.16.100, this packet no longer actually
// contains any deltas.
type MoveActorDelta struct {
	// EntityRuntimeID is the runtime ID of the entity that is being moved. The packet works provided a non-player
	// entity with this runtime ID is present.
	ActorRuntimeID uint64
	// Position is the new position that the entity was moved to.
	NewPositionX protocol.Optional[float32]
	NewPositionY protocol.Optional[float32]
	NewPositionZ protocol.Optional[float32]
	// Rotation is the new absolute rotation. Unlike the position, it is not actually a delta. If any of the
	// values of this rotation are not sent, these values are 0 and no flag for them is present.
	RotationX     protocol.Optional[int8]
	RotationY     protocol.Optional[int8]
	RotationYHead protocol.Optional[int8]
	// OnGround specifies whether the entity is on the ground after applying the update.
	IsOnGround bool
	// ForceMove specifies whether the client should snap the entity to its new position without interpolation.
	ForceMove bool
	// ForceMoveLocalEntity specifies whether the client should also snap an entity it owns locally.
	ForceMoveLocalEntity bool
	// ForceCompletion specifies whether the client should complete any in-progress local movement before applying
	// the update.
	ForceCompletion bool
	// Ticks is the number of ticks over which the movement should be interpolated by the client.
	Ticks uint64
}

// ID ...
func (*MoveActorDelta) ID() uint32 {
	return IDMoveActorDelta
}

func (pk *MoveActorDelta) Marshal(io protocol.IO) {
	io.ActorRuntimeID(&pk.ActorRuntimeID)
	protocol.OptionalFunc(io, &pk.NewPositionX, io.Float32)
	protocol.OptionalFunc(io, &pk.NewPositionY, io.Float32)
	protocol.OptionalFunc(io, &pk.NewPositionZ, io.Float32)
	protocol.OptionalFunc(io, &pk.RotationX, io.Int8)
	protocol.OptionalFunc(io, &pk.RotationY, io.Int8)
	protocol.OptionalFunc(io, &pk.RotationYHead, io.Int8)
	io.Bool(&pk.IsOnGround)
	io.Bool(&pk.ForceMove)
	io.Bool(&pk.ForceMoveLocalEntity)
	io.Bool(&pk.ForceCompletion)
	io.Varuint64(&pk.Ticks)
}
