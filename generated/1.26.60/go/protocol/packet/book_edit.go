// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// BookEdit is sent by the client when it edits a book. It is sent each time a modification was made and the
// player stops its typing 'session', rather than simply after closing the book.
type BookEdit struct {
	BookSlot  int32
	Operation protocol.BookEditAction
}

// ID ...
func (*BookEdit) ID() uint32 {
	return IDBookEdit
}

func (pk *BookEdit) Marshal(io protocol.IO) {
	io.Varint32(&pk.BookSlot)
	protocol.MarshalBookEditAction(io, &pk.Operation)
}
