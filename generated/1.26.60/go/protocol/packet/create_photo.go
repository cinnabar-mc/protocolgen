// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// CreatePhoto is a packet that allows players to export photos from their portfolios into items in their
// inventory. This packet only works on the Education Edition version of Minecraft.
type CreatePhoto struct {
	// EntityUniqueID is the unique ID of the entity.
	RawID uint64
	// PhotoName is the name of the photo.
	PhotoName string
	// ItemName is the name of the photo as an item.
	PhotoItemName string
}

// ID ...
func (*CreatePhoto) ID() uint32 {
	return IDCreatePhoto
}

func (pk *CreatePhoto) Marshal(io protocol.IO) {
	io.Uint64(&pk.RawID)
	io.String(&pk.PhotoName)
	io.String(&pk.PhotoItemName)
}
