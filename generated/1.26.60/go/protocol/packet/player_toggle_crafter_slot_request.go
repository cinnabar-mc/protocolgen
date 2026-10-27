// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// PlayerToggleCrafterSlotRequest is sent by the client when it tries to toggle the state of a slot within a
// Crafter.
type PlayerToggleCrafterSlotRequest struct {
	// PosX is the X position of the Crafter that is being modified.
	PosX int32
	// PosY is the Y position of the Crafter that is being modified.
	PosY int32
	// PosZ is the Z position of the Crafter that is being modified.
	PosZ int32
	// Slot is the index of the slot that was toggled. This should be a value between 0 and 8.
	SlotIndex uint8
	// Disabled is the new state of the slot. If true, the slot is disabled, if false, the slot is enabled.
	IsDisabled bool
}

// ID ...
func (*PlayerToggleCrafterSlotRequest) ID() uint32 {
	return IDPlayerToggleCrafterSlotRequest
}

func (pk *PlayerToggleCrafterSlotRequest) Marshal(io protocol.IO) {
	io.Int32(&pk.PosX)
	io.Int32(&pk.PosY)
	io.Int32(&pk.PosZ)
	io.Uint8(&pk.SlotIndex)
	io.Bool(&pk.IsDisabled)
}
