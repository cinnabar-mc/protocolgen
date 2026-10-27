// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// PlayerLocation is sent by the server to the client to either update a player's position on the locator bar,
// or remove them completely. The client will determine how to render the player on the locator bar based on
// their own distance to Position.
type PlayerLocation struct {
	// EntityUniqueID is the unique ID of the entity. The unique ID is a value that remains consistent across
	// different sessions of the same world.
	TargetActorID int64
	Location      protocol.PlayerLocationData
}

// ID ...
func (*PlayerLocation) ID() uint32 {
	return IDPlayerLocation
}

func (pk *PlayerLocation) Marshal(io protocol.IO) {
	io.ActorUniqueID(&pk.TargetActorID)
	protocol.MarshalPlayerLocationData(io, &pk.Location)
}
