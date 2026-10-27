// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// SetLocalPlayerAsInitialised is sent by the client in response to a PlayStatus packet with the status set to
// 3. The packet marks the moment at which the client is fully initialised and can receive any packet without
// discarding it.
type SetLocalPlayerAsInitialized struct {
	// EntityRuntimeID is the entity runtime ID the player was assigned earlier in the login sequence in the
	// StartGame packet.
	PlayerID uint64
}

// ID ...
func (*SetLocalPlayerAsInitialized) ID() uint32 {
	return IDSetLocalPlayerAsInitialized
}

func (pk *SetLocalPlayerAsInitialized) Marshal(io protocol.IO) {
	io.ActorRuntimeID(&pk.PlayerID)
}
