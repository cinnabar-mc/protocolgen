// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

type BlockCommandData struct {
	BlockPosition    BlockPos
	CommandBlockMode uint32
	RedstoneMode     bool
	IsConditional    bool
}

func (*BlockCommandData) tagCommandBlockUpdateData() uint32 { return 1 }

// Marshal reads or writes BlockCommandData using its canonical wire layout.
func (x *BlockCommandData) Marshal(io IO) {
	x.BlockPosition.Marshal(io)
	io.Varuint32(&x.CommandBlockMode)
	io.Bool(&x.RedstoneMode)
	io.Bool(&x.IsConditional)
}
