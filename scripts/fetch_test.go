package scripts

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// gitFixture runs Git in an isolated test repository with a local commit identity.
func gitFixture(t *testing.T, root string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=Fixture", "GIT_AUTHOR_EMAIL=fixture@example.invalid", "GIT_COMMITTER_NAME=Fixture", "GIT_COMMITTER_EMAIL=fixture@example.invalid")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func TestFetchCinnabarUpdatesReusedOrigins(t *testing.T) {
	root := t.TempDir()
	checkout, valentine := filepath.Join(root, "checkout"), filepath.Join(root, "valentine")
	for _, name := range []string{"one", "two"} {
		repo := filepath.Join(root, name)
		if err := os.MkdirAll(repo, 0755); err != nil {
			t.Fatal(err)
		}
		gitFixture(t, repo, "init", "--quiet")
		for _, dir := range []string{"fixtures", "codec", "tools/fixturegen"} {
			if err := os.MkdirAll(filepath.Join(repo, dir), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(repo, dir, "value"), []byte(name), 0600); err != nil {
				t.Fatal(err)
			}
		}
		gitFixture(t, repo, "add", ".")
		gitFixture(t, repo, "commit", "--quiet", "-m", name)
		commit := gitFixture(t, repo, "rev-parse", "HEAD")
		corpus := map[string]any{"source": map[string]string{"repository": repo, "commit": commit, "directory": "fixtures"}, "valentine": map[string]string{"commit": commit, "directory": "codec"}}
		data, err := json.Marshal(corpus)
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(root, "corpus.json")
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
		out, err := exec.Command("bash", "fetch-cinnabar.sh", path, checkout, valentine).CombinedOutput()
		if err != nil {
			t.Fatalf("checkout %s: %v\n%s", name, err, out)
		}
		for dir, file := range map[string]string{checkout: "fixtures/value", valentine: "codec/value"} {
			if got := gitFixture(t, dir, "remote", "get-url", "origin"); got != repo {
				t.Fatalf("origin = %q, want %q", got, repo)
			}
			content, err := os.ReadFile(filepath.Join(dir, file))
			if err != nil || string(content) != name {
				t.Fatalf("checkout content = %q, %v", content, err)
			}
		}
	}
}
