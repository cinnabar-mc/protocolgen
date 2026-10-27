// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// MovementEffect is sent by the server to the client to update specific movement effects to allow the client
// to predict its movement. For example, fireworks used during gliding will send this packet to tell the
// client the exact duration of the boost.
type MovementEffect struct {
	// EntityRuntimeID is the runtime ID of the entity. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	TargetRuntimeID uint64
	// Type is the type of movement effect being updated. It is one of the constants found above.
	EffectID protocol.MovementEffectType
	// Duration is the duration of the effect, measured in ticks.
	EffectDuration int32
	// Tick is the server tick at which the packet was sent. It is used in relation to
	// CorrectPlayerMovePrediction.
	Tick uint64
}

// ID ...
func (*MovementEffect) ID() uint32 {
	return IDMovementEffect
}

func (pk *MovementEffect) Marshal(io protocol.IO) {
	io.ActorRuntimeID(&pk.TargetRuntimeID)
	pk.EffectID.Marshal(io)
	io.Varint32(&pk.EffectDuration)
	io.PlayerInputTick(&pk.Tick)
}
