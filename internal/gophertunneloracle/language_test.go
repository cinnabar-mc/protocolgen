package gophertunneloracle

import (
	"testing"

	"protocolgen/internal/manifest"
)

func TestFiniteUnionDoesNotCoverAnOracleDefault(t *testing.T) {
	packet := fixtureManifest(manifest.Union(manifest.Primitive("var_u32"),
		manifest.Variant{Value: 0, Name: "Zero", Encode: manifest.Void()},
		manifest.Variant{Value: 1, Name: "One", Encode: manifest.Void()},
	))
	source := extraction{Packets: []sourcePacket{{ID: 1, Operations: []sourceOperation{
		{Kind: "primitive", Code: "var_u32", Field: "Mode"},
		{Kind: "switch", CompareTo: "Mode", Variants: []sourceVariant{{Value: 0, Values: []int64{0}, Discriminant: true}}, HasDefault: true},
	}}}}
	report := Compare(packet, source, fixtureLock(), AcceptedFile{}, "fixture.json")
	if report.Counts.Divergence != 1 {
		t.Fatalf("default accepts tags beyond 0 and 1: %#v", report.Counts)
	}
	witness := compareLanguages(canonicalPacketExpr(packet.Packets[0]), sourceSequenceExpr(source.Packets[0].Operations))
	if len(witness.gophertunnel) == 0 || witness.gophertunnel[0].Token != "VARIANT:!0,1" {
		t.Fatalf("missing residual tag domain: %#v", witness)
	}
	source.Packets[0].Operations[1].HasDefault = false
	source.Packets[0].Operations[1].Variants = append(source.Packets[0].Operations[1].Variants, sourceVariant{Value: 1, Values: []int64{1}, Discriminant: true})
	report = Compare(packet, source, fixtureLock(), AcceptedFile{}, "fixture.json")
	if report.Counts.Agreement != 1 {
		t.Fatalf("matching finite tag domains disagree: %#v", report.Counts)
	}
}

func TestWildcardVariantLanguagesRetainTheirEntireDomains(t *testing.T) {
	zero := atomExpr(variantAtom(0, "Mode"))
	one := atomExpr(variantAtom(1, "Mode"))
	all := atomExpr(wildcardAtom(nil, "Mode"))
	exceptZero := atomExpr(wildcardAtom([]string{"0"}, "Mode"))
	exceptOne := atomExpr(wildcardAtom([]string{"1"}, "Mode"))
	exceptBoth := atomExpr(wildcardAtom([]string{"0", "1"}, "Mode"))
	u8 := atomExpr(primitiveAtom("u8", "Value", "u8"))
	u16 := atomExpr(primitiveAtom("u16le", "Value", "u16le"))
	for _, tc := range []struct {
		name        string
		left, right shapeExpr
		equal       bool
	}{
		{"finite versus default", altExpr(zero, one), altExpr(zero, exceptZero), false},
		{"different exclusions", exceptZero, exceptBoth, false},
		{"all versus partition", all, altExpr(zero, exceptZero), true},
		{"overlapping defaults", all, altExpr(exceptZero, exceptOne), true},
		{"overlapping payloads", altExpr(concatExpr(all, u8), concatExpr(one, u16)), altExpr(concatExpr(exceptOne, u8), concatExpr(one, altExpr(u8, u16))), true},
		{"default payload matters", altExpr(concatExpr(zero, u8), concatExpr(exceptZero, u16)), altExpr(concatExpr(zero, u8), concatExpr(one, u16)), false},
		{"explicit payload matters", altExpr(concatExpr(zero, u8), concatExpr(exceptZero, u16)), concatExpr(all, u16), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, pair := range [][2]shapeExpr{{tc.left, tc.right}, {tc.right, tc.left}} {
				if equal := compareLanguages(pair[0], pair[1]) == nil; equal != tc.equal {
					t.Fatalf("language equality = %v, want %v", equal, tc.equal)
				}
			}
		})
	}
}
