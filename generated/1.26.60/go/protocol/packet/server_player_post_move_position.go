// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"

	"github.com/go-gl/mathgl/mgl32"
)

// ServerPlayerPostMovePosition is sent by the server with the player's position after movement processing.
type ServerPlayerPostMovePosition struct {
	// Position is the player's position after the server has processed movement.
	Pos mgl32.Vec3
}

// ID ...
func (*ServerPlayerPostMovePosition) ID() uint32 {
	return IDServerPlayerPostMovePosition
}

func (pk *ServerPlayerPostMovePosition) Marshal(io protocol.IO) {
	io.Vec3(&pk.Pos)
}
