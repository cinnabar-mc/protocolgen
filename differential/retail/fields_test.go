package retail

import (
	"math"
	"reflect"
	"slices"
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

func testSchema(t *testing.T) *Schema {
	t.Helper()
	schema, err := LoadSchema("../../generated/1.26.51/rust-contract/src", "../../generated/1.26.51/go/protocol")
	if err != nil {
		t.Fatal(err)
	}
	return schema
}

func fieldDiff(t *testing.T, debug string, pk any) string {
	t.Helper()
	schema := testSchema(t)
	r, err := ParseRustDebug(debug)
	if err != nil {
		t.Fatal(err)
	}
	g, err := goTree(reflect.ValueOf(pk))
	if err != nil {
		t.Fatal(err)
	}
	return diffTrees(r, g, "MovePlayer", schema)
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

// Both codecs must define the same unions with the same discriminants, or the
// variant identity check would have gaps.
func TestUnionTablesAgree(t *testing.T) {
	s := testSchema(t)
	if len(s.RustUnions) != len(s.GoUnions) {
		t.Errorf("%d Rust unions, %d Go unions", len(s.RustUnions), len(s.GoUnions))
	}
	for name, goVariants := range s.GoUnions {
		rustVariants, ok := s.RustUnions[name]
		if !ok {
			t.Errorf("Go union %s has no Rust counterpart", name)
			continue
		}
		goTags, rustTags := map[string]bool{}, map[string]bool{}
		for _, tag := range goVariants {
			goTags[tag] = true
		}
		for _, tag := range rustVariants {
			rustTags[tag] = true
		}
		if !reflect.DeepEqual(goTags, rustTags) {
			t.Errorf("%s: Go discriminants %v, Rust %v", name, goTags, rustTags)
		}
	}
}

const (
	textDebug       = `Text(Text { localize: false, body: MessageOnly { message_type: Raw, message: "hi" }, sender_xuid: "", platform_id: "", filtered_message: None })`
	disconnectEmpty = `Disconnect(Disconnect { reason: Kicked, messages: Empty })`
	disconnectData  = `Disconnect(Disconnect { reason: Kicked, messages: DisconnectPacketMessages { message: "bye", filtered_message: "" } })`
)

func treeDiff(t *testing.T, debug string, pk any, root string) string {
	t.Helper()
	r, err := ParseRustDebug(debug)
	if err != nil {
		t.Fatal(err)
	}
	g, err := goTree(reflect.ValueOf(pk))
	if err != nil {
		t.Fatal(err)
	}
	if name := reflect.TypeOf(pk).Elem().Name(); !slices.Contains(r.tags, name) {
		return "packet " + name
	}
	return diffTrees(r, g, root, testSchema(t))
}

func text(body genprotocol.TextData, xuid string) *genpacket.Text {
	return &genpacket.Text{Body: body, SenderSXUID: xuid}
}

// Every identity the value trees keep must be able to fail on its own: a
// union variant with equal fields, a unit variant, a field name, the packet
// type, a newtype's type, a NaN, a negative zero and a lossy string.
func TestFieldsCatchEveryIdentity(t *testing.T) {
	messageOnly := &genprotocol.MessageOnly{MessageType: genprotocol.TextPacketTypeRaw, Message: "hi"}
	kicked := genprotocol.ConnectionDisconnectFailReasonKicked
	for _, tc := range []struct {
		name  string
		debug string
		pk    any
		want  string // "" for agreement
	}{
		{"text agrees", textDebug, text(messageOnly, ""), ""},
		{"union variant with equal fields", textDebug, text(&genprotocol.AuthorAndMessage{MessageType: genprotocol.TextPacketTypeRaw, Message: "hi"}, ""), "TextData"},
		{"unit variant agrees", disconnectEmpty, &genpacket.Disconnect{Reason: kicked, Messages: &genprotocol.DisconnectMessagesEmpty{}}, ""},
		{"unit variant against data", disconnectEmpty, &genpacket.Disconnect{Reason: kicked, Messages: &genprotocol.DisconnectMessagesData{}}, "discriminant"},
		{"data against unit variant", disconnectData, &genpacket.Disconnect{Reason: kicked, Messages: &genprotocol.DisconnectMessagesEmpty{}}, "discriminant"},
		{"renamed variant agrees", disconnectData, &genpacket.Disconnect{Reason: kicked, Messages: &genprotocol.DisconnectMessagesData{Message: "bye"}}, ""},
		{"field name", strings.Replace(textDebug, "platform_id", "platform_name", 1), text(messageOnly, ""), "PlatformID"},
		{"aliased field name differs in value", textDebug, text(messageOnly, "x"), "SenderSXUID"},
		{"packet type", strings.Replace(textDebug, "Text(Text {", "SetTitle(SetTitle {", 1), text(messageOnly, ""), "packet"},
		{"lossy string", strings.Replace(textDebug, `"hi"`, `"h\u{FFFD}"`, 1), text(messageOnly, ""), "Message"},
		{"lossy string agrees", strings.Replace(textDebug, `"hi"`, `"h\u{FFFD}"`, 1), text(&genprotocol.MessageOnly{Message: "h\xff"}, ""), ""},
	} {
		got := treeDiff(t, tc.debug, tc.pk, "Packet")
		if tc.want == "" && got != "" || tc.want != "" && !strings.Contains(got, tc.want) {
			t.Errorf("%s: diff %q, want one naming %q", tc.name, got, tc.want)
		}
	}
	for _, tc := range []struct {
		name, debug string
		mutate      func(*genpacket.MovePlayer)
		want        string
	}{
		{"nan against a number", strings.Replace(movePlayerDebug, "64.0", "NaN", 1), func(*genpacket.MovePlayer) {}, "Position[1]"},
		{"negative zero", strings.Replace(movePlayerDebug, "F32LE(30.75)", "F32LE(-0.0)", 1), func(pk *genpacket.MovePlayer) { pk.YHeadRotation = 0 }, "YHeadRotation"},
		{"nan agrees", strings.Replace(movePlayerDebug, "F32LE(30.75)", "F32LE(NaN)", 1), func(pk *genpacket.MovePlayer) { pk.YHeadRotation = float32(math.NaN()) }, ""},
		// Go flattens this newtype to uint64, so only the value is comparable;
		// the Rust newtype is the field's static type and cannot vary by value.
		{"flattened newtype", strings.Replace(movePlayerDebug, "PlayerInputTick(1234)", "ActorUniqueID(1234)", 1), func(*genpacket.MovePlayer) {}, ""},
	} {
		pk := generatedMovePlayer()
		tc.mutate(pk)
		if got := fieldDiff(t, tc.debug, pk); tc.want == "" && got != "" || tc.want != "" && !strings.Contains(got, tc.want) {
			t.Errorf("%s: diff %q, want one naming %q", tc.name, got, tc.want)
		}
	}
	// A Go one-field struct must be the newtype Rust names around the value.
	g, err := goTree(reflect.ValueOf(genprotocol.DimensionType{Value: 1}))
	if err != nil {
		t.Fatal(err)
	}
	for debug, want := range map[string]bool{"DimensionType(1)": true, "ActorRuntimeID(1)": false, "DimensionType(2)": false} {
		r, err := ParseRustDebug(debug)
		if err != nil {
			t.Fatal(err)
		}
		if got := diffTrees(r, g, "DimensionType", testSchema(t)) == ""; got != want {
			t.Errorf("%s against DimensionType{1}: agree %v, want %v", debug, got, want)
		}
	}
}
