package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"protocolgen/internal/manifest"
	"protocolgen/internal/naming"
)

func TestBuildIndexSuppressesEveryFlattenedWrapper(t *testing.T) {
	inner := manifest.Node{Kind: manifest.KindStruct, TypeID: "InnerPayload", Fields: []manifest.Field{
		{Name: "First", Encode: manifest.Node{Kind: manifest.KindPrimitive}},
		{Name: "Second", Encode: manifest.Node{Kind: manifest.KindPrimitive}},
	}}
	outer := manifest.Node{Kind: manifest.KindStruct, TypeID: "FooPayload", Fields: []manifest.Field{{Name: "Inner", Encode: inner}}}
	m := manifest.Manifest{Packets: []manifest.Packet{{Name: "FooPacket", Fields: []manifest.Field{{Name: "Payload", Encode: outer}}}}}
	got, err := buildIndex(m, naming.Overlay{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.owners) != 1 || !got.owners[0].Packet || got.owners[0].TypeID != "InnerPayload" || len(got.owners[0].Fields) != 2 {
		t.Fatalf("unexpected owners: %+v", got.owners)
	}
}

func TestConstEvaluatorPreservesTypedValues(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "fixture.go", `package protocol
const (
	Mask = ^uint8(0)
	Wide = ^uint16(0)
	Typed uint8 = 0
	Referenced = ^Typed
	Shift = uint16(1) << 8
	Next = iota
	Last
)
const Alias uint8 = Mask
const Complement = ^Alias
const Bad = missing(1)
`, 0)
	if err != nil {
		t.Fatal(err)
	}
	evaluator := newConstEvaluator(fset, []*ast.File{file})
	values := map[string]int64{}
	for _, declaration := range file.Decls {
		for _, value := range evaluator.group(declaration.(*ast.GenDecl)) {
			values[value.Name] = value.Value
		}
	}
	for name, want := range map[string]int64{"Mask": 255, "Wide": 65535, "Referenced": 255, "Shift": 256, "Next": 5, "Last": 6, "Complement": 0} {
		if got, ok := values[name]; !ok || got != want {
			t.Errorf("%s = %d (%v), want %d", name, got, ok, want)
		}
	}
	if value, exists := values["Bad"]; exists {
		t.Fatalf("unsupported call accepted as %d", value)
	}
}
