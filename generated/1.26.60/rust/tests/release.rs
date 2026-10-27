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
    assert_eq!(decoded.encode_to_vec(), bytes);
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
        assert_eq!(decoded.encode_to_vec(), bytes);
    }
}
