package ingest

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestCorrectionDigestsDistinguishLargeIntegers(t *testing.T) {
	left, err := canonicalDigest(json.Number("18446744073709551614"))
	if err != nil {
		t.Fatal(err)
	}
	right, err := canonicalDigest(json.Number("18446744073709551615"))
	if err != nil {
		t.Fatal(err)
	}
	if left == right {
		t.Fatal("adjacent uint64 integers have the same correction digest")
	}
	equivalent, err := canonicalDigest(json.Number("184467440737095516150e-1"))
	if err != nil {
		t.Fatal(err)
	}
	if right != equivalent {
		t.Fatal("equivalent integral spellings have different digests")
	}
}

func TestCorrectionRejectsChangedLargeInteger(t *testing.T) {
	old := json.Number("18446744073709551614")
	digest, err := canonicalDigest(old)
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	writeJSON(t, filepath.Join(directory, "change.json"), map[string]any{"operations": []any{map[string]any{"id": "fix", "file": "Value.json", "pointer": "/maximum", "pre_patch_sha256": digest, "why": "correct bound", "replace": 0}}})
	_, err = applyCorrections(map[string]any{"Value.json": map[string]any{"maximum": json.Number("18446744073709551615")}}, directory, fixturePin("mojang"))
	if err == nil || !strings.Contains(err.Error(), "stale correction") {
		t.Fatalf("changed integer accepted: %v", err)
	}
}
