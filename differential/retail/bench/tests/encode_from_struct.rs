//! Packets built as Rust structs must encode exactly as gophertunnel encodes
//! the same values built as its structs; `differential/retail/encode.go` holds
//! the Go side and writes the expected entries to `testdata/encode`.

use std::path::Path;

use bedrock_protocol_1_26_51::enums::{
    PlayerPositionModeComponentPositionMode, SoftEnumUpdateType,
};
use bedrock_protocol_1_26_51::packets::{
    LevelChunk, MovePlayer, NetworkStackLatency, Packet, SetActorMotion, UpdateBlock,
    UpdateSoftEnum,
};
use bedrock_protocol_1_26_51::types::{
    ActorRuntimeID, BlockPos, ChunkPos, DimensionType, MovePlayerTeleportData, PlayerInputTick,
};
use bedrock_protocol_1_26_51::wire::{self, Peer};

fn cases() -> Vec<(&'static str, Peer, Packet)> {
    let chunk: Vec<u8> = (0..1024u32).map(|i| (i * 7) as u8).collect();
    vec![
        (
            "move_player",
            Peer::Server,
            Packet::from(MovePlayer {
                player_runtime_id: ActorRuntimeID(42),
                position: glam::Vec3::new(1.25, 64.0, -2.5),
                rotation: glam::Vec2::new(10.5, 20.25),
                y_head_rotation: wire::F32LE(30.75),
                position_mode: PlayerPositionModeComponentPositionMode::Teleport,
                on_ground: true,
                riding_runtime_id: ActorRuntimeID(0),
                teleport_data: Some(MovePlayerTeleportData {
                    teleportation_cause: wire::I32LE(3),
                    source_actor_type: wire::I32LE(87),
                }),
                tick: PlayerInputTick(1234),
            }),
        ),
        (
            "set_actor_motion",
            Peer::Server,
            Packet::from(SetActorMotion {
                target_runtime_id: ActorRuntimeID(9001),
                motion: glam::Vec3::new(0.125, -0.0784, 3.5),
                tick: PlayerInputTick(77),
            }),
        ),
        (
            "update_block",
            Peer::Server,
            Packet::from(UpdateBlock {
                block_position: BlockPos {
                    x: wire::ZigZag32(-30_000_000),
                    y: wire::ZigZag32(-64),
                    z: wire::ZigZag32(29_999_999),
                },
                block_runtime_id: wire::VarUInt(0xdead_beef),
                flags: wire::VarUInt(3),
                layer: wire::VarUInt(1),
            }),
        ),
        (
            "network_stack_latency",
            Peer::Server,
            Packet::from(NetworkStackLatency {
                creation_time: wire::U64LE(1_760_000_000_123),
                is_from_server: true,
            }),
        ),
        (
            "update_soft_enum",
            Peer::Server,
            Packet::from(UpdateSoftEnum {
                enum_name: wire::Str::from("Players"),
                values: vec![
                    wire::Str::from("Steve"),
                    wire::Str::from("Alex"),
                    wire::Str::from("§cZuri"),
                ],
                update_type: SoftEnumUpdateType::Replace,
            }),
        ),
        (
            "level_chunk",
            Peer::Server,
            Packet::from(LevelChunk {
                chunk_position: ChunkPos {
                    x: wire::ZigZag32(-3),
                    z: wire::ZigZag32(12),
                },
                dimension_id: DimensionType(1),
                sub_chunks_count: wire::VarUInt(4),
                client_request_sub_chunk_limit: None,
                cache_enabled: false,
                cache_metadata: Vec::new(),
                serialized_chunk_data: bytes::Bytes::from(chunk),
            }),
        ),
    ]
}

#[test]
fn rust_structs_encode_like_gophertunnel() {
    let dir = Path::new(env!("CARGO_MANIFEST_DIR")).join("../../../testdata/encode");
    for (name, sender, packet) in cases() {
        let want = std::fs::read(dir.join(format!("{name}.bin")))
            .expect("read the committed gophertunnel entry");
        let mut frame = Vec::new();
        packet
            .encode_frame(sender, 0, 0, &mut frame)
            .expect("encode");
        let mut reader = wire::Reader::new(&frame);
        let length = reader.read_var_u32().expect("length prefix") as usize;
        let entry = &frame[reader.position()..];
        assert_eq!(length, entry.len(), "{name}: length prefix");
        assert_eq!(
            entry,
            want.as_slice(),
            "{name}: Rust and gophertunnel encodings differ"
        );
    }
}
