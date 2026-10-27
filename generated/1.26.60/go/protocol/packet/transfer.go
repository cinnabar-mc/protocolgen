// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// Transfer is sent by the server to transfer a player from the current server to another. Doing so will fully
// disconnect the client, bring it back to the main menu and make it connect to the next server.
type Transfer struct {
	// ServerAddress is the address of the new server, which might be either a hostname or an actual IP address.
	ServerAddress string
	// ServerPort is the UDP port of the new server.
	ServerPort uint16
	// ReloadWorld currently has an unknown usage.
	ReloadWorld bool
	// GatheringsConfiguration optionally identifies the gathering being joined on the target server.
	GatheringsConfiguration protocol.Optional[protocol.ServerConfigurationGatheringsConfigurationJoinInfo]
}

// ID ...
func (*Transfer) ID() uint32 {
	return IDTransfer
}

func (pk *Transfer) Marshal(io protocol.IO) {
	io.String(&pk.ServerAddress)
	io.Uint16(&pk.ServerPort)
	io.Bool(&pk.ReloadWorld)
	protocol.OptionalMarshaler(io, &pk.GatheringsConfiguration)
}
