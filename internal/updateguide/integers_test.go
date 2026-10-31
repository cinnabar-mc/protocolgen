package updateguide

import (
	"strings"
	"testing"
)

func TestGeneratePreservesLargeEnumIntegers(t *testing.T) {
	directory := t.TempDir()
	writeFixture(t, directory, "Mode.json", `{"title":"Mode","x-protocol-version":2,"type":"string","enum":["A"],"x-underlying-type":"int64","x-enum-binary-value":[9007199254740993]}`)
	got, err := Generate([]byte("# Bedrock protocol changes — 1 to 2\n\n## Modified Enums\n\n### Mode\n"), directory)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "ModeA = 9007199254740993") {
		t.Fatalf("enum integer rounded: %s", got)
	}
}

func TestLoadDocumentsRejectsTrailingJSON(t *testing.T) {
	directory := t.TempDir()
	writeFixture(t, directory, "Mode.json", `{"title":"Mode","x-protocol-version":2} {}`)
	if _, err := loadDocuments(directory); err == nil {
		t.Fatal("trailing JSON accepted")
	}
}

func TestGenerateAcceptsIntegralDecimalMetadata(t *testing.T) {
	directory := t.TempDir()
	writeFixture(t, directory, "Mode.json", `{"title":"Mode","x-protocol-version":2.0,"enum":["A"],"x-enum-binary-value":[7e0]}`)
	got, err := Generate([]byte("# Bedrock protocol changes — 1 to 2\n\n## Modified Enums\n\n### Mode\n"), directory)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "ModeA = 7") {
		t.Fatalf("integral literal lost: %s", got)
	}
}
