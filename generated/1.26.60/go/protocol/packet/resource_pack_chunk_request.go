// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// ResourcePackChunkRequest is sent by the client to request a chunk of data from a particular resource pack,
// that it has obtained information about in a ResourcePackDataInfo packet.
type ResourcePackChunkRequest struct {
	// ResourceName is the unique ID of the resource pack that the chunk of data is requested from.
	ResourceName string
	// Chunk is the requested chunk index of the chunk. It is a number that starts at 0 and is incremented for
	// each resource pack data chunk requested.
	Chunk int32
}

// ID ...
func (*ResourcePackChunkRequest) ID() uint32 {
	return IDResourcePackChunkRequest
}

func (pk *ResourcePackChunkRequest) Marshal(io protocol.IO) {
	io.String(&pk.ResourceName)
	io.Int32(&pk.Chunk)
	protocol.Minimum(io, &pk.Chunk, 0)
}
