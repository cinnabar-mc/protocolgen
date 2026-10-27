// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// LevelChunk is sent by the server to provide the client with a chunk of a world data (16xYx16 blocks).
// Typically, a certain amount of chunks is sent to the client before sending it the spawn PlayStatus packet,
// so that the client spawns in a loaded world.
type LevelChunk struct {
	// Position contains the X and Z coordinates of the chunk sent. You can convert a block coordinate to a chunk
	// coordinate by right-shifting it four bits.
	ChunkPosition  protocol.ChunkPos
	DimensionID    protocol.DimensionType
	SubChunksCount uint32
	// SubChunkLimit is the maximum amount of sub-chunks a client will request when in request mode. A value of -1
	// means there is no limit.
	ClientRequestSubChunkLimit protocol.Optional[int32]
	// CacheEnabled specifies if the client blob cache should be enabled. This system is based on hashes of blobs
	// which are consistent and saved by the client in combination with that blob, so that the server does not
	// have the same chunk multiple times. If the client does not yet have a blob with the hash sent, it will send
	// a ClientCacheBlobStatus packet containing the hashes is does not have the data of.
	CacheEnabled        bool
	CacheMetadata       []protocol.SubChunkMetadata
	SerializedChunkData []byte
	IsClientBiomeUpdate bool
}

// ID ...
func (*LevelChunk) ID() uint32 {
	return IDLevelChunk
}

func (pk *LevelChunk) Marshal(io protocol.IO) {
	pk.ChunkPosition.Marshal(io)
	pk.DimensionID.Marshal(io)
	io.Varuint32(&pk.SubChunksCount)
	protocol.Maximum(io, &pk.SubChunksCount, 64)
	protocol.OptionalFunc(io, &pk.ClientRequestSubChunkLimit, func(value *int32) {
		io.Varint32(value)
		protocol.Minimum(io, value, -1)
		protocol.Maximum(io, value, 64)
	})
	io.Bool(&pk.CacheEnabled)
	protocol.SliceLimits(io, &pk.CacheMetadata, 0, 65)
	io.ByteSlice(&pk.SerializedChunkData)
	io.Bool(&pk.IsClientBiomeUpdate)
}
