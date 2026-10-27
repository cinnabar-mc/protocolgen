// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// MapInfoRequest is sent by the client to request the server to deliver information of a certain map in the
// inventory of the player. The server should respond with a ClientBoundMapItemData packet.
type MapInfoRequest struct {
	// MapID is the unique identifier that represents the map that is requested over network. It remains
	// consistent across sessions.
	MapUniqueID int64
	// ClientPixels is a slice of pixels sent from the client to notify the server about the pixels that it isn't
	// aware of.
	ClientPixelsList []protocol.PixelRequest
}

// ID ...
func (*MapInfoRequest) ID() uint32 {
	return IDMapInfoRequest
}

func (pk *MapInfoRequest) Marshal(io protocol.IO) {
	io.ActorUniqueID(&pk.MapUniqueID)
	protocol.FuncSliceLimits(io, &pk.ClientPixelsList, io.Uint32, 0, 16384, func(value *protocol.PixelRequest) {
		value.Marshal(io)
	})
}
