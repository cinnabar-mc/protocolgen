package gophertunneloracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"protocolgen/internal/manifest"
)

// extractFixture parses a small packet package without compiling the oracle source.
func extractFixture(t *testing.T, source string) (extraction, error) {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "minecraft/protocol/packet")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "fixture.go"), []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	return Extract(root)
}

func TestRecursiveManifestAndSourceUseEqualExpansionDepth(t *testing.T) {
	source, err := extractFixture(t, `package packet
import "github.com/sandertv/gophertunnel/minecraft/protocol"
const IDFixture = 1
type Fixture struct {Root Node}
func (*Fixture) ID() uint32 {return IDFixture}
func (p *Fixture) Marshal(io protocol.IO) {protocol.Single(io,&p.Root)}
type Node struct {Value uint8; Children []Node}
func (n *Node) Marshal(io protocol.IO) {io.Uint8(&n.Value);protocol.Slice(io,&n.Children)}
`)
	if err != nil {
		t.Fatal(err)
	}
	node := manifest.Struct(manifest.Field{Ordinal: 0, Name: "Value", Encode: manifest.Primitive("u8")}, manifest.Field{Ordinal: 1, Name: "Children", Encode: manifest.Array(manifest.Primitive("var_u32"), manifest.Recursive("Node"))})
	node.TypeID = "Node"
	report := Compare(fixtureManifest(node), source, fixtureLock(), emptyAccepted(), "fixture.json")
	if report.Counts.Agreement != 1 {
		t.Fatalf("recursive comparison: %+v", report.Packets)
	}
	node.Fields[0].Encode = manifest.Primitive("u16le")
	report = Compare(fixtureManifest(node), source, fixtureLock(), emptyAccepted(), "fixture.json")
	if report.Counts.Divergence != 1 {
		t.Fatalf("changed recursive body not detected: %+v", report.Packets)
	}
}

func TestHoistingDiscardsClosedSelectorRejections(t *testing.T) {
	primitive := manifest.Primitive("u8")
	variants := []manifest.Variant{{Value: 0, Name: "Zero", Encode: manifest.Void()}, {Value: 1, Name: "One", Encode: manifest.Void()}}
	got := sourceSequenceExpr([]sourceOperation{
		{Kind: "primitive", Code: "u8", Field: "Mode"},
		{Kind: "switch", CompareTo: "Mode", Variants: []sourceVariant{{Value: 0, Discriminant: true}, {Value: 1, Discriminant: true}}},
		{Kind: "conditional", CompareTo: "Mode", Variants: []sourceVariant{{Value: 2, Discriminant: true, Ops: []sourceOperation{{Kind: "primitive", Code: "u16le"}}}}, HasDefault: true},
	})
	if witness := compareLanguages(manifestExpr(manifest.Union(primitive, variants...)), got); witness != nil {
		t.Fatalf("rejected path survived: %+v", witness)
	}
	variants = append(variants, manifest.Variant{Value: 2, Name: "Two", Encode: manifest.Primitive("u16le")})
	if compareLanguages(manifestExpr(manifest.Union(primitive, variants...)), got) == nil {
		t.Fatal("invalid tag agreed")
	}
}

func TestConstantsPreserveParallelDeclarationsAndRepetitions(t *testing.T) {
	source, err := extractFixture(t, `package packet
const (
 IDOne, IDTwo = 1 + iota*2, 2 + iota*2
 IDThree, IDFour
)
type One struct{Value uint8}; type Two struct{Value uint16}; type Three struct{};type Four struct{}
func (*One) ID()uint32{return IDOne};func (*Two)ID()uint32{return IDTwo};func (*Three)ID()uint32{return IDThree};func (*Four)ID()uint32{return IDFour}
func(p *One)Marshal(io IO){io.Uint8(&p.Value)};func(p *Two)Marshal(io IO){io.Uint16(&p.Value)};func(*Three)Marshal(io IO){};func(*Four)Marshal(io IO){}
`)
	if err != nil {
		t.Fatal(err)
	}
	if len(source.Packets) != 4 {
		t.Fatalf("packets: %+v", source.Packets)
	}
	for index, packet := range source.Packets {
		if packet.ID != uint32(index+1) {
			t.Errorf("packet %s ID = %d", packet.Name, packet.ID)
		}
	}
	if len(source.Packets[1].Operations) != 1 || source.Packets[1].Operations[0].Code != "u16le" {
		t.Fatal("second packet lost its marshal")
	}
}

func TestExtractionRejectsDuplicatePacketIDs(t *testing.T) {
	_, err := extractFixture(t, `package packet
const IDOne, IDTwo = 1,1
type One struct{};type Two struct{}
func(*One)ID()uint32{return IDOne};func(*Two)ID()uint32{return IDTwo}
func(*One)Marshal(io IO){};func(*Two)Marshal(io IO){}
`)
	if err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("duplicate packet ID: %v", err)
	}
}

func TestByteArraysNormalizeScalarGroupingAndSignedness(t *testing.T) {
	for _, tc := range []struct {
		node      manifest.Node
		operation sourceOperation
	}{
		{manifest.FixedArray(16, manifest.FixedArray(16, manifest.Primitive("u8"))), sourceOperation{Kind: "fixed_array", Length: 256, Element: []sourceOperation{{Kind: "primitive", Code: "u8"}}}},
		{manifest.FixedArray(16, manifest.Primitive("u8")), sourceOperation{Kind: "fixed_array", Length: 16, Element: []sourceOperation{{Kind: "primitive", Code: "i8"}}}},
	} {
		if witness := compareLanguages(manifestExpr(tc.node), sourceOperationExpr(tc.operation)); witness != nil {
			t.Fatalf("byte grouping diverges: %+v", witness)
		}
	}
}
