// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// ScriptMessage is used to communicate custom messages from the client to the server, or from the server to
// the client. While the name may suggest this packet is used for the discontinued scripting API, it is likely
// instead for the GameTest framework.
type ScriptMessage struct {
	// Identifier is the identifier of the message, used by either party to identify the message data sent.
	MessageID string
	// Data contains the data of the message.
	MessageValue []byte
}

// ID ...
func (*ScriptMessage) ID() uint32 {
	return IDScriptMessage
}

func (pk *ScriptMessage) Marshal(io protocol.IO) {
	io.String(&pk.MessageID)
	io.ByteSlice(&pk.MessageValue)
}
