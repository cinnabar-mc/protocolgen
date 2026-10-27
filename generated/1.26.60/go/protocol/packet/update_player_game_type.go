// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// UpdatePlayerGameType is sent by the server to change the game mode of a player. It is functionally
// identical to the SetPlayerGameType packet.
type UpdatePlayerGameType struct {
	// GameType is the new game type of the player. It is one of the constants that can be found in
	// set_player_game_type.go. Some of these game types require additional flags to be set in an UpdateAbilities
	// packet for the game mode to obtain its full functionality.
	PlayerGameType protocol.LevelGameType
	// PlayerUniqueID is the entity unique ID of the player that should have its game mode updated. If this packet
	// is sent to other clients with the player unique ID of another player, nothing happens.
	TargetPlayer int64
	// Tick is the server tick at which the packet was sent. It is used in relation to
	// CorrectPlayerMovePrediction.
	Tick uint64
}

// ID ...
func (*UpdatePlayerGameType) ID() uint32 {
	return IDUpdatePlayerGameType
}

func (pk *UpdatePlayerGameType) Marshal(io protocol.IO) {
	pk.PlayerGameType.Marshal(io)
	io.ActorUniqueID(&pk.TargetPlayer)
	io.PlayerInputTick(&pk.Tick)
}
