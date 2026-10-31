// Package wire is the deterministic synthetic-golden interpreter used by v2
// conformance tests. It consumes manifest nodes directly; it is not an emitter
// and does not contain a second protocol schema.
package wire

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"reflect"

	"protocolgen/internal/manifest"
)

type OptionalValue struct {
	Present bool
	Value   any
}

type UnionValue struct {
	Tag   int64
	Value any
}

type MapEntry struct {
	Key   any
	Value any
}

type ConditionalValue struct {
	Case  string
	Value any
}

func Encode(node manifest.Node, value any) ([]byte, error) {
	var buffer bytes.Buffer
	if err := encodeNode(&buffer, node, value); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func DecodeAll(node manifest.Node, data []byte) (any, error) {
	reader := decoder{data: data}
	value, err := reader.decodeNode(node)
	if err != nil {
		return nil, err
	}
	if reader.remaining() != 0 {
		return nil, fmt.Errorf("decode left %d trailing bytes", reader.remaining())
	}
	return value, nil
}

func encodeNode(buffer *bytes.Buffer, node manifest.Node, value any) error {
	switch node.Kind {
	case manifest.KindVoid:
		return nil
	case manifest.KindPrimitive:
		if node.Primitive == nil {
			return fmt.Errorf("primitive has no shape")
		}
		return encodePrimitive(buffer, node.Primitive.Code, value)
	case manifest.KindString:
		text, ok := value.(string)
		if !ok {
			return fmt.Errorf("string node received %T", value)
		}
		data := []byte(text)
		if err := encodeLength(buffer, node.Prefix, uint64(len(data))); err != nil {
			return err
		}
		_, _ = buffer.Write(data)
		return nil
	case manifest.KindBytes:
		data, err := byteValue(value)
		if err != nil {
			return err
		}
		if err := encodeLength(buffer, node.Prefix, uint64(len(data))); err != nil {
			return err
		}
		_, _ = buffer.Write(data)
		return nil
	case manifest.KindBitset:
		return encodeBitset(buffer, node.Length, value)
	case manifest.KindArray, manifest.KindFixedArray:
		values, err := sliceValue(value)
		if err != nil {
			return err
		}
		if node.Kind == manifest.KindArray {
			if err := encodeLength(buffer, node.Prefix, uint64(len(values))); err != nil {
				return err
			}
		} else if uint64(len(values)) != node.Length {
			return fmt.Errorf("fixed array length %d does not equal %d", len(values), node.Length)
		}
		for _, item := range values {
			if node.Element == nil {
				return fmt.Errorf("array has no element")
			}
			if err := encodeNode(buffer, *node.Element, item); err != nil {
				return err
			}
		}
		return nil
	case manifest.KindSequence:
		return encodeSequence(buffer, node.Elements, value)
	case manifest.KindOptional:
		present, inner := optionalValue(value)
		if err := encodePrimitive(buffer, "bool", present); err != nil {
			return err
		}
		if !present {
			return nil
		}
		if node.Value == nil {
			return fmt.Errorf("optional has no value")
		}
		return encodeNode(buffer, *node.Value, inner)
	case manifest.KindStruct:
		object, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("struct node received %T; use map[string]any", value)
		}
		for _, field := range node.Fields {
			fieldValue, ok := object[field.Name]
			if !ok {
				return fmt.Errorf("struct is missing field %s", field.Name)
			}
			if err := encodeNode(buffer, field.Encode, fieldValue); err != nil {
				return fmt.Errorf("field %s: %w", field.Name, err)
			}
		}
		return nil
	case manifest.KindMap:
		entries, ok := value.([]MapEntry)
		if !ok {
			return fmt.Errorf("ordered map received %T; use []MapEntry", value)
		}
		if err := encodeLength(buffer, node.Prefix, uint64(len(entries))); err != nil {
			return err
		}
		for _, entry := range entries {
			if node.Key == nil || node.Value == nil {
				return fmt.Errorf("map has no key/value")
			}
			if err := encodeNode(buffer, *node.Key, entry.Key); err != nil {
				return err
			}
			if err := encodeNode(buffer, *node.Value, entry.Value); err != nil {
				return err
			}
		}
		return nil
	case manifest.KindUnion:
		choice, ok := value.(UnionValue)
		if !ok {
			return fmt.Errorf("union received %T; use UnionValue", value)
		}
		if node.Control == nil {
			return fmt.Errorf("union has no control")
		}
		if err := encodeNode(buffer, *node.Control, choice.Tag); err != nil {
			return err
		}
		for _, variant := range node.Variants {
			if variant.Value == choice.Tag {
				return encodeNode(buffer, variant.Encode, choice.Value)
			}
		}
		return fmt.Errorf("union has no variant for tag %d", choice.Tag)
	case manifest.KindConditional:
		conditional, ok := value.(ConditionalValue)
		if !ok {
			return fmt.Errorf("conditional received %T; use ConditionalValue", value)
		}
		for _, oneCase := range node.Cases {
			if oneCase.Value != conditional.Case {
				continue
			}
			return encodeSequence(buffer, oneCase.Encode, conditional.Value)
		}
		if node.Default != nil {
			return encodeNode(buffer, *node.Default, conditional.Value)
		}
		return fmt.Errorf("conditional has no case %q", conditional.Case)
	case manifest.KindEnum:
		if node.Primitive == nil {
			return fmt.Errorf("enum has no underlying primitive")
		}
		return encodePrimitive(buffer, node.Primitive.Code, value)
	case manifest.KindReserved, manifest.KindIgnored:
		if node.Element == nil {
			return fmt.Errorf("compatibility node has no element")
		}
		return encodeNode(buffer, *node.Element, value)
	case manifest.KindRecursive:
		return fmt.Errorf("recursive node %s requires a profile codec", node.Target)
	case manifest.KindOpaque, manifest.KindUnresolved:
		return fmt.Errorf("%s node blocks encoding: %s", node.Kind, node.Reason)
	default:
		return fmt.Errorf("unsupported node kind %q", node.Kind)
	}
}

func encodeSequence(buffer *bytes.Buffer, nodes []manifest.Node, value any) error {
	values, err := sequenceValues(value, len(nodes))
	if err != nil {
		return err
	}
	for index, node := range nodes {
		if err := encodeNode(buffer, node, values[index]); err != nil {
			return err
		}
	}
	return nil
}

func encodeBitset(buffer *bytes.Buffer, bits uint64, value any) error {
	values, err := sliceValue(value)
	if err != nil {
		return fmt.Errorf("bitset: %w", err)
	}
	wordCount := int((bits + 63) / 64)
	if len(values) != wordCount {
		return fmt.Errorf("%d-bit bitset has %d words, want %d", bits, len(values), wordCount)
	}
	number := new(big.Int)
	for wordIndex, raw := range values {
		word, err := integerValue(raw)
		if err != nil {
			return fmt.Errorf("bitset word %d: %w", wordIndex, err)
		}
		for bit := 0; bit < 64; bit++ {
			position := uint64(wordIndex*64 + bit)
			if position >= bits {
				if word>>bit != 0 {
					return fmt.Errorf("%d-bit bitset has set bits outside its bound", bits)
				}
				break
			}
			if word&(uint64(1)<<bit) != 0 {
				number.SetBit(number, int(position), 1)
			}
		}
	}
	if number.Sign() == 0 {
		return buffer.WriteByte(0)
	}
	mask := big.NewInt(0x7f)
	chunk := new(big.Int)
	for number.Sign() != 0 {
		chunk.And(number, mask)
		byteValue := byte(chunk.Uint64())
		number.Rsh(number, 7)
		if number.Sign() != 0 {
			byteValue |= 0x80
		}
		if err := buffer.WriteByte(byteValue); err != nil {
			return err
		}
	}
	return nil
}

type decoder struct {
	data []byte
	off  int
}

func (d *decoder) decodeBitset(bits uint64) ([]uint64, error) {
	words := make([]uint64, (bits+63)/64)
	for offset := uint64(0); offset < bits; offset += 7 {
		data, err := d.take(1)
		if err != nil {
			return nil, err
		}
		value := data[0] & 0x7f
		for bit := uint64(0); bit < 7; bit++ {
			position := offset + bit
			if value&(1<<bit) == 0 {
				continue
			}
			if position >= bits {
				return nil, fmt.Errorf("%d-bit bitset overflow", bits)
			}
			words[position/64] |= uint64(1) << (position % 64)
		}
		if data[0]&0x80 == 0 {
			return words, nil
		}
	}
	return nil, fmt.Errorf("%d-bit bitset continuation exceeds its bound", bits)
}

func (d *decoder) decodeNode(node manifest.Node) (any, error) {
	switch node.Kind {
	case manifest.KindVoid:
		return struct{}{}, nil
	case manifest.KindPrimitive:
		if node.Primitive == nil {
			return nil, fmt.Errorf("primitive has no shape")
		}
		return d.decodePrimitive(node.Primitive.Code)
	case manifest.KindString:
		length, err := d.decodeLength(node.Prefix)
		if err != nil {
			return nil, err
		}
		data, err := d.take(length)
		if err != nil {
			return nil, err
		}
		return string(data), nil
	case manifest.KindBytes:
		length, err := d.decodeLength(node.Prefix)
		if err != nil {
			return nil, err
		}
		data, err := d.take(length)
		if err != nil {
			return nil, err
		}
		return append([]byte(nil), data...), nil
	case manifest.KindBitset:
		return d.decodeBitset(node.Length)
	case manifest.KindArray, manifest.KindFixedArray:
		length := node.Length
		if node.Kind == manifest.KindArray {
			var err error
			length, err = d.decodeLength(node.Prefix)
			if err != nil {
				return nil, err
			}
		}
		if length > uint64(d.remaining())*16+1<<20 {
			return nil, fmt.Errorf("array length %d is unreasonable for remaining buffer %d", length, d.remaining())
		}
		values := make([]any, 0, length)
		for index := uint64(0); index < length; index++ {
			if node.Element == nil {
				return nil, fmt.Errorf("array has no element")
			}
			value, err := d.decodeNode(*node.Element)
			if err != nil {
				return nil, fmt.Errorf("array element %d: %w", index, err)
			}
			values = append(values, value)
		}
		return values, nil
	case manifest.KindSequence:
		values := make([]any, 0, len(node.Elements))
		for index, element := range node.Elements {
			value, err := d.decodeNode(element)
			if err != nil {
				return nil, fmt.Errorf("sequence element %d: %w", index, err)
			}
			values = append(values, value)
		}
		return values, nil
	case manifest.KindOptional:
		present, err := d.decodePrimitive("bool")
		if err != nil {
			return nil, err
		}
		isPresent, ok := present.(bool)
		if !ok {
			return nil, fmt.Errorf("presence codec returned %T", present)
		}
		if !isPresent {
			return OptionalValue{}, nil
		}
		if node.Value == nil {
			return nil, fmt.Errorf("optional has no value")
		}
		value, err := d.decodeNode(*node.Value)
		return OptionalValue{Present: true, Value: value}, err
	case manifest.KindStruct:
		object := make(map[string]any, len(node.Fields))
		for _, field := range node.Fields {
			shape := field.Encode
			if field.Decode != nil {
				shape = *field.Decode
			}
			value, err := d.decodeNode(shape)
			if err != nil {
				return nil, fmt.Errorf("field %s: %w", field.Name, err)
			}
			object[field.Name] = value
		}
		return object, nil
	case manifest.KindMap:
		length, err := d.decodeLength(node.Prefix)
		if err != nil {
			return nil, err
		}
		entries := make([]MapEntry, 0)
		for index := uint64(0); index < length; index++ {
			key, err := d.decodeNode(*node.Key)
			if err != nil {
				return nil, err
			}
			value, err := d.decodeNode(*node.Value)
			if err != nil {
				return nil, err
			}
			entries = append(entries, MapEntry{Key: key, Value: value})
		}
		return entries, nil
	case manifest.KindUnion:
		if node.Control == nil {
			return nil, fmt.Errorf("union has no control")
		}
		control, err := d.decodeNode(*node.Control)
		if err != nil {
			return nil, err
		}
		tag, err := integerValue(control)
		if err != nil {
			return nil, err
		}
		for _, variant := range node.Variants {
			if variant.Value == int64(tag) {
				shape := variant.Encode
				if variant.Decode != nil {
					shape = *variant.Decode
				}
				value, err := d.decodeNode(shape)
				return UnionValue{Tag: int64(tag), Value: value}, err
			}
		}
		return nil, fmt.Errorf("union has no variant for tag %d", tag)
	case manifest.KindConditional:
		return nil, fmt.Errorf("conditional decode needs discriminator context %q", node.CompareTo)
	case manifest.KindEnum:
		if node.Primitive == nil {
			return nil, fmt.Errorf("enum has no underlying primitive")
		}
		return d.decodePrimitive(node.Primitive.Code)
	case manifest.KindReserved, manifest.KindIgnored:
		if node.Element == nil {
			return nil, fmt.Errorf("compatibility node has no element")
		}
		return d.decodeNode(*node.Element)
	case manifest.KindRecursive:
		return nil, fmt.Errorf("recursive node %s requires a profile decoder", node.Target)
	case manifest.KindOpaque, manifest.KindUnresolved:
		return nil, fmt.Errorf("%s node blocks decoding: %s", node.Kind, node.Reason)
	default:
		return nil, fmt.Errorf("unsupported node kind %q", node.Kind)
	}
}

func (d *decoder) decodePrimitive(code string) (any, error) {
	switch code {
	case "bool":
		data, err := d.take(1)
		if err != nil {
			return nil, err
		}
		if data[0] > 1 {
			return nil, fmt.Errorf("invalid bool byte %d", data[0])
		}
		return data[0] == 1, nil
	case "i8":
		data, err := d.take(1)
		if err != nil {
			return nil, err
		}
		return int8(data[0]), nil
	case "u8":
		data, err := d.take(1)
		if err != nil {
			return nil, err
		}
		return data[0], nil
	case "i16le", "u16le", "i16be", "u16be":
		data, err := d.take(2)
		if err != nil {
			return nil, err
		}
		var order binary.ByteOrder = binary.LittleEndian
		if code[len(code)-2:] == "be" {
			order = binary.BigEndian
		}
		value := order.Uint16(data)
		if code[0] == 'i' {
			return int16(value), nil
		}
		return value, nil
	case "i32le", "u32le", "i32be", "u32be", "f32le", "f32be":
		data, err := d.take(4)
		if err != nil {
			return nil, err
		}
		var order binary.ByteOrder = binary.LittleEndian
		if code[len(code)-2:] == "be" {
			order = binary.BigEndian
		}
		value := order.Uint32(data)
		switch code[0] {
		case 'i':
			return int32(value), nil
		case 'f':
			return math.Float32frombits(value), nil
		default:
			return value, nil
		}
	case "i64le", "u64le", "i64be", "u64be", "f64le", "f64be":
		data, err := d.take(8)
		if err != nil {
			return nil, err
		}
		var order binary.ByteOrder = binary.LittleEndian
		if code[len(code)-2:] == "be" {
			order = binary.BigEndian
		}
		value := order.Uint64(data)
		switch code[0] {
		case 'i':
			return int64(value), nil
		case 'f':
			return math.Float64frombits(value), nil
		default:
			return value, nil
		}
	case "var_i32":
		value, err := d.readUvarint32()
		return int32(value), err
	case "var_u32":
		value, err := d.readUvarint32()
		return uint32(value), err
	case "var_i64":
		value, err := d.readUvarint()
		return int64(value), err
	case "zigzag_i32":
		value, err := d.readUvarint32()
		return int32((value >> 1) ^ uint64(-int64(value&1))), err
	case "zigzag_i64":
		value, err := d.readUvarint()
		return int64((value >> 1) ^ uint64(-int64(value&1))), err
	case "var_u64":
		return d.readUvarint()
	case "uuid":
		data, err := d.take(16)
		if err != nil {
			return nil, err
		}
		value := append([]byte(nil), data...)
		reverse(value[:8])
		reverse(value[8:])
		return value, nil
	case "nbt_le":
		return nil, fmt.Errorf("NBT primitive requires a bounded schema codec")
	default:
		return nil, fmt.Errorf("unsupported primitive code %q", code)
	}
}

func encodePrimitive(buffer *bytes.Buffer, code string, value any) error {
	switch code {
	case "bool":
		boolean, ok := value.(bool)
		if !ok {
			return fmt.Errorf("bool codec received %T", value)
		}
		if boolean {
			buffer.WriteByte(1)
		} else {
			buffer.WriteByte(0)
		}
		return nil
	case "i8", "u8":
		unsigned, err := integerValue(value)
		if err != nil {
			return err
		}
		buffer.WriteByte(byte(unsigned))
		return nil
	case "i16le", "u16le", "i16be", "u16be":
		unsigned, err := integerValue(value)
		if err != nil {
			return err
		}
		var writeOrder binary.ByteOrder = binary.LittleEndian
		if code[len(code)-2:] == "be" {
			writeOrder = binary.BigEndian
		}
		var data [2]byte
		writeOrder.PutUint16(data[:], uint16(unsigned))
		_, _ = buffer.Write(data[:])
		return nil
	case "i32le", "u32le", "i32be", "u32be":
		unsigned, err := integerValue(value)
		if err != nil {
			return err
		}
		var writeOrder binary.ByteOrder = binary.LittleEndian
		if code[len(code)-2:] == "be" {
			writeOrder = binary.BigEndian
		}
		var data [4]byte
		writeOrder.PutUint32(data[:], uint32(unsigned))
		_, _ = buffer.Write(data[:])
		return nil
	case "i64le", "u64le", "i64be", "u64be":
		unsigned, err := integerValue(value)
		if err != nil {
			return err
		}
		var writeOrder binary.ByteOrder = binary.LittleEndian
		if code[len(code)-2:] == "be" {
			writeOrder = binary.BigEndian
		}
		var data [8]byte
		writeOrder.PutUint64(data[:], unsigned)
		_, _ = buffer.Write(data[:])
		return nil
	case "f32le", "f32be":
		float, err := floatValue(value)
		if err != nil {
			return err
		}
		var writeOrder binary.ByteOrder = binary.LittleEndian
		if code[len(code)-2:] == "be" {
			writeOrder = binary.BigEndian
		}
		var data [4]byte
		writeOrder.PutUint32(data[:], math.Float32bits(float32(float)))
		_, _ = buffer.Write(data[:])
		return nil
	case "f64le", "f64be":
		float, err := floatValue(value)
		if err != nil {
			return err
		}
		var writeOrder binary.ByteOrder = binary.LittleEndian
		if code[len(code)-2:] == "be" {
			writeOrder = binary.BigEndian
		}
		var data [8]byte
		writeOrder.PutUint64(data[:], math.Float64bits(float))
		_, _ = buffer.Write(data[:])
		return nil
	case "var_i32":
		integer, err := integerValue(value)
		if err != nil {
			return err
		}
		writeUvarint(buffer, uint64(uint32(int32(integer))))
		return nil
	case "var_u32":
		integer, err := integerValue(value)
		if err != nil {
			return err
		}
		writeUvarint(buffer, uint64(uint32(integer)))
		return nil
	case "var_i64":
		integer, err := integerValue(value)
		if err != nil {
			return err
		}
		writeUvarint(buffer, uint64(int64(integer)))
		return nil
	case "zigzag_i32":
		integer, err := integerValue(value)
		if err != nil {
			return err
		}
		signed := int32(integer)
		encoded := uint32(uint64(signed<<1) ^ uint64(signed>>31))
		writeUvarint(buffer, uint64(encoded))
		return nil
	case "zigzag_i64":
		integer, err := integerValue(value)
		if err != nil {
			return err
		}
		signed := int64(integer)
		encoded := uint64(signed<<1) ^ uint64(signed>>63)
		writeUvarint(buffer, encoded)
		return nil
	case "var_u64":
		integer, err := integerValue(value)
		if err != nil {
			return err
		}
		writeUvarint(buffer, integer)
		return nil
	case "uuid":
		data, err := byteValue(value)
		if err != nil || len(data) != 16 {
			if err != nil {
				return err
			}
			return fmt.Errorf("uuid has length %d", len(data))
		}
		wire := append([]byte(nil), data...)
		reverse(wire[:8])
		reverse(wire[8:])
		_, _ = buffer.Write(wire)
		return nil
	case "nbt_le":
		return fmt.Errorf("NBT primitive requires a bounded schema codec")
	default:
		return fmt.Errorf("unsupported primitive code %q", code)
	}
}

func reverse(data []byte) {
	for left, right := 0, len(data)-1; left < right; left, right = left+1, right-1 {
		data[left], data[right] = data[right], data[left]
	}
}

func encodeLength(buffer *bytes.Buffer, node *manifest.Node, length uint64) error {
	if node == nil {
		return fmt.Errorf("missing length prefix")
	}
	if node.Kind != manifest.KindPrimitive || node.Primitive == nil || node.Primitive.Width == 0 {
		return fmt.Errorf("invalid length prefix")
	}
	bits := node.Primitive.Width
	if node.Primitive.Signed {
		bits--
	}
	if bits < 64 && length >= uint64(1)<<bits {
		return fmt.Errorf("length %d does not fit %s", length, node.Primitive.Code)
	}
	return encodeNode(buffer, *node, length)
}

func (d *decoder) decodeLength(node *manifest.Node) (uint64, error) {
	if node == nil {
		return 0, fmt.Errorf("missing length prefix")
	}
	value, err := d.decodeNode(*node)
	if err != nil {
		return 0, err
	}
	rv := reflect.ValueOf(value)
	if rv.IsValid() && rv.CanInt() && rv.Int() < 0 {
		return 0, fmt.Errorf("negative collection length %d", rv.Int())
	}
	return integerValue(value)
}

func (d *decoder) take(length uint64) ([]byte, error) {
	if length > uint64(d.remaining()) {
		return nil, fmt.Errorf("need %d bytes, have %d", length, d.remaining())
	}
	start := d.off
	d.off += int(length)
	return d.data[start:d.off], nil
}

func (d *decoder) remaining() int { return len(d.data) - d.off }

// readUvarint32 applies the 32-bit payload and encoded-width limits.
func (d *decoder) readUvarint32() (uint64, error) { return d.readUvarintBits(32) }

// readUvarint applies the 64-bit payload and encoded-width limits.
func (d *decoder) readUvarint() (uint64, error) { return d.readUvarintBits(64) }

// readUvarintBits rejects payload bits and continuation beyond the target width.
func (d *decoder) readUvarintBits(width uint) (uint64, error) {
	var value uint64
	for shift := uint(0); shift < width; shift += 7 {
		data, err := d.take(1)
		if err != nil {
			return 0, err
		}
		b := data[0]
		if width-shift < 7 && b>>(width-shift) != 0 {
			return 0, fmt.Errorf("varint overflow")
		}
		value |= uint64(b&0x7f) << shift
		if b&0x80 == 0 {
			return value, nil
		}
	}
	return 0, fmt.Errorf("varint too long")
}

func writeUvarint(buffer *bytes.Buffer, value uint64) {
	for value >= 0x80 {
		buffer.WriteByte(byte(value) | 0x80)
		value >>= 7
	}
	buffer.WriteByte(byte(value))
}

func integerValue(value any) (uint64, error) {
	if value == nil {
		return 0, fmt.Errorf("nil is not an integer")
	}
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return uint64(rv.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint(), nil
	default:
		return 0, fmt.Errorf("%T is not an integer", value)
	}
}

func floatValue(value any) (float64, error) {
	rv := reflect.ValueOf(value)
	if !rv.IsValid() {
		return 0, fmt.Errorf("nil is not a float")
	}
	switch rv.Kind() {
	case reflect.Float32, reflect.Float64:
		return rv.Float(), nil
	default:
		return 0, fmt.Errorf("%T is not a float", value)
	}
}

func byteValue(value any) ([]byte, error) {
	if data, ok := value.([]byte); ok {
		return data, nil
	}
	rv := reflect.ValueOf(value)
	if rv.IsValid() && (rv.Kind() == reflect.Array || rv.Kind() == reflect.Slice) && rv.Type().Elem().Kind() == reflect.Uint8 {
		data := make([]byte, rv.Len())
		for index := range data {
			data[index] = byte(rv.Index(index).Uint())
		}
		return data, nil
	}
	return nil, fmt.Errorf("%T is not bytes", value)
}

func sliceValue(value any) ([]any, error) {
	if values, ok := value.([]any); ok {
		return values, nil
	}
	rv := reflect.ValueOf(value)
	if !rv.IsValid() || (rv.Kind() != reflect.Array && rv.Kind() != reflect.Slice) {
		return nil, fmt.Errorf("%T is not an array", value)
	}
	result := make([]any, rv.Len())
	for index := range result {
		result[index] = rv.Index(index).Interface()
	}
	return result, nil
}

func optionalValue(value any) (bool, any) {
	if value == nil {
		return false, nil
	}
	if optional, ok := value.(OptionalValue); ok {
		return optional.Present, optional.Value
	}
	return true, value
}

func sequenceValues(value any, length int) ([]any, error) {
	if length == 1 {
		if values, ok := value.([]any); ok && len(values) == 1 {
			return values, nil
		}
		return []any{value}, nil
	}
	values, err := sliceValue(value)
	if err != nil || len(values) != length {
		return nil, fmt.Errorf("conditional sequence needs %d values, got %T", length, value)
	}
	return values, nil
}
