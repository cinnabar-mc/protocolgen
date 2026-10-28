package sourcelock

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
			outside := t.TempDir()
			if err := os.WriteFile(filepath.Join(outside, "outside.json"), []byte(`{"outside":true}`), 0o644); err != nil {
				t.Fatal(err)
			}
			link := filepath.Join(root, name)
			makeDirectoryLink(t, outside, link)
			if _, err := os.Stat(filepath.Join(link, "outside.json")); err != nil {
				t.Fatalf("directory link fixture does not reach its outside-root target: %v", err)
			}
			if _, err := DigestDirectory(root); err == nil {
				t.Fatal("DigestDirectory accepted a directory link")
			}
		})
	}
}

func makeDirectoryLink(t *testing.T, target, link string) {
	t.Helper()
	if runtime.GOOS != "windows" {
		if err := os.Symlink(target, link); err != nil {
			t.Fatal(err)
		}
		return
	}
	// Directory junctions exercise the same no-follow boundary without
	// requiring Windows Developer Mode or an elevated test process.
	if output, err := exec.Command("cmd", "/c", "mklink", "/J", link, target).CombinedOutput(); err != nil {
		t.Fatalf("create directory junction: %v: %s", err, output)
	}
}

func TestLoadValidatesConfiguredSymbolLockPath(t *testing.T) {
	for _, configured := range []string{"symbols-lock.json", "../outside.json", "/outside.json"} {
		t.Run(configured, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "source-lock.json")
			data := `{
  "schema_version": 1,
  "target": {"minecraft_version": "1.26.51", "protocol_version": 2193},
  "symbols_lock": "` + configured + `",
  "sources": [{"id":"wire","kind":"synthetic","revision":"fixture","digest":"fixture:wire"}]
}`
			if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
				t.Fatal(err)
			}
			lock, err := Load(path)
			if configured == "symbols-lock.json" {
				if err != nil || lock.SymbolsLock != configured {
					t.Fatalf("configured supplement lost: %+v, %v", lock, err)
				}
			} else if err == nil {
				t.Fatal("accepted a symbols lock outside the configured snapshot")
			}
		})
	}
}
