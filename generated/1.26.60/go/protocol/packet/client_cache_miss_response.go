// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// ClientCacheMissResponse is part of the blob cache protocol. It is sent by the server in response to a
// ClientCacheBlobStatus packet and contains the blob data of all blobs that the client acknowledged not to
// have yet.
type ClientCacheMissResponse struct {
	// MissingBlobs is a list of all blobs that the client sent misses for in the ClientCacheBlobStatus. These
	// blobs hold the data of the blobs with the hashes they are matched with.
	MissingBlobs []protocol.MissingBlobData
}

// ID ...
func (*ClientCacheMissResponse) ID() uint32 {
	return IDClientCacheMissResponse
}

func (pk *ClientCacheMissResponse) Marshal(io protocol.IO) {
	protocol.SliceLimits(io, &pk.MissingBlobs, 0, 4095)
}
