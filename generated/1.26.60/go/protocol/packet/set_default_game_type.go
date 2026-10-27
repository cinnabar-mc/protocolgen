// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// SetDefaultGameType is sent by the client when it toggles the default game type in the settings UI, and is
// sent by the server when it actually changes the default game type, resulting in the toggle being changed in
// the settings UI.
type SetDefaultGameType struct {
	// GameType is the new game type that is set. When sent by the client, this is the requested new default game
	// type.
	DefaultGameType protocol.GameType
}

// ID ...
func (*SetDefaultGameType) ID() uint32 {
	return IDSetDefaultGameType
}

func (pk *SetDefaultGameType) Marshal(io protocol.IO) {
	pk.DefaultGameType.Marshal(io)
}
