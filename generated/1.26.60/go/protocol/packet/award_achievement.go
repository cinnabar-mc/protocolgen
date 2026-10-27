// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// AwardAchievement is sent by the server to award an achievement to a player.
type AwardAchievement struct {
	// AchievementID is the ID of the achievement that should be awarded to the player. The values for these IDs
	// are currently unknown.
	AchievementID int32
}

// ID ...
func (*AwardAchievement) ID() uint32 {
	return IDAwardAchievement
}

func (pk *AwardAchievement) Marshal(io protocol.IO) {
	io.Int32(&pk.AchievementID)
}
