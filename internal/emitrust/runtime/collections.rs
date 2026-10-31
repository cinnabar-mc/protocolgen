
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

/// Reads a fixed-width collection count under schema and reader bounds.
pub fn decode_collection_u32le_limits<T: Decode>(
    reader: &mut Reader<'_>,
    min_element_size: usize,
    minimum: u64,
    maximum: u64,
) -> DecodeResult<Vec<T>> {
    let declared = u64::from(u32::from_le_bytes(reader.read_bytes::<4>()?));
    let count = reader.checked_count_limits(declared, min_element_size, minimum, maximum)?;
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

pub fn encode_map_limits<K: Encode, V: Encode>(writer: &mut Writer, entries: &[(K, V)], minimum: u64, maximum: u64) {
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
            return Err(DecodeError::NbtMalformed("bitset bits outside declared width"));
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
