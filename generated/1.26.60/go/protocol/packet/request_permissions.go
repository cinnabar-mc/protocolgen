// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// RequestPermissions is a packet sent from the client to the server to request permissions that the client
// does not currently have. It can only be sent by operators and host in vanilla Minecraft.
type RequestPermissions struct {
	// TargetPlayerIDSRawID is the unique ID of the player. The unique ID is unique for the entire world and is
	// often used in packets. Most servers send an EntityUniqueID equal to the EntityRuntimeID.
	TargetPlayerIDSRawID int64
	// PlayerPermissionLevel is the current permission level of the player. This is one of the constants that may
	// be found in the AdventureSettings packet.
	PlayerPermissionLevel int32
	// CustomPermissionFlags contains the requested permission flags.
	CustomPermissionFlags uint16
}

// ID ...
func (*RequestPermissions) ID() uint32 {
	return IDRequestPermissions
}

func (pk *RequestPermissions) Marshal(io protocol.IO) {
	io.Int64(&pk.TargetPlayerIDSRawID)
	io.Varint32(&pk.PlayerPermissionLevel)
	io.Uint16(&pk.CustomPermissionFlags)
}
