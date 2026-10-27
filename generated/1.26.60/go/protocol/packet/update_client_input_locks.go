// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// UpdateClientInputLocks is sent by the server to the client to lock specific player inputs such as camera
// rotation, movement, jumping, sneaking, mounting or individual directional movement.
type UpdateClientInputLocks struct {
	// InputLockComponentData is a set of flags that specify which client inputs are disabled, such as whether the
	// player can move, rotate the camera, jump, sneak or mount/dismount entities. It is a combination of the
	// ClientInputLock constants above.
	InputLockComponentData uint32
}

// ID ...
func (*UpdateClientInputLocks) ID() uint32 {
	return IDUpdateClientInputLocks
}

func (pk *UpdateClientInputLocks) Marshal(io protocol.IO) {
	io.Varuint32(&pk.InputLockComponentData)
}
