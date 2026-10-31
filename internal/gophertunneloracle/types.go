// Package gophertunneloracle compares the canonical manifest with the
// hand-written gophertunnel packet marshals. It is deliberately source based:
// the checkout is parsed with go/ast and is never imported, compiled, or
// allowed to become a second protocol schema.
package gophertunneloracle

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	LockSchemaVersion     = 1
	ReportSchemaVersion   = 1
	AcceptedSchemaVersion = 1
	defaultCacheSubdir    = "protocolgen/gophertunnel"
)

// Lock pins the independent oracle. Commit must be a full 40-character SHA.
type Lock struct {
	SchemaVersion    int    `json:"schema_version"`
	MinecraftVersion string `json:"minecraft_version"`
	ProtocolVersion  int    `json:"protocol_version"`
	Gophertunnel     struct {
		Repo   string `json:"repo"`
		Commit string `json:"commit"`
	} `json:"gophertunnel"`
}

// AcceptedDivergence records a reviewed difference. Both the reason and the
// concrete evidence that would settle it are required so the baseline cannot
// silently turn an arbitrary mismatch into an accepted result.
type AcceptedDivergence struct {
	ID                uint32     `json:"id"`
	Fingerprint       string     `json:"fingerprint"`
	Name              string     `json:"name"`
	Reason            string     `json:"reason"`
	WhatWouldSettleIt string     `json:"what_would_settle_it"`
	Evidence          []Evidence `json:"evidence"`
}

type Evidence struct {
	Locator string `json:"locator"`
	Summary string `json:"summary"`
}

type AcceptedFile struct {
	SchemaVersion    int                  `json:"schema_version"`
	MinecraftVersion string               `json:"minecraft_version"`
	ProtocolVersion  int                  `json:"protocol_version"`
	Divergences      []AcceptedDivergence `json:"divergences"`
}

// Options controls one verification run.
type Options struct {
	ManifestPath     string
	LockPath         string
	AcceptedPath     string
	GophertunnelPath string
	CacheDir         string
	ReportPath       string
	FailOnUnaccepted bool
}

type Counts struct {
	Agreement      int `json:"AGREEMENT"`
	Divergence     int `json:"DIVERGENCE"`
	Unresolved     int `json:"UNRESOLVED"`
	NoOraclePacket int `json:"NO_ORACLE_PACKET"`
}

type Report struct {
	SchemaVersion         int              `json:"schema_version"`
	MinecraftVersion      string           `json:"minecraft_version"`
	ProtocolVersion       int              `json:"protocol_version"`
	Manifest              string           `json:"manifest"`
	Gophertunnel          OracleSource     `json:"gophertunnel"`
	Normalization         Normalization    `json:"normalization"`
	Counts                Counts           `json:"counts"`
	Accepted              []uint32         `json:"accepted_divergences,omitempty"`
	Unaccepted            []uint32         `json:"unaccepted_divergences,omitempty"`
	ResolvedAccepted      []uint32         `json:"resolved_accepted,omitempty"`
	OracleOnly            []PacketIdentity `json:"oracle_only_packets,omitempty"`
	UnresolvedDiagnostics []diagnostic     `json:"unresolved_diagnostics,omitempty"`
	Packets               []PacketResult   `json:"packets"`
}

type OracleSource struct {
	Repo   string `json:"repo"`
	Commit string `json:"commit"`
}

type Normalization struct {
	FixedWidth    string   `json:"fixed_width"`
	FixedGrouping string   `json:"fixed_array_grouping"`
	Strings       string   `json:"strings_and_bytes"`
	ByteArrays    string   `json:"byte_arrays"`
	UUID          string   `json:"uuid"`
	PreencodedNBT string   `json:"preencoded_nbt"`
	Colour        string   `json:"colour"`
	Preserved     []string `json:"preserved_distinctions"`
}

type PacketIdentity struct {
	ID   uint32 `json:"id"`
	Name string `json:"name"`
}

type PacketResult struct {
	ID                   uint32       `json:"id"`
	Name                 string       `json:"name"`
	GophertunnelName     string       `json:"gophertunnel_name,omitempty"`
	Classification       string       `json:"classification"`
	Fingerprint          string       `json:"fingerprint,omitempty"`
	OperationCount       int          `json:"operation_count,omitempty"`
	Reasons              []string     `json:"reasons,omitempty"`
	ManifestSequence     []string     `json:"manifest_sequence,omitempty"`
	GophertunnelSequence []string     `json:"gophertunnel_sequence,omitempty"`
	Differences          []Difference `json:"differences,omitempty"`
	Paths                []PathResult `json:"paths,omitempty"`
}

type PathResult struct {
	Classification         string       `json:"classification"`
	ManifestConstraint     string       `json:"manifest_constraint,omitempty"`
	GophertunnelConstraint string       `json:"gophertunnel_constraint,omitempty"`
	GophertunnelSite       string       `json:"gophertunnel_site,omitempty"`
	Reasons                []string     `json:"reasons,omitempty"`
	ManifestSequence       []string     `json:"manifest_sequence,omitempty"`
	GophertunnelSequence   []string     `json:"gophertunnel_sequence,omitempty"`
	Differences            []Difference `json:"differences,omitempty"`
}

type Difference struct {
	Position          int    `json:"position"`
	Manifest          string `json:"manifest,omitempty"`
	ManifestField     string `json:"manifest_field,omitempty"`
	Gophertunnel      string `json:"gophertunnel,omitempty"`
	GophertunnelField string `json:"gophertunnel_field,omitempty"`
}

type atom struct {
	Token   string
	Field   string
	Display string
	Site    string
}

// sourceOperation is the extractor's tree-shaped intermediate form. Keeping
// structure until comparison is what lets fixed arrays, option presence, and
// union discriminants remain visible.
type sourceOperation struct {
	Kind           string
	Field          string
	Code           string
	Prefix         string
	Encoding       string
	Presence       string
	Length         uint64
	Control        string
	CompareTo      string
	Predicate      string
	ConsumesPrefix bool
	Element        []sourceOperation
	Value          []sourceOperation
	Variants       []sourceVariant
	Default        []sourceOperation
	HasDefault     bool
	VariantValue   int64
	Reason         string
	Site           string
	TypeName       string
}

type sourceVariant struct {
	Value        int64
	Values       []int64
	Name         string
	Constraint   string
	Site         string
	Discriminant bool // Values were resolved from an integer comparison, not a bool guard.
	Negated      bool // The branch runs when the compared field is not in Values.
	Ops          []sourceOperation
}

type sourcePacket struct {
	ID         uint32
	Name       string
	Operations []sourceOperation
}

type diagnostic struct {
	Packet string `json:"packet,omitempty"`
	Type   string `json:"type,omitempty"`
	Field  string `json:"field,omitempty"`
	File   string `json:"file"`
	Line   int    `json:"line"`
	Reason string `json:"reason"`
	Raw    string `json:"raw,omitempty"`
}

type extraction struct {
	Packets     []sourcePacket
	Diagnostics []diagnostic
}

var comparisonSHA = regexp.MustCompile(`^sha256:[a-f0-9]{64}$`)

var fullSHA = regexp.MustCompile(`^[0-9a-fA-F]{40}$`)

func LoadLock(path string) (Lock, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Lock{}, fmt.Errorf("read gophertunnel lock: %w", err)
	}
	var lock Lock
	if err := json.Unmarshal(data, &lock); err != nil {
		return Lock{}, fmt.Errorf("parse gophertunnel lock: %w", err)
	}
	if lock.SchemaVersion != LockSchemaVersion {
		return Lock{}, fmt.Errorf("gophertunnel lock schema_version %d is not v%d", lock.SchemaVersion, LockSchemaVersion)
	}
	if lock.MinecraftVersion == "" || lock.ProtocolVersion <= 0 || lock.Gophertunnel.Repo == "" {
		return Lock{}, fmt.Errorf("gophertunnel lock is missing repository or protocol identity")
	}
	if !fullSHA.MatchString(lock.Gophertunnel.Commit) {
		return Lock{}, fmt.Errorf("gophertunnel lock commit must be a full 40-character SHA, got %q", lock.Gophertunnel.Commit)
	}
	lock.Gophertunnel.Commit = strings.ToLower(lock.Gophertunnel.Commit)
	return lock, nil
}

func LoadAccepted(path string) (AcceptedFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return AcceptedFile{}, fmt.Errorf("read accepted gophertunnel divergences: %w", err)
	}
	var accepted AcceptedFile
	if err := json.Unmarshal(data, &accepted); err != nil {
		return AcceptedFile{}, fmt.Errorf("parse accepted gophertunnel divergences: %w", err)
	}
	if accepted.SchemaVersion != AcceptedSchemaVersion {
		return AcceptedFile{}, fmt.Errorf("accepted-divergences schema_version %d is not v%d", accepted.SchemaVersion, AcceptedSchemaVersion)
	}
	if accepted.ProtocolVersion <= 0 || accepted.MinecraftVersion == "" {
		return AcceptedFile{}, fmt.Errorf("accepted-divergences is missing protocol identity")
	}
	seen := map[uint32]bool{}
	for _, entry := range accepted.Divergences {
		if entry.ID == 0 || entry.Name == "" || strings.TrimSpace(entry.Reason) == "" || strings.TrimSpace(entry.WhatWouldSettleIt) == "" || len(entry.Evidence) == 0 {
			return AcceptedFile{}, fmt.Errorf("accepted divergence %d is incomplete", entry.ID)
		}
		if !comparisonSHA.MatchString(entry.Fingerprint) {
			return AcceptedFile{}, fmt.Errorf("accepted divergence %d has no exact comparison fingerprint", entry.ID)
		}
		for _, evidence := range entry.Evidence {
			if strings.TrimSpace(evidence.Locator) == "" || strings.TrimSpace(evidence.Summary) == "" {
				return AcceptedFile{}, fmt.Errorf("accepted divergence %d has incomplete evidence", entry.ID)
			}
		}
		if seen[entry.ID] {
			return AcceptedFile{}, fmt.Errorf("duplicate accepted divergence packet id %d", entry.ID)
		}
		seen[entry.ID] = true
	}
	return accepted, nil
}

// checkAcceptedEvidence keeps the reviewed baseline citing the oracle that was
// actually run. An entry may cite any independent source, but a locator that
// names a gophertunnel checkout must be the locked repository at the locked
// commit; otherwise the "evidence" is a dead or moving link and the entry would
// silence a divergence nobody can re-read.
func checkAcceptedEvidence(lock Lock, accepted AcceptedFile) error {
	base := strings.TrimSuffix(lock.Gophertunnel.Repo, ".git")
	for _, entry := range accepted.Divergences {
		for _, evidence := range entry.Evidence {
			locator := strings.TrimSpace(evidence.Locator)
			if !strings.HasPrefix(locator, "http") || !strings.Contains(strings.ToLower(locator), "gophertunnel") {
				continue
			}
			if !strings.HasPrefix(locator, base+"/blob/"+lock.Gophertunnel.Commit+"/") && !strings.HasPrefix(locator, base+"/tree/"+lock.Gophertunnel.Commit) {
				return fmt.Errorf("accepted divergence %d cites gophertunnel evidence %q outside the locked oracle %s@%s", entry.ID, locator, base, lock.Gophertunnel.Commit)
			}
		}
	}
	return nil
}

func resolveCheckout(lock Lock, requested, cacheDir string) (string, error) {
	if requested != "" {
		path, err := filepath.Abs(requested)
		if err != nil {
			return "", fmt.Errorf("resolve gophertunnel checkout: %w", err)
		}
		if err := verifyCheckout(path, lock.Gophertunnel.Commit); err != nil {
			return "", err
		}
		return path, nil
	}
	if cacheDir == "" {
		base, err := os.UserCacheDir()
		if err != nil {
			return "", fmt.Errorf("find user cache directory: %w", err)
		}
		cacheDir = filepath.Join(base, defaultCacheSubdir)
	}
	target := filepath.Join(cacheDir, lock.Gophertunnel.Commit)
	if _, err := os.Stat(target); err == nil {
		if verificationErr := verifyCheckout(target, lock.Gophertunnel.Commit); verificationErr == nil {
			return target, nil
		} else {
			// Older failed clones left only Git metadata, with no source checkout to preserve.
			entries, err := os.ReadDir(target)
			if err != nil || len(entries) != 1 || entries[0].Name() != ".git" {
				return "", fmt.Errorf("cached gophertunnel checkout %s: %w", target, verificationErr)
			}
			if err := os.RemoveAll(target); err != nil {
				return "", fmt.Errorf("remove incomplete oracle cache: %w", err)
			}
		}
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("inspect gophertunnel cache: %w", err)
	}
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return "", fmt.Errorf("create gophertunnel cache: %w", err)
	}
	temporary, err := os.MkdirTemp(cacheDir, "."+lock.Gophertunnel.Commit+"-")
	if err != nil {
		return "", fmt.Errorf("create temporary oracle checkout: %w", err)
	}
	defer os.RemoveAll(temporary)
	if err := runGit("clone", "--no-checkout", "--filter=blob:none", lock.Gophertunnel.Repo, temporary); err != nil {
		return "", fmt.Errorf("clone gophertunnel %s: %w", lock.Gophertunnel.Repo, err)
	}
	if err := runGitAt(temporary, "fetch", "--depth=1", "origin", lock.Gophertunnel.Commit); err != nil {
		return "", fmt.Errorf("fetch pinned gophertunnel commit %s: %w", lock.Gophertunnel.Commit, err)
	}
	if err := runGitAt(temporary, "checkout", "--detach", lock.Gophertunnel.Commit); err != nil {
		return "", fmt.Errorf("checkout gophertunnel %s: %w", lock.Gophertunnel.Commit, err)
	}
	if err := verifyCheckout(temporary, lock.Gophertunnel.Commit); err != nil {
		return "", err
	}
	if err := os.Rename(temporary, target); err != nil {
		if verificationErr := verifyCheckout(target, lock.Gophertunnel.Commit); verificationErr != nil {
			return "", fmt.Errorf("publish oracle checkout: %w", err)
		}
	}
	return target, nil
}

func verifyCheckout(path, wanted string) error {
	info, err := os.Stat(filepath.Join(path, "minecraft", "protocol", "packet"))
	if err != nil || !info.IsDir() {
		return fmt.Errorf("gophertunnel checkout %q has no minecraft/protocol/packet directory", path)
	}
	got, err := gitOutput(path, "rev-parse", "HEAD")
	if err != nil {
		return fmt.Errorf("read gophertunnel checkout revision: %w", err)
	}
	got = strings.ToLower(strings.TrimSpace(got))
	if got != strings.ToLower(wanted) {
		return fmt.Errorf("gophertunnel checkout is %s, lock requires %s", got, wanted)
	}
	status, err := gitOutput(path, "status", "--porcelain", "--untracked-files=all", "--ignored", "--", "minecraft/protocol")
	if err != nil {
		return fmt.Errorf("inspect oracle source tree: %w", err)
	}
	if strings.TrimSpace(status) != "" {
		return fmt.Errorf("gophertunnel checkout has modified or untracked protocol sources")
	}
	return nil
}

func runGit(args ...string) error {
	_, err := gitOutput("", args...)
	return err
}

func runGitAt(directory string, args ...string) error {
	_, err := gitOutput(directory, args...)
	return err
}

func gitOutput(directory string, args ...string) (string, error) {
	// Kept in a separate file-level helper so tests can exercise lock/check-out
	// policy without making the extractor depend on the shell.
	return gitCommand(directory, args...)
}

func sortIDs(ids []uint32) {
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
}
