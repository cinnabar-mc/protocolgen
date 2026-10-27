// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// SetLastHurtBy is sent by the server to let the client know what entity type it was last hurt by. At this
// moment, the packet is useless and should not be used. There is no behaviour that depends on if this packet
// is sent or not.
type SetLastHurtBy struct {
	// EntityType is the numerical type of the entity that the player was last hurt by.
	LastHurtBy protocol.ActorType
}

// ID ...
func (*SetLastHurtBy) ID() uint32 {
	return IDSetLastHurtBy
}

func (pk *SetLastHurtBy) Marshal(io protocol.IO) {
	pk.LastHurtBy.Marshal(io)
}
