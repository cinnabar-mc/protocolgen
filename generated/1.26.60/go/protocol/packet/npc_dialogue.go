// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// NPCDialogue is a packet that allows the client to display dialog boxes for interacting with NPCs.
type NpcDialogue struct {
	// EntityUniqueID is the unique ID of the NPC being requested.
	NpcIDRawID uint64
	// ActionType is the type of action for the packet.
	NpcDialogueActionType protocol.NpcDialogueActionType
	// Dialogue is the text that the client should see.
	Dialogue string
	// SceneName is the identifier of the scene. If this is left empty, the client will use the last scene sent to
	// it. https://docs.microsoft.com/en-us/minecraft/creator/documents/npcdialogue.
	SceneName string
	// NPCName is the name of the NPC to be displayed to the client.
	NpcName string
	// ActionJSON is the JSON string of the buttons/actions the server can perform.
	ActionJSON string
}

// ID ...
func (*NpcDialogue) ID() uint32 {
	return IDNpcDialogue
}

func (pk *NpcDialogue) Marshal(io protocol.IO) {
	io.Uint64(&pk.NpcIDRawID)
	pk.NpcDialogueActionType.Marshal(io)
	io.String(&pk.Dialogue)
	io.String(&pk.SceneName)
	io.String(&pk.NpcName)
	io.String(&pk.ActionJSON)
}
