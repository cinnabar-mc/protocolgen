// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// ServerSettingsRequest is sent by the client to request the settings specific to the server. These settings
// are shown in a separate tab client-side, and have the same structure as a custom form.
type ServerSettingsRequest struct {
}

// ID ...
func (*ServerSettingsRequest) ID() uint32 {
	return IDServerSettingsRequest
}

func (pk *ServerSettingsRequest) Marshal(io protocol.IO) {
}
