// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// ServerBoundDataStore is sent by the client to update a data store property on the server.
type ServerboundDataStore struct {
	// Update contains the data store update.
	Update protocol.BedrockDDUIDataStoreUpdate
}

// ID ...
func (*ServerboundDataStore) ID() uint32 {
	return IDServerboundDataStore
}

func (pk *ServerboundDataStore) Marshal(io protocol.IO) {
	pk.Update.Marshal(io)
}
