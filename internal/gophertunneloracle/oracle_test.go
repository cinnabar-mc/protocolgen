package gophertunneloracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"protocolgen/internal/manifest"
)

func TestNormalizationAppliesOnlyDocumentedByteEquivalences(t *testing.T) {
	m := fixtureManifest(
		manifest.Primitive("i16le"),
		manifest.String(manifest.Primitive("var_u32")),
		manifest.Bytes(manifest.Primitive("var_u32")),
		manifest.Array(manifest.Primitive("var_u32"), manifest.Primitive("u8")),
		manifest.Primitive("uuid"),
	)
	source := extraction{Packets: []sourcePacket{{ID: 1, Name: "Fixture", Operations: []sourceOperation{
		{Kind: "primitive", Code: "u16le"},
		{Kind: "bytes", Prefix: "var_u32"},
		{Kind: "string", Prefix: "var_u32"},
		{Kind: "array", Prefix: "var_u32", Element: []sourceOperation{{Kind: "primitive", Code: "u8"}}},
		{Kind: "uuid"},
	}}}}
	lock := fixtureLock()
	report := Compare(m, source, lock, AcceptedFile{SchemaVersion: AcceptedSchemaVersion, MinecraftVersion: "1.26.40", ProtocolVersion: 2168}, "manifest.json")
	if report.Counts.Agreement != 1 || report.Counts.Divergence != 0 {
		t.Fatalf("counts = %#v, want one agreement: %#v", report.Counts, report.Packets)
	}
}

func manifestExpr(node manifest.Node) shapeExpr {
	return (&canonicalContext{ancestors: map[string]manifest.Node{}, depth: map[string]int{}}).node("Field", node)
}

func TestUUIDNormalizesOnlyItsExactSixteenByteShape(t *testing.T) {
	uuid := manifestExpr(manifest.FixedArray(16, manifest.Primitive("u8")))
	if witness := compareLanguages(uuid, sourceOperationExpr(sourceOperation{Kind: "uuid", Field: "UUID"})); witness != nil {
		t.Fatalf("UUID shape did not normalize: %#v", witness)
	}
	short := sourceOperationExpr(sourceOperation{Kind: "fixed_array", Length: 15, Element: []sourceOperation{{Kind: "primitive", Code: "u8"}}})
	if compareLanguages(uuid, short) == nil {
		t.Fatal("UUID normalization collapsed a non-16-byte fixed array")
	}
}

func TestFixedArrayGroupingNormalizesOnlyWireEquivalentScalarLayout(t *testing.T) {
	nested := manifestExpr(manifest.FixedArray(16, manifest.FixedArray(16, manifest.Primitive("i8"))))
	flat := sourceOperationExpr(sourceOperation{Kind: "fixed_array", Length: 256, Element: []sourceOperation{{Kind: "primitive", Code: "i8"}}})
	if witness := compareLanguages(nested, flat); witness != nil {
		t.Fatalf("nested and flat fixed arrays did not normalize: %#v", witness)
	}
	short := sourceOperationExpr(sourceOperation{Kind: "fixed_array", Length: 255, Element: []sourceOperation{{Kind: "primitive", Code: "i8"}}})
	if compareLanguages(nested, short) == nil {
		t.Fatal("different scalar counts were normalized as equivalent")
	}
}

func TestNormalizationPreservesWireShapeDistinctions(t *testing.T) {
	for name, pair := range map[string]struct {
		want sourceOperation
		got  sourceOperation
	}{
		"endianness": {
			want: sourceOperation{Kind: "primitive", Code: "u32le"},
			got:  sourceOperation{Kind: "primitive", Code: "u32be"},
		},
		"varint-family": {
			want: sourceOperation{Kind: "primitive", Code: "var_u32"},
			got:  sourceOperation{Kind: "primitive", Code: "zigzag_i32"},
		},
		"option-presence": {
			want: sourceOperation{Kind: "optional", Presence: "bool", Value: []sourceOperation{{Kind: "primitive", Code: "u8"}}},
			got:  sourceOperation{Kind: "optional", Presence: "u8", Value: []sourceOperation{{Kind: "primitive", Code: "u8"}}},
		},
		"array-prefix": {
			want: sourceOperation{Kind: "array", Prefix: "var_u32", Element: []sourceOperation{{Kind: "primitive", Code: "u16le"}}},
			got:  sourceOperation{Kind: "array", Prefix: "u8", Element: []sourceOperation{{Kind: "primitive", Code: "u16le"}}},
		},
		"fixed-array-length": {
			want: sourceOperation{Kind: "fixed_array", Length: 2, Element: []sourceOperation{{Kind: "primitive", Code: "u8"}}},
			got:  sourceOperation{Kind: "fixed_array", Length: 3, Element: []sourceOperation{{Kind: "primitive", Code: "u8"}}},
		},
		"float-versus-integer": {
			want: sourceOperation{Kind: "primitive", Code: "f32le"},
			got:  sourceOperation{Kind: "primitive", Code: "i32le"},
		},
		"union-discriminant": {
			want: sourceOperation{Kind: "union", Control: "var_u32", Variants: []sourceVariant{{Value: 0}, {Value: 1}}},
			got:  sourceOperation{Kind: "union", Control: "var_u32", Variants: []sourceVariant{{Value: 0}, {Value: 2}}},
		},
	} {
		if compareLanguages(sourceOperationExpr(pair.want), sourceOperationExpr(pair.got)) == nil {
			t.Errorf("%s was collapsed", name)
		}
	}
}

// A little-endian colour int and BEARGB are the same four bytes; a plain
// big-endian int is not.
func TestColourIntMatchesBEARGB(t *testing.T) {
	want := manifestExpr(manifest.Primitive("i32le"))
	if witness := compareLanguages(want, sourceOperationExpr(sourceOperation{Kind: "primitive", Code: "argb32"})); witness != nil {
		t.Fatalf("colour did not normalize: %#v", witness)
	}
	if compareLanguages(want, sourceOperationExpr(sourceOperation{Kind: "primitive", Code: "i32be"})) == nil {
		t.Fatal("a plain big-endian int was accepted as a colour")
	}
}

// A bool-guarded field and a manifest optional are the same bytes.
func TestBoolGuardedFieldMatchesManifestOptional(t *testing.T) {
	want := manifestExpr(manifest.Optional(manifest.Primitive("u16le")))
	got := sourceSequenceExpr([]sourceOperation{
		{Kind: "primitive", Code: "bool", Field: "Field.Has"},
		{Kind: "conditional", CompareTo: "Field.Has", Variants: []sourceVariant{{Value: 1, Values: []int64{1}, Ops: []sourceOperation{{Kind: "primitive", Code: "u16le"}}}}, HasDefault: true},
	})
	if witness := compareLanguages(want, got); witness != nil {
		t.Fatalf("bool guard did not match optional: %#v", witness)
	}
}

// A conditional on a discriminant read earlier compares as that union variant,
// with the else branch covering every other discriminant.
func TestDiscriminantConditionalHoistsToTheControlRead(t *testing.T) {
	variants := []manifest.Variant{}
	for value := int64(0); value < 4; value++ {
		payload := manifest.Struct(manifest.Field{Ordinal: 0, Name: "Value", Encode: manifest.Primitive("u16le"), Symmetry: manifest.Symmetric})
		if value == 3 {
			payload = manifest.Struct(
				manifest.Field{Ordinal: 0, Name: "Value", Encode: manifest.Primitive("u16le"), Symmetry: manifest.Symmetric},
				manifest.Field{Ordinal: 1, Name: "Flag", Encode: manifest.Primitive("bool"), Symmetry: manifest.Symmetric},
			)
		}
		variants = append(variants, manifest.Variant{Value: value, Name: "V", Encode: payload})
	}
	want := manifestExpr(manifest.Union(manifest.Primitive("var_u32"), variants...))
	cases := make([]sourceVariant, len(variants))
	for index, variant := range variants {
		cases[index] = sourceVariant{Value: variant.Value, Discriminant: true}
	}
	closedSelector := sourceOperation{Kind: "switch", CompareTo: "Field.Type", Variants: cases}
	got := sourceSequenceExpr([]sourceOperation{
		{Kind: "primitive", Code: "var_u32", Field: "Field.Type"},
		closedSelector,
		{Kind: "primitive", Code: "u16le", Field: "Field.Value"},
		{Kind: "conditional", CompareTo: "Field.Type", Variants: []sourceVariant{{Values: []int64{3}, Discriminant: true, Ops: []sourceOperation{{Kind: "primitive", Code: "bool"}}}}, HasDefault: true},
	})
	if witness := compareLanguages(want, got); witness != nil {
		t.Fatalf("discriminant conditional was not hoisted: %#v", witness)
	}
	negated := sourceSequenceExpr([]sourceOperation{
		{Kind: "primitive", Code: "var_u32", Field: "Field.Type"},
		closedSelector,
		{Kind: "primitive", Code: "u16le", Field: "Field.Value"},
		{Kind: "conditional", CompareTo: "Field.Type", Variants: []sourceVariant{{Values: []int64{3}, Discriminant: true, Negated: true}}, Default: []sourceOperation{{Kind: "primitive", Code: "bool"}}, HasDefault: true},
	})
	if witness := compareLanguages(want, negated); witness != nil {
		t.Fatalf("negated discriminant conditional was not hoisted: %#v", witness)
	}
	if compareLanguages(want, sourceSequenceExpr([]sourceOperation{
		{Kind: "primitive", Code: "var_u32", Field: "Field.Type"},
		closedSelector,
		{Kind: "primitive", Code: "u16le", Field: "Field.Value"},
		{Kind: "conditional", CompareTo: "Field.Type", Variants: []sourceVariant{{Values: []int64{2}, Discriminant: true, Ops: []sourceOperation{{Kind: "primitive", Code: "bool"}}}}, HasDefault: true},
	})) == nil {
		t.Fatal("a different discriminant value was accepted")
	}
}

func TestLanguageWitnessNamesTheFirstDivergingAtom(t *testing.T) {
	want := manifestExpr(manifest.Struct(
		manifest.Field{Ordinal: 0, Name: "A", Encode: manifest.Primitive("u8"), Symmetry: manifest.Symmetric},
		manifest.Field{Ordinal: 1, Name: "B", Encode: manifest.Optional(manifest.Primitive("u16le")), Symmetry: manifest.Symmetric},
	))
	got := sourceSequenceExpr([]sourceOperation{
		{Kind: "primitive", Code: "u8", Field: "A"},
		{Kind: "optional", Presence: "bool", Field: "B", Value: []sourceOperation{{Kind: "primitive", Code: "u32le", Field: "B"}}},
	})
	witness := compareLanguages(want, got)
	if witness == nil {
		t.Fatal("different optional payload widths compared equal")
	}
	if len(witness.prefix) != 3 || witness.manifest[0].Token != "P:FIXED16LE" || witness.gophertunnel[0].Token != "P:FIXED32LE" {
		t.Fatalf("witness = %#v", witness)
	}
}

func TestRuntimeBranchesBecomeUnresolved(t *testing.T) {
	m := fixtureManifest(manifest.Primitive("u8"))
	source := extraction{Packets: []sourcePacket{{ID: 1, Name: "Fixture", Operations: []sourceOperation{
		{Kind: "unresolved", Reason: "runtime conditional branch", Site: "fixture.go:10"},
	}}}}
	report := Compare(m, source, fixtureLock(), AcceptedFile{SchemaVersion: AcceptedSchemaVersion, MinecraftVersion: "1.26.40", ProtocolVersion: 2168}, "manifest.json")
	if report.Counts.Unresolved != 1 || report.Counts.Agreement != 0 {
		t.Fatalf("counts = %#v", report.Counts)
	}
	if !strings.Contains(report.Packets[0].Reasons[0], "runtime conditional") {
		t.Fatalf("reasons = %#v", report.Packets[0].Reasons)
	}
}

func TestExtractRecursesThroughSliceCallbackWithoutTypeChecking(t *testing.T) {
	root := t.TempDir()
	packetDir := filepath.Join(root, "minecraft", "protocol", "packet")
	if err := os.MkdirAll(packetDir, 0o755); err != nil {
		t.Fatal(err)
	}
	idSource := `package packet
const ( IDFixture = iota + 1 )
`
	packetSource := `package packet
import "github.com/sandertv/gophertunnel/minecraft/protocol"
type Fixture struct { Values []uint8; Value uint16 }
func (*Fixture) ID() uint32 { return IDFixture }
func (pk *Fixture) Marshal(io protocol.IO) {
    protocol.FuncSlice(io, &pk.Values, io.Uint8)
    io.Uint16(&pk.Value)
}
`
	if err := os.WriteFile(filepath.Join(packetDir, "id.go"), []byte(idSource), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(packetDir, "fixture.go"), []byte(packetSource), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := Extract(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Packets) != 1 || result.Packets[0].ID != 1 {
		t.Fatalf("packets = %#v", result.Packets)
	}
	if len(result.Packets[0].Operations) != 2 || result.Packets[0].Operations[0].Kind != "array" || result.Packets[0].Operations[1].Code != "u16le" {
		t.Fatalf("operations = %#v", result.Packets[0].Operations)
	}
}

func TestExtractLeavesPartiallyBoundBooleanGuardUnresolved(t *testing.T) {
	root := t.TempDir()
	packetDir := filepath.Join(root, "minecraft", "protocol", "packet")
	if err := os.MkdirAll(packetDir, 0o755); err != nil {
		t.Fatal(err)
	}
	idSource := `package packet
const (
	IDFixture = iota + 1
	ModeA = 1
	ModeB = 2
)
`
	packetSource := `package packet
import "github.com/sandertv/gophertunnel/minecraft/protocol"
type Fixture struct { Mode uint8; Flag bool; Value uint16 }
func (*Fixture) ID() uint32 { return IDFixture }
func (pk *Fixture) Marshal(io protocol.IO) {
	io.Uint8(&pk.Mode)
	switch pk.Mode {
	case ModeA:
		io.Uint16(&pk.Value)
	case ModeB:
		io.Bool(&pk.Flag)
	}
	if pk.Flag {
		io.Uint16(&pk.Value)
	}
}
`
	if err := os.WriteFile(filepath.Join(packetDir, "id.go"), []byte(idSource), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(packetDir, "fixture.go"), []byte(packetSource), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := Extract(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Packets) != 1 {
		t.Fatalf("packets = %#v", result.Packets)
	}
	got := sourceSequenceExpr(result.Packets[0].Operations)
	if got.kind != "unknown" {
		t.Fatalf("partially bound boolean guard was accepted: %s", expressionKey(got))
	}
}

// A hand-written type switch that writes a constant discriminant at the start
// of each case compares as a manifest union.
func TestExtractResolvesTypeSwitchDiscriminants(t *testing.T) {
	root := t.TempDir()
	packetDir := filepath.Join(root, "minecraft", "protocol", "packet")
	if err := os.MkdirAll(packetDir, 0o755); err != nil {
		t.Fatal(err)
	}
	idSource := `package packet
const (
	IDFixture = iota + 1
	TypeBool = 1
)
`
	packetSource := `package packet
import "github.com/sandertv/gophertunnel/minecraft/protocol"
type Fixture struct { Value any }
func (*Fixture) ID() uint32 { return IDFixture }
func (pk *Fixture) Marshal(io protocol.IO) {
	switch v := pk.Value.(type) {
	case bool:
		id := uint32(TypeBool)
		io.Varuint32(&id)
		io.Bool(&v)
	case float32:
		id := uint32(2)
		io.Varuint32(&id)
		io.Float32(&v)
	}
}
`
	if err := os.WriteFile(filepath.Join(packetDir, "id.go"), []byte(idSource), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(packetDir, "fixture.go"), []byte(packetSource), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := Extract(root)
	if err != nil {
		t.Fatal(err)
	}
	want := manifestExpr(manifest.Union(manifest.Primitive("var_u32"),
		manifest.Variant{Value: 1, Name: "Bool", Encode: manifest.Primitive("bool")},
		manifest.Variant{Value: 2, Name: "Float", Encode: manifest.Primitive("f32le")},
	))
	got := sourceSequenceExpr(result.Packets[0].Operations)
	if witness := compareLanguages(want, got); witness != nil {
		t.Fatalf("type switch discriminants were not resolved: %#v\n%s", witness, expressionKey(got))
	}
}

func TestExtractFollowsStaticallyResolvableLocalWireCall(t *testing.T) {
	root := t.TempDir()
	packetDir := filepath.Join(root, "minecraft", "protocol", "packet")
	if err := os.MkdirAll(packetDir, 0o755); err != nil {
		t.Fatal(err)
	}
	idSource := `package packet
const IDFixture = 1
`
	packetSource := `package packet
import "github.com/sandertv/gophertunnel/minecraft/protocol"
type Fixture struct { Value uint16 }
func (*Fixture) ID() uint32 { return IDFixture }
func encodeValue(io protocol.IO, value *uint16) { io.Uint16(value) }
func (pk *Fixture) Marshal(io protocol.IO) { encodeValue(io, &pk.Value) }
`
	if err := os.WriteFile(filepath.Join(packetDir, "id.go"), []byte(idSource), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(packetDir, "fixture.go"), []byte(packetSource), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := Extract(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Packets) != 1 || len(result.Packets[0].Operations) != 1 {
		t.Fatalf("packets = %#v", result.Packets)
	}
	operation := result.Packets[0].Operations[0]
	if operation.Kind != "primitive" || operation.Code != "u16le" {
		t.Fatalf("operation = %#v, want resolved local call", operation)
	}
}

func TestExtractTraversesNestedFunctionLiteral(t *testing.T) {
	root := t.TempDir()
	packetDir := filepath.Join(root, "minecraft", "protocol", "packet")
	if err := os.MkdirAll(packetDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(packetDir, "fixture.go"), []byte(`package packet
import "github.com/sandertv/gophertunnel/minecraft/protocol"
const IDFixture = 1
type Fixture struct { Value uint16 }
func (*Fixture) ID() uint32 { return IDFixture }
func (pk *Fixture) Marshal(io protocol.IO) {
	encode := func() { io.Uint16(&pk.Value) }
	encode()
}
`), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := Extract(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Packets) != 1 || len(result.Packets[0].Operations) != 1 || result.Packets[0].Operations[0].Code != "u16le" {
		t.Fatalf("operations = %#v, want nested function wire operation", result.Packets)
	}
}

func TestComparePreservesExactUnionVariantValues(t *testing.T) {
	union := manifest.Union(
		manifest.Primitive("var_u32"),
		manifest.Variant{Value: 0, Name: "Zero", Encode: manifest.Void()},
		manifest.Variant{Value: 1, Name: "One", Encode: manifest.Void()},
	)
	m := fixtureManifest(union)
	operations := []sourceOperation{
		{Kind: "primitive", Code: "var_u32"},
		{Kind: "switch", CompareTo: "Field", Variants: []sourceVariant{
			{Values: []int64{0}},
			{Values: []int64{2}},
		}},
	}
	source := extraction{Packets: []sourcePacket{{ID: 1, Name: "Fixture", Operations: operations}}}
	report := Compare(m, source, fixtureLock(), emptyAccepted(), "manifest.json")
	if report.Counts.Divergence != 1 || report.Counts.Agreement != 0 {
		t.Fatalf("counts = %#v, packets = %#v", report.Counts, report.Packets)
	}
	if len(report.Packets[0].Differences) == 0 || len(report.Packets[0].ManifestSequence) == 0 || report.Packets[0].Fingerprint == "" {
		t.Fatalf("variant witness evidence missing: %#v", report.Packets[0])
	}
}

func TestOptionalUnionsCompareWithoutCartesianProduct(t *testing.T) {
	fields := make([]manifest.Node, 6)
	operations := make([]sourceOperation, 0, len(fields))
	for index := range fields {
		fields[index] = manifest.Optional(manifest.Union(
			manifest.Primitive("var_u32"),
			manifest.Variant{Value: 0, Name: "Zero", Encode: manifest.Primitive("u8")},
			manifest.Variant{Value: 1, Name: "One", Encode: manifest.Primitive("u8")},
		))
		operations = append(operations, sourceOperation{Kind: "optional", Field: "Field", Presence: "bool", Value: []sourceOperation{
			{Kind: "primitive", Code: "var_u32"},
			{Kind: "union", Control: "", Variants: []sourceVariant{
				{Value: 0, Ops: []sourceOperation{{Kind: "primitive", Code: "u8"}}},
				{Value: 1, Ops: []sourceOperation{{Kind: "primitive", Code: "u8"}}},
			}},
		}})
	}
	m := fixtureManifest(fields...)
	source := extraction{Packets: []sourcePacket{{ID: 1, Name: "Fixture", Operations: operations}}}
	report := Compare(m, source, fixtureLock(), emptyAccepted(), "manifest.json")
	if report.Counts.Agreement != 1 || report.Counts.Unresolved != 0 {
		t.Fatalf("counts = %#v, packets = %#v", report.Counts, report.Packets)
	}
}

func TestExternalLengthIsCoalescedWithTheFollowingArray(t *testing.T) {
	expression := sourceSequenceExpr([]sourceOperation{
		{Kind: "primitive", Code: "u32le"},
		{Kind: "array", Prefix: "u32le", ConsumesPrefix: true, Element: []sourceOperation{{Kind: "primitive", Code: "u8"}}},
	})
	if key := expressionKey(expression); key != "t:LEN:FIXED32LE" {
		t.Fatalf("external length was not coalesced: %s", key)
	}
}

func TestReviewedHelperRequiresPinnedRevision(t *testing.T) {
	root := t.TempDir()
	protocolDir := filepath.Join(root, "minecraft", "protocol")
	packetDir := filepath.Join(protocolDir, "packet")
	if err := os.MkdirAll(packetDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(protocolDir, "writer.go"), []byte(`package protocol
type IO interface{}
type Color struct{}
type Writer struct{}
func (w *Writer) Float32(*float32) {}
func (w *Writer) RGB(x *Color) { var value float32; w.Float32(&value); w.Float32(&value); w.Float32(&value) }
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(packetDir, "fixture.go"), []byte(`package packet
import "example/protocol"
const IDFixture = 1
type Fixture struct { Value uint32 }
func (*Fixture) ID() uint32 { return IDFixture }
func (pk *Fixture) Marshal(io protocol.IO) { io.RGB(nil) }
`), 0o644); err != nil {
		t.Fatal(err)
	}
	unpinned, err := ExtractAtRevision(root, "1111111111111111111111111111111111111111")
	if err != nil {
		t.Fatal(err)
	}
	if len(unpinned.Packets) != 1 || len(unpinned.Packets[0].Operations) != 1 || unpinned.Packets[0].Operations[0].Kind != "unresolved" {
		t.Fatalf("unpinned helper was admitted: %#v", unpinned.Packets)
	}
	pinned, err := ExtractAtRevision(root, "77149b94f8024826e34efda6f625c366596118d2")
	if err != nil {
		t.Fatal(err)
	}
	if len(pinned.Packets) != 1 || len(pinned.Packets[0].Operations) != 3 || pinned.Packets[0].Operations[0].Code != "f32le" {
		t.Fatalf("pinned helper was not expanded: %#v", pinned.Packets)
	}
}

func emptyAccepted() AcceptedFile {
	return AcceptedFile{SchemaVersion: AcceptedSchemaVersion, MinecraftVersion: "1.26.40", ProtocolVersion: 2168}
}

func TestLoadAcceptedRequiresEvidence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "accepted.json")
	data := `{"schema_version":1,"minecraft_version":"1.26.40","protocol_version":2168,"divergences":[{"id":1,"name":"P","reason":"reviewed","what_would_settle_it":"capture","evidence":[]}]}`
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadAccepted(path); err == nil {
		t.Fatal("LoadAccepted accepted an entry without evidence")
	}
}

func TestAcceptedEvidenceMustCiteTheLockedOracle(t *testing.T) {
	lock := fixtureLock()
	lock.Gophertunnel.Repo = "https://github.com/example/gophertunnel.git"
	entry := AcceptedDivergence{ID: 1, Name: "P", Reason: "reviewed", WhatWouldSettleIt: "capture"}
	for _, locator := range []string{
		"https://github.com/other/gophertunnel/blob/" + lock.Gophertunnel.Commit + "/minecraft/protocol/packet/p.go",
		"https://github.com/example/gophertunnel/blob/1111111111111111111111111111111111111111/minecraft/protocol/packet/p.go",
	} {
		entry.Evidence = []Evidence{{Locator: locator, Summary: "pinned marshal"}}
		file := AcceptedFile{SchemaVersion: AcceptedSchemaVersion, MinecraftVersion: "1.26.40", ProtocolVersion: 2168, Divergences: []AcceptedDivergence{entry}}
		if err := checkAcceptedEvidence(lock, file); err == nil {
			t.Errorf("locator %q outside the locked oracle was accepted", locator)
		}
	}

	entry.Evidence = []Evidence{
		{Locator: "https://github.com/example/gophertunnel/blob/" + lock.Gophertunnel.Commit + "/minecraft/protocol/packet/p.go", Summary: "pinned marshal"},
		{Locator: "https://github.com/CloudburstMC/Protocol/blob/fbbeee7/Serializer.java", Summary: "independent serializer"},
	}
	file := AcceptedFile{SchemaVersion: AcceptedSchemaVersion, MinecraftVersion: "1.26.40", ProtocolVersion: 2168, Divergences: []AcceptedDivergence{entry}}
	if err := checkAcceptedEvidence(lock, file); err != nil {
		t.Fatalf("locked-oracle and independent evidence were rejected: %v", err)
	}
}

// TestCheckedInOracleBaselineMatchesTheCanonicalManifest keeps the committed
// lock and reviewed baseline from drifting away from the manifest without a
// gophertunnel checkout being available.
func TestCheckedInOracleBaselineMatchesTheCanonicalManifest(t *testing.T) {
	canonical, err := manifest.Load(filepath.Join("..", "..", "generated", "1.26.60", "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	lock, err := LoadLock(filepath.Join("..", "..", "tools", "gophertunnel-oracle", "lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	accepted, err := LoadAccepted(filepath.Join("..", "..", "tools", "gophertunnel-oracle", "accepted-divergences.json"))
	if err != nil {
		t.Fatal(err)
	}
	if lock.MinecraftVersion != canonical.Target.MinecraftVersion || lock.ProtocolVersion != canonical.Target.ProtocolVersion {
		t.Fatalf("lock targets %s/%d, manifest targets %s/%d", lock.MinecraftVersion, lock.ProtocolVersion, canonical.Target.MinecraftVersion, canonical.Target.ProtocolVersion)
	}
	if accepted.MinecraftVersion != canonical.Target.MinecraftVersion || accepted.ProtocolVersion != canonical.Target.ProtocolVersion {
		t.Fatalf("baseline targets %s/%d, manifest targets %s/%d", accepted.MinecraftVersion, accepted.ProtocolVersion, canonical.Target.MinecraftVersion, canonical.Target.ProtocolVersion)
	}
	if err := checkAcceptedEvidence(lock, accepted); err != nil {
		t.Fatal(err)
	}
	names := make(map[uint32]string, len(canonical.Packets))
	for _, packet := range canonical.Packets {
		names[packet.ID] = packet.Name
	}
	for _, entry := range accepted.Divergences {
		name, ok := names[entry.ID]
		if !ok {
			t.Errorf("accepted divergence %d (%s) has no manifest packet", entry.ID, entry.Name)
			continue
		}
		if name != entry.Name {
			t.Errorf("accepted divergence %d names %q, manifest packet is %q", entry.ID, entry.Name, name)
		}
	}
}

func fixtureManifest(nodes ...manifest.Node) manifest.Manifest {
	fields := make([]manifest.Field, len(nodes))
	for index, node := range nodes {
		fields[index] = manifest.Field{Ordinal: index, Name: "Field", Encode: node, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}
	}
	return manifest.Manifest{
		SchemaVersion: 2,
		Target:        manifest.Target{MinecraftVersion: "1.26.40", ProtocolVersion: 2168},
		Sources:       []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:oracle", MinecraftVersion: "1.26.40", ProtocolVersion: 2168}},
		Packets:       []manifest.Packet{{ID: 1, Name: "Fixture", Direction: manifest.DirectionUnknown, Fields: fields}},
	}
}

func fixtureLock() Lock {
	var lock Lock
	lock.SchemaVersion = LockSchemaVersion
	lock.MinecraftVersion = "1.26.40"
	lock.ProtocolVersion = 2168
	lock.Gophertunnel.Repo = "fixture"
	lock.Gophertunnel.Commit = "0123456789012345678901234567890123456789"
	return lock
}

func TestNBTEncodingIsPartOfTheWireLanguage(t *testing.T) {
	for _, encoding := range []manifest.NBTEncoding{manifest.NBTNetwork, manifest.NBTPersistent, manifest.NBTNetworkCompound} {
		want := manifestExpr(manifest.NBT(encoding))
		for _, sourceEncoding := range []manifest.NBTEncoding{manifest.NBTNetwork, manifest.NBTPersistent, manifest.NBTNetworkCompound} {
			got := sourceOperationExpr(sourceOperation{Kind: "primitive", Code: "nbt_le", Encoding: string(sourceEncoding)})
			if (compareLanguages(want, got) == nil) != (encoding == sourceEncoding) {
				t.Errorf("%s compared with %s", encoding, sourceEncoding)
			}
		}
	}
	for _, op := range []sourceOperation{{Kind: "primitive", Code: "nbt_le"}, {Kind: "primitive", Code: "raw_bytes", Field: "Field.SerialisedOffers"}} {
		if got := sourceOperationExpr(op); got.kind != "unknown" {
			t.Errorf("unknown NBT encoding accepted: %#v", got)
		}
	}
}

func TestDecodeOverridesCannotReceiveAgreement(t *testing.T) {
	decode := manifest.Primitive("u16le")
	plain := manifest.Primitive("u8")
	for _, node := range []manifest.Node{
		manifest.Struct(manifest.Field{Name: "Nested", Encode: plain, Decode: &decode, Symmetry: manifest.Asymmetric}),
		manifest.Union(plain, manifest.Variant{Value: 1, Name: "One", Encode: plain, Decode: &decode}),
	} {
		if got := manifestExpr(node); got.kind != "unknown" {
			t.Errorf("nested decode override accepted: %#v", got)
		}
	}
	m := fixtureManifest(plain)
	m.Packets[0].Fields[0].Decode = &decode
	m.Packets[0].Fields[0].Symmetry = manifest.Asymmetric
	source := extraction{Packets: []sourcePacket{{ID: 1, Operations: []sourceOperation{{Kind: "primitive", Code: "u8"}}}}}
	report := Compare(m, source, fixtureLock(), AcceptedFile{}, "fixture.json")
	if report.Counts.Unresolved != 1 {
		t.Fatalf("decode override classified as %#v", report.Counts)
	}
}

func TestBooleanGuardPolarityAffectsAgreement(t *testing.T) {
	want := manifestExpr(manifest.Optional(manifest.Primitive("u16le")))
	for _, tc := range []struct {
		value   int64
		negated bool
		matches bool
	}{{1, false, true}, {0, false, false}, {1, true, false}, {0, true, true}} {
		got := sourceSequenceExpr([]sourceOperation{
			{Kind: "primitive", Code: "bool", Field: "Has"},
			{Kind: "conditional", CompareTo: "Has", Variants: []sourceVariant{{Value: tc.value, Values: []int64{tc.value}, Negated: tc.negated, Ops: []sourceOperation{{Kind: "primitive", Code: "u16le"}}}}, HasDefault: true},
		})
		if got.kind == "unknown" || (compareLanguages(want, got) == nil) != tc.matches {
			t.Errorf("guard %d negated=%v: %#v", tc.value, tc.negated, got)
		}
	}
	conditional := manifest.Node{Kind: manifest.KindConditional, CompareTo: "Mode", Cases: []manifest.Case{{Value: "1", Encode: []manifest.Node{manifest.Primitive("u8")}}, {Value: "2", Encode: []manifest.Node{manifest.Primitive("u16le")}}}}
	before := manifestExpr(conditional)
	conditional.Cases[0].Value, conditional.Cases[1].Value = "2", "1"
	after := manifestExpr(conditional)
	if before.kind != "unknown" && after.kind != "unknown" && compareLanguages(before, after) == nil {
		t.Fatal("conditional case values were discarded")
	}
	if got := sourceOperationExpr(sourceOperation{Kind: "conditional", CompareTo: "Missing", HasDefault: true, Variants: []sourceVariant{{Value: 1, Ops: []sourceOperation{{Kind: "primitive", Code: "u8"}}}}}); got.kind != "unknown" {
		t.Fatal("unbound predicate accepted")
	}
}

func TestExtractPreservesBooleanPredicatesAndNBTEncodings(t *testing.T) {
	for _, tc := range []struct {
		predicate string
		matches   bool
	}{{"pk.Flag", true}, {"!pk.Flag", false}, {"pk.Flag == false", false}, {"pk.Flag != false", true}, {"pk.Flag && pk.Other", false}} {
		t.Run(tc.predicate, func(t *testing.T) {
			root := t.TempDir()
			dir := filepath.Join(root, "minecraft/protocol/packet")
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			source := `package packet
import "github.com/sandertv/gophertunnel/minecraft/protocol"
import "github.com/sandertv/gophertunnel/minecraft/nbt"
const IDFixture = 1
type Fixture struct { Flag, Other bool; Value uint16; Data map[string]any }
func (*Fixture) ID() uint32 { return IDFixture }
func (pk *Fixture) Marshal(io protocol.IO) {
 io.Bool(&pk.Flag)
 if ` + tc.predicate + ` { io.Uint16(&pk.Value) }
 io.NBT(&pk.Data, nbt.NetworkLittleEndian)
 io.NBT(&pk.Data, nbt.LittleEndian)
}`
			if err := os.WriteFile(filepath.Join(dir, "fixture.go"), []byte(source), 0o644); err != nil {
				t.Fatal(err)
			}
			result, err := Extract(root)
			if err != nil {
				t.Fatal(err)
			}
			want := canonicalPacketExpr(fixtureManifest(manifest.Optional(manifest.Primitive("u16le")), manifest.NBT(manifest.NBTNetwork), manifest.NBT(manifest.NBTPersistent)).Packets[0])
			got := sourceSequenceExpr(result.Packets[0].Operations)
			matches := got.kind != "unknown" && compareLanguages(want, got) == nil
			if matches != tc.matches {
				t.Fatalf("predicate comparison = %v, want %v: %s", matches, tc.matches, expressionKey(got))
			}
		})
	}
}
