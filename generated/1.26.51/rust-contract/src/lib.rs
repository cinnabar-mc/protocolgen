// Code generated from canonical protocol manifest v2. DO NOT EDIT.

// Variant layouts follow the manifest; the emitter boxes by its own size estimate.
#![allow(clippy::large_enum_variant)]

pub const GAME_VERSION: &str = "1.26.51";
pub const PROTOCOL_VERSION: i32 = 2193;
/// Rust codec contract this crate implements; see protocolgen's
/// `docs/rust-codec-contract.md`.
pub const CODEC_CONTRACT: u32 = 2;
/// SHA-256 of the stable serialization of the manifest this crate was generated from.
pub const MANIFEST_SHA256: &str =
    "724f8bb63843ea0fbc654c647656cc40410724676f5c757b200c0dff97563ab8";

pub mod enums;
pub mod fields;
pub mod packets;
pub mod types;
pub mod wire;
