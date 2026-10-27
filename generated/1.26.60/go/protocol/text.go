// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

type TextData interface {
	Marshaler
	tagTextData() uint32
}

// MarshalTextData reads or writes the TextData union using its canonical wire layout.
func MarshalTextData(io IO, x *TextData) {
	Union(io, x, io.Varuint32, TextData.tagTextData, func(tag uint32) TextData {
		switch tag {
		case 0:
			return new(MessageOnly)
		case 1:
			return new(AuthorAndMessage)
		case 2:
			return new(MessageAndParams)
		}
		return nil
	})
}

type TextPacketType uint8

const (
	TextPacketTypeRaw                    TextPacketType = 0
	TextPacketTypeChat                   TextPacketType = 1
	TextPacketTypeTranslate              TextPacketType = 2
	TextPacketTypePopup                  TextPacketType = 3
	TextPacketTypeJukeboxPopup           TextPacketType = 4
	TextPacketTypeTip                    TextPacketType = 5
	TextPacketTypeSystemMessage          TextPacketType = 6
	TextPacketTypeWhisper                TextPacketType = 7
	TextPacketTypeAnnouncement           TextPacketType = 8
	TextPacketTypeTextObjectWhisper      TextPacketType = 9
	TextPacketTypeTextObject             TextPacketType = 10
	TextPacketTypeTextObjectAnnouncement TextPacketType = 11
)

// Marshal reads or writes TextPacketType through its uint8 wire encoding.
func (x *TextPacketType) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type TextProcessingEventOrigin int32

const (
	TextProcessingEventOriginUnknown            TextProcessingEventOrigin = -1
	TextProcessingEventOriginServerChatPublic   TextProcessingEventOrigin = 0
	TextProcessingEventOriginServerChatWhisper  TextProcessingEventOrigin = 1
	TextProcessingEventOriginSignText           TextProcessingEventOrigin = 2
	TextProcessingEventOriginAnvilText          TextProcessingEventOrigin = 3
	TextProcessingEventOriginBookAndQuillText   TextProcessingEventOrigin = 4
	TextProcessingEventOriginCommandBlockText   TextProcessingEventOrigin = 5
	TextProcessingEventOriginBlockActorDataText TextProcessingEventOrigin = 6
	TextProcessingEventOriginJoinEventText      TextProcessingEventOrigin = 7
	TextProcessingEventOriginLeaveEventText     TextProcessingEventOrigin = 8
	TextProcessingEventOriginSlashCommandChat   TextProcessingEventOrigin = 9
	TextProcessingEventOriginCartographyText    TextProcessingEventOrigin = 10
	TextProcessingEventOriginKickCommand        TextProcessingEventOrigin = 11
	TextProcessingEventOriginTitleCommand       TextProcessingEventOrigin = 12
	TextProcessingEventOriginSummonCommand      TextProcessingEventOrigin = 13
	TextProcessingEventOriginServerForm         TextProcessingEventOrigin = 14
	TextProcessingEventOriginDataDrivenUI       TextProcessingEventOrigin = 15
)

// Marshal reads or writes TextProcessingEventOrigin through its int32 wire encoding.
func (x *TextProcessingEventOrigin) Marshal(io IO) { io.Int32((*int32)(x)) }
