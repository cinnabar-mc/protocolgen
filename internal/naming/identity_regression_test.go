package naming

import (
	"protocolgen/internal/manifest"
	"testing"
)

func TestUnnamedUnionIdentityIncludesWireShapes(t *testing.T) {
	first := manifest.Union(manifest.Primitive("u8"), manifest.Variant{Value: 0, Name: "Family::Value", Encode: manifest.Primitive("u8")})
	for _, decodeOnly := range []bool{false, true} {
		second := manifest.Union(manifest.Primitive("u8"), manifest.Variant{Value: 0, Name: "Family::Value", Encode: manifest.Primitive("u8")})
		changed := manifest.Primitive("u16le")
		if decodeOnly {
			second.Variants[0].Decode = &changed
		} else {
			second.Variants[0].Encode = changed
		}
		if IdentityKey(first) == IdentityKey(second) {
			t.Error("distinct payloads share identity")
		}
		r := NewResolver(Overlay{})
		if _, err := r.Resolve(first, "First", PublicTypeName); err != nil {
			t.Fatal(err)
		}
		if _, err := r.Resolve(second, "Second", PublicTypeName); err == nil {
			t.Error("incompatible union reused public name")
		}
	}
}

func TestQualifiedArtifactsRequireReview(t *testing.T) {
	for _, name := range []string{"Foo::Empty0", "Foo::T12"} {
		if !LooksLikeArtifact(name) {
			t.Errorf("missed %q", name)
		}
		m := manifest.Manifest{Packets: []manifest.Packet{{Fields: []manifest.Field{{Encode: manifest.Node{Kind: manifest.KindStruct, TypeID: name}}}}}}
		if err := ValidateRequiredEntries(m, Overlay{}); err == nil {
			t.Errorf("accepted unreviewed %q", name)
		}
	}
}

func TestDecodeVariantArtifactsRequireReview(t *testing.T) {
	for _, location := range []string{"packet", "field", "variant", "case"} {
		t.Run(location, func(t *testing.T) {
			decoded := manifest.Enum("u8", manifest.EnumValue{Value: 0, Name: "Empty0"})
			decoded.TypeID = "Mode"
			field := manifest.Field{Encode: manifest.Void()}
			switch location {
			case "packet":
				field.Decode = &decoded
			case "field":
				field.Encode = manifest.Struct(manifest.Field{Encode: manifest.Void(), Decode: &decoded})
			case "variant":
				field.Encode = manifest.Union(manifest.Primitive("u8"), manifest.Variant{Value: 0, Name: "Value", Encode: manifest.Void(), Decode: &decoded})
			case "case":
				field.Encode = manifest.Node{Kind: manifest.KindConditional, Cases: []manifest.Case{{Encode: []manifest.Node{manifest.Void()}, Decode: []manifest.Node{decoded}}}}
			}
			m := manifest.Manifest{Packets: []manifest.Packet{{Fields: []manifest.Field{field}}}}
			if err := ValidateRequiredEntries(m, Overlay{}); err == nil {
				t.Fatal("accepted unreviewed decode variant")
			}
			if err := ValidateRequiredEntries(m, Overlay{Names: map[string]string{"Mode": "Mode"}}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestUnnamedUnionIdentityIgnoresPayloadProvenance(t *testing.T) {
	makeUnion := func(pin string) manifest.Node {
		payload := manifest.Struct(manifest.Field{Name: "Value", Encode: manifest.Primitive("u8"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{pin}}})
		return manifest.Union(manifest.Primitive("u8"), manifest.Variant{Value: 0, Name: "Family::Value", Encode: payload})
	}
	first, second := makeUnion("first"), makeUnion("second")
	if IdentityKey(first) != IdentityKey(second) {
		t.Fatal("provenance changed the wire identity")
	}
	resolver := NewResolver(Overlay{})
	if _, err := resolver.Resolve(first, "First", PublicTypeName); err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.Resolve(second, "Second", PublicTypeName); err != nil {
		t.Fatal(err)
	}
}

func TestUnnamedUnionIdentityPreservesExactPayloadTags(t *testing.T) {
	makeUnion := func(tag int64) manifest.Node {
		payload := manifest.Enum("u64le", manifest.EnumValue{Value: tag, Name: "Value"})
		return manifest.Union(manifest.Primitive("u8"), manifest.Variant{Value: 0, Name: "Family::Value", Encode: payload})
	}
	if IdentityKey(makeUnion(1<<53)) == IdentityKey(makeUnion(1<<53+1)) {
		t.Fatal("distinct payload tags share an identity")
	}
}

func TestUnnamedUnionIdentityKeepsNestedNamedReferences(t *testing.T) {
	makeUnion := func(value string) manifest.Node {
		enum := manifest.Enum("u8", manifest.EnumValue{Value: 0, Name: value})
		enum.TypeID = "Mode"
		payload := manifest.Struct(manifest.Field{Name: "Mode", Encode: enum, Symmetry: manifest.Symmetric})
		return manifest.Union(manifest.Primitive("u8"), manifest.Variant{Value: 0, Name: "Family::Value", Encode: payload})
	}
	if IdentityKey(makeUnion("First")) != IdentityKey(makeUnion("Second")) {
		t.Fatal("a shared named type changed its enclosing identity")
	}
}
