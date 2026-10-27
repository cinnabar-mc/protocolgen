// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// UpdateAttributes is sent by the server to update an amount of attributes of any entity in the world. These
// attributes include ones such as the health or the movement speed of the entity.
type UpdateAttributes struct {
	// EntityRuntimeID is the runtime ID of the entity. The runtime ID is unique for each world session, and
	// entities are generally identified in packets using this runtime ID.
	TargetRuntimeID uint64
	// Attributes is a slice of new attributes that the entity gets. It includes attributes such as its health,
	// movement speed, etc. Note that only changed attributes have to be sent in this packet. It is not required
	// to send attributes that did not have their values changed.
	AttributeList []protocol.AttributeData
	// Tick is the server tick at which the packet was sent. It is used in relation to
	// CorrectPlayerMovePrediction.
	Tick uint64
}

// ID ...
func (*UpdateAttributes) ID() uint32 {
	return IDUpdateAttributes
}

func (pk *UpdateAttributes) Marshal(io protocol.IO) {
	io.ActorRuntimeID(&pk.TargetRuntimeID)
	protocol.Slice(io, &pk.AttributeList)
	io.PlayerInputTick(&pk.Tick)
}
