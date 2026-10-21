// Command entitybreedinggen projects locked official breeding definitions into
// the existing entity catalog without requiring the independent BDS base inputs.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bedrock-mc/protocolgen/data/internal/generator"
	"github.com/bedrock-mc/protocolgen/data/source"
)

func main() {
	lock := flag.String("lock", "", "source lock path; defaults to the bundled release")
	output := flag.String("out", "../generated/data", "existing generated catalog root")
	cacheRoot, _ := os.UserCacheDir()
	cache := flag.String("cache", filepath.Join(cacheRoot, "protocolgen-data"), "authenticated input cache")
	check := flag.Bool("check", false, "verify the projection without writing")
	flag.Parse()
	if err := run(*lock, *cache, *output, *check); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run verifies the base lock before projecting into its existing entity domain.
func run(lockPath, cache, output string, check bool) error {
	sources, err := source.Open(lockPath)
	if err != nil {
		return err
	}
	digest, err := sources.Lock.SHA256()
	if err != nil {
		return err
	}
	manifest, err := os.ReadFile(filepath.Join(output, "semantic_sources.json"))
	if err != nil {
		return err
	}
	var base struct {
		SourceLockSHA256 string `json:"source_lock_sha256"`
	}
	if err := json.Unmarshal(manifest, &base); err != nil {
		return err
	}
	if base.SourceLockSHA256 != digest {
		return fmt.Errorf("entity base snapshot and breeding inputs use different source locks")
	}
	paths, err := filepath.Glob(filepath.Join(output, "entity", "*_generated.go"))
	if err != nil {
		return err
	}
	files := make(map[string][]byte)
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !bytes.HasPrefix(content, []byte(generator.GeneratedMarker+"\n")) {
			continue
		}
		files["entity/"+filepath.Base(path)] = content
	}
	if len(files) == 0 {
		return fmt.Errorf("no generated base entities in %s", output)
	}
	inputs, err := generator.ReadEntityBreedingInputs(context.Background(), sources, cache)
	if err != nil {
		return err
	}
	generated, count, err := generator.GenerateEntityBreeding(files, inputs)
	if err != nil {
		return err
	}
	if check {
		for name, want := range generated {
			got, err := os.ReadFile(filepath.Join(output, filepath.FromSlash(name)))
			if err != nil {
				return err
			}
			if !bytes.Equal(got, want) {
				return fmt.Errorf("%s differs from locked breeding projection", name)
			}
		}
		return nil
	}
	if err := generator.Write(output, generator.GeneratedMarker, generated); err != nil {
		return err
	}
	fmt.Printf("generated %d supported entity breeding descriptors\n", count)
	return nil
}
