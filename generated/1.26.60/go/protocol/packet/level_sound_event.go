// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"

	"github.com/go-gl/mathgl/mgl32"
)

// LevelSoundEvent is sent by the server to make any kind of built-in sound heard to a player. It is sent to,
// for example, play a stepping sound or a shear sound. The packet is also sent by the client, in which case
// it could be forwarded by the server to the other players online. If possible, the packets from the client
// should be ignored however, and the server should play them on its own accord.
type LevelSoundEvent struct {
	// SoundEvent is the type of the sound to play. It is one of the constants above. Some of the sound types
	// require additional data, which is set in the ExtraData field.
	SoundEvent string
	// Position is the position of the sound event. The player will be able to hear the direction of the sound
	// based on what position is sent here.
	Position mgl32.Vec3
	// Data is a packed integer that some sound types use to provide extra data. An example of this is the note
	// sound, which is composed of a pitch and an instrument type.
	Data int32
	// ActorIdentifier is the string entity type of the entity that emitted the sound, for example
	// 'minecraft:skeleton'. Some sound types use this entity type for additional data.
	ActorIdentifier string
	// IsBaby specifies if the sound should be that of a baby mob. It is most notably used for parrot imitations,
	// which will change based on if this field is set to true or not.
	IsBaby bool
	// IsGlobal specifies if the sound should be played relatively or not. If set to true, the sound will have
	// full volume, regardless of where the Position is, whereas if set to false, the sound's volume will be based
	// on the distance to Position.
	IsGlobal bool
	// ActorUniqueID is the unique ID of a source entity. The unique ID is a value that remains consistent across
	// different sessions of the same world, but most servers simply fill the runtime ID of the entity out for
	// this field.
	ActorUniqueID int64
	// FireAtPosition is the position in the same world at which the event should fire. If this is not present,
	// the position entity will be used instead.
	FireAtPosition protocol.Optional[mgl32.Vec3]
}

// ID ...
func (*LevelSoundEvent) ID() uint32 {
	return IDLevelSoundEvent
}

func (pk *LevelSoundEvent) Marshal(io protocol.IO) {
	io.String(&pk.SoundEvent)
	io.Vec3(&pk.Position)
	io.Varint32(&pk.Data)
	io.String(&pk.ActorIdentifier)
	io.Bool(&pk.IsBaby)
	io.Bool(&pk.IsGlobal)
	io.Int64(&pk.ActorUniqueID)
	protocol.OptionalFunc(io, &pk.FireAtPosition, io.Vec3)
}
