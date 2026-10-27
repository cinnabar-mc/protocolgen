// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

// BlockPos is the position of a block. It is composed of three integers, and is typically written as either 3
// varint32s or a varint32, varuint32 and varint32.
type BlockPos struct {
	X int32
	Y int32
	Z int32
}

// Marshal reads or writes BlockPos using its canonical wire layout.
func (x *BlockPos) Marshal(io IO) {
	io.Varint32(&x.X)
	io.Varint32(&x.Y)
	io.Varint32(&x.Z)
}

// BlockPos is the position of a block. It is composed of three integers, and is typically written as either 3
// varint32s or a varint32, varuint32 and varint32.
type BlockPosition struct {
	X int32
	Y int32
	Z int32
}

// Marshal reads or writes BlockPosition using its canonical wire layout.
func (x *BlockPosition) Marshal(io IO) {
	io.Varint32(&x.X)
	io.Varint32(&x.Y)
	io.Varint32(&x.Z)
}
