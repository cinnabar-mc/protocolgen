package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// writeToolFixture writes one input for a build-tool integration test.
func writeToolFixture(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

// runTool runs a fixture command and requires a successful exit.
func runTool(t *testing.T, directory, name string, args ...string) string {
	t.Helper()
	command := exec.Command(name, args...)
	command.Dir = directory
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, output)
	}
	return strings.TrimSpace(string(output))
}

// initToolRepository creates an isolated committed tree without project hooks.
func initToolRepository(t *testing.T, directory string) {
	t.Helper()
	runTool(t, directory, "git", "init", "--quiet")
	runTool(t, directory, "git", "add", ".")
	runTool(t, directory, "git", "-c", "user.name=Test", "-c", "user.email=test@example.com", "commit", "--quiet", "-m", "Initial fixture")
}

// fixtureMakefile copies the real recipes and appends only test command overrides.
func fixtureMakefile(t *testing.T, directory, overrides string) string {
	t.Helper()
	data, err := os.ReadFile("../../Makefile")
	if err != nil {
		t.Fatal(err)
	}
	index, err := os.ReadFile("../../data/source/releases.json")
	if err != nil {
		t.Fatal(err)
	}
	writeToolFixture(t, filepath.Join(directory, "data/source/releases.json"), string(index))
	path := filepath.Join(directory, "Makefile")
	writeToolFixture(t, path, string(data)+"\n"+overrides)
	return path
}

func TestMakePassesFilesystemPathsAsSingleArguments(t *testing.T) {
	directory := t.TempDir()
	fixtureMakefile(t, directory, "define emit-rust\n\t@:\nendef\n")
	log := filepath.Join(directory, "arguments.jsonl")
	stub := filepath.Join(directory, "capture.py")
	writeToolFixture(t, stub, "import json, sys\nwith open(sys.argv[1], 'a') as output: output.write(json.dumps(sys.argv[2:]) + '\\n')\n")
	index, err := filepath.Abs("../../data/source/releases.json")
	if err != nil {
		t.Fatal(err)
	}
	command := "python3 '" + stub + "' '" + log + "'"
	wants := map[string]string{"-root": "mojang source path", "-mojang": "mojang source path", "-endstone": "endstone source path", "-bds-binary": "server binary path"}
	for _, target := range []string{"regen", "ingest-1.26.60", "vanilla-data"} {
		runTool(t, directory, "make", target, "GO="+command, "PROTOCOLGEN="+command, "RELEASE_INDEX="+index, "MOJANG_DIR=mojang source path", "ENDSTONE_DIR=endstone source path", "BDS_BINARY=server binary path")
	}
	data, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		var args []string
		if err := json.Unmarshal([]byte(line), &args); err != nil {
			t.Fatal(err)
		}
		for i, arg := range args {
			want, ok := wants[arg]
			if !ok {
				continue
			}
			if arg == "-root" && i+1 < len(args) && args[i+1] == "endstone source path" {
				continue
			}
			seen[arg] = true
			if i+1 >= len(args) || args[i+1] != want {
				t.Errorf("%s lost path %q in %q", arg, want, args)
			}
		}
	}
	for flag := range wants {
		if !seen[flag] {
			t.Errorf("did not exercise %s", flag)
		}
	}
}

func TestParallelVerifyFinishesRegenerationBeforeHotfix(t *testing.T) {
	directory := t.TempDir()
	fixtureMakefile(t, directory, "define emit-rust\n\t@:\nendef\n")
	log := filepath.Join(directory, "commands.jsonl")
	stub := filepath.Join(directory, "capture.py")
	writeToolFixture(t, stub, "import json, sys\nwith open(sys.argv[1], 'a') as output: output.write(json.dumps(sys.argv[2:]) + '\\n')\n")
	initToolRepository(t, directory)
	command := "python3 '" + stub + "' '" + log + "'"
	runTool(t, directory, "make", "-j3", "verify", "GO="+command, "PROTOCOLGEN="+command, "MOJANG_DIR=mojang", "ENDSTONE_DIR=endstone")
	data, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	regenEnd, hotfixStart := -1, -1
	for i, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		var args []string
		if err := json.Unmarshal([]byte(line), &args); err != nil {
			t.Fatal(err)
		}
		if len(args) > 0 && args[0] == "vet" {
			regenEnd = i
		}
		if len(args) > 0 && args[0] == "hotfix" {
			hotfixStart = i
		}
	}
	if regenEnd < 0 || hotfixStart <= regenEnd {
		t.Fatalf("hotfix started at %d before regeneration finished at %d: %s", hotfixStart, regenEnd, data)
	}
}

func TestMakeDriftChecksIncludeStagedChanges(t *testing.T) {
	for target, outputDir := range map[string]string{"verify-1.26.60": "generated/1.26.60", "verify-data": "generated/data"} {
		for _, state := range []string{"modified", "added"} {
			t.Run(target+"/"+state, func(t *testing.T) {
				directory := t.TempDir()
				fixtureMakefile(t, directory, "regen-1.26.60 generate-data:\n\t@:\n")
				path := filepath.Join(directory, outputDir, "existing.txt")
				writeToolFixture(t, path, "committed")
				initToolRepository(t, directory)
				if state == "added" {
					path = filepath.Join(directory, outputDir, "new.txt")
				}
				writeToolFixture(t, path, "regenerated")
				runTool(t, directory, "git", "add", path)
				command := exec.Command("make", target)
				command.Dir = directory
				output, err := command.CombinedOutput()
				if err == nil {
					t.Fatalf("staged %s escaped drift check: %s", state, output)
				}
			})
		}
	}
}

func TestMakeRustEmissionPreservesOverlayArguments(t *testing.T) {
	directory := t.TempDir()
	fixtureMakefile(t, directory, "")
	log := filepath.Join(directory, "arguments.jsonl")
	stub := filepath.Join(directory, "capture.py")
	writeToolFixture(t, stub, "import json, sys\nwith open(sys.argv[1], 'a') as output: output.write(json.dumps(sys.argv[2:]) + '\\n')\n")
	bin := filepath.Join(directory, "bin")
	cargo := filepath.Join(bin, "cargo")
	writeToolFixture(t, cargo, "#!/bin/sh\nexit 0\n")
	if err := os.Chmod(cargo, 0755); err != nil {
		t.Fatal(err)
	}
	snapshot := "snapshot with spaces"
	command := exec.Command("make", "emit-rust", "SNAPSHOT_DIR="+snapshot, "PROTOCOLGEN=python3 '"+stub+"' '"+log+"'")
	command.Dir = directory
	command.Env = append(os.Environ(), "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("emit-rust: %v: %s", err, output)
	}
	data, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	var args []string
	if err := json.Unmarshal(data, &args); err != nil {
		t.Fatal(err)
	}
	want := []string{"emit-rust", "-manifest", snapshot + "/manifest.json", "-out", snapshot + "/rust", "-naming", snapshot + "/naming.json", "-domains", snapshot + "/domains.json", "-docs", snapshot + "/docs.json"}
	if strings.Join(args, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("overlay arguments = %q, want %q", args, want)
	}
}

// fetchFixture creates a local source repository and valid corpus metadata.
func fetchFixture(t *testing.T) (string, map[string]any) {
	t.Helper()
	repository := t.TempDir()
	for _, directory := range []string{"fixtures", "codec-a", "codec-b"} {
		writeToolFixture(t, filepath.Join(repository, directory, "fixture.txt"), directory)
	}
	writeToolFixture(t, filepath.Join(repository, "tools", "fixturegen", "go.mod"), "module fixture\n")
	initToolRepository(t, repository)
	commit := runTool(t, repository, "git", "rev-parse", "HEAD")
	return repository, map[string]any{"source": map[string]any{"repository": repository, "commit": commit, "directory": "fixtures"}, "valentine": map[string]any{"commit": commit, "directory": "codec-a"}}
}

// writeCorpus serializes metadata used by the fetch script.
func writeCorpus(t *testing.T, path string, metadata map[string]any) {
	t.Helper()
	data, err := json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}
	writeToolFixture(t, path, string(data))
}

func TestFetchRejectsMissingDirectoriesBeforeCheckout(t *testing.T) {
	script, err := filepath.Abs("../../scripts/fetch-cinnabar.sh")
	if err != nil {
		t.Fatal(err)
	}
	for _, section := range []string{"source", "valentine"} {
		t.Run(section, func(t *testing.T) {
			_, metadata := fetchFixture(t)
			delete(metadata[section].(map[string]any), "directory")
			directory := t.TempDir()
			corpus := filepath.Join(directory, "corpus.json")
			writeCorpus(t, corpus, metadata)
			checkout, valentine := filepath.Join(directory, "checkout"), filepath.Join(directory, "valentine")
			output, err := exec.Command("bash", script, corpus, checkout, valentine).CombinedOutput()
			if err == nil {
				t.Errorf("missing %s.directory accepted: %s", section, output)
			}
			for _, path := range []string{checkout, valentine} {
				if _, err := os.Stat(path); !os.IsNotExist(err) {
					t.Errorf("checkout began before validation: %s, %v", path, err)
				}
			}
		})
	}
}

func TestFetchRefreshesSparsePathsAtMatchingCommit(t *testing.T) {
	script, err := filepath.Abs("../../scripts/fetch-cinnabar.sh")
	if err != nil {
		t.Fatal(err)
	}
	_, metadata := fetchFixture(t)
	directory := t.TempDir()
	corpus, checkout, valentine := filepath.Join(directory, "corpus.json"), filepath.Join(directory, "checkout"), filepath.Join(directory, "valentine")
	writeCorpus(t, corpus, metadata)
	runTool(t, directory, "bash", script, corpus, checkout, valentine)
	metadata["valentine"].(map[string]any)["directory"] = "codec-b"
	writeCorpus(t, corpus, metadata)
	runTool(t, directory, "bash", script, corpus, checkout, valentine)
	if _, err := os.Stat(filepath.Join(valentine, "codec-b", "fixture.txt")); err != nil {
		t.Fatalf("new sparse path not materialized: %v", err)
	}
	if _, err := os.Stat(filepath.Join(valentine, "codec-a", "fixture.txt")); !os.IsNotExist(err) {
		t.Fatalf("old sparse path retained: %v", err)
	}
}
