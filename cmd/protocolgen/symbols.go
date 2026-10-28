package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"protocolgen/internal/manifest"
	"protocolgen/internal/mojangsymbols"
	"protocolgen/internal/sourcelock"
)

func runSupplementSymbols(args []string) error {
	fs := flag.NewFlagSet("supplement-symbols", flag.ContinueOnError)
	manifestPath := fs.String("manifest", "", "canonical manifest input")
	lockPath := fs.String("symbols-lock", "", "pinned Mojang reference enum catalog lock")
	catalogPath := fs.String("symbols-catalog", "", "Mojang html/enums.html at the symbols lock revision; overrides the lock's relative catalog path")
	outPath := fs.String("out", "", "canonical manifest output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *manifestPath == "" || *lockPath == "" || *outPath == "" {
		return fmt.Errorf("-manifest, -symbols-lock, and -out are required")
	}
	result, err := manifest.Load(*manifestPath)
	if err != nil {
		return err
	}
	if err := mojangsymbols.AddMetadataSymbols(&result, *lockPath, *catalogPath); err != nil {
		return err
	}
	if err := manifest.Write(*outPath, result); err != nil {
		return err
	}
	fmt.Printf("manifest v2: %d non-wire symbol groups -> %s\n", len(result.Symbols), *outPath)
	return nil
}

func addConfiguredSymbols(result *manifest.Manifest, lock sourcelock.Lock, lockPath, override, catalog string) error {
	path := override
	if path == "" && lock.SymbolsLock != "" {
		path = filepath.Join(filepath.Dir(lockPath), lock.SymbolsLock)
	}
	if path == "" {
		if catalog != "" {
			return fmt.Errorf("-symbols-catalog given but %s configures no symbols lock", lockPath)
		}
		return nil
	}
	// A configured but absent lock/catalog is an error, never silent omission.
	return mojangsymbols.AddMetadataSymbols(result, path, catalog)
}
