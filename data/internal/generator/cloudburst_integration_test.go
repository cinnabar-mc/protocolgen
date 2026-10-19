package generator

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/bedrock-mc/protocolgen/data/internal/nativeblock"
	"github.com/bedrock-mc/protocolgen/data/source"
)

// TestLockedCloudburstOutput verifies full committed block, biome and named-shape output when
// authenticated source files are available. CI supplies them; unit tests stay offline.
func TestLockedCloudburstOutput(t *testing.T) {
	dir := os.Getenv("PROTOCOLGEN_CLOUDBURST_DIR")
	if dir == "" {
		t.Skip("set PROTOCOLGEN_CLOUDBURST_DIR to verify the complete locked extract")
	}
	inputs, err := source.Open("")
	if err != nil {
		t.Fatal(err)
	}
	if err := inputs.ValidateCloudburst(dir); err != nil {
		t.Fatal(err)
	}
	if inputs.Revision("liquid_clip_omissions") != inputs.Lock.CloudburstRevision() {
		t.Fatal("liquid clip omissions must match the Cloudburst revision")
	}
	encoded, err := inputs.Read(context.Background(), "liquid_clip_omissions", "")
	if err != nil {
		t.Fatal(err)
	}
	var omissions []LiquidClipOmission
	if err := json.Unmarshal(encoded, &omissions); err != nil {
		t.Fatal(err)
	}
	biomeInput, err := inputs.Read(context.Background(), "native_biome_ids", "")
	if err != nil {
		t.Fatal(err)
	}
	biomeIDs, err := DecodeNativeBiomeIDs(biomeInput)
	if err != nil {
		t.Fatal(err)
	}
	encodedBiomes, err := inputs.Read(context.Background(), "native_biomes", "")
	if err != nil {
		t.Fatal(err)
	}
	nativeBiomes, err := DecodeNativeBiomes(encodedBiomes)
	if err != nil {
		t.Fatal(err)
	}
	native, err := nativeblock.Read(context.Background(), inputs, "")
	if err != nil {
		t.Fatal(err)
	}
	files, _, err := GenerateCloudburst(Config{CloudburstDir: dir, BiomeIDs: biomeIDs, NativeBiomes: nativeBiomes, LiquidClipOmissions: omissions, NativeBlocks: &native})
	if err != nil {
		t.Fatal(err)
	}
	catalog := filepath.Join("..", "..", "..", "generated", "data")
	for name, want := range files {
		got, err := os.ReadFile(filepath.Join(catalog, filepath.FromSlash(name)))
		if err != nil || !bytes.Equal(got, want) {
			t.Errorf("%s differs from the locked extract; regenerate the catalog (%v)", name, err)
		}
	}
	for _, pkg := range []string{"block", "biome", "voxelshape"} {
		entries, err := os.ReadDir(filepath.Join(catalog, pkg))
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := filepath.Join(pkg, entry.Name())
			content, err := os.ReadFile(filepath.Join(catalog, name))
			if err != nil {
				t.Fatal(err)
			}
			if bytes.HasPrefix(content, []byte(GeneratedMarker)) {
				if _, exists := files[filepath.ToSlash(name)]; !exists {
					t.Errorf("%s is a stale generated file", name)
				}
			}
		}
	}
}
