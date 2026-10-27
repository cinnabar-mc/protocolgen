// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

type ServerboundCursorItemDrag struct {
	State protocol.CursorItemDragState
}

// ID ...
func (*ServerboundCursorItemDrag) ID() uint32 {
	return IDServerboundCursorItemDrag
}

func (pk *ServerboundCursorItemDrag) Marshal(io protocol.IO) {
	pk.State.Marshal(io)
}
