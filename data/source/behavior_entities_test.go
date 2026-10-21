package source

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestBehaviorEntities_AuthenticateArchiveAndSelectDefinitions(t *testing.T) {
	archive := sampleArchive(t, []tar.Header{
		{Name: "samples/behavior_pack/entities/test.json", Typeflag: tar.TypeReg},
		{Name: "samples/behavior_pack/items/test.json", Typeflag: tar.TypeReg},
		{Name: "samples/resource_pack/entity/test.json", Typeflag: tar.TypeReg},
	})
	cache := t.TempDir()
	digest := fmt.Sprintf("%x", sha256.Sum256(archive))
	sources := &Sources{Lock: Lock{
		Inputs:    map[string]Input{"resource_pack": {Upstream: "samples", Archive: true, SHA256: digest}},
		Upstreams: map[string]Upstream{"samples": {Repository: "Mojang/bedrock-samples", Revision: "unused-local-fixture"}},
	}}
	path := filepath.Join(cache, digest)
	if err := os.WriteFile(path, archive, 0o600); err != nil {
		t.Fatal(err)
	}
	files, err := sources.BehaviorEntities(context.Background(), cache)
	if err != nil || len(files) != 1 || string(files["behavior_pack/entities/test.json"]) != "{}" {
		t.Fatalf("BehaviorEntities = %v, %v", files, err)
	}
	if err := os.WriteFile(path, []byte("changed"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := sources.BehaviorEntities(context.Background(), cache); err == nil {
		t.Fatal("unauthenticated cached archive was accepted")
	}
}

func TestBehaviorEntities_RejectUnsafeOrAmbiguousEntries(t *testing.T) {
	for _, test := range []struct {
		name    string
		headers []tar.Header
	}{
		{"no entities", []tar.Header{{Name: "samples/resource_pack/manifest.json", Typeflag: tar.TypeReg}}},
		{"traversal", []tar.Header{{Name: "samples/behavior_pack/entities/../../escape.json", Typeflag: tar.TypeReg}}},
		{"link", []tar.Header{{Name: "samples/behavior_pack/entities/test.json", Typeflag: tar.TypeSymlink, Linkname: "elsewhere"}}},
		{"duplicate", []tar.Header{{Name: "samples/behavior_pack/entities/test.json", Typeflag: tar.TypeReg}, {Name: "samples/behavior_pack/entities/test.json", Typeflag: tar.TypeReg}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := readBehaviorEntities(bytes.NewReader(sampleArchive(t, test.headers))); err == nil {
				t.Fatal("unsafe or ambiguous archive accepted")
			}
		})
	}
}
