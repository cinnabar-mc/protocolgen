// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// CameraAimAssistPresets is sent by the server to the client to provide a list of categories and presets that
// can be used when sending a CameraAimAssist packet or a CameraInstruction including aim assist.
type CameraAimAssistPresets struct {
	// CameraAimAssistPresets is a list of categories which can be referenced by one of the Presets.
	CameraAimAssistPresets []protocol.CameraAimAssistCategoryDefinition
	// CameraAimAssistCategories is a list of presets which define a base for how aim assist should behave
	CameraAimAssistCategories []protocol.CameraAimAssistPresetDefinition
	// Operation is the operation to perform with the presets. It is one of the constants above.
	Operation protocol.CameraAimAssistPresetOperation
}

// ID ...
func (*CameraAimAssistPresets) ID() uint32 {
	return IDCameraAimAssistPresets
}

func (pk *CameraAimAssistPresets) Marshal(io protocol.IO) {
	protocol.Slice(io, &pk.CameraAimAssistPresets)
	protocol.Slice(io, &pk.CameraAimAssistCategories)
	pk.Operation.Marshal(io)
}
