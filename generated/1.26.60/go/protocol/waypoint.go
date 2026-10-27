// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

import (
	"github.com/google/uuid"
)

// LocatorBarWaypoint represents a waypoint entry in the locator bar packet.
type LocatorBarWaypoint struct {
	// GroupHandle is the UUID handle for the waypoint group.
	GroupHandle WaypointGroupWaypointHandle
	// Waypoint contains the waypoint data.
	ServerWaypointPayload ServerWaypoint
	// Action determines the action for this waypoint. It is one of the WaypointAction constants.
	ActionFlag ServerWaypointGroupAction
}

// Marshal reads or writes LocatorBarWaypoint using its canonical wire layout.
func (x *LocatorBarWaypoint) Marshal(io IO) {
	x.GroupHandle.Marshal(io)
	x.ServerWaypointPayload.Marshal(io)
	x.ActionFlag.Marshal(io)
}

type WaypointGroupWaypointHandle struct {
	UUID uuid.UUID
}

// Marshal reads or writes WaypointGroupWaypointHandle using its canonical wire layout.
func (x *WaypointGroupWaypointHandle) Marshal(io IO) {
	io.UUID(&x.UUID)
}
