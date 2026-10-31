package docs

import (
	"reflect"
	"strings"
	"testing"

	"protocolgen/internal/manifest"
)

func docsFixture() manifest.Manifest {
	node := manifest.Node{Kind: manifest.KindStruct, TypeID: "Shared", Fields: []manifest.Field{{Name: "Value", Encode: manifest.Primitive("u8"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}
	return manifest.Manifest{
		SchemaVersion: 2,
		Target:        manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 1},
		Sources:       []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "1", Digest: "fixture", MinecraftVersion: "fixture", ProtocolVersion: 1}},
		Packets:       []manifest.Packet{{Name: "FixturePacket", Fields: []manifest.Field{{Name: "Shared Value", Encode: node, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}},
	}
}

func TestValidateOverlayRejectsStaleTypeAndField(t *testing.T) {
	m := docsFixture()
	document := Document{
		SchemaVersion: 1,
		Target:        m.Target,
		Entries: []Entry{
			{TypeID: "Missing", Doc: "stale"},
			{TypeID: "Shared", Field: "Missing", Doc: "stale"},
		},
	}
	if err := ValidateOverlay(m, document); err == nil || !strings.Contains(err.Error(), "Missing") {
		t.Fatalf("ValidateOverlay error = %v, want stale reference", err)
	}
}

func TestCoverageCountsSharedAndPacketDocs(t *testing.T) {
	m := docsFixture()
	overlay := Overlay{
		Types:  map[string]string{"Shared": "shared", "FixturePacket": "packet"},
		Fields: map[string]string{"Shared\x00Value": "value docs", "FixturePacket\x00Shared Value": "packet field docs"},
	}
	coverage := CoverageOf(m, overlay)
	if coverage.TypesDocumented != 2 || coverage.TypesTotal != 2 || coverage.FieldsDocumented != 2 || coverage.FieldsTotal != 2 {
		t.Fatalf("CoverageOf = %+v, want 2/2 types and fields", coverage)
	}
}

func TestLeadWithRewritesForeignIdentifier(t *testing.T) {
	cases := []struct{ text, from, to, want string }{
		{"Values contains the index.", "Values", "SubCommandValues", "SubCommandValues contains the index."},
		{"Index is the index.", "Index", "`sub_command_first_value`", "`sub_command_first_value` is the index."},
		{"IndexValue is the index.", "Index", "Other", "IndexValue is the index."},
		{"The command shown to players.", "Name", "Command", "The command shown to players."},
		{"Name's owner.", "Name", "Target", "Target's owner."},
		{"", "Name", "X", ""},
	}
	for _, c := range cases {
		if got := LeadWith(c.text, c.from, c.to); got != c.want {
			t.Fatalf("LeadWith(%q, %q, %q) = %q, want %q", c.text, c.from, c.to, got, c.want)
		}
	}
}

func TestCommentsPreserveParagraphs(t *testing.T) {
	for _, tc := range []struct {
		name   string
		render func(string) []string
		prefix string
	}{
		{"Go", GoComments, "//"}, {"Rust", RustComments, "///"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.render("First paragraph.\r\n\r\nSecond paragraph.")
			want := []string{tc.prefix + " First paragraph.", tc.prefix, tc.prefix + " Second paragraph."}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("comments = %q, want %q", got, want)
			}
		})
	}
}
