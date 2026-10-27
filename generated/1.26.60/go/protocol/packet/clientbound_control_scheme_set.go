// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// ClientBoundControlSchemeSet is sent by the server upon the client's request or the usage of the vanilla
// /controlscheme command. It is used to set the control scheme of the client, often used in combination with
// custom cameras.
type ClientboundControlSchemeSet struct {
	// ControlScheme is the control scheme that the client should use. It is one of the following: -
	// ControlSchemeLockedPlayerRelativeStrafe is the default behaviour, this cannot be set when the client is in
	// a custom camera. - ControlSchemeCameraRelative makes movement relative to the camera's transform, with the
	// client's rotation being relative to the client's movement. - ControlSchemeCameraRelativeStrafe makes
	// movement relative to the camera's transform, with the client's rotation being locked. -
	// ControlSchemePlayerRelative makes movement relative to the player's transform, meaning holding left/right
	// will make the player turn in a circle. - ControlSchemePlayerRelativeStrafe makes movement the same as the
	// default behaviour, but can be used in a custom camera.
	ControlScheme protocol.ControlScheme
}

// ID ...
func (*ClientboundControlSchemeSet) ID() uint32 {
	return IDClientboundControlSchemeSet
}

func (pk *ClientboundControlSchemeSet) Marshal(io protocol.IO) {
	pk.ControlScheme.Marshal(io)
}
