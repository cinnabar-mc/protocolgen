// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

type SetPassengerOfBlock struct {
	Passenger            int64
	PassengerOfBlockData protocol.Optional[protocol.PassengerOfBlockArguments]
}

// ID ...
func (*SetPassengerOfBlock) ID() uint32 {
	return IDSetPassengerOfBlock
}

func (pk *SetPassengerOfBlock) Marshal(io protocol.IO) {
	io.ActorUniqueID(&pk.Passenger)
	protocol.OptionalMarshaler(io, &pk.PassengerOfBlockData)
}
