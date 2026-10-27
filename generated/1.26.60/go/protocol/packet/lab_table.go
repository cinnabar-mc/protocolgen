// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// LabTable is sent by the client to let the server know it started a chemical reaction in Education Edition,
// and is sent by the server to other clients to show the effects. The packet is only functional if Education
// features are enabled.
type LabTable struct {
	// Type is the type of the action that was executed. It is one of the constants above. Typically, only
	// LabTableActionCombine is sent by the client, whereas LabTableActionReact is sent by the server.
	Type protocol.LabTableType
	// Position is the position at which the lab table used was located.
	Position protocol.BlockPos
	// Reaction is the type of the reaction that took place as a result of the items put into the lab table. The
	// reaction type can be either that of an item or a particle, depending on whatever the result was of the
	// reaction.
	Reaction protocol.LabTableReactionType
}

// ID ...
func (*LabTable) ID() uint32 {
	return IDLabTable
}

func (pk *LabTable) Marshal(io protocol.IO) {
	pk.Type.Marshal(io)
	pk.Position.Marshal(io)
	pk.Reaction.Marshal(io)
}
