package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommandRejectsMissingSourcesWithoutReplacingOutputs(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "seed")
	if output, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		t.Fatalf("build: %v: %s", err, output)
	}
	manifest, err := filepath.Abs("../../testdata/parity/v2-small.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, state := range []string{"missing", "empty", "missing packet", "wrong package", "tests only"} {
		t.Run(state, func(t *testing.T) {
			directory := t.TempDir()
			root := filepath.Join(directory, "checkout")
			if state != "missing" {
				protocol := filepath.Join(root, "minecraft", "protocol")
				if err := os.MkdirAll(protocol, 0755); err != nil {
					t.Fatal(err)
				}
				if state != "empty" {
					filename, content := "fixture.go", "package protocol\ntype Fixture struct{}\n"
					if state == "tests only" {
						filename = "fixture_test.go"
					}
					if state == "wrong package" {
						content = "package other\ntype Fixture struct{}\n"
					}
					if err := os.WriteFile(filepath.Join(protocol, filename), []byte(content), 0644); err != nil {
						t.Fatal(err)
					}
				}
			}
			naming := filepath.Join(directory, "naming.json")
			if err := os.WriteFile(naming, []byte(`{"schema_version":1,"target":{"minecraft_version":"1.26.40","protocol_version":2168},"entries":[]}`), 0644); err != nil {
				t.Fatal(err)
			}
			args := []string{"-manifest", manifest, "-naming", naming, "-gophertunnel", root}
			var outputs []string
			for _, flag := range []string{"-domains-out", "-docs-out"} {
				path := filepath.Join(directory, strings.TrimPrefix(flag, "-"))
				if err := os.WriteFile(path, []byte("sentinel"), 0644); err != nil {
					t.Fatal(err)
				}
				args = append(args, flag, path)
				outputs = append(outputs, path)
			}
			output, err := exec.Command(binary, args...).CombinedOutput()
			if err == nil || !strings.Contains(string(output), "source") {
				t.Errorf("invalid checkout accepted or unclear error: %v: %s", err, output)
			}
			for _, path := range outputs {
				got, err := os.ReadFile(path)
				if err != nil || string(got) != "sentinel" {
					t.Errorf("output replaced: %s = %q, %v", path, got, err)
				}
			}
		})
	}
}
