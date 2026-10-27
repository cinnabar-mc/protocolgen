// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// ServerPresenceInfo is sent by the server to provide the client with presence info.
type ServerPresenceInfo struct {
	// PresenceConfiguration is the presence info to set, or nothing to fall back to the default.
	PresenceConfiguration protocol.Optional[protocol.ServerConfigurationPresenceConfiguration]
}

// ID ...
func (*ServerPresenceInfo) ID() uint32 {
	return IDServerPresenceInfo
}

func (pk *ServerPresenceInfo) Marshal(io protocol.IO) {
	protocol.OptionalMarshaler(io, &pk.PresenceConfiguration)
}
