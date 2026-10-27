// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// ClientBoundDataStore is sent by the server to update, change or remove data store entries on the client.
type ClientboundDataStore struct {
	// Updates is an array of data store changes. Each entry has its own change type discriminator.
	Updates []protocol.BedrockDDUI
}

// ID ...
func (*ClientboundDataStore) ID() uint32 {
	return IDClientboundDataStore
}

func (pk *ClientboundDataStore) Marshal(io protocol.IO) {
	protocol.FuncSliceLimits(io, &pk.Updates, io.Varuint32, 0, 500, func(value *protocol.BedrockDDUI) {
		protocol.MarshalBedrockDDUI(io, value)
	})
}
