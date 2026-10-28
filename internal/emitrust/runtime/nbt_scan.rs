
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

