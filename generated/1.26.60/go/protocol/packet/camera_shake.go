// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// CameraShake is sent by the server to make the camera shake client-side. This feature was added for map-
// making partners.
type CameraShake struct {
	// Intensity is the intensity of the shaking. The client limits this value to 4, so anything higher may not
	// work.
	Intensity float32
	// Duration is the number of seconds the camera will shake for.
	Seconds float32
	// Type is the type of shake, and is one of the constants listed above. The different type affects how the
	// shake looks in game.
	ShakeType protocol.CameraShakeType
	// Action is the action to be performed, and is one of the constants listed above. Currently the different
	// actions will either add or stop shaking the client.
	ShakeAction protocol.CameraShakeAction
}

// ID ...
func (*CameraShake) ID() uint32 {
	return IDCameraShake
}

func (pk *CameraShake) Marshal(io protocol.IO) {
	io.Float32(&pk.Intensity)
	io.Float32(&pk.Seconds)
	pk.ShakeType.Marshal(io)
	pk.ShakeAction.Marshal(io)
}
