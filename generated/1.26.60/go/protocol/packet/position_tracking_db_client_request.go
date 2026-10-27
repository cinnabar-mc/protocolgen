// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// PositionTrackingDBClientRequest is a packet sent by the client to request the position and dimension of a
// 'tracking ID'. These IDs are tracked in a database by the server. In 1.16, this is used for lodestones. The
// client will send this request to find the position a lodestone compass needs to point to. If found, it will
// point to the lodestone. If not, it will start spinning around. A PositionTrackingDBServerBroadcast packet
// should be sent in response to this packet.
type PositionTrackingDBClientRequest struct {
	// RequestAction is the action that should be performed upon the receiving of the packet. It is one of the
	// constants found above.
	Action  protocol.PositionTrackingDBClientRequestAction
	IDValue protocol.PositionTrackingID
}

// ID ...
func (*PositionTrackingDBClientRequest) ID() uint32 {
	return IDPositionTrackingDBClientRequest
}

func (pk *PositionTrackingDBClientRequest) Marshal(io protocol.IO) {
	pk.Action.Marshal(io)
	pk.IDValue.Marshal(io)
}
