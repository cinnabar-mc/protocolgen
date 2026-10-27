// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"

	"github.com/go-gl/mathgl/mgl32"
)

// MotionPredictionHints is sent by the server to the client. There is a predictive movement component for
// entities. This packet fills the "history" of that component and entity movement is computed based on the
// points. Vanilla sends this packet instead of the SetActorMotion packet when 'spatial optimisations' are
// enabled.
type MotionPredictionHints struct {
	// EntityRuntimeID is the runtime ID of the entity whose velocity is sent to the client.
	MRuntimeID uint64
	// Velocity is the server-calculated velocity of the entity at the point of sending the packet.
	MMotion mgl32.Vec3
	// OnGround specifies if the server currently thinks the entity is on the ground.
	MOnGround bool
}

// ID ...
func (*MotionPredictionHints) ID() uint32 {
	return IDMotionPredictionHints
}

func (pk *MotionPredictionHints) Marshal(io protocol.IO) {
	io.ActorRuntimeID(&pk.MRuntimeID)
	io.Vec3(&pk.MMotion)
	io.Bool(&pk.MOnGround)
}
