// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// ContainerClose is sent by the server to close a container the player currently has opened, which was opened
// using the ContainerOpen packet, or by the client to tell the server it closed a particular container, such
// as the crafting grid.
type ContainerClose struct {
	// WindowID is the ID representing the window of the container that should be closed. It must be equal to the
	// one sent in the ContainerOpen packet to close the designated window.
	ContainerID uint8
	// ContainerType is the type of container that the server is trying to close. This is used to validate on the
	// client side whether or not the server's close request is valid.
	ContainerType uint8
	// ServerSide determines whether or not the container was force-closed by the server. If this value is not set
	// correctly, the client may ignore the packet and respond with a PacketViolationWarning.
	ServerInitiatedClose bool
}

// ID ...
func (*ContainerClose) ID() uint32 {
	return IDContainerClose
}

func (pk *ContainerClose) Marshal(io protocol.IO) {
	io.Uint8(&pk.ContainerID)
	io.Uint8(&pk.ContainerType)
	io.Bool(&pk.ServerInitiatedClose)
}
