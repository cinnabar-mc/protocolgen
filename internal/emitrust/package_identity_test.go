package emitrust

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"protocolgen/internal/emitter"
	"protocolgen/internal/manifest"
)

// Exercise Cargo's recursive Git-package discovery, which may select the wrong
// profile when two generated crates have the same package name and version.
func TestContractGitPackageIdentity(t *testing.T) {
	for _, tool := range []string{"cargo", "git"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("missing fixture: %s is required for Cargo Git package selection", tool)
		}
	}
	profiles, err := filepath.Glob("../../generated/*/rust-contract/Cargo.toml")
	if err != nil || len(profiles) == 0 {
		t.Fatalf("discover active contract profiles: %v (%d profiles)", err, len(profiles))
	}
	for _, profile := range profiles {
		m, err := manifest.Load(filepath.Join(filepath.Dir(filepath.Dir(profile)), "manifest.json"))
		if err != nil {
			t.Fatal(err)
		}
		t.Run(m.Target.MinecraftVersion, func(t *testing.T) {
			// Package identity depends only on the active manifest's game version;
			// a small schema keeps discovery independent of optional codec dependencies.
			fixture := contractFixture()
			fixture.Target.MinecraftVersion = m.Target.MinecraftVersion
			fixture.Sources[0].MinecraftVersion = m.Target.MinecraftVersion
			repository := filepath.Join(t.TempDir(), "repository")
			suffix := strings.NewReplacer(".", "-", "_", "-").Replace(m.Target.MinecraftVersion)
			plainName, contractName := "bedrock-protocol-"+suffix, "bedrock-protocol-contract-"+suffix
			for _, output := range []struct {
				profile, name string
				contract      int
			}{{"plain", plainName, 0}, {"contract", contractName, CodecContractVersion}} {
				files, err := GenerateFilesWithOptions(fixture, Options{CodecContract: output.contract})
				if err != nil {
					t.Fatal(err)
				}
				directory := filepath.Join(repository, output.profile)
				if err := emitter.WriteFiles(directory, files); err != nil {
					t.Fatal(err)
				}
				var pkg struct{ Name string }
				if err := json.Unmarshal(packageIdentityCommand(t, repository, "cargo", "read-manifest", "--manifest-path", filepath.Join(directory, "Cargo.toml")), &pkg); err != nil {
					t.Fatal(err)
				}
				if pkg.Name != output.name {
					t.Fatalf("%s package = %q, want %q", output.profile, pkg.Name, output.name)
				}
			}
			packageIdentityCommand(t, repository, "git", "init", "-q")
			packageIdentityCommand(t, repository, "git", "add", ".")
			packageIdentityCommand(t, repository, "git", "-c", "user.name=Protocolgen Fixture", "-c", "user.email=fixture@example.invalid", "-c", "commit.gpgsign=false", "-c", "core.hooksPath=/dev/null", "commit", "-qm", "Generated profiles")
			consumer := filepath.Join(t.TempDir(), "consumer")
			// A local registry replacement and bytes patch prevent network access;
			// Cargo does not compile the codec or this metadata-only dependency stub.
			cargo := fmt.Sprintf("[package]\nname = %q\nversion = %q\nedition = %q\n[dependencies]\nselected = { package = %q, git = %q }\n[patch.crates-io]\nbytes = { path = %q }\n", "consumer", "0.1.0", "2024", contractName, "file://"+filepath.ToSlash(repository), "bytes")
			if err := emitter.WriteFiles(consumer, map[string]string{
				"Cargo.toml": cargo, "src/lib.rs": "",
				".cargo/config.toml": fmt.Sprintf("[source.crates-io]\nreplace-with = \"fixture-registry\"\n[source.fixture-registry]\ndirectory = %q\n", filepath.Join(consumer, "empty-registry")),
				"bytes/Cargo.toml":   "[package]\nname = \"bytes\"\nversion = \"1.0.0\"\nedition = \"2024\"\n",
				"bytes/src/lib.rs":   "",
			}); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Join(consumer, "empty-registry"), 0o755); err != nil {
				t.Fatal(err)
			}
			var metadata struct {
				Packages []struct {
					Name         string
					ManifestPath string `json:"manifest_path"`
				}
			}
			if err := json.Unmarshal(packageIdentityCommand(t, consumer, "cargo", "metadata", "--format-version=1"), &metadata); err != nil {
				t.Fatal(err)
			}
			for _, pkg := range metadata.Packages {
				if pkg.Name == contractName && filepath.Base(filepath.Dir(pkg.ManifestPath)) == "contract" {
					return
				}
			}
			t.Fatal("Cargo Git dependency did not select the emitted contract profile")
		})
	}
}

// packageIdentityCommand runs metadata-only tools against an isolated fixture.
func packageIdentityCommand(t *testing.T, directory, tool string, args ...string) []byte {
	t.Helper()
	command := exec.Command(tool, args...)
	command.Dir = directory
	command.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null")
	output, err := command.Output()
	if err != nil {
		if failure, ok := err.(*exec.ExitError); ok {
			t.Fatalf("%s %v: %v\n%s", tool, args, err, failure.Stderr)
		}
		t.Fatalf("%s %v: %v", tool, args, err)
	}
	return output
}
