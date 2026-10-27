// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// CameraSpline is sent by the server to define camera spline paths.
type CameraSpline struct {
	// CameraDataSplines is a list of camera spline definitions.
	CameraDataSplines []protocol.CameraSplineDefinition
}

// ID ...
func (*CameraSpline) ID() uint32 {
	return IDCameraSpline
}

func (pk *CameraSpline) Marshal(io protocol.IO) {
	protocol.Slice(io, &pk.CameraDataSplines)
}
