package retail

import (
	"bytes"
	"testing"
)

// The committed synthetic corpus must be exactly what the generator writes, so
// both benches and the retail run read inputs a reviewer can regenerate.
func TestSyntheticCorpusIsCurrent(t *testing.T) {
	want, err := SyntheticCorpus()
	if err != nil {
		t.Fatal(err)
	}
	got, err := ReadSyntheticCorpus(SyntheticDir)
	if err != nil {
		t.Fatalf("%v; run go run ./cmd/retail synth", err)
	}
	if len(got) != len(want) {
		t.Fatalf("manifest lists %d files, the generator writes %d; run go run ./cmd/retail synth", len(got), len(want))
	}
	for i := range want {
		if got[i].Name != want[i].Name || got[i].File != want[i].File || !bytes.Equal(got[i].Batch, want[i].Batch) {
			t.Errorf("%s differs from the generator's output; run go run ./cmd/retail synth", want[i].File)
		}
	}
}
