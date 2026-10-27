// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// ClientboundUpdateSoundData updates a playing sound by its server handle. The event starts with one byte
// selecting the update, followed by that update's payload.
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
