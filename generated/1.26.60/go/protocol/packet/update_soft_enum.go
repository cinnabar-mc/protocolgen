// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// UpdateSoftEnum is sent by the server to update a soft enum, also known as a dynamic enum, previously sent
// in the AvailableCommands packet. It is sent whenever the enum should get new options or when some of its
// options should be removed. The UpdateSoftEnum packet will apply for enums that have been set in the
// AvailableCommands packet with the 'Dynamic' field of the CommandEnum set to true.
type UpdateSoftEnum struct {
	// EnumName is the type of the enum. This type must be identical to the one set in the AvailableCommands
	// packet, because the client uses this to recognise which enum to update.
	EnumName string
	// Values is a list of options that should be updated. Depending on the ActionType field, either these options
	// will be added to the enum, the enum options will be set to these options or all of these options will be
	// removed from the enum.
	Values []string
	// UpdateType is the type of the action to execute on the enum. The Options field has a different result,
	// depending on what ActionType is used.
	UpdateType protocol.SoftEnumUpdateType
}

// ID ...
func (*UpdateSoftEnum) ID() uint32 {
	return IDUpdateSoftEnum
}

func (pk *UpdateSoftEnum) Marshal(io protocol.IO) {
	io.String(&pk.EnumName)
	protocol.FuncSlice(io, &pk.Values, io.Varuint32, io.String)
	pk.UpdateType.Marshal(io)
}
