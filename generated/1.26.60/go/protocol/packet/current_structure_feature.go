// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// CurrentStructureFeature is sent by the server to let the client know the name of the structure feature that
// the player is currently occupying.
type CurrentStructureFeature struct {
	// CurrentStructureFeature is the identifier of the structure feature that the player is currently occupying.
	// If the player is not occupying any structure feature, this field is empty.
	CurrentStructureFeature string
}

// ID ...
func (*CurrentStructureFeature) ID() uint32 {
	return IDCurrentStructureFeature
}

func (pk *CurrentStructureFeature) Marshal(io protocol.IO) {
	io.String(&pk.CurrentStructureFeature)
}
