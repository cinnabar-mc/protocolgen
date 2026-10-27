// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// NPCRequest is sent by the client when it interacts with an NPC. The packet is specifically made for
// Education Edition, where NPCs are available to use.
type NpcRequest struct {
	// EntityRuntimeID is the runtime ID of the NPC entity that the player interacted with. It is the same as sent
	// by the server when spawning the entity.
	NPCRuntimeID uint64
	// RequestType is the type of the request, which depends on the permission that the player has. It will be
	// either a type that indicates that the NPC should show its dialog, or that it should open the editing
	// window.
	RequestType protocol.RequestType
	// CommandString is the command string set in the NPC. It may consist of multiple commands, depending on what
	// the player set in it.
	Actions string
	// ActionType is the type of the action to execute.
	ActionIndex uint8
	// SceneName is the name of the scene. This can be left empty to specify the last scene that the player was
	// sent.
	SceneName string
}

// ID ...
func (*NpcRequest) ID() uint32 {
	return IDNpcRequest
}

func (pk *NpcRequest) Marshal(io protocol.IO) {
	io.ActorRuntimeID(&pk.NPCRuntimeID)
	pk.RequestType.Marshal(io)
	io.String(&pk.Actions)
	io.Uint8(&pk.ActionIndex)
	io.String(&pk.SceneName)
}
