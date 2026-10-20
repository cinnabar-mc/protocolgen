//! Loads the committed synthetic corpus (`testdata/bench`) for the criterion
//! benches, the allocation report and the encode-from-struct tests.

use std::path::{Path, PathBuf};

use bedrock_protocol_1_26_51::packets::{Decoded, Packet};
use bedrock_protocol_1_26_51::wire::{self, Frames, Limits, Peer, RawFrame};
use bytes::Bytes;
use serde::Deserialize;

/// Bench sets, as `differential/retail/synthetic.go` names them.
pub const SET_BATCH: &str = "batch";

/// One manifest entry of the synthetic corpus.
#[derive(Deserialize)]
pub struct File {
    pub name: String,
    pub file: String,
    pub sender: String,
    pub set: String,
    pub packets: Vec<u32>,
    pub sha256: String,
}

/// A synthetic file with its batch framed and every entry decoded once.
pub struct Case {
    pub file: File,
    pub batch: Bytes,
    pub sender: Peer,
    pub limits: Limits,
    pub frames: Vec<RawFrame>,
    pub packets: Vec<Packet>,
}

impl Case {
    /// The bytes one entry occupies after its length prefix, which is what
    /// the Go benches report as throughput.
    pub fn entry_len(&self, index: usize) -> usize {
        let frame = self.frames[index].frame();
        let mut reader = wire::Reader::new(frame);
        reader.read_var_u32().expect("length prefix");
        frame.len() - reader.position()
    }
}

/// The committed corpus directory.
pub fn corpus_dir() -> PathBuf {
    Path::new(env!("CARGO_MANIFEST_DIR")).join("../../../testdata/bench")
}

/// Reads and decodes every synthetic file; panics on any file the contract
/// crate cannot decode, since the retail run already requires that.
pub fn load() -> Vec<Case> {
    let dir = corpus_dir();
    let manifest =
        std::fs::read(dir.join("manifest.json")).expect("read testdata/bench/manifest.json");
    let files: Vec<File> = serde_json::from_slice(&manifest).expect("parse manifest.json");
    files
        .into_iter()
        .map(|file| {
            let batch =
                Bytes::from(std::fs::read(dir.join(&file.file)).expect("read synthetic file"));
            let sender = match file.sender.as_str() {
                "server" => Peer::Server,
                _ => Peer::Client,
            };
            let limits = Limits::receiving_from(sender);
            let frames: Vec<RawFrame> = Frames::from_batch(batch.clone(), usize::MAX)
                .expect("frame batch")
                .map(|frame| frame.expect("frame entry"))
                .collect();
            let packets = frames
                .iter()
                .map(|frame| {
                    let decoded = Packet::decode_frame(frame, sender, &limits);
                    assert_eq!(
                        Packet::decode_framed(frame.frame(), sender, &limits),
                        decoded,
                        "{}: borrowed and owned frames differ",
                        file.file
                    );
                    match decoded {
                        Ok(Decoded::Known(packet)) => packet,
                        Ok(Decoded::Unknown(_)) => panic!("{}: unknown packet", file.file),
                        Err(error) => panic!("{}: {error}", file.file),
                    }
                })
                .collect();
            Case {
                file,
                batch,
                sender,
                limits,
                frames,
                packets,
            }
        })
        .collect()
}
