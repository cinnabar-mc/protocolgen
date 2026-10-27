// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"

	"github.com/go-gl/mathgl/mgl32"
)

// ActorEvent is sent by the server when a particular event happens that has to do with an entity. Some of
// these events are entity-specific, for example a wolf shaking itself dry, but others are used for each
// entity, such as dying.
type ActorEvent struct {
	// EntityRuntimeID is the runtime ID of the entity. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	TargetRuntimeID uint64
	// EventType is the ID of the event to be called. It is one of the constants that can be found above.
	EventID protocol.ActorEventType
	// EventData is optional data associated with a particular event. The data has a different function for
	// different events, however most events don't use this field at all.
	Data int32
	// FireAtPosition is the position in the same world at which the event should fire. If this is not present,
	// the position entity will be used instead.
	FireAtPosition protocol.Optional[mgl32.Vec3]
}

// ID ...
func (*ActorEvent) ID() uint32 {
	return IDActorEvent
}

func (pk *ActorEvent) Marshal(io protocol.IO) {
	io.ActorRuntimeID(&pk.TargetRuntimeID)
	pk.EventID.Marshal(io)
	io.Varint32(&pk.Data)
	protocol.OptionalFunc(io, &pk.FireAtPosition, io.Vec3)
}
