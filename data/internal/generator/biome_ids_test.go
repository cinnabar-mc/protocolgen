package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLockedBiomeIDsMatchRetailCatalog(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "source", "inputs", "1.26.51", "native", "biome_id_map.json"))
	if err != nil {
		t.Fatal(err)
	}
	ids, err := DecodeNativeBiomeIDs(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 89 || ids["minecraft:ocean"] != 0 || ids["minecraft:plains"] != 1 || ids["minecraft:dappled_forest"] != 195 {
		t.Fatalf("unexpected retail biome IDs: count=%d ocean=%d plains=%d dappled_forest=%d", len(ids), ids["minecraft:ocean"], ids["minecraft:plains"], ids["minecraft:dappled_forest"])
	}
}

func TestBiomeIDsMustCoverCloudburstNames(t *testing.T) {
	path := filepath.Join("testdata", "cloudburst", "stripped_biome_definitions.json")
	for _, test := range []struct {
		name string
		ids  map[string]int32
		want string
	}{
		{"missing", map[string]int32{}, "cover 0 names"},
		{"wrong name", map[string]int32{"minecraft:ocean": 0}, `"minecraft:plains" has no locked numeric ID`},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, _, err := generateBiomes(path, test.ids, nil); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("generateBiomes error = %v, want %q", err, test.want)
			}
		})
	}
}
