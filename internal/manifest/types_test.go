package manifest

import (
	"fmt"
	"strings"
	"testing"
)

func TestValidateAcceptsCompleteV2WireVocabulary(t *testing.T) {
	manifest := Manifest{
		SchemaVersion: 2,
		Target:        Target{MinecraftVersion: "fixture", ProtocolVersion: 2168},
		Sources:       []SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture-2168", Digest: "sha256:fixture"}},
		Packets: []Packet{{
			ID:        1,
			Name:      "Vocabulary",
			Direction: DirectionClientbound,
			Fields: []Field{
				{Ordinal: 0, Name: "Nested", Semantic: "NestedValue", Encode: Optional(Optional(Struct(
					Field{Ordinal: 0, Name: "Values", Encode: Array(Primitive("var_u32"), FixedArray(2, Primitive("u16le"))), Symmetry: Symmetric, Provenance: Provenance{Pins: []string{"fixture"}}},
				))), Symmetry: Symmetric, Provenance: Provenance{Pins: []string{"fixture"}}},
				{Ordinal: 1, Name: "Choice", Encode: Union(
					Primitive("var_u32"),
					Variant{Value: 0, Name: "None", Encode: Void()},
					Variant{Value: 7, Name: "Payload", Encode: Primitive("bytes")},
				), Symmetry: Symmetric, Provenance: Provenance{Pins: []string{"fixture"}}},
				{Ordinal: 2, Name: "Mode", Encode: Enum("u8", EnumValue{Name: "Ready", Value: 4}), Symmetry: Symmetric, Provenance: Provenance{Pins: []string{"fixture"}}},
				{Ordinal: 3, Name: "Compatibility", Encode: Reserved(Primitive("u8")), Reserved: true, Ignored: true, Symmetry: Symmetric, Provenance: Provenance{Pins: []string{"fixture"}}},
				{Ordinal: 4, Name: "Inline", Encode: Sequence(Primitive("u8"), String(Primitive("var_u32"))), Symmetry: Symmetric, Provenance: Provenance{Pins: []string{"fixture"}}},
			},
		}},
	}

	if err := Validate(manifest); err != nil {
		t.Fatalf("Validate: %v", err)
	}

	data, err := MarshalStable(manifest)
	if err != nil {
		t.Fatalf("MarshalStable: %v", err)
	}
	if !strings.Contains(string(data), `"schema_version": 2`) || !strings.Contains(string(data), `"kind": "optional"`) {
		t.Fatalf("stable manifest omitted canonical fields:\n%s", data)
	}
}

func TestValidateRejectsReachableUnresolvedNode(t *testing.T) {
	manifest := Manifest{
		SchemaVersion: 2,
		Target:        Target{MinecraftVersion: "fixture", ProtocolVersion: 2168},
		Sources:       []SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture-2168", Digest: "sha256:fixture"}},
		Packets: []Packet{{
			ID: 1, Name: "Blocked", Direction: DirectionClientbound,
			Fields: []Field{{Ordinal: 0, Name: "Unknown", Encode: Unresolved("dynamic-value", true), Symmetry: Symmetric, Provenance: Provenance{Pins: []string{"fixture"}}}},
		}},
	}

	err := Validate(manifest)
	if err == nil || !strings.Contains(err.Error(), "unresolved") {
		t.Fatalf("Validate error = %v, want reachable unresolved failure", err)
	}
	manifest.Packets[0].Fields[0].Encode.Reachable = false
	if err := Validate(manifest); err == nil || !strings.Contains(err.Error(), "unresolved") {
		t.Fatalf("Validate non-reachable flag error = %v, want fail-closed unresolved failure", err)
	}
}

func TestValidateRejectsUnknownPacketDirection(t *testing.T) {
	value := Manifest{
		SchemaVersion: SchemaVersion,
		Target:        Target{MinecraftVersion: "fixture", ProtocolVersion: 2168},
		Sources:       []SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture-2168", Digest: "sha256:fixture"}},
		Packets:       []Packet{{ID: 1, Name: "UnknownDirection", Direction: DirectionUnknown}},
	}
	if err := Validate(value); err == nil || !strings.Contains(err.Error(), "unknown direction") {
		t.Fatalf("Validate error = %v, want unknown direction failure", err)
	}
}

func TestPrimitiveVarintAndZigZagRemainDistinct(t *testing.T) {
	varint, err := PrimitiveForCode("var_i32")
	if err != nil {
		t.Fatal(err)
	}
	zigzag, err := PrimitiveForCode("zigzag_i32")
	if err != nil {
		t.Fatal(err)
	}
	if varint.Signed != zigzag.Signed || varint.Width != zigzag.Width || varint.ZigZag || !zigzag.ZigZag {
		t.Fatalf("varint=%+v zigzag=%+v", varint, zigzag)
	}
}

func TestValidateRequiresExplicitNBTEncoding(t *testing.T) {
	base := Manifest{
		SchemaVersion: SchemaVersion,
		Target:        Target{MinecraftVersion: "fixture", ProtocolVersion: 2168},
		Sources:       []SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "sha256:fixture"}},
		Packets: []Packet{{ID: 1, Name: "NBT", Direction: DirectionClientbound, Fields: []Field{{
			Ordinal: 0, Name: "Value", Encode: Primitive("nbt_le"), Symmetry: Symmetric,
			Provenance: Provenance{Pins: []string{"fixture"}},
		}}}},
	}
	if err := Validate(base); err == nil || !strings.Contains(err.Error(), "NBT encoding") {
		t.Fatalf("Validate error = %v, want missing NBT encoding", err)
	}

	base.Packets[0].Fields[0].Encode = NBT(NBTNetwork)
	if err := Validate(base); err != nil {
		t.Fatalf("Validate network NBT: %v", err)
	}
	base.Packets[0].Fields[0].Encode = NBT(NBTPersistent)
	if err := Validate(base); err != nil {
		t.Fatalf("Validate persistent NBT: %v", err)
	}
}

func TestValidateRejectsMixedProtocolTargetSources(t *testing.T) {
	manifest := Manifest{
		SchemaVersion: 2,
		Target:        Target{MinecraftVersion: "1.26.40", ProtocolVersion: 2168},
		Sources: []SourcePin{
			{ID: "old", Kind: "mojang", Revision: "old", Digest: "sha256:old", MinecraftVersion: "1.26.40", ProtocolVersion: 2168},
			{ID: "new", Kind: "mojang", Revision: "new", Digest: "sha256:new", MinecraftVersion: "1.26.50", ProtocolVersion: 2169},
		},
		Packets: []Packet{{ID: 1, Name: "P", Direction: DirectionUnknown}},
	}
	if err := Validate(manifest); err == nil || !strings.Contains(err.Error(), "mix") {
		t.Fatalf("Validate error = %v, want mixed-source failure", err)
	}
}

func TestValidateRequiresIndependentAdjudicationEvidence(t *testing.T) {
	value := adjudicatedFixture()
	value.Adjudications[0].Evidence[0].Locator = "https://github.com/example/endstone-docs/blob/rev/status.json"
	if err := Validate(value); err == nil || !strings.Contains(err.Error(), "no evidence independent") {
		t.Fatalf("Validate error = %v, want same-repository evidence rejection", err)
	}

	value.Adjudications[0].Evidence[0].Locator = "https://github.com/example/wire-oracle/blob/rev/capture.bin"
	if err := Validate(value); err != nil {
		t.Fatalf("Validate independent evidence: %v", err)
	}

	value = adjudicatedFixture()
	value.Adjudications[0].Evidence[0].Locator = "https://raw.githubusercontent.com/example/endstone-docs/rev/status.json"
	if err := Validate(value); err == nil || !strings.Contains(err.Error(), "no evidence independent") {
		t.Fatalf("Validate error = %v, want raw GitHub alias rejection", err)
	}

	value = adjudicatedFixture()
	value.Adjudications[0].Evidence[0].Locator = "https://www.github.com/example/endstone-docs/blob/rev/status.json"
	if err := Validate(value); err == nil || !strings.Contains(err.Error(), "no evidence independent") {
		t.Fatalf("Validate error = %v, want www GitHub alias rejection", err)
	}

	value = adjudicatedFixture()
	value.Adjudications[0].Evidence[0].SourceID = ""
	if err := Validate(value); err == nil || !strings.Contains(err.Error(), "unpinned evidence source") {
		t.Fatalf("Validate error = %v, want empty evidence source rejection", err)
	}
}

func TestValidateRejectsEnumOrdinalsOutsidePrimitiveRange(t *testing.T) {
	value := adjudicatedFixture()
	value.Packets[0].Fields[0].Encode = Enum("u8", EnumValue{Value: -1, Name: "Negative"})
	if err := Validate(value); err == nil || !strings.Contains(err.Error(), "does not fit u8") {
		t.Fatalf("Validate error = %v, want negative u8 enum rejection", err)
	}
	value.Packets[0].Fields[0].Encode = Enum("u8", EnumValue{Value: 255, Name: "Maximum"})
	if err := Validate(value); err != nil {
		t.Fatalf("Validate u8 maximum: %v", err)
	}
}

func TestValidateRejectsUnionValuesOutsideControlRange(t *testing.T) {
	value := adjudicatedFixture()
	value.Packets[0].Fields[0].Encode = Union(Primitive("u8"), Variant{Value: -1, Name: "Negative", Encode: Void()})
	if err := Validate(value); err == nil || !strings.Contains(err.Error(), "does not fit u8") {
		t.Fatalf("Validate error = %v, want negative u8 union rejection", err)
	}
}

func TestValidateRejectsDuplicateAdjudicationFingerprints(t *testing.T) {
	value := adjudicatedFixture()
	value.Adjudications[0].Claims = append(value.Adjudications[0].Claims, value.Adjudications[0].Claims[0])
	if err := Validate(value); err == nil || !strings.Contains(err.Error(), "more than once") {
		t.Fatalf("Validate error = %v, want duplicate fingerprint rejection", err)
	}
}

func adjudicatedFixture() Manifest {
	return Manifest{
		SchemaVersion: SchemaVersion,
		Target:        Target{MinecraftVersion: "fixture", ProtocolVersion: 2168},
		Sources: []SourcePin{
			{ID: "endstone", Kind: "endstone", Revision: "rev", Digest: "sha256:endstone", Locator: "https://github.com/example/endstone-docs/tree/rev"},
			{ID: "mojang", Kind: "mojang", Revision: "rev", Digest: "sha256:mojang", Locator: "https://github.com/example/mojang-docs/tree/rev"},
		},
		Packets: []Packet{{ID: 1, Name: "Packet", Direction: DirectionClientbound, Fields: []Field{{
			Ordinal: 0, Name: "Value", Encode: Primitive("u8"), Symmetry: Symmetric,
			Provenance: Provenance{Pins: []string{"endstone"}, Evidence: []Evidence{{SourceID: "endstone", Locator: "https://github.com/example/wire-oracle/blob/rev/capture.bin"}}},
		}}}},
		Adjudications: []Adjudication{{
			ID: "choose-endstone", Target: "Packet.Value", PrePatchContextSHA256: "sha256:context",
			Claims:         []ClaimFingerprint{{SourceID: "endstone", Digest: "sha256:endstone-claim"}, {SourceID: "mojang", Digest: "sha256:mojang-claim"}},
			SelectedSource: "endstone",
			Evidence:       []Evidence{{SourceID: "endstone", Locator: "https://github.com/example/wire-oracle/blob/rev/capture.bin"}},
			Reason:         "independent wire evidence selects the Endstone shape",
		}},
	}
}

func TestValidateEnumAliasesRetainNameAndRangeChecks(t *testing.T) {
	for _, test := range []struct {
		name      string
		values    []EnumValue
		wantError string
	}{
		{name: "aliases", values: []EnumValue{{Name: "Survival", Value: 0}, {Name: "WorldDefault", Value: 0}}},
		{name: "duplicate name", values: []EnumValue{{Name: "Survival", Value: 0}, {Name: "Survival", Value: 1}}, wantError: "unique nonempty name"},
		{name: "empty name", values: []EnumValue{{Name: "", Value: 0}}, wantError: "unique nonempty name"},
		{name: "out of range aliases", values: []EnumValue{{Name: "TooLarge", Value: 256}, {Name: "Alias", Value: 256}}, wantError: "does not fit u8"},
	} {
		t.Run(test.name, func(t *testing.T) {
			value := adjudicatedFixture()
			value.Packets[0].Fields[0].Encode = Enum("u8", test.values...)
			err := Validate(value)
			if test.wantError == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("Validate error = %v, want %q", err, test.wantError)
			}
		})
	}
}

func TestValidateRejectsDuplicateUnionTags(t *testing.T) {
	value := adjudicatedFixture()
	value.Packets[0].Fields[0].Encode = Union(Primitive("u8"),
		Variant{Name: "First", Value: 0, Encode: Void()},
		Variant{Name: "Alias", Value: 0, Encode: Void()},
	)
	if err := Validate(value); err == nil || !strings.Contains(err.Error(), "duplicate or empty explicit identity") {
		t.Fatalf("Validate error = %v, want duplicate union tag rejection", err)
	}
}

// Prose in a pattern field compiles as a regex but matches no real value.
func TestValidateRejectsUnanchoredPatterns(t *testing.T) {
	value := adjudicatedFixture()
	text := String(Primitive("var_u32"))
	text.Constraints = &Constraints{Pattern: "A string in the format of <uuid>_<semver>"}
	value.Packets[0].Fields[0].Encode = text
	if err := Validate(value); err == nil || !strings.Contains(err.Error(), "not anchored") {
		t.Fatalf("Validate error = %v, want prose pattern rejection", err)
	}
	text.Constraints = &Constraints{Pattern: `^\w+:\w+$`}
	value.Packets[0].Fields[0].Encode = text
	if err := Validate(value); err != nil {
		t.Fatalf("Validate anchored pattern: %v", err)
	}
}

func TestMergeEnumVariantsKeepsEverySiteValue(t *testing.T) {
	first := Enum("u8", EnumValue{Name: "raw", Value: 0}, EnumValue{Name: "tip", Value: 5}).Variants
	second := Enum("u8", EnumValue{Name: "chat", Value: 1}, EnumValue{Name: "raw", Value: 0}, EnumValue{Name: "plain", Value: 0}).Variants
	merged, err := MergeEnumVariants("TextPacketType", first, second)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, variant := range merged {
		got = append(got, fmt.Sprintf("%s=%d", variant.Name, variant.Value))
	}
	if want := "raw=0 plain=0 chat=1 tip=5"; strings.Join(got, " ") != want {
		t.Fatalf("merged = %v, want %s", got, want)
	}
	if merged, err := MergeEnumVariants("TextPacketType", first, first[:1]); err != nil || len(merged) != 2 || merged[1].Value != 5 {
		t.Fatalf("merging a subset reordered or grew the enum: %+v, %v", merged, err)
	}
}

// One name bound to two values would silently rename a Rust variant.
func TestMergeEnumVariantsRejectsConflictingName(t *testing.T) {
	first := Enum("u8", EnumValue{Name: "chat", Value: 1}).Variants
	second := Enum("u8", EnumValue{Name: "chat", Value: 7}).Variants
	_, err := MergeEnumVariants("TextPacketType", first, second)
	if err == nil || !strings.Contains(err.Error(), `enum TextPacketType names "chat" both 1 and 7`) {
		t.Fatalf("error = %v, want name conflict", err)
	}
}

func TestSchemaPatternsRequireEveryAlternativeToBeAnchored(t *testing.T) {
	for _, tc := range []struct {
		pattern string
		valid   bool
	}{
		{`^foo|bar$`, false}, {`^(foo|bar)$`, true}, {`^foo\\$`, true}, {`^foo\$`, false},
		{`(?m)^foo$`, false}, {`(?i)^foo$`, true}, {`\A(foo|bar)\z`, true}, {`^foo$|^bar$`, true},
		{`(^foo$)?`, false}, {`(^foo$)+`, true}, {`(^foo$){0,2}`, false},
	} {
		t.Run(tc.pattern, func(t *testing.T) {
			if got := IsSchemaPattern(tc.pattern); got != tc.valid {
				t.Errorf("IsSchemaPattern = %t, want %t", got, tc.valid)
			}
			value := adjudicatedFixture()
			node := String(Primitive("var_u32"))
			node.Constraints = &Constraints{Pattern: tc.pattern}
			value.Packets[0].Fields[0].Encode = node
			if err := Validate(value); (err == nil) != tc.valid {
				t.Errorf("Validate = %v", err)
			}
		})
	}
}
