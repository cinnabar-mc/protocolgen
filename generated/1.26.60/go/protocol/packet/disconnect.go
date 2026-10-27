// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// Disconnect may be sent by the server to disconnect the client using an optional message to send as the
// disconnect screen.
type Disconnect struct {
	// Reason is the reason for the disconnection. This affects the error code displayed on the Ore UI
	// disconnection screen and is one of the constants above.
	Reason   protocol.ConnectionDisconnectFailReason
	Messages protocol.DisconnectMessages
}

// ID ...
func (*Disconnect) ID() uint32 {
	return IDDisconnect
}

func (pk *Disconnect) Marshal(io protocol.IO) {
	pk.Reason.Marshal(io)
	protocol.MarshalDisconnectMessages(io, &pk.Messages)
}
