//! The generated codec-contract crate: frames with `wire::Frames`, decodes with
//! `Packet::decode_frame` under the sender's default limits, and re-encodes
//! with `encode_frame`. Retail inputs also run the element-limit probe, large
//! inputs the allocation-failure probe and unknown-id inputs the zero-copy
//! probe.

use bedrock_protocol_1_26_51::packets::{Decoded, Packet};
use bedrock_protocol_1_26_51::wire::{
    self, DecodeError, Frames, Limits, PacketHeader, Peer, RawFrame,
};
use bytes::Bytes;

use crate::alloc;
use crate::verdict::{DIFFERS, Input, OK, SAME, Sender, UNKNOWN, Verdict, digest, guarded};

/// The allocation-failure probe refuses allocations above this: the large
/// input's collection reserves more, and packet boxes stay well below it.
const ALLOC_PROBE_BYTES: usize = 64 * 1024;

pub fn run(input: &Input, lines: &mut Vec<Verdict>) {
    let sender = match input.sender {
        Sender::Server => Peer::Server,
        Sender::Client => Peer::Client,
    };
    let limits = Limits::receiving_from(sender);
    each_frame(input, &input.case, lines, |frame, verdict| {
        decode(frame, sender, &limits, verdict)
    });
    match input.case.as_str() {
        "retail" => {
            let zero = Limits::unbounded().with_elements(0);
            each_frame(input, "element-limit", lines, |frame, verdict| {
                decode(frame, sender, &zero, verdict)
            });
        }
        "large" => each_frame(input, "alloc-failure", lines, |frame, verdict| {
            let result = alloc::refusing_above(ALLOC_PROBE_BYTES, || {
                Packet::decode_frame(frame, sender, &limits).map(drop)
            });
            match result {
                Ok(()) => verdict.outcome = OK,
                Err(error) => verdict.fail(class(&error), &error),
            }
        }),
        "unknown-id" => each_frame(input, "zero-copy", lines, |frame, verdict| {
            zero_copy(&input.batch, frame, sender, verdict)
        }),
        _ => {}
    }
}

fn each_frame(
    input: &Input,
    case: &str,
    lines: &mut Vec<Verdict>,
    mut decode: impl FnMut(&RawFrame, &mut Verdict),
) {
    let frames = match Frames::from_batch(input.batch.clone(), usize::MAX) {
        Ok(frames) => frames,
        Err(error) => return lines.push(Verdict::frame_error(input, case, 0, error)),
    };
    for (entry, frame) in frames.enumerate() {
        let frame = match frame {
            Ok(frame) => frame,
            Err(error) => return lines.push(Verdict::frame_error(input, case, entry, error)),
        };
        let mut verdict = Verdict::new(input, case, entry);
        verdict.header = Some(header_text(frame.header()));
        verdict.payload = Some(digest(after_length(frame.frame())));
        guarded(&mut verdict, |verdict| decode(&frame, verdict));
        lines.push(verdict);
    }
}

fn decode(frame: &RawFrame, sender: Peer, limits: &Limits, verdict: &mut Verdict) {
    match Packet::decode_frame(frame, sender, limits) {
        Ok(Decoded::Known(packet)) => {
            verdict.outcome = OK;
            let fields = format!("{packet:?}");
            verdict.fields = Some(digest(fields.as_bytes()));
            verdict.debug = Some(fields.clone());
            let header = frame.header();
            let mut out = Vec::new();
            if let Err(error) = packet.encode_frame(
                sender,
                header.sender_subclient,
                header.target_subclient,
                &mut out,
            ) {
                verdict.detail = Some(format!("encode: {error}"));
                return;
            }
            let Some(Ok(reencoded)) = Frames::new(Bytes::from(out), 1).next() else {
                verdict.detail = Some("the re-encoded entry does not frame".to_string());
                return;
            };
            verdict.encoded = Some(digest(after_length(reencoded.frame())));
            verdict.roundtrip = Some(match Packet::decode_frame(&reencoded, sender, limits) {
                Ok(Decoded::Known(again)) if format!("{again:?}") == fields => SAME,
                _ => DIFFERS,
            });
        }
        Ok(Decoded::Unknown(unknown)) => {
            verdict.outcome = UNKNOWN;
            let mut out = Vec::new();
            match unknown.encode_frame(&mut out) {
                Ok(()) => verdict.encoded = Some(digest(after_length(&out))),
                Err(error) => verdict.detail = Some(format!("forward: {error}")),
            }
        }
        Err(error) => verdict.fail(class(&error), &error),
    }
}

/// Frame and unknown bodies must be slices of the batch, and a compacted
/// frame must not be.
fn zero_copy(batch: &Bytes, frame: &RawFrame, sender: Peer, verdict: &mut Verdict) {
    let mut failures = Vec::new();
    if !shares(batch, &frame.body()) {
        failures.push("the frame body copies the batch");
    }
    if shares(batch, &frame.clone().compact().body()) {
        failures.push("the compacted frame still shares the batch");
    }
    match Packet::decode_frame(frame, sender, &Limits::receiving_from(sender)) {
        Ok(Decoded::Unknown(unknown)) if !shares(batch, &unknown.body) => {
            failures.push("the unknown body copies the batch")
        }
        Ok(Decoded::Unknown(_)) => {}
        Ok(Decoded::Known(_)) => failures.push("the entry decoded as a known packet"),
        Err(_) => failures.push("the entry did not decode"),
    }
    if failures.is_empty() {
        verdict.outcome = OK;
    } else {
        verdict.fail("zero-copy", failures.join("; "));
    }
}

/// Whether a non-empty `part` lies inside `batch`'s allocation.
fn shares(batch: &[u8], part: &[u8]) -> bool {
    let start = batch.as_ptr() as usize;
    let at = part.as_ptr() as usize;
    part.is_empty() || (at >= start && at + part.len() <= start + batch.len())
}

fn after_length(entry: &[u8]) -> &[u8] {
    let mut reader = wire::Reader::new(entry);
    match reader.read_var_u32() {
        Ok(_) => &entry[reader.position()..],
        Err(_) => entry,
    }
}

fn header_text(header: PacketHeader) -> String {
    format!(
        "{}:{}:{}",
        header.id, header.sender_subclient, header.target_subclient
    )
}

fn class(error: &DecodeError) -> &'static str {
    match error {
        DecodeError::UnexpectedEof { .. } | DecodeError::LengthNotRepresentable { .. } => "eof",
        DecodeError::VarIntOverlong | DecodeError::VarIntTooLarge => "varint",
        DecodeError::NegativeLength(_) => "length",
        DecodeError::LengthLimitExceeded { .. } | DecodeError::RecursiveDepthExceeded => "limit",
        DecodeError::AllocationFailed { .. } => "alloc",
        DecodeError::SchemaConstraint(_) => "schema",
        DecodeError::UnknownVariant { .. } => "variant",
        DecodeError::InvalidBitset(_) => "bitset",
        DecodeError::NbtDepthExceeded | DecodeError::NbtMalformed(_) => "nbt",
        DecodeError::UnknownPacketId(_) => "unknown-id",
        DecodeError::UnexpectedDirection(_) => "direction",
        DecodeError::TrailingBytes(_) => "trailing",
        DecodeError::InvalidBatchHeader(_) | DecodeError::TooManyFrames { .. } => "frame",
    }
}
