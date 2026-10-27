// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// SetHud is sent by the server to set the visibility of individual HUD elements on the client.
type SetHud struct {
	// Elements is a list of HUD elements that are being modified. The values can be any of the HudElement
	// constants above.
	HudElement []protocol.HudElement
	// Visibility represents the new visibility of the specified Elements. It can be any of the HudVisibility
	// constants above.
	HudVisible protocol.HudVisibility
}

// ID ...
func (*SetHud) ID() uint32 {
	return IDSetHud
}

func (pk *SetHud) Marshal(io protocol.IO) {
	protocol.Slice(io, &pk.HudElement)
	pk.HudVisible.Marshal(io)
}
