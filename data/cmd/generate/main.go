// Command generate converts pinned Cloudburst and vanilla BDS inputs into the
// active generated/data catalog.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bedrock-mc/protocolgen/data/internal/generator"
	"github.com/bedrock-mc/protocolgen/data/internal/nativeblock"
	"github.com/bedrock-mc/protocolgen/data/source"
)

func main() {
	lock := flag.String("lock", "", "source lock path; defaults to the bundled release")
	cloudburst := flag.String("cloudburst", "", "path to the locked CloudburstMC/Data files")
	bds := flag.String("bds", "", "path to the locked Bedrock Dedicated Server inputs")
	output := flag.String("out", "../generated/data", "generated catalog root")
	cacheRoot, _ := os.UserCacheDir()
	cache := flag.String("cache", filepath.Join(cacheRoot, "protocolgen-data"), "authenticated input cache")
	flag.Parse()

	stats, err := run(*lock, *cache, *cloudburst, *bds, *output)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf(
		"generated %d block states, %d biomes, %d voxel shapes, %d entities, and %d foods\n",
		stats.BlockStates, stats.Biomes, stats.VoxelShapes, stats.Entities, stats.Foods,
	)
	if stats.UnavailableLiquidClipShapes > 0 {
		fmt.Printf("marked %d reviewed liquid clip shapes unavailable; see semantic_sources.json\n", stats.UnavailableLiquidClipShapes)
	}
}

// run authenticates every input and derives generation versions from the lock.
func run(lockPath, cache, cloudburstDir, bdsDir, outputDir string) (generator.Stats, error) {
	sources, err := source.Open(lockPath)
	if err != nil {
		return generator.Stats{}, err
	}
	if err := sources.ValidateSemantic(cloudburstDir, bdsDir); err != nil {
		return generator.Stats{}, err
	}
	if sources.Revision("liquid_clip_omissions") != sources.Lock.CloudburstRevision() {
		return generator.Stats{}, fmt.Errorf("liquid clip omissions must match the Cloudburst revision")
	}
	omitted, err := sources.Read(context.Background(), "liquid_clip_omissions", cache)
	if err != nil {
		return generator.Stats{}, err
	}
	var omissions []generator.LiquidClipOmission
	if err := json.Unmarshal(omitted, &omissions); err != nil {
		return generator.Stats{}, fmt.Errorf("decode liquid clip omissions: %w", err)
	}
	encodedBiomeIDs, err := sources.Read(context.Background(), "native_biome_ids", cache)
	if err != nil {
		return generator.Stats{}, fmt.Errorf("read biome IDs: %w", err)
	}
	biomeIDs, err := generator.DecodeNativeBiomeIDs(encodedBiomeIDs)
	if err != nil {
		return generator.Stats{}, err
	}
	encodedBiomes, err := sources.Read(context.Background(), "native_biomes", cache)
	if err != nil {
		return generator.Stats{}, err
	}
	nativeBiomes, err := generator.DecodeNativeBiomes(encodedBiomes)
	if err != nil {
		return generator.Stats{}, err
	}
	digest, err := sources.Lock.SHA256()
	if err != nil {
		return generator.Stats{}, err
	}
	native, err := nativeblock.Read(context.Background(), sources, cache)
	if err != nil {
		return generator.Stats{}, err
	}
	breeding, err := generator.ReadEntityBreedingInputs(context.Background(), sources, cache)
	if err != nil {
		return generator.Stats{}, err
	}
	files, stats, err := generator.Generate(generator.Config{
		CloudburstDir:       cloudburstDir,
		BDSDir:              bdsDir,
		CloudburstRef:       sources.Lock.CloudburstRevision(),
		BDSVersion:          sources.Lock.Semantic.BDSVersion,
		BiomeIDs:            biomeIDs,
		NativeBiomes:        nativeBiomes,
		LiquidClipOmissions: omissions,
		SourceLockSHA256:    digest,
		NativeBlocks:        &native,
		NativeRef:           sources.Revision("native_block_states"),
		EntityBreeding:      &breeding,
	})
	if err != nil {
		return generator.Stats{}, err
	}
	return stats, generator.Write(outputDir, generator.GeneratedMarker, files)
}
