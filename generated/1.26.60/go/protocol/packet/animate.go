// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// Animate is sent by the server to send a player animation from one player to all viewers of that player. It
// is used for a couple of actions, such as arm swimming and critical hits.
type Animate struct {
	// ActionType is the ID of the animation action to execute. It is one of the action type constants that may be
	// found above.
	Action protocol.AnimateAction
	// EntityRuntimeID is the runtime ID of the player that the animation should be played upon. The runtime ID is
	// unique for each world session, and entities are generally identified in packets using this runtime ID.
	TargetActorRuntimeID uint64
	// Data ...
	Data float32
	// SwingSource is the source for swing actions. It is one of the action type constants that may be found
	// above.
	SwingSource protocol.Optional[string]
	Hand        protocol.HandSlot
}

// ID ...
func (*Animate) ID() uint32 {
	return IDAnimate
}

func (pk *Animate) Marshal(io protocol.IO) {
	pk.Action.Marshal(io)
	io.ActorRuntimeID(&pk.TargetActorRuntimeID)
	io.Float32(&pk.Data)
	protocol.OptionalFunc(io, &pk.SwingSource, io.String)
	pk.Hand.Marshal(io)
}
