package retail

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func committedCorpus(t *testing.T) Corpus {
	t.Helper()
	corpus, err := LoadCorpus("corpus.json")
	if err != nil {
		t.Fatal(err)
	}
	return corpus
}

func TestCommittedExceptionsAreReviewable(t *testing.T) {
	oracle, err := GophertunnelReplacement("go.mod")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := LoadExceptions("exceptions.json", committedCorpus(t), oracle); err != nil {
		t.Fatal(err)
	}
}

// The corpus directions decide which fixtures get a wrong-sender case, so
// they must be the manifest's.
func TestCorpusDirectionsMatchTheManifest(t *testing.T) {
	corpus := committedCorpus(t)
	data, err := os.ReadFile(filepath.Join("..", "..", "generated", corpus.MinecraftVersion, "directions.json"))
	if err != nil {
		t.Fatal(err)
	}
	var directions struct {
		Packets []struct {
			ID        uint32 `json:"id"`
			Direction string `json:"direction"`
		} `json:"packets"`
	}
	if err := json.Unmarshal(data, &directions); err != nil {
		t.Fatal(err)
	}
	manifest := map[uint32]string{}
	for _, packet := range directions.Packets {
		manifest[packet.ID] = map[string]string{"server": "clientbound", "client": "serverbound", "both": "bidirectional"}[packet.Direction]
	}
	for _, fixture := range corpus.Fixtures {
		if got := manifest[fixture.ID]; got != fixture.Direction {
			t.Errorf("%s: corpus says %s, manifest says %q", fixture.File, fixture.Direction, got)
		}
		if fixture.Direction == "serverbound" && fixture.Sender != SenderClient || fixture.Direction == "clientbound" && fixture.Sender != SenderServer {
			t.Errorf("%s: %s packet sent by the %s", fixture.File, fixture.Direction, fixture.Sender)
		}
	}
}

// The Rust runner names the contract crate, which is versioned, so it must
// follow the corpus version.
func TestRunnerTargetsTheCorpusVersion(t *testing.T) {
	corpus := committedCorpus(t)
	manifest, err := os.ReadFile(filepath.Join("runner", "Cargo.toml"))
	if err != nil {
		t.Fatal(err)
	}
	dashed := "bedrock-protocol-" + strings.ReplaceAll(corpus.MinecraftVersion, ".", "-")
	_, dependency, found := strings.Cut(string(manifest), dashed+" = {")
	if !found {
		t.Fatalf("runner/Cargo.toml does not depend on %s", dashed)
	}
	dependency, _, found = strings.Cut(dependency, "}")
	if !found {
		t.Fatalf("runner/Cargo.toml has an unterminated dependency for %s", dashed)
	}
	contract := "bedrock-protocol-contract-" + strings.ReplaceAll(corpus.MinecraftVersion, ".", "-")
	for _, field := range []string{
		`package = "` + contract + `"`,
		`path = "../../../generated/` + corpus.MinecraftVersion + `/rust-contract"`,
	} {
		if !strings.Contains(dependency, field) {
			t.Fatalf("runner/Cargo.toml dependency %s is missing %s", dashed, field)
		}
	}
	source, err := os.ReadFile(filepath.Join("runner", "src", "contract.rs"))
	if err != nil {
		t.Fatal(err)
	}
	if crate := strings.ReplaceAll(dashed, "-", "_"); !strings.Contains(string(source), "use "+crate+"::") {
		t.Fatalf("runner/src/contract.rs does not use %s", crate)
	}
}

func TestGophertunnelReplacementIsRead(t *testing.T) {
	path := filepath.Join(t.TempDir(), "go.mod")
	mod := "module fixturegen\n\nreplace github.com/sandertv/gophertunnel => github.com/fork/gophertunnel v1.2.3\n"
	if err := os.WriteFile(path, []byte(mod), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := GophertunnelReplacement(path)
	if err != nil || got != "github.com/fork/gophertunnel v1.2.3" {
		t.Fatalf("got %q, %v", got, err)
	}
}
