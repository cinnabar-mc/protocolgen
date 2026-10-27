// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// BiomeDefinitionList is sent by the server to let the client know all biomes that are available and
// implemented on the server side. When enabled, it also includes information for the client to accurately
// recreate the server-side generation in vanilla worlds/servers for increased performance.
type BiomeDefinitionList struct {
	MapOfBiomeNamesToData []protocol.OrderedEntry[uint16, protocol.BiomeDefinitionData]
	// StringList is a makeshift dictionary implementation Mojang created to try and reduce the size of the
	// overall packet. It is a list of common strings that are used in the biome definitions, such as biome names,
	// float values or query expressions.
	StringList protocol.BiomeStringList
}

// ID ...
func (*BiomeDefinitionList) ID() uint32 {
	return IDBiomeDefinitionList
}

func (pk *BiomeDefinitionList) Marshal(io protocol.IO) {
	protocol.OrderedMap(io, &pk.MapOfBiomeNamesToData, io.Varuint32, io.Uint16, func(value *protocol.BiomeDefinitionData) {
		value.Marshal(io)
	})
	pk.StringList.Marshal(io)
}
