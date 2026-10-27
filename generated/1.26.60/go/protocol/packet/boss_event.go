// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// BossEvent is sent by the server to make a specific 'boss event' occur in the world. It includes features
// such as showing a boss bar to the player and turning the sky dark.
type BossEvent struct {
	// BossEntityUniqueID is the unique ID of the boss entity that the boss event sent involves. By default, the
	// health percentage and title of the boss bar depend on the health and name tag of this entity. If
	// BossEntityUniqueID is the same as the client's entity unique ID, its HealthPercentage and BossBarTitle can
	// be freely altered.
	TargetActorID int64
	// EventType is the type of the event. It is one of the BossEvent constants above.
	EventType protocol.BossEventUpdateType
	// BossBarTitle is the title shown above the boss bar. It may be set to set a different title if the
	// BossEntityUniqueID matches the client's entity unique ID.
	Name string
	// FilteredBossBarTitle is a filtered version of BossBarTitle with all the profanity removed. The client will
	// use this over BossBarTitle if this field is not empty and they have the "Filter Profanity" setting enabled.
	FilteredName string
	// HealthPercentage is the percentage of health that is shown in the boss bar (0.0-1.0). The HealthPercentage
	// may be set to a specific value if the BossEntityUniqueID matches the client's entity unique ID.
	HealthPercent float32
	// Colour is the colour of the boss bar that is shown when a player is subscribed. It is one of the
	// BossEventColour constants listed above.
	Color protocol.BossBarColor
	// Overlay is the overlay of the boss bar that is shown on top of the boss bar when a player is subscribed. It
	// is one of the BossEventOverlay constants listed above.
	Overlay protocol.BossBarOverlay
}

// ID ...
func (*BossEvent) ID() uint32 {
	return IDBossEvent
}

func (pk *BossEvent) Marshal(io protocol.IO) {
	io.ActorUniqueID(&pk.TargetActorID)
	pk.EventType.Marshal(io)
	io.StringLimits(&pk.Name, 0, 256)
	io.StringLimits(&pk.FilteredName, 0, 256)
	io.Float32(&pk.HealthPercent)
	pk.Color.Marshal(io)
	pk.Overlay.Marshal(io)
}
