// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

type PositionTrackingDBClientRequestAction uint8

const (
	PositionTrackingDBClientRequestActionQuery PositionTrackingDBClientRequestAction = 0
)

// Marshal reads or writes PositionTrackingDBClientRequestAction through its uint8 wire encoding.
func (x *PositionTrackingDBClientRequestAction) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type PositionTrackingDBServerBroadcastAction uint8

const (
	PositionTrackingDBServerBroadcastActionUpdate   PositionTrackingDBServerBroadcastAction = 0
	PositionTrackingDBServerBroadcastActionDestroy  PositionTrackingDBServerBroadcastAction = 1
	PositionTrackingDBServerBroadcastActionNotFound PositionTrackingDBServerBroadcastAction = 2
)

// Marshal reads or writes PositionTrackingDBServerBroadcastAction through its uint8 wire encoding.
func (x *PositionTrackingDBServerBroadcastAction) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type PositionTrackingID struct {
	Value int32
}

// Marshal reads or writes PositionTrackingID using its canonical wire layout.
func (x *PositionTrackingID) Marshal(io IO) {
	io.Varint32(&x.Value)
}
