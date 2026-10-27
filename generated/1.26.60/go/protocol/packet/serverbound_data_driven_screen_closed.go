// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// ServerBoundDataDrivenScreenClosed is sent by the client when a data-driven UI screen is closed.
type ServerboundDataDrivenScreenClosed struct {
	// FormID is the unique instance ID of the form that was closed.
	FormID uint32
	// CloseReason is the reason the screen was closed. It is one of the DataDrivenScreenCloseReason constants.
	CloseReason string
}

// ID ...
func (*ServerboundDataDrivenScreenClosed) ID() uint32 {
	return IDServerboundDataDrivenScreenClosed
}

func (pk *ServerboundDataDrivenScreenClosed) Marshal(io protocol.IO) {
	io.Uint32(&pk.FormID)
	io.String(&pk.CloseReason)
}
