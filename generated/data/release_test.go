package data_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/cinnabar-mc/protocolgen/generated/data"
	"github.com/cinnabar-mc/protocolgen/generated/data/block"
	"github.com/cinnabar-mc/protocolgen/generated/data/registry"
)

// TestReleaseMetadata verifies that both projections used the same complete lock.
func TestReleaseMetadata(t *testing.T) {
	release, err := os.ReadFile("release.json")
	if err != nil {
		t.Fatal(err)
	}
	semantic, err := os.ReadFile("semantic_sources.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := validateReleaseMetadata(release, semantic); err != nil {
		t.Fatal(err)
	}
	var changed map[string]any
	if err := json.Unmarshal(semantic, &changed); err != nil {
		t.Fatal(err)
	}
	changed["source_lock_sha256"] = "another lock"
	mismatched, err := json.Marshal(changed)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateReleaseMetadata(release, mismatched); err == nil {
		t.Fatal("accepted semantic output from another source lock")
	}
}

// validateReleaseMetadata compares the manifests with the compiled catalog facts.
func validateReleaseMetadata(releaseJSON, semanticJSON []byte) error {
	var release struct {
		MinecraftVersion string          `json:"minecraft_version"`
		ProtocolVersion  int             `json:"protocol_version"`
		SourceLockSHA256 string          `json:"source_lock_sha256"`
		SourceLock       json.RawMessage `json:"source_lock"`
		Target           json.RawMessage `json:"target"`
	}
	if err := json.Unmarshal(releaseJSON, &release); err != nil {
		return err
	}
	if release.MinecraftVersion != data.MinecraftVersion || release.ProtocolVersion != data.ProtocolVersion {
		return fmt.Errorf("release target differs from the compiled catalog")
	}
	identity, err := json.Marshal(struct {
		Lock   json.RawMessage `json:"lock"`
		Target json.RawMessage `json:"target"`
	}{release.SourceLock, release.Target})
	if err != nil {
		return err
	}
	var canonical bytes.Buffer
	if err := json.Compact(&canonical, identity); err != nil {
		return err
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(canonical.Bytes()))
	for _, identity := range []string{release.SourceLockSHA256, data.SourceLockSHA256, data.SemanticSourceLockSHA256, registry.SourceLockSHA256, block.MiningSourceLockSHA256} {
		if digest != identity {
			return fmt.Errorf("catalog source-lock identities differ; regenerate both projections")
		}
	}
	var lock struct {
		Release   string `json:"release"`
		Upstreams map[string]struct {
			Revision string `json:"revision"`
		} `json:"upstreams"`
		Inputs map[string]struct {
			Upstream string `json:"upstream"`
		} `json:"inputs"`
		Semantic struct {
			Cloudburst string `json:"cloudburst"`
			BDSVersion string `json:"bds_version"`
		} `json:"semantic"`
	}
	if err := json.Unmarshal(release.SourceLock, &lock); err != nil {
		return err
	}
	var target struct {
		MinecraftVersion string `json:"minecraft_version"`
		ProtocolVersion  int    `json:"protocol_version"`
		Snapshot         string `json:"snapshot"`
		Channel          string `json:"channel"`
	}
	if err := json.Unmarshal(release.Target, &target); err != nil {
		return err
	}
	if target.MinecraftVersion != data.MinecraftVersion || target.ProtocolVersion != data.ProtocolVersion || target.Snapshot != data.ProtocolSnapshot || target.Channel != data.ReleaseChannel || lock.Release != target.Snapshot || lock.Upstreams[lock.Semantic.Cloudburst].Revision != data.CloudburstRef || lock.Semantic.BDSVersion != data.BDSVersion || lock.Inputs["liquid_clip_omissions"].Upstream != lock.Semantic.Cloudburst {
		return fmt.Errorf("compiled source pins differ from the release lock")
	}
	var semantic struct {
		SchemaVersion    int         `json:"schema_version"`
		CloudburstRef    string      `json:"cloudburst_ref"`
		BDSVersion       string      `json:"bds_version"`
		SourceLockSHA256 string      `json:"source_lock_sha256"`
		Counts           data.Counts `json:"counts"`
	}
	if err := json.Unmarshal(semanticJSON, &semantic); err != nil {
		return err
	}
	if semantic.SchemaVersion != data.SchemaVersion || semantic.SourceLockSHA256 != digest ||
		semantic.CloudburstRef != data.CloudburstRef || semantic.BDSVersion != data.BDSVersion ||
		semantic.Counts != data.GeneratedCounts {
		return fmt.Errorf("semantic manifest differs from the compiled catalog")
	}
	return nil
}
