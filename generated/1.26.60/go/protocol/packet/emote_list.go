// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"

	"github.com/google/uuid"
)

// EmoteList is sent by the client every time it joins the server and when it equips new emotes. It may be
// used by the server to find out which emotes the client has available. If the player has no emotes equipped,
// this packet is not sent. Under certain circumstances, this packet is also sent from the server to the
// client, but I was unable to find when this is done.
type EmoteList struct {
	// PlayerRuntimeID is the runtime ID of the player that owns the emote pieces below. If sent by the client,
	// this player runtime ID is always that of the player itself.
	RuntimeID uint64
	// EmotePieces is a list of emote pieces that the player with the runtime ID above has.
	EmotePieceIds []uuid.UUID
}

// ID ...
func (*EmoteList) ID() uint32 {
	return IDEmoteList
}

func (pk *EmoteList) Marshal(io protocol.IO) {
	io.ActorRuntimeID(&pk.RuntimeID)
	protocol.FuncSlice(io, &pk.EmotePieceIds, io.Varuint32, io.UUID)
}
