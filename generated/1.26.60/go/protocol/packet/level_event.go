// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"

	"github.com/go-gl/mathgl/mgl32"
)

// LevelEvent is sent by the server to make a certain event in the level occur. It ranges from particles, to
// sounds, and other events such as starting rain and block breaking.
type LevelEvent struct {
	// EventID is the ID of the event that is being 'called'. It is one of the events found in the constants
	// above.
	EventID int32
	// Position is the position of the level event. Practically every event requires this Vec3 set for it, as
	// particles, sounds and block editing relies on it.
	Position mgl32.Vec3
	// Data is an integer holding additional data of the event. The type of data held depends on the EventType.
	Data int32
}

// ID ...
func (*LevelEvent) ID() uint32 {
	return IDLevelEvent
}

func (pk *LevelEvent) Marshal(io protocol.IO) {
	io.Varint32(&pk.EventID)
	io.Vec3(&pk.Position)
	io.Varint32(&pk.Data)
}
