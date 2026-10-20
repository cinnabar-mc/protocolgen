package retail

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
)

// Senders, as the corpus records them.
const (
	SenderServer = "server"
	SenderClient = "client"
)

const gophertunnelModule = "github.com/sandertv/gophertunnel"

// Corpus is differential/retail/corpus.json.
type Corpus struct {
	SchemaVersion    int       `json:"schema_version"`
	MinecraftVersion string    `json:"minecraft_version"`
	ProtocolVersion  int       `json:"protocol_version"`
	Source           Source    `json:"source"`
	Valentine        Pin       `json:"valentine"`
	Format           string    `json:"format"`
	Fixtures         []Fixture `json:"fixtures"`
}

// Source pins the Cinnabar commit and directory holding the fixtures.
type Source struct {
	Repository string `json:"repository"`
	Commit     string `json:"commit"`
	Directory  string `json:"directory"`
}

// Pin is a directory at another commit of the source repository.
type Pin struct {
	Commit    string `json:"commit"`
	Directory string `json:"directory"`
}

// Fixture is one hash-pinned batch. File names are unique and key every
// verdict.
type Fixture struct {
	Name      string `json:"name"`
	File      string `json:"file"`
	Packet    string `json:"packet"`
	ID        uint32 `json:"id"`
	Direction string `json:"direction"`
	Sender    string `json:"sender"`
	SHA256    string `json:"sha256"`
}

// LoadCorpus reads and checks a corpus file.
func LoadCorpus(path string) (Corpus, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Corpus{}, err
	}
	var corpus Corpus
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&corpus); err != nil {
		return Corpus{}, fmt.Errorf("%s: %w", path, err)
	}
	if corpus.SchemaVersion != 1 || corpus.Source.Commit == "" || corpus.Source.Directory == "" || corpus.Valentine.Commit == "" || corpus.Valentine.Directory == "" || len(corpus.Fixtures) == 0 {
		return Corpus{}, fmt.Errorf("%s: incomplete corpus", path)
	}
	files := map[string]bool{}
	for _, fixture := range corpus.Fixtures {
		if fixture.File == "" || filepath.Base(fixture.File) != fixture.File || files[fixture.File] {
			return Corpus{}, fmt.Errorf("%s: fixture file %q is empty, nested or repeated", path, fixture.File)
		}
		files[fixture.File] = true
		if fixture.Sender != SenderServer && fixture.Sender != SenderClient {
			return Corpus{}, fmt.Errorf("%s: fixture %s has sender %q", path, fixture.File, fixture.Sender)
		}
		if _, err := hex.DecodeString(fixture.SHA256); err != nil || len(fixture.SHA256) != sha256.Size*2 {
			return Corpus{}, fmt.Errorf("%s: fixture %s has an invalid SHA-256", path, fixture.File)
		}
	}
	return corpus, nil
}

// ReadFixtures reads every fixture from a Cinnabar checkout at the pinned
// commit, rejecting any whose SHA-256 differs from the corpus.
func (c Corpus) ReadFixtures(cinnabar string) (map[string][]byte, error) {
	head, err := exec.Command("git", "-C", cinnabar, "rev-parse", "HEAD").Output()
	if err != nil {
		return nil, fmt.Errorf("read Cinnabar checkout commit: %w", err)
	}
	if got := strings.TrimSpace(string(head)); got != c.Source.Commit {
		return nil, fmt.Errorf("checkout %s is at %s, corpus pins %s", cinnabar, got, c.Source.Commit)
	}
	batches := make(map[string][]byte, len(c.Fixtures))
	for _, fixture := range c.Fixtures {
		data, err := os.ReadFile(filepath.Join(cinnabar, c.Source.Directory, fixture.File))
		if err != nil {
			return nil, err
		}
		if sum := sha256.Sum256(data); hex.EncodeToString(sum[:]) != fixture.SHA256 {
			return nil, fmt.Errorf("fixture %s does not match its pinned SHA-256", fixture.File)
		}
		batches[fixture.File] = data
	}
	return batches, nil
}

// CheckOraclePin fails unless this binary links the gophertunnel commit that
// encoded the fixtures, as pinned by fixturegen's go.mod in the checkout.
func CheckOraclePin(cinnabar string) error {
	want, err := GophertunnelReplacement(filepath.Join(cinnabar, "tools", "fixturegen", "go.mod"))
	if err != nil {
		return err
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return errors.New("binary has no module build information")
	}
	for _, dep := range info.Deps {
		if dep.Path != gophertunnelModule {
			continue
		}
		got := dep.Path + " " + dep.Version
		if dep.Replace != nil {
			got = dep.Replace.Path + " " + dep.Replace.Version
		}
		if got != want {
			return fmt.Errorf("oracle is %s, fixturegen pins %s", got, want)
		}
		return nil
	}
	return errors.New("binary does not link gophertunnel")
}

// GophertunnelReplacement returns "path version" of the gophertunnel
// replacement in a go.mod file.
func GophertunnelReplacement(goMod string) (string, error) {
	file, err := os.Open(goMod)
	if err != nil {
		return "", err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(strings.TrimPrefix(strings.TrimSpace(scanner.Text()), "replace"))
		if len(fields) == 4 && fields[0] == gophertunnelModule && fields[1] == "=>" {
			return fields[2] + " " + fields[3], nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("%s does not replace %s", goMod, gophertunnelModule)
}
