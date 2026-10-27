// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

type ClientboundMatchmakingState struct {
	State           protocol.MatchmakingState
	DestinationName string
	Options         protocol.Optional[protocol.MatchmakingStateOptions]
}

// ID ...
func (*ClientboundMatchmakingState) ID() uint32 {
	return IDClientboundMatchmakingState
}

func (pk *ClientboundMatchmakingState) Marshal(io protocol.IO) {
	pk.State.Marshal(io)
	io.StringLimits(&pk.DestinationName, 0, 100)
	protocol.OptionalMarshaler(io, &pk.Options)
}
