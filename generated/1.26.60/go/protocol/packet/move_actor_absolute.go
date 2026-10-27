// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"

	"github.com/go-gl/mathgl/mgl32"
)

// MoveActorAbsolute is sent by the server to move an entity to an absolute position. It is typically used for
// movements where high accuracy isn't needed, such as for long range teleporting.
type MoveActorAbsolute struct {
	// EntityRuntimeID is the runtime ID of the entity. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	ActorRuntimeID uint64
	Header         uint8
	// Position is the position to spawn the entity on. If the entity is on a distance that the player cannot see
	// it, the entity will still show up if the player moves closer.
	Position      mgl32.Vec3
	RotationX     uint8
	RotationY     uint8
	RotationYHead uint8
}

// ID ...
func (*MoveActorAbsolute) ID() uint32 {
	return IDMoveActorAbsolute
}

func (pk *MoveActorAbsolute) Marshal(io protocol.IO) {
	io.ActorRuntimeID(&pk.ActorRuntimeID)
	io.Uint8(&pk.Header)
	io.Vec3(&pk.Position)
	io.Uint8(&pk.RotationX)
	io.Uint8(&pk.RotationY)
	io.Uint8(&pk.RotationYHead)
}
