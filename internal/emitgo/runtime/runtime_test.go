package protocol

import (
	"bytes"
	"image/color"
	"math"
	"strings"
	"testing"

	"github.com/google/uuid"
)

type sliceItem struct {
	value uint8
}

func (x *sliceItem) Marshal(io IO) {
	io.Uint8(&x.value)
}

func TestSliceGenericRoundTrip(t *testing.T) {
	want := []sliceItem{{value: 1}, {value: 2}, {value: 3}}
	writer := NewWriter()
	Slice(writer, &want)
	if err := writer.Err(); err != nil {
		t.Fatal(err)
	}

	reader := NewReader(writer.Data())
	got := []sliceItem(nil)
	Slice(reader, &got)
	if err := reader.Err(); err != nil {
		t.Fatal(err)
	}
	if len(got) != len(want) {
		t.Fatalf("decoded %d values, want %d", len(got), len(want))
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("value %d = %#v, want %#v", index, got[index], want[index])
		}
	}
}

func TestVarintsRoundTrip(t *testing.T) {
	writer := NewWriter()
	u32 := uint32(math.MaxUint32)
	i32 := int32(math.MinInt32)
	u64 := uint64(math.MaxUint64)
	i64 := int64(math.MinInt64)
	writer.Varuint32(&u32)
	writer.Varint32(&i32)
	writer.Varuint64(&u64)
	writer.Varint64(&i64)
	if err := writer.Err(); err != nil {
		t.Fatal(err)
	}

	reader := NewReader(writer.Data())
	var gotU32 uint32
	var gotI32 int32
	var gotU64 uint64
	var gotI64 int64
	reader.Varuint32(&gotU32)
	reader.Varint32(&gotI32)
	reader.Varuint64(&gotU64)
	reader.Varint64(&gotI64)
	if err := reader.Err(); err != nil {
		t.Fatal(err)
	}
	if gotU32 != u32 || gotI32 != i32 || gotU64 != u64 || gotI64 != i64 {
		t.Fatalf("round trip = %d, %d, %d, %d", gotU32, gotI32, gotU64, gotI64)
	}
}

func Test32BitVarintsRejectSixByteEncoding(t *testing.T) {
	reader := NewReader([]byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x00})
	var value uint32
	reader.Varuint32(&value)
	if reader.Err() == nil {
		t.Fatal("six-byte uint32 varint was accepted")
	}
}

func TestReaderRejectsSliceOverLimit(t *testing.T) {
	reader := NewReader(nil)
	if reader.SliceLength(maxSliceLength + 1) {
		t.Fatal("SliceLength accepted a collection over the default limit")
	}
	if reader.Err() == nil {
		t.Fatal("SliceLength did not record an error")
	}
}

func TestFuncSliceAcceptsRaisedLimit(t *testing.T) {
	data := append([]byte{0x88, 0x27}, make([]byte, 5000)...)
	reader := NewReaderWithLimit(data, 1<<20)
	values := []uint8(nil)
	FuncSlice(reader, &values, reader.Varuint32, reader.Uint8)
	if err := reader.Err(); err != nil {
		t.Fatal(err)
	}
	if len(values) != 5000 {
		t.Fatalf("decoded %d values, want 5000", len(values))
	}
}

func TestReaderWithoutLimitAcceptsLargeSlice(t *testing.T) {
	reader := NewReaderWithoutLimit(nil)
	if !reader.SliceLength(maxSliceLength + 1) {
		t.Fatal("unlimited reader rejected a collection")
	}
}

func TestUUIDRoundTrip(t *testing.T) {
	want := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	writer := NewWriter()
	writer.UUID(&want)
	if err := writer.Err(); err != nil {
		t.Fatal(err)
	}

	reader := NewReader(writer.Data())
	var got uuid.UUID
	reader.UUID(&got)
	if err := reader.Err(); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("UUID = %s, want %s", got, want)
	}
	if !bytes.Equal(writer.Data(), []byte{0x77, 0x66, 0x55, 0x44, 0x33, 0x22, 0x11, 0x00, 0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x99, 0x88}) {
		t.Fatalf("wire UUID = %x", writer.Data())
	}
}

func TestReaderRejectsTruncatedString(t *testing.T) {
	reader := NewReader([]byte{3, 'a'})
	var value string
	reader.String(&value)
	if reader.Err() == nil {
		t.Fatal("truncated string was accepted")
	}
}

func TestStringLimitsRejectBeforeReadingOrWritingPayload(t *testing.T) {
	reader := NewReader([]byte{5, 'h', 'e', 'l', 'l', 'o'})
	var decoded string
	reader.StringLimits(&decoded, 0, 4)
	if reader.Err() == nil || reader.Remaining() != 5 {
		t.Fatalf("reader error = %v, remaining = %d; want limit error before payload", reader.Err(), reader.Remaining())
	}

	writer := NewWriter()
	value := "hello"
	writer.StringLimits(&value, 0, 4)
	if writer.Err() == nil || len(writer.Data()) != 0 {
		t.Fatalf("writer error = %v, data = %x; want limit error before output", writer.Err(), writer.Data())
	}
}

func TestFuncSliceLimitsRejectsDeclaredCountBeforeAllocation(t *testing.T) {
	reader := NewReader([]byte{5, 1, 2, 3, 4, 5})
	var values []byte
	FuncSliceLimits(reader, &values, reader.Varuint32, 0, 4, reader.Uint8)
	if reader.Err() == nil || values != nil || reader.Remaining() != 5 {
		t.Fatalf("error = %v, values = %v, remaining = %d", reader.Err(), values, reader.Remaining())
	}
}

func TestNumericLimitsRejectOutOfRangeValues(t *testing.T) {
	writer := NewWriter()
	value := int32(65)
	Maximum(writer, &value, int32(64))
	if writer.Err() == nil {
		t.Fatal("Maximum accepted an out-of-range value")
	}
}

func TestPatternRejectsNonMatchingString(t *testing.T) {
	writer := NewWriter()
	value := "123"
	Pattern(writer, &value, "^[a-z]+$")
	if writer.Err() == nil {
		t.Fatal("Pattern accepted a non-matching string")
	}
}

func TestReaderErrorsIncludeByteOffset(t *testing.T) {
	reader := NewReader([]byte{2})
	var value string
	reader.String(&value)
	if err := reader.Err(); err == nil || !strings.Contains(err.Error(), "byte offset 1") {
		t.Fatalf("error = %v, want byte offset 1", reader.Err())
	}
}

func TestWriterResetClearsState(t *testing.T) {
	writer := NewWriter()
	value := uint8(1)
	writer.Uint8(&value)
	data := writer.Data()
	writer.InvalidValue(nil, "test")
	writer.Reset()
	if writer.Err() != nil || len(writer.Data()) != 0 {
		t.Fatalf("after Reset: err=%v data=%x", writer.Err(), writer.Data())
	}
	value = 2
	writer.Uint8(&value)
	if len(data) != 1 || data[0] != 2 {
		t.Fatalf("Data did not reflect documented buffer reuse: %x", data)
	}
}

func TestWriterUUIDDoesNotAllocate(t *testing.T) {
	value := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	writer := NewWriter()
	writer.data = make([]byte, 0, 16)
	allocs := testing.AllocsPerRun(100, func() {
		writer.Reset()
		writer.UUID(&value)
	})
	if allocs != 0 {
		t.Fatalf("UUID allocations = %v, want zero", allocs)
	}
}

func TestNBTScannersReadNestedPersistentCompound(t *testing.T) {
	data := persistentNBTFixture()
	data = append(data, 0x7f)
	reader := NewReader(data)
	var got []byte
	reader.NBT(&got, NBTPersistent)
	if err := reader.Err(); err != nil {
		t.Fatalf("persistent NBT: %v", err)
	}
	if !bytes.Equal(got, data[:len(data)-1]) || reader.Remaining() != 1 {
		t.Fatalf("persistent result=%x remaining=%d", got, reader.Remaining())
	}
	var marker byte
	reader.Uint8(&marker)
	if marker != 0x7f || reader.Err() != nil {
		t.Fatalf("marker=%x err=%v", marker, reader.Err())
	}
}

func TestNBTScannersReadNestedNetworkCompound(t *testing.T) {
	data := networkNBTFixture()
	data = append(data, 0x7f)
	reader := NewReader(data)
	var got []byte
	reader.NBT(&got, NBTNetwork)
	if err := reader.Err(); err != nil {
		t.Fatalf("network NBT: %v", err)
	}
	if !bytes.Equal(got, data[:len(data)-1]) || reader.Remaining() != 1 {
		t.Fatalf("network result=%x remaining=%d", got, reader.Remaining())
	}
	var marker byte
	reader.Uint8(&marker)
	if marker != 0x7f || reader.Err() != nil {
		t.Fatalf("marker=%x err=%v", marker, reader.Err())
	}
}

func TestNBTScannersRejectTruncation(t *testing.T) {
	for _, test := range []struct {
		name string
		data []byte
		enc  NBTEncoding
	}{
		{"persistent", persistentNBTFixture(), NBTPersistent},
		{"network", networkNBTFixture(), NBTNetwork},
	} {
		t.Run(test.name, func(t *testing.T) {
			for length := 0; length < len(test.data); length++ {
				reader := NewReader(test.data[:length])
				var value []byte
				reader.NBT(&value, test.enc)
				if reader.Err() == nil {
					t.Fatalf("prefix length %d was accepted", length)
				}
			}
		})
	}
}

func TestNetworkNBTRegressionBlockActorDataShape(t *testing.T) {
	data := networkCompoundStart()
	data = appendNetworkNamedString(data, "id", "Chest")
	data = appendNetworkNamedInt32(data, "x", 42)
	data = appendNetworkNamedCompound(data, "data", func(value []byte) []byte {
		value = appendNetworkNamedList(value, "Items", 10, [][]byte{
			networkCompoundPayload(networkNamedInt32(nil, "Slot", 2)),
		})
		return value
	})
	data = append(data, 0)
	reader := NewReader(append(data, 0xaa))
	var value []byte
	reader.NBT(&value, NBTNetwork)
	if err := reader.Err(); err != nil || reader.Remaining() != 1 {
		t.Fatalf("BlockActorData-shaped NBT err=%v remaining=%d", reader.Err(), reader.Remaining())
	}
}

func TestNBTWriterEchoesRawBytesWithoutFormatKnowledge(t *testing.T) {
	data := []byte{0x0a, 0x00, 0x0a, 0x01, 0x78, 0x00}
	for _, encoding := range []NBTEncoding{NBTNetwork, NBTPersistent} {
		writer := NewWriter()
		writer.NBT(&data, encoding)
		if err := writer.Err(); err != nil || !bytes.Equal(writer.Data(), data) {
			t.Fatalf("encoding=%v err=%v data=%x", encoding, err, writer.Data())
		}
	}
}

func appendPersistentNamed(data []byte, name string, payload []byte) []byte {
	data = append(data, payload[0])
	data = append(data, byte(len(name)), 0)
	data = append(data, name...)
	return append(data, payload[1:]...)
}

func persistentNBTFixture() []byte {
	data := []byte{10, 0, 0}
	data = appendPersistentNamed(data, "byte", []byte{1, 0x7f})
	data = appendPersistentNamed(data, "short", []byte{2, 0x34, 0x12})
	data = appendPersistentNamed(data, "int", []byte{3, 0x78, 0x56, 0x34, 0x12})
	data = appendPersistentNamed(data, "long", append([]byte{4}, []byte{0xef, 0xcd, 0xab, 0x89, 0x67, 0x45, 0x23, 0x01}...))
	data = appendPersistentNamed(data, "float", []byte{5, 0, 0, 0, 0})
	data = appendPersistentNamed(data, "double", append([]byte{6}, make([]byte, 8)...))
	data = appendPersistentNamed(data, "bytes", append([]byte{7, 2, 0, 0, 0}, []byte{1, 2}...))
	data = appendPersistentNamed(data, "string", append([]byte{8, 3, 0}, []byte("abc")...))
	list := []byte{9, 3, 2, 0, 0, 0, 1, 0, 0, 0, 0xff, 0xff, 0xff, 0xff}
	data = appendPersistentNamed(data, "list", list)
	data = appendPersistentNamed(data, "empty", []byte{9, 0, 0, 0, 0, 0})
	nested := appendPersistentNamed(nil, "value", []byte{3, 7, 0, 0, 0})
	data = appendPersistentNamed(data, "nested", append(append([]byte{10}, nested...), 0))
	data = appendPersistentNamed(data, "intarray", append([]byte{11, 2, 0, 0, 0}, []byte{1, 0, 0, 0, 2, 0, 0, 0}...))
	data = appendPersistentNamed(data, "longarray", append([]byte{12, 2, 0, 0, 0}, make([]byte, 16)...))
	return append(data, 0)
}

func networkNBTFixture() []byte {
	data := networkCompoundStart()
	data = appendNetworkNamed(data, "byte", 1, []byte{0x7f})
	data = appendNetworkNamed(data, "short", 2, []byte{0x34, 0x12})
	data = appendNetworkNamed(data, "int", 3, appendNetworkVarint(nil, 123))
	data = appendNetworkNamed(data, "long", 4, appendNetworkVarint64(nil, -456))
	data = appendNetworkNamed(data, "float", 5, []byte{0, 0, 0, 0})
	data = appendNetworkNamed(data, "double", 6, make([]byte, 8))
	data = appendNetworkNamed(data, "bytes", 7, append(appendNetworkVarint(nil, 2), []byte{1, 2}...))
	stringPayload := appendNetworkUvarint(nil, 3)
	stringPayload = append(stringPayload, "abc"...)
	data = appendNetworkNamed(data, "string", 8, stringPayload)
	list := []byte{3}
	list = appendNetworkVarint(list, 2)
	list = appendNetworkVarint(list, 1)
	list = appendNetworkVarint(list, -1)
	data = appendNetworkNamed(data, "list", 9, list)
	data = appendNetworkNamed(data, "empty", 9, []byte{0, 0})
	nested := appendNetworkNamed(nil, "value", 3, appendNetworkVarint(nil, 7))
	data = appendNetworkNamed(data, "nested", 10, append(nested, 0))
	intArray := appendNetworkVarint(nil, 2)
	intArray = appendNetworkVarint(intArray, 1)
	intArray = appendNetworkVarint(intArray, 2)
	data = appendNetworkNamed(data, "intarray", 11, intArray)
	longArray := appendNetworkVarint(nil, 2)
	longArray = append(longArray, appendNetworkVarint64(nil, 3)...)
	longArray = append(longArray, appendNetworkVarint64(nil, 4)...)
	data = appendNetworkNamed(data, "longarray", 12, longArray)
	return append(data, 0)
}

func appendNetworkUvarint(data []byte, value uint32) []byte {
	for value >= 0x80 {
		data = append(data, byte(value)|0x80)
		value >>= 7
	}
	return append(data, byte(value))
}

func appendNetworkVarint(data []byte, value int32) []byte {
	raw := uint32(value) << 1
	if value < 0 {
		raw = ^raw
	}
	return appendNetworkUvarint(data, raw)
}

func appendNetworkVarint64(data []byte, value int64) []byte {
	raw := uint64(value) << 1
	if value < 0 {
		raw = ^raw
	}
	for raw >= 0x80 {
		data = append(data, byte(raw)|0x80)
		raw >>= 7
	}
	return append(data, byte(raw))
}

func appendNetworkNamed(data []byte, name string, tag byte, payload []byte) []byte {
	data = append(data, tag)
	data = appendNetworkUvarint(data, uint32(len(name)))
	data = append(data, name...)
	return append(data, payload...)
}

func networkCompoundStart() []byte { return []byte{10, 0} }
func networkCompoundPayload(payload []byte) []byte {
	return append(payload, 0)
}
func networkNamedInt32(data []byte, name string, value int32) []byte {
	return appendNetworkNamed(data, name, 3, appendNetworkVarint(nil, value))
}
func appendNetworkNamedInt32(data []byte, name string, value int32) []byte {
	return appendNetworkNamed(data, name, 3, appendNetworkVarint(nil, value))
}
func appendNetworkNamedString(data []byte, name, value string) []byte {
	payload := appendNetworkUvarint(nil, uint32(len(value)))
	payload = append(payload, value...)
	return appendNetworkNamed(data, name, 8, payload)
}
func appendNetworkNamedCompound(data []byte, name string, fn func([]byte) []byte) []byte {
	payload := fn(nil)
	payload = append(payload, 0)
	return appendNetworkNamed(data, name, 10, payload)
}
func appendNetworkNamedList(data []byte, name string, element byte, values [][]byte) []byte {
	payload := []byte{element}
	payload = appendNetworkVarint(payload, int32(len(values)))
	for _, value := range values {
		payload = append(payload, value...)
	}
	return appendNetworkNamed(data, name, 9, payload)
}

// RGBA is one little-endian ARGB int, so blue is the low byte on the wire.
func TestRGBAIsALittleEndianARGBInt(t *testing.T) {
	writer := NewWriter()
	colour := color.RGBA{R: 0x11, G: 0x22, B: 0x33, A: 0x44}
	writer.RGBA(&colour)
	if got := writer.Data(); !bytes.Equal(got, []byte{0x33, 0x22, 0x11, 0x44}) {
		t.Fatalf("encoded %x", got)
	}
	var decoded color.RGBA
	NewReader(writer.Data()).RGBA(&decoded)
	if decoded != colour {
		t.Fatalf("decoded %#v", decoded)
	}
}

// TestFuncSliceBoundsAllocation checks truncated counts without reserving their declared size.
func TestFuncSliceBoundsAllocation(t *testing.T) {
	for _, count := range []uint32{1 << 20, uint32(min(uint64(math.MaxUint32), uint64(math.MaxInt)))} {
		for _, payload := range [][]byte{nil, {7}} {
			writer := NewWriter()
			writer.Varuint32(&count)
			data := append(writer.Data(), payload...)
			reader := NewReaderWithoutLimit(data)
			var values []uint64
			calls := 0
			FuncSlice(reader, &values, reader.Varuint32, func(value *uint64) {
				calls++
				if cap(values) > len(payload)+1 {
					t.Fatalf("reserved %d elements from %d payload bytes", cap(values), len(payload))
				}
				reader.Uint64(value)
			})
			if reader.Err() == nil || calls != 1 || len(values) != 0 {
				t.Fatalf("error = %v, calls = %d, values = %v", reader.Err(), calls, values)
			}
		}
	}
}

func TestFuncSliceAllowsElementsWithoutWireBytes(t *testing.T) {
	reader := NewReaderWithoutLimit([]byte{5})
	var values []struct{}
	FuncSlice(reader, &values, reader.Varuint32, func(*struct{}) {})
	if reader.Err() != nil || len(values) != 5 {
		t.Fatalf("error = %v, length = %d", reader.Err(), len(values))
	}
}
