// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// CommandRequest is sent by the client to request the execution of a server-side command. Although some
// servers support sending commands using the Text packet, this packet is guaranteed to have the correct
// result.
type CommandRequest struct {
	// Command is the raw entered command line. The client does no parsing of the command line by itself (unlike
	// it did in the early stages), but lets the server do that.
	Command string
	// Origin is the data specifying the origin of the command. In other words, the source that the command was
	// from, such as the player itself or a websocket server.
	Origin protocol.CommandOriginData
	// IsInternal specifies if the command request internal. Setting it to false seems to work and the usage of
	// this field is not known.
	IsInternal bool
	// Version is the version of the command that is being executed. This field currently has no purpose or
	// functionality.
	Version string
}

// ID ...
func (*CommandRequest) ID() uint32 {
	return IDCommandRequest
}

func (pk *CommandRequest) Marshal(io protocol.IO) {
	io.StringLimits(&pk.Command, 0, 1000)
	pk.Origin.Marshal(io)
	io.Bool(&pk.IsInternal)
	io.String(&pk.Version)
}
