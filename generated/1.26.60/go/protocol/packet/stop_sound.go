// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// StopSound is sent by the server to stop a sound playing to the player, such as a playing music disk track
// or other long-lasting sounds.
type StopSound struct {
	// SoundName is the name of the sound that should be stopped from playing. If no sound with this name is
	// currently active, the packet is ignored.
	SoundName string
	// StopAllSounds specifies if all sounds currently playing to the player should be stopped. If set to true,
	// the SoundName field may be left empty.
	StopAllSounds bool
	// StopMusicLegacy is currently unknown.
	StopMusicLegacy bool
}

// ID ...
func (*StopSound) ID() uint32 {
	return IDStopSound
}

func (pk *StopSound) Marshal(io protocol.IO) {
	io.String(&pk.SoundName)
	io.Bool(&pk.StopAllSounds)
	io.Bool(&pk.StopMusicLegacy)
}
