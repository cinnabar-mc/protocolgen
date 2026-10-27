// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"

	"github.com/go-gl/mathgl/mgl32"
)

// CameraAimAssist is sent by the server to the client to set up aim assist for the client's camera.
type CameraAimAssist struct {
	// PresetID is the ID of the preset that has previously been defined in the CameraAimAssistPresets packet.
	PresetID string
	// ViewAngle is the maximum angle around the playes's cursor that the aim assist should check for a target, if
	// TargetMode is set to protocol.AimAssistTargetModeAngle.
	ViewAngle mgl32.Vec2
	// Distance is the maximum distance from the player's cursor should check for a target, if TargetMode is set
	// to protocol.AimAssistTargetModeDistance.
	Distance float32
	// TargetMode is the mode that the camera should use for detecting targets. This is currently one of
	// protocol.AimAssistTargetModeAngle or protocol.AimAssistTargetModeDistance.
	TargetMode protocol.TargetMode
	// Action is the action that should be performed with the aim assist. This is one of the constants above.
	Action protocol.CameraAimAssistAction
	// ShowDebugRender specifies if debug render should be shown.
	ShowDebugRender bool
}

// ID ...
func (*CameraAimAssist) ID() uint32 {
	return IDCameraAimAssist
}

func (pk *CameraAimAssist) Marshal(io protocol.IO) {
	io.String(&pk.PresetID)
	io.Vec2(&pk.ViewAngle)
	io.Float32(&pk.Distance)
	protocol.Minimum(io, &pk.Distance, 1)
	protocol.Maximum(io, &pk.Distance, 16)
	pk.TargetMode.Marshal(io)
	pk.Action.Marshal(io)
	io.Bool(&pk.ShowDebugRender)
}
