package manifest

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"reflect"
	"regexp"
	"sort"
	"strings"
)

func Validate(m Manifest) error {
	if m.SchemaVersion != SchemaVersion {
		return fmt.Errorf("schema_version %d is not v%d", m.SchemaVersion, SchemaVersion)
	}
	if m.Target.MinecraftVersion == "" || m.Target.ProtocolVersion <= 0 {
		return fmt.Errorf("target must pin minecraft_version and positive protocol_version")
	}
	if len(m.Sources) == 0 {
		return fmt.Errorf("manifest has no source pins")
	}
	if len(m.Packets) == 0 {
		return fmt.Errorf("manifest has no packets")
	}
	sourceIDs := make(map[string]bool, len(m.Sources))
	for i, source := range m.Sources {
		if source.ID == "" || source.Kind == "" || source.Revision == "" || source.Digest == "" {
			return fmt.Errorf("sources[%d] is missing an immutable pin", i)
		}
		if sourceIDs[source.ID] {
			return fmt.Errorf("duplicate source pin %q", source.ID)
		}
		sourceIDs[source.ID] = true
		if source.ProtocolVersion != 0 && source.ProtocolVersion != m.Target.ProtocolVersion {
			return fmt.Errorf("mixes protocol %d source %q into target protocol %d", source.ProtocolVersion, source.ID, m.Target.ProtocolVersion)
		}
		baseVersion := ""
		if m.Derivation != nil {
			baseVersion = m.Derivation.BaseTarget.MinecraftVersion
		}
		if source.MinecraftVersion != "" && source.MinecraftVersion != m.Target.MinecraftVersion && source.MinecraftVersion != baseVersion {
			return fmt.Errorf("mixes Minecraft %s source %q into target Minecraft %s", source.MinecraftVersion, source.ID, m.Target.MinecraftVersion)
		}
	}

	packetIDs := map[uint32]bool{}
	for i, packet := range m.Packets {
		if packetIDs[packet.ID] {
			return fmt.Errorf("packets[%d] duplicates packet id %d", i, packet.ID)
		}
		packetIDs[packet.ID] = true
		if packet.Name == "" {
			return fmt.Errorf("packets[%d] has no name", i)
		}
		if !validDirection(packet.Direction) {
			return fmt.Errorf("packet %s has invalid direction %q", packet.Name, packet.Direction)
		}
		if packet.Direction == DirectionUnknown {
			return fmt.Errorf("packet %s has unknown direction", packet.Name)
		}
		previous := -1
		for fieldIndex, field := range packet.Fields {
			path := fmt.Sprintf("packet %s field[%d]", packet.Name, fieldIndex)
			if field.Ordinal < 0 || field.Ordinal <= previous {
				return fmt.Errorf("%s ordinals are not strictly increasing", path)
			}
			previous = field.Ordinal
			if err := validateField(field, path, sourceIDs); err != nil {
				return err
			}
		}
	}
	if err := validateAdjudications(m, sourceIDs); err != nil {
		return err
	}
	if err := validateOverrides(m, sourceIDs); err != nil {
		return err
	}
	return validateDerivation(m, sourceIDs)
}

func validDirection(direction Direction) bool {
	switch direction {
	case DirectionClientbound, DirectionServerbound, DirectionBidirectional, DirectionUnknown:
		return true
	default:
		return false
	}
}

func validateField(field Field, path string, sourceIDs map[string]bool) error {
	if field.Name == "" {
		return fmt.Errorf("%s has no name", path)
	}
	if len(field.Provenance.Pins) == 0 {
		return fmt.Errorf("%s has no machine-derived provenance pins", path)
	}
	if field.Symmetry != Symmetric && field.Symmetry != Asymmetric {
		return fmt.Errorf("%s has invalid symmetry %q", path, field.Symmetry)
	}
	if field.Symmetry == Symmetric && field.Decode != nil {
		return fmt.Errorf("%s declares symmetric wire but has a decode shape", path)
	}
	if field.Symmetry == Asymmetric && field.Decode == nil {
		return fmt.Errorf("%s declares asymmetric wire without a decode shape", path)
	}
	if field.Decode != nil && reflect.DeepEqual(field.Encode, *field.Decode) {
		return fmt.Errorf("%s declares asymmetry with identical encode/decode shapes", path)
	}
	if err := validateNode(field.Encode, path+".encode", sourceIDs); err != nil {
		return err
	}
	if field.Decode != nil {
		if err := validateNode(*field.Decode, path+".decode", sourceIDs); err != nil {
			return err
		}
	}
	if field.Reserved && field.Encode.Kind != KindReserved {
		return fmt.Errorf("%s reserved metadata disagrees with node kind", path)
	}
	if field.Ignored && field.Encode.Kind != KindIgnored && field.Encode.Kind != KindReserved {
		return fmt.Errorf("%s ignored metadata disagrees with node kind", path)
	}
	if field.Encode.Kind == KindReserved && !field.Reserved && !field.Ignored {
		return fmt.Errorf("%s reserved node is missing compatibility metadata", path)
	}
	if field.Encode.Kind == KindIgnored && !field.Ignored {
		return fmt.Errorf("%s ignored node is missing compatibility metadata", path)
	}
	return validateProvenance(field.Provenance, sourceIDs, path)
}

func validateNode(node Node, path string, sourceIDs map[string]bool) error {
	if err := validateConstraints(node, path); err != nil {
		return err
	}
	switch node.Kind {
	case KindPrimitive:
		if node.Primitive == nil {
			return fmt.Errorf("%s primitive has no exact shape", path)
		}
		canonical, ok := primitiveTable[node.Primitive.Code]
		if !ok || !reflect.DeepEqual(canonical, *node.Primitive) {
			return fmt.Errorf("%s primitive %q has incomplete or invented signedness/width/endianness", path, node.Primitive.Code)
		}
		if node.Primitive.Code == "nbt_le" && !ValidNBTEncoding(node.Encoding) {
			return fmt.Errorf("%s NBT encoding must be %q, %q or %q", path, NBTNetwork, NBTPersistent, NBTNetworkCompound)
		}
	case KindString, KindBytes:
		if node.Prefix == nil {
			return fmt.Errorf("%s has no explicit length prefix", path)
		}
		if node.Kind == KindString && (node.Representation != "text" || node.Encoding == "") {
			return fmt.Errorf("%s must distinguish text encoding from arbitrary bytes", path)
		}
		if node.Kind == KindBytes && (node.Representation != "bytes" || node.Encoding != "") {
			return fmt.Errorf("%s bytes node has text metadata", path)
		}
		if err := validatePrefix(*node.Prefix, path+".prefix"); err != nil {
			return err
		}
	case KindBitset:
		if node.Length == 0 || node.Representation != "bitset" {
			return fmt.Errorf("%s bitset must have a positive bit length and bitset representation", path)
		}
	case KindArray:
		if node.Prefix == nil || node.Element == nil {
			return fmt.Errorf("%s array must have explicit prefix and element", path)
		}
		if err := validatePrefix(*node.Prefix, path+".prefix"); err != nil {
			return err
		}
		if err := validateNode(*node.Element, path+".element", sourceIDs); err != nil {
			return err
		}
	case KindFixedArray:
		if node.Element == nil || node.Length == 0 {
			return fmt.Errorf("%s fixed array must have positive length and element", path)
		}
		if err := validateNode(*node.Element, path+".element", sourceIDs); err != nil {
			return err
		}
	case KindSequence:
		if len(node.Elements) == 0 {
			return fmt.Errorf("%s sequence must contain at least one node", path)
		}
		for index, element := range node.Elements {
			if err := validateNode(element, fmt.Sprintf("%s.elements[%d]", path, index), sourceIDs); err != nil {
				return err
			}
		}
	case KindOptional:
		if node.Value == nil {
			return fmt.Errorf("%s optional has no value shape", path)
		}
		if err := validateNode(*node.Value, path+".value", sourceIDs); err != nil {
			return err
		}
	case KindStruct:
		previous := -1
		for i, field := range node.Fields {
			if field.Ordinal < 0 || field.Ordinal <= previous {
				return fmt.Errorf("%s.fields[%d] ordinals are not strictly increasing", path, i)
			}
			previous = field.Ordinal
			if err := validateField(field, fmt.Sprintf("%s.fields[%d]", path, i), sourceIDs); err != nil {
				return err
			}
		}
	case KindMap:
		if node.Prefix == nil || node.Key == nil || node.Value == nil || node.Representation != "ordered_entries" {
			return fmt.Errorf("%s map must be ordered entries with prefix, key, and value", path)
		}
		if err := validatePrefix(*node.Prefix, path+".prefix"); err != nil {
			return err
		}
		if err := validateNode(*node.Key, path+".key", sourceIDs); err != nil {
			return err
		}
		if err := validateNode(*node.Value, path+".value", sourceIDs); err != nil {
			return err
		}
	case KindUnion:
		if node.Control == nil || len(node.Variants) == 0 {
			return fmt.Errorf("%s union must have explicit control and variants", path)
		}
		if err := validateControl(*node.Control, path+".control"); err != nil {
			return err
		}
		seenValues := map[int64]bool{}
		seenNames := map[string]bool{}
		for i, variant := range node.Variants {
			if variant.Name == "" || seenNames[variant.Name] || seenValues[variant.Value] {
				return fmt.Errorf("%s variant[%d] has duplicate or empty explicit identity", path, i)
			}
			seenNames[variant.Name], seenValues[variant.Value] = true, true
			if !enumValueFitsPrimitive(variant.Value, *node.Control.Primitive) {
				return fmt.Errorf("%s variant[%d] value %d does not fit %s", path, i, variant.Value, node.Control.Primitive.Code)
			}
			if err := validateNode(variant.Encode, fmt.Sprintf("%s.variants[%d].encode", path, i), sourceIDs); err != nil {
				return err
			}
			if variant.Decode != nil {
				if err := validateNode(*variant.Decode, fmt.Sprintf("%s.variants[%d].decode", path, i), sourceIDs); err != nil {
					return err
				}
			}
		}
	case KindConditional:
		if node.CompareTo == "" || len(node.Cases) == 0 {
			return fmt.Errorf("%s conditional must retain discriminator and cases", path)
		}
		for i, oneCase := range node.Cases {
			if oneCase.Value == "" || len(oneCase.Encode) == 0 {
				return fmt.Errorf("%s case[%d] has no explicit value or shape", path, i)
			}
			for j, child := range oneCase.Encode {
				if err := validateNode(child, fmt.Sprintf("%s.cases[%d].encode[%d]", path, i, j), sourceIDs); err != nil {
					return err
				}
			}
			for j, child := range oneCase.Decode {
				if err := validateNode(child, fmt.Sprintf("%s.cases[%d].decode[%d]", path, i, j), sourceIDs); err != nil {
					return err
				}
			}
		}
		if node.Default != nil {
			if err := validateNode(*node.Default, path+".default", sourceIDs); err != nil {
				return err
			}
		}
	case KindEnum:
		if node.Primitive == nil {
			return fmt.Errorf("%s enum has no exact underlying primitive", path)
		}
		canonical, ok := primitiveTable[node.Primitive.Code]
		if !ok || canonical.Width == 0 || !reflect.DeepEqual(canonical, *node.Primitive) {
			return fmt.Errorf("%s enum underlying primitive is not exact", path)
		}
		seenNames := map[string]bool{}
		for i, variant := range node.Variants {
			if variant.Name == "" || seenNames[variant.Name] {
				return fmt.Errorf("%s enum value[%d] lacks a unique nonempty name", path, i)
			}
			seenNames[variant.Name] = true
			if !enumValueFitsPrimitive(variant.Value, *node.Primitive) {
				return fmt.Errorf("%s enum value[%d] ordinal %d does not fit %s", path, i, variant.Value, node.Primitive.Code)
			}
		}
	case KindReserved, KindIgnored:
		if node.Element == nil {
			return fmt.Errorf("%s compatibility node has no preserved wire shape", path)
		}
		if err := validateNode(*node.Element, path+".element", sourceIDs); err != nil {
			return err
		}
	case KindRecursive:
		if node.TypeID == "" && node.Target == "" {
			return fmt.Errorf("%s recursive node has no type identity", path)
		}
	case KindOpaque, KindUnresolved:
		// Manifest nodes are embedded only beneath packet fields; there is no
		// detached type table in which an unresolved definition could be safely
		// parked. Treat even a producer-marked non-reachable node as reachable
		// here so a parser cannot accidentally make a fail-open claim by lying in
		// the flag.
		return fmt.Errorf("%s %s node blocks generation: %s", path, node.Kind, node.Reason)
	case KindVoid:
		// Explicitly consumes no bytes.
	default:
		return fmt.Errorf("%s has unknown node kind %q", path, node.Kind)
	}
	return nil
}

func validateConstraints(node Node, path string) error {
	constraints := node.Constraints
	if constraints == nil {
		return nil
	}
	checkBounds := func(name string, min, max *uint64) error {
		if min != nil && max != nil && *min > *max {
			return fmt.Errorf("%s %s minimum %d exceeds maximum %d", path, name, *min, *max)
		}
		return nil
	}
	if err := checkBounds("length", constraints.MinLength, constraints.MaxLength); err != nil {
		return err
	}
	if err := checkBounds("items", constraints.MinItems, constraints.MaxItems); err != nil {
		return err
	}
	if err := checkBounds("properties", constraints.MinProperties, constraints.MaxProperties); err != nil {
		return err
	}
	if constraints.Minimum != nil && (math.IsNaN(*constraints.Minimum) || math.IsInf(*constraints.Minimum, 0)) || constraints.Maximum != nil && (math.IsNaN(*constraints.Maximum) || math.IsInf(*constraints.Maximum, 0)) {
		return fmt.Errorf("%s numeric constraints must be finite", path)
	}
	if constraints.Minimum != nil && constraints.Maximum != nil && *constraints.Minimum > *constraints.Maximum {
		return fmt.Errorf("%s numeric minimum %g exceeds maximum %g", path, *constraints.Minimum, *constraints.Maximum)
	}
	allowed := func(values ...*uint64) bool {
		for _, value := range values {
			if value != nil {
				return true
			}
		}
		return false
	}
	if allowed(constraints.MinLength, constraints.MaxLength) && node.Kind != KindString && node.Kind != KindBytes {
		return fmt.Errorf("%s length constraints require a string or bytes node", path)
	}
	if constraints.Pattern != "" {
		if node.Kind != KindString {
			return fmt.Errorf("%s pattern constraint requires a string node", path)
		}
		if _, err := regexp.Compile(constraints.Pattern); err != nil {
			return fmt.Errorf("%s has invalid pattern constraint: %w", path, err)
		}
		if !IsSchemaPattern(constraints.Pattern) {
			return fmt.Errorf("%s pattern constraint %q is not anchored with ^ and $", path, constraints.Pattern)
		}
	}
	if allowed(constraints.MinItems, constraints.MaxItems) && node.Kind != KindArray && node.Kind != KindFixedArray {
		return fmt.Errorf("%s item constraints require an array node", path)
	}
	if allowed(constraints.MinProperties, constraints.MaxProperties) && node.Kind != KindMap {
		return fmt.Errorf("%s property constraints require a map node", path)
	}
	if (constraints.Minimum != nil || constraints.Maximum != nil) && node.Kind != KindPrimitive && node.Kind != KindEnum {
		return fmt.Errorf("%s numeric constraints require a primitive or enum node", path)
	}
	if constraints.Minimum != nil || constraints.Maximum != nil {
		if node.Primitive == nil || node.Primitive.Code == "bool" || node.Primitive.Code == "uuid" || node.Primitive.Code == "nbt_le" {
			return fmt.Errorf("%s numeric constraints require a numeric primitive", path)
		}
		if node.Primitive.Code != "f32le" && node.Primitive.Code != "f32be" && node.Primitive.Code != "f64le" && node.Primitive.Code != "f64be" {
			if constraints.Minimum != nil && math.Trunc(*constraints.Minimum) != *constraints.Minimum || constraints.Maximum != nil && math.Trunc(*constraints.Maximum) != *constraints.Maximum {
				return fmt.Errorf("%s integer constraints must use integral bounds", path)
			}
		}
	}
	if node.Kind == KindFixedArray {
		if constraints.MinItems != nil && node.Length < *constraints.MinItems || constraints.MaxItems != nil && node.Length > *constraints.MaxItems {
			return fmt.Errorf("%s fixed array length %d is outside item constraints", path, node.Length)
		}
	}
	return nil
}

// IsSchemaPattern reports whether value is a regular expression anchored to
// the whole string. Source docs also carry prose in pattern fields; it is not.
func IsSchemaPattern(value string) bool {
	if !strings.HasPrefix(value, "^") || !strings.HasSuffix(value, "$") || strings.HasSuffix(value, `\$`) {
		return false
	}
	_, err := regexp.Compile(value)
	return err == nil
}

func validatePrefix(node Node, path string) error {
	if node.Kind != KindPrimitive || node.Primitive == nil {
		return fmt.Errorf("%s must be a primitive length codec", path)
	}
	if !isIntegerPrimitive(node.Primitive.Code) {
		return fmt.Errorf("%s has non-integer length codec %q", path, node.Primitive.Code)
	}
	return validateNode(node, path, nil)
}

func validateControl(node Node, path string) error {
	if node.Kind != KindPrimitive || node.Primitive == nil || !isIntegerPrimitive(node.Primitive.Code) {
		return fmt.Errorf("%s must be an explicit numeric discriminator codec", path)
	}
	return validateNode(node, path, nil)
}

func isIntegerPrimitive(code string) bool {
	switch code {
	case "i8", "u8", "i16le", "u16le", "i16be", "u16be",
		"i32le", "u32le", "i32be", "u32be", "i64le", "u64le",
		"i64be", "u64be", "var_i32", "var_u32", "var_i64", "var_u64",
		"zigzag_i32", "zigzag_i64":
		return true
	default:
		return false
	}
}

func validateProvenance(provenance Provenance, sourceIDs map[string]bool, path string) error {
	if len(provenance.Pins) == 0 {
		return fmt.Errorf("%s has no provenance pins", path)
	}
	previous := ""
	for _, pin := range provenance.Pins {
		if !sourceIDs[pin] {
			return fmt.Errorf("%s provenance references unknown source %q", path, pin)
		}
		if pin <= previous {
			return fmt.Errorf("%s provenance pins are not sorted and unique", path)
		}
		previous = pin
	}
	for i, evidence := range provenance.Evidence {
		if evidence.SourceID == "" || evidence.Locator == "" || !sourceIDs[evidence.SourceID] {
			return fmt.Errorf("%s evidence[%d] has no pinned source locator", path, i)
		}
	}
	return nil
}

func validateAdjudications(m Manifest, sourceIDs map[string]bool) error {
	seen := map[string]bool{}
	sourceRepositories := map[string]string{}
	for _, source := range m.Sources {
		sourceRepositories[source.ID] = repositoryIdentity(source.Locator)
	}
	for i, adjudication := range m.Adjudications {
		if adjudication.ID == "" || adjudication.Target == "" || adjudication.PrePatchContextSHA256 == "" || adjudication.SelectedSource == "" || adjudication.Reason == "" {
			return fmt.Errorf("adjudications[%d] is incomplete", i)
		}
		if seen[adjudication.ID] {
			return fmt.Errorf("duplicate adjudication %q", adjudication.ID)
		}
		seen[adjudication.ID] = true
		if !sourceIDs[adjudication.SelectedSource] {
			return fmt.Errorf("adjudication %q selects unknown source %q", adjudication.ID, adjudication.SelectedSource)
		}
		if len(adjudication.Claims) == 0 {
			return fmt.Errorf("adjudication %q must fingerprint every available source claim", adjudication.ID)
		}
		claimSources := map[string]bool{}
		for _, claim := range adjudication.Claims {
			if claim.Digest == "" || !sourceIDs[claim.SourceID] {
				return fmt.Errorf("adjudication %q contains an incomplete claim fingerprint", adjudication.ID)
			}
			if claimSources[claim.SourceID] {
				return fmt.Errorf("adjudication %q fingerprints source %q more than once", adjudication.ID, claim.SourceID)
			}
			claimSources[claim.SourceID] = true
		}
		if len(adjudication.Evidence) == 0 {
			return fmt.Errorf("adjudication %q has no cited evidence", adjudication.ID)
		}
		selectedRepository := sourceRepositories[adjudication.SelectedSource]
		independent := false
		for _, evidence := range adjudication.Evidence {
			if evidence.SourceID == "" || !sourceIDs[evidence.SourceID] || evidence.Locator == "" {
				return fmt.Errorf("adjudication %q cites an unpinned evidence source", adjudication.ID)
			}
			evidenceRepository := repositoryIdentity(evidence.Locator)
			if evidenceRepository != "" && evidenceRepository != selectedRepository {
				independent = true
			}
		}
		if selectedRepository == "" || !independent {
			return fmt.Errorf("adjudication %q has no evidence independent of selected source %q", adjudication.ID, adjudication.SelectedSource)
		}
	}
	return nil
}

func repositoryIdentity(locator string) string {
	parsed, err := url.Parse(locator)
	if err != nil || parsed.Host == "" {
		return ""
	}
	host := strings.TrimPrefix(strings.TrimSuffix(strings.ToLower(parsed.Hostname()), "."), "www.")
	segments := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if host == "github.com" && len(segments) >= 2 {
		return host + "/" + strings.ToLower(segments[0]) + "/" + strings.TrimSuffix(strings.ToLower(segments[1]), ".git")
	}
	if host == "raw.githubusercontent.com" && len(segments) >= 2 {
		return "github.com/" + strings.ToLower(segments[0]) + "/" + strings.TrimSuffix(strings.ToLower(segments[1]), ".git")
	}
	if host == "api.github.com" && len(segments) >= 3 && segments[0] == "repos" {
		return "github.com/" + strings.ToLower(segments[1]) + "/" + strings.TrimSuffix(strings.ToLower(segments[2]), ".git")
	}
	if host == "codeload.github.com" && len(segments) >= 2 {
		return "github.com/" + strings.ToLower(segments[0]) + "/" + strings.TrimSuffix(strings.ToLower(segments[1]), ".git")
	}
	if len(segments) != 0 && segments[0] != "" {
		return host + "/" + strings.ToLower(segments[0])
	}
	return host
}

func enumValueFitsPrimitive(value int64, primitive PrimitiveShape) bool {
	if primitive.Code == "bool" {
		return value == 0 || value == 1
	}
	if strings.HasPrefix(primitive.Code, "f") || primitive.Width == 0 || primitive.Width > 64 {
		return false
	}
	if primitive.Signed {
		if primitive.Width == 64 {
			return true
		}
		limit := int64(1) << (primitive.Width - 1)
		return value >= -limit && value < limit
	}
	if value < 0 {
		return false
	}
	if primitive.Width == 64 {
		return true
	}
	return uint64(value) < uint64(1)<<primitive.Width
}

func validateOverrides(m Manifest, sourceIDs map[string]bool) error {
	seen := map[string]bool{}
	for i, proof := range m.Overrides {
		if proof.ID == "" || proof.Target == "" || proof.PrePatchNodeSHA256 == "" || proof.PrePatchContextSHA256 == "" || proof.PostPatchNodeSHA256 == "" || proof.Reason == "" || len(proof.Evidence) == 0 {
			return fmt.Errorf("overrides[%d] is incomplete", i)
		}
		if seen[proof.ID] {
			return fmt.Errorf("duplicate override proof %q", proof.ID)
		}
		seen[proof.ID] = true
		for _, evidence := range proof.Evidence {
			if evidence.SourceID == "" || evidence.Locator == "" || !sourceIDs[evidence.SourceID] {
				return fmt.Errorf("override %q cites an unpinned evidence source", proof.ID)
			}
		}
	}
	return nil
}

func validateDerivation(m Manifest, sourceIDs map[string]bool) error {
	if m.Derivation == nil {
		return nil
	}
	proof := m.Derivation
	if proof.BaseTarget.MinecraftVersion == "" || proof.BaseTarget.ProtocolVersion != m.Target.ProtocolVersion || proof.BaseTarget.MinecraftVersion == m.Target.MinecraftVersion {
		return fmt.Errorf("derivation must change Minecraft version while retaining protocol %d", m.Target.ProtocolVersion)
	}
	if !validSHA256(proof.BaseManifestSHA256) || len(proof.Operations) == 0 {
		return fmt.Errorf("derivation has no valid base manifest fingerprint or operations")
	}
	seen := map[string]bool{}
	sourceVersions := make(map[string]string, len(m.Sources))
	for _, source := range m.Sources {
		sourceVersions[source.ID] = source.MinecraftVersion
	}
	for index, operation := range proof.Operations {
		if operation.ID == "" || seen[operation.ID] || operation.PacketID == 0 || operation.FieldOrdinal < 0 || operation.Path == "" || operation.Operation == "" || operation.Reason == "" || !validSHA256(operation.PrePatchNodeSHA256) || !validSHA256(operation.PostPatchNodeSHA256) || len(operation.Evidence) == 0 {
			return fmt.Errorf("derivation operation %d is incomplete or duplicated", index)
		}
		seen[operation.ID] = true
		for _, evidence := range operation.Evidence {
			if evidence.SourceID == "" || evidence.Locator == "" || !sourceIDs[evidence.SourceID] {
				return fmt.Errorf("derivation operation %q cites an unpinned evidence source", operation.ID)
			}
			if sourceVersions[evidence.SourceID] != m.Target.MinecraftVersion {
				return fmt.Errorf("derivation operation %q cites evidence outside target Minecraft %s", operation.ID, m.Target.MinecraftVersion)
			}
		}
	}
	return nil
}

func validSHA256(value string) bool {
	if len(value) != len("sha256:")+64 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil
}

// MarshalStable returns canonical pretty JSON. JSON object keys are encoded in
// deterministic order by encoding/json; slices are sorted on their semantic
// identities in a copy so generation does not depend on filesystem iteration.
func MarshalStable(m Manifest) ([]byte, error) {
	if err := Validate(m); err != nil {
		return nil, err
	}
	normalized := normalize(m)
	data, err := json.MarshalIndent(normalized, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func normalize(m Manifest) Manifest {
	data, _ := json.Marshal(m)
	var out Manifest
	_ = json.Unmarshal(data, &out)
	sort.Slice(out.Sources, func(i, j int) bool { return out.Sources[i].ID < out.Sources[j].ID })
	sort.Slice(out.Packets, func(i, j int) bool {
		if out.Packets[i].ID != out.Packets[j].ID {
			return out.Packets[i].ID < out.Packets[j].ID
		}
		return out.Packets[i].Name < out.Packets[j].Name
	})
	for packetIndex := range out.Packets {
		packet := &out.Packets[packetIndex]
		sort.SliceStable(packet.Fields, func(i, j int) bool {
			if packet.Fields[i].Ordinal != packet.Fields[j].Ordinal {
				return packet.Fields[i].Ordinal < packet.Fields[j].Ordinal
			}
			return packet.Fields[i].Name < packet.Fields[j].Name
		})
		for fieldIndex := range packet.Fields {
			normalizeField(&packet.Fields[fieldIndex])
		}
	}
	if out.Derivation != nil {
		sort.Slice(out.Derivation.Operations, func(i, j int) bool {
			return out.Derivation.Operations[i].ID < out.Derivation.Operations[j].ID
		})
		for index := range out.Derivation.Operations {
			sortEvidence(out.Derivation.Operations[index].Evidence)
		}
	}
	sort.Slice(out.Adjudications, func(i, j int) bool { return out.Adjudications[i].ID < out.Adjudications[j].ID })
	for index := range out.Adjudications {
		adjudication := &out.Adjudications[index]
		sort.SliceStable(adjudication.Claims, func(i, j int) bool {
			return adjudication.Claims[i].SourceID < adjudication.Claims[j].SourceID
		})
		sortEvidence(adjudication.Evidence)
	}
	sort.Slice(out.Overrides, func(i, j int) bool { return out.Overrides[i].ID < out.Overrides[j].ID })
	for index := range out.Overrides {
		sortEvidence(out.Overrides[index].Evidence)
	}
	return out
}

func normalizeField(field *Field) {
	sort.Strings(field.Provenance.Pins)
	sort.Strings(field.Compatibility)
	sortEvidence(field.Provenance.Evidence)
	normalizeNode(&field.Encode)
	if field.Decode != nil {
		normalizeNode(field.Decode)
	}
}

func normalizeNode(node *Node) {
	if node.Prefix != nil {
		normalizeNode(node.Prefix)
	}
	if node.Element != nil {
		normalizeNode(node.Element)
	}
	if node.Value != nil {
		normalizeNode(node.Value)
	}
	for index := range node.Elements {
		normalizeNode(&node.Elements[index])
	}
	sort.SliceStable(node.Fields, func(i, j int) bool {
		if node.Fields[i].Ordinal != node.Fields[j].Ordinal {
			return node.Fields[i].Ordinal < node.Fields[j].Ordinal
		}
		return node.Fields[i].Name < node.Fields[j].Name
	})
	for index := range node.Fields {
		normalizeField(&node.Fields[index])
	}
	if node.Key != nil {
		normalizeNode(node.Key)
	}
	if node.Control != nil {
		normalizeNode(node.Control)
	}
	for index := range node.Variants {
		variant := &node.Variants[index]
		normalizeNode(&variant.Encode)
		if variant.Decode != nil {
			normalizeNode(variant.Decode)
		}
	}
	sort.SliceStable(node.Variants, func(i, j int) bool {
		if node.Variants[i].Value != node.Variants[j].Value {
			return node.Variants[i].Value < node.Variants[j].Value
		}
		// Alias order determines the canonical decoded name, so preserve it.
		if node.Kind == KindEnum {
			return false
		}
		return node.Variants[i].Name < node.Variants[j].Name
	})
	for index := range node.Cases {
		oneCase := &node.Cases[index]
		for childIndex := range oneCase.Encode {
			normalizeNode(&oneCase.Encode[childIndex])
		}
		for childIndex := range oneCase.Decode {
			normalizeNode(&oneCase.Decode[childIndex])
		}
	}
	sort.SliceStable(node.Cases, func(i, j int) bool { return node.Cases[i].Value < node.Cases[j].Value })
	if node.Default != nil {
		normalizeNode(node.Default)
	}
}

func sortEvidence(evidence []Evidence) {
	sort.SliceStable(evidence, func(i, j int) bool {
		left, right := evidence[i], evidence[j]
		if left.SourceID != right.SourceID {
			return left.SourceID < right.SourceID
		}
		if left.Locator != right.Locator {
			return left.Locator < right.Locator
		}
		if left.ClaimFingerprint != right.ClaimFingerprint {
			return left.ClaimFingerprint < right.ClaimFingerprint
		}
		return left.Summary < right.Summary
	})
}

func canonicalBytes(value any) ([]byte, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, data); err != nil {
		return nil, err
	}
	return compact.Bytes(), nil
}
