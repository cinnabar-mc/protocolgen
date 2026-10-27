// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// RemoveObjective is sent by the server to remove a scoreboard objective. It is used to stop showing a
// scoreboard to a player.
type RemoveObjective struct {
	// ObjectiveName is the name of the objective that the scoreboard currently active has. This name must be
	// identical to the one sent in the SetDisplayObjective packet.
	ObjectiveName string
}

// ID ...
func (*RemoveObjective) ID() uint32 {
	return IDRemoveObjective
}

func (pk *RemoveObjective) Marshal(io protocol.IO) {
	io.String(&pk.ObjectiveName)
}
