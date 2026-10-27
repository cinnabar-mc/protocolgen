// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// AgentAction is an Education Edition packet sent from the server to the client to return a response to a
// previously requested action.
type AgentActionEvent struct {
	// Identifier is a JSON identifier referenced in the initial action.
	RequestID string
	// Action represents the action type that was requested. It is one of the constants defined above.
	Action protocol.AgentActionType
	// Response is a JSON string containing the response to the action.
	Response string
}

// ID ...
func (*AgentActionEvent) ID() uint32 {
	return IDAgentActionEvent
}

func (pk *AgentActionEvent) Marshal(io protocol.IO) {
	io.String(&pk.RequestID)
	pk.Action.Marshal(io)
	io.String(&pk.Response)
}
