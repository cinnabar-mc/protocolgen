// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// UpdateAbilities is a packet sent from the server to the client to update the abilities of the player. It,
// along with the UpdateAdventureSettings packet, are replacements of the AdventureSettings packet since
// v1.19.10.
type UpdateAbilities struct {
	// Data represents various data about the abilities of a player, such as ability layers or permissions.
	Data protocol.SerializedAbilitiesData
}

// ID ...
func (*UpdateAbilities) ID() uint32 {
	return IDUpdateAbilities
}

func (pk *UpdateAbilities) Marshal(io protocol.IO) {
	pk.Data.Marshal(io)
}
