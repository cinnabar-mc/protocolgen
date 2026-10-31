package gophertunneloracle

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// checkoutFixture creates a committed protocol source tree without network access.
func checkoutFixture(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "minecraft/protocol/packet"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "minecraft/protocol/packet/fixture.go"), []byte("package packet\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"init"}, {"add", "."}, {"-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-m", "Fixture"}} {
		if err := runGitAt(root, args...); err != nil {
			t.Fatal(err)
		}
	}
	revision, err := gitOutput(root, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	return root, strings.TrimSpace(revision)
}

func TestCheckoutRejectsChangedOracleSources(t *testing.T) {
	for _, name := range []string{"fixture.go", "extra.go", "ignored.go"} {
		t.Run(name, func(t *testing.T) {
			root, revision := checkoutFixture(t)
			if name == "ignored.go" {
				if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte("ignored.go\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.WriteFile(filepath.Join(root, "minecraft/protocol/packet", name), []byte("package changed\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := verifyCheckout(root, revision); err == nil {
				t.Fatal("dirty source tree accepted")
			}
		})
	}
}

func TestCheckoutInitializationRetriesAfterFetchFailure(t *testing.T) {
	root, revision := checkoutFixture(t)
	lock := fixtureLock()
	lock.Gophertunnel.Repo, lock.Gophertunnel.Commit = root, revision
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	marker := filepath.Join(bin, "fail-fetch")
	if err := os.WriteFile(marker, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	wrapper := "#!/bin/sh\nif test \"$1\" = fetch && test -f " + strconv.Quote(marker) + "; then rm " + strconv.Quote(marker) + "; exit 1; fi\nexec " + strconv.Quote(realGit) + " \"$@\"\n"
	if err := os.WriteFile(filepath.Join(bin, "git"), []byte(wrapper), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	cache := t.TempDir()
	if _, err := resolveCheckout(lock, "", cache); err == nil {
		t.Fatal("expected fetch failure")
	}
	if _, err := resolveCheckout(lock, "", cache); err != nil {
		t.Fatalf("retry failed: %v", err)
	}
}

func TestCheckoutRecoversLegacyIncompleteCache(t *testing.T) {
	root, revision := checkoutFixture(t)
	lock := fixtureLock()
	lock.Gophertunnel.Repo, lock.Gophertunnel.Commit = root, revision
	cache := t.TempDir()
	if err := runGit("clone", "--no-checkout", root, filepath.Join(cache, revision)); err != nil {
		t.Fatal(err)
	}
	if _, err := resolveCheckout(lock, "", cache); err != nil {
		t.Fatal(err)
	}
}
