// ---------------------------------------------------------------------------
// Runtime tests
// ---------------------------------------------------------------------------

#[cfg(test)]
mod runtime_tests {
    use super::*;

    fn frame(id: u32, body: &[u8]) -> Vec<u8> {
        let mut out = Vec::new();
        encode_frame_with(&mut out, PacketHeader::new(id), usize::MAX, |writer| writer.write_all(body))
            .expect("frame encodes");
        out
    }

    /// Empty values consume no payload bytes but still obey the element limit.
    #[test]
    fn zero_byte_collections_obey_element_limits() {
        assert_eq!(Reader::new(&[]).checked_count(FieldId(0), 3, 0), Ok(3));
        let limits = Limits::unbounded().with_field(FieldId(0), 2);
        assert!(matches!(Reader::new(&[]).with_limits(&limits).checked_count(FieldId(0), 3, 0), Err(DecodeError::LengthLimitExceeded { .. })));
    }

    /// Fixed-width strings and buffers preserve their declared prefix bytes.
    #[test]
    fn fixed_width_buffers_round_trip() {
        let mut out = Vec::new();
        encode_string_u32le(&mut Writer::new(&mut out), &Str::from("hi"), 0, 2).unwrap();
        assert_eq!(out, [2, 0, 0, 0, b'h', b'i']);
        assert_eq!(decode_string_u32le(&mut Reader::new(&out), FieldId(0), 0, 2).unwrap(), "hi");
        assert!(decode_bytes_u32le(&mut Reader::new(&out), FieldId(0), 0, 1).is_err());
    }

    #[test]
    fn varint_rejects_overlong_encodings() {
        assert_eq!(Reader::new(&[0x80, 0x00]).read_var_u32(), Err(DecodeError::VarIntOverlong));
        assert_eq!(Reader::new(&[0x00]).read_var_u32(), Ok(0));
    }

    #[test]
    fn varint_round_trips_boundary_values() {
        for value in [0u64, 1, 127, 128, u32::MAX as u64, u64::MAX] {
            let mut out = Vec::new();
            Writer::new(&mut out).write_var_u64(value).unwrap();
            assert_eq!(Reader::new(&out).read_var_u64(), Ok(value), "varint {value}");
        }
    }

    /// A declared count the input cannot hold fails before any allocation.
    #[test]
    fn impossible_counts_fail_before_allocating() {
        let mut input = Vec::new();
        Writer::new(&mut input).write_var_u32(u32::MAX).unwrap();
        let result = decode_collection::<U64LE>(&mut Reader::new(&input), FieldId(0), 8);
        assert!(matches!(result, Err(DecodeError::LengthNotRepresentable { .. })));
    }

    /// A truncated count reports truncation, not a schema or limit violation.
    #[test]
    fn truncation_is_reported_before_bounds() {
        let mut input = Vec::new();
        Writer::new(&mut input).write_var_u32(1000).unwrap();
        let limits = Limits::unbounded().with_field(FieldId(0), 4);
        let truncated = decode_collection_limits::<U8>(
            &mut Reader::new(&input).with_limits(&limits),
            FieldId(0),
            1,
            0,
            4,
        );
        assert!(matches!(truncated, Err(DecodeError::LengthNotRepresentable { .. })));
        let text = decode_string_limits(&mut Reader::new(&input).with_limits(&limits), FieldId(0), 0, 4);
        assert!(matches!(text, Err(DecodeError::LengthNotRepresentable { .. })));

        input.resize(input.len() + 1000, 0);
        let schema = decode_collection_limits::<U8>(
            &mut Reader::new(&input).with_limits(&limits),
            FieldId(0),
            1,
            0,
            4,
        );
        assert_eq!(schema, Err(DecodeError::SchemaConstraint("length above maximum")));
        let limited = decode_collection_limits::<U8>(
            &mut Reader::new(&input).with_limits(&limits),
            FieldId(0),
            1,
            0,
            2000,
        );
        assert!(matches!(limited, Err(DecodeError::LengthLimitExceeded { .. })));
    }

    /// Bits at or beyond the declared width are dropped in both directions.
    #[test]
    fn bitsets_mask_bits_beyond_their_width() {
        for words in [[1u64 << 10, 0], [0, 1], [0, 1 << 63], [u64::MAX, u64::MAX]] {
            let mut out = Vec::new();
            encode_bitset(&mut Writer::new(&mut out), &words, 10).unwrap();
            let masked = words[0] & ((1 << 10) - 1);
            assert_eq!(decode_bitset::<1>(&mut Reader::new(&out), 10), Ok([masked]));
        }
        let mut out = Vec::new();
        encode_bitset(&mut Writer::new(&mut out), &[1 << 9 | 1, 0], 10).unwrap();
        assert_eq!(out, vec![0x81, 0x04]);

        // 0x7c sets bits 9..=13 of a 10-bit set; only bit 9 survives.
        assert_eq!(decode_bitset::<1>(&mut Reader::new(&[0x81, 0x7c]), 10), Ok([1 << 9 | 1]));
        assert!(decode_bitset::<1>(&mut Reader::new(&[0x80, 0x80, 0x00]), 10).is_err());
        let mut widest = vec![0xff; 18];
        widest.push(0x7f);
        assert_eq!(decode_bitset::<3>(&mut Reader::new(&widest), 131), Ok([u64::MAX, u64::MAX, 7]));
        widest[18] = 0xff;
        widest.push(0x00);
        assert_eq!(
            decode_bitset::<3>(&mut Reader::new(&widest), 131),
            Err(DecodeError::InvalidBitset("unterminated after its widest encoding"))
        );
    }

    #[test]
    fn field_limits_override_the_default() {
        let mut input = Vec::new();
        Writer::new(&mut input).write_str(b"hello").unwrap();
        let limits = Limits::unbounded().with_field(FieldId(3), 4);
        let mut reader = Reader::new(&input).with_limits(&limits);
        assert_eq!(
            reader.read_str(FieldId(3)),
            Err(DecodeError::LengthLimitExceeded { field: FieldId(3), limit: 4, actual: 5 })
        );
        let mut reader = Reader::new(&input).with_limits(&limits);
        assert_eq!(reader.read_str(FieldId(4)).unwrap(), "hello");
    }

    #[test]
    fn client_defaults_bound_collections_and_server_defaults_do_not() {
        let mut input = Vec::new();
        Writer::new(&mut input).write_var_u32(5000).unwrap();
        input.resize(input.len() + 5000, 0);
        let client = Limits::receiving_from(Peer::Client);
        let result = decode_collection::<U8>(&mut Reader::new(&input).with_limits(&client), FieldId(0), 1);
        assert!(matches!(result, Err(DecodeError::LengthLimitExceeded { .. })));
        let server = Limits::receiving_from(Peer::Server);
        let values = decode_collection::<U8>(&mut Reader::new(&input).with_limits(&server), FieldId(0), 1);
        assert_eq!(values.map(|values| values.len()), Ok(5000));
    }

    /// Strings keep their exact bytes, so non-UTF-8 text survives a round trip.
    #[test]
    fn strings_preserve_invalid_utf8() {
        let raw = bytes::Bytes::from_static(&[0x02, 0xff, 0xfe]);
        let value = Str::decode_shared(&raw, &Limits::unbounded()).unwrap();
        assert!(value.to_str().is_err());
        assert_eq!(value.encode_to_vec().unwrap(), raw.as_ref());
    }

    #[test]
    fn shared_strings_do_not_copy() {
        let raw = bytes::Bytes::from(vec![0x03, b'a', b'b', b'c']);
        let value = Str::decode_shared(&raw, &Limits::unbounded()).unwrap();
        assert_eq!(value.as_bytes().as_ptr(), raw[1..].as_ptr());
    }

    struct FailsMidway;

    impl Encode for FailsMidway {
        fn encode(&self, writer: &mut Writer<'_>) -> EncodeResult<()> {
            writer.write_all(&[1, 2])?;
            Err(EncodeError::SchemaConstraint("fails midway"))
        }
    }

    #[test]
    fn writer_limit_fails_and_encode_into_restores_the_buffer() {
        let mut out = vec![0xaa];
        let mut writer = Writer::with_limit(&mut out, 2);
        assert!(writer.write_all(&[1, 2, 3]).is_err());

        let mut out = vec![9];
        assert!(encode_into(&FailsMidway, &mut out).is_err());
        assert_eq!(out, vec![9]);
    }

    #[test]
    fn frames_keep_exact_boundaries() {
        let mut batch = vec![BATCH_HEADER];
        batch.extend(frame(1000, &[1, 2, 3]));
        batch.extend(frame(9, &[]));
        let frames: Vec<_> = Frames::from_batch(batch.into(), 8).unwrap().collect::<Result<_, _>>().unwrap();
        assert_eq!(frames.len(), 2);
        assert_eq!(frames[0].id(), 1000);
        assert_eq!(frames[0].body().as_ref(), &[1, 2, 3]);
        assert_eq!(frames[1].id(), 9);
        assert!(frames[1].body().is_empty());
    }

    /// A body reader borrows its original frame and cannot consume the next entry.
    #[test]
    fn frame_body_reader_borrows_only_its_body() {
        let mut batch = vec![BATCH_HEADER];
        batch.extend(frame(1000, &[1, 2, 3]));
        batch.extend(frame(9, &[4, 5]));
        let batch = bytes::Bytes::from(batch);
        let frame = Frames::from_batch(batch.clone(), 8).unwrap().next().unwrap().unwrap();
        let mut reader = frame.body_reader();
        assert!(std::ptr::eq(reader.shared.unwrap(), frame.frame()));
        assert_eq!(reader.position(), 0);
        assert_eq!(reader.remaining(), 3);
        let body = reader.take_shared(3).unwrap();
        assert_eq!(body.as_ptr(), batch[4..].as_ptr());
        assert_eq!(&body[..], &[1, 2, 3]);
        assert!(matches!(reader.read_u8(), Err(DecodeError::UnexpectedEof { .. })));
        drop(frame);
        drop(batch);
        assert_eq!(&body[..], &[1, 2, 3]);
    }

    #[test]
    fn frame_encoding_moves_short_prefixes() {
        let body = vec![7u8; 300];
        let encoded = frame(1, &body);
        // 301 bytes of header and body need a two-byte prefix.
        assert_eq!(&encoded[..3], &[0xad, 0x02, 0x01]);
        assert_eq!(encoded.len(), 2 + 1 + 300);
    }

    /// Capacity guesses may grow or shrink the prefix without changing bytes or limits.
    #[test]
    fn frame_encoding_is_independent_of_reused_capacity() {
        for capacity in [0, 1, 127, 128, 16_383, 16_384] {
            for body_len in [0, 125, 126, 127, 300, 16_382, 16_383] {
                let body = vec![7; body_len];
                let mut expected = vec![0xaa];
                let mut writer = Writer::new(&mut expected);
                writer.write_var_u32(body_len as u32 + 2).unwrap();
                writer.write_var_u32(1000).unwrap();
                writer.write_all(&body).unwrap();

                let mut out = Vec::with_capacity(capacity);
                out.push(0xaa);
                encode_frame_with(&mut out, PacketHeader::new(1000), body_len + 2, |writer| writer.write_all(&body)).unwrap();
                assert_eq!(out, expected, "capacity {capacity}, body {body_len}");
                assert!(encode_frame_with(&mut out, PacketHeader::new(1000), body_len + 1, |writer| writer.write_all(&body)).is_err());
                assert_eq!(out, expected, "limit failure changed prior output");
            }
        }
    }

    #[test]
    fn truncated_frames_and_frame_floods_fail() {
        let truncated = bytes::Bytes::from_static(&[0x05, 0x01]);
        let mut frames = Frames::new(truncated, 8);
        assert!(frames.next().unwrap().is_err());
        assert!(frames.next().is_none());

        let mut flood = Vec::new();
        for _ in 0..3 {
            flood.extend(frame(1, &[]));
        }
        let results: Vec<_> = Frames::new(flood.into(), 2).collect();
        assert_eq!(results.len(), 3);
        assert_eq!(results[2], Err(DecodeError::TooManyFrames { limit: 2 }));
    }

    #[test]
    fn subclients_out_of_range_are_rejected() {
        let header = PacketHeader { id: 1, sender_subclient: 4, target_subclient: 0 };
        let mut out = vec![9];
        assert!(encode_frame_with(&mut out, header, usize::MAX, |_| Ok(())).is_err());
        assert_eq!(out, vec![9]);
    }
}
