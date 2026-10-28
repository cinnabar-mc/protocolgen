// Code generated from canonical protocol manifest v2. DO NOT EDIT.

// Variant layouts follow the manifest; the emitter boxes by its own size estimate.
#![allow(clippy::large_enum_variant)]

pub const GAME_VERSION: &str = "1.26.60";
pub const PROTOCOL_VERSION: i32 = 2225;
/// Rust codec contract this crate implements; see protocolgen's
/// `docs/rust-codec-contract.md`.
pub const CODEC_CONTRACT: u32 = 2;
/// SHA-256 of the stable serialization of the manifest this crate was generated from.
pub const MANIFEST_SHA256: &str =
    "3f790cd30b299a515fffe591e8da3823893e86d9b04481ed091b2bab5e07d7a9";

pub mod enums;
pub mod fields;
pub mod packets;
pub mod types;
pub mod wire;
