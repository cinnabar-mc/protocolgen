// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

import (
	"math"
	"regexp"
	"sync"
)

// Marshaler is implemented by every generated struct, enum, and union payload.
type Marshaler interface {
	Marshal(IO)
}

// PtrMarshaler constrains a value whose pointer implements Marshaler.
type PtrMarshaler[T any] interface {
	Marshaler
	*T
}

type integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type number interface {
	integer | ~float32 | ~float64
}

var schemaPatterns sync.Map

// Pattern validates a schema-published regular expression. Manifests validate
// patterns before generation; the cache avoids recompiling them per packet.
func Pattern(io IO, x *string, pattern string) {
	compiledValue, ok := schemaPatterns.Load(pattern)
	if !ok {
		compiled, err := regexp.Compile(pattern)
		if err != nil {
			io.InvalidValue(pattern, "invalid schema pattern")
			return
		}
		compiledValue, _ = schemaPatterns.LoadOrStore(pattern, compiled)
	}
	if !compiledValue.(*regexp.Regexp).MatchString(*x) {
		io.InvalidValue(*x, "string does not match schema pattern")
	}
}

// Minimum and Maximum validate schema-published numeric bounds after the wire
// operation, so the same generated call checks both encoded and decoded data.
func Minimum[T number](io IO, x *T, minimum T) {
	if *x < minimum {
		io.InvalidValue(*x, "value is below schema minimum")
	}
}

func Maximum[T number](io IO, x *T, maximum T) {
	if *x > maximum {
		io.InvalidValue(*x, "value exceeds schema maximum")
	}
}

// OptionalFunc marshals a bool-prefixed optional value.
func OptionalFunc[T any](io IO, x *Optional[T], f func(*T)) {
	io.Bool(&x.set)
	if x.set {
		f(&x.val)
	}
}

// DoubleOptionalFunc marshals an optional nested inside an always-present
// outer optional, as used by several Cereal fields.
func DoubleOptionalFunc[T any](io IO, x *Optional[T], f func(*T)) {
	outer := true
	io.Bool(&outer)
	if outer {
		OptionalFunc(io, x, f)
	} else {
		*x = Optional[T]{}
	}
}

// OptionalMarshaler marshals an Optional whose value pointer marshals itself.
func OptionalMarshaler[T any, A PtrMarshaler[T]](io IO, x *Optional[T]) {
	OptionalFunc(io, x, func(value *T) { A(value).Marshal(io) })
}

// Union marshals a tag-discriminated payload. The tag encoder fixes the wire
// width; tagOf reports the tag of the payload being written and variant
// allocates the payload for a decoded tag, returning nil for unknown tags.
func Union[T Marshaler, C integer](io IO, x *T, tag func(*C), tagOf func(T) C, variant func(C) T) {
	var value C
	if io.Reading() {
		tag(&value)
		payload := variant(value)
		if any(payload) == nil {
			io.InvalidValue(value, "unknown union tag")
			return
		}
		*x = payload
	} else {
		if any(*x) == nil {
			io.InvalidValue(*x, "missing union value")
			return
		}
		value = tagOf(*x)
		tag(&value)
	}
	(*x).Marshal(io)
}

// sliceReader is implemented by decoders that allocate slices from a wire
// count. Writers intentionally do not implement it, so the same helpers work
// in both directions without exposing direction checks in packet methods.
type sliceReader interface {
	SliceLength(length uint64) bool
}

// boundedSliceReader exposes the input budget and terminal decode error so
// collections reserve only available input and stop at the first failed element.
type boundedSliceReader interface {
	Remaining() int
	Err() error
}

const maxSliceLength = 4096

// Slice marshals a varuint32-prefixed slice whose element pointer marshals itself.
func Slice[T any, A PtrMarshaler[T]](io IO, x *[]T) {
	SliceLimits[T, A](io, x, 0, math.MaxUint64)
}

// SliceLimits is Slice with schema-published element-count bounds.
func SliceLimits[T any, A PtrMarshaler[T]](io IO, x *[]T, min, max uint64) {
	FuncSliceLimits(io, x, io.Varuint32, min, max, func(value *T) { A(value).Marshal(io) })
}

// FuncSlice marshals a length-prefixed slice using a count encoder and element
// callback. The count encoder determines the exact wire prefix type.
func FuncSlice[T any, C integer](io IO, x *[]T, count func(*C), f func(*T)) {
	FuncSliceLimits(io, x, count, 0, math.MaxUint64, f)
}

// FuncSliceLimits is FuncSlice with schema-published element-count bounds.
func FuncSliceLimits[T any, C integer](io IO, x *[]T, count func(*C), minimum, maximum uint64, f func(*T)) {
	length, ok := collectionLength(io, count, len(*x), minimum, maximum)
	if !ok {
		return
	}
	if io.Reading() {
		if reader, ok := io.(boundedSliceReader); ok {
			*x = make([]T, min(length, reader.Remaining()))
			var zero T
			for i := 0; i < length && reader.Err() == nil; i++ {
				if i == len(*x) {
					*x = append(*x, zero)
				}
				f(&(*x)[i])
				if reader.Err() != nil {
					*x = (*x)[:i]
					return
				}
			}
			return
		}
		*x = make([]T, length)
	}
	for i := range *x {
		f(&(*x)[i])
	}
}

// OrderedMap marshals an ordered map representation while preserving duplicate
// keys and source order.
func OrderedMap[K, V any, C integer](io IO, x *[]OrderedEntry[K, V], count func(*C), key func(*K), value func(*V)) {
	OrderedMapLimits(io, x, count, 0, math.MaxUint64, key, value)
}

// OrderedMapLimits is OrderedMap with schema-published entry-count bounds.
func OrderedMapLimits[K, V any, C integer](io IO, x *[]OrderedEntry[K, V], count func(*C), min, max uint64, key func(*K), value func(*V)) {
	FuncSliceLimits(io, x, count, min, max, func(entry *OrderedEntry[K, V]) {
		key(&entry.Key)
		value(&entry.Value)
	})
}

// collectionLength marshals a collection count through its wire prefix and
// returns the element count to iterate. Readers validate the decoded count
// against the schema bounds and the decoder limit before any allocation;
// writers validate the live length against the bounds and the prefix width.
func collectionLength[C integer](io IO, count func(*C), length int, min, max uint64) (int, bool) {
	if !io.Reading() {
		if !schemaLength(io, uint64(length), min, max) {
			return 0, false
		}
		n := C(length)
		if int(n) != length || n < 0 {
			io.InvalidValue(length, "collection length exceeds its wire count")
			return 0, false
		}
		count(&n)
		return length, true
	}
	reader, ok := io.(sliceReader)
	if !ok {
		io.InvalidValue(io, "reader does not implement SliceLength")
		return 0, false
	}
	var n C
	count(&n)
	if n < 0 {
		io.InvalidValue(n, "negative collection length")
		return 0, false
	}
	if uint64(n) > math.MaxInt {
		io.InvalidValue(n, "collection length overflows int")
		return 0, false
	}
	if !schemaLength(io, uint64(n), min, max) || !reader.SliceLength(uint64(n)) {
		return 0, false
	}
	return int(n), true
}

func schemaLength(io IO, value, min, max uint64) bool {
	if value < min || value > max {
		io.InvalidValue(value, "collection length outside schema limits")
		return false
	}
	return true
}
