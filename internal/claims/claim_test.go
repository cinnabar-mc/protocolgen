package claims

import (
	"testing"

	"protocolgen/internal/manifest"
)

func TestFingerprintIncludesCompleteClaimContext(t *testing.T) {
	claim := Claim{
		SourceID: "endstone",
		Locator:  "packets/Vocabulary.field",
		PacketID: 1, PacketName: "Vocabulary", FieldPath: "Vocabulary.Value", Ordinal: 0, Name: "Value",
		Encode: manifest.Primitive("u8"), Symmetry: manifest.Symmetric,
	}
	first, err := Fingerprint(claim)
	if err != nil {
		t.Fatalf("Fingerprint: %v", err)
	}
	claim.Name = "ChangedName"
	second, err := Fingerprint(claim)
	if err != nil {
		t.Fatalf("Fingerprint changed claim: %v", err)
	}
	if first == second {
		t.Fatalf("claim fingerprint ignored semantic context")
	}
}

func TestWireFingerprintIgnoresSourceMetadataAndConstraints(t *testing.T) {
	minimum := 1.0
	left := Claim{
		SourceID: "mojang",
		Locator:  "old.json",
		PacketID: 7, PacketName: "ExamplePacket", FieldPath: "ExamplePacket.Value", Ordinal: 0, Name: "Value",
		Semantic: "Old Value", Encode: manifest.Primitive("var_u32"), Symmetry: manifest.Symmetric,
	}
	left.Encode.Constraints = &manifest.Constraints{Minimum: &minimum}
	right := left
	right.SourceID = "endstone"
	right.Locator = "packets/ExamplePacket.json"
	right.Semantic = "Renamed Value"
	right.Encode.Constraints = nil

	leftFingerprint, err := WireFingerprint(left)
	if err != nil {
		t.Fatal(err)
	}
	rightFingerprint, err := WireFingerprint(right)
	if err != nil {
		t.Fatal(err)
	}
	if leftFingerprint != rightFingerprint {
		t.Fatalf("wire fingerprints differ: %s != %s", leftFingerprint, rightFingerprint)
	}
}

func TestWireFingerprintRetainsRecursiveReferences(t *testing.T) {
	for _, nested := range []bool{false, true} {
		makeClaim := func(target string) Claim {
			node := manifest.Node{Kind: manifest.KindRecursive, TypeID: target}
			if nested {
				node = manifest.Optional(node)
			}
			return Claim{Encode: node}
		}
		left, err := WireFingerprint(makeClaim("A"))
		if err != nil {
			t.Fatal(err)
		}
		right, err := WireFingerprint(makeClaim("B"))
		if err != nil {
			t.Fatal(err)
		}
		if left == right {
			t.Fatalf("recursive references collapsed (nested=%v)", nested)
		}
		leftField, err := FieldWireFingerprint(1, manifest.DirectionClientbound, manifest.Field{Encode: makeClaim("A").Encode})
		if err != nil {
			t.Fatal(err)
		}
		rightField, err := FieldWireFingerprint(1, manifest.DirectionClientbound, manifest.Field{Encode: makeClaim("B").Encode})
		if err != nil {
			t.Fatal(err)
		}
		if leftField == rightField {
			t.Fatal("field recursive references collapsed")
		}
	}
}
