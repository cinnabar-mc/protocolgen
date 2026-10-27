// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// ModalFormRequest is sent by the server to make the client open a form. This form may be either a modal form
// which has two options, a menu form for a selection of options and a custom form for properties.
type ModalFormRequest struct {
	// FormID is an ID used to identify the form. The ID is saved by the client and sent back when the player
	// submits the form, so that the server can identify which form was submitted.
	FormID     uint32
	FormUIJSON string
}

// ID ...
func (*ModalFormRequest) ID() uint32 {
	return IDModalFormRequest
}

func (pk *ModalFormRequest) Marshal(io protocol.IO) {
	io.Varuint32(&pk.FormID)
	io.String(&pk.FormUIJSON)
}
