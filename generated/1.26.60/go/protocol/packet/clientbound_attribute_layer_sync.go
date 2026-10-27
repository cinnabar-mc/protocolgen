// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// ClientBoundAttributeLayerSync is sent by the server to synchronise attribute layers with the client.
type ClientboundAttributeLayerSync struct {
	Data protocol.AttributeLayerSyncData
}

// ID ...
func (*ClientboundAttributeLayerSync) ID() uint32 {
	return IDClientboundAttributeLayerSync
}

func (pk *ClientboundAttributeLayerSync) Marshal(io protocol.IO) {
	protocol.MarshalAttributeLayerSyncData(io, &pk.Data)
}
