// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

import (
	"encoding/binary"
	"fmt"
	"image/color"
	"math"
	"math/bits"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
)

// Writer encodes one generated protocol value into an in-memory byte slice.
// InvalidValue records a terminal error; callers should inspect Err after the
// generated Marshal method returns.
type Writer struct {
	data []byte
	err  error
}

// NewWriter creates an empty Writer.
func NewWriter() *Writer { return &Writer{} }

// Reading reports that this IO implementation encodes values.
func (*Writer) Reading() bool { return false }

// Err returns the first encoding error, if any.
func (w *Writer) Err() error { return w.err }

// Data returns the encoded bytes. It aliases the Writer buffer and is invalid
// after a subsequent write or Reset; copy it when the bytes must be retained.
func (w *Writer) Data() []byte { return w.data }

// Reset clears the encoded data and error while retaining the buffer capacity.
func (w *Writer) Reset() {
	w.data = w.data[:0]
	w.err = nil
}

func (w *Writer) fail(err error) {
	if w.err == nil {
		w.err = err
	}
}

// InvalidValue records a malformed model or an unsupported wire value.
func (w *Writer) InvalidValue(value any, context string) {
	w.fail(fmt.Errorf("invalid value %v for %s", value, context))
}

func (w *Writer) writeByte(value byte) {
	if w.err == nil {
		w.data = append(w.data, value)
	}
}

func (w *Writer) write(value []byte) {
	if w.err == nil {
		w.data = append(w.data, value...)
	}
}

func (w *Writer) writeUvarint(value uint64) {
	if w.err != nil {
		return
	}
	for value >= 0x80 {
		w.data = append(w.data, byte(value)|0x80)
		value >>= 7
	}
	w.data = append(w.data, byte(value))
}

func (w *Writer) Bool(x *bool) {
	if *x {
		w.writeByte(1)
	} else {
		w.writeByte(0)
	}
}

func (w *Writer) Int8(x *int8)   { w.writeByte(byte(*x)) }
func (w *Writer) Uint8(x *uint8) { w.writeByte(*x) }

func (w *Writer) Int16(x *int16) {
	var data [2]byte
	binary.LittleEndian.PutUint16(data[:], uint16(*x))
	w.write(data[:])
}

func (w *Writer) Uint16(x *uint16) {
	var data [2]byte
	binary.LittleEndian.PutUint16(data[:], *x)
	w.write(data[:])
}

func (w *Writer) BEInt16(x *int16) {
	var data [2]byte
	binary.BigEndian.PutUint16(data[:], uint16(*x))
	w.write(data[:])
}

func (w *Writer) BEUint16(x *uint16) {
	var data [2]byte
	binary.BigEndian.PutUint16(data[:], *x)
	w.write(data[:])
}

func (w *Writer) Int32(x *int32) {
	var data [4]byte
	binary.LittleEndian.PutUint32(data[:], uint32(*x))
	w.write(data[:])
}

func (w *Writer) Uint32(x *uint32) {
	var data [4]byte
	binary.LittleEndian.PutUint32(data[:], *x)
	w.write(data[:])
}

func (w *Writer) BEInt32(x *int32) {
	var data [4]byte
	binary.BigEndian.PutUint32(data[:], uint32(*x))
	w.write(data[:])
}

func (w *Writer) BEUint32(x *uint32) {
	var data [4]byte
	binary.BigEndian.PutUint32(data[:], *x)
	w.write(data[:])
}

func (w *Writer) Int64(x *int64) {
	var data [8]byte
	binary.LittleEndian.PutUint64(data[:], uint64(*x))
	w.write(data[:])
}

func (w *Writer) Uint64(x *uint64) {
	var data [8]byte
	binary.LittleEndian.PutUint64(data[:], *x)
	w.write(data[:])
}

func (w *Writer) BEInt64(x *int64) {
	var data [8]byte
	binary.BigEndian.PutUint64(data[:], uint64(*x))
	w.write(data[:])
}

func (w *Writer) BEUint64(x *uint64) {
	var data [8]byte
	binary.BigEndian.PutUint64(data[:], *x)
	w.write(data[:])
}

func (w *Writer) Float32(x *float32) {
	var data [4]byte
	binary.LittleEndian.PutUint32(data[:], math.Float32bits(*x))
	w.write(data[:])
}

func (w *Writer) Float64(x *float64) {
	var data [8]byte
	binary.LittleEndian.PutUint64(data[:], math.Float64bits(*x))
	w.write(data[:])
}

func (w *Writer) BEFloat32(x *float32) {
	var data [4]byte
	binary.BigEndian.PutUint32(data[:], math.Float32bits(*x))
	w.write(data[:])
}

func (w *Writer) BEFloat64(x *float64) {
	var data [8]byte
	binary.BigEndian.PutUint64(data[:], math.Float64bits(*x))
	w.write(data[:])
}

func (w *Writer) Varint32(x *int32) {
	value := uint32(*x)<<1 ^ uint32(*x>>31)
	w.writeUvarint(uint64(value))
}

func (w *Writer) Varuint32(x *uint32) { w.writeUvarint(uint64(*x)) }

func (w *Writer) Varint64(x *int64) {
	value := uint64(*x<<1) ^ uint64(*x>>63)
	w.writeUvarint(value)
}

func (w *Writer) Varuint64(x *uint64)     { w.writeUvarint(*x) }
func (w *Writer) SignedVarint32(x *int32) { w.writeUvarint(uint64(uint32(*x))) }
func (w *Writer) SignedVarint64(x *int64) { w.writeUvarint(uint64(*x)) }

func (w *Writer) ActorRuntimeID(x *uint64)          { w.Varuint64(x) }
func (w *Writer) ActorRuntimeIDVarint64(x *int64)   { w.Varint64(x) }
func (w *Writer) ActorRuntimeIDVaruint32(x *uint32) { w.Varuint32(x) }
func (w *Writer) ActorUniqueID(x *int64)            { w.Varint64(x) }
func (w *Writer) ActorUniqueIDInt64(x *int64)       { w.Int64(x) }
func (w *Writer) ActorUniqueIDUint64(x *uint64)     { w.Uint64(x) }
func (w *Writer) ActorUniqueIDVaruint64(x *uint64)  { w.Varuint64(x) }
func (w *Writer) PlayerInputTick(x *uint64)         { w.Varuint64(x) }

func (w *Writer) String(x *string) {
	w.StringLimits(x, 0, ^uint64(0))
}

func (w *Writer) StringLimits(x *string, min, max uint64) {
	data := []byte(*x)
	if uint64(len(data)) < min || uint64(len(data)) > max {
		w.InvalidValue(len(data), "string length outside schema limits")
		return
	}
	if uint64(len(data)) > uint64(^uint32(0)) {
		w.InvalidValue(len(data), "string exceeds uint32 length")
		return
	}
	length := uint32(len(data))
	w.Varuint32(&length)
	w.write(data)
}

func (w *Writer) ByteSlice(x *[]byte) {
	w.ByteSliceLimits(x, 0, ^uint64(0))
}

func (w *Writer) ByteSliceLimits(x *[]byte, min, max uint64) {
	if uint64(len(*x)) < min || uint64(len(*x)) > max {
		w.InvalidValue(len(*x), "byte slice length outside schema limits")
		return
	}
	if uint64(len(*x)) > uint64(^uint32(0)) {
		w.InvalidValue(len(*x), "byte slice exceeds uint32 length")
		return
	}
	length := uint32(len(*x))
	w.Varuint32(&length)
	w.write(*x)
}

// NBT copies already encoded bytes. The encoding is needed by readers only;
// writers do not reinterpret raw NBT values.
func (w *Writer) NBT(x *[]byte, _ NBTEncoding) {
	if w.err != nil {
		return
	}
	w.write(*x)
}

func (w *Writer) UUID(x *uuid.UUID) {
	var data [16]byte
	copy(data[:], x[:])
	reverseBytes(data[:8])
	reverseBytes(data[8:])
	w.write(data[:])
}

func (w *Writer) UUIDBytes(x *[16]byte) {
	var value uuid.UUID
	copy(value[:], x[:])
	w.UUID(&value)
}

func (w *Writer) Vec2(x *mgl32.Vec2) {
	w.Float32(&(*x)[0])
	w.Float32(&(*x)[1])
}

func (w *Writer) Vec3(x *mgl32.Vec3) {
	w.Float32(&(*x)[0])
	w.Float32(&(*x)[1])
	w.Float32(&(*x)[2])
}

func (w *Writer) RGBA(x *color.RGBA) {
	value := uint32(x.B) | uint32(x.G)<<8 | uint32(x.R)<<16 | uint32(x.A)<<24
	w.Uint32(&value)
}

func (w *Writer) Bitset(words []uint64, width uint64) {
	wordCount := width / 64
	if width%64 != 0 {
		wordCount++
	}
	if uint64(len(words)) != wordCount {
		w.InvalidValue(len(words), "bitset word count does not match declared width")
		return
	}
	if width%64 != 0 && words[len(words)-1]>>(width%64) != 0 {
		w.InvalidValue(words[len(words)-1], "bitset contains bits outside its declared width")
		return
	}
	last := -1
	for index := len(words) - 1; index >= 0 && last < 0; index-- {
		if words[index] != 0 {
			last = index*64 + bits.Len64(words[index]) - 1
		}
	}
	if last < 0 {
		w.writeByte(0)
		return
	}
	groups := uint64(last)/7 + 1
	for group := uint64(0); group < groups; group++ {
		offset := group * 7
		remaining := min(width-offset, 7)
		var value byte
		for bit := uint64(0); bit < remaining; bit++ {
			index := offset + bit
			if words[index/64]&(uint64(1)<<(index%64)) != 0 {
				value |= 1 << bit
			}
		}
		if group+1 < groups {
			value |= 0x80
		}
		w.writeByte(value)
	}
}
