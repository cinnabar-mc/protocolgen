// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// SetDifficulty is sent by the server to update the client-side difficulty of the client. The actual effect
// of this packet on the client isn't very significant, as the difficulty is handled server-side.
type SetDifficulty struct {
	// Difficulty is the new difficulty that the world has.
	Difficulty uint32
}

// ID ...
func (*SetDifficulty) ID() uint32 {
	return IDSetDifficulty
}

func (pk *SetDifficulty) Marshal(io protocol.IO) {
	io.Varuint32(&pk.Difficulty)
}
