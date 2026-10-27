// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"

	"github.com/go-gl/mathgl/mgl32"
)

// ChangeDimension is sent by the server to the client to send a dimension change screen client-side. Once the
// screen is cleared client-side, the client will send a PlayerAction packet with
// PlayerActionDimensionChangeDone.
type ChangeDimension struct {
	DimensionID protocol.DimensionType
	// Position is the position in the new dimension that the player is spawned in.
	Position mgl32.Vec3
	// Respawn specifies if the dimension change was respawn based, meaning that the player died in one dimension
	// and got respawned into another. The client will send a PlayerAction packet with
	// PlayerActionDimensionChangeRequest if it dies in another dimension, indicating that it needs a
	// DimensionChange packet with Respawn set to true.
	Respawn bool
	// LoadingScreenID is a unique ID for the loading screen that the player is currently in. The client will
	// update the server on its state through the ServerBoundLoadingScreen packet, and it can be used to not send
	// specific packets to the client if it is changing dimensions. This field should be unique for every
	// ChangeDimension packet sent.
	LoadingScreenID protocol.Optional[uint32]
}

// ID ...
func (*ChangeDimension) ID() uint32 {
	return IDChangeDimension
}

func (pk *ChangeDimension) Marshal(io protocol.IO) {
	pk.DimensionID.Marshal(io)
	io.Vec3(&pk.Position)
	io.Bool(&pk.Respawn)
	protocol.OptionalFunc(io, &pk.LoadingScreenID, io.Uint32)
}
