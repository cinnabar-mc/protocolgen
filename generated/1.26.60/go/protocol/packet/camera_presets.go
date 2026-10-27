// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// CameraPresets gives the client a list of custom camera presets.
type CameraPresets struct {
	// Presets is a list of camera presets that can be used by other cameras. The order of this list is important
	// because the index of presets is used as a pointer in the CameraInstruction packet.
	Presets []protocol.CameraPreset
}

// ID ...
func (*CameraPresets) ID() uint32 {
	return IDCameraPresets
}

func (pk *CameraPresets) Marshal(io protocol.IO) {
	protocol.Slice(io, &pk.Presets)
}
