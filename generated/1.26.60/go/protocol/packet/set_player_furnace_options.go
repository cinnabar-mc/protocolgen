// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// SetPlayerFurnaceOptions is a bidirectional packet that can be used to update the options a player has
// selected in a furnace-like container's UI.
type SetPlayerFurnaceOptions struct {
	// FurnaceType is the type of furnace that the options apply to. It is one of the FurnaceType constants above.
	FurnaceType protocol.FurnaceType
	// FurnaceOptions holds the options that the player has selected for the furnace type above.
	FurnaceOptions protocol.FurnaceOptions
}

// ID ...
func (*SetPlayerFurnaceOptions) ID() uint32 {
	return IDSetPlayerFurnaceOptions
}

func (pk *SetPlayerFurnaceOptions) Marshal(io protocol.IO) {
	pk.FurnaceType.Marshal(io)
	pk.FurnaceOptions.Marshal(io)
}
