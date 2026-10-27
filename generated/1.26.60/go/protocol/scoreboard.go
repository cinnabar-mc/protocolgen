// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

type ScoreboardID struct {
	ScoreboardID int64
}

// Marshal reads or writes ScoreboardID using its canonical wire layout.
func (x *ScoreboardID) Marshal(io IO) {
	io.Varint64(&x.ScoreboardID)
}

type ScoreboardIdentityPacketInfo struct {
	ScoreboardID   ScoreboardID
	PlayerUniqueID Optional[int64]
}

// Marshal reads or writes ScoreboardIdentityPacketInfo using its canonical wire layout.
func (x *ScoreboardIdentityPacketInfo) Marshal(io IO) {
	x.ScoreboardID.Marshal(io)
	OptionalFunc(io, &x.PlayerUniqueID, io.Varint64)
}

type ScoreboardIdentityPacketType uint8

const (
	ScoreboardIdentityPacketTypeUpdate ScoreboardIdentityPacketType = 0
	ScoreboardIdentityPacketTypeRemove ScoreboardIdentityPacketType = 1
)

// Marshal reads or writes ScoreboardIdentityPacketType through its uint8 wire encoding.
func (x *ScoreboardIdentityPacketType) Marshal(io IO) { io.Uint8((*uint8)(x)) }
