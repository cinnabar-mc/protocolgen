// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// SendPartyDestinationCookie is sent by the server to a client with a party destination cookie.
type SendPartyDestinationCookie struct {
	// Cookie is the opaque party destination cookie.
	Cookie string
	// Intent is the intent of the cookie. It is one of the PartyDestinationCookieIntent constants.
	Intent string
	// DestinationName is the name of the destination the cookie refers to.
	DestinationName string
}

// ID ...
func (*SendPartyDestinationCookie) ID() uint32 {
	return IDSendPartyDestinationCookie
}

func (pk *SendPartyDestinationCookie) Marshal(io protocol.IO) {
	io.StringLimits(&pk.Cookie, 0, 2048)
	io.String(&pk.Intent)
	io.StringLimits(&pk.DestinationName, 0, 64)
}
