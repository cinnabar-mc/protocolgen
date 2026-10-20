package retail

import (
	"reflect"
	"strings"
	"testing"

	"github.com/go-gl/mathgl/mgl32"
	genprotocol "protocolgen/generated/1.26.51/go/protocol"
	genpacket "protocolgen/generated/1.26.51/go/protocol/packet"
)

const movePlayerDebug = `MovePlayer(MovePlayer { player_runtime_id: ActorRuntimeID(42), position: Vec3(1.25, 64.0, -2.5), ` +
	`rotation: Vec2(10.5, 20.25), y_head_rotation: F32LE(30.75), position_mode: Teleport, on_ground: true, ` +
	`riding_runtime_id: ActorRuntimeID(0), teleport_data: Some(MovePlayerTeleportData { teleportation_cause: I32LE(3), ` +
	`source_actor_type: I32LE(87) }), tick: PlayerInputTick(1234) })`

func generatedMovePlayer() *genpacket.MovePlayer {
	return &genpacket.MovePlayer{
		PlayerRuntimeID: 42, Position: mgl32.Vec3{1.25, 64, -2.5}, Rotation: mgl32.Vec2{10.5, 20.25}, YHeadRotation: 30.75,
		PositionMode: genprotocol.PlayerPositionModeComponentPositionModeTeleport, OnGround: true,
		TeleportData: genprotocol.Option(genprotocol.MovePlayerTeleportData{TeleportationCause: 3, SourceActorType: 87}), Tick: 1234,
	}
}

func fieldDiff(t *testing.T, debug string, pk any) string {
	t.Helper()
	enums, err := LoadRustEnums("../../generated/1.26.51/rust-contract/src")
	if err != nil {
		t.Fatal(err)
	}
	r, err := ParseRustDebug(debug)
	if err != nil {
		t.Fatal(err)
	}
	g, err := goTree(reflect.ValueOf(pk))
	if err != nil {
		t.Fatal(err)
	}
	return diffTrees(r, g, "MovePlayer", enums)
}

func TestFieldsAgreeAcrossRenderings(t *testing.T) {
	if diff := fieldDiff(t, movePlayerDebug, generatedMovePlayer()); diff != "" {
		t.Fatal(diff)
	}
}

// Every leaf kind must be able to fail: a float, an enum variant, an
// optional's presence and a nested integer.
func TestFieldsReportEachDifference(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*genpacket.MovePlayer)
		want   string
	}{
		{"float", func(pk *genpacket.MovePlayer) { pk.Position[1] = 64.5 }, "Position[1]"},
		{"enum", func(pk *genpacket.MovePlayer) { pk.PositionMode = 1 }, "PositionMode"},
		{"optional", func(pk *genpacket.MovePlayer) {
			pk.TeleportData = genprotocol.Optional[genprotocol.MovePlayerTeleportData]{}
		}, "TeleportData"},
		{"nested", func(pk *genpacket.MovePlayer) {
			pk.TeleportData = genprotocol.Option(genprotocol.MovePlayerTeleportData{TeleportationCause: 4, SourceActorType: 87})
		}, "TeleportationCause"},
	} {
		pk := generatedMovePlayer()
		tc.mutate(pk)
		if diff := fieldDiff(t, movePlayerDebug, pk); !strings.Contains(diff, tc.want) {
			t.Errorf("%s: diff %q does not name %s", tc.name, diff, tc.want)
		}
	}
}

func TestRustDebugStringsAndBytesUnescape(t *testing.T) {
	n, err := ParseRustDebug(`X { s: "a\"b\\c\n\u{e9}", b: b"\x00\xffz\t", v: [], n: None }`)
	if err != nil {
		t.Fatal(err)
	}
	if n.items[0].text != "a\"b\\c\né" || n.items[1].text != "\x00\xffz\t" || len(n.items[2].items) != 0 || n.items[3].kind != nodeNone {
		t.Fatalf("parsed %+v", n.items)
	}
}

// Each maximal ill-formed subpart becomes one U+FFFD, as from_utf8_lossy
// renders it: two for ff ff, one for a truncated three-byte sequence.
func TestLossyMatchesRustReplacement(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"\xff\xff", "\uFFFD\uFFFD"},
		{"a\xe2\x82b", "a\uFFFDb"},
		{"\xe2\x82\xac\xe2", "\u20ac\uFFFD"},
		{"\xf0\x80\x80", "\uFFFD\uFFFD\uFFFD"},
		{"\xed\xa0\x80", "\uFFFD\uFFFD\uFFFD"},
		{"\xf4\x8f\xbf", "\uFFFD"},
		{"\xc0\xaf", "\uFFFD\uFFFD"},
		{"ok", "ok"},
	} {
		if got := lossy(tc.in); got != tc.want {
			t.Errorf("lossy(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
