package retail

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-gl/mathgl/mgl32"
	gtprotocol "github.com/sandertv/gophertunnel/minecraft/protocol"
	gtpacket "github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	genprotocol "protocolgen/generated/1.26.51/go/protocol"
	genpacket "protocolgen/generated/1.26.51/go/protocol/packet"
)

// EncodeDir holds the gophertunnel encodings of the encode-from-struct
// cases, relative to this module. bench/tests/encode_from_struct.rs builds the
// same values as Rust structs and must encode the same bytes.
const EncodeDir = "../../testdata/encode"

// EncodeCase is one packet value built independently in gophertunnel and the
// generated Go codec, which must encode it to the same entry.
type EncodeCase struct {
	Name         string
	Gophertunnel gtpacket.Packet
	Generated    genpacket.Packet
}

// EncodeCases returns the encode-from-struct cases.
func EncodeCases() []EncodeCase {
	chunk := make([]byte, 1024)
	for i := range chunk {
		chunk[i] = byte(i * 7)
	}
	return []EncodeCase{
		{"move_player",
			&gtpacket.MovePlayer{EntityRuntimeID: 42, Position: mgl32.Vec3{1.25, 64, -2.5}, Pitch: 10.5, Yaw: 20.25, HeadYaw: 30.75,
				Mode: gtpacket.MoveModeTeleport, OnGround: true, TeleportData: gtprotocol.Option(gtprotocol.TeleportData{TeleportCause: 3, TeleportSourceEntityType: 87}), Tick: 1234},
			&genpacket.MovePlayer{PlayerRuntimeID: 42, Position: mgl32.Vec3{1.25, 64, -2.5}, Rotation: mgl32.Vec2{10.5, 20.25}, YHeadRotation: 30.75,
				PositionMode: genprotocol.PlayerPositionModeComponentPositionModeTeleport, OnGround: true,
				TeleportData: genprotocol.Option(genprotocol.MovePlayerTeleportData{TeleportationCause: 3, SourceActorType: 87}), Tick: 1234}},
		{"set_actor_motion",
			&gtpacket.SetActorMotion{EntityRuntimeID: 9001, Velocity: mgl32.Vec3{0.125, -0.0784, 3.5}, Tick: 77},
			&genpacket.SetActorMotion{TargetRuntimeID: 9001, Motion: mgl32.Vec3{0.125, -0.0784, 3.5}, Tick: 77}},
		{"update_block",
			&gtpacket.UpdateBlock{Position: gtprotocol.BlockPos{-30_000_000, -64, 29_999_999}, NewBlockRuntimeID: 0xdead_beef, Flags: 3, Layer: 1},
			&genpacket.UpdateBlock{BlockPosition: genprotocol.BlockPos{X: -30_000_000, Y: -64, Z: 29_999_999}, BlockRuntimeID: 0xdead_beef, Flags: 3, Layer: 1}},
		{"network_stack_latency",
			&gtpacket.NetworkStackLatency{Timestamp: 1_760_000_000_123, NeedsResponse: true},
			&genpacket.NetworkStackLatency{CreationTime: 1_760_000_000_123, IsFromServer: true}},
		{"update_soft_enum",
			&gtpacket.UpdateSoftEnum{EnumType: "Players", Options: []string{"Steve", "Alex", "§cZuri"}, ActionType: gtpacket.SoftEnumActionSet},
			&genpacket.UpdateSoftEnum{EnumName: "Players", Values: []string{"Steve", "Alex", "§cZuri"}, UpdateType: genprotocol.SoftEnumUpdateTypeReplace}},
		{"level_chunk",
			&gtpacket.LevelChunk{Position: gtprotocol.ChunkPos{-3, 12}, Dimension: 1, SubChunkCount: 4, RawPayload: chunk},
			&genpacket.LevelChunk{ChunkPosition: genprotocol.ChunkPos{X: -3, Z: 12}, DimensionID: genprotocol.DimensionType{Value: 1}, SubChunksCount: 4, SerializedChunkData: chunk}},
	}
}

// EncodeCaseEntries returns each case's gophertunnel-encoded entry, header
// varint and body, keyed by case name.
func EncodeCaseEntries() (map[string][]byte, error) {
	out := map[string][]byte{}
	for _, c := range EncodeCases() {
		payload, err := encodeGophertunnel(gtpacket.Header{PacketID: c.Gophertunnel.ID()}, c.Gophertunnel)
		if err != nil {
			return nil, fmt.Errorf("encode %s: %w", c.Name, err)
		}
		out[c.Name] = payload
	}
	return out, nil
}

// WriteEncodeCases writes each case's entry to dir/<name>.bin.
func WriteEncodeCases(dir string) error {
	entries, err := EncodeCaseEntries()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for name, payload := range entries {
		if err := os.WriteFile(filepath.Join(dir, name+".bin"), payload, 0o644); err != nil {
			return err
		}
	}
	return nil
}
