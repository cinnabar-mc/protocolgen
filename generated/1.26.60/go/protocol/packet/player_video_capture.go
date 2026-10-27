// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// PlayerVideoCapture packet is sent by the server to start or stop video recording for a player. This packet
// only works on development builds and has no effect on retail builds. When recording, the client will save
// individual frames to '/LocalCache/minecraftpe' in the format specified below.
type PlayerVideoCapture struct {
	// Action is the action to perform with the video capture. It is one of the constants above.
	Action protocol.PlayerVideoCaptureData
}

// ID ...
func (*PlayerVideoCapture) ID() uint32 {
	return IDPlayerVideoCapture
}

func (pk *PlayerVideoCapture) Marshal(io protocol.IO) {
	protocol.MarshalPlayerVideoCaptureData(io, &pk.Action)
}
