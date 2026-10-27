// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// ClientMovementPredictionSync is sent by the client to the server periodically if the client has received
// movement corrections from the server, containing information about client-predictions that are relevant to
// movement.
type ClientMovementPredictionSync struct {
	ActorDataFlag      protocol.ActorDataFlagComponent
	ActorBoundingBox   protocol.ActorDataBoundingBoxComponent
	MovementAttributes [9]float32
	// EntityUniqueID is the unique ID of the entity that the prediction data applies to.
	ActorUniqueID    int64
	ActorFlyingState bool
}

// ID ...
func (*ClientMovementPredictionSync) ID() uint32 {
	return IDClientMovementPredictionSync
}

func (pk *ClientMovementPredictionSync) Marshal(io protocol.IO) {
	pk.ActorDataFlag.Marshal(io)
	pk.ActorBoundingBox.Marshal(io)
	for index1 := range pk.MovementAttributes {
		io.Float32(&pk.MovementAttributes[index1])
	}
	io.ActorUniqueID(&pk.ActorUniqueID)
	io.Bool(&pk.ActorFlyingState)
}
