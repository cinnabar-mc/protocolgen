package reconcile

import (
	"strings"
	"testing"

	"protocolgen/internal/claims"
	"protocolgen/internal/direction"
	"protocolgen/internal/manifest"
	"protocolgen/internal/nbtencoding"
)

func TestReconcileRejectsUnfingerprintedSourceDisagreement(t *testing.T) {
	target := manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}
	results := []claims.Result{
		{Pin: sourcePin("endstone"), Target: target, Claims: []claims.Claim{testClaim("endstone", manifest.Primitive("u8"))}},
		{Pin: sourcePin("mojang"), Target: target, Claims: []claims.Claim{testClaim("mojang", manifest.Primitive("u16le"))}},
	}
	_, err := Reconcile(target, results, nil)
	if err == nil || !strings.Contains(err.Error(), "adjudication") {
		t.Fatalf("Reconcile error = %v, want missing adjudication", err)
	}
}

func TestReconcileRejectsStaleAdjudicationFingerprint(t *testing.T) {
	target := manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}
	left := testClaim("endstone", manifest.Primitive("u8"))
	right := testClaim("mojang", manifest.Primitive("u16le"))
	context, err := claims.ContextFingerprint(target, []claims.Claim{left, right})
	if err != nil {
		t.Fatalf("ContextFingerprint: %v", err)
	}
	left.Name = "MutatedAfterReview"
	adj := manifest.Adjudication{
		ID: "choose-endstone", Target: left.FieldPath, PrePatchContextSHA256: context,
		Claims:         []manifest.ClaimFingerprint{{SourceID: "endstone", Digest: mustFingerprint(t, testClaim("endstone", manifest.Primitive("u8")))}, {SourceID: "mojang", Digest: mustFingerprint(t, right)}},
		SelectedSource: "endstone", Evidence: []manifest.Evidence{{SourceID: "endstone", Locator: "https://github.com/example/wire-oracle/blob/rev/capture.bin"}}, Reason: "fixture evidence",
	}
	_, err = Reconcile(target, []claims.Result{{Pin: sourcePin("endstone"), Target: target, Claims: []claims.Claim{left}}, {Pin: sourcePin("mojang"), Target: target, Claims: []claims.Claim{right}}}, []manifest.Adjudication{adj})
	if err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("Reconcile error = %v, want stale fingerprint failure", err)
	}
}

func TestReconcileMergesIdenticalClaimsAndOnlyPinsSources(t *testing.T) {
	target := manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}
	claimA := testClaim("endstone", manifest.Primitive("u8"))
	claimB := testClaim("mojang", manifest.Primitive("u8"))
	m, err := Reconcile(target, []claims.Result{{Pin: sourcePin("endstone"), Target: target, Claims: []claims.Claim{claimA}}, {Pin: sourcePin("mojang"), Target: target, Claims: []claims.Claim{claimB}}}, nil)
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	field := m.Packets[0].Fields[0]
	if len(field.Provenance.Pins) != 2 || len(field.Provenance.Evidence) != 0 {
		t.Fatalf("provenance = %+v, want two pins and no detailed evidence", field.Provenance)
	}
}

func TestReconcileWithDirectionsAppliesReviewedOverlay(t *testing.T) {
	target := manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}
	table := direction.Table{
		SchemaVersion: direction.SchemaVersion,
		Target:        target,
		Source: direction.Source{
			Repository: "https://github.com/example/gophertunnel",
			Revision:   "0123456789012345678901234567890123456789",
			Locator:    "https://github.com/example/gophertunnel/blob/0123456789012345678901234567890123456789/minecraft/protocol/packet/pool.go",
			SHA256:     "sha256:0123456789012345678901234567890123456789012345678901234567890123",
		},
		Packets: []direction.Entry{{
			ID: 1, Name: "Vocabulary", Direction: direction.DirectionBoth,
			Evidence: direction.Evidence{Locator: "https://github.com/example/gophertunnel/blob/0123456789012345678901234567890123456789/minecraft/protocol/packet/pool.go", Summary: "The packet ID is registered in both direction pools."},
		}},
	}
	m, err := ReconcileWithDirections(target, []claims.Result{
		{Pin: sourcePin("endstone"), Target: target, Claims: []claims.Claim{testClaim("endstone", manifest.Primitive("u8"))}},
		{Pin: sourcePin("mojang"), Target: target, Claims: []claims.Claim{testClaim("mojang", manifest.Primitive("u8"))}},
	}, nil, table)
	if err != nil {
		t.Fatalf("ReconcileWithDirections: %v", err)
	}
	if got := m.Packets[0].Direction; got != manifest.DirectionBidirectional {
		t.Fatalf("direction = %q, want %q", got, manifest.DirectionBidirectional)
	}
}

func TestReconcileRequiresAdjudicationWhenOnlyOneSourceIsConcrete(t *testing.T) {
	target := manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}
	unresolved := testClaim("mojang", manifest.Unresolved("source omits the wire shape", true))
	concrete := testClaim("endstone", manifest.Primitive("u8"))

	_, err := Reconcile(target, []claims.Result{
		{Pin: sourcePin("mojang"), Target: target, Claims: []claims.Claim{unresolved}},
		{Pin: sourcePin("endstone"), Target: target, Claims: []claims.Claim{concrete}},
	}, nil)
	if err == nil || !strings.Contains(err.Error(), "adjudication") {
		t.Fatalf("Reconcile error = %v, want missing adjudication", err)
	}
}

func TestMergeNodeFillsMissingTextSemantics(t *testing.T) {
	partial := manifest.String(manifest.Primitive("var_u32"))
	partial.Encoding = ""
	partial.Representation = ""
	concrete := manifest.String(manifest.Primitive("var_u32"))
	merged, ok := mergeNode(partial, concrete)
	if !ok || merged.Encoding != "utf8" || merged.Representation != "text" {
		t.Fatalf("mergeNode = %#v, %v; want concrete text semantics", merged, ok)
	}
}

func TestMergeNodeRetainsComplementaryConstraints(t *testing.T) {
	unconstrained := manifest.String(manifest.Primitive("var_u32"))
	constrained := manifest.String(manifest.Primitive("var_u32"))
	constrained.Constraints = &manifest.Constraints{MinLength: pointerTo(uint64(1)), MaxLength: pointerTo(uint64(65536))}
	merged, ok := mergeNode(unconstrained, constrained)
	if !ok || merged.Constraints == nil || merged.Constraints.MinLength == nil || *merged.Constraints.MinLength != 1 || merged.Constraints.MaxLength == nil || *merged.Constraints.MaxLength != 65536 {
		t.Fatalf("mergeNode = %#v, %v; want both string constraints", merged, ok)
	}
}

func TestMergeNodeUsesLooserSourceLimits(t *testing.T) {
	left := manifest.String(manifest.Primitive("var_u32"))
	left.Constraints = &manifest.Constraints{MinLength: pointerTo(uint64(2)), MaxLength: pointerTo(uint64(512))}
	right := manifest.String(manifest.Primitive("var_u32"))
	right.Constraints = &manifest.Constraints{MinLength: pointerTo(uint64(1)), MaxLength: pointerTo(uint64(1000))}
	merged, ok := mergeNode(left, right)
	if !ok || merged.Constraints == nil || merged.Constraints.MinLength == nil || *merged.Constraints.MinLength != 1 || merged.Constraints.MaxLength == nil || *merged.Constraints.MaxLength != 1000 {
		t.Fatalf("mergeNode = %#v, %v; want looser limits 1..1000", merged, ok)
	}
}

func pointerTo[T any](value T) *T { return &value }

func TestReconcileRequiresAdjudicationWhenAnotherSourceHasNestedUnresolvedShape(t *testing.T) {
	target := manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}
	incomplete := manifest.Node{Kind: manifest.KindStruct, Fields: []manifest.Field{{Ordinal: 0, Name: "Mode", Encode: manifest.Unresolved("missing enum values", true), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"mojang"}}}}}
	complete := manifest.Node{Kind: manifest.KindStruct, Fields: []manifest.Field{{Ordinal: 0, Name: "Mode", Encode: manifest.Primitive("u8"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"endstone"}}}}}

	_, err := Reconcile(target, []claims.Result{
		{Pin: sourcePin("mojang"), Target: target, Claims: []claims.Claim{testClaim("mojang", incomplete)}},
		{Pin: sourcePin("endstone"), Target: target, Claims: []claims.Claim{testClaim("endstone", complete)}},
	}, nil)
	if err == nil || !strings.Contains(err.Error(), "adjudication") {
		t.Fatalf("Reconcile error = %v, want missing adjudication", err)
	}
}

func TestReconcileRejectsComplementaryClaimsWithoutIndependentFullShapes(t *testing.T) {
	target := manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}
	left := manifest.Node{Kind: manifest.KindStruct, Fields: []manifest.Field{
		{Ordinal: 0, Name: "First", Encode: manifest.Primitive("u8"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"endstone"}}},
		{Ordinal: 1, Name: "Second", Encode: manifest.Unresolved("missing", true), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"endstone"}}},
	}}
	right := manifest.Node{Kind: manifest.KindStruct, Fields: []manifest.Field{
		{Ordinal: 0, Name: "First", Encode: manifest.Unresolved("missing", true), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"mojang"}}},
		{Ordinal: 1, Name: "Second", Encode: manifest.Primitive("u16le"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"mojang"}}},
	}}

	_, err := Reconcile(target, []claims.Result{
		{Pin: sourcePin("endstone"), Target: target, Claims: []claims.Claim{testClaim("endstone", left)}},
		{Pin: sourcePin("mojang"), Target: target, Claims: []claims.Claim{testClaim("mojang", right)}},
	}, nil)
	if err == nil || !strings.Contains(err.Error(), "adjudication") {
		t.Fatalf("Reconcile error = %v, want missing adjudication", err)
	}
}

func TestReconcilePreservesSourcePinWhenAllClaimsAreUnresolved(t *testing.T) {
	target := manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}
	_, err := Reconcile(target, []claims.Result{{
		Pin: sourcePin("endstone"), Target: target,
		Claims: []claims.Claim{testClaim("endstone", manifest.Unresolved("missing selector", true))},
	}}, nil)
	if err == nil || !strings.Contains(err.Error(), "adjudication") {
		t.Fatalf("Reconcile error = %v, want missing adjudication", err)
	}
}

func TestReconcileAcceptsFingerprintedSingletonWithIndependentEvidence(t *testing.T) {
	target := manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}
	claim := testClaim("endstone", manifest.Primitive("u8"))
	context, err := claims.ContextFingerprint(target, []claims.Claim{claim})
	if err != nil {
		t.Fatalf("ContextFingerprint: %v", err)
	}
	adjudication := manifest.Adjudication{
		ID: "confirm-singleton", Target: claim.FieldPath, PrePatchContextSHA256: context,
		Claims:         []manifest.ClaimFingerprint{{SourceID: claim.SourceID, Digest: mustFingerprint(t, claim)}},
		SelectedSource: claim.SourceID,
		Evidence:       []manifest.Evidence{{SourceID: claim.SourceID, Locator: "https://github.com/example/wire-oracle/blob/rev/capture.bin"}},
		Reason:         "an independent wire capture confirms the only available source claim",
	}
	m, err := Reconcile(target, []claims.Result{{Pin: sourcePin("endstone"), Target: target, Claims: []claims.Claim{claim}}}, []manifest.Adjudication{adjudication})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	field := m.Packets[0].Fields[0]
	if len(field.Provenance.Pins) != 1 || len(field.Provenance.Evidence) != 1 || len(m.Adjudications) != 1 {
		t.Fatalf("field provenance = %+v, adjudications = %+v", field.Provenance, m.Adjudications)
	}
}

func TestReconcileAdjudicationPreservesCompatibleIncompleteSemanticMetadata(t *testing.T) {
	target := manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}
	complete := testClaim("endstone", manifest.Node{Kind: manifest.KindStruct, Fields: []manifest.Field{
		{Ordinal: 0, Name: "Mode", Encode: manifest.Primitive("u8"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"endstone"}}},
	}})
	incomplete := testClaim("mojang", manifest.Node{Kind: manifest.KindStruct, Semantic: "NamedData", TypeID: "NamedData.json#", Fields: []manifest.Field{
		{Ordinal: 0, Name: "Mode", Encode: manifest.Unresolved("missing enum", true), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"mojang"}}},
	}})
	group := []claims.Claim{complete, incomplete}
	context, err := claims.ContextFingerprint(target, group)
	if err != nil {
		t.Fatalf("ContextFingerprint: %v", err)
	}
	adjudication := manifest.Adjudication{
		ID: "confirm-complete-wire", Target: complete.FieldPath, PrePatchContextSHA256: context,
		Claims: []manifest.ClaimFingerprint{
			{SourceID: complete.SourceID, Digest: mustFingerprint(t, complete)},
			{SourceID: incomplete.SourceID, Digest: mustFingerprint(t, incomplete)},
		},
		SelectedSource: complete.SourceID,
		Evidence:       []manifest.Evidence{{SourceID: complete.SourceID, Locator: "https://github.com/example/wire-oracle/blob/rev/serializer.go"}},
		Reason:         "the independent serializer confirms the complete wire claim",
	}
	m, err := Reconcile(target, []claims.Result{
		{Pin: sourcePin("endstone"), Target: target, Claims: []claims.Claim{complete}},
		{Pin: sourcePin("mojang"), Target: target, Claims: []claims.Claim{incomplete}},
	}, []manifest.Adjudication{adjudication})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	field := m.Packets[0].Fields[0]
	if field.Encode.Kind != manifest.KindStruct || field.Encode.TypeID != "NamedData.json#" || field.Encode.Fields[0].Encode.Kind != manifest.KindPrimitive {
		t.Fatalf("field = %#v, want complete wire with compatible semantic metadata", field)
	}
	if got := field.Provenance.Pins; len(got) != 1 || got[0] != "endstone" {
		t.Fatalf("pins = %v, want only adjudicated complete source", got)
	}
}

func TestReconcileAdjudicationPreservesMetadataAcrossWireRepresentationChoice(t *testing.T) {
	target := manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}
	bytesClaim := testClaim("endstone", manifest.Bytes(manifest.Primitive("var_u32")))
	prefix := manifest.Primitive("var_u32")
	stringClaim := testClaim("mojang", manifest.Node{
		Kind: manifest.KindString, Semantic: "ConnectionRequest", TypeID: "ConnectionRequest.json#",
		Prefix: &prefix, Encoding: "utf8", Representation: "text",
	})
	group := []claims.Claim{bytesClaim, stringClaim}
	context, err := claims.ContextFingerprint(target, group)
	if err != nil {
		t.Fatalf("ContextFingerprint: %v", err)
	}
	adjudication := manifest.Adjudication{
		ID: "binary-representation", Target: bytesClaim.FieldPath, PrePatchContextSHA256: context,
		Claims: []manifest.ClaimFingerprint{
			{SourceID: bytesClaim.SourceID, Digest: mustFingerprint(t, bytesClaim)},
			{SourceID: stringClaim.SourceID, Digest: mustFingerprint(t, stringClaim)},
		},
		SelectedSource: bytesClaim.SourceID,
		Evidence:       []manifest.Evidence{{SourceID: bytesClaim.SourceID, Locator: "https://github.com/example/wire-oracle/blob/rev/serializer.go"}},
		Reason:         "the serializer identifies an arbitrary byte payload",
	}
	m, err := Reconcile(target, []claims.Result{
		{Pin: sourcePin("endstone"), Target: target, Claims: []claims.Claim{bytesClaim}},
		{Pin: sourcePin("mojang"), Target: target, Claims: []claims.Claim{stringClaim}},
	}, []manifest.Adjudication{adjudication})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	field := m.Packets[0].Fields[0]
	if field.Encode.Kind != manifest.KindBytes || field.Encode.TypeID != "ConnectionRequest.json#" {
		t.Fatalf("field = %#v, want byte wire representation with semantic TypeID", field)
	}
}

func TestReconcileReportsEveryProvenanceGap(t *testing.T) {
	target := manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}
	first := testClaim("endstone", manifest.Primitive("u8"))
	second := testClaim("endstone", manifest.Primitive("u16le"))
	second.PacketID = 2
	second.PacketName = "SecondPacket"
	second.FieldPath = "SecondPacket.Value"

	_, err := Reconcile(target, []claims.Result{{
		Pin: sourcePin("endstone"), Target: target, Claims: []claims.Claim{first, second},
	}}, nil)
	if err == nil || !strings.Contains(err.Error(), "2 provenance gap(s)") || !strings.Contains(err.Error(), first.FieldPath) || !strings.Contains(err.Error(), second.FieldPath) {
		t.Fatalf("Reconcile error = %v, want both provenance gaps", err)
	}
}

func TestReconcileRejectsDuplicateClaimsFromOneSource(t *testing.T) {
	target := manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}
	claim := testClaim("endstone", manifest.Primitive("u8"))
	_, err := Reconcile(target, []claims.Result{{
		Pin: sourcePin("endstone"), Target: target, Claims: []claims.Claim{claim, claim},
	}}, nil)
	if err == nil || !strings.Contains(err.Error(), "duplicate claims") {
		t.Fatalf("Reconcile error = %v, want duplicate same-source rejection", err)
	}
}

func TestReconcileMergesByteEquivalentEnumAndPrimitiveClaims(t *testing.T) {
	target := manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}
	primitiveClaim := testClaim("endstone", manifest.Primitive("u8"))
	enumClaim := testClaim("mojang", manifest.Node{
		Kind: manifest.KindEnum, Semantic: "Mode", TypeID: "Mode.json#", Primitive: manifest.Primitive("u8").Primitive,
		Variants: []manifest.Variant{{Value: 0, Name: "Ready", Encode: manifest.Void()}, {Value: 1, Name: "Done", Encode: manifest.Void()}},
	})

	m, err := Reconcile(target, []claims.Result{
		{Pin: sourcePin("endstone"), Target: target, Claims: []claims.Claim{primitiveClaim}},
		{Pin: sourcePin("mojang"), Target: target, Claims: []claims.Claim{enumClaim}},
	}, nil)
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	field := m.Packets[0].Fields[0]
	if field.Encode.Kind != manifest.KindEnum || len(field.Provenance.Pins) != 2 {
		t.Fatalf("field = %#v, want semantic enum with both source pins", field)
	}
}

func TestReconcileIgnoresNestedSourceNamesAndProvenanceForWireComparison(t *testing.T) {
	target := manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}
	leftNode := manifest.Node{Kind: manifest.KindStruct, Semantic: "Left", TypeID: "Left", Fields: []manifest.Field{{Ordinal: 0, Name: "First", Encode: manifest.Primitive("u16le"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"endstone"}}}}}
	rightNode := manifest.Node{Kind: manifest.KindStruct, Semantic: "Right", TypeID: "Right.json#", Fields: []manifest.Field{{Ordinal: 0, Name: "Other", Encode: manifest.Primitive("u16le"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"mojang"}}}}}

	m, err := Reconcile(target, []claims.Result{
		{Pin: sourcePin("endstone"), Target: target, Claims: []claims.Claim{testClaim("endstone", leftNode)}},
		{Pin: sourcePin("mojang"), Target: target, Claims: []claims.Claim{testClaim("mojang", rightNode)}},
	}, nil)
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	if got := m.Packets[0].Fields[0].Provenance.Pins; len(got) != 2 {
		t.Fatalf("pins = %v, want both sources", got)
	}
}

func testClaim(sourceID string, node manifest.Node) claims.Claim {
	return claims.Claim{SourceID: sourceID, Locator: "fixture/Vocabulary.Value", PacketID: 1, PacketName: "Vocabulary", Direction: manifest.DirectionClientbound, FieldPath: "Vocabulary.Value", Ordinal: 0, Name: "Value", Encode: node, Symmetry: manifest.Symmetric}
}

func sourcePin(id string) manifest.SourcePin {
	return manifest.SourcePin{ID: id, Kind: id, Revision: "fixture-2168", Digest: "sha256:" + id, Locator: "https://github.com/example/" + id + "-docs/tree/rev", MinecraftVersion: "fixture", ProtocolVersion: 2168}
}

func mustFingerprint(t *testing.T, claim claims.Claim) string {
	t.Helper()
	digest, err := claims.Fingerprint(claim)
	if err != nil {
		t.Fatalf("Fingerprint: %v", err)
	}
	return digest
}

func TestReconcileRetainsEmptyPackets(t *testing.T) {
	target := manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}
	result := claims.Result{
		Pin: sourcePin("endstone"), Target: target,
		Packets: []claims.PacketClaim{{SourceID: "endstone", Locator: "packets/EmptyPacket.json", ID: 4, Name: "EmptyPacket", Direction: manifest.DirectionClientbound}},
	}
	m, err := Reconcile(target, []claims.Result{result}, nil)
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	if len(m.Packets) != 1 || m.Packets[0].ID != 4 || len(m.Packets[0].Fields) != 0 {
		t.Fatalf("packets = %#v, want one empty packet", m.Packets)
	}
}

func TestMergeOptionalWrapperAroundComplementaryConcreteShape(t *testing.T) {
	concrete := manifest.Union(manifest.Primitive("var_u32"), manifest.Variant{Value: 0, Name: "None", Encode: manifest.Void()})
	partial := manifest.Optional(manifest.Unresolved("published oneOf omitted selectors", true))
	merged, ok := mergeNode(partial, concrete)
	if !ok || merged.Kind != manifest.KindOptional || merged.Value == nil || merged.Value.Kind != manifest.KindUnion {
		t.Fatalf("mergeNode = %#v, %v; want optional concrete union", merged, ok)
	}
}

func TestReconcileRequiresAdjudicationForVariantDecodeDisagreement(t *testing.T) {
	target := manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}
	var results []claims.Result
	for i, source := range []string{"first", "second", "third"} {
		decode := manifest.Primitive("u16le")
		if i == 2 {
			decode = manifest.Primitive("u32le")
		}
		node := manifest.Union(manifest.Primitive("u8"), manifest.Variant{Value: 1, Name: "One", Encode: manifest.Primitive("u8"), Decode: &decode})
		results = append(results, claims.Result{Pin: sourcePin(source), Target: target, Claims: []claims.Claim{testClaim(source, node)}})
	}
	if _, err := Reconcile(target, results, nil); err == nil || !strings.Contains(err.Error(), "adjudication") {
		t.Fatalf("Reconcile error = %v", err)
	}
	left := manifest.Union(manifest.Primitive("u8"), manifest.Variant{Value: 1, Name: "One", Encode: manifest.Primitive("u8")})
	decode := manifest.Primitive("u8")
	decode.Semantic = "Payload"
	right := manifest.Union(manifest.Primitive("u8"), manifest.Variant{Value: 1, Name: "One", Encode: manifest.Primitive("u8"), Decode: &decode})
	merged, ok := mergeNode(left, right)
	if !ok || merged.Variants[0].Decode == nil || merged.Variants[0].Decode.Semantic != "Payload" {
		t.Fatalf("compatible decode metadata was lost: %#v, %v", merged, ok)
	}
}

func TestReconcileNestedEnrichmentDoesNotMutateClaims(t *testing.T) {
	target := manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}
	wrap := func(node manifest.Node) manifest.Node {
		return manifest.Struct(manifest.Field{Ordinal: 0, Name: "Nested", Encode: node, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"endstone"}}})
	}
	left := testClaim("endstone", wrap(manifest.Bytes(manifest.Primitive("var_u32"))))
	text := manifest.String(manifest.Primitive("var_u32"))
	text.TypeID = "Payload"
	right := testClaim("mojang", wrap(text))
	group := []claims.Claim{left, right}
	context, err := claims.ContextFingerprint(target, group)
	if err != nil {
		t.Fatal(err)
	}
	adj := manifest.Adjudication{ID: "binary", Target: left.FieldPath, PrePatchContextSHA256: context,
		Claims:         []manifest.ClaimFingerprint{{SourceID: left.SourceID, Digest: mustFingerprint(t, left)}, {SourceID: right.SourceID, Digest: mustFingerprint(t, right)}},
		SelectedSource: left.SourceID, Evidence: []manifest.Evidence{{SourceID: left.SourceID, Locator: "fixture://serializer"}}, Reason: "binary payload"}
	results := []claims.Result{{Pin: sourcePin("endstone"), Target: target, Claims: []claims.Claim{left}}, {Pin: sourcePin("mojang"), Target: target, Claims: []claims.Claim{right}}}
	for range 2 {
		if _, err := Reconcile(target, results, []manifest.Adjudication{adj}); err != nil {
			t.Fatal(err)
		}
		got, err := claims.ContextFingerprint(target, group)
		if err != nil {
			t.Fatal(err)
		}
		if got != context {
			t.Error("reconciliation mutated input claim fingerprints")
		}
	}
}

func TestReconcileAppliesNBTEncodingToAsymmetricDecode(t *testing.T) {
	target := manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}
	var results []claims.Result
	for _, source := range []string{"endstone", "mojang"} {
		claim := testClaim(source, manifest.Primitive("u8"))
		decode := manifest.Primitive("nbt_le")
		claim.Decode = &decode
		claim.Symmetry = manifest.Asymmetric
		results = append(results, claims.Result{Pin: sourcePin(source), Target: target, Claims: []claims.Claim{claim}})
	}
	encodings := nbtencoding.Table{SchemaVersion: nbtencoding.SchemaVersion, Target: target,
		Source: nbtencoding.Source{Repository: "https://example.invalid/oracle", Revision: strings.Repeat("a", 40), Locator: "fixture://encoding", SHA256: "sha256:" + strings.Repeat("b", 64)},
		Fields: []nbtencoding.Entry{{PacketID: 1, PacketName: "Vocabulary", FieldOrdinal: 0, FieldName: "Value", Path: "decode", Encoding: manifest.NBTNetwork, Evidence: nbtencoding.Evidence{Locator: "fixture://encoding", Summary: "network encoding"}}}}
	m, err := reconcile(target, results, nil, nil, &encodings)
	if err != nil {
		t.Fatal(err)
	}
	if m.Packets[0].Fields[0].Decode.Encoding != string(manifest.NBTNetwork) {
		t.Fatal("decode encoding was not applied")
	}
}
