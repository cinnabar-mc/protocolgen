package differential

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const gophertunnelModule = "github.com/sandertv/gophertunnel"

// Every module under differential/ must link the gophertunnel fork commit
// gophertunnel.json names; scripts/pin-gophertunnel.sh rewrites them all.
func TestGophertunnelPinsMatchTheLock(t *testing.T) {
	data, err := os.ReadFile("gophertunnel.json")
	if err != nil {
		t.Fatal(err)
	}
	var lock struct {
		Module  string `json:"module"`
		Version string `json:"version"`
	}
	if err := json.Unmarshal(data, &lock); err != nil || lock.Module == "" || lock.Version == "" {
		t.Fatalf("gophertunnel.json: incomplete lock (%v)", err)
	}
	want := lock.Module + " " + lock.Version
	modules := 0
	err = filepath.WalkDir(".", func(path string, entry fs.DirEntry, err error) error {
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
		if got := replacement(string(text)); got != want {
			t.Errorf("%s replaces %s with %q, gophertunnel.json pins %q", path, gophertunnelModule, got, want)
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
