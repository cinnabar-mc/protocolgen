// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

type ServerboundRegisterAudioContent struct {
	Registrations []protocol.AudioContentRegistrationEntry
}

// ID ...
func (*ServerboundRegisterAudioContent) ID() uint32 {
	return IDServerboundRegisterAudioContent
}

func (pk *ServerboundRegisterAudioContent) Marshal(io protocol.IO) {
	protocol.SliceLimits(io, &pk.Registrations, 0, 64)
}
