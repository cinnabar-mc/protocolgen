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
    UnexpectedEof { needed: usize, remaining: usize },
    /// A varint used more bytes than its canonical encoding requires.
    VarIntOverlong,
    /// A varint did not terminate within the width of its target type.
    VarIntTooLarge,
    InvalidUtf8,
    NegativeLength(i64),
    /// A declared element count exceeded the configured bound.
    LengthLimitExceeded { limit: usize, actual: usize },
    /// A declared element count could not be covered by the remaining bytes.
    LengthNotRepresentable { needed: usize, remaining: usize },
    SchemaConstraint(&'static str),
    UnknownVariant { type_name: &'static str, value: i64 },
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
                write!(f, "unexpected end of input: needed {needed}, {remaining} remaining")
            }
            Self::VarIntOverlong => write!(f, "varint is not canonically encoded"),
            Self::VarIntTooLarge => write!(f, "varint exceeds the width of its target type"),
            Self::InvalidUtf8 => write!(f, "string is not valid UTF-8"),
            Self::NegativeLength(value) => write!(f, "negative length {value}"),
            Self::LengthLimitExceeded { limit, actual } => {
                write!(f, "declared length {actual} exceeds limit {limit}")
            }
            Self::LengthNotRepresentable { needed, remaining } => {
                write!(f, "declared length needs {needed} bytes, {remaining} remaining")
            }
            Self::SchemaConstraint(reason) => write!(f, "schema constraint violated: {reason}"),
            Self::UnknownVariant { type_name, value } => {
                write!(f, "unknown {type_name} discriminant {value}")
            }
            Self::NbtDepthExceeded => write!(f, "NBT nesting exceeds {MAX_NBT_DEPTH}"),
            Self::RecursiveDepthExceeded => {
                write!(f, "recursive value depth exceeds {MAX_RECURSIVE_VALUE_DEPTH}")
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
        Self { shared: Some(source), ..Self::new(source) }
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
            return Err(DecodeError::UnexpectedEof { needed: count, remaining: self.remaining() });
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
            return Err(DecodeError::LengthLimitExceeded { limit, actual: declared as usize });
        }
        let count = declared as usize;
        let needed = count.saturating_mul(min_element_size);
        if needed > self.remaining() {
            return Err(DecodeError::LengthNotRepresentable { needed, remaining: self.remaining() });
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
        Self { buf: Vec::with_capacity(capacity) }
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

pub fn encode_collection_limits<T: Encode>(writer: &mut Writer, values: &[T], minimum: u64, maximum: u64) {
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
    assert!(validate_length(actual as u64, minimum, maximum).is_ok(), "length outside schema limits");
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

pub fn decode_string_limits(reader: &mut Reader<'_>, minimum: u64, maximum: u64) -> DecodeResult<String> {
    let declared = u64::from(reader.read_var_u32()?);
    validate_length(declared, minimum, maximum)?;
    let count = reader.checked_count_with(declared, 1, reader.byte_buffer_limit())?;
    let bytes = reader.take(count)?;
    std::str::from_utf8(bytes).map(str::to_owned).map_err(|_| DecodeError::InvalidUtf8)
}

pub fn encode_bytes_limits(writer: &mut Writer, value: &[u8], minimum: u64, maximum: u64) {
    assert_length(value.len(), minimum, maximum);
    writer.write_byte_slice(value);
}

pub fn decode_bytes_limits(reader: &mut Reader<'_>, minimum: u64, maximum: u64) -> DecodeResult<bytes::Bytes> {
    let declared = u64::from(reader.read_var_u32()?);
    validate_length(declared, minimum, maximum)?;
    let count = reader.checked_count_with(declared, 1, reader.byte_buffer_limit())?;
    reader.take_shared(count)
}

pub fn assert_number_limits<T: PartialOrd + Copy>(value: T, minimum: Option<T>, maximum: Option<T>) {
    assert!(validate_number_limits(value, minimum, maximum).is_ok(), "number outside schema limits");
}

pub fn validate_number_limits<T: PartialOrd + Copy>(value: T, minimum: Option<T>, maximum: Option<T>) -> DecodeResult<()> {
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
pub fn decode_string_u32le(reader: &mut Reader<'_>, minimum: u64, maximum: u64) -> DecodeResult<String> {
    let declared = u64::from(u32::from_le_bytes(reader.read_bytes::<4>()?));
    validate_length(declared, minimum, maximum)?;
    let count = reader.checked_count_with(declared, 1, reader.byte_buffer_limit())?;
    std::str::from_utf8(reader.take(count)?).map(str::to_owned).map_err(|_| DecodeError::InvalidUtf8)
}

/// Writes a fixed-width length and a schema-bounded byte buffer.
pub fn encode_bytes_u32le(writer: &mut Writer, value: &[u8], minimum: u64, maximum: u64) {
    assert_length(value.len(), minimum, maximum);
    let count = u32::try_from(value.len()).expect("buffer length exceeds u32");
    writer.write_all(&count.to_le_bytes());
    writer.write_all(value);
}

/// Reads a fixed-width buffer length under byte and schema bounds.
pub fn decode_bytes_u32le(reader: &mut Reader<'_>, minimum: u64, maximum: u64) -> DecodeResult<bytes::Bytes> {
    let declared = u64::from(u32::from_le_bytes(reader.read_bytes::<4>()?));
    validate_length(declared, minimum, maximum)?;
    let count = reader.checked_count_with(declared, 1, reader.byte_buffer_limit())?;
    reader.take_shared(count)
}
