package nbtencoding

import (
	"strings"
	"testing"

	"protocolgen/internal/manifest"
)

func TestApplyAssignsTopLevelAndNestedNBTNodes(t *testing.T) {
	target := manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}
	nested := manifest.Struct(manifest.Field{
		Ordinal: 0, Name: "Nested", Encode: manifest.Primitive("nbt_le"), Symmetry: manifest.Symmetric,
		Provenance: manifest.Provenance{Pins: []string{"fixture"}},
	})
	conditional := manifest.Node{Kind: manifest.KindConditional, CompareTo: "mode", Cases: []manifest.Case{{Value: "x", Encode: []manifest.Node{manifest.Primitive("nbt_le")}}}}
	m := manifest.Manifest{
		SchemaVersion: manifest.SchemaVersion,
		Target:        target,
		Sources:       []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "sha256:fixture"}},
		Packets: []manifest.Packet{{ID: 1, Name: "Packet", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{
			{Ordinal: 0, Name: "Top", Encode: manifest.Primitive("nbt_le"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
			{Ordinal: 1, Name: "Container", Encode: nested, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
			{Ordinal: 2, Name: "Conditional", Encode: conditional, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		}}},
	}
	table := Table{
		SchemaVersion: SchemaVersion,
		Target:        target,
		Source:        Source{Repository: "https://example.invalid/gophertunnel", Revision: strings.Repeat("a", 40), Locator: "https://example.invalid/source", SHA256: "sha256:" + strings.Repeat("b", 64)},
		Fields: []Entry{
			{PacketID: 1, PacketName: "Packet", FieldOrdinal: 0, FieldName: "Top", Path: "encode", Encoding: manifest.NBTNetwork, Evidence: Evidence{Locator: "packet/top.go", Summary: "network NBT"}},
			{PacketID: 1, PacketName: "Packet", FieldOrdinal: 1, FieldName: "Container", Path: "encode.fields[0].encode", Encoding: manifest.NBTPersistent, Evidence: Evidence{Locator: "packet/container.go", Summary: "persistent NBT"}},
			{PacketID: 1, PacketName: "Packet", FieldOrdinal: 2, FieldName: "Conditional", Path: "encode.cases[0].encode[0]", Encoding: manifest.NBTNetwork, Evidence: Evidence{Locator: "packet/conditional.go", Summary: "network NBT"}},
		},
	}
	if err := table.Apply(&m); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if got := m.Packets[0].Fields[0].Encode.Encoding; got != string(manifest.NBTNetwork) {
		t.Fatalf("top encoding = %q", got)
	}
	if got := m.Packets[0].Fields[1].Encode.Fields[0].Encode.Encoding; got != string(manifest.NBTPersistent) {
		t.Fatalf("nested encoding = %q", got)
	}
	if err := manifest.Validate(m); err != nil {
		t.Fatalf("Validate after Apply: %v", err)
	}
}

func TestApplyRejectsMissingAndUnknownNBTEntries(t *testing.T) {
	target := manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}
	m := manifest.Manifest{
		SchemaVersion: manifest.SchemaVersion,
		Target:        target,
		Sources:       []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "sha256:fixture"}},
		Packets: []manifest.Packet{{ID: 1, Name: "Packet", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{{
			Ordinal: 0, Name: "Value", Encode: manifest.Primitive("nbt_le"), Symmetry: manifest.Symmetric,
			Provenance: manifest.Provenance{Pins: []string{"fixture"}},
		}}}},
	}
	table := Table{SchemaVersion: SchemaVersion, Target: target, Source: validSource(), Fields: nil}
	if err := table.Apply(&m); err == nil || !strings.Contains(err.Error(), "missing NBT encoding entry") {
		t.Fatalf("Apply error = %v, want missing entry", err)
	}
	table.Fields = []Entry{{PacketID: 99, PacketName: "Unknown", FieldOrdinal: 0, FieldName: "Value", Path: "encode", Encoding: manifest.NBTNetwork, Evidence: validEvidence()}}
	if err := table.Apply(&m); err == nil || !strings.Contains(err.Error(), "unknown packet") {
		t.Fatalf("Apply error = %v, want unknown packet", err)
	}
}

func validSource() Source {
	return Source{Repository: "https://example.invalid/gophertunnel", Revision: strings.Repeat("a", 40), Locator: "https://example.invalid/source", SHA256: "sha256:" + strings.Repeat("b", 64)}
}

func validEvidence() Evidence {
	return Evidence{Locator: "packet/value.go", Summary: "network NBT"}
}

func TestApplyCoversDecodeOverridesAtEveryLevel(t *testing.T) {
	for _, path := range []string{"decode", "encode.fields[0].decode", "encode.variants[0].decode", "encode.cases[0].decode[0]"} {
		t.Run(path, func(t *testing.T) {
			nbt := manifest.Primitive("nbt_le")
			plain := manifest.Primitive("u8")
			field := manifest.Field{Ordinal: 0, Name: "Value", Encode: plain, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}
			switch path {
			case "decode":
				field.Decode = &nbt
				field.Symmetry = manifest.Asymmetric
			case "encode.fields[0].decode":
				inner := field
				inner.Decode = &nbt
				inner.Symmetry = manifest.Asymmetric
				field.Encode = manifest.Struct(inner)
			case "encode.variants[0].decode":
				field.Encode = manifest.Union(plain, manifest.Variant{Value: 1, Name: "One", Encode: plain, Decode: &nbt})
			case "encode.cases[0].decode[0]":
				field.Encode = manifest.Node{Kind: manifest.KindConditional, CompareTo: "Mode", Cases: []manifest.Case{{Value: "1", Encode: []manifest.Node{plain}, Decode: []manifest.Node{nbt}}}}
			}
			target := manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}
			m := manifest.Manifest{SchemaVersion: manifest.SchemaVersion, Target: target, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "sha256:fixture"}}, Packets: []manifest.Packet{{ID: 1, Name: "Packet", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{field}}}}
			table := Table{SchemaVersion: SchemaVersion, Target: target, Source: validSource()}
			if err := table.Validate(target, m.Packets); err == nil || !strings.Contains(err.Error(), "missing NBT encoding entry") {
				t.Errorf("missing decode coverage error = %v", err)
			}
			table.Fields = []Entry{{PacketID: 1, PacketName: "Packet", FieldOrdinal: 0, FieldName: "Value", Path: path, Encoding: manifest.NBTNetwork, Evidence: validEvidence()}}
			if err := table.Apply(&m); err != nil {
				t.Fatal(err)
			}
			paths := nbtNodes(&m.Packets[0].Fields[0])
			if len(paths) != 1 || paths[path] == nil {
				t.Fatalf("NBT paths = %v", paths)
			}
		})
	}
}

func TestApplyRejectsConflictingAssignmentsToSharedNodes(t *testing.T) {
	shared := manifest.Optional(manifest.Primitive("nbt_le"))
	m := manifest.Manifest{SchemaVersion: 2, Target: manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 1}, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "1", Digest: "fixture"}}, Packets: []manifest.Packet{{ID: 1, Name: "Packet", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{
		{Ordinal: 0, Name: "First", Encode: shared, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		{Ordinal: 1, Name: "Second", Encode: shared, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
	}}}}
	table := Table{SchemaVersion: 1, Target: m.Target, Source: validSource(), Fields: []Entry{
		{PacketID: 1, PacketName: "Packet", FieldOrdinal: 0, FieldName: "First", Path: "encode.value", Encoding: manifest.NBTNetwork, Evidence: validEvidence()},
		{PacketID: 1, PacketName: "Packet", FieldOrdinal: 1, FieldName: "Second", Path: "encode.value", Encoding: manifest.NBTPersistent, Evidence: validEvidence()},
	}}
	if err := table.Apply(&m); err == nil {
		t.Fatal("accepted conflicting assignments to one shared NBT node")
	}
	if shared.Value.Encoding != "" {
		t.Fatal("failed application changed the shared node")
	}
	table.Fields[1].Encoding = manifest.NBTNetwork
	if err := table.Apply(&m); err != nil {
		t.Fatal(err)
	}
}
