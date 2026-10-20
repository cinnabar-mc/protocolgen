package retail

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	genpacket "protocolgen/generated/1.26.51/go/protocol/packet"
)

// A value built as generated-Go structs must encode exactly as the same
// value built as gophertunnel structs, and as the committed entry the Rust
// encode-from-struct test checks against.
func TestGeneratedGoEncodesFromStructLikeGophertunnel(t *testing.T) {
	if err := CheckEncodeDir(EncodeDir, EncodeCases()); err != nil {
		t.Fatalf("%v; run go run ./cmd/retail synth", err)
	}
	entries, err := EncodeCaseEntries()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range EncodeCases() {
		committed, err := os.ReadFile(filepath.Join(EncodeDir, c.Name+".bin"))
		if err != nil {
			t.Fatalf("%v; run go run ./cmd/retail synth", err)
		}
		if !bytes.Equal(committed, entries[c.Name]) {
			t.Errorf("%s: committed entry is stale; run go run ./cmd/retail synth", c.Name)
		}
		body, err := genpacket.Encode(c.Generated)
		if err != nil {
			t.Fatalf("%s: %v", c.Name, err)
		}
		if got := NewEntry(c.Generated.ID(), body).Payload; !bytes.Equal(got, entries[c.Name]) {
			t.Errorf("%s: generated Go encodes %x, gophertunnel %x", c.Name, got, entries[c.Name])
		}
	}
}
