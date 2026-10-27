// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// PlayerUpdateEntityOverrides is sent by the server to modify an entity's properties individually.
type PlayerUpdateEntityOverrides struct {
	// EntityUniqueID is the unique ID of the entity. The unique ID is a value that remains consistent across
	// different sessions of the same world, but most servers simply fill the runtime ID of the entity out for
	// this field.
	TargetID int64
	// PropertyIndex is the index of the property to modify. The index is unique for each property of an entity.
	PropertyIndex uint32
	Update        protocol.PlayerUpdateEntityOverridesData
}

// ID ...
func (*PlayerUpdateEntityOverrides) ID() uint32 {
	return IDPlayerUpdateEntityOverrides
}

func (pk *PlayerUpdateEntityOverrides) Marshal(io protocol.IO) {
	io.ActorUniqueID(&pk.TargetID)
	io.Varuint32(&pk.PropertyIndex)
	protocol.MarshalPlayerUpdateEntityOverridesData(io, &pk.Update)
}
