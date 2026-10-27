// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// LecternUpdate is sent by the client to update the server on which page was opened in a book on a lectern,
// or if the book should be removed from it.
type LecternUpdate struct {
	// Page is the page number in the book that was opened by the player on the lectern.
	NewPageToShow uint8
	// PageCount is the number of pages that the book opened in the lectern has.
	TotalPages uint8
	// Position is the position of the lectern that was updated. If no lectern is at the block position, the
	// packet should be ignored.
	PositionOfLecternToUpdate protocol.BlockPos
}

// ID ...
func (*LecternUpdate) ID() uint32 {
	return IDLecternUpdate
}

func (pk *LecternUpdate) Marshal(io protocol.IO) {
	io.Uint8(&pk.NewPageToShow)
	io.Uint8(&pk.TotalPages)
	pk.PositionOfLecternToUpdate.Marshal(io)
}
