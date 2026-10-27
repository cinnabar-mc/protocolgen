// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// LocatorBar is sent by the server to add, remove or update waypoints on the client's locator bar.
type LocatorBar struct {
	// Waypoints is a slice of waypoints to add, remove or update.
	Waypoints []protocol.LocatorBarWaypoint
}

// ID ...
func (*LocatorBar) ID() uint32 {
	return IDLocatorBar
}

func (pk *LocatorBar) Marshal(io protocol.IO) {
	protocol.SliceLimits(io, &pk.Waypoints, 0, 40000)
}
