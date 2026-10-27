use bedrock_protocol_1_26_60::packets::ClientboundUpdateSoundData;
use bedrock_protocol_1_26_60::types::{SerializedSkinImage, SoundDataEvent};
use bedrock_protocol_1_26_60::wire::{Decode, Encode};

/// Ordinary RGBA images are byte buffers, independent of the collection element limit.
#[test]
fn ordinary_skin_image_is_a_byte_buffer() {
    let mut bytes = vec![64, 0, 0, 0, 64, 0, 0, 0, 0x80, 0x80, 1];
    bytes.extend(vec![0xff; 64 * 64 * 4]);
    let decoded = SerializedSkinImage::decode_exact(&bytes).unwrap();
    assert_eq!(decoded.width.0, 64);
    assert_eq!(decoded.height.0, 64);
    assert_eq!(decoded.image_bytes.len(), 64 * 64 * 4);
    assert_eq!(decoded.encode_to_vec().unwrap(), bytes);
}

/// A stop or volume update consumes one selector and only its selected payload.
#[test]
fn sound_update_uses_one_selector() {
    let stop = [8, 7, 6, 5, 4, 3, 2, 1, 0];
    let volume = [8, 7, 6, 5, 4, 3, 2, 1, 1, 0, 0, 0, 0x3f];
    for (bytes, tag) in [(&stop[..], 0), (&volume[..], 1)] {
        let decoded = ClientboundUpdateSoundData::decode_exact(bytes).unwrap();
        assert_eq!(decoded.server_sound_handle.0, 0x0102030405060708);
        match &decoded.event {
            SoundDataEvent::Stop => assert_eq!(tag, 0),
            SoundDataEvent::SetVolume { volume } => {
                assert_eq!(tag, 1);
                assert_eq!(volume.0, 0.5);
            }
            other => panic!("unexpected sound event: {other:?}"),
        }
        assert_eq!(decoded.encode_to_vec().unwrap(), bytes);
    }
}

/// Generic level events carry a rootless compound body, retaining its end tag.
#[test]
fn generic_event_compound_body() {
    use bedrock_protocol_1_26_60::packets::LevelEventGeneric;
    for bytes in [&[2, 0][..], &[2, 3, 1, b'x', 2, 0]] {
        let decoded = LevelEventGeneric::decode_exact(bytes).unwrap();
        assert_eq!(decoded.ctd.0.as_ref(), &bytes[1..]);
        assert_eq!(decoded.encode_to_vec().unwrap(), bytes);
        assert!(LevelEventGeneric::decode_exact(&bytes[..bytes.len() - 1]).is_err());
    }
}

/// A structure response always has a presence marker, even when no NBT follows.
#[test]
fn structure_response_presence() {
    use bedrock_protocol_1_26_60::packets::StructureTemplateDataResponse;
    for bytes in [&[1, b'a', 0, 1][..], &[1, b'a', 1, 10, 0, 0, 1]] {
        let decoded = StructureTemplateDataResponse::decode_exact(bytes).unwrap();
        assert_eq!(decoded.structure_nbt.is_some(), bytes[2] == 1);
        assert_eq!(decoded.encode_to_vec().unwrap(), bytes);
    }
}

/// Stop is byte zero; start is byte one with a fixed frame rate and file prefix.
#[test]
fn video_capture_action_values() {
    use bedrock_protocol_1_26_60::packets::PlayerVideoCapture;
    use bedrock_protocol_1_26_60::types::PlayerVideoCaptureData;
    for bytes in [&[0][..], &[1, 30, 0, 0, 0, 1, b'a']] {
        let decoded = PlayerVideoCapture::decode_exact(bytes).unwrap();
        assert_eq!(
            matches!(
                decoded.action,
                PlayerVideoCaptureData::StartVideoCapture { .. }
            ),
            bytes[0] == 1
        );
        assert_eq!(decoded.encode_to_vec().unwrap(), bytes);
    }
}

/// Spline easing strings immediately follow the key-frame floats.
#[test]
fn spline_easing_is_required() {
    use bedrock_protocol_1_26_60::types::{
        CameraSplineProgressKeyFrame, CameraSplineRotationKeyFrame,
    };
    let mut progress = vec![0; 8];
    progress.extend_from_slice(b"\x06linear");
    let decoded = CameraSplineProgressKeyFrame::decode_exact(&progress).unwrap();
    assert_eq!(decoded.easing, "linear");
    assert_eq!(decoded.encode_to_vec().unwrap(), progress);
    let mut rotation = vec![0; 16];
    rotation.extend_from_slice(b"\x06linear");
    let decoded = CameraSplineRotationKeyFrame::decode_exact(&rotation).unwrap();
    assert_eq!(decoded.easing, "linear");
    assert_eq!(decoded.encode_to_vec().unwrap(), rotation);
}

/// Each optional height map has sixteen length-prefixed rows of sixteen heights.
#[test]
fn height_map_rows() {
    use bedrock_protocol_1_26_60::types::HeightmapData;
    let mut bytes = vec![1, 1];
    for _ in 0..16 {
        bytes.push(16);
        bytes.extend_from_slice(&[0; 16]);
    }
    bytes.extend_from_slice(&[0, 0]);
    let decoded = HeightmapData::decode_exact(&bytes).unwrap();
    assert_eq!(decoded.encode_to_vec().unwrap(), bytes);
    bytes[2] = 15;
    assert!(HeightmapData::decode_exact(&bytes).is_err());
}
