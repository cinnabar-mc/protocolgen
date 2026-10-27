// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// Text is sent by the client to the server to send chat messages, and by the server to the client to forward
// or send messages, which may be chat, popups, tips etc.
type Text struct {
	Localize bool
	Body     protocol.TextData
	// XUID is the XBOX Live user ID of the player that sent the message. It is only set for packets of
	// TextTypeChat. When sent to a player, the player will only be shown the chat message if a player with this
	// XUID is present in the player list and not muted, or if the XUID is empty.
	SenderSXUID string
	PlatformID  string
	// FilteredMessage is a filtered version of Message with all the profanity removed. The client will use this
	// over Message if this field is not empty and they have the "Filter Profanity" setting enabled.
	FilteredMessage protocol.Optional[string]
}

// ID ...
func (*Text) ID() uint32 {
	return IDText
}

func (pk *Text) Marshal(io protocol.IO) {
	io.Bool(&pk.Localize)
	protocol.MarshalTextData(io, &pk.Body)
	io.StringLimits(&pk.SenderSXUID, 0, 64)
	io.StringLimits(&pk.PlatformID, 0, 256)
	protocol.OptionalFunc(io, &pk.FilteredMessage, func(value *string) {
		io.StringLimits(value, 0, 65536)
	})
}
