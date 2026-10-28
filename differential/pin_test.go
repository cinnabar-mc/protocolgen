package differential

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const gophertunnelModule = "github.com/sandertv/gophertunnel"

// Every module under differential/ must link the gophertunnel fork commit named by the
// gophertunnel.json beside its go.mod; scripts/pin-gophertunnel.sh rewrites them.
func TestGophertunnelPinsMatchTheLock(t *testing.T) {
	modules := 0
	err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || entry.Name() != "go.mod" {
			return err
		}
		text, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !strings.Contains(string(text), gophertunnelModule+" ") {
			return nil
		}
		modules++
		want, err := lockedGophertunnel(filepath.Join(filepath.Dir(path), "gophertunnel.json"))
		if err != nil {
			t.Error(err)
			return nil
		}
		if got := replacement(string(text)); got != want {
			t.Errorf("%s replaces %s with %q, its gophertunnel.json pins %q", path, gophertunnelModule, got, want)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if modules < 2 {
		t.Fatalf("found %d modules linking gophertunnel, want the differential and retail modules", modules)
	}
}

// lockedGophertunnel returns "module version" from a gophertunnel.json lock.
func lockedGophertunnel(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var lock struct {
		Module  string `json:"module"`
		Version string `json:"version"`
	}
	if err := json.Unmarshal(data, &lock); err != nil || lock.Module == "" || lock.Version == "" {
		return "", fmt.Errorf("%s: incomplete lock (%v)", path, err)
	}
	return lock.Module + " " + lock.Version, nil
}

// replacement returns "path version" of the gophertunnel replace directive in
// a go.mod file, or "" when it has none.
func replacement(goMod string) string {
	for _, line := range strings.Split(goMod, "\n") {
		fields := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "replace"))
		if len(fields) == 4 && fields[0] == gophertunnelModule && fields[1] == "=>" {
			return fields[2] + " " + fields[3]
		}
	}
	return ""
}
