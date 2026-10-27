// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// SyncWorldClocks is sent by the server to initialise and synchronise world clocks with the client.
type SyncWorldClocks struct {
	Data protocol.SyncWorldClocksData
}

// ID ...
func (*SyncWorldClocks) ID() uint32 {
	return IDSyncWorldClocks
}

func (pk *SyncWorldClocks) Marshal(io protocol.IO) {
	protocol.MarshalSyncWorldClocksData(io, &pk.Data)
}
