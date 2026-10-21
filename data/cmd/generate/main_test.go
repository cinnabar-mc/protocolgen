package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bedrock-mc/protocolgen/data/source"
)

func TestRunUsesLockedIdentities(t *testing.T) {
	lockPath, cloudburst, bds := lockedFixture(t)
	output := t.TempDir()
	stats, err := run(lockPath, "", cloudburst, bds, output)
	if err != nil {
		t.Fatal(err)
	}
	if stats.BlockStates != 2 || stats.Entities != 2 || stats.Foods != 1 {
		t.Fatalf("unexpected generated counts: %+v", stats)
	}
	data, err := os.ReadFile(filepath.Join(output, "semantic_sources.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	locked, err := source.Open(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	digest, err := locked.Lock.SHA256()
	if err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{
		"cloudburst_ref": strings.Repeat("1", 40), "bds_version": "locked-bds",
		"source_lock_sha256": digest,
	} {
		if manifest[key] != want {
			t.Errorf("generated %s = %v, want %s", key, manifest[key], want)
		}
	}
}

func TestRunRejectsChangedInputsBeforeWriting(t *testing.T) {
	lockPath, cloudburst, bds := lockedFixture(t)
	if err := os.WriteFile(filepath.Join(cloudburst, "blocks.json"), []byte("[]"), 0o644); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "output")
	if _, err := run(lockPath, "", cloudburst, bds, output); err == nil || !strings.Contains(err.Error(), "digest is") {
		t.Fatalf("run error = %v, want source digest rejection", err)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatalf("invalid inputs created an output directory: %v", err)
	}
}

func TestRunRequiresBreedingArchiveBeforeWriting(t *testing.T) {
	lockPath, cloudburst, bds := lockedFixture(t)
	encoded, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	var lock source.Lock
	if err := json.Unmarshal(encoded, &lock); err != nil {
		t.Fatal(err)
	}
	delete(lock.Inputs, "resource_pack")
	encoded, err = json.Marshal(lock)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(lockPath, encoded, 0o644); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(t.TempDir(), "output")
	if _, err := run(lockPath, "", cloudburst, bds, output); err == nil || !strings.Contains(err.Error(), "no sample archive") {
		t.Fatalf("run = %v, want required breeding source", err)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatalf("missing source wrote output: %v", err)
	}
}

// lockedFixture creates a complete local lock over the synthetic generator inputs.
func lockedFixture(t *testing.T) (lockPath, cloudburst, bds string) {
	t.Helper()
	cloudburst, cloudburstFiles := copyLockedTree(t, "cloudburst")
	bds, bdsFiles := copyLockedTree(t, "bds")
	dir := t.TempDir()
	omissions := []byte("[]")
	if err := os.WriteFile(filepath.Join(dir, "omissions.json"), omissions, 0o644); err != nil {
		t.Fatal(err)
	}
	lock := source.Lock{
		SchemaVersion: 2, Release: "1.26.51",
		Upstreams: map[string]source.Upstream{"cloudburst": {Repository: "example/data", Revision: strings.Repeat("1", 40)}},
		Inputs: map[string]source.Input{
			"liquid_clip_omissions": {Path: "omissions.json", Upstream: "cloudburst", SHA256: fmt.Sprintf("%x", sha256.Sum256(omissions))},
		},
		Semantic: source.SemanticInputs{
			Cloudburst: "cloudburst", BDSVersion: "locked-bds",
			CloudburstFiles: cloudburstFiles, BDSFiles: bdsFiles,
		},
	}
	blocks, err := os.ReadFile(filepath.Join(cloudburst, "blocks.json"))
	if err != nil {
		t.Fatal(err)
	}
	var raw []map[string]any
	if err := json.Unmarshal(blocks, &raw); err != nil {
		t.Fatal(err)
	}
	properties := make(map[string]any)
	states := make(map[string]any)
	for i, block := range raw {
		properties[block["name"].(string)] = map[string]any{"requiresCorrectToolForDrops": block["requiresCorrectToolForDrops"]}
		state := make(map[string]any)
		for _, key := range []string{"friction", "explosionResistance", "lightEmission", "flameOdds", "burnOdds"} {
			state[key] = block[key]
		}
		state["hashedNetworkId"] = block["blockStateHash"]
		state["destroySpeed"] = block["hardness"]
		state["lightDamping"] = block["lightDampening"]
		state["solid"] = block["isSolid"]
		states[fmt.Sprint(i)] = state
	}
	biomeData, err := os.ReadFile(filepath.Join(cloudburst, "stripped_biome_definitions.json"))
	if err != nil {
		t.Fatal(err)
	}
	var biomes map[string]map[string]any
	if err := json.Unmarshal(biomeData, &biomes); err != nil {
		t.Fatal(err)
	}
	for _, biome := range biomes {
		biome["mapWaterColor"] = "#ff336699"
	}
	for key, value := range map[string]any{
		"native_biome_ids":        map[string]int{"minecraft:plains": 1},
		"native_biomes":           biomes,
		"native_version":          map[string]any{"protocol": 2193, "version": "1.26.51"},
		"native_block_properties": properties,
		"native_block_states":     states,
		"native_block_tags":       map[string][]string{},
	} {
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		filename := key + ".json"
		if err := os.WriteFile(filepath.Join(dir, filename), data, 0o644); err != nil {
			t.Fatal(err)
		}
		lock.Inputs[key] = source.Input{Path: filename, Revision: strings.Repeat("2", 40), SHA256: fmt.Sprintf("%x", sha256.Sum256(data))}
	}
	var archive bytes.Buffer
	gz := gzip.NewWriter(&archive)
	tarWriter := tar.NewWriter(gz)
	document := []byte(`{"minecraft:entity":{"description":{"identifier":"minecraft:donkey"}}}`)
	if err := tarWriter.WriteHeader(&tar.Header{Name: "samples/behavior_pack/entities/donkey.json", Typeflag: tar.TypeReg, Mode: 0o644, Size: int64(len(document))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tarWriter.Write(document); err != nil {
		t.Fatal(err)
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	lock.Upstreams["samples"] = source.Upstream{Repository: "Mojang/bedrock-samples", Revision: strings.Repeat("3", 40)}
	for name, fixture := range map[string]struct {
		data     []byte
		upstream string
	}{
		"resource_pack": {archive.Bytes(), "samples"},
		"item_registry": {[]byte(`[{"name":"minecraft:wheat","id":1,"componentBased":false,"version":1}]`), "cloudburst"},
	} {
		filename := name + ".input"
		if err := os.WriteFile(filepath.Join(dir, filename), fixture.data, 0o644); err != nil {
			t.Fatal(err)
		}
		lock.Inputs[name] = source.Input{Path: filename, Upstream: fixture.upstream, SHA256: fmt.Sprintf("%x", sha256.Sum256(fixture.data))}
	}
	data, err := json.Marshal(lock)
	if err != nil {
		t.Fatal(err)
	}
	lockPath = filepath.Join(dir, "lock.json")
	if err := os.WriteFile(lockPath, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return lockPath, cloudburst, bds
}

// copyLockedTree copies and hashes a synthetic source tree for a CLI test.
func copyLockedTree(t *testing.T, name string) (string, map[string]string) {
	t.Helper()
	root := t.TempDir()
	if err := os.CopyFS(root, os.DirFS(filepath.Join("..", "..", "internal", "generator", "testdata", name))); err != nil {
		t.Fatal(err)
	}
	manifest := make(map[string]string)
	if err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		manifest[filepath.ToSlash(relative)] = fmt.Sprintf("%x", sha256.Sum256(data))
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return root, manifest
}
