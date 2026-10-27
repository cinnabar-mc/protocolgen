// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// ClientBoundTextureShift is sent by the server to control texture shift animations on the client.
type ClientboundTextureShift struct {
	// ActionID is the texture shift action to perform. It is one of the constants above.
	ActionID protocol.ClientboundTextureShiftAction
	// CollectionName is the name of the texture shift collection.
	CollectionName string
	// FromStep is the step to shift from.
	FromStep string
	// ToStep is the step to shift to.
	ToStep string
	// AllSteps is a list of all steps in the texture shift.
	AllSteps []string
	// CurrentLengthTicks is the current length of the shift in ticks.
	CurrentLengthInTicks uint64
	// TotalLengthTicks is the total length of the shift in ticks.
	TotalLengthInTicks uint64
	// Enabled specifies if the texture shift is enabled.
	Enabled bool
}

// ID ...
func (*ClientboundTextureShift) ID() uint32 {
	return IDClientboundTextureShift
}

func (pk *ClientboundTextureShift) Marshal(io protocol.IO) {
	pk.ActionID.Marshal(io)
	io.String(&pk.CollectionName)
	io.String(&pk.FromStep)
	io.String(&pk.ToStep)
	protocol.FuncSlice(io, &pk.AllSteps, io.Varuint32, io.String)
	io.Varuint64(&pk.CurrentLengthInTicks)
	io.Varuint64(&pk.TotalLengthInTicks)
	io.Bool(&pk.Enabled)
}
