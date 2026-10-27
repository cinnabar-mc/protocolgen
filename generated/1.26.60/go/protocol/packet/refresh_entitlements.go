// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// RefreshEntitlements is sent by the client to the server to refresh the entitlements of the player.
type RefreshEntitlements struct {
}

// ID ...
func (*RefreshEntitlements) ID() uint32 {
	return IDRefreshEntitlements
}

func (pk *RefreshEntitlements) Marshal(io protocol.IO) {
}
