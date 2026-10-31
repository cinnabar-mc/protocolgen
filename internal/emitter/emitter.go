// Package emitter provides the common orchestration contract for language
// backends. Backends receive the same validated manifest and reviewed overlays,
// then return a deterministic set of relative output files.
package emitter

import (
	"crypto/rand"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"protocolgen/internal/docs"
	"protocolgen/internal/domains"
	"protocolgen/internal/layout"
	"protocolgen/internal/manifest"
	"protocolgen/internal/naming"
	"protocolgen/internal/semantics"
)

type Input struct {
	Manifest  manifest.Manifest
	Naming    naming.Overlay
	Domains   domains.Overlay
	Docs      docs.Overlay
	Layout    layout.Overlay // only when Config.LayoutPath is set; never defaulted from the manifest directory
	Semantics semantics.Overlay
}

type Backend interface {
	Generate(Input) (map[string]string, error)
}

type Func func(Input) (map[string]string, error)

func (f Func) Generate(input Input) (map[string]string, error) { return f(input) }

type Config struct {
	ManifestPath  string
	NamingPath    string
	DomainsPath   string
	DocsPath      string
	LayoutPath    string
	SemanticsPath string
	OutputDir     string
}

type Result struct {
	FileCount int
	Coverage  docs.Coverage
}

// Run loads and validates the common generator inputs, invokes one language
// backend, and safely replaces its generated output files.
func Run(config Config, backend Backend) (Result, error) {
	if config.ManifestPath == "" || config.OutputDir == "" {
		return Result{}, fmt.Errorf("manifest path and output directory are required")
	}
	if backend == nil {
		return Result{}, fmt.Errorf("emitter backend is required")
	}
	m, err := manifest.Load(config.ManifestPath)
	if err != nil {
		return Result{}, err
	}
	namingOverlay, err := loadOptionalOverlay(config.ManifestPath, config.NamingPath, "naming.json", "naming", func(path string) (naming.Overlay, error) {
		return naming.LoadOverlay(path, m)
	})
	if err != nil {
		return Result{}, err
	}
	domainOverlay, err := loadOptionalOverlay(config.ManifestPath, config.DomainsPath, "domains.json", "domains", func(path string) (domains.Overlay, error) {
		return domains.LoadOverlay(path, m)
	})
	if err != nil {
		return Result{}, err
	}
	docOverlay, err := loadOptionalOverlay(config.ManifestPath, config.DocsPath, "docs.json", "docs", func(path string) (docs.Overlay, error) {
		return docs.LoadOverlay(path, m)
	})
	if err != nil {
		return Result{}, err
	}
	semanticOverlay, err := loadOptionalOverlay(config.ManifestPath, config.SemanticsPath, "semantics.json", "semantics", func(path string) (semantics.Overlay, error) {
		return semantics.LoadOverlay(path, m)
	})
	if err != nil {
		return Result{}, err
	}
	var layoutOverlay layout.Overlay
	if config.LayoutPath != "" {
		layoutOverlay, err = layout.LoadOverlay(config.LayoutPath, m)
		if err != nil {
			return Result{}, err
		}
	}
	files, err := backend.Generate(Input{Manifest: m, Naming: namingOverlay, Domains: domainOverlay, Docs: docOverlay, Layout: layoutOverlay, Semantics: semanticOverlay})
	if err != nil {
		return Result{}, err
	}
	if err := WriteFiles(config.OutputDir, files); err != nil {
		return Result{}, err
	}
	return Result{FileCount: len(files), Coverage: docs.CoverageOf(m, docOverlay)}, nil
}

func loadOptionalOverlay[T any](manifestPath, explicitPath, defaultName, label string, load func(string) (T, error)) (T, error) {
	path := explicitPath
	if path == "" {
		path = filepath.Join(filepath.Dir(manifestPath), defaultName)
	}
	if _, err := os.Stat(path); err != nil {
		var zero T
		if os.IsNotExist(err) && explicitPath == "" {
			return zero, nil
		}
		return zero, fmt.Errorf("stat %s overlay: %w", label, err)
	}
	return load(path)
}

// WriteFiles stages every replacement before changing generated output. Each
// rename is atomic, and all filesystem access stays within the output root.
func WriteFiles(directory string, files map[string]string) error {
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return err
	}
	root, err := os.OpenRoot(directory)
	if err != nil {
		return err
	}
	defer root.Close()
	desired := make(map[string]bool, len(files))
	names := make([]string, 0, len(files))
	for name := range files {
		clean := filepath.Clean(name)
		if clean == "." || filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || desired[clean] {
			return fmt.Errorf("emitter returned unsafe or duplicate filename %q", name)
		}
		desired[clean] = true
		names = append(names, name)
	}
	sort.Strings(names)
	stage := ".protocolgen-" + rand.Text()
	if err := root.Mkdir(stage, 0o700); err != nil {
		return err
	}
	defer root.RemoveAll(stage)
	for _, name := range names {
		clean := filepath.Clean(name)
		if err := root.MkdirAll(filepath.Dir(clean), 0o755); err != nil {
			return err
		}
		if info, err := root.Lstat(clean); err == nil {
			if !info.Mode().IsRegular() {
				return fmt.Errorf("output %q is not a regular file", name)
			}
		} else if !os.IsNotExist(err) {
			return err
		}
		staged := filepath.Join(stage, clean)
		if err := root.MkdirAll(filepath.Dir(staged), 0o755); err != nil {
			return err
		}
		if err := root.WriteFile(staged, []byte(files[name]), 0o644); err != nil {
			return err
		}
	}
	for _, name := range names {
		clean := filepath.Clean(name)
		if err := root.Rename(filepath.Join(stage, clean), clean); err != nil {
			return err
		}
	}
	return removeStaleGeneratedFiles(root, desired)
}

// removeStaleGeneratedFiles removes only obsolete regular generated files.
func removeStaleGeneratedFiles(root *os.Root, desired map[string]bool) error {
	const generatedHeader = "Code generated from canonical protocol manifest v2. DO NOT EDIT."
	return fs.WalkDir(root.FS(), ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.Type().IsRegular() || desired[filepath.FromSlash(path)] {
			return nil
		}
		data, err := root.ReadFile(path)
		if err != nil {
			return err
		}
		contents := strings.TrimLeft(string(data), "/# ")
		if strings.HasPrefix(contents, generatedHeader) {
			return root.Remove(path)
		}
		return nil
	})
}
