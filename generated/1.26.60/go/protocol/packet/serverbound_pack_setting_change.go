// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"

	"github.com/google/uuid"
)

// ServerBoundPackSettingChange is sent by the client to the server when it changes a setting for a specific
// pack in the pack settings UI.
type ServerboundPackSettingChange struct {
	// PackID is the UUID of the pack.
	PackID          uuid.UUID
	PackSettingName string
	// PackSetting is the new setting value applied to the pack.
	PackSettingValue protocol.ServerboundPackSettingChangePackSettingValue
}

// ID ...
func (*ServerboundPackSettingChange) ID() uint32 {
	return IDServerboundPackSettingChange
}

func (pk *ServerboundPackSettingChange) Marshal(io protocol.IO) {
	io.UUID(&pk.PackID)
	io.StringLimits(&pk.PackSettingName, 0, 128)
	protocol.MarshalServerboundPackSettingChangePackSettingValue(io, &pk.PackSettingValue)
}
