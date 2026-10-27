// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// ResourcePacksReadyForValidation is sent by the client to inform the server that the client has finished
// loading resource packs and is ready for validation.
type ResourcePacksReadyForValidation struct {
}

// ID ...
func (*ResourcePacksReadyForValidation) ID() uint32 {
	return IDResourcePacksReadyForValidation
}

func (pk *ResourcePacksReadyForValidation) Marshal(io protocol.IO) {
}
