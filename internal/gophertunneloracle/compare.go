package gophertunneloracle

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"protocolgen/internal/manifest"
)

func CompareFile(options Options) (Report, error) {
	canonical, err := manifest.Load(options.ManifestPath)
	if err != nil {
		return Report{}, err
	}
	lock, err := LoadLock(options.LockPath)
	if err != nil {
		return Report{}, err
	}
	accepted, err := LoadAccepted(options.AcceptedPath)
	if err != nil {
		return Report{}, err
	}
	if lock.ProtocolVersion != canonical.Target.ProtocolVersion || lock.MinecraftVersion != canonical.Target.MinecraftVersion {
		return Report{}, fmt.Errorf("gophertunnel lock targets Minecraft %s/protocol %d, manifest targets %s/%d", lock.MinecraftVersion, lock.ProtocolVersion, canonical.Target.MinecraftVersion, canonical.Target.ProtocolVersion)
	}
	if accepted.ProtocolVersion != canonical.Target.ProtocolVersion || accepted.MinecraftVersion != canonical.Target.MinecraftVersion {
		return Report{}, fmt.Errorf("accepted-divergences targets Minecraft %s/protocol %d, manifest targets %s/%d", accepted.MinecraftVersion, accepted.ProtocolVersion, canonical.Target.MinecraftVersion, canonical.Target.ProtocolVersion)
	}
	if err := checkAcceptedEvidence(lock, accepted); err != nil {
		return Report{}, err
	}
	checkout, err := resolveCheckout(lock, options.GophertunnelPath, options.CacheDir)
	if err != nil {
		return Report{}, err
	}
	extracted, err := ExtractAtRevision(checkout, lock.Gophertunnel.Commit)
	if err != nil {
		return Report{}, err
	}
	report := Compare(canonical, extracted, lock, accepted, options.ManifestPath)
	if options.ReportPath != "" {
		if err := WriteReport(options.ReportPath, report); err != nil {
			return report, err
		}
	}
	if options.FailOnUnaccepted && len(report.Unaccepted) > 0 {
		return report, fmt.Errorf("%d unaccepted gophertunnel wire divergence(s): %s", len(report.Unaccepted), joinIDs(report.Unaccepted))
	}
	if options.FailOnUnaccepted && len(report.ResolvedAccepted) > 0 {
		return report, fmt.Errorf("stale accepted gophertunnel divergence(s) require review: %s", joinIDs(report.ResolvedAccepted))
	}
	return report, nil
}

// Compare is the pure comparison layer. Extraction and checkout are kept
// outside it so normalization rules can be tested with tiny synthetic trees.
func Compare(canonical manifest.Manifest, source extraction, lock Lock, accepted AcceptedFile, manifestPath string) Report {
	report := Report{
		SchemaVersion:         ReportSchemaVersion,
		MinecraftVersion:      canonical.Target.MinecraftVersion,
		ProtocolVersion:       canonical.Target.ProtocolVersion,
		Manifest:              manifestPath,
		Gophertunnel:          OracleSource{Repo: lock.Gophertunnel.Repo, Commit: lock.Gophertunnel.Commit},
		Normalization:         defaultNormalization(),
		UnresolvedDiagnostics: source.Diagnostics,
	}
	acceptedByID := make(map[uint32]AcceptedDivergence, len(accepted.Divergences))
	for _, entry := range accepted.Divergences {
		acceptedByID[entry.ID] = entry
	}
	sourceByID := make(map[uint32]sourcePacket, len(source.Packets))
	for _, packet := range source.Packets {
		sourceByID[packet.ID] = packet
	}
	canonicalIDs := make(map[uint32]bool, len(canonical.Packets))
	divergentIDs := make(map[uint32]bool)
	for _, packet := range canonical.Packets {
		canonicalIDs[packet.ID] = true
		result := PacketResult{ID: packet.ID, Name: packet.Name}
		oracle, ok := sourceByID[packet.ID]
		if !ok {
			result.Classification = "NO_ORACLE_PACKET"
			report.Counts.NoOraclePacket++
			result.Reasons = []string{"gophertunnel has no packet with this ID"}
			report.Packets = append(report.Packets, result)
			continue
		}
		result.GophertunnelName = oracle.Name
		want := canonicalPacketExpr(packet)
		got := sourceSequenceExpr(oracle.Operations)
		if want.kind == "unknown" || got.kind == "unknown" {
			result.Classification = "UNRESOLVED"
			result.Reasons = uniqueStrings(append(append([]string{}, want.reasons...), got.reasons...))
			report.Counts.Unresolved++
			report.Packets = append(report.Packets, result)
			continue
		}
		witness := compareLanguages(want, got)
		if witness == nil {
			result.Classification = "AGREEMENT"
			if isLinear(want) {
				result.OperationCount = tokenCount(want)
			}
			report.Counts.Agreement++
			report.Packets = append(report.Packets, result)
			continue
		}
		result.Classification = "DIVERGENCE"
		divergentIDs[packet.ID] = true
		manifestSequence := append(append([]atom{}, witness.prefix...), witness.manifest...)
		sourceSequence := append(append([]atom{}, witness.prefix...), witness.gophertunnel...)
		result.ManifestSequence = atomDisplays(manifestSequence)
		result.GophertunnelSequence = atomDisplays(sourceSequence)
		result.Differences = differences(manifestSequence, sourceSequence)
		result.Paths = []PathResult{{
			Classification:       "DIVERGENCE",
			ManifestConstraint:   "shortest wire path on which the languages differ",
			GophertunnelSite:     firstAtomSite(witness.gophertunnel),
			ManifestSequence:     result.ManifestSequence,
			GophertunnelSequence: result.GophertunnelSequence,
			Differences:          result.Differences,
		}}
		report.Counts.Divergence++
		var err error
		result.Fingerprint, err = divergenceFingerprint(packet, oracle, result.Paths, lock)
		if err != nil {
			result.Reasons = append(result.Reasons, "cannot fingerprint divergence: "+err.Error())
		}
		if entry, accepted := acceptedByID[packet.ID]; accepted && result.Fingerprint != "" && entry.Fingerprint == result.Fingerprint {
			report.Accepted = append(report.Accepted, packet.ID)
		} else {
			if _, exists := acceptedByID[packet.ID]; exists {
				result.Reasons = append(result.Reasons, "reviewed comparison fingerprint changed; re-review both input shapes")
			}
			report.Unaccepted = append(report.Unaccepted, packet.ID)
		}
		report.Packets = append(report.Packets, result)
	}
	for _, packet := range source.Packets {
		if !canonicalIDs[packet.ID] {
			report.OracleOnly = append(report.OracleOnly, PacketIdentity{ID: packet.ID, Name: packet.Name})
		}
	}
	for _, entry := range accepted.Divergences {
		if !divergentIDs[entry.ID] {
			report.ResolvedAccepted = append(report.ResolvedAccepted, entry.ID)
		}
	}
	sort.Slice(report.Packets, func(i, j int) bool { return report.Packets[i].ID < report.Packets[j].ID })
	sort.Slice(report.OracleOnly, func(i, j int) bool { return report.OracleOnly[i].ID < report.OracleOnly[j].ID })
	sortIDs(report.Accepted)
	sortIDs(report.Unaccepted)
	sortIDs(report.ResolvedAccepted)
	return report
}

func defaultNormalization() Normalization {
	return Normalization{
		FixedWidth:    "Signed and unsigned fixed-width integers with identical width and endianness are equivalent.",
		FixedGrouping: "Fixed-array wrapper grouping is normalized by expanding scalar wire values; the repeated scalar count remains significant.",
		Strings:       "A length-prefixed UTF-8 string and a length-prefixed byte slice with the same prefix are wire-equivalent.",
		ByteArrays:    "A prefixed array of single u8 elements is equivalent to a byte slice with the same prefix.",
		UUID:          "UUID is compared as 16 bytes at its wire position; gophertunnel's internal UUID byte ordering is intentionally not validated.",
		PreencodedNBT: "RawBytes remain unresolved unless their encoding can be established from source.",
		Colour:        "A little-endian 32-bit colour int and gophertunnel's BEARGB (channels swapped, then big-endian) are the same four bytes.",
		Preserved: []string{
			"integer width",
			"endianness",
			"fixed-width versus varint",
			"varint versus zigzag",
			"float versus integer",
			"option presence",
			"array prefix type",
			"fixed-array scalar count",
			"union control and variant discriminants",
		},
	}
}

func manifestPrefix(node *manifest.Node) (string, error) {
	if node == nil || node.Kind != manifest.KindPrimitive || node.Primitive == nil {
		return "", fmt.Errorf("length prefix is not an explicit primitive")
	}
	return node.Primitive.Code, nil
}

func manifestPrimitive(node *manifest.Node) (string, error) {
	if node == nil || node.Kind != manifest.KindPrimitive || node.Primitive == nil {
		return "", fmt.Errorf("union control is not an explicit primitive")
	}
	return node.Primitive.Code, nil
}

func canonicalPrimitive(code string) string {
	switch code {
	case "i8", "u8":
		return "FIXED8"
	case "i16le", "u16le":
		return "FIXED16LE"
	case "i16be", "u16be":
		return "FIXED16BE"
	case "i32le", "u32le", "argb32":
		// BEARGB swaps the channels and writes big-endian, which is byte for
		// byte a little-endian ARGB int.
		return "FIXED32LE"
	case "i32be", "u32be":
		return "FIXED32BE"
	case "i64le", "u64le":
		return "FIXED64LE"
	case "i64be", "u64be":
		return "FIXED64BE"
	default:
		return code
	}
}

func isManifestU8(node manifest.Node) bool {
	return node.Kind == manifest.KindPrimitive && node.Primitive != nil && node.Primitive.Code == "u8"
}

func isSourceU8(operations []sourceOperation) bool {
	return len(operations) == 1 && operations[0].Kind == "primitive" && operations[0].Code == "u8"
}

func atomDisplays(atoms []atom) []string {
	result := make([]string, len(atoms))
	for index, current := range atoms {
		result[index] = current.Display
	}
	return result
}

func firstAtomSite(atoms []atom) string {
	for _, current := range atoms {
		if current.Site != "" {
			return current.Site
		}
	}
	return ""
}

func differences(left, right []atom) []Difference {
	n, m := len(left), len(right)
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}
	for i := n; i >= 0; i-- {
		for j := m; j >= 0; j-- {
			switch {
			case i == n:
				dp[i][j] = m - j
			case j == m:
				dp[i][j] = n - i
			case left[i].Token == right[j].Token:
				dp[i][j] = dp[i+1][j+1]
			default:
				dp[i][j] = minInt(1+dp[i+1][j+1], 1+dp[i+1][j], 1+dp[i][j+1])
			}
		}
	}
	var result []Difference
	i, j, position := 0, 0, 0
	for i < n || j < m {
		if i < n && j < m && left[i].Token == right[j].Token {
			i++
			j++
			position++
			continue
		}
		if i < n && j < m && dp[i][j] == 1+dp[i+1][j+1] {
			result = append(result, Difference{Position: position, Manifest: left[i].Display, ManifestField: left[i].Field, Gophertunnel: right[j].Display, GophertunnelField: right[j].Field})
			i++
			j++
			position++
			continue
		}
		if i < n && dp[i][j] == 1+dp[i+1][j] {
			result = append(result, Difference{Position: position, Manifest: left[i].Display, ManifestField: left[i].Field, Gophertunnel: "missing"})
			i++
			position++
			continue
		}
		result = append(result, Difference{Position: position, Manifest: "missing", Gophertunnel: right[j].Display, GophertunnelField: right[j].Field})
		j++
		position++
	}
	return result
}

func minInt(values ...int) int {
	result := values[0]
	for _, value := range values[1:] {
		if value < result {
			result = value
		}
	}
	return result
}

func uniqueStrings(input []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(input))
	for _, value := range input {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}

func containsID(ids []uint32, wanted uint32) bool {
	for _, id := range ids {
		if id == wanted {
			return true
		}
	}
	return false
}

func joinIDs(ids []uint32) string {
	parts := make([]string, len(ids))
	for index, id := range ids {
		parts[index] = fmt.Sprintf("%d", id)
	}
	return strings.Join(parts, ", ")
}

func WriteReport(path string, report Report) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal gophertunnel report: %w", err)
	}
	data = append(data, '\n')
	if parent := filepath.Dir(path); parent != "." && parent != "" {
		if err := os.MkdirAll(parent, 0o755); err != nil {
			return fmt.Errorf("create gophertunnel report directory: %w", err)
		}
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write gophertunnel report: %w", err)
	}
	return nil
}

func Summary(report Report, reportPath string) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "gophertunnel oracle: %s @ %s\n", report.Gophertunnel.Repo, report.Gophertunnel.Commit)
	fmt.Fprintf(&builder, "Compared %d manifest packets for Minecraft %s / protocol %d\n", len(report.Packets), report.MinecraftVersion, report.ProtocolVersion)
	fmt.Fprintf(&builder, "AGREEMENT: %d  DIVERGENCE: %d  UNRESOLVED: %d  NO_ORACLE_PACKET: %d\n", report.Counts.Agreement, report.Counts.Divergence, report.Counts.Unresolved, report.Counts.NoOraclePacket)
	fmt.Fprintf(&builder, "accepted divergences: %d  unaccepted divergences: %d  resolved accepted entries: %d\n", len(report.Accepted), len(report.Unaccepted), len(report.ResolvedAccepted))
	if len(report.Unaccepted) > 0 {
		fmt.Fprintf(&builder, "unaccepted packet IDs: %s\n", joinIDs(report.Unaccepted))
	}
	if len(report.ResolvedAccepted) > 0 {
		fmt.Fprintf(&builder, "review baseline entries no longer diverging: %s\n", joinIDs(report.ResolvedAccepted))
	}
	if len(report.OracleOnly) > 0 {
		fmt.Fprintf(&builder, "oracle-only packets: %d\n", len(report.OracleOnly))
	}
	if reportPath != "" {
		fmt.Fprintf(&builder, "JSON report: %s\n", reportPath)
	}
	return builder.String()
}
