// Code generated from canonical protocol manifest v2. DO NOT EDIT.

//! Wire runtime for Rust codec contract v2 (`docs/rust-codec-contract.md`).
//!
//! Decoding is slice-based: every declared length is checked against the
//! remaining input and the reader's [`Limits`] before anything is allocated,
//! and collection storage grows fallibly. Encoding is fallible and appends to a
//! caller-owned buffer.

use std::fmt;

/// Largest compound/list nesting depth accepted by the NBT scanners.
pub const MAX_NBT_DEPTH: u32 = 512;

/// Default element ceiling when receiving from a client, matching the Go
/// reader's `maxSliceLength`.
pub const CLIENT_COLLECTION_LIMIT: usize = 4096;

/// Largest number of bytes reserved up front for one decoded collection; the
/// rest grows fallibly as elements decode.
pub const MAX_EAGER_RESERVE_BYTES: usize = 1 << 20;

/// Widest encoding of a `var_u32`.
pub const MAX_VAR_U32_LEN: usize = 5;

/// Identifies one length-prefixed field site; the generated `fields` module
/// names every site this crate decodes.
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord)]
pub struct FieldId(pub u16);

impl FieldId {
    /// A length read outside any named site, such as inside NBT.
    pub const NONE: Self = Self(u16::MAX);
}

#[derive(Clone, Debug, PartialEq, Eq)]
pub enum DecodeError {
    UnexpectedEof { needed: usize, remaining: usize },
    /// A varint used more bytes than its canonical encoding requires.
    VarIntOverlong,
    /// A varint did not terminate within the width of its target type.
    VarIntTooLarge,
    NegativeLength(i64),
    /// A declared length exceeded the limit configured for its field.
    LengthLimitExceeded { field: FieldId, limit: usize, actual: u64 },
    /// A declared length could not be covered by the remaining bytes.
    LengthNotRepresentable { field: FieldId, needed: usize, remaining: usize },
    /// The allocator refused storage for a decoded value.
    AllocationFailed { requested: usize },
    SchemaConstraint(&'static str),
    UnknownVariant { type_name: &'static str, value: i64 },
    InvalidBitset(&'static str),
    NbtDepthExceeded,
    /// A recursive schema edge exceeded the client-owned decoded-tree depth budget.
    RecursiveDepthExceeded,
    NbtMalformed(&'static str),
    UnknownPacketId(u32),
    /// A known packet id that the sending peer is not permitted to use.
    UnexpectedDirection(u32),
    /// Bytes remained inside a region that declared its own length.
    TrailingBytes(usize),
    InvalidBatchHeader(Option<u8>),
    TooManyFrames { limit: usize },
}

impl fmt::Display for DecodeError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::UnexpectedEof { needed, remaining } => {
                write!(f, "unexpected end of input: needed {needed}, {remaining} remaining")
            }
            Self::VarIntOverlong => write!(f, "varint is not canonically encoded"),
            Self::VarIntTooLarge => write!(f, "varint exceeds the width of its target type"),
            Self::NegativeLength(value) => write!(f, "negative length {value}"),
            Self::LengthLimitExceeded { field, limit, actual } => {
                write!(f, "field {} declared length {actual} exceeds limit {limit}", field.0)
            }
            Self::LengthNotRepresentable { field, needed, remaining } => {
                write!(f, "field {} declared length needs {needed} bytes, {remaining} remaining", field.0)
            }
            Self::AllocationFailed { requested } => {
                write!(f, "could not allocate {requested} bytes for a decoded value")
            }
            Self::SchemaConstraint(reason) => write!(f, "schema constraint violated: {reason}"),
            Self::UnknownVariant { type_name, value } => {
                write!(f, "unknown {type_name} discriminant {value}")
            }
            Self::InvalidBitset(reason) => write!(f, "invalid bitset: {reason}"),
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
            Self::InvalidBatchHeader(actual) => {
                write!(f, "batch does not start with 0xfe: {actual:?}")
            }
            Self::TooManyFrames { limit } => write!(f, "batch holds more than {limit} packets"),
        }
    }
}

impl std::error::Error for DecodeError {}

pub type DecodeResult<T> = Result<T, DecodeError>;

#[derive(Clone, Debug, PartialEq, Eq)]
pub enum EncodeError {
    /// The output would exceed the writer's configured limit.
    OutputLimitExceeded { limit: usize, needed: usize },
    /// The allocator refused more output space.
    AllocationFailed { requested: usize },
    /// A length does not fit its wire prefix.
    LengthOverflow(usize),
    SchemaConstraint(&'static str),
    /// The local peer may not send this packet id.
    UnexpectedDirection(u32),
    InvalidSubclient { sender: u8, target: u8 },
}

impl fmt::Display for EncodeError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::OutputLimitExceeded { limit, needed } => {
                write!(f, "encoding needs {needed} bytes, limit is {limit}")
            }
            Self::AllocationFailed { requested } => {
                write!(f, "could not allocate {requested} output bytes")
            }
            Self::LengthOverflow(len) => write!(f, "length {len} does not fit its wire prefix"),
            Self::SchemaConstraint(reason) => write!(f, "schema constraint violated: {reason}"),
            Self::UnexpectedDirection(id) => write!(f, "packet id {id} cannot be sent by this peer"),
            Self::InvalidSubclient { sender, target } => {
                write!(f, "subclient ids must be 0..=3, got sender {sender} and target {target}")
            }
        }
    }
}

impl std::error::Error for EncodeError {}

pub type EncodeResult<T> = Result<T, EncodeError>;

/// Decode bounds: a default element ceiling, a default byte ceiling for
/// strings and byte buffers, and per-field overrides of either.
///
/// Every declared length is also bounded by the input remaining, so the
/// defaults only need to be finite when a peer is untrusted.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Limits {
    elements: usize,
    bytes: usize,
    fields: Vec<Option<usize>>,
}

static UNBOUNDED: Limits = Limits::unbounded();

impl Limits {
    /// No ceiling beyond the remaining input.
    pub const fn unbounded() -> Self {
        Self { elements: usize::MAX, bytes: usize::MAX, fields: Vec::new() }
    }

    /// The defaults for decoding packets sent by `sender`: unbounded from a
    /// server, and [`CLIENT_COLLECTION_LIMIT`] elements from a client.
    pub fn receiving_from(sender: Peer) -> Self {
        match sender {
            Peer::Server => Self::unbounded(),
            Peer::Client => Self { elements: CLIENT_COLLECTION_LIMIT, ..Self::unbounded() },
        }
    }

    pub fn with_elements(mut self, limit: usize) -> Self {
        self.elements = limit;
        self
    }

    pub fn with_bytes(mut self, limit: usize) -> Self {
        self.bytes = limit;
        self
    }

    /// Bounds one field site: an element count for collections and maps, a
    /// byte length for strings and byte buffers.
    pub fn with_field(mut self, field: FieldId, limit: usize) -> Self {
        self.set_field(field, limit);
        self
    }

    pub fn set_field(&mut self, field: FieldId, limit: usize) {
        if field == FieldId::NONE {
            return;
        }
        let index = usize::from(field.0);
        if self.fields.len() <= index {
            self.fields.resize(index + 1, None);
        }
        self.fields[index] = Some(limit);
    }

    pub fn field(&self, field: FieldId) -> Option<usize> {
        self.fields.get(usize::from(field.0)).copied().flatten()
    }

    fn elements_for(&self, field: FieldId) -> usize {
        self.field(field).unwrap_or(self.elements)
    }

    fn bytes_for(&self, field: FieldId) -> usize {
        self.field(field).unwrap_or(self.bytes)
    }
}

impl Default for Limits {
    fn default() -> Self {
        Self::unbounded()
    }
}

/// Cursor over a byte slice. Every read is bounds-checked; no read allocates
/// beyond the bytes it returns.
///
/// Constructing with [`Reader::from_shared`] lets strings, byte buffers and
/// NBT decode as refcounted slices of the source instead of copies.
#[derive(Clone, Debug)]
pub struct Reader<'a> {
    buf: &'a [u8],
    pos: usize,
    shared: Option<&'a bytes::Bytes>,
    limits: &'a Limits,
    recursive_depth: usize,
}

impl<'a> Reader<'a> {
    pub fn new(buf: &'a [u8]) -> Self {
        Self {
            buf,
            pos: 0,
            shared: None,
            limits: &UNBOUNDED,
            recursive_depth: 0,
        }
    }

    /// Reads from `source` and shares its allocation with decoded byte fields.
    pub fn from_shared(source: &'a bytes::Bytes) -> Self {
        Self { shared: Some(source), ..Self::new(source) }
    }

    pub fn with_limits(mut self, limits: &'a Limits) -> Self {
        self.limits = limits;
        self
    }

    pub fn limits(&self) -> &'a Limits {
        self.limits
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

    /// Errors unless the input is fully consumed.
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

    /// Default byte ceiling; the NBT scanner bounds its inner lengths by it.
    pub fn byte_buffer_limit(&self) -> usize {
        self.limits.bytes
    }

    /// Checks a declared count against `limit` and the remaining input, where
    /// `min_element_size` is the fewest bytes one element can occupy.
    pub fn checked_count_with(
        &self,
        declared: u64,
        min_element_size: usize,
        limit: usize,
    ) -> DecodeResult<usize> {
        self.checked(FieldId::NONE, declared, min_element_size, limit)
    }

    fn checked(
        &self,
        field: FieldId,
        declared: u64,
        min_element_size: usize,
        limit: usize,
    ) -> DecodeResult<usize> {
        self.checked_bounds(field, declared, min_element_size, 0, u64::MAX, limit)
    }

    /// Checks truncation first, then the schema bounds, then `limit`, so a
    /// short input is never reported as a recoverable bound violation.
    fn checked_bounds(
        &self,
        field: FieldId,
        declared: u64,
        min_element_size: usize,
        minimum: u64,
        maximum: u64,
        limit: usize,
    ) -> DecodeResult<usize> {
        let remaining = self.remaining();
        let needed = declared.saturating_mul(min_element_size as u64);
        if needed > remaining as u64 {
            return Err(DecodeError::LengthNotRepresentable {
                field,
                needed: usize::try_from(needed).unwrap_or(usize::MAX),
                remaining,
            });
        }
        validate_length(declared, minimum, maximum)?;
        if declared > limit as u64 {
            return Err(DecodeError::LengthLimitExceeded { field, limit, actual: declared });
        }
        Ok(declared as usize)
    }

    /// Validates the declared element count of `field` before allocation.
    pub fn checked_count(
        &self,
        field: FieldId,
        declared: u64,
        min_element_size: usize,
    ) -> DecodeResult<usize> {
        self.checked(field, declared, min_element_size, self.limits.elements_for(field))
    }

    /// Applies schema-published count bounds in addition to the field limit.
    pub fn checked_count_limits(
        &self,
        field: FieldId,
        declared: u64,
        min_element_size: usize,
        minimum: u64,
        maximum: u64,
    ) -> DecodeResult<usize> {
        self.checked_bounds(
            field,
            declared,
            min_element_size,
            minimum,
            maximum,
            self.limits.elements_for(field),
        )
    }

    fn checked_byte_len(&self, field: FieldId, declared: u64) -> DecodeResult<usize> {
        self.checked(field, declared, 1, self.limits.bytes_for(field))
    }

    fn checked_byte_bounds(
        &self,
        field: FieldId,
        declared: u64,
        minimum: u64,
        maximum: u64,
    ) -> DecodeResult<usize> {
        self.checked_bounds(field, declared, 1, minimum, maximum, self.limits.bytes_for(field))
    }

    /// Reads the `var_u32`-prefixed string at `field`, keeping its exact bytes.
    pub fn read_str(&mut self, field: FieldId) -> DecodeResult<Str> {
        let declared = u64::from(self.read_var_u32()?);
        let count = self.checked_byte_len(field, declared)?;
        Ok(Str(self.take_shared(count)?))
    }

    /// Reads the `var_u32`-prefixed byte buffer at `field`.
    pub fn read_shared_bytes(&mut self, field: FieldId) -> DecodeResult<bytes::Bytes> {
        let declared = u64::from(self.read_var_u32()?);
        let count = self.checked_byte_len(field, declared)?;
        self.take_shared(count)
    }

    /// Borrows the `var_u32`-prefixed byte buffer at `field` from the input.
    pub fn read_byte_slice(&mut self, field: FieldId) -> DecodeResult<&'a [u8]> {
        let declared = u64::from(self.read_var_u32()?);
        let count = self.checked_byte_len(field, declared)?;
        self.take(count)
    }
}

/// Creates storage for `count` decoded elements, reserving at most
/// [`MAX_EAGER_RESERVE_BYTES`] up front.
pub fn alloc_vec<T>(count: usize) -> DecodeResult<Vec<T>> {
    let size = std::mem::size_of::<T>().max(1);
    let eager = count.min(MAX_EAGER_RESERVE_BYTES / size);
    let mut out = Vec::new();
    out.try_reserve_exact(eager)
        .map_err(|_| DecodeError::AllocationFailed { requested: eager.saturating_mul(size) })?;
    Ok(out)
}

/// Appends a decoded element, growing the storage fallibly.
#[inline]
pub fn push<T>(out: &mut Vec<T>, value: T) -> DecodeResult<()> {
    if out.len() == out.capacity() {
        grow(out)?;
    }
    out.push(value);
    Ok(())
}

#[cold]
#[inline(never)]
fn grow<T>(out: &mut Vec<T>) -> DecodeResult<()> {
    out.try_reserve(1).map_err(|_| DecodeError::AllocationFailed {
        requested: out.len().saturating_add(1).saturating_mul(std::mem::size_of::<T>()),
    })
}

/// Appends encoded bytes to a caller-owned buffer. Every write is fallible:
/// it fails at the configured output limit or when the allocator refuses.
#[derive(Debug)]
pub struct Writer<'a> {
    buf: &'a mut Vec<u8>,
    start: usize,
    limit: usize,
}

impl<'a> Writer<'a> {
    /// Appends to `buf` after its current contents, without an output limit.
    pub fn new(buf: &'a mut Vec<u8>) -> Self {
        Self::with_limit(buf, usize::MAX)
    }

    /// Appends to `buf`, failing once this writer would add more than `limit`
    /// bytes.
    pub fn with_limit(buf: &'a mut Vec<u8>, limit: usize) -> Self {
        let start = buf.len();
        Self { buf, start, limit }
    }

    /// Bytes this writer has appended.
    pub fn written(&self) -> usize {
        self.buf.len() - self.start
    }

    pub fn as_written(&self) -> &[u8] {
        &self.buf[self.start..]
    }

    /// Ensures `additional` more bytes fit the limit and are allocated.
    pub fn reserve(&mut self, additional: usize) -> EncodeResult<()> {
        let needed = self.written().saturating_add(additional);
        if needed > self.limit {
            return Err(EncodeError::OutputLimitExceeded { limit: self.limit, needed });
        }
        self.buf
            .try_reserve(additional)
            .map_err(|_| EncodeError::AllocationFailed { requested: additional })
    }

    pub fn write_u8(&mut self, value: u8) -> EncodeResult<()> {
        self.reserve(1)?;
        self.buf.push(value);
        Ok(())
    }

    pub fn write_all(&mut self, bytes: &[u8]) -> EncodeResult<()> {
        self.reserve(bytes.len())?;
        self.buf.extend_from_slice(bytes);
        Ok(())
    }

    pub fn write_var_uint(&mut self, value: u64) -> EncodeResult<()> {
        let mut scratch = [0u8; 10];
        let len = put_var_uint(&mut scratch, value);
        self.write_all(&scratch[..len])
    }

    pub fn write_var_u32(&mut self, value: u32) -> EncodeResult<()> {
        self.write_var_uint(u64::from(value))
    }

    pub fn write_var_u64(&mut self, value: u64) -> EncodeResult<()> {
        self.write_var_uint(value)
    }

    pub fn write_zigzag_i32(&mut self, value: i32) -> EncodeResult<()> {
        self.write_var_uint(u64::from(((value << 1) ^ (value >> 31)) as u32))
    }

    pub fn write_zigzag_i64(&mut self, value: i64) -> EncodeResult<()> {
        self.write_var_uint(((value << 1) ^ (value >> 63)) as u64)
    }

    /// Writes a length as a `var_u32`, rejecting lengths above `u32::MAX`.
    pub fn write_len_var_u32(&mut self, len: usize) -> EncodeResult<()> {
        let value = u32::try_from(len).map_err(|_| EncodeError::LengthOverflow(len))?;
        self.write_var_u32(value)
    }

    /// Writes a length as a little-endian `u32`, rejecting lengths above `u32::MAX`.
    pub fn write_len_u32le(&mut self, len: usize) -> EncodeResult<()> {
        let value = u32::try_from(len).map_err(|_| EncodeError::LengthOverflow(len))?;
        self.write_all(&value.to_le_bytes())
    }

    pub fn write_str(&mut self, value: &[u8]) -> EncodeResult<()> {
        self.write_len_var_u32(value.len())?;
        self.write_all(value)
    }

    pub fn write_byte_slice(&mut self, value: &[u8]) -> EncodeResult<()> {
        self.write_len_var_u32(value.len())?;
        self.write_all(value)
    }
}

/// Writes `value` as LEB128 into `out` and returns the byte count.
fn put_var_uint(out: &mut [u8], mut value: u64) -> usize {
    let mut len = 0;
    while value >= 0x80 {
        out[len] = (value as u8) | 0x80;
        value >>= 7;
        len += 1;
    }
    out[len] = value as u8;
    len + 1
}

pub trait Encode {
    fn encode(&self, writer: &mut Writer<'_>) -> EncodeResult<()>;

    fn encode_to_vec(&self) -> EncodeResult<Vec<u8>> {
        let mut out = Vec::new();
        encode_into(self, &mut out)?;
        Ok(out)
    }
}

/// Appends `value` to `out`; on error `out` is restored to its prior length.
pub fn encode_into<T: Encode + ?Sized>(value: &T, out: &mut Vec<u8>) -> EncodeResult<()> {
    let start = out.len();
    let result = value.encode(&mut Writer::new(out));
    if result.is_err() {
        out.truncate(start);
    }
    result
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

    /// Decodes all of `input` under `limits`, sharing its allocation with
    /// string, byte and NBT fields.
    fn decode_shared(input: &bytes::Bytes, limits: &Limits) -> DecodeResult<Self> {
        let mut reader = Reader::from_shared(input).with_limits(limits);
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
            fn encode(&self, writer: &mut Writer<'_>) -> EncodeResult<()> {
                writer.write_all(&self.0.$write())
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
            fn encode(&self, writer: &mut Writer<'_>) -> EncodeResult<()> {
                writer.write_all(&self.0.$write())
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
    ($name:ident, $inner:ty, $write:ident, $read:ident) => {
        #[derive(Clone, Copy, Debug, Default, PartialEq, Eq, Hash, PartialOrd, Ord)]
        pub struct $name(pub $inner);

        impl Encode for $name {
            fn encode(&self, writer: &mut Writer<'_>) -> EncodeResult<()> {
                writer.$write(self.0)
            }
        }

        impl Decode for $name {
            fn decode(reader: &mut Reader<'_>) -> DecodeResult<Self> {
                reader.$read().map(Self)
            }
        }
    };
}

var_codec!(VarUInt, u32, write_var_u32, read_var_u32);
var_codec!(VarULong, u64, write_var_u64, read_var_u64);
var_codec!(ZigZag32, i32, write_zigzag_i32, read_zigzag_i32);
var_codec!(ZigZag64, i64, write_zigzag_i64, read_zigzag_i64);

/// Signed non-zigzag varint: the two's-complement value zero-extended to the
/// target width.
#[derive(Clone, Copy, Debug, Default, PartialEq, Eq, Hash, PartialOrd, Ord)]
pub struct VarInt(pub i32);

impl Encode for VarInt {
    fn encode(&self, writer: &mut Writer<'_>) -> EncodeResult<()> {
        writer.write_var_u32(self.0 as u32)
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
    fn encode(&self, writer: &mut Writer<'_>) -> EncodeResult<()> {
        writer.write_var_u64(self.0 as u64)
    }
}

impl Decode for VarLong {
    fn decode(reader: &mut Reader<'_>) -> DecodeResult<Self> {
        reader.read_var_u64().map(|value| Self(value as i64))
    }
}

impl Encode for bool {
    fn encode(&self, writer: &mut Writer<'_>) -> EncodeResult<()> {
        writer.write_u8(u8::from(*self))
    }
}

impl Decode for bool {
    fn decode(reader: &mut Reader<'_>) -> DecodeResult<Self> {
        Ok(reader.read_u8()? != 0)
    }
}

/// A length-prefixed protocol string, kept as the exact bytes received.
///
/// Bedrock strings are byte strings on the wire. UTF-8 is checked when the
/// text is read, so a malformed string never fails the surrounding packet and
/// always re-encodes byte-identically.
#[derive(Clone, Default, PartialEq, Eq, Hash, PartialOrd, Ord)]
pub struct Str(bytes::Bytes);

impl Str {
    pub const fn from_static(value: &'static str) -> Self {
        Self(bytes::Bytes::from_static(value.as_bytes()))
    }

    pub fn as_bytes(&self) -> &[u8] {
        &self.0
    }

    pub fn len(&self) -> usize {
        self.0.len()
    }

    pub fn is_empty(&self) -> bool {
        self.0.is_empty()
    }

    pub fn to_str(&self) -> Result<&str, std::str::Utf8Error> {
        std::str::from_utf8(&self.0)
    }

    pub fn to_string_lossy(&self) -> std::borrow::Cow<'_, str> {
        String::from_utf8_lossy(&self.0)
    }

    pub fn into_bytes(self) -> bytes::Bytes {
        self.0
    }
}

impl fmt::Debug for Str {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        fmt::Debug::fmt(&self.to_string_lossy(), f)
    }
}

impl From<bytes::Bytes> for Str {
    fn from(value: bytes::Bytes) -> Self {
        Self(value)
    }
}

impl From<String> for Str {
    fn from(value: String) -> Self {
        Self(bytes::Bytes::from(value))
    }
}

impl From<&str> for Str {
    fn from(value: &str) -> Self {
        Self(bytes::Bytes::copy_from_slice(value.as_bytes()))
    }
}

impl AsRef<[u8]> for Str {
    fn as_ref(&self) -> &[u8] {
        &self.0
    }
}

impl PartialEq<str> for Str {
    fn eq(&self, other: &str) -> bool {
        self.as_bytes() == other.as_bytes()
    }
}

impl PartialEq<&str> for Str {
    fn eq(&self, other: &&str) -> bool {
        self.as_bytes() == other.as_bytes()
    }
}

impl Encode for Str {
    fn encode(&self, writer: &mut Writer<'_>) -> EncodeResult<()> {
        writer.write_str(self.as_bytes())
    }
}

impl Decode for Str {
    fn decode(reader: &mut Reader<'_>) -> DecodeResult<Self> {
        reader.read_str(FieldId::NONE)
    }
}

impl<T: Encode> Encode for Option<T> {
    fn encode(&self, writer: &mut Writer<'_>) -> EncodeResult<()> {
        match self {
            Some(value) => {
                writer.write_u8(1)?;
                value.encode(writer)
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
    fn encode(&self, writer: &mut Writer<'_>) -> EncodeResult<()> {
        for value in self {
            value.encode(writer)?;
        }
        Ok(())
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
pub fn encode_collection<T: Encode>(writer: &mut Writer<'_>, values: &[T]) -> EncodeResult<()> {
    writer.write_len_var_u32(values.len())?;
    for value in values {
        value.encode(writer)?;
    }
    Ok(())
}

pub fn encode_collection_limits<T: Encode>(
    writer: &mut Writer<'_>,
    values: &[T],
    minimum: u64,
    maximum: u64,
) -> EncodeResult<()> {
    check_length(values.len(), minimum, maximum)?;
    encode_collection(writer, values)
}

fn decode_elements<T: Decode>(reader: &mut Reader<'_>, count: usize) -> DecodeResult<Vec<T>> {
    let mut out = alloc_vec(count)?;
    for _ in 0..count {
        push(&mut out, T::decode(reader)?)?;
    }
    Ok(out)
}

/// Reads the `VarUInt`-prefixed collection at `field`, bounding the declared
/// count before reserving.
pub fn decode_collection<T: Decode>(
    reader: &mut Reader<'_>,
    field: FieldId,
    min_element_size: usize,
) -> DecodeResult<Vec<T>> {
    let declared = u64::from(reader.read_var_u32()?);
    let count = reader.checked_count(field, declared, min_element_size)?;
    decode_elements(reader, count)
}

pub fn decode_collection_limits<T: Decode>(
    reader: &mut Reader<'_>,
    field: FieldId,
    min_element_size: usize,
    minimum: u64,
    maximum: u64,
) -> DecodeResult<Vec<T>> {
    let declared = u64::from(reader.read_var_u32()?);
    let count = reader.checked_count_limits(field, declared, min_element_size, minimum, maximum)?;
    decode_elements(reader, count)
}

/// Rejects an encoded length outside the schema bounds.
pub fn check_length(actual: usize, minimum: u64, maximum: u64) -> EncodeResult<()> {
    validate_length(actual as u64, minimum, maximum)
        .map_err(|_| EncodeError::SchemaConstraint("length outside schema limits"))
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

pub fn encode_string_limits(
    writer: &mut Writer<'_>,
    value: &Str,
    minimum: u64,
    maximum: u64,
) -> EncodeResult<()> {
    check_length(value.len(), minimum, maximum)?;
    writer.write_str(value.as_bytes())
}

pub fn decode_string_limits(
    reader: &mut Reader<'_>,
    field: FieldId,
    minimum: u64,
    maximum: u64,
) -> DecodeResult<Str> {
    let declared = u64::from(reader.read_var_u32()?);
    let count = reader.checked_byte_bounds(field, declared, minimum, maximum)?;
    Ok(Str(reader.take_shared(count)?))
}

pub fn encode_bytes_limits(
    writer: &mut Writer<'_>,
    value: &[u8],
    minimum: u64,
    maximum: u64,
) -> EncodeResult<()> {
    check_length(value.len(), minimum, maximum)?;
    writer.write_byte_slice(value)
}

pub fn decode_bytes_limits(
    reader: &mut Reader<'_>,
    field: FieldId,
    minimum: u64,
    maximum: u64,
) -> DecodeResult<bytes::Bytes> {
    let declared = u64::from(reader.read_var_u32()?);
    let count = reader.checked_byte_bounds(field, declared, minimum, maximum)?;
    reader.take_shared(count)
}

/// Rejects an encoded number outside the schema bounds.
pub fn check_number_limits<T: PartialOrd + Copy>(
    value: T,
    minimum: Option<T>,
    maximum: Option<T>,
) -> EncodeResult<()> {
    validate_number_limits(value, minimum, maximum)
        .map_err(|_| EncodeError::SchemaConstraint("number outside schema limits"))
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

// ---------------------------------------------------------------------------
// Collections, maps, and bitsets
// ---------------------------------------------------------------------------

/// Writes a `u32le`-prefixed collection.
pub fn encode_collection_u32le<T: Encode>(
    writer: &mut Writer<'_>,
    values: &[T],
) -> EncodeResult<()> {
    writer.write_len_u32le(values.len())?;
    for value in values {
        value.encode(writer)?;
    }
    Ok(())
}

/// Reads the `u32le`-prefixed collection at `field` under the same bounds as
/// [`decode_collection`].
pub fn decode_collection_u32le<T: Decode>(
    reader: &mut Reader<'_>,
    field: FieldId,
    min_element_size: usize,
) -> DecodeResult<Vec<T>> {
    let declared = u64::from(u32::from_le_bytes(reader.read_bytes::<4>()?));
    let count = reader.checked_count(field, declared, min_element_size)?;
    decode_elements(reader, count)
}

/// Reads a fixed-width collection count under schema and reader bounds.
pub fn decode_collection_u32le_limits<T: Decode>(
    reader: &mut Reader<'_>,
    field: FieldId,
    min_element_size: usize,
    minimum: u64,
    maximum: u64,
) -> DecodeResult<Vec<T>> {
    let declared = u64::from(u32::from_le_bytes(reader.read_bytes::<4>()?));
    let count = reader.checked_count_limits(field, declared, min_element_size, minimum, maximum)?;
    decode_elements(reader, count)
}

/// Writes a `VarUInt`-prefixed map as ordered key/value pairs.
pub fn encode_map<K: Encode, V: Encode>(
    writer: &mut Writer<'_>,
    entries: &[(K, V)],
) -> EncodeResult<()> {
    writer.write_len_var_u32(entries.len())?;
    for (key, value) in entries {
        key.encode(writer)?;
        value.encode(writer)?;
    }
    Ok(())
}

pub fn encode_map_limits<K: Encode, V: Encode>(
    writer: &mut Writer<'_>,
    entries: &[(K, V)],
    minimum: u64,
    maximum: u64,
) -> EncodeResult<()> {
    check_length(entries.len(), minimum, maximum)?;
    encode_map(writer, entries)
}

fn decode_entries<K: Decode, V: Decode>(
    reader: &mut Reader<'_>,
    count: usize,
) -> DecodeResult<Vec<(K, V)>> {
    let mut out = alloc_vec(count)?;
    for _ in 0..count {
        let key = K::decode(reader)?;
        let value = V::decode(reader)?;
        push(&mut out, (key, value))?;
    }
    Ok(out)
}

/// Reads the `VarUInt`-prefixed map at `field`, preserving wire order.
pub fn decode_map<K: Decode, V: Decode>(
    reader: &mut Reader<'_>,
    field: FieldId,
    min_entry_size: usize,
) -> DecodeResult<Vec<(K, V)>> {
    let declared = u64::from(reader.read_var_u32()?);
    let count = reader.checked_count(field, declared, min_entry_size)?;
    decode_entries(reader, count)
}

pub fn decode_map_limits<K: Decode, V: Decode>(
    reader: &mut Reader<'_>,
    field: FieldId,
    min_entry_size: usize,
    minimum: u64,
    maximum: u64,
) -> DecodeResult<Vec<(K, V)>> {
    let declared = u64::from(reader.read_var_u32()?);
    let count = reader.checked_count_limits(field, declared, min_entry_size, minimum, maximum)?;
    decode_entries(reader, count)
}

/// Writes `bits` bits using seven payload bits per continuation byte.
///
/// Trailing zero groups are omitted, so an all-zero bitset is a single `0`
/// byte. Set bits at or beyond `bits` are not written, as the client drops
/// them on read.
pub fn encode_bitset(writer: &mut Writer<'_>, words: &[u64], bits: u64) -> EncodeResult<()> {
    let masked = |index: usize| match bits.saturating_sub(index as u64 * 64) {
        0 => 0,
        width if width >= 64 => words[index],
        width => words[index] & ((1u64 << width) - 1),
    };
    let Some(last) = (0..words.len()).rev().find_map(|index| {
        let word = masked(index);
        if word == 0 { None } else { Some(index as u64 * 64 + u64::from(63 - word.leading_zeros())) }
    }) else {
        return writer.write_u8(0);
    };
    let groups = last / 7 + 1;
    for group in 0..groups {
        let offset = group * 7;
        let mut value = 0u8;
        for bit in 0..7 {
            let index = offset + bit;
            if index <= last && words[(index / 64) as usize] & (1u64 << (index % 64)) != 0 {
                value |= 1 << bit;
            }
        }
        if group + 1 < groups {
            value |= 0x80;
        }
        writer.write_u8(value)?;
    }
    Ok(())
}

/// Reads a seven-bits-per-byte bitset, dropping payload bits at or beyond
/// `bits` as the client does. Only an encoding still unterminated after
/// `ceil(bits / 7)` bytes is an error.
pub fn decode_bitset<const N: usize>(reader: &mut Reader<'_>, bits: u64) -> DecodeResult<[u64; N]> {
    let mut words = [0u64; N];
    let mut offset = 0u64;
    for _ in 0..bits.div_ceil(7).max(1) {
        let value = reader.read_u8()?;
        for bit in 0..7 {
            let index = offset + bit;
            if index >= bits || value & (1 << bit) == 0 {
                continue;
            }
            if let Some(word) = words.get_mut((index / 64) as usize) {
                *word |= 1u64 << (index % 64);
            }
        }
        if value & 0x80 == 0 {
            return Ok(words);
        }
        offset += 7;
    }
    Err(DecodeError::InvalidBitset("unterminated after its widest encoding"))
}

// ---------------------------------------------------------------------------
// Length-prefixed byte buffers
// ---------------------------------------------------------------------------

impl Encode for bytes::Bytes {
    fn encode(&self, writer: &mut Writer<'_>) -> EncodeResult<()> {
        writer.write_byte_slice(self)
    }
}

impl Decode for bytes::Bytes {
    fn decode(reader: &mut Reader<'_>) -> DecodeResult<Self> {
        reader.read_shared_bytes(FieldId::NONE)
    }
}

/// Writes a fixed-width length and a schema-bounded string.
pub fn encode_string_u32le(writer: &mut Writer<'_>, value: &Str, minimum: u64, maximum: u64) -> EncodeResult<()> {
    encode_bytes_u32le(writer, value.as_bytes(), minimum, maximum)
}

/// Reads a fixed-width string length under byte and schema bounds.
pub fn decode_string_u32le(reader: &mut Reader<'_>, field: FieldId, minimum: u64, maximum: u64) -> DecodeResult<Str> {
    Ok(Str(decode_bytes_u32le(reader, field, minimum, maximum)?))
}

/// Writes a fixed-width length and a schema-bounded byte buffer.
pub fn encode_bytes_u32le(writer: &mut Writer<'_>, value: &[u8], minimum: u64, maximum: u64) -> EncodeResult<()> {
    check_length(value.len(), minimum, maximum)?;
    writer.write_len_u32le(value.len())?;
    writer.write_all(value)
}

/// Reads a fixed-width buffer length under byte and schema bounds.
pub fn decode_bytes_u32le(reader: &mut Reader<'_>, field: FieldId, minimum: u64, maximum: u64) -> DecodeResult<bytes::Bytes> {
    let declared = u64::from(u32::from_le_bytes(reader.read_bytes::<4>()?));
    let count = reader.checked_byte_bounds(field, declared, minimum, maximum)?;
    reader.take_shared(count)
}
