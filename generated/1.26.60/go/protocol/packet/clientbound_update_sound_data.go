// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// ClientboundUpdateSoundData is sent by the server to update a sound that is currently playing, identified by
// the handle that the server sent in the PlaySound packet that started it. Each optional field is a Cereal
// union slot that may hold any SoundDataUpdate variant; its name does not constrain the variant on the wire.
type ClientboundUpdateSoundData struct {
	// ServerSoundHandle is the server-side handle of the sound to update.
	ServerSoundHandle protocol.ServerSoundHandle
	Event             protocol.SoundDataEvent
}

// ID ...
func (*ClientboundUpdateSoundData) ID() uint32 {
	return IDClientboundUpdateSoundData
}

func (pk *ClientboundUpdateSoundData) Marshal(io protocol.IO) {
	pk.ServerSoundHandle.Marshal(io)
	protocol.MarshalSoundDataEvent(io, &pk.Event)
}
