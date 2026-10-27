// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// ClientToServerHandshake is sent by the client in response to a ServerToClientHandshake packet sent by the
// server. It is the first encrypted packet in the login handshake and serves as a confirmation that
// encryption is correctly initialised client side.
type ClientToServerHandshake struct {
}

// ID ...
func (*ClientToServerHandshake) ID() uint32 {
	return IDClientToServerHandshake
}

func (pk *ClientToServerHandshake) Marshal(io protocol.IO) {
}
