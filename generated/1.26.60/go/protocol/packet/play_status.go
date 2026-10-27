// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// PlayStatus is sent by the server to update a player on the play status. This includes failed statuses due
// to a mismatched version, but also success statuses.
type PlayStatus struct {
	// Status is the status of the packet. It is one of the constants found above.
	Status protocol.PlayStatusType
}

// ID ...
func (*PlayStatus) ID() uint32 {
	return IDPlayStatus
}

func (pk *PlayStatus) Marshal(io protocol.IO) {
	pk.Status.Marshal(io)
}
