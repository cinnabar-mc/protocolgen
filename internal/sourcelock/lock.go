// Package sourcelock loads and verifies the explicit source pins used by a
// protocol generation run. It never downloads or vendors source corpora.
package sourcelock

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"protocolgen/internal/manifest"
)

const SchemaVersion uint32 = 1

type Lock struct {
	SchemaVersion uint32               `json:"schema_version"`
	Target        manifest.Target      `json:"target"`
	Sources       []manifest.SourcePin `json:"sources"`
	SymbolsLock   string               `json:"symbols_lock,omitempty"`
	// Non-wire reference catalogs may describe a different source version.
	Catalog   string           `json:"catalog,omitempty"`
	AppliesTo *manifest.Target `json:"applies_to,omitempty"`
}

func Load(path string) (Lock, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Lock{}, fmt.Errorf("read source lock: %w", err)
	}
	var lock Lock
	if err := json.Unmarshal(data, &lock); err != nil {
		return Lock{}, fmt.Errorf("parse source lock: %w", err)
	}
	if lock.SchemaVersion != SchemaVersion {
		return Lock{}, fmt.Errorf("source lock schema_version %d is not v%d", lock.SchemaVersion, SchemaVersion)
	}
	if lock.Target.MinecraftVersion == "" || lock.Target.ProtocolVersion <= 0 {
		return Lock{}, fmt.Errorf("source lock has no complete target")
	}
	if len(lock.Sources) == 0 {
		return Lock{}, fmt.Errorf("source lock has no sources")
	}
	if lock.SymbolsLock != "" && !filepath.IsLocal(lock.SymbolsLock) {
		return Lock{}, fmt.Errorf("source lock symbols_lock must be a local relative path")
	}
	seen := map[string]bool{}
	for _, source := range lock.Sources {
		if source.ID == "" || source.Kind == "" || source.Revision == "" || source.Digest == "" {
			return Lock{}, fmt.Errorf("source lock contains an incomplete source pin")
		}
		if seen[source.ID] {
			return Lock{}, fmt.Errorf("source lock duplicates source %q", source.ID)
		}
		seen[source.ID] = true
		if source.ProtocolVersion != 0 && source.ProtocolVersion != lock.Target.ProtocolVersion {
			return Lock{}, fmt.Errorf("source lock mixes protocol %d into target protocol %d", source.ProtocolVersion, lock.Target.ProtocolVersion)
		}
		if source.MinecraftVersion != "" && source.MinecraftVersion != lock.Target.MinecraftVersion {
			return Lock{}, fmt.Errorf("source lock mixes Minecraft %s into target Minecraft %s", source.MinecraftVersion, lock.Target.MinecraftVersion)
		}
	}
	return lock, nil
}

func (l Lock) Source(id string) (manifest.SourcePin, bool) {
	for _, source := range l.Sources {
		if source.ID == id {
			return source, true
		}
	}
	return manifest.SourcePin{}, false
}

func VerifyDirectory(root string, pin manifest.SourcePin) error {
	got, err := DigestDirectory(root)
	if err != nil {
		return err
	}
	if got != pin.Digest {
		return fmt.Errorf("source %q digest mismatch: lock has %s, directory has %s", pin.ID, pin.Digest, got)
	}
	return nil
}

// DigestDirectory hashes sorted relative paths and their bytes. Git metadata
// and symlinks are excluded/rejected so a checkout's object database cannot
// affect a source pin and a path cannot escape the locked root.
func DigestDirectory(root string) (string, error) {
	type entry struct {
		path string
		data []byte
	}
	var entries []entry
	root = filepath.Clean(root)
	err := filepath.WalkDir(root, func(path string, directoryEntry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if directoryEntry.IsDir() && (directoryEntry.Name() == ".git" || directoryEntry.Name() == ".hg" || directoryEntry.Name() == ".svn") {
			return fs.SkipDir
		}
		if directoryEntry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("source tree contains unsupported symlink %s", rel)
		}
		if directoryEntry.Name() == ".git" && directoryEntry.Type().IsRegular() {
			return nil
		}
		if directoryEntry.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		entries = append(entries, entry{path: filepath.ToSlash(rel), data: data})
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("hash source directory %s: %w", root, err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].path < entries[j].path })
	hash := sha256.New()
	for _, item := range entries {
		fmt.Fprintf(hash, "%d:%s\n", len(item.path), item.path)
		fmt.Fprintf(hash, "%d:", len(item.data))
		_, _ = hash.Write(item.data)
		_, _ = hash.Write([]byte{'\n'})
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), nil
}

func Write(path string, lock Lock) error {
	if lock.SchemaVersion == 0 {
		lock.SchemaVersion = SchemaVersion
	}
	data, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if parent := filepath.Dir(path); parent != "." && parent != "" {
		if err := os.MkdirAll(parent, 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, data, 0o644)
}

func IsSyntheticDigest(digest string) bool { return strings.HasPrefix(digest, "fixture:") }
