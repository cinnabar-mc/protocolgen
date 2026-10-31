package emitter

import (
	"os"
	"path/filepath"
	"testing"

	"protocolgen/internal/manifest"
)

func TestRunLoadsAdjacentOverlaysAndWritesEmitterFiles(t *testing.T) {
	directory := t.TempDir()
	manifestPath := filepath.Join(directory, "manifest.json")
	target := manifest.Target{MinecraftVersion: "1.0.0", ProtocolVersion: 1}
	m := manifest.Manifest{
		SchemaVersion: manifest.SchemaVersion,
		Target:        target,
		Sources:       []manifest.SourcePin{{ID: "fixture", Kind: "fixture", Revision: "v1", Digest: "sha256:fixture"}},
		Packets: []manifest.Packet{{
			ID: 1, Name: "PingPacket", Direction: manifest.DirectionBidirectional,
			Fields: []manifest.Field{{
				Ordinal: 0, Name: "Value", Encode: manifest.Primitive("u8"), Symmetry: manifest.Symmetric,
				Provenance: manifest.Provenance{Pins: []string{"fixture"}},
			}},
		}},
	}
	if err := manifest.Write(manifestPath, m); err != nil {
		t.Fatal(err)
	}
	for name, contents := range map[string]string{
		"naming.json":  `{"schema_version":1,"target":{"minecraft_version":"1.0.0","protocol_version":1},"entries":[]}`,
		"domains.json": `{"schema_version":1,"target":{"minecraft_version":"1.0.0","protocol_version":1},"entries":[]}`,
		"docs.json":    `{"schema_version":1,"target":{"minecraft_version":"1.0.0","protocol_version":1},"entries":[]}`,
	} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	output := filepath.Join(directory, "out")
	result, err := Run(Config{ManifestPath: manifestPath, OutputDir: output}, Func(func(input Input) (map[string]string, error) {
		if input.Naming.Names == nil || input.Domains.Domains == nil || input.Docs.Types == nil {
			t.Fatal("Run did not load the adjacent reviewed overlays")
		}
		return map[string]string{"src/generated.txt": "generated\n"}, nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	if result.FileCount != 1 {
		t.Fatalf("Run file count = %d, want 1", result.FileCount)
	}
	data, err := os.ReadFile(filepath.Join(output, "src", "generated.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "generated\n" {
		t.Fatalf("generated file = %q", data)
	}
}

func TestWriteFilesSupportsModulesAndRemovesOnlyStaleGeneratedFiles(t *testing.T) {
	directory := t.TempDir()
	stale := filepath.Join(directory, "stale.rs")
	staleCargo := filepath.Join(directory, "stale.toml")
	keep := filepath.Join(directory, "notes.txt")
	if err := os.WriteFile(stale, []byte("// Code generated from canonical protocol manifest v2. DO NOT EDIT.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keep, []byte("keep me\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(staleCargo, []byte("# Code generated from canonical protocol manifest v2. DO NOT EDIT.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := WriteFiles(directory, map[string]string{"packets/login.rs": "generated\n"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(directory, "packets", "login.rs")); err != nil {
		t.Fatalf("nested output was not written: %v", err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatalf("stale generated file was not removed: %v", err)
	}
	if _, err := os.Stat(staleCargo); !os.IsNotExist(err) {
		t.Fatalf("stale generated Cargo file was not removed: %v", err)
	}
	if _, err := os.Stat(keep); err != nil {
		t.Fatalf("non-generated file was removed: %v", err)
	}
}

func TestWriteFilesRejectsParentTraversal(t *testing.T) {
	if err := WriteFiles(t.TempDir(), map[string]string{"../escape.rs": "bad"}); err == nil {
		t.Fatal("WriteFiles accepted parent traversal")
	}
}

func TestWriteFilesPreservesOutputOnParentConflict(t *testing.T) {
	directory := t.TempDir()
	for name, text := range map[string]string{"stale.rs": "// Code generated from canonical protocol manifest v2. DO NOT EDIT.\n", "packets": "handwritten"} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := WriteFiles(directory, map[string]string{"packets/login.rs": "generated"}); err == nil {
		t.Fatal("accepted parent-file conflict")
	}
	if data, err := os.ReadFile(filepath.Join(directory, "stale.rs")); err != nil || len(data) == 0 {
		t.Fatalf("lost previous output: %q, %v", data, err)
	}
	if data, err := os.ReadFile(filepath.Join(directory, "packets")); err != nil || string(data) != "handwritten" {
		t.Fatalf("changed handwritten file: %q, %v", data, err)
	}
}

func TestWriteFilesCannotFollowExternalSymlinks(t *testing.T) {
	for _, parent := range []bool{false, true} {
		directory, external := t.TempDir(), t.TempDir()
		victim := filepath.Join(external, "victim.rs")
		if err := os.WriteFile(victim, []byte("original"), 0o644); err != nil {
			t.Fatal(err)
		}
		link, target, name := filepath.Join(directory, "victim.rs"), victim, "victim.rs"
		if parent {
			link, target, name = filepath.Join(directory, "link"), external, "link/victim.rs"
		}
		if err := os.Symlink(target, link); err != nil {
			t.Fatal(err)
		}
		if err := WriteFiles(directory, map[string]string{name: "replacement"}); err == nil {
			t.Error("accepted escaping symlink")
		}
		if data, err := os.ReadFile(victim); err != nil || string(data) != "original" {
			t.Fatalf("external file changed: %q, %v", data, err)
		}
	}
}
