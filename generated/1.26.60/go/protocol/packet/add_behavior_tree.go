// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// AddBehaviourTree is sent by the server to the client. The packet is currently unused by both client and
// server.
type AddBehaviorTree struct {
	// BehaviourTree is an unused string.
	BehaviorTreeStructureJSON string
}

// ID ...
func (*AddBehaviorTree) ID() uint32 {
	return IDAddBehaviorTree
}

func (pk *AddBehaviorTree) Marshal(io protocol.IO) {
	io.String(&pk.BehaviorTreeStructureJSON)
}
