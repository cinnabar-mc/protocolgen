// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// SetActorLink is sent by the server to initiate an entity link client-side, meaning one entity will start
// riding another.
type SetActorLink struct {
	// Link is the link to be set client-side. It links two entities together, so that one entity rides another.
	// Note that players that see those entities later will not see the link, unless it is also sent in the
	// AddActor and AddPlayer packets.
	Link protocol.EntityLink
}

// ID ...
func (*SetActorLink) ID() uint32 {
	return IDSetActorLink
}

func (pk *SetActorLink) Marshal(io protocol.IO) {
	pk.Link.Marshal(io)
}
