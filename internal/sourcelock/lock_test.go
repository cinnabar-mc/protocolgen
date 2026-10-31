package sourcelock

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadRejectsMixedProtocolPins(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source-lock.json")
	data := []byte(`{
  "schema_version": 1,
  "target": {"minecraft_version": "1.26.40", "protocol_version": 2168},
  "sources": [
    {"id":"mojang-2168","kind":"mojang","revision":"a","digest":"sha256:a","minecraft_version":"1.26.40","protocol_version":2168},
    {"id":"endstone-2169","kind":"endstone","revision":"b","digest":"sha256:b","minecraft_version":"1.26.50","protocol_version":2169}
  ]
}`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil || !strings.Contains(err.Error(), "mix") {
		t.Fatalf("Load error = %v, want mixed target rejection", err)
	}
}

func TestDirectoryDigestIsStableAndExcludesGitMetadata(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "schema.json"), []byte(`{"x":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	first, err := DigestDirectory(root)
	if err != nil {
		t.Fatalf("DigestDirectory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".git", "HEAD"), []byte("unrelated"), 0o644); err != nil {
		t.Fatal(err)
	}
	second, err := DigestDirectory(root)
	if err != nil {
		t.Fatalf("DigestDirectory second: %v", err)
	}
	if first != second {
		t.Fatalf("digest changed after .git metadata: %q != %q", first, second)
	}
	if err := os.RemoveAll(filepath.Join(root, ".git")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".git"), []byte("gitdir: /different/worktree/location\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	worktree, err := DigestDirectory(root)
	if err != nil {
		t.Fatalf("DigestDirectory worktree: %v", err)
	}
	if first != worktree {
		t.Fatalf("digest changed for worktree metadata: %q != %q", first, worktree)
	}
}

func TestDirectoryDigestRejectsSymlinksIncludingGitMetadata(t *testing.T) {
	for _, name := range []string{"schema.json", ".git"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			if err := os.Symlink(t.TempDir(), filepath.Join(root, name)); err != nil {
				t.Fatal(err)
			}
			if _, err := DigestDirectory(root); err == nil || !strings.Contains(err.Error(), "unsupported symlink") {
				t.Fatalf("DigestDirectory error = %v, want symlink rejection", err)
			}
		})
	}
}

func TestDirectoryDigestRejectsNonDirectoryRoots(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "file")
	if err := os.WriteFile(file, []byte("content"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(root, link); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{file, link} {
		if _, err := DigestDirectory(path); err == nil {
			t.Errorf("accepted root %s", path)
		}
	}
	if _, err := DigestDirectory(t.TempDir()); err != nil {
		t.Fatal(err)
	}
}
