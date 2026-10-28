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
            fn encode(&self, writer: &mut Writer<'_>) -> EncodeResult<()> {
                writer.write_all(&self.0)
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
        for body in [&[0][..], &[3, 1, b'x', 2, 0], &[10, 1, b'n', 3, 1, b'x', 2, 0, 0]] {
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
