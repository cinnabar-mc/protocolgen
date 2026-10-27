// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// TrimData is sent by the server to the client when they first join the server. It contains a list of all the
// patterns and materials that can be applied via armour trims.
type TrimData struct {
	// TrimPatternList is a list of patterns that can be applied to armour. Each pattern has its own style and
	// texture that is defined through resource packs.
	TrimPatternList []protocol.TrimPattern
	// TrimMaterialList is a list of materials that can be applied to armour. These are mostly different ores that
	// have different colours for more customization.
	TrimMaterialList []protocol.TrimMaterial
}

// ID ...
func (*TrimData) ID() uint32 {
	return IDTrimData
}

func (pk *TrimData) Marshal(io protocol.IO) {
	protocol.Slice(io, &pk.TrimPatternList)
	protocol.Slice(io, &pk.TrimMaterialList)
}
