// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// SyncActorProperty is an alternative to synced actor data.
type SyncActorProperty struct {
	// PropertyData ...
	PropertyData []byte
}

// ID ...
func (*SyncActorProperty) ID() uint32 {
	return IDSyncActorProperty
}

func (pk *SyncActorProperty) Marshal(io protocol.IO) {
	io.NBT(&pk.PropertyData, protocol.NBTNetwork)
}
