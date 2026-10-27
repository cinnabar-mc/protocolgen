// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// ModalFormResponse is sent by the client in response to a ModalFormRequest, after the player has submitted
// the form sent. It contains the options/properties selected by the player, or a JSON encoded 'null' if the
// form was closed by clicking the X at the top right corner of the form.
type ModalFormResponse struct {
	// FormID is the form ID of the form the client has responded to. It is the same as the ID sent in the
	// ModalFormRequest, and may be used to identify which form was submitted.
	FormID uint32
	// JSONResponse is a JSON encoded value representing the response of the player. For a modal form, the
	// response is either true or false, for a menu form, the response is an integer specifying the index of the
	// button clicked, and for a custom form, the response is an array containing a value for each element.
	JSONResponse protocol.Optional[string]
	// FormCancelReason represents the reason why the form was cancelled. It is one of the constants above.
	FormCancelReason protocol.Optional[protocol.ModalFormCancelReason]
}

// ID ...
func (*ModalFormResponse) ID() uint32 {
	return IDModalFormResponse
}

func (pk *ModalFormResponse) Marshal(io protocol.IO) {
	io.Varuint32(&pk.FormID)
	protocol.OptionalFunc(io, &pk.JSONResponse, io.String)
	protocol.OptionalMarshaler(io, &pk.FormCancelReason)
}
