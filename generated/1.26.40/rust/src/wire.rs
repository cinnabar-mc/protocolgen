// Code generated from canonical protocol manifest v2. DO NOT EDIT.

//! Wire runtime for the generated Bedrock protocol types.
//!
//! Decoding is slice-based rather than `std::io::Read` so that every
//! length-prefixed field can be bounded against the bytes actually remaining
//! before anything is allocated.

use std::fmt;

/// Default bound on the element count of a length-prefixed collection.
///
/// It is a default rather than a hard cap because a peer may legitimately send
/// a larger collection; `Reader::set_collection_limit` raises or lowers it per
/// decode, matching the Go backend's `NewReaderWithLimit`.
pub const MAX_COLLECTION_ELEMENTS: usize = 4096;

/// Default bound on a length-prefixed byte buffer or string.
pub const MAX_BYTE_BUFFER_LEN: usize = 16 * 1024 * 1024;

/// Largest compound/list nesting depth accepted by the NBT scanners.
pub const MAX_NBT_DEPTH: u32 = 512;

#[derive(Clone, Debug, PartialEq, Eq)]
pub enum DecodeError {
    UnexpectedEof {
        needed: usize,
        remaining: usize,
    },
    /// A varint used more bytes than its canonical encoding requires.
    VarIntOverlong,
    /// A varint did not terminate within the width of its target type.
    VarIntTooLarge,
    InvalidUtf8,
    NegativeLength(i64),
    /// A declared element count exceeded the configured bound.
    LengthLimitExceeded {
        limit: usize,
        actual: usize,
    },
    /// A declared element count could not be covered by the remaining bytes.
    LengthNotRepresentable {
        needed: usize,
        remaining: usize,
    },
    SchemaConstraint(&'static str),
    UnknownVariant {
        type_name: &'static str,
        value: i64,
    },
    NbtDepthExceeded,
    /// A recursive schema edge exceeded the client-owned decoded-tree depth budget.
    RecursiveDepthExceeded,
    NbtMalformed(&'static str),
    UnknownPacketId(u32),
    /// A known packet id that the sending peer is not permitted to use.
    UnexpectedDirection(u32),
    /// Bytes remained inside a region that declared its own length.
    TrailingBytes(usize),
}

impl fmt::Display for DecodeError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::UnexpectedEof { needed, remaining } => {
                write!(
                    f,
                    "unexpected end of input: needed {needed}, {remaining} remaining"
                )
            }
            Self::VarIntOverlong => write!(f, "varint is not canonically encoded"),
            Self::VarIntTooLarge => write!(f, "varint exceeds the width of its target type"),
            Self::InvalidUtf8 => write!(f, "string is not valid UTF-8"),
            Self::NegativeLength(value) => write!(f, "negative length {value}"),
            Self::LengthLimitExceeded { limit, actual } => {
                write!(f, "declared length {actual} exceeds limit {limit}")
            }
            Self::LengthNotRepresentable { needed, remaining } => {
                write!(
                    f,
                    "declared length needs {needed} bytes, {remaining} remaining"
                )
            }
            Self::SchemaConstraint(reason) => write!(f, "schema constraint violated: {reason}"),
            Self::UnknownVariant { type_name, value } => {
                write!(f, "unknown {type_name} discriminant {value}")
            }
            Self::NbtDepthExceeded => write!(f, "NBT nesting exceeds {MAX_NBT_DEPTH}"),
            Self::RecursiveDepthExceeded => {
                write!(
                    f,
                    "recursive value depth exceeds {MAX_RECURSIVE_VALUE_DEPTH}"
                )
            }
            Self::NbtMalformed(reason) => write!(f, "malformed NBT: {reason}"),
            Self::UnknownPacketId(id) => write!(f, "unknown packet id {id}"),
            Self::UnexpectedDirection(id) => {
                write!(f, "packet id {id} cannot be sent by this peer")
            }
            Self::TrailingBytes(count) => write!(f, "{count} unread bytes in declared region"),
        }
    }
}

impl std::error::Error for DecodeError {}

pub type DecodeResult<T> = Result<T, DecodeError>;

/// Cursor over a byte slice. Every read is bounds-checked; no read allocates.
///
/// Constructing with [`Reader::from_shared`] lets byte-buffer and NBT fields
/// decode as refcounted slices of the original buffer instead of copies.
#[derive(Clone, Debug)]
pub struct Reader<'a> {
    buf: &'a [u8],
    pos: usize,
    shared: Option<&'a bytes::Bytes>,
    collection_limit: usize,
    byte_buffer_limit: usize,
    recursive_depth: usize,
}

impl<'a> Reader<'a> {
    pub fn new(buf: &'a [u8]) -> Self {
        Self {
            buf,
            pos: 0,
            shared: None,
            collection_limit: MAX_COLLECTION_ELEMENTS,
            byte_buffer_limit: MAX_BYTE_BUFFER_LEN,
            recursive_depth: 0,
        }
    }

    /// Reads from `source` and shares its allocation with decoded byte buffers.
    pub fn from_shared(source: &'a bytes::Bytes) -> Self {
        Self {
            shared: Some(source),
            ..Self::new(source)
        }
    }

    pub fn collection_limit(&self) -> usize {
        self.collection_limit
    }

    /// Raises or lowers the element bound applied to every length-prefixed
    /// collection decoded through this reader.
    pub fn set_collection_limit(&mut self, limit: usize) -> &mut Self {
        self.collection_limit = limit;
        self
    }

    pub fn byte_buffer_limit(&self) -> usize {
        self.byte_buffer_limit
    }

    pub fn set_byte_buffer_limit(&mut self, limit: usize) -> &mut Self {
        self.byte_buffer_limit = limit;
        self
    }

    /// Refcounts `slice` against the source buffer when this reader shares one,
    /// and copies otherwise.
    ///
    /// Private because `Bytes::slice_ref` panics on a slice from a different
    /// allocation. Callers use `take_shared`, which can only pass its own.
    fn share(&self, slice: &[u8]) -> bytes::Bytes {
        match self.shared {
            Some(source) => source.slice_ref(slice),
            None => bytes::Bytes::copy_from_slice(slice),
        }
    }

    /// Reads `count` bytes without copying when this reader shares its source.
    pub fn take_shared(&mut self, count: usize) -> DecodeResult<bytes::Bytes> {
        let slice = self.take(count)?;
        Ok(self.share(slice))
    }

    pub fn remaining(&self) -> usize {
        self.buf.len() - self.pos
    }

    pub fn position(&self) -> usize {
        self.pos
    }

    pub fn is_empty(&self) -> bool {
        self.remaining() == 0
    }

    /// Errors unless the input is fully consumed. Callers that decode a region
    /// with a declared length use this to reject trailing bytes.
    pub fn expect_consumed(&self) -> DecodeResult<()> {
        match self.remaining() {
            0 => Ok(()),
            count => Err(DecodeError::TrailingBytes(count)),
        }
    }

    pub fn take(&mut self, count: usize) -> DecodeResult<&'a [u8]> {
        if count > self.remaining() {
            return Err(DecodeError::UnexpectedEof {
                needed: count,
                remaining: self.remaining(),
            });
        }
        let start = self.pos;
        self.pos += count;
        Ok(&self.buf[start..self.pos])
    }

    pub fn read_u8(&mut self) -> DecodeResult<u8> {
        Ok(self.take(1)?[0])
    }

    pub fn read_bytes<const N: usize>(&mut self) -> DecodeResult<[u8; N]> {
        let slice = self.take(N)?;
        let mut out = [0u8; N];
        out.copy_from_slice(slice);
        Ok(out)
    }

    /// Reads an unsigned LEB128 varint, rejecting both non-canonical encodings
    /// and encodings wider than `max_bits`.
    pub fn read_var_uint(&mut self, max_bits: u32) -> DecodeResult<u64> {
        let mut value = 0u64;
        let mut shift = 0u32;
        loop {
            let byte = self.read_u8()?;
            let payload = u64::from(byte & 0x7f);
            let remaining_bits = max_bits - shift;
            if remaining_bits < 7 && payload >= (1u64 << remaining_bits) {
                return Err(DecodeError::VarIntTooLarge);
            }
            value |= payload << shift;
            if byte & 0x80 == 0 {
                // A canonical varint never ends in a zero continuation group.
                if shift > 0 && byte == 0 {
                    return Err(DecodeError::VarIntOverlong);
                }
                return Ok(value);
            }
            shift += 7;
            if shift >= max_bits {
                return Err(DecodeError::VarIntTooLarge);
            }
        }
    }

    pub fn read_var_u32(&mut self) -> DecodeResult<u32> {
        Ok(self.read_var_uint(32)? as u32)
    }

    pub fn read_var_u64(&mut self) -> DecodeResult<u64> {
        self.read_var_uint(64)
    }

    pub fn read_zigzag_i32(&mut self) -> DecodeResult<i32> {
        let raw = self.read_var_uint(32)? as u32;
        Ok(((raw >> 1) as i32) ^ -((raw & 1) as i32))
    }

    pub fn read_zigzag_i64(&mut self) -> DecodeResult<i64> {
        let raw = self.read_var_uint(64)?;
        Ok(((raw >> 1) as i64) ^ -((raw & 1) as i64))
    }

    /// Validates a declared element count before it is used to reserve memory.
    ///
    /// `min_element_size` is the smallest number of bytes one element can
    /// occupy, so a count that could not possibly be covered by the remaining
    /// input is rejected before allocation rather than after a partial read.
    pub fn checked_count(&self, declared: u64, min_element_size: usize) -> DecodeResult<usize> {
        self.checked_count_with(declared, min_element_size, self.collection_limit)
    }

    /// Applies schema-published count bounds before allocation in addition to
    /// the reader's configurable global limit.
    pub fn checked_count_limits(
        &self,
        declared: u64,
        min_element_size: usize,
        minimum: u64,
        maximum: u64,
    ) -> DecodeResult<usize> {
        validate_length(declared, minimum, maximum)?;
        self.checked_count(declared, min_element_size)
    }

    /// Applies an explicit bound instead of this reader's configured one.
    pub fn checked_count_with(
        &self,
        declared: u64,
        min_element_size: usize,
        limit: usize,
    ) -> DecodeResult<usize> {
        if declared > limit as u64 {
            return Err(DecodeError::LengthLimitExceeded {
                limit,
                actual: declared as usize,
            });
        }
        let count = declared as usize;
        let needed = count.saturating_mul(min_element_size);
        if needed > self.remaining() {
            return Err(DecodeError::LengthNotRepresentable {
                needed,
                remaining: self.remaining(),
            });
        }
        Ok(count)
    }

    /// Borrows a length-prefixed UTF-8 string without copying it.
    pub fn read_str(&mut self) -> DecodeResult<&'a str> {
        let declared = self.read_var_u32()? as u64;
        let count = self.checked_count_with(declared, 1, self.byte_buffer_limit)?;
        let bytes = self.take(count)?;
        std::str::from_utf8(bytes).map_err(|_| DecodeError::InvalidUtf8)
    }

    /// Borrows a length-prefixed byte buffer without copying it.
    pub fn read_byte_slice(&mut self) -> DecodeResult<&'a [u8]> {
        let declared = self.read_var_u32()? as u64;
        let count = self.checked_count_with(declared, 1, self.byte_buffer_limit)?;
        self.take(count)
    }
}

/// Growable output buffer. Encoding is infallible.
#[derive(Clone, Debug, Default)]
pub struct Writer {
    buf: Vec<u8>,
}

impl Writer {
    pub fn new() -> Self {
        Self { buf: Vec::new() }
    }

    pub fn with_capacity(capacity: usize) -> Self {
        Self {
            buf: Vec::with_capacity(capacity),
        }
    }

    pub fn into_inner(self) -> Vec<u8> {
        self.buf
    }

    pub fn as_slice(&self) -> &[u8] {
        &self.buf
    }

    pub fn len(&self) -> usize {
        self.buf.len()
    }

    pub fn is_empty(&self) -> bool {
        self.buf.is_empty()
    }

    pub fn write_u8(&mut self, value: u8) {
        self.buf.push(value);
    }

    pub fn write_all(&mut self, bytes: &[u8]) {
        self.buf.extend_from_slice(bytes);
    }

    pub fn write_var_uint(&mut self, mut value: u64) {
        while value >= 0x80 {
            self.buf.push((value as u8) | 0x80);
            value >>= 7;
        }
        self.buf.push(value as u8);
    }

    pub fn write_var_u32(&mut self, value: u32) {
        self.write_var_uint(u64::from(value));
    }

    pub fn write_var_u64(&mut self, value: u64) {
        self.write_var_uint(value);
    }

    pub fn write_zigzag_i32(&mut self, value: i32) {
        self.write_var_uint(u64::from(((value << 1) ^ (value >> 31)) as u32));
    }

    pub fn write_zigzag_i64(&mut self, value: i64) {
        self.write_var_uint(((value << 1) ^ (value >> 63)) as u64);
    }

    pub fn write_str(&mut self, value: &str) {
        self.write_var_u32(value.len() as u32);
        self.write_all(value.as_bytes());
    }

    pub fn write_byte_slice(&mut self, value: &[u8]) {
        self.write_var_u32(value.len() as u32);
        self.write_all(value);
    }
}

pub trait Encode {
    fn encode(&self, writer: &mut Writer);

    fn encode_to_vec(&self) -> Vec<u8> {
        let mut writer = Writer::new();
        self.encode(&mut writer);
        writer.into_inner()
    }
}

pub trait Decode: Sized {
    fn decode(reader: &mut Reader<'_>) -> DecodeResult<Self>;

    /// Decodes from a complete slice and rejects trailing bytes.
    fn decode_exact(input: &[u8]) -> DecodeResult<Self> {
        let mut reader = Reader::new(input);
        let value = Self::decode(&mut reader)?;
        reader.expect_consumed()?;
        Ok(value)
    }
}

macro_rules! fixed_codec {
    ($name:ident, $inner:ty, $size:expr, $write:ident, $read:ident) => {
        #[derive(Clone, Copy, Debug, Default, PartialEq, Eq, Hash, PartialOrd, Ord)]
        pub struct $name(pub $inner);

        impl Encode for $name {
            fn encode(&self, writer: &mut Writer) {
                writer.write_all(&self.0.$write());
            }
        }

        impl Decode for $name {
            fn decode(reader: &mut Reader<'_>) -> DecodeResult<Self> {
                Ok(Self(<$inner>::$read(reader.read_bytes::<$size>()?)))
            }
        }
    };
}

macro_rules! fixed_float_codec {
    ($name:ident, $inner:ty, $size:expr, $write:ident, $read:ident) => {
        #[derive(Clone, Copy, Debug, Default, PartialEq, PartialOrd)]
        pub struct $name(pub $inner);

        impl Encode for $name {
            fn encode(&self, writer: &mut Writer) {
                writer.write_all(&self.0.$write());
            }
        }

        impl Decode for $name {
            fn decode(reader: &mut Reader<'_>) -> DecodeResult<Self> {
                Ok(Self(<$inner>::$read(reader.read_bytes::<$size>()?)))
            }
        }
    };
}

fixed_codec!(I8, i8, 1, to_le_bytes, from_le_bytes);
fixed_codec!(U8, u8, 1, to_le_bytes, from_le_bytes);
fixed_codec!(I16LE, i16, 2, to_le_bytes, from_le_bytes);
fixed_codec!(I16BE, i16, 2, to_be_bytes, from_be_bytes);
fixed_codec!(U16LE, u16, 2, to_le_bytes, from_le_bytes);
fixed_codec!(U16BE, u16, 2, to_be_bytes, from_be_bytes);
fixed_codec!(I32LE, i32, 4, to_le_bytes, from_le_bytes);
fixed_codec!(I32BE, i32, 4, to_be_bytes, from_be_bytes);
fixed_codec!(U32LE, u32, 4, to_le_bytes, from_le_bytes);
fixed_codec!(U32BE, u32, 4, to_be_bytes, from_be_bytes);
fixed_codec!(I64LE, i64, 8, to_le_bytes, from_le_bytes);
fixed_codec!(I64BE, i64, 8, to_be_bytes, from_be_bytes);
fixed_codec!(U64LE, u64, 8, to_le_bytes, from_le_bytes);
fixed_codec!(U64BE, u64, 8, to_be_bytes, from_be_bytes);
fixed_float_codec!(F32LE, f32, 4, to_le_bytes, from_le_bytes);
fixed_float_codec!(F32BE, f32, 4, to_be_bytes, from_be_bytes);
fixed_float_codec!(F64LE, f64, 8, to_le_bytes, from_le_bytes);
fixed_float_codec!(F64BE, f64, 8, to_be_bytes, from_be_bytes);

macro_rules! var_codec {
    ($name:ident, $inner:ty, $bits:expr, $write:ident, $read:ident) => {
        #[derive(Clone, Copy, Debug, Default, PartialEq, Eq, Hash, PartialOrd, Ord)]
        pub struct $name(pub $inner);

        impl Encode for $name {
            fn encode(&self, writer: &mut Writer) {
                writer.$write(self.0);
            }
        }

        impl Decode for $name {
            fn decode(reader: &mut Reader<'_>) -> DecodeResult<Self> {
                reader.$read().map(Self)
            }
        }
    };
}

var_codec!(VarUInt, u32, 32, write_var_u32, read_var_u32);
var_codec!(VarULong, u64, 64, write_var_u64, read_var_u64);
var_codec!(ZigZag32, i32, 32, write_zigzag_i32, read_zigzag_i32);
var_codec!(ZigZag64, i64, 64, write_zigzag_i64, read_zigzag_i64);

/// Signed non-zigzag varint: the two's-complement value zero-extended to the
/// target width.
#[derive(Clone, Copy, Debug, Default, PartialEq, Eq, Hash, PartialOrd, Ord)]
pub struct VarInt(pub i32);

impl Encode for VarInt {
    fn encode(&self, writer: &mut Writer) {
        writer.write_var_u32(self.0 as u32);
    }
}

impl Decode for VarInt {
    fn decode(reader: &mut Reader<'_>) -> DecodeResult<Self> {
        reader.read_var_u32().map(|value| Self(value as i32))
    }
}

#[derive(Clone, Copy, Debug, Default, PartialEq, Eq, Hash, PartialOrd, Ord)]
pub struct VarLong(pub i64);

impl Encode for VarLong {
    fn encode(&self, writer: &mut Writer) {
        writer.write_var_u64(self.0 as u64);
    }
}

impl Decode for VarLong {
    fn decode(reader: &mut Reader<'_>) -> DecodeResult<Self> {
        reader.read_var_u64().map(|value| Self(value as i64))
    }
}

impl Encode for bool {
    fn encode(&self, writer: &mut Writer) {
        writer.write_u8(u8::from(*self));
    }
}

impl Decode for bool {
    fn decode(reader: &mut Reader<'_>) -> DecodeResult<Self> {
        Ok(reader.read_u8()? != 0)
    }
}

impl Encode for String {
    fn encode(&self, writer: &mut Writer) {
        writer.write_str(self);
    }
}

impl Decode for String {
    fn decode(reader: &mut Reader<'_>) -> DecodeResult<Self> {
        reader.read_str().map(str::to_owned)
    }
}

impl<T: Encode> Encode for Option<T> {
    fn encode(&self, writer: &mut Writer) {
        match self {
            Some(value) => {
                writer.write_u8(1);
                value.encode(writer);
            }
            None => writer.write_u8(0),
        }
    }
}

impl<T: Decode> Decode for Option<T> {
    fn decode(reader: &mut Reader<'_>) -> DecodeResult<Self> {
        match reader.read_u8()? {
            0 => Ok(None),
            _ => T::decode(reader).map(Some),
        }
    }
}

impl<T: Encode, const N: usize> Encode for [T; N] {
    fn encode(&self, writer: &mut Writer) {
        for value in self {
            value.encode(writer);
        }
    }
}

impl<T: Decode + Default + Copy, const N: usize> Decode for [T; N] {
    fn decode(reader: &mut Reader<'_>) -> DecodeResult<Self> {
        let mut out = [T::default(); N];
        for slot in out.iter_mut() {
            *slot = T::decode(reader)?;
        }
        Ok(out)
    }
}

/// Writes a `VarUInt`-prefixed collection.
pub fn encode_collection<T: Encode>(writer: &mut Writer, values: &[T]) {
    writer.write_var_u32(values.len() as u32);
    for value in values {
        value.encode(writer);
    }
}

pub fn encode_collection_limits<T: Encode>(
    writer: &mut Writer,
    values: &[T],
    minimum: u64,
    maximum: u64,
) {
    assert_length(values.len(), minimum, maximum);
    encode_collection(writer, values);
}

/// Reads a `VarUInt`-prefixed collection, bounding the declared count against
/// the reader's collection limit and the remaining input before reserving.
pub fn decode_collection<T: Decode>(
    reader: &mut Reader<'_>,
    min_element_size: usize,
) -> DecodeResult<Vec<T>> {
    let declared = u64::from(reader.read_var_u32()?);
    let count = reader.checked_count(declared, min_element_size)?;
    let mut out = Vec::with_capacity(count);
    for _ in 0..count {
        out.push(T::decode(reader)?);
    }
    Ok(out)
}

pub fn decode_collection_limits<T: Decode>(
    reader: &mut Reader<'_>,
    min_element_size: usize,
    minimum: u64,
    maximum: u64,
) -> DecodeResult<Vec<T>> {
    let declared = u64::from(reader.read_var_u32()?);
    let count = reader.checked_count_limits(declared, min_element_size, minimum, maximum)?;
    let mut out = Vec::with_capacity(count);
    for _ in 0..count {
        out.push(T::decode(reader)?);
    }
    Ok(out)
}

pub fn assert_length(actual: usize, minimum: u64, maximum: u64) {
    assert!(
        validate_length(actual as u64, minimum, maximum).is_ok(),
        "length outside schema limits"
    );
}

fn validate_length(actual: u64, minimum: u64, maximum: u64) -> DecodeResult<()> {
    if actual < minimum {
        return Err(DecodeError::SchemaConstraint("length below minimum"));
    }
    if actual > maximum {
        return Err(DecodeError::SchemaConstraint("length above maximum"));
    }
    Ok(())
}

pub fn encode_string_limits(writer: &mut Writer, value: &str, minimum: u64, maximum: u64) {
    assert_length(value.len(), minimum, maximum);
    writer.write_str(value);
}

pub fn decode_string_limits(
    reader: &mut Reader<'_>,
    minimum: u64,
    maximum: u64,
) -> DecodeResult<String> {
    let declared = u64::from(reader.read_var_u32()?);
    validate_length(declared, minimum, maximum)?;
    let count = reader.checked_count_with(declared, 1, reader.byte_buffer_limit())?;
    let bytes = reader.take(count)?;
    std::str::from_utf8(bytes)
        .map(str::to_owned)
        .map_err(|_| DecodeError::InvalidUtf8)
}

pub fn encode_bytes_limits(writer: &mut Writer, value: &[u8], minimum: u64, maximum: u64) {
    assert_length(value.len(), minimum, maximum);
    writer.write_byte_slice(value);
}

pub fn decode_bytes_limits(
    reader: &mut Reader<'_>,
    minimum: u64,
    maximum: u64,
) -> DecodeResult<bytes::Bytes> {
    let declared = u64::from(reader.read_var_u32()?);
    validate_length(declared, minimum, maximum)?;
    let count = reader.checked_count_with(declared, 1, reader.byte_buffer_limit())?;
    reader.take_shared(count)
}

pub fn assert_number_limits<T: PartialOrd + Copy>(
    value: T,
    minimum: Option<T>,
    maximum: Option<T>,
) {
    assert!(
        validate_number_limits(value, minimum, maximum).is_ok(),
        "number outside schema limits"
    );
}

pub fn validate_number_limits<T: PartialOrd + Copy>(
    value: T,
    minimum: Option<T>,
    maximum: Option<T>,
) -> DecodeResult<()> {
    if minimum.is_some_and(|minimum| value < minimum) {
        return Err(DecodeError::SchemaConstraint("number below minimum"));
    }
    if maximum.is_some_and(|maximum| value > maximum) {
        return Err(DecodeError::SchemaConstraint("number above maximum"));
    }
    Ok(())
}

/// Writes a fixed-width length and a schema-bounded string.
pub fn encode_string_u32le(writer: &mut Writer, value: &str, minimum: u64, maximum: u64) {
    encode_bytes_u32le(writer, value.as_bytes(), minimum, maximum);
}

/// Reads a fixed-width string length under byte and schema bounds.
pub fn decode_string_u32le(
    reader: &mut Reader<'_>,
    minimum: u64,
    maximum: u64,
) -> DecodeResult<String> {
    let declared = u64::from(u32::from_le_bytes(reader.read_bytes::<4>()?));
    validate_length(declared, minimum, maximum)?;
    let count = reader.checked_count_with(declared, 1, reader.byte_buffer_limit())?;
    std::str::from_utf8(reader.take(count)?)
        .map(str::to_owned)
        .map_err(|_| DecodeError::InvalidUtf8)
}

/// Writes a fixed-width length and a schema-bounded byte buffer.
pub fn encode_bytes_u32le(writer: &mut Writer, value: &[u8], minimum: u64, maximum: u64) {
    assert_length(value.len(), minimum, maximum);
    let count = u32::try_from(value.len()).expect("buffer length exceeds u32");
    writer.write_all(&count.to_le_bytes());
    writer.write_all(value);
}

/// Reads a fixed-width buffer length under byte and schema bounds.
pub fn decode_bytes_u32le(
    reader: &mut Reader<'_>,
    minimum: u64,
    maximum: u64,
) -> DecodeResult<bytes::Bytes> {
    let declared = u64::from(u32::from_le_bytes(reader.read_bytes::<4>()?));
    validate_length(declared, minimum, maximum)?;
    let count = reader.checked_count_with(declared, 1, reader.byte_buffer_limit())?;
    reader.take_shared(count)
}
/// Client-owned ceiling on recursive schema edges along a decoded value's path.
/// This bounds decoder stack use and the resulting owned tree's clone/drop depth;
/// it is independent of collection/byte limits and is not a native protocol maximum.
pub const MAX_RECURSIVE_VALUE_DEPTH: usize = 64;

impl Reader<'_> {
    /// Decodes one recursive child, restoring depth capacity on success or a decode error.
    /// Refusal consumes no child bytes; cursor changes made by `decode` are not rolled back.
    pub fn with_recursive_value<T>(
        &mut self,
        decode: impl FnOnce(&mut Self) -> DecodeResult<T>,
    ) -> DecodeResult<T> {
        if self.recursive_depth == MAX_RECURSIVE_VALUE_DEPTH {
            return Err(DecodeError::RecursiveDepthExceeded);
        }
        self.recursive_depth += 1;
        let result = decode(self);
        self.recursive_depth -= 1;
        result
    }
}

// ---------------------------------------------------------------------------
// Collections, maps, and bitsets
// ---------------------------------------------------------------------------

/// Writes a `u32le`-prefixed collection.
pub fn encode_collection_u32le<T: Encode>(writer: &mut Writer, values: &[T]) {
    writer.write_all(&(values.len() as u32).to_le_bytes());
    for value in values {
        value.encode(writer);
    }
}

/// Reads a `u32le`-prefixed collection under the same bounds as
/// [`decode_collection`].
pub fn decode_collection_u32le<T: Decode>(
    reader: &mut Reader<'_>,
    min_element_size: usize,
) -> DecodeResult<Vec<T>> {
    let declared = u64::from(u32::from_le_bytes(reader.read_bytes::<4>()?));
    let count = reader.checked_count(declared, min_element_size)?;
    let mut out = Vec::with_capacity(count);
    for _ in 0..count {
        out.push(T::decode(reader)?);
    }
    Ok(out)
}

/// Writes a `VarUInt`-prefixed map as ordered key/value pairs.
pub fn encode_map<K: Encode, V: Encode>(writer: &mut Writer, entries: &[(K, V)]) {
    writer.write_var_u32(entries.len() as u32);
    for (key, value) in entries {
        key.encode(writer);
        value.encode(writer);
    }
}

pub fn encode_map_limits<K: Encode, V: Encode>(
    writer: &mut Writer,
    entries: &[(K, V)],
    minimum: u64,
    maximum: u64,
) {
    assert_length(entries.len(), minimum, maximum);
    encode_map(writer, entries);
}

/// Reads a `VarUInt`-prefixed map, preserving wire order.
pub fn decode_map<K: Decode, V: Decode>(
    reader: &mut Reader<'_>,
    min_entry_size: usize,
) -> DecodeResult<Vec<(K, V)>> {
    let declared = u64::from(reader.read_var_u32()?);
    let count = reader.checked_count(declared, min_entry_size)?;
    let mut out = Vec::with_capacity(count);
    for _ in 0..count {
        let key = K::decode(reader)?;
        let value = V::decode(reader)?;
        out.push((key, value));
    }
    Ok(out)
}

pub fn decode_map_limits<K: Decode, V: Decode>(
    reader: &mut Reader<'_>,
    min_entry_size: usize,
    minimum: u64,
    maximum: u64,
) -> DecodeResult<Vec<(K, V)>> {
    let declared = u64::from(reader.read_var_u32()?);
    let count = reader.checked_count_limits(declared, min_entry_size, minimum, maximum)?;
    let mut out = Vec::with_capacity(count);
    for _ in 0..count {
        let key = K::decode(reader)?;
        let value = V::decode(reader)?;
        out.push((key, value));
    }
    Ok(out)
}

/// Writes `bits` bits using seven payload bits per continuation byte.
///
/// Trailing zero groups are omitted, so an all-zero bitset is a single `0`
/// byte.
pub fn encode_bitset(writer: &mut Writer, words: &[u64], bits: u64) {
    let mut last = 0u64;
    let mut found = false;
    for (index, word) in words.iter().enumerate() {
        if *word == 0 {
            continue;
        }
        found = true;
        for bit in (0..64).rev() {
            if word & (1u64 << bit) != 0 {
                last = index as u64 * 64 + bit;
                break;
            }
        }
    }
    if !found {
        writer.write_u8(0);
        return;
    }
    let groups = last / 7 + 1;
    for group in 0..groups {
        let offset = group * 7;
        let width = (bits - offset).min(7);
        let mut value = 0u8;
        for bit in 0..width {
            let index = offset + bit;
            if words[(index / 64) as usize] & (1u64 << (index % 64)) != 0 {
                value |= 1 << bit;
            }
        }
        if group + 1 < groups {
            value |= 0x80;
        }
        writer.write_u8(value);
    }
}

/// Reads a seven-bits-per-byte bitset, rejecting payload bits outside `bits`.
pub fn decode_bitset<const N: usize>(reader: &mut Reader<'_>, bits: u64) -> DecodeResult<[u64; N]> {
    let mut words = [0u64; N];
    let mut offset = 0u64;
    while offset < bits {
        let value = reader.read_u8()?;
        let width = (bits - offset).min(7);
        if width < 7 && u64::from(value & 0x7f) >= (1u64 << width) {
            return Err(DecodeError::NbtMalformed(
                "bitset bits outside declared width",
            ));
        }
        for bit in 0..width {
            if value & (1 << bit) != 0 {
                let index = offset + bit;
                words[(index / 64) as usize] |= 1u64 << (index % 64);
            }
        }
        if value & 0x80 == 0 {
            return Ok(words);
        }
        if offset + 7 >= bits {
            return Err(DecodeError::NbtMalformed("bitset exceeds declared width"));
        }
        offset += 7;
    }
    Ok(words)
}

// ---------------------------------------------------------------------------
// Length-prefixed byte buffers
// ---------------------------------------------------------------------------

impl Encode for bytes::Bytes {
    fn encode(&self, writer: &mut Writer) {
        writer.write_byte_slice(self);
    }
}

impl Decode for bytes::Bytes {
    fn decode(reader: &mut Reader<'_>) -> DecodeResult<Self> {
        let slice = reader.read_byte_slice()?;
        Ok(reader.share(slice))
    }
}

// ---------------------------------------------------------------------------
// NBT
// ---------------------------------------------------------------------------

/// Integer and length encoding used by an NBT payload.
///
/// Bedrock uses both forms and they are not interchangeable: the network form
/// varint-encodes integers and lengths, the persistent form stores them as
/// fixed-width little endian. A payload read with the wrong variant desyncs the
/// surrounding packet rather than failing locally.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum NbtVariant {
    Network,
    /// A network compound body without its root tag ID and name.
    NetworkCompound,
    Persistent,
}

const TAG_END: u8 = 0;
const TAG_BYTE: u8 = 1;
const TAG_SHORT: u8 = 2;
const TAG_INT: u8 = 3;
const TAG_LONG: u8 = 4;
const TAG_FLOAT: u8 = 5;
const TAG_DOUBLE: u8 = 6;
const TAG_BYTE_ARRAY: u8 = 7;
const TAG_STRING: u8 = 8;
const TAG_LIST: u8 = 9;
const TAG_COMPOUND: u8 = 10;
const TAG_INT_ARRAY: u8 = 11;
const TAG_LONG_ARRAY: u8 = 12;

impl<'a> Reader<'a> {
    fn nbt_len(&mut self, variant: NbtVariant) -> DecodeResult<usize> {
        let declared = match variant {
            NbtVariant::Network | NbtVariant::NetworkCompound => i64::from(self.read_zigzag_i32()?),
            NbtVariant::Persistent => i64::from(i32::from_le_bytes(self.read_bytes::<4>()?)),
        };
        if declared < 0 {
            return Err(DecodeError::NegativeLength(declared));
        }
        self.checked_count_with(declared as u64, 1, self.byte_buffer_limit())
    }

    fn nbt_name(&mut self, variant: NbtVariant) -> DecodeResult<()> {
        let count = match variant {
            NbtVariant::Network | NbtVariant::NetworkCompound => {
                let declared = u64::from(self.read_var_u32()?);
                self.checked_count_with(declared, 1, self.byte_buffer_limit())?
            }
            NbtVariant::Persistent => usize::from(u16::from_le_bytes(self.read_bytes::<2>()?)),
        };
        self.take(count)?;
        Ok(())
    }

    fn scan_nbt_payload(&mut self, tag: u8, variant: NbtVariant, depth: u32) -> DecodeResult<()> {
        if depth > MAX_NBT_DEPTH {
            return Err(DecodeError::NbtDepthExceeded);
        }
        match tag {
            TAG_BYTE => {
                self.take(1)?;
            }
            TAG_SHORT => {
                self.take(2)?;
            }
            TAG_INT => match variant {
                NbtVariant::Network | NbtVariant::NetworkCompound => {
                    self.read_zigzag_i32()?;
                }
                NbtVariant::Persistent => {
                    self.take(4)?;
                }
            },
            TAG_LONG => match variant {
                NbtVariant::Network | NbtVariant::NetworkCompound => {
                    self.read_zigzag_i64()?;
                }
                NbtVariant::Persistent => {
                    self.take(8)?;
                }
            },
            TAG_FLOAT => {
                self.take(4)?;
            }
            TAG_DOUBLE => {
                self.take(8)?;
            }
            TAG_BYTE_ARRAY => {
                let count = self.nbt_len(variant)?;
                self.take(count)?;
            }
            TAG_STRING => self.nbt_name(variant)?,
            TAG_LIST => {
                let element = self.read_u8()?;
                let count = self.nbt_len(variant)?;
                if element == TAG_END && count != 0 {
                    return Err(DecodeError::NbtMalformed("non-empty list of TAG_End"));
                }
                for _ in 0..count {
                    self.scan_nbt_payload(element, variant, depth + 1)?;
                }
            }
            TAG_COMPOUND => loop {
                let entry = self.read_u8()?;
                if entry == TAG_END {
                    break;
                }
                self.nbt_name(variant)?;
                self.scan_nbt_payload(entry, variant, depth + 1)?;
            },
            TAG_INT_ARRAY => {
                let count = self.nbt_len(variant)?;
                for _ in 0..count {
                    match variant {
                        NbtVariant::Network | NbtVariant::NetworkCompound => {
                            self.read_zigzag_i32()?;
                        }
                        NbtVariant::Persistent => {
                            self.take(4)?;
                        }
                    }
                }
            }
            TAG_LONG_ARRAY => {
                let count = self.nbt_len(variant)?;
                for _ in 0..count {
                    match variant {
                        NbtVariant::Network | NbtVariant::NetworkCompound => {
                            self.read_zigzag_i64()?;
                        }
                        NbtVariant::Persistent => {
                            self.take(8)?;
                        }
                    }
                }
            }
            _ => return Err(DecodeError::NbtMalformed("unknown tag")),
        }
        Ok(())
    }

    /// Advances past exactly one root NBT value and borrows its bytes.
    pub fn read_nbt(&mut self, variant: NbtVariant) -> DecodeResult<&'a [u8]> {
        let start = self.pos;
        if variant == NbtVariant::NetworkCompound {
            self.scan_nbt_payload(TAG_COMPOUND, NbtVariant::Network, 0)?;
            return Ok(&self.buf[start..self.pos]);
        }
        let tag = self.read_u8()?;
        if tag != TAG_END {
            self.nbt_name(variant)?;
            self.scan_nbt_payload(tag, variant, 0)?;
        }
        Ok(&self.buf[start..self.pos])
    }
}

macro_rules! nbt_codec {
    ($name:ident, $variant:expr, $doc:expr) => {
        #[doc = $doc]
        #[derive(Clone, Debug, PartialEq, Eq)]
        pub struct $name(pub bytes::Bytes);

        impl Default for $name {
            /// A lone TAG_End, which is the empty payload the scanner accepts.
            fn default() -> Self {
                Self(bytes::Bytes::from_static(&[0u8]))
            }
        }

        impl $name {
            pub const VARIANT: NbtVariant = $variant;
        }

        impl Encode for $name {
            fn encode(&self, writer: &mut Writer) {
                writer.write_all(&self.0);
            }
        }

        impl Decode for $name {
            fn decode(reader: &mut Reader<'_>) -> DecodeResult<Self> {
                let payload = reader.read_nbt($variant)?;
                Ok(Self(reader.share(payload)))
            }
        }
    };
}

nbt_codec!(
    NetworkNbt,
    NbtVariant::Network,
    "NBT with varint integers and lengths."
);
nbt_codec!(
    NetworkCompoundNbt,
    NbtVariant::NetworkCompound,
    "Network NBT compound body without a root tag ID or name, ending in TAG_End."
);
nbt_codec!(
    PersistentNbt,
    NbtVariant::Persistent,
    "NBT with fixed-width little-endian integers and lengths."
);

#[cfg(test)]
mod nbt_tests {
    use super::*;

    /// The same bytes mean different things under the two variants, so the
    /// scanner must consume a different number of them.
    #[test]
    fn network_and_persistent_integers_differ() {
        // TAG_Int("", 1): network varint-encodes the value, persistent uses i32le.
        let network = [0x03u8, 0x00, 0x02];
        let mut reader = Reader::new(&network);
        assert_eq!(reader.read_nbt(NbtVariant::Network), Ok(&network[..]));

        let persistent = [0x03u8, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00];
        let mut reader = Reader::new(&persistent);
        assert_eq!(reader.read_nbt(NbtVariant::Persistent), Ok(&persistent[..]));

        // Reading the persistent payload as network leaves bytes unconsumed.
        let mut reader = Reader::new(&persistent);
        reader.read_nbt(NbtVariant::Network).expect("scan");
        assert_ne!(reader.remaining(), 0);
    }

    /// Compound bodies omit the root header but keep their end tag and packet boundary.
    #[test]
    fn network_compound_framing() {
        for body in [
            &[0][..],
            &[3, 1, b'x', 2, 0],
            &[10, 1, b'n', 3, 1, b'x', 2, 0, 0],
        ] {
            let mut bytes = body.to_vec();
            bytes.push(0x7f);
            let mut reader = Reader::new(&bytes);
            assert_eq!(reader.read_nbt(NbtVariant::NetworkCompound), Ok(body));
            assert_eq!(reader.remaining(), 1);
            for end in 0..body.len() {
                let mut reader = Reader::new(&body[..end]);
                assert!(reader.read_nbt(NbtVariant::NetworkCompound).is_err());
            }
        }
    }

    #[test]
    fn empty_root_is_a_single_end_tag() {
        let mut reader = Reader::new(&[0x00]);
        assert_eq!(reader.read_nbt(NbtVariant::Network), Ok(&[0x00][..]));
        assert_eq!(reader.remaining(), 0);
    }

    #[test]
    fn nesting_is_bounded() {
        // Root compound with an empty name, then a chain of compounds nested
        // deeper than the cap. Each entry is TAG_Compound plus a zero-length
        // name; the scanner errors before reaching the closing TAG_Ends.
        let depth = MAX_NBT_DEPTH + 8;
        let mut deep = vec![0x0a, 0x00];
        for _ in 0..depth {
            deep.extend_from_slice(&[0x0a, 0x00]);
        }
        deep.extend(std::iter::repeat_n(0x00, depth as usize + 1));
        let mut reader = Reader::new(&deep);
        assert_eq!(
            reader.read_nbt(NbtVariant::Network),
            Err(DecodeError::NbtDepthExceeded)
        );
    }

    #[test]
    fn truncated_payloads_fail_rather_than_over_read() {
        let mut reader = Reader::new(&[0x03, 0x00]);
        assert!(reader.read_nbt(NbtVariant::Persistent).is_err());
    }
}

// ---------------------------------------------------------------------------
// UUID
// ---------------------------------------------------------------------------

impl Encode for uuid::Uuid {
    fn encode(&self, writer: &mut Writer) {
        // Bedrock transmits a UUID as two little-endian 64-bit halves, not as
        // the raw byte order returned by `Uuid::as_bytes`.
        let (high, low) = self.as_u64_pair();
        writer.write_all(&high.to_le_bytes());
        writer.write_all(&low.to_le_bytes());
    }
}

impl Decode for uuid::Uuid {
    fn decode(reader: &mut Reader<'_>) -> DecodeResult<Self> {
        let high = u64::from_le_bytes(reader.read_bytes::<8>()?);
        let low = u64::from_le_bytes(reader.read_bytes::<8>()?);
        Ok(uuid::Uuid::from_u64_pair(high, low))
    }
}

// ---------------------------------------------------------------------------
// Vectors
// ---------------------------------------------------------------------------

impl Encode for glam::Vec2 {
    fn encode(&self, writer: &mut Writer) {
        writer.write_all(&self.x.to_le_bytes());
        writer.write_all(&self.y.to_le_bytes());
    }
}

impl Decode for glam::Vec2 {
    fn decode(reader: &mut Reader<'_>) -> DecodeResult<Self> {
        let x = f32::from_le_bytes(reader.read_bytes::<4>()?);
        let y = f32::from_le_bytes(reader.read_bytes::<4>()?);
        Ok(glam::Vec2::new(x, y))
    }
}

impl Encode for glam::Vec3 {
    fn encode(&self, writer: &mut Writer) {
        writer.write_all(&self.x.to_le_bytes());
        writer.write_all(&self.y.to_le_bytes());
        writer.write_all(&self.z.to_le_bytes());
    }
}

impl Decode for glam::Vec3 {
    fn decode(reader: &mut Reader<'_>) -> DecodeResult<Self> {
        let x = f32::from_le_bytes(reader.read_bytes::<4>()?);
        let y = f32::from_le_bytes(reader.read_bytes::<4>()?);
        let z = f32::from_le_bytes(reader.read_bytes::<4>()?);
        Ok(glam::Vec3::new(x, y, z))
    }
}

// ---------------------------------------------------------------------------
// Schema string patterns
// ---------------------------------------------------------------------------

use std::collections::HashMap;
use std::sync::{LazyLock, Mutex};

static SCHEMA_PATTERNS: LazyLock<Mutex<HashMap<&'static str, regex::Regex>>> =
    LazyLock::new(|| Mutex::new(HashMap::new()));

pub fn assert_pattern(value: &str, pattern: &'static str) {
    assert!(
        validate_pattern(value, pattern).is_ok(),
        "string does not match schema pattern"
    );
}

pub fn validate_pattern(value: &str, pattern: &'static str) -> DecodeResult<()> {
    let mut patterns = SCHEMA_PATTERNS
        .lock()
        .expect("schema pattern cache poisoned");
    let compiled = patterns.entry(pattern).or_insert_with(|| {
        regex::Regex::new(pattern).expect("manifest contains an invalid schema pattern")
    });
    if compiled.is_match(value) {
        Ok(())
    } else {
        Err(DecodeError::SchemaConstraint(
            "string does not match pattern",
        ))
    }
}

#[cfg(test)]
mod pattern_tests {
    use super::*;

    #[test]
    fn patterns_are_enforced() {
        assert_eq!(validate_pattern("abc", "^[a-z]+$"), Ok(()));
        assert_eq!(
            validate_pattern("123", "^[a-z]+$"),
            Err(DecodeError::SchemaConstraint(
                "string does not match pattern"
            ))
        );
    }
}

// ---------------------------------------------------------------------------
// Runtime tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod runtime_tests {
    use super::*;

    /// Empty values consume no payload bytes but still obey the element limit.
    #[test]
    fn zero_byte_collections_obey_element_limits() {
        let mut reader = Reader::new(&[]);
        assert_eq!(reader.checked_count(3, 0), Ok(3));
        reader.set_collection_limit(2);
        assert!(matches!(
            reader.checked_count(3, 0),
            Err(DecodeError::LengthLimitExceeded { .. })
        ));
    }

    /// Fixed-width strings and buffers preserve their declared prefix bytes.
    #[test]
    fn fixed_width_buffers_round_trip() {
        let mut writer = Writer::new();
        encode_string_u32le(&mut writer, "hi", 0, 2);
        assert_eq!(writer.as_slice(), [2, 0, 0, 0, b'h', b'i']);
        assert_eq!(
            decode_string_u32le(&mut Reader::new(writer.as_slice()), 0, 2).unwrap(),
            "hi"
        );
        assert!(decode_bytes_u32le(&mut Reader::new(writer.as_slice()), 0, 1).is_err());
    }

    #[test]
    fn varint_rejects_overlong_encodings() {
        // 0x80 0x00 encodes 0 in two bytes; the canonical form is one byte.
        assert_eq!(
            Reader::new(&[0x80, 0x00]).read_var_u32(),
            Err(DecodeError::VarIntOverlong)
        );
        assert_eq!(Reader::new(&[0x00]).read_var_u32(), Ok(0));
    }

    /// A ten-byte varint must not silently drop the bits that do not fit.
    #[test]
    fn varint_rejects_values_wider_than_their_target() {
        let ten_bytes = [0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x7f];
        assert_eq!(
            Reader::new(&ten_bytes).read_var_u64(),
            Err(DecodeError::VarIntTooLarge)
        );
        let six_bytes = [0xff, 0xff, 0xff, 0xff, 0xff, 0x7f];
        assert_eq!(
            Reader::new(&six_bytes).read_var_u32(),
            Err(DecodeError::VarIntTooLarge)
        );
    }

    #[test]
    fn varint_round_trips_boundary_values() {
        for value in [0u64, 1, 127, 128, u32::MAX as u64, u64::MAX] {
            let mut writer = Writer::new();
            writer.write_var_u64(value);
            assert_eq!(
                Reader::new(writer.as_slice()).read_var_u64(),
                Ok(value),
                "varint {value}"
            );
        }
    }

    #[test]
    fn zigzag_round_trips_boundary_values() {
        for value in [0i32, -1, 1, i32::MIN, i32::MAX] {
            let mut writer = Writer::new();
            writer.write_zigzag_i32(value);
            assert_eq!(
                Reader::new(writer.as_slice()).read_zigzag_i32(),
                Ok(value),
                "zigzag {value}"
            );
        }
    }

    /// A declared count is rejected before it is used to reserve memory.
    #[test]
    fn declared_counts_are_bounded_before_allocation() {
        let reader = Reader::new(&[0u8; 4]);
        assert_eq!(
            reader.checked_count(u64::from(u32::MAX), 1),
            Err(DecodeError::LengthLimitExceeded {
                limit: MAX_COLLECTION_ELEMENTS,
                actual: u32::MAX as usize,
            })
        );
        // Within the limit, but more elements than the input could hold.
        assert_eq!(
            reader.checked_count(100, 8),
            Err(DecodeError::LengthNotRepresentable {
                needed: 800,
                remaining: 4
            })
        );
        assert_eq!(reader.checked_count(4, 1), Ok(4));
    }

    /// A peer may legitimately exceed the default bound, so the limit is a
    /// reader setting rather than a hard cap.
    #[test]
    fn the_collection_limit_is_configurable() {
        let data = [0u8; 8192];
        let mut reader = Reader::new(&data);
        assert_eq!(reader.collection_limit(), MAX_COLLECTION_ELEMENTS);
        assert!(reader.checked_count(8192, 1).is_err());
        reader.set_collection_limit(8192);
        assert_eq!(reader.checked_count(8192, 1), Ok(8192));
    }

    /// Byte buffers decoded from a shared source refcount it instead of copying.
    #[test]
    fn shared_readers_do_not_copy_byte_buffers() {
        let source = bytes::Bytes::from_static(&[0x04, 0xde, 0xad, 0xbe, 0xef]);
        let mut reader = Reader::from_shared(&source);
        let decoded = <bytes::Bytes as Decode>::decode(&mut reader).expect("decode");
        assert_eq!(decoded, source.slice(1..));
        assert_eq!(decoded.as_ptr(), source[1..].as_ptr(), "buffer was copied");

        // An unshared reader still decodes correctly, by copying.
        let mut reader = Reader::new(&source);
        let copied = <bytes::Bytes as Decode>::decode(&mut reader).expect("decode");
        assert_eq!(copied, source.slice(1..));
    }

    #[test]
    fn strings_reject_invalid_utf8() {
        assert_eq!(
            Reader::new(&[0x01, 0xff]).read_str(),
            Err(DecodeError::InvalidUtf8)
        );
    }

    #[test]
    fn schema_string_limit_is_checked_before_payload_read() {
        let mut reader = Reader::new(&[0x05, b'h', b'e', b'l', b'l', b'o']);
        assert_eq!(
            decode_string_limits(&mut reader, 0, 4),
            Err(DecodeError::SchemaConstraint("length above maximum"))
        );
        assert_eq!(reader.remaining(), 5);
    }

    #[test]
    fn schema_number_limits_reject_out_of_range_values() {
        assert_eq!(
            validate_number_limits(65i32, Some(-1), Some(64)),
            Err(DecodeError::SchemaConstraint("number above maximum"))
        );
        assert_eq!(validate_number_limits(64i32, Some(-1), Some(64)), Ok(()));
    }

    #[test]
    fn trailing_bytes_are_rejected() {
        assert_eq!(
            Reader::new(&[0x00, 0x01]).expect_consumed(),
            Err(DecodeError::TrailingBytes(2))
        );
    }
}

#[cfg(test)]
mod bitset_tests {
    use super::*;

    #[test]
    fn bitsets_round_trip() {
        for words in [
            [0u64, 0, 0],
            [1, 0, 0],
            [0, 0, 1 << 2],
            [u64::MAX, u64::MAX, 0b111],
        ] {
            let mut writer = Writer::new();
            encode_bitset(&mut writer, &words, 131);
            let decoded: [u64; 3] =
                decode_bitset(&mut Reader::new(writer.as_slice()), 131).expect("decode");
            assert_eq!(decoded, words, "bitset {words:?}");
        }
    }

    #[test]
    fn empty_bitset_is_one_byte() {
        let mut writer = Writer::new();
        encode_bitset(&mut writer, &[0u64; 3], 131);
        assert_eq!(writer.as_slice(), &[0x00]);
    }
}
