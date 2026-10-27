// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// ServerSettingsResponse is optionally sent by the server in response to a ServerSettingsRequest from the
// client. It is structured the same as a ModalFormRequest packet, and if filled out correctly, will show a
// specific tab for the server in the settings of the client. A ModalFormResponse packet is sent by the client
// in response to a ServerSettingsResponse, when the client fills out the settings and closes the settings
// again.
type ServerSettingsResponse struct {
	// FormID is an ID used to identify the form. The ID is saved by the client and sent back when the player
	// submits the form, so that the server can identify which form was submitted.
	FormID     uint32
	FormUIJSON string
}

// ID ...
func (*ServerSettingsResponse) ID() uint32 {
	return IDServerSettingsResponse
}

func (pk *ServerSettingsResponse) Marshal(io protocol.IO) {
	io.Varuint32(&pk.FormID)
	io.String(&pk.FormUIJSON)
}
