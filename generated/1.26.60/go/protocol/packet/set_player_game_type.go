// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// SetPlayerGameType is sent by the server to update the game type, which is otherwise known as the game mode,
// of a player.
type SetPlayerGameType struct {
	// GameType is the new game type of the player. It is one of the constants that can be found above. Some of
	// these game types require additional flags to be set in an AdventureSettings packet for the game mode to
	// obtain its full functionality.
	PlayerGameType protocol.LevelGameType
}

// ID ...
func (*SetPlayerGameType) ID() uint32 {
	return IDSetPlayerGameType
}

func (pk *SetPlayerGameType) Marshal(io protocol.IO) {
	pk.PlayerGameType.Marshal(io)
}
