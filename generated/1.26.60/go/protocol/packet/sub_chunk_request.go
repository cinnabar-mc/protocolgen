// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// SubChunkRequest requests specific sub-chunks from the server using a center point.
type SubChunkRequest struct {
	DimensionType              protocol.DimensionType
	SubChunkPositionOffsetList []protocol.SubChunkPosOffset
	CenterPos                  protocol.SubChunkPos
}

// ID ...
func (*SubChunkRequest) ID() uint32 {
	return IDSubChunkRequest
}

func (pk *SubChunkRequest) Marshal(io protocol.IO) {
	pk.DimensionType.Marshal(io)
	protocol.SliceLimits(io, &pk.SubChunkPositionOffsetList, 0, 8192)
	pk.CenterPos.Marshal(io)
}
