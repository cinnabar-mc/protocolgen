package source

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSourceLockAuthenticatesLocalInputs(t *testing.T) {
	dir := t.TempDir()
	payload := []byte("synthetic input")
	if err := os.WriteFile(filepath.Join(dir, "input.json"), payload, 0o644); err != nil {
		t.Fatal(err)
	}
	lock := Lock{SchemaVersion: 2, Release: "1.26.51", Upstreams: map[string]Upstream{"cloudburst": {Repository: "example/data", Revision: strings.Repeat("1", 40)}}, Semantic: SemanticInputs{Cloudburst: "cloudburst"}, Inputs: map[string]Input{
		"fixture": {Path: "input.json", SHA256: fmt.Sprintf("%x", sha256.Sum256(payload)), Revision: "fixture"},
	}}
	encoded, _ := json.Marshal(lock)
	path := filepath.Join(dir, "lock.json")
	if err := os.WriteFile(path, encoded, 0o644); err != nil {
		t.Fatal(err)
	}
	sources, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := sources.Read(context.Background(), "fixture", ""); err != nil || string(got) != string(payload) {
		t.Fatalf("verified read = %q, %v", got, err)
	}
	if err := os.WriteFile(filepath.Join(dir, "input.json"), []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := sources.Read(context.Background(), "fixture", ""); err == nil || !strings.Contains(err.Error(), "digest") {
		t.Fatalf("changed source was accepted: %v", err)
	}
	if _, err := sources.Read(context.Background(), "unknown", ""); err == nil {
		t.Fatal("unlocked source was accepted")
	}
}

func TestBundledSourceInputsAndTarget(t *testing.T) {
	sources, err := Open("")
	if err != nil {
		t.Fatal(err)
	}
	if sources.Lock.Target.MinecraftVersion != "1.26.60" || sources.Lock.Target.ProtocolVersion != 2225 {
		t.Fatalf("unexpected migrated target: %+v", sources.Lock)
	}
	for name, input := range sources.Lock.Inputs {
		if input.Path != "" {
			if _, err := sources.Read(context.Background(), name, ""); err != nil {
				t.Fatal(err)
			}
		}
	}
}

// TestGroupedSourcesRejectDrift checks relationships that hashes alone cannot prove.
func TestGroupedSourcesRejectDrift(t *testing.T) {
	tests := []struct {
		name   string
		change func(*Lock)
	}{
		{"unknown release", func(l *Lock) { l.Release = "missing" }},
		{"duplicate revision", func(l *Lock) {
			i := l.Inputs["item_registry"]
			i.Revision = l.CloudburstRevision()
			l.Inputs["item_registry"] = i
		}},
		{"mixed registry", func(l *Lock) { i := l.Inputs["item_registry"]; i.Upstream = "allay"; l.Inputs["item_registry"] = i }},
		{"unknown upstream", func(l *Lock) {
			i := l.Inputs["native_item_properties"]
			i.Upstream = "missing"
			l.Inputs["native_item_properties"] = i
		}},
		{"moving ref", func(l *Lock) { u := l.Upstreams["cloudburst"]; u.Revision = "master"; l.Upstreams["cloudburst"] = u }},
		{"ambiguous input", func(l *Lock) { i := l.Inputs["item_registry"]; i.Path = "local.json"; l.Inputs["item_registry"] = i }},
		{"URL path escape", func(l *Lock) { i := l.Inputs["item_registry"]; i.File = "../other.json"; l.Inputs["item_registry"] = i }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sources, err := Open("")
			if err != nil {
				t.Fatal(err)
			}
			test.change(&sources.Lock)
			encoded, err := json.Marshal(sources.Lock)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), "lock.json")
			if err := os.WriteFile(path, encoded, 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := Open(path); err == nil {
				t.Fatal("accepted an inconsistent source lock")
			}
		})
	}
}

// TestSourceURLsDeriveFromOneRevision prevents stale URLs after a pin update.
func TestSourceURLsDeriveFromOneRevision(t *testing.T) {
	sources, err := Open("")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"block_metadata", "resource_pack"} {
		if !strings.Contains(sources.URL(name), sources.Revision(name)) {
			t.Fatalf("%s URL does not use its source revision", name)
		}
	}
	upstream := sources.Lock.Upstreams["cloudburst"]
	upstream.Revision = strings.Repeat("a", 40)
	sources.Lock.Upstreams["cloudburst"] = upstream
	if got := sources.URL("item_registry"); got != "https://raw.githubusercontent.com/CloudburstMC/Data/"+upstream.Revision+"/runtime_item_states.json" {
		t.Fatal(got)
	}
	if sources.URL("liquid_clip_omissions") != "" || sources.URL("missing") != "" {
		t.Fatal("local or unknown input has a URL")
	}
	before, err := sources.Lock.SHA256()
	if err != nil {
		t.Fatal(err)
	}
	sources.Lock.Target.ProtocolVersion++
	after, err := sources.Lock.SHA256()
	if err != nil {
		t.Fatal(err)
	}
	if before == after {
		t.Fatal("release target is absent from lock identity")
	}
}
