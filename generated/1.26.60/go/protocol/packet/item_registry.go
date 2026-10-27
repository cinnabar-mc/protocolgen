// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// ItemRegistry is sent by the server to send the client a list of available items and attach client-side
// components to a custom item. This packet was formerly known as the ItemComponent packet before 1.21.60,
// which did not include item definitions but only the components.
type ItemRegistry struct {
	// ItemData is a list of all items with their legacy IDs which are available in the game. Failing to send any
	// of the items that are in the game will crash mobile clients. Any custom components are also attached to the
	// items in this list.
	ItemData []protocol.ItemData
}

// ID ...
func (*ItemRegistry) ID() uint32 {
	return IDItemRegistry
}

func (pk *ItemRegistry) Marshal(io protocol.IO) {
	protocol.Slice(io, &pk.ItemData)
}
