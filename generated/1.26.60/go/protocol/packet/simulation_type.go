// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// SimulationType is an in-progress packet. We currently do not know the use case.
type SimulationType struct {
	// SimType is the simulation type selected.
	SimType protocol.SimulationTypeEnum
}

// ID ...
func (*SimulationType) ID() uint32 {
	return IDSimulationType
}

func (pk *SimulationType) Marshal(io protocol.IO) {
	pk.SimType.Marshal(io)
}
