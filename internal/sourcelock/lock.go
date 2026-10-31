// Package sourcelock loads and verifies the explicit source pins used by a
// protocol generation run. It never downloads or vendors source corpora.
package sourcelock

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
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
		size int64
	}
	var entries []entry
	root = filepath.Clean(root)
	info, err := os.Lstat(root)
	if err != nil {
		return "", fmt.Errorf("inspect source directory: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("source root %s must be a directory, got %s", root, info.Mode().Type())
	}
	err = filepath.WalkDir(root, func(path string, directoryEntry fs.DirEntry, err error) error {
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
		if !directoryEntry.Type().IsRegular() {
			return fmt.Errorf("source tree contains unsupported file type %s: %s", rel, directoryEntry.Type())
		}
		info, err := directoryEntry.Info()
		if err != nil {
			return err
		}
		entries = append(entries, entry{path: filepath.ToSlash(rel), size: info.Size()})
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("hash source directory %s: %w", root, err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].path < entries[j].path })
	hash := sha256.New()
	buffer := make([]byte, 32*1024)
	for _, item := range entries {
		file, err := os.Open(filepath.Join(root, filepath.FromSlash(item.path)))
		if err != nil {
			return "", fmt.Errorf("open source file %s: %w", item.path, err)
		}
		info, err := file.Stat()
		if err != nil || !info.Mode().IsRegular() || info.Size() != item.size {
			file.Close()
			return "", fmt.Errorf("source file %s changed while hashing", item.path)
		}
		fmt.Fprintf(hash, "%d:%s\n%d:", len(item.path), item.path, item.size)
		count, copyErr := io.CopyBuffer(hash, io.LimitReader(file, item.size+1), buffer)
		closeErr := file.Close()
		if copyErr != nil {
			return "", fmt.Errorf("hash source file %s: %w", item.path, copyErr)
		}
		if closeErr != nil {
			return "", fmt.Errorf("close source file %s: %w", item.path, closeErr)
		}
		if count != item.size {
			return "", fmt.Errorf("source file %s changed while hashing: read %d bytes, expected %d", item.path, count, item.size)
		}
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
