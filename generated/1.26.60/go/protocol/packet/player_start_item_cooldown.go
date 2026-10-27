// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

type PlayerStartItemCooldown struct {
	ItemCategory  string
	DurationTicks int32
}

// ID ...
func (*PlayerStartItemCooldown) ID() uint32 {
	return IDPlayerStartItemCooldown
}

func (pk *PlayerStartItemCooldown) Marshal(io protocol.IO) {
	io.String(&pk.ItemCategory)
	io.Varint32(&pk.DurationTicks)
}
