// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

type ClientboundPlayAudioContent struct {
	SharedMetadata  protocol.AudioSignedAudioContent
	PlaybackContent protocol.AudioSignedAudioContent
	PlaybackType    string
	PlaySound       protocol.AudioSound
}

// ID ...
func (*ClientboundPlayAudioContent) ID() uint32 {
	return IDClientboundPlayAudioContent
}

func (pk *ClientboundPlayAudioContent) Marshal(io protocol.IO) {
	pk.SharedMetadata.Marshal(io)
	pk.PlaybackContent.Marshal(io)
	io.String(&pk.PlaybackType)
	pk.PlaySound.Marshal(io)
}
