// Command protocolgen is the v2 protocol pipeline. Sources are lowered into
// independent claims, reconciled into one version-pinned manifest, and only
// then handed to emitters or parity adapters.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"protocolgen/internal/changelog"
	"protocolgen/internal/claims"
	"protocolgen/internal/direction"
	"protocolgen/internal/emitgo"
	"protocolgen/internal/emitrust"
	"protocolgen/internal/emitter"
	"protocolgen/internal/gophertunneloracle"
	"protocolgen/internal/hotfix"
	"protocolgen/internal/ingest"
	"protocolgen/internal/manifest"
	"protocolgen/internal/nbtencoding"
	"protocolgen/internal/parity"
	"protocolgen/internal/reconcile"
	"protocolgen/internal/sourcelock"
	"protocolgen/internal/updateguide"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "reconcile":
		err = runReconcile(os.Args[2:])
	case "reconcile-claims":
		err = runReconcileClaims(os.Args[2:])
	case "supplement-symbols":
		err = runSupplementSymbols(os.Args[2:])
	case "carry-adjudications":
		err = runCarryAdjudications(os.Args[2:])
	case "adjudicate-claims":
		err = runAdjudicateClaims(os.Args[2:])
	case "ingest":
		err = runIngest(os.Args[2:])
	case "validate":
		err = runValidate(os.Args[2:])
	case "emit-go":
		err = runEmitGo(os.Args[2:])
	case "emit-rust":
		err = runEmitRust(os.Args[2:])
	case "parity":
		err = runParity(os.Args[2:])
	case "verify-gophertunnel":
		err = runVerifyGophertunnel(os.Args[2:])
	case "update-guide":
		err = runUpdateGuide(os.Args[2:])
	case "changelog":
		err = runChangelog(os.Args[2:])
	case "hash-source":
		err = runHashSource(os.Args[2:])
	case "hotfix":
		err = runHotfix(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `usage: protocolgen <reconcile|reconcile-claims|supplement-symbols|carry-adjudications|adjudicate-claims|ingest|validate|emit-go|emit-rust|parity|verify-gophertunnel|changelog|update-guide|hash-source|hotfix> [flags]

reconcile lowers one or both explicit source checkouts and writes manifest v2.
reconcile-claims reconciles two or more previously validated claims files and writes manifest v2.
supplement-symbols adds source-locked, non-wire actor metadata vocabularies to a canonical manifest.
carry-adjudications re-fingerprints prior adjudications only when the selected target claim remains byte-equivalent.
adjudicate-claims fingerprints explicit reviewed source selections against target claims.
ingest lowers one source to auditable machine-derived claims JSON.
validate checks a canonical manifest and all fingerprint metadata.
emit-go and emit-rust consume only a canonical manifest.
parity compares a canonical manifest with Axolotl's public v1 wire manifest.
verify-gophertunnel compares a canonical manifest with a pinned gophertunnel checkout and writes a JSON report.
changelog diffs two corrected Mojang schema directories into human-readable Markdown.
update-guide turns a protocol changelog and its target corrected schemas into gophertunnel transcription snippets.
hash-source prints the deterministic source-tree digest for a lock file.
hotfix derives a fingerprinted same-protocol target from a reconciled base manifest.`)
}

type adjudicationSelectionDocument struct {
	SchemaVersion uint32                  `json:"schema_version"`
	Target        manifest.Target         `json:"target"`
	Selections    []adjudicationSelection `json:"selections"`
}

type adjudicationSelection struct {
	ID             string              `json:"id"`
	Target         string              `json:"target"`
	SelectedSource string              `json:"selected_source"`
	Evidence       []manifest.Evidence `json:"evidence"`
	Reason         string              `json:"reason"`
}

// runAdjudicateClaims fingerprints explicit reviewed selections against all claims for each target field.
func runAdjudicateClaims(args []string) error {
	fs := flag.NewFlagSet("adjudicate-claims", flag.ContinueOnError)
	lockPath := fs.String("lock", "", "target source lock JSON")
	var claimPaths stringListFlag
	fs.Var(&claimPaths, "claims", "target claims JSON; repeat for each source")
	selectionsPath := fs.String("selections", "", "reviewed source selections JSON")
	existingPath := fs.String("existing", "", "existing carried adjudications JSON")
	outPath := fs.String("out", "", "complete adjudications JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *lockPath == "" || len(claimPaths) < 2 || *selectionsPath == "" || *outPath == "" {
		return fmt.Errorf("-lock, at least two -claims, -selections, and -out are required")
	}
	lock, err := sourcelock.Load(*lockPath)
	if err != nil {
		return err
	}
	groups := map[string][]claims.Claim{}
	for _, path := range claimPaths {
		result, err := loadClaims(path)
		if err != nil {
			return err
		}
		pin, ok := lock.Source(result.Pin.ID)
		if !ok || pin != result.Pin || result.Target != lock.Target {
			return fmt.Errorf("claims %s do not exactly match the target source lock", path)
		}
		for _, claim := range result.Claims {
			groups[claim.FieldPath] = append(groups[claim.FieldPath], claim)
		}
	}
	data, err := os.ReadFile(*selectionsPath)
	if err != nil {
		return fmt.Errorf("read selections: %w", err)
	}
	var document adjudicationSelectionDocument
	if err := json.Unmarshal(data, &document); err != nil {
		return fmt.Errorf("parse selections: %w", err)
	}
	if document.SchemaVersion != 1 || document.Target != lock.Target || len(document.Selections) == 0 {
		return fmt.Errorf("selections do not identify the target snapshot")
	}
	var output []manifest.Adjudication
	if *existingPath != "" {
		output, err = manifest.LoadAdjudications(*existingPath)
		if err != nil {
			return err
		}
	}
	seen := map[string]bool{}
	for _, adjudication := range output {
		seen[adjudication.Target] = true
	}
	for index, selection := range document.Selections {
		if selection.ID == "" || selection.Target == "" || selection.SelectedSource == "" || selection.Reason == "" || len(selection.Evidence) == 0 {
			return fmt.Errorf("selection %d is incomplete", index)
		}
		if seen[selection.Target] {
			return fmt.Errorf("selection target %q is already adjudicated", selection.Target)
		}
		group := groups[selection.Target]
		selected, ok := claimFromSource(group, selection.SelectedSource)
		if !ok {
			return fmt.Errorf("selection %q has no %q claim", selection.Target, selection.SelectedSource)
		}
		context, err := claims.ContextFingerprint(lock.Target, group)
		if err != nil {
			return err
		}
		fingerprints := make([]manifest.ClaimFingerprint, 0, len(group))
		for _, claim := range group {
			fingerprint, err := claims.Fingerprint(claim)
			if err != nil {
				return err
			}
			fingerprints = append(fingerprints, manifest.ClaimFingerprint{SourceID: claim.SourceID, Digest: fingerprint})
		}
		sort.Slice(fingerprints, func(i, j int) bool { return fingerprints[i].SourceID < fingerprints[j].SourceID })
		selectedFingerprint, err := claims.Fingerprint(selected)
		if err != nil {
			return err
		}
		evidence := append([]manifest.Evidence(nil), selection.Evidence...)
		evidence = append(evidence, manifest.Evidence{SourceID: selected.SourceID, Locator: selected.Locator, ClaimFingerprint: selectedFingerprint, Summary: "Exact target source claim selected by this reviewed adjudication."})
		output = append(output, manifest.Adjudication{ID: selection.ID, Target: selection.Target, PrePatchContextSHA256: context, Claims: fingerprints, SelectedSource: selection.SelectedSource, Evidence: evidence, Reason: selection.Reason})
		seen[selection.Target] = true
	}
	sort.Slice(output, func(i, j int) bool { return output[i].ID < output[j].ID })
	return writeJSON(*outPath, output)
}

type carryReportEntry struct {
	ID     string `json:"id"`
	Target string `json:"target"`
	Reason string `json:"reason"`
}

type carryReport struct {
	Carried int                `json:"carried"`
	Skipped []carryReportEntry `json:"skipped"`
}

// runCarryAdjudications safely carries prior decisions across a protocol bump when their selected wire shape is unchanged.
func runCarryAdjudications(args []string) error {
	fs := flag.NewFlagSet("carry-adjudications", flag.ContinueOnError)
	basePath := fs.String("base", "", "previous canonical manifest")
	lockPath := fs.String("lock", "", "target source lock JSON")
	var claimPaths stringListFlag
	fs.Var(&claimPaths, "claims", "target claims JSON; repeat for each source")
	outPath := fs.String("out", "", "carried adjudications JSON")
	reportPath := fs.String("report", "", "carry report JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *basePath == "" || *lockPath == "" || len(claimPaths) < 2 || *outPath == "" || *reportPath == "" {
		return fmt.Errorf("-base, -lock, at least two -claims, -out, and -report are required")
	}
	base, err := manifest.Load(*basePath)
	if err != nil {
		return err
	}
	lock, err := sourcelock.Load(*lockPath)
	if err != nil {
		return err
	}
	groups := map[string][]claims.Claim{}
	for _, path := range claimPaths {
		result, err := loadClaims(path)
		if err != nil {
			return err
		}
		pin, ok := lock.Source(result.Pin.ID)
		if !ok || pin != result.Pin || result.Target != lock.Target {
			return fmt.Errorf("claims %s do not exactly match the target source lock", path)
		}
		for _, claim := range result.Claims {
			groups[claim.FieldPath] = append(groups[claim.FieldPath], claim)
		}
	}
	baseFields := manifestFieldsByPath(base)
	carried := make([]manifest.Adjudication, 0, len(base.Adjudications))
	report := carryReport{}
	for _, previous := range base.Adjudications {
		field, ok := baseFields[previous.Target]
		if !ok {
			report.Skipped = append(report.Skipped, carryReportEntry{ID: previous.ID, Target: previous.Target, Reason: "previous canonical field is unavailable"})
			continue
		}
		group := groups[previous.Target]
		selected, ok := claimFromSource(group, previous.SelectedSource)
		if !ok {
			report.Skipped = append(report.Skipped, carryReportEntry{ID: previous.ID, Target: previous.Target, Reason: "selected source claim is unavailable"})
			continue
		}
		oldWire, err := claims.FieldWireFingerprint(field.PacketID, field.Direction, field.Field)
		if err != nil {
			return err
		}
		newWire, err := claims.WireFingerprint(selected)
		if err != nil {
			return err
		}
		if oldWire != newWire {
			report.Skipped = append(report.Skipped, carryReportEntry{ID: previous.ID, Target: previous.Target, Reason: "selected source wire shape changed"})
			continue
		}
		context, err := claims.ContextFingerprint(lock.Target, group)
		if err != nil {
			return err
		}
		fingerprints := make([]manifest.ClaimFingerprint, 0, len(group))
		for _, claim := range group {
			fingerprint, err := claims.Fingerprint(claim)
			if err != nil {
				return err
			}
			fingerprints = append(fingerprints, manifest.ClaimFingerprint{SourceID: claim.SourceID, Digest: fingerprint})
		}
		sort.Slice(fingerprints, func(i, j int) bool { return fingerprints[i].SourceID < fingerprints[j].SourceID })
		selectedFingerprint, err := claims.Fingerprint(selected)
		if err != nil {
			return err
		}
		evidence := append([]manifest.Evidence(nil), previous.Evidence...)
		evidence = append(evidence, manifest.Evidence{
			SourceID: selected.SourceID, Locator: selected.Locator, ClaimFingerprint: selectedFingerprint,
			Summary: "The exact target source claim is byte-equivalent to the previous canonical wire field.",
		})
		previous.ID = carriedAdjudicationID(previous.ID, base.Target.ProtocolVersion, lock.Target.ProtocolVersion)
		previous.PrePatchContextSHA256 = context
		previous.Claims = fingerprints
		previous.Evidence = evidence
		previous.Reason += " Carried forward only after byte-equivalence with the previous canonical field was verified."
		carried = append(carried, previous)
	}
	report.Carried = len(carried)
	if err := writeJSON(*outPath, carried); err != nil {
		return err
	}
	if err := writeJSON(*reportPath, report); err != nil {
		return err
	}
	fmt.Printf("carried %d adjudications; skipped %d -> %s\n", len(carried), len(report.Skipped), *outPath)
	return nil
}

type manifestFieldRef struct {
	PacketID  uint32
	Direction manifest.Direction
	Field     manifest.Field
}

// manifestFieldsByPath indexes top-level canonical fields using adjudication target paths.
func manifestFieldsByPath(value manifest.Manifest) map[string]manifestFieldRef {
	result := map[string]manifestFieldRef{}
	for _, packet := range value.Packets {
		for _, field := range packet.Fields {
			result[packet.Name+"."+field.Name] = manifestFieldRef{PacketID: packet.ID, Direction: packet.Direction, Field: field}
		}
	}
	return result
}

// claimFromSource selects one exact source claim from a disagreement group.
func claimFromSource(group []claims.Claim, sourceID string) (claims.Claim, bool) {
	for _, claim := range group {
		if claim.SourceID == sourceID {
			return claim, true
		}
	}
	return claims.Claim{}, false
}

// carriedAdjudicationID updates an established protocol-prefixed identifier without changing its descriptive suffix.
func carriedAdjudicationID(id string, fromProtocol, toProtocol int) string {
	prefix := fmt.Sprintf("protocol-%d-", fromProtocol)
	if strings.HasPrefix(id, prefix) {
		return fmt.Sprintf("protocol-%d-%s", toProtocol, strings.TrimPrefix(id, prefix))
	}
	return fmt.Sprintf("protocol-%d-carried-%s", toProtocol, id)
}

// writeJSON writes one deterministic indented JSON artifact.
func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && filepath.Dir(path) != "." {
		return err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

type stringListFlag []string

// String formats repeated string flags for flag package diagnostics.
func (values *stringListFlag) String() string {
	return fmt.Sprint([]string(*values))
}

// Set appends one repeated string flag value.
func (values *stringListFlag) Set(value string) error {
	*values = append(*values, value)
	return nil
}

// runReconcileClaims reconciles independently ingested, source-locked claims without assigning source precedence.
func runReconcileClaims(args []string) error {
	fs := flag.NewFlagSet("reconcile-claims", flag.ContinueOnError)
	lockPath := fs.String("lock", "", "source lock JSON")
	var claimPaths stringListFlag
	fs.Var(&claimPaths, "claims", "validated claims JSON; repeat for each independent source")
	adjudicationsPath := fs.String("adjudications", "", "fingerprinted adjudications JSON")
	directionsPath := fs.String("directions", "", "reviewed packet-direction JSON")
	nbtEncodingsPath := fs.String("nbt-encodings", "", "reviewed per-field NBT encoding JSON")
	symbolsLock := fs.String("symbols-lock", "", "override the persisted Mojang reference catalog lock")
	outPath := fs.String("out", "", "canonical manifest v2 output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *lockPath == "" || *directionsPath == "" || *nbtEncodingsPath == "" || *outPath == "" || len(claimPaths) < 2 {
		return fmt.Errorf("-lock, at least two -claims, -directions, -nbt-encodings, and -out are required")
	}
	lock, err := sourcelock.Load(*lockPath)
	if err != nil {
		return err
	}
	directions, err := direction.Load(*directionsPath)
	if err != nil {
		return err
	}
	nbtEncodings, err := nbtencoding.Load(*nbtEncodingsPath)
	if err != nil {
		return err
	}
	results := make([]claims.Result, 0, len(claimPaths))
	seen := map[string]bool{}
	for _, path := range claimPaths {
		result, err := loadClaims(path)
		if err != nil {
			return err
		}
		pin, ok := lock.Source(result.Pin.ID)
		if !ok || pin != result.Pin {
			return fmt.Errorf("claims %s source pin %q does not exactly match the source lock", path, result.Pin.ID)
		}
		if seen[result.Pin.ID] {
			return fmt.Errorf("claims contain duplicate source %q", result.Pin.ID)
		}
		seen[result.Pin.ID] = true
		results = append(results, result)
	}
	var adjudications []manifest.Adjudication
	if *adjudicationsPath != "" {
		adjudications, err = manifest.LoadAdjudications(*adjudicationsPath)
		if err != nil {
			return err
		}
	}
	result, err := reconcile.ReconcileWithDirectionsAndNBT(lock.Target, results, adjudications, directions, nbtEncodings)
	if err != nil {
		return err
	}
	if err := addConfiguredSymbols(&result, lock, *lockPath, *symbolsLock); err != nil {
		return err
	}
	if err := manifest.Write(*outPath, result); err != nil {
		return err
	}
	fmt.Printf("manifest v2: %d packets, %d source pins, %d adjudications -> %s\n", len(result.Packets), len(result.Sources), len(result.Adjudications), *outPath)
	return nil
}

// loadClaims reads one auditable claims file produced by the ingest command.
func loadClaims(path string) (claims.Result, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return claims.Result{}, fmt.Errorf("read claims %s: %w", path, err)
	}
	var result claims.Result
	if err := json.Unmarshal(data, &result); err != nil {
		return claims.Result{}, fmt.Errorf("parse claims %s: %w", path, err)
	}
	if result.Pin.ID == "" || result.Target.MinecraftVersion == "" || result.Target.ProtocolVersion == 0 {
		return claims.Result{}, fmt.Errorf("claims %s are incomplete", path)
	}
	return result, nil
}

func runHotfix(args []string) error {
	fs := flag.NewFlagSet("hotfix", flag.ContinueOnError)
	basePath := fs.String("base", "", "fully reconciled base manifest")
	specPath := fs.String("spec", "", "same-protocol hotfix specification")
	outPath := fs.String("out", "", "derived canonical manifest")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *basePath == "" || *specPath == "" || *outPath == "" {
		return fmt.Errorf("-base, -spec, and -out are required")
	}
	baseBytes, err := os.ReadFile(*basePath)
	if err != nil {
		return fmt.Errorf("read base manifest: %w", err)
	}
	base, err := manifest.Load(*basePath)
	if err != nil {
		return err
	}
	spec, err := hotfix.LoadSpec(*specPath)
	if err != nil {
		return err
	}
	result, err := hotfix.Apply(base, baseBytes, spec)
	if err != nil {
		return err
	}
	if err := manifest.Write(*outPath, result); err != nil {
		return err
	}
	fmt.Printf("hotfix manifest: Minecraft %s / protocol %d -> %s\n", result.Target.MinecraftVersion, result.Target.ProtocolVersion, *outPath)
	return nil
}

func runChangelog(args []string) error {
	fs := flag.NewFlagSet("changelog", flag.ContinueOnError)
	fromPath := fs.String("from", "", "previous corrected Mojang JSON schema directory")
	toPath := fs.String("to", "", "target corrected Mojang JSON schema directory")
	fromBranch := fs.String("from-branch", "", "previous upstream branch")
	toBranch := fs.String("to-branch", "", "target upstream branch")
	fromUpstream := fs.String("from-upstream", "", "previous upstream commit")
	toUpstream := fs.String("to-upstream", "", "target upstream commit")
	fromFixer := fs.String("from-fixer", "", "previous fixer commit")
	toFixer := fs.String("to-fixer", "", "target fixer commit")
	outPath := fs.String("out", "", "human-readable protocol changelog Markdown output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *fromPath == "" || *toPath == "" || *outPath == "" {
		return fmt.Errorf("-from, -to, and -out are required")
	}
	if *fromBranch == "" || *toBranch == "" || *fromUpstream == "" || *toUpstream == "" || *fromFixer == "" || *toFixer == "" {
		return fmt.Errorf("provenance flags -from-branch, -to-branch, -from-upstream, -to-upstream, -from-fixer, and -to-fixer are required")
	}
	data, err := changelog.Generate(changelog.Config{
		FromDir:      *fromPath,
		ToDir:        *toPath,
		FromBranch:   *fromBranch,
		ToBranch:     *toBranch,
		FromUpstream: *fromUpstream,
		ToUpstream:   *toUpstream,
		FromFixer:    *fromFixer,
		ToFixer:      *toFixer,
	})
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(*outPath), 0o755); err != nil {
		return fmt.Errorf("create changelog directory: %w", err)
	}
	if err := os.WriteFile(*outPath, data, 0o644); err != nil {
		return fmt.Errorf("write changelog: %w", err)
	}
	return nil
}

func runUpdateGuide(args []string) error {
	fs := flag.NewFlagSet("update-guide", flag.ContinueOnError)
	changelogPath := fs.String("changelog", "", "human-readable protocol changelog Markdown")
	schemasPath := fs.String("schemas", "", "target corrected Mojang JSON schema directory")
	outPath := fs.String("out", "", "gophertunnel update guide Markdown output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *changelogPath == "" || *schemasPath == "" || *outPath == "" {
		return fmt.Errorf("-changelog, -schemas, and -out are required")
	}
	changelog, err := os.ReadFile(*changelogPath)
	if err != nil {
		return fmt.Errorf("read changelog: %w", err)
	}
	guide, err := updateguide.Generate(changelog, *schemasPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(*outPath), 0o755); err != nil {
		return fmt.Errorf("create update guide directory: %w", err)
	}
	if err := os.WriteFile(*outPath, guide, 0o644); err != nil {
		return fmt.Errorf("write update guide: %w", err)
	}
	return nil
}

func runReconcile(args []string) error {
	fs := flag.NewFlagSet("reconcile", flag.ContinueOnError)
	lockPath := fs.String("lock", "", "source lock JSON")
	mojangRoot := fs.String("mojang", "", "local raw Mojang checkout or json directory")
	mojangID := fs.String("mojang-id", "mojang", "source-lock id for -mojang")
	mojangCorrections := fs.String("mojang-corrections", "", "fingerprinted correction directory for -mojang")
	endstoneRoot := fs.String("endstone", "", "local Endstone dump checkout")
	endstoneID := fs.String("endstone-id", "endstone", "source-lock id for -endstone")
	endstoneCorrections := fs.String("endstone-corrections", "", "fingerprinted correction directory for -endstone")
	adjudicationsPath := fs.String("adjudications", "", "fingerprinted adjudications JSON")
	directionsPath := fs.String("directions", "", "reviewed packet-direction JSON")
	nbtEncodingsPath := fs.String("nbt-encodings", "", "reviewed per-field NBT encoding JSON")
	symbolsLock := fs.String("symbols-lock", "", "override the persisted Mojang reference catalog lock")
	outPath := fs.String("out", "", "canonical manifest v2 output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *lockPath == "" || *directionsPath == "" || *nbtEncodingsPath == "" || *outPath == "" {
		return fmt.Errorf("-lock, -directions, -nbt-encodings, and -out are required")
	}
	lock, err := sourcelock.Load(*lockPath)
	if err != nil {
		return err
	}
	directions, err := direction.Load(*directionsPath)
	if err != nil {
		return err
	}
	nbtEncodings, err := nbtencoding.Load(*nbtEncodingsPath)
	if err != nil {
		return err
	}
	var results []claims.Result
	if *mojangRoot != "" {
		result, err := loadSource(lock, "mojang", *mojangID, *mojangRoot, *mojangCorrections)
		if err != nil {
			return err
		}
		results = append(results, result)
	}
	if *endstoneRoot != "" {
		result, err := loadSource(lock, "endstone", *endstoneID, *endstoneRoot, *endstoneCorrections)
		if err != nil {
			return err
		}
		results = append(results, result)
	}
	if len(results) == 0 {
		return fmt.Errorf("at least one of -mojang or -endstone is required")
	}
	var adjudications []manifest.Adjudication
	if *adjudicationsPath != "" {
		adjudications, err = manifest.LoadAdjudications(*adjudicationsPath)
		if err != nil {
			return err
		}
	}
	result, err := reconcile.ReconcileWithDirectionsAndNBT(lock.Target, results, adjudications, directions, nbtEncodings)
	if err != nil {
		return err
	}
	if err := addConfiguredSymbols(&result, lock, *lockPath, *symbolsLock); err != nil {
		return err
	}
	if err := manifest.Write(*outPath, result); err != nil {
		return err
	}
	fmt.Printf("manifest v2: %d packets, %d source pins, %d adjudications -> %s\n", len(result.Packets), len(result.Sources), len(result.Adjudications), *outPath)
	return nil
}

func runIngest(args []string) error {
	fs := flag.NewFlagSet("ingest", flag.ContinueOnError)
	kind := fs.String("kind", "", "mojang or endstone")
	root := fs.String("root", "", "source checkout")
	lockPath := fs.String("lock", "", "source lock JSON")
	id := fs.String("id", "", "source-lock id")
	corrections := fs.String("corrections", "", "fingerprinted correction directory")
	outPath := fs.String("out", "", "claims JSON output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *kind == "" || *root == "" || *lockPath == "" || *id == "" || *outPath == "" {
		return fmt.Errorf("-kind, -root, -lock, -id, and -out are required")
	}
	lock, err := sourcelock.Load(*lockPath)
	if err != nil {
		return err
	}
	result, err := loadSource(lock, *kind, *id, *root, *corrections)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(*outPath), 0o755); err != nil && filepath.Dir(*outPath) != "." {
		return err
	}
	if err := os.WriteFile(*outPath, data, 0o644); err != nil {
		return err
	}
	fmt.Printf("%s claims: %d -> %s\n", *kind, len(result.Claims), *outPath)
	return nil
}

func loadSource(lock sourcelock.Lock, kind, id, root, corrections string) (claims.Result, error) {
	pin, ok := lock.Source(id)
	if !ok {
		return claims.Result{}, fmt.Errorf("source-lock has no %q pin", id)
	}
	if pin.Kind != kind {
		return claims.Result{}, fmt.Errorf("source %q is pinned as kind %q, not %q", id, pin.Kind, kind)
	}
	if !sourcelock.IsSyntheticDigest(pin.Digest) {
		if err := sourcelock.VerifyDirectory(root, pin); err != nil {
			return claims.Result{}, err
		}
	}
	switch kind {
	case "mojang":
		return ingest.ParseMojang(root, pin, corrections)
	case "endstone":
		return ingest.ParseEndstone(root, pin, corrections)
	default:
		return claims.Result{}, fmt.Errorf("unknown source kind %q", kind)
	}
}

func runValidate(args []string) error {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	path := fs.String("manifest", "", "canonical manifest v2 JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *path == "" {
		return fmt.Errorf("-manifest is required")
	}
	m, err := manifest.Load(*path)
	if err != nil {
		return err
	}
	fmt.Printf("valid manifest v%d: %d packets, protocol %d\n", m.SchemaVersion, len(m.Packets), m.Target.ProtocolVersion)
	return nil
}

func runEmitGo(args []string) error {
	fs := flag.NewFlagSet("emit-go", flag.ContinueOnError)
	manifestPath := fs.String("manifest", "", "canonical manifest v2 JSON")
	out := fs.String("out", "", "generated Go source directory")
	namingPath := fs.String("naming", "", "reviewed naming overlay JSON; defaults to naming.json beside the manifest")
	domainsPath := fs.String("domains", "", "reviewed domain overlay JSON; defaults to domains.json beside the manifest")
	docsPath := fs.String("docs", "", "reviewed documentation overlay JSON; defaults to docs.json beside the manifest")
	protocolImport := fs.String("protocol-import", "", "import path of the generated protocol package")
	layoutPath := fs.String("layout", "", "reviewed layout overlay JSON that places constants and names fields like a gophertunnel checkout; never defaulted")
	semanticsPath := fs.String("semantics", "", "reviewed semantics overlay JSON marking actor identifier fields; defaults to semantics.json beside the manifest")
	nativeTypes := fs.Bool("native-types", true, "map canonical semantic shapes to established Go types such as uuid.UUID and mgl32 vectors")
	packetRuntime := fs.Bool("packet-runtime", true, "emit the packet interface and ID methods")
	packetPools := fs.Bool("packet-pools", true, "emit packet factory pools")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *manifestPath == "" || *out == "" {
		return fmt.Errorf("-manifest and -out are required")
	}
	if *protocolImport == "" {
		return fmt.Errorf("-protocol-import is required")
	}
	result, err := emitter.Run(emitter.Config{
		ManifestPath:  *manifestPath,
		NamingPath:    *namingPath,
		DomainsPath:   *domainsPath,
		DocsPath:      *docsPath,
		LayoutPath:    *layoutPath,
		SemanticsPath: *semanticsPath,
		OutputDir:     *out,
	}, emitter.Func(func(input emitter.Input) (map[string]string, error) {
		return emitgo.GenerateWithOptions(input.Manifest, emitgo.Options{
			ProtocolImportPath: *protocolImport,
			Naming:             input.Naming,
			Domains:            input.Domains,
			Docs:               input.Docs,
			Layout:             input.Layout,
			Semantics:          input.Semantics,
			NativeTypes:        *nativeTypes,
			EmitPacketRuntime:  *packetRuntime,
			EmitPacketPools:    *packetPools,
		})
	}))
	if err != nil {
		return err
	}
	fmt.Printf("Go emitter: %d files -> %s\n", result.FileCount, *out)
	fmt.Printf("Go docs coverage: types %d/%d, fields %d/%d\n", result.Coverage.TypesDocumented, result.Coverage.TypesTotal, result.Coverage.FieldsDocumented, result.Coverage.FieldsTotal)
	return nil
}

func runEmitRust(args []string) error {
	fs := flag.NewFlagSet("emit-rust", flag.ContinueOnError)
	manifestPath := fs.String("manifest", "", "canonical manifest v2 JSON")
	out := fs.String("out", "", "generated Rust source directory")
	namingPath := fs.String("naming", "", "reviewed naming overlay JSON; defaults to naming.json beside the manifest")
	domainsPath := fs.String("domains", "", "reviewed domain overlay JSON; defaults to domains.json beside the manifest")
	docsPath := fs.String("docs", "", "reviewed documentation overlay JSON; defaults to docs.json beside the manifest")
	codecContract := fs.Int("codec-contract", 0, "Rust codec contract to emit (docs/rust-codec-contract.md); 0 keeps the profile-neutral output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *manifestPath == "" || *out == "" {
		return fmt.Errorf("-manifest and -out are required")
	}
	result, err := emitter.Run(emitter.Config{
		ManifestPath: *manifestPath,
		NamingPath:   *namingPath,
		DomainsPath:  *domainsPath,
		DocsPath:     *docsPath,
		OutputDir:    *out,
	}, emitter.Func(func(input emitter.Input) (map[string]string, error) {
		return emitrust.GenerateFilesWithOptions(input.Manifest, emitrust.Options{Naming: input.Naming, Domains: input.Domains, Docs: input.Docs, CodecContract: *codecContract})
	}))
	if err != nil {
		return err
	}
	fmt.Printf("Rust emitter: %d files -> %s\n", result.FileCount, *out)
	fmt.Printf("Rust docs coverage: types %d/%d, fields %d/%d\n", result.Coverage.TypesDocumented, result.Coverage.TypesTotal, result.Coverage.FieldsDocumented, result.Coverage.FieldsTotal)
	return nil
}

func runParity(args []string) error {
	fs := flag.NewFlagSet("parity", flag.ContinueOnError)
	manifestPath := fs.String("manifest", "", "canonical manifest v2 JSON")
	axolotlPath := fs.String("axolotl", "", "Axolotl public v1 wire manifest JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *manifestPath == "" || *axolotlPath == "" {
		return fmt.Errorf("-manifest and -axolotl are required")
	}
	m, err := manifest.Load(*manifestPath)
	if err != nil {
		return err
	}
	if err := parity.CompareFile(m, *axolotlPath); err != nil {
		return err
	}
	fmt.Println("Axolotl v1 parity: byte-equivalence normalization matched")
	return nil
}

func runVerifyGophertunnel(args []string) error {
	fs := flag.NewFlagSet("verify-gophertunnel", flag.ContinueOnError)
	manifestPath := fs.String("manifest", "", "canonical manifest v2 JSON")
	lockPath := fs.String("lock", "tools/gophertunnel-oracle/lock.json", "gophertunnel oracle lock JSON")
	acceptedPath := fs.String("accepted", "tools/gophertunnel-oracle/accepted-divergences.json", "reviewed accepted-divergences JSON")
	gophertunnelPath := fs.String("gophertunnel", "", "existing gophertunnel checkout at the locked commit")
	cacheDir := fs.String("cache-dir", "", "cache directory for a runtime clone when -gophertunnel is omitted")
	reportPath := fs.String("report", "gophertunnel-report.json", "machine-readable report output JSON")
	failOnUnaccepted := fs.Bool("fail-on-unaccepted", true, "exit non-zero when a divergence is not in the accepted baseline")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *manifestPath == "" {
		return fmt.Errorf("-manifest is required")
	}
	report, err := gophertunneloracle.CompareFile(gophertunneloracle.Options{
		ManifestPath:     *manifestPath,
		LockPath:         *lockPath,
		AcceptedPath:     *acceptedPath,
		GophertunnelPath: *gophertunnelPath,
		CacheDir:         *cacheDir,
		ReportPath:       *reportPath,
		FailOnUnaccepted: *failOnUnaccepted,
	})
	if report.SchemaVersion != 0 {
		fmt.Print(gophertunneloracle.Summary(report, *reportPath))
	}
	return err
}

func runHashSource(args []string) error {
	fs := flag.NewFlagSet("hash-source", flag.ContinueOnError)
	root := fs.String("root", "", "source checkout")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *root == "" {
		return fmt.Errorf("-root is required")
	}
	digest, err := sourcelock.DigestDirectory(*root)
	if err != nil {
		return err
	}
	fmt.Println(digest)
	return nil
}
