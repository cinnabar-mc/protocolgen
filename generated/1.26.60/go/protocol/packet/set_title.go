// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// SetTitle is sent by the server to make a title, subtitle or action bar shown to a player. It has several
// fields that allow setting the duration of the titles.
type SetTitle struct {
	// ActionType is the type of the action that should be executed upon the title of a player. It is one of the
	// constants above and specifies the response of the client to the packet.
	TitleType protocol.TitleType
	// Text is the text of the title, which has a different meaning depending on the ActionType that the packet
	// has. The text is the text of a title, subtitle or action bar, depending on the type set.
	TitleText string
	// FadeInDuration is the duration that the title takes to fade in on the screen of the player. It is measured
	// in 20ths of a second (AKA in ticks).
	FadeInTime int32
	// RemainDuration is the duration that the title remains on the screen of the player. It is measured in 20ths
	// of a second (AKA in ticks).
	StayTime int32
	// FadeOutDuration is the duration that the title takes to fade out of the screen of the player. It is
	// measured in 20ths of a second (AKA in ticks).
	FadeOutTime int32
	// Xuid is the XBOX Live user ID of the player, which will remain consistent as long as the player is logged
	// in with the XBOX Live account. It is empty if the user is not logged into its XBL account.
	Xuid string
	// PlatformOnlineID is either a uint64 or an empty string.
	PlatformOnlineID string
	// FilteredMessage is a filtered version of Message with all the profanity removed. The client will use this
	// over Message if this field is not empty and they have the "Filter Profanity" setting enabled.
	FilteredTitleMessage string
}

// ID ...
func (*SetTitle) ID() uint32 {
	return IDSetTitle
}

func (pk *SetTitle) Marshal(io protocol.IO) {
	pk.TitleType.Marshal(io)
	io.String(&pk.TitleText)
	io.Varint32(&pk.FadeInTime)
	io.Varint32(&pk.StayTime)
	io.Varint32(&pk.FadeOutTime)
	io.String(&pk.Xuid)
	io.String(&pk.PlatformOnlineID)
	io.String(&pk.FilteredTitleMessage)
}
