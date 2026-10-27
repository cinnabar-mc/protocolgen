// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// AgentAnimation is an Education Edition packet sent from the server to the client to make an agent perform
// an animation.
type AgentAnimation struct {
	// Animation is the ID of the animation that the agent should perform. As of its implementation, there are no
	// IDs that can be used in the regular client.
	AgentAnimation protocol.AgentAnimationType
	// EntityRuntimeID is the runtime ID of the entity. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	RuntimeID uint64
}

// ID ...
func (*AgentAnimation) ID() uint32 {
	return IDAgentAnimation
}

func (pk *AgentAnimation) Marshal(io protocol.IO) {
	pk.AgentAnimation.Marshal(io)
	io.ActorRuntimeID(&pk.RuntimeID)
}
