// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// RequestAbility is a packet sent by the client to the server to request permission for a specific ability
// from the server. These abilities are defined above.
type RequestAbility struct {
	// Ability is the ability that the client is requesting. This is one of the constants defined in the
	// protocol/ability.go file.
	Ability   int32
	ValueType protocol.RequestAbilityType
	Bool      bool
	Float     float32
}

// ID ...
func (*RequestAbility) ID() uint32 {
	return IDRequestAbility
}

func (pk *RequestAbility) Marshal(io protocol.IO) {
	io.Varint32(&pk.Ability)
	protocol.Minimum(io, &pk.Ability, 0)
	protocol.Maximum(io, &pk.Ability, 19)
	pk.ValueType.Marshal(io)
	io.Bool(&pk.Bool)
	io.Float32(&pk.Float)
}
