// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// RecordStarted is sent by the server to notify the client that a record started playing at a specific block
// position, such as when a music disc is inserted into a jukebox.
type RecordStarted struct {
	// Position is the position of the block that the record started playing at.
	BlockPosition     protocol.BlockPos
	ServerSoundHandle protocol.ServerSoundHandle
}

// ID ...
func (*RecordStarted) ID() uint32 {
	return IDRecordStarted
}

func (pk *RecordStarted) Marshal(io protocol.IO) {
	pk.BlockPosition.Marshal(io)
	pk.ServerSoundHandle.Marshal(io)
}
