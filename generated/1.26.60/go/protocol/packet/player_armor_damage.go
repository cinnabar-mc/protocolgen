// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// PlayerArmourDamage is sent by the server to damage the armour of a player. It is a very efficient packet,
// but generally it's much easier to just send a slot update for the damaged armour.
type PlayerArmorDamage struct {
	// List is a list of armour entries indicating which pieces of armour should receive damage.
	ArmorSlotAndDamagePairs []protocol.ArmorSlotAndDamagePair
}

// ID ...
func (*PlayerArmorDamage) ID() uint32 {
	return IDPlayerArmorDamage
}

func (pk *PlayerArmorDamage) Marshal(io protocol.IO) {
	protocol.SliceLimits(io, &pk.ArmorSlotAndDamagePairs, 0, 5)
}
