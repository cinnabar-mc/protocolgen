// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

// ChunkPos is the position of a chunk. It is composed of two integers and is written as two varint32s.
type ChunkPos struct {
	X int32
	Z int32
}

// Marshal reads or writes ChunkPos using its canonical wire layout.
func (x *ChunkPos) Marshal(io IO) {
	io.Varint32(&x.X)
	io.Varint32(&x.Z)
}

// SubChunkPos is the position of a sub-chunk. The X and Z coordinates are the coordinates of the chunk, and
// the Y coordinate is the absolute sub-chunk index.
type SubChunkPos struct {
	SubchunkPositionX int32
	SubchunkPositionY int32
	SubchunkPositionZ int32
}

// Marshal reads or writes SubChunkPos using its canonical wire layout.
func (x *SubChunkPos) Marshal(io IO) {
	io.Int32(&x.SubchunkPositionX)
	io.Int32(&x.SubchunkPositionY)
	io.Int32(&x.SubchunkPositionZ)
}
