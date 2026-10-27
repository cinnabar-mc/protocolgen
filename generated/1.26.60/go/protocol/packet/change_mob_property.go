// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// ChangeMobProperty is a packet sent from the server to the client to change one of the properties of a mob
// client-side.
type ChangeMobProperty struct {
	// EntityUniqueID is the unique ID of the entity whose property is being changed.
	ActorID int64
	// Property is the name of the property being updated.
	PropertyName string
	// BoolValue is set if the property value is a bool type. If the type is not a bool, this field is ignored.
	BoolComponentValue bool
	// StringValue is set if the property value is a string type. If the type is not a string, this field is
	// ignored.
	StringComponentValue string
	// IntValue is set if the property value is an int type. If the type is not an int, this field is ignored.
	IntComponentValue int32
	// FloatValue is set if the property value is a float type. If the type is not a float, this field is ignored.
	FloatComponentValue float32
}

// ID ...
func (*ChangeMobProperty) ID() uint32 {
	return IDChangeMobProperty
}

func (pk *ChangeMobProperty) Marshal(io protocol.IO) {
	io.ActorUniqueID(&pk.ActorID)
	io.String(&pk.PropertyName)
	io.Bool(&pk.BoolComponentValue)
	io.String(&pk.StringComponentValue)
	io.Varint32(&pk.IntComponentValue)
	io.Float32(&pk.FloatComponentValue)
}
