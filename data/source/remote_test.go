package source

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestReadRemoteAuthenticatesCache preserves digest checks without reading bundled inputs.
func TestReadRemoteAuthenticatesCache(t *testing.T) {
	payload := []byte("synthetic remote metadata")
	digest := fmt.Sprintf("%x", sha256.Sum256(payload))
	cache := t.TempDir()
	path := filepath.Join(cache, digest)
	inputs := &Sources{Lock: Lock{Inputs: map[string]Input{
		"remote": {SHA256: digest},
		"local":  {Path: "inputs/local.json", SHA256: digest},
	}}}
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	if got, err := inputs.ReadRemote(context.Background(), "remote", cache); err != nil || !bytes.Equal(got, payload) {
		t.Fatalf("authenticated remote cache = %q, %v", got, err)
	}
	for _, name := range []string{"local", "missing"} {
		if _, err := inputs.ReadRemote(context.Background(), name, cache); err == nil {
			t.Fatalf("accepted non-remote input %q", name)
		}
	}
	if err := os.WriteFile(path, []byte("changed"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := inputs.ReadRemote(context.Background(), "remote", cache); err == nil || !strings.Contains(err.Error(), "digest") {
		t.Fatalf("changed cached bytes accepted: %v", err)
	}
}

// TestRemoteClientOmitsGeneratorPayloads checks the linked executable, not source imports.
func TestRemoteClientOmitsGeneratorPayloads(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "remote-client")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	command := exec.Command("go", "build", "-o", binary, "./testdata/remoteclient")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build remote client: %v\n%s", err, output)
	}
	if output, err := exec.Command(binary).CombinedOutput(); err != nil {
		t.Fatalf("run remote client: %v\n%s", err, output)
	}
	contents, err := os.ReadFile(binary)
	if err != nil {
		t.Fatal(err)
	}
	err = fs.WalkDir(bundled, "inputs", func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		payload, err := bundled.ReadFile(path)
		if err != nil {
			return err
		}
		// Small sequences can occur in ordinary program strings by coincidence.
		if len(payload) >= 1024 && bytes.Contains(contents, payload) {
			t.Errorf("remote-only executable contains generator input %s (%d bytes)", path, len(payload))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
