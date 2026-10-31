
// ---------------------------------------------------------------------------
// Runtime tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod runtime_tests {
    use super::*;

    /// Empty values consume no payload bytes but still obey the element limit.
    /// Fixed-width counts enforce the same schema bounds as varints.
    #[test]
    fn fixed_width_collections_reject_excess_items() {
        let input = [2, 0, 0, 0, 5, 0, 6, 0];
        assert!(decode_collection_u32le_limits::<U16LE>(&mut Reader::new(&input), 2, 0, 1).is_err());
        let valid = [1, 0, 0, 0, 5, 0];
        assert_eq!(decode_collection_u32le_limits::<U16LE>(&mut Reader::new(&valid), 2, 0, 1).unwrap(), vec![U16LE(5)]);
    }

    #[test]
    fn zero_byte_collections_obey_element_limits() {
        let mut reader = Reader::new(&[]);
        assert_eq!(reader.checked_count(3, 0), Ok(3));
        reader.set_collection_limit(2);
        assert!(matches!(reader.checked_count(3, 0), Err(DecodeError::LengthLimitExceeded { .. })));
    }

    /// Fixed-width strings and buffers preserve their declared prefix bytes.
    #[test]
    fn fixed_width_buffers_round_trip() {
        let mut writer = Writer::new();
        encode_string_u32le(&mut writer, "hi", 0, 2);
        assert_eq!(writer.as_slice(), [2, 0, 0, 0, b'h', b'i']);
        assert_eq!(decode_string_u32le(&mut Reader::new(writer.as_slice()), 0, 2).unwrap(), "hi");
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
            Err(DecodeError::LengthNotRepresentable { needed: 800, remaining: 4 })
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
        for words in [[0u64, 0, 0], [1, 0, 0], [0, 0, 1 << 2], [u64::MAX, u64::MAX, 0b111]] {
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
