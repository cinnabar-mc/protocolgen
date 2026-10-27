// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// TickingAreasLoadStatus is sent by the server to the client to notify the client of a ticking area's loading
// status.
type TickingAreasLoadStatus struct {
	// WaitingForPreload is true if the server is waiting for the area's preload.
	WaitingForPreload bool
}

// ID ...
func (*TickingAreasLoadStatus) ID() uint32 {
	return IDTickingAreasLoadStatus
}

func (pk *TickingAreasLoadStatus) Marshal(io protocol.IO) {
	io.Bool(&pk.WaitingForPreload)
}
