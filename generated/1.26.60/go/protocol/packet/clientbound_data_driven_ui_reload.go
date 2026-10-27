// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// ClientBoundDataDrivenUIReload is sent by the server to reload the data-driven UI on the client.
type ClientboundDataDrivenUIReload struct {
}

// ID ...
func (*ClientboundDataDrivenUIReload) ID() uint32 {
	return IDClientboundDataDrivenUIReload
}

func (pk *ClientboundDataDrivenUIReload) Marshal(io protocol.IO) {
}
