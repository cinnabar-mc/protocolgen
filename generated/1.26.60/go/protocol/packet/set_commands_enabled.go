// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// SetCommandsEnabled is sent by the server to enable or disable the ability to execute commands for the
// client. If disabled, the client itself will stop the execution of commands.
type SetCommandsEnabled struct {
	// CommandsEnabled defines if the commands should be enabled, or if false, disabled.
	CommandsEnabled bool
}

// ID ...
func (*SetCommandsEnabled) ID() uint32 {
	return IDSetCommandsEnabled
}

func (pk *SetCommandsEnabled) Marshal(io protocol.IO) {
	io.Bool(&pk.CommandsEnabled)
}
