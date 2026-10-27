// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// RequestNetworkSettings is sent by the client to request network settings, such as compression, from the
// server.
type RequestNetworkSettings struct {
	// ClientNetworkVersion is the protocol version of the player. The player is disconnected if the protocol is
	// incompatible with the protocol of the server.
	ClientNetworkVersion int32
}

// ID ...
func (*RequestNetworkSettings) ID() uint32 {
	return IDRequestNetworkSettings
}

func (pk *RequestNetworkSettings) Marshal(io protocol.IO) {
	io.BEInt32(&pk.ClientNetworkVersion)
	protocol.Minimum(io, &pk.ClientNetworkVersion, 2225)
	protocol.Maximum(io, &pk.ClientNetworkVersion, 2225)
}
