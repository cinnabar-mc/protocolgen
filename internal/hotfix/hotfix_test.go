package hotfix

import (
	"testing"

	"protocolgen/internal/manifest"
)

func TestApplyWrapsOnlyTheFingerprintedOptionalNode(t *testing.T) {
	base := testManifest()
	baseBytes, err := manifest.MarshalStable(base)
	if err != nil {
		t.Fatal(err)
	}
	node := base.Packets[0].Fields[0].Encode
	pre, err := nodeDigest(node)
	if err != nil {
		t.Fatal(err)
	}
	spec := Spec{
		SchemaVersion:      1,
		Target:             manifest.Target{MinecraftVersion: "1.26.44", ProtocolVersion: 2168},
		BaseManifestSHA256: bytesDigest(baseBytes),
		Sources: []manifest.SourcePin{{
			ID: "hotfix", Kind: "cloudburst", Revision: "2222222222222222222222222222222222222222",
			Digest:           "sha256:2222222222222222222222222222222222222222222222222222222222222222",
			MinecraftVersion: "1.26.44", ProtocolVersion: 2168,
		}},
		Operations: []Operation{{
			ID: "nested", PacketID: 108, FieldOrdinal: 0, Path: "encode", Operation: "wrap_optional",
			PrePatchNodeSHA256: pre, Reason: "wire hotfix",
			Evidence: []manifest.Evidence{{SourceID: "hotfix", Locator: "fixture://hotfix"}},
		}},
	}
	mutated := testManifest()
	mutated.Packets[0].Direction = manifest.DirectionServerbound
	if got, err := Apply(mutated, baseBytes, spec); err == nil && got.Packets[0].Direction != base.Packets[0].Direction {
		t.Fatal("Apply retained changes outside fingerprinted bytes")
	}
	got, err := Apply(base, baseBytes, spec)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if got.Target != spec.Target || got.Derivation == nil {
		t.Fatalf("derived target/proof = %#v / %#v", got.Target, got.Derivation)
	}
	outer := got.Packets[0].Fields[0].Encode
	if outer.Kind != manifest.KindOptional || outer.Value == nil || outer.Value.Kind != manifest.KindOptional {
		t.Fatalf("derived node = %#v, want nested optional", outer)
	}
	if base.Packets[0].Fields[0].Encode.Value == nil || base.Packets[0].Fields[0].Encode.Value.Kind != manifest.KindString {
		t.Fatal("Apply mutated the base manifest")
	}
}

func TestApplyRejectsStaleBaseAndNodeFingerprints(t *testing.T) {
	base := testManifest()
	baseBytes, err := manifest.MarshalStable(base)
	if err != nil {
		t.Fatal(err)
	}
	spec := Spec{SchemaVersion: 1, Target: manifest.Target{MinecraftVersion: "1.26.44", ProtocolVersion: 2168}, BaseManifestSHA256: "sha256:" + string(make([]byte, 64))}
	if _, err := Apply(base, baseBytes, spec); err == nil {
		t.Fatal("Apply accepted a stale base-manifest fingerprint")
	}
}

func TestApplyRejectsEvidenceSourceFromTheBaseVersion(t *testing.T) {
	base := testManifest()
	baseBytes, err := manifest.MarshalStable(base)
	if err != nil {
		t.Fatal(err)
	}
	node, err := nodeDigest(base.Packets[0].Fields[0].Encode)
	if err != nil {
		t.Fatal(err)
	}
	spec := Spec{
		SchemaVersion: 1, Target: manifest.Target{MinecraftVersion: "1.26.44", ProtocolVersion: 2168},
		BaseManifestSHA256: bytesDigest(baseBytes),
		Sources: []manifest.SourcePin{{
			ID: "stale", Kind: "fixture", Revision: "2222222222222222222222222222222222222222",
			Digest:           "sha256:2222222222222222222222222222222222222222222222222222222222222222",
			MinecraftVersion: "1.26.40", ProtocolVersion: 2168,
		}},
		Operations: []Operation{{
			ID: "nested", PacketID: 108, FieldOrdinal: 0, Path: "encode", Operation: "wrap_optional",
			PrePatchNodeSHA256: node, Reason: "wire hotfix",
			Evidence: []manifest.Evidence{{SourceID: "stale", Locator: "fixture://stale"}},
		}},
	}
	if _, err := Apply(base, baseBytes, spec); err == nil {
		t.Fatal("Apply accepted target evidence pinned to the base version")
	}
}

func testManifest() manifest.Manifest {
	return manifest.Manifest{
		SchemaVersion: manifest.SchemaVersion,
		Target:        manifest.Target{MinecraftVersion: "1.26.40", ProtocolVersion: 2168},
		Sources: []manifest.SourcePin{{
			ID: "base", Kind: "fixture", Revision: "1111111111111111111111111111111111111111",
			Digest:           "sha256:1111111111111111111111111111111111111111111111111111111111111111",
			MinecraftVersion: "1.26.40", ProtocolVersion: 2168,
		}},
		Packets: []manifest.Packet{{ID: 108, Name: "SetScorePacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{{
			Ordinal: 0, Name: "Score Info", Encode: manifest.Optional(manifest.String(manifest.Primitive("var_u32"))),
			Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"base"}},
		}}}},
	}
}
