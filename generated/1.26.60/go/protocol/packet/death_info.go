// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// DeathInfo is a packet sent from the server to the client expected to be sent when a player dies. It
// contains messages related to the player's death, which are shown on the death screen as of v1.19.10.
type DeathInfo struct {
	// DeathCauseAttackName is the cause of the player's death, such as "suffocation" or "suicide".
	DeathCauseAttackName string
	// DeathCauseMessageList is a list of death messages to be shown on the death screen.
	DeathCauseMessageList []string
}

// ID ...
func (*DeathInfo) ID() uint32 {
	return IDDeathInfo
}

func (pk *DeathInfo) Marshal(io protocol.IO) {
	io.String(&pk.DeathCauseAttackName)
	protocol.FuncSlice(io, &pk.DeathCauseMessageList, io.Varuint32, io.String)
}
