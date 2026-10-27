// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

type SubChunkData struct {
	SubChunkPosOffset     SubChunkPosOffset
	SubChunkRequestResult SubChunkRequestResult
	SerializedSubChunk    Optional[[]byte]
	HeightMapData         HeightmapData
	BlobID                Optional[uint64]
}

// Marshal reads or writes SubChunkData using its canonical wire layout.
func (x *SubChunkData) Marshal(io IO) {
	x.SubChunkPosOffset.Marshal(io)
	x.SubChunkRequestResult.Marshal(io)
	OptionalFunc(io, &x.SerializedSubChunk, io.ByteSlice)
	x.HeightMapData.Marshal(io)
	OptionalFunc(io, &x.BlobID, io.Uint64)
}

type SubChunkMetadata struct {
	BlobID uint64
}

// Marshal reads or writes SubChunkMetadata using its canonical wire layout.
func (x *SubChunkMetadata) Marshal(io IO) {
	io.Uint64(&x.BlobID)
}

type SubChunkPosOffset struct {
	SubchunkOffsetX int8
	SubchunkOffsetY int8
	SubchunkOffsetZ int8
}

// Marshal reads or writes SubChunkPosOffset using its canonical wire layout.
func (x *SubChunkPosOffset) Marshal(io IO) {
	io.Int8(&x.SubchunkOffsetX)
	io.Int8(&x.SubchunkOffsetY)
	io.Int8(&x.SubchunkOffsetZ)
}

type SubChunkRequestResult uint8

const (
	SubChunkRequestResultSuccess               SubChunkRequestResult = 1
	SubChunkRequestResultLevelChunkDoesntExist SubChunkRequestResult = 2
	SubChunkRequestResultWrongDimension        SubChunkRequestResult = 3
	SubChunkRequestResultPlayerDoesntExist     SubChunkRequestResult = 4
	SubChunkRequestResultIndexOutOfBounds      SubChunkRequestResult = 5
	SubChunkRequestResultSuccessAllAir         SubChunkRequestResult = 6
)

// Marshal reads or writes SubChunkRequestResult through its uint8 wire encoding.
func (x *SubChunkRequestResult) Marshal(io IO) { io.Uint8((*uint8)(x)) }
