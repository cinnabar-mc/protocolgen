// Package source authenticates the inputs shared by Bedrock data generators.
package source

import (
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//go:embed lock.json releases.json
var manifests embed.FS

//go:embed inputs
var bundled embed.FS

// Upstream pins a public GitHub repository once for all of its input files.
type Upstream struct {
	Repository string `json:"repository"`
	Revision   string `json:"revision"`
}

// Input identifies a local file or a file/archive from a locked upstream.
// Revision is only used for local inputs without an upstream relationship.
type Input struct {
	Path        string `json:"path,omitempty"`
	Upstream    string `json:"upstream,omitempty"`
	File        string `json:"file,omitempty"`
	Archive     bool   `json:"archive,omitempty"`
	SHA256      string `json:"sha256"`
	Revision    string `json:"revision,omitempty"`
	Description string `json:"description,omitempty"`
}

// Lock records input identities and selects one release from releases.json.
// Target is resolved by Open and included in the complete lock digest.
type Lock struct {
	SchemaVersion int                 `json:"schema_version"`
	Release       string              `json:"release"`
	Upstreams     map[string]Upstream `json:"upstreams"`
	Inputs        map[string]Input    `json:"inputs"`
	Semantic      SemanticInputs      `json:"semantic"`
	Target        Release             `json:"-"`
}

// SHA256 identifies the complete input lock and its resolved release record.
func (l Lock) SHA256() (string, error) {
	canonical, err := json.Marshal(struct {
		Lock   Lock    `json:"lock"`
		Target Release `json:"target"`
	}{l, l.Target})
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(canonical)
	return hex.EncodeToString(digest[:]), nil
}

// CloudburstRevision returns the single revision shared by the registry and facts.
func (l Lock) CloudburstRevision() string {
	return l.Upstreams[l.Semantic.Cloudburst].Revision
}

// Sources resolves files relative to a lock and verifies every input digest.
type Sources struct {
	Lock Lock
	dir  string
}

// Open reads an explicit lock, or the bundled release lock when path is empty.
func Open(path string) (*Sources, error) {
	var data []byte
	var err error
	s := &Sources{}
	if path == "" {
		data, err = manifests.ReadFile("lock.json")
	} else {
		data, err = os.ReadFile(path)
		s.dir = filepath.Dir(path)
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &s.Lock); err != nil {
		return nil, fmt.Errorf("decode source lock: %w", err)
	}
	if s.Lock.SchemaVersion != 2 || len(s.Lock.Inputs) == 0 {
		return nil, fmt.Errorf("source lock has an unsupported schema or no inputs")
	}
	releases, err := Releases()
	if err != nil {
		return nil, err
	}
	target, ok := releases.Releases[s.Lock.Release]
	if !ok {
		return nil, fmt.Errorf("unknown source release %q", s.Lock.Release)
	}
	s.Lock.Target = target
	for name, upstream := range s.Lock.Upstreams {
		revision, err := hex.DecodeString(upstream.Revision)
		parts := strings.Split(upstream.Repository, "/")
		if name == "" || len(parts) != 2 || parts[0] == "" || parts[1] == "" || strings.ContainsAny(upstream.Repository, "\\ ?#%:") || strings.Contains(upstream.Repository, "..") || err != nil || len(revision) != 20 {
			return nil, fmt.Errorf("upstream %q has an invalid repository or commit", name)
		}
	}
	for name, input := range s.Lock.Inputs {
		digest, err := hex.DecodeString(input.SHA256)
		if name == "" || err != nil || len(digest) != sha256.Size {
			return nil, fmt.Errorf("source %q has an invalid digest", name)
		}
		if input.Upstream != "" {
			if _, ok := s.Lock.Upstreams[input.Upstream]; !ok || input.Revision != "" {
				return nil, fmt.Errorf("source %q has an unknown upstream or duplicate revision", name)
			}
		} else if input.Path == "" || input.Revision == "" {
			return nil, fmt.Errorf("source %q needs an upstream or local revision", name)
		}
		locations := 0
		for _, path := range []string{input.Path, input.File} {
			if path != "" {
				locations++
				if !filepath.IsLocal(path) || strings.ContainsAny(path, "\\?#%") {
					return nil, fmt.Errorf("source %q has an invalid path", name)
				}
			}
		}
		if input.Archive {
			locations++
		}
		if locations != 1 {
			return nil, fmt.Errorf("source %q needs exactly one local path, upstream file or archive", name)
		}
	}
	if s.Lock.CloudburstRevision() == "" {
		return nil, fmt.Errorf("source lock has no Cloudburst upstream")
	}
	for _, name := range []string{"item_registry", "data_driven_blocks", "item_components", "liquid_clip_omissions"} {
		if input, ok := s.Lock.Inputs[name]; ok && input.Upstream != s.Lock.Semantic.Cloudburst {
			return nil, fmt.Errorf("source %q must use the semantic Cloudburst upstream", name)
		}
	}
	return s, nil
}

// Revision returns the pinned revision for an input, or empty for an unknown name.
func (s *Sources) Revision(name string) string {
	input := s.Lock.Inputs[name]
	if input.Upstream != "" {
		return s.Lock.Upstreams[input.Upstream].Revision
	}
	return input.Revision
}

// URL derives an immutable download URL from the upstream commit and file path.
// Local files and unknown names have no download URL.
func (s *Sources) URL(name string) string {
	input := s.Lock.Inputs[name]
	if input.Path != "" || input.Upstream == "" {
		return ""
	}
	upstream := s.Lock.Upstreams[input.Upstream]
	if input.Archive {
		return "https://codeload.github.com/" + upstream.Repository + "/tar.gz/" + upstream.Revision
	}
	return "https://raw.githubusercontent.com/" + upstream.Repository + "/" + upstream.Revision + "/" + input.File
}

// Read loads one locked source, including bundled generator inputs. Runtime
// consumers that only need downloads should use ReadRemote to avoid embedding inputs.
func (s *Sources) Read(ctx context.Context, name, cache string) ([]byte, error) {
	input, ok := s.Lock.Inputs[name]
	if !ok {
		return nil, fmt.Errorf("source %q is not in the lock", name)
	}
	if input.Path == "" {
		return s.ReadRemote(ctx, name, cache)
	}
	var data []byte
	var err error
	if s.dir == "" {
		data, err = bundled.ReadFile(input.Path)
	} else {
		data, err = os.ReadFile(filepath.Join(s.dir, input.Path))
	}
	return authenticate(name, input, data, err)
}

// ReadRemote loads only a locked upstream input, rejecting local inputs. Downloaded
// and cached bytes are authenticated without linking bundled generator payloads.
func (s *Sources) ReadRemote(ctx context.Context, name, cache string) ([]byte, error) {
	input, ok := s.Lock.Inputs[name]
	if !ok {
		return nil, fmt.Errorf("source %q is not in the lock", name)
	}
	if input.Path != "" {
		return nil, fmt.Errorf("source %q is local; use Read for generator inputs", name)
	}
	var data []byte
	var err error
	cachePath := ""
	if cache != "" {
		cachePath = filepath.Join(cache, input.SHA256)
		data, err = os.ReadFile(cachePath)
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("read source cache: %w", err)
		}
	}
	if data == nil {
		request, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, s.URL(name), nil)
		if reqErr != nil {
			return nil, reqErr
		}
		client := &http.Client{Timeout: 2 * time.Minute}
		response, reqErr := client.Do(request)
		if reqErr != nil {
			return nil, reqErr
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("fetch source %q: %s", name, response.Status)
		}
		const limit = 256 << 20
		data, err = io.ReadAll(io.LimitReader(response.Body, limit+1))
		if len(data) > limit {
			return nil, fmt.Errorf("source %q exceeds %d bytes", name, limit)
		}
		if err == nil && fmt.Sprintf("%x", sha256.Sum256(data)) == input.SHA256 && cachePath != "" {
			if err := os.MkdirAll(cache, 0o755); err != nil {
				return nil, err
			}
			if err := writeCache(cachePath, data); err != nil {
				return nil, err
			}
		}
	}
	return authenticate(name, input, data, err)
}

// authenticate checks read errors and requires the exact digest declared by the lock.
func authenticate(name string, input Input, data []byte, err error) ([]byte, error) {
	if err != nil {
		return nil, fmt.Errorf("read source %q: %w", name, err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != input.SHA256 {
		return nil, fmt.Errorf("source %q digest is %s, want %s", name, got, input.SHA256)
	}
	return data, nil
}

// writeCache installs a verified input atomically so concurrent generators never
// observe a partially written cache entry.
func writeCache(path string, data []byte) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".source-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err := file.Write(data); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}

// ValidateNative checks the authenticated native snapshot against the selected release and protocol.
func (s *Sources) ValidateNative(ctx context.Context, cache string) error {
	data, err := s.Read(ctx, "native_version", cache)
	if err != nil {
		return err
	}
	var version struct {
		Protocol int    `json:"protocol"`
		Version  string `json:"version"`
	}
	if err := json.Unmarshal(data, &version); err != nil {
		return fmt.Errorf("decode native_version: %w", err)
	}
	if version.Protocol != s.Lock.Target.ProtocolVersion || version.Version != s.Lock.Release {
		return fmt.Errorf("native snapshot %s/%d does not match release %s/%d", version.Version, version.Protocol, s.Lock.Release, s.Lock.Target.ProtocolVersion)
	}
	return nil
}
