// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// SetScoreboardIdentity is sent by the server to change the identity type of one of the entries on a
// scoreboard. This is used to change, for example, an entry pointing to a player, to a fake player when it
// leaves the server, and to change it back to a real player when it joins again. In non-vanilla situations,
// the packet is quite useless.
type SetScoreboardIdentity struct {
	// ScoreboardIdentityPacketType is the type of the action to execute. The action is either
	// ScoreboardIdentityActionRegister to associate an identity with the entry, or ScoreboardIdentityActionClear
	// to remove associations with an entity.
	ScoreboardIdentityPacketType protocol.ScoreboardIdentityPacketType
	// ScoreboardIdentityInfo is a list of all entries in the packet. Each of these entries points to one of the
	// entries on a scoreboard. Depending on ActionType, their identity will either be registered or cleared.
	ScoreboardIdentityInfo []protocol.ScoreboardIdentityPacketInfo
}

// ID ...
func (*SetScoreboardIdentity) ID() uint32 {
	return IDSetScoreboardIdentity
}

func (pk *SetScoreboardIdentity) Marshal(io protocol.IO) {
	pk.ScoreboardIdentityPacketType.Marshal(io)
	protocol.Slice(io, &pk.ScoreboardIdentityInfo)
}
