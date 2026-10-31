// Package reconcile turns independent source claims into the canonical v2
// manifest. It has no precedence rule: equal claims merge source pins and
// different claims require a fingerprinted, evidenced adjudication.
package reconcile

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strings"

	"protocolgen/internal/claims"
	"protocolgen/internal/direction"
	"protocolgen/internal/manifest"
	"protocolgen/internal/nbtencoding"
)

type fieldKey struct {
	PacketID uint32
	Ordinal  int
}

type packetMetadata struct {
	Name      string
	Direction manifest.Direction
}

func Reconcile(target manifest.Target, results []claims.Result, adjudications []manifest.Adjudication) (manifest.Manifest, error) {
	return reconcile(target, results, adjudications, nil, nil)
}

func ReconcileWithDirections(target manifest.Target, results []claims.Result, adjudications []manifest.Adjudication, table direction.Table) (manifest.Manifest, error) {
	return reconcile(target, results, adjudications, &table, nil)
}

func ReconcileWithDirectionsAndNBT(target manifest.Target, results []claims.Result, adjudications []manifest.Adjudication, directions direction.Table, encodings nbtencoding.Table) (manifest.Manifest, error) {
	return reconcile(target, results, adjudications, &directions, &encodings)
}

func reconcile(target manifest.Target, results []claims.Result, adjudications []manifest.Adjudication, table *direction.Table, encodings *nbtencoding.Table) (manifest.Manifest, error) {
	if len(results) == 0 {
		return manifest.Manifest{}, fmt.Errorf("reconcile has no source results")
	}
	allPins := make([]manifest.SourcePin, 0, len(results))
	pins := map[string]manifest.SourcePin{}
	allOverrides := map[string]manifest.OverrideProof{}
	groups := map[fieldKey][]claims.Claim{}
	claimSources := map[fieldKey]map[string]bool{}
	packetMetadataByID := map[uint32]packetMetadata{}
	for resultIndex, result := range results {
		if result.Pin.ID == "" {
			return manifest.Manifest{}, fmt.Errorf("source result %d has no source pin", resultIndex)
		}
		if result.Target != target {
			return manifest.Manifest{}, fmt.Errorf("source %q target does not match; mixing protocol snapshots is forbidden", result.Pin.ID)
		}
		if old, ok := pins[result.Pin.ID]; ok && !reflect.DeepEqual(old, result.Pin) {
			return manifest.Manifest{}, fmt.Errorf("source pin %q is not stable", result.Pin.ID)
		}
		if _, ok := pins[result.Pin.ID]; !ok {
			pins[result.Pin.ID] = result.Pin
			allPins = append(allPins, result.Pin)
		}
		for _, packet := range result.Packets {
			if packet.SourceID != result.Pin.ID {
				return manifest.Manifest{}, fmt.Errorf("packet %d has source %q but result is %q", packet.ID, packet.SourceID, result.Pin.ID)
			}
			metadata := packetMetadata{Name: packet.Name, Direction: packet.Direction}
			if old, ok := packetMetadataByID[packet.ID]; ok {
				if old.Name != metadata.Name {
					return manifest.Manifest{}, fmt.Errorf("packet %d name disagrees: %q vs %q", packet.ID, old.Name, metadata.Name)
				}
				direction, ok := mergeDirection(old.Direction, metadata.Direction)
				if !ok {
					return manifest.Manifest{}, fmt.Errorf("packet %d direction disagrees: %q vs %q", packet.ID, old.Direction, metadata.Direction)
				}
				metadata.Direction = direction
			}
			packetMetadataByID[packet.ID] = metadata
		}
		for _, claim := range result.Claims {
			if claim.SourceID != result.Pin.ID {
				return manifest.Manifest{}, fmt.Errorf("claim %s has source %q but result is %q", claim.FieldPath, claim.SourceID, result.Pin.ID)
			}
			key := fieldKey{PacketID: claim.PacketID, Ordinal: claim.Ordinal}
			if claimSources[key] == nil {
				claimSources[key] = map[string]bool{}
			}
			if claimSources[key][claim.SourceID] {
				return manifest.Manifest{}, fmt.Errorf("source %q supplied duplicate claims for %s", claim.SourceID, claim.FieldPath)
			}
			claimSources[key][claim.SourceID] = true
			groups[key] = append(groups[key], claim)
		}
		for _, proof := range result.Overrides {
			if old, ok := allOverrides[proof.ID]; ok && !reflect.DeepEqual(old, proof) {
				return manifest.Manifest{}, fmt.Errorf("override proof %q is not stable", proof.ID)
			}
			allOverrides[proof.ID] = proof
		}
	}
	if len(groups) == 0 && len(packetMetadataByID) == 0 {
		return manifest.Manifest{}, fmt.Errorf("reconcile has no packet claims")
	}

	keys := make([]fieldKey, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].PacketID != keys[j].PacketID {
			return keys[i].PacketID < keys[j].PacketID
		}
		return keys[i].Ordinal < keys[j].Ordinal
	})

	packets := map[uint32]*manifest.Packet{}
	for id, metadata := range packetMetadataByID {
		packets[id] = &manifest.Packet{ID: id, Name: metadata.Name, Direction: metadata.Direction}
	}
	usedAdjudications := map[string]bool{}
	var selectionErrors []string
	for _, key := range keys {
		group := groups[key]
		sort.SliceStable(group, func(i, j int) bool { return group[i].SourceID < group[j].SourceID })
		selected, pinsForField, evidence, used, err := selectClaim(target, group, adjudications)
		if err != nil {
			selectionErrors = append(selectionErrors, err.Error())
			continue
		}
		if used != "" {
			usedAdjudications[used] = true
		}
		packet := packets[selected.PacketID]
		if packet == nil {
			packet = &manifest.Packet{ID: selected.PacketID, Name: selected.PacketName, Direction: selected.Direction}
			packets[selected.PacketID] = packet
		}
		direction, directionsAgree := mergeDirection(packet.Direction, selected.Direction)
		if packet.Name != selected.PacketName || !directionsAgree {
			return manifest.Manifest{}, fmt.Errorf("packet %d metadata disagrees without a field adjudication", selected.PacketID)
		}
		packet.Direction = direction
		field := manifest.Field{
			Ordinal: selected.Ordinal, Name: selected.Name, Semantic: selected.Semantic, TypeID: selected.TypeID,
			Encode: selected.Encode, Decode: selected.Decode, Symmetry: selected.Symmetry,
			Reserved: selected.Reserved, Ignored: selected.Ignored, Compatibility: append([]string(nil), selected.Compatibility...),
			Provenance: manifest.Provenance{Pins: pinsForField, Evidence: evidence},
		}
		packet.Fields = append(packet.Fields, field)
	}
	if len(selectionErrors) != 0 {
		return manifest.Manifest{}, fmt.Errorf("reconciliation blocked by %d provenance gap(s):\n- %s", len(selectionErrors), strings.Join(selectionErrors, "\n- "))
	}

	packetsOut := make([]manifest.Packet, 0, len(packets))
	for _, packet := range packets {
		sort.Slice(packet.Fields, func(i, j int) bool { return packet.Fields[i].Ordinal < packet.Fields[j].Ordinal })
		packetsOut = append(packetsOut, *packet)
	}
	sort.Slice(packetsOut, func(i, j int) bool { return packetsOut[i].ID < packetsOut[j].ID })
	sort.Slice(allPins, func(i, j int) bool { return allPins[i].ID < allPins[j].ID })

	used := make([]manifest.Adjudication, 0, len(usedAdjudications))
	for _, adjudication := range adjudications {
		if usedAdjudications[adjudication.ID] {
			used = append(used, adjudication)
		}
	}
	proofs := make([]manifest.OverrideProof, 0, len(allOverrides))
	for _, proof := range allOverrides {
		proofs = append(proofs, proof)
	}
	sort.Slice(proofs, func(i, j int) bool { return proofs[i].ID < proofs[j].ID })
	result := manifest.Manifest{SchemaVersion: manifest.SchemaVersion, Target: target, Sources: allPins, Packets: packetsOut, Adjudications: used, Overrides: proofs}
	if table != nil {
		if err := table.Apply(&result); err != nil {
			return manifest.Manifest{}, err
		}
	}
	if encodings != nil {
		// Each overlay path owns its nodes, even when source trees shared pointers.
		data, err := json.Marshal(result)
		if err != nil {
			return manifest.Manifest{}, err
		}
		var owned manifest.Manifest
		if err := json.Unmarshal(data, &owned); err != nil {
			return manifest.Manifest{}, err
		}
		result = owned
		if err := encodings.Apply(&result); err != nil {
			return manifest.Manifest{}, err
		}
	}
	if err := manifest.Validate(result); err != nil {
		return manifest.Manifest{}, err
	}
	return result, nil
}

func mergeDirection(left, right manifest.Direction) (manifest.Direction, bool) {
	if left == manifest.DirectionUnknown {
		return right, true
	}
	if right == manifest.DirectionUnknown || left == right {
		return left, true
	}
	return "", false
}

func selectClaim(target manifest.Target, group []claims.Claim, adjudications []manifest.Adjudication) (claims.Claim, []string, []manifest.Evidence, string, error) {
	if len(group) == 0 {
		return claims.Claim{}, nil, nil, "", fmt.Errorf("empty claim group")
	}
	merged, pinIDs, compatible := mergeClaimGroupCandidate(group)
	if compatible && len(pinIDs) >= 2 {
		return merged, pinIDs, nil, "", nil
	}

	context, err := claims.ContextFingerprint(target, group)
	if err != nil {
		return claims.Claim{}, nil, nil, "", err
	}
	for _, adjudication := range adjudications {
		if adjudication.Target != group[0].FieldPath {
			continue
		}
		if adjudication.PrePatchContextSHA256 != context {
			return claims.Claim{}, nil, nil, "", fmt.Errorf(
				"stale adjudication %q for %s: pre-patch context fingerprint is %s, current context is %s",
				adjudication.ID,
				adjudication.Target,
				adjudication.PrePatchContextSHA256,
				context,
			)
		}
		if err := matchClaimFingerprints(adjudication, group); err != nil {
			return claims.Claim{}, nil, nil, "", fmt.Errorf("stale adjudication %q for %s: %w", adjudication.ID, adjudication.Target, err)
		}
		var selected *claims.Claim
		for i := range group {
			if group[i].SourceID == adjudication.SelectedSource {
				selected = &group[i]
				break
			}
		}
		if selected == nil {
			return claims.Claim{}, nil, nil, "", fmt.Errorf("adjudication %q selects a claim not present in disagreement", adjudication.ID)
		}
		selectedClaim := *selected
		if compatible && containsString(pinIDs, selected.SourceID) {
			selectedClaim = merged
		}
		for _, claim := range group {
			if claim.SourceID != selected.SourceID && hasBinaryTextRepresentationDifference(selectedClaim.Encode, claim.Encode) {
				selectedClaim = enrichClaimMetadata(selectedClaim, claim)
			}
		}
		return selectedClaim, []string{selected.SourceID}, append([]manifest.Evidence(nil), adjudication.Evidence...), adjudication.ID, nil
	}
	return claims.Claim{}, nil, nil, "", fmt.Errorf("source claims for %s lack two byte-equivalent complete claims (%s); an evidenced fingerprinted adjudication is required", group[0].FieldPath, claimCoverageSummary(group))
}

func claimCoverageSummary(group []claims.Claim) string {
	parts := make([]string, 0, len(group))
	for _, claim := range group {
		coverage := "complete"
		if hasEvidenceGap(claim.Encode) || claim.Decode != nil && hasEvidenceGap(*claim.Decode) {
			coverage = "incomplete"
		}
		parts = append(parts, claim.SourceID+"="+coverage)
	}
	return strings.Join(parts, ",")
}

func mergeClaimGroupCandidate(group []claims.Claim) (claims.Claim, []string, bool) {
	selected := group[0]
	for _, claim := range group[1:] {
		if semanticScore(claim) > semanticScore(selected) {
			selected = claim
		}
	}
	merged := selected
	for _, claim := range group {
		if claim.PacketID != merged.PacketID || claim.Ordinal != merged.Ordinal || claim.Symmetry != merged.Symmetry || claim.Reserved != merged.Reserved || claim.Ignored != merged.Ignored || !reflect.DeepEqual(claim.Compatibility, merged.Compatibility) {
			return claims.Claim{}, nil, false
		}
		direction, ok := mergeDirection(merged.Direction, claim.Direction)
		if !ok {
			return claims.Claim{}, nil, false
		}
		merged.Direction = direction
		encode, ok := mergeNode(merged.Encode, claim.Encode)
		if !ok {
			return claims.Claim{}, nil, false
		}
		merged.Encode = encode
		leftDecode, rightDecode := merged.Encode, claim.Encode
		if merged.Decode != nil {
			leftDecode = *merged.Decode
		}
		if claim.Decode != nil {
			rightDecode = *claim.Decode
		}
		decode, ok := mergeNode(leftDecode, rightDecode)
		if !ok {
			return claims.Claim{}, nil, false
		}
		if reflect.DeepEqual(wireNode(merged.Encode), wireNode(decode)) {
			merged.Decode = nil
		} else {
			merged.Decode = &decode
		}
	}
	var pinIDs []string
	mergedWire := normalizedWireComparable(merged)
	for _, claim := range group {
		if !hasEvidenceGap(claim.Encode) && (claim.Decode == nil || !hasEvidenceGap(*claim.Decode)) && reflect.DeepEqual(normalizedWireComparable(claim), mergedWire) {
			pinIDs = append(pinIDs, claim.SourceID)
		}
	}
	pinIDs = mergeStrings(nil, pinIDs)
	return merged, pinIDs, true
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func enrichClaimMetadata(selected, other claims.Claim) claims.Claim {
	if semanticScore(other) > semanticScore(selected) {
		selected.Semantic = other.Semantic
		selected.TypeID = other.TypeID
	}
	selected.Encode = enrichNodeMetadata(selected.Encode, other.Encode)
	if selected.Decode != nil && other.Decode != nil {
		decode := enrichNodeMetadata(*selected.Decode, *other.Decode)
		selected.Decode = &decode
	}
	return selected
}

func enrichNodeMetadata(selected, other manifest.Node) manifest.Node {
	if nodeSemanticScore(other) > nodeSemanticScore(selected) {
		selected.Semantic = other.Semantic
		selected.TypeID = other.TypeID
	}
	switch {
	case selected.Kind == manifest.KindStruct && other.Kind == manifest.KindStruct:
		selected.Fields = append([]manifest.Field(nil), selected.Fields...)
		for i := range selected.Fields {
			for j := range other.Fields {
				if selected.Fields[i].Ordinal != other.Fields[j].Ordinal {
					continue
				}
				if selected.Fields[i].TypeID == "" && other.Fields[j].TypeID != "" && !hasEvidenceGap(other.Fields[j].Encode) {
					selected.Fields[i].TypeID = other.Fields[j].TypeID
					selected.Fields[i].Semantic = other.Fields[j].Semantic
				}
				selected.Fields[i].Encode = enrichNodeMetadata(selected.Fields[i].Encode, other.Fields[j].Encode)
				if selected.Fields[i].Decode != nil && other.Fields[j].Decode != nil {
					decode := enrichNodeMetadata(*selected.Fields[i].Decode, *other.Fields[j].Decode)
					selected.Fields[i].Decode = &decode
				}
				break
			}
		}
	case selected.Kind == manifest.KindArray && other.Kind == manifest.KindArray,
		selected.Kind == manifest.KindFixedArray && other.Kind == manifest.KindFixedArray:
		selected.Element = enrichNodeMetadataPointer(selected.Element, other.Element)
	case selected.Kind == manifest.KindOptional && other.Kind == manifest.KindOptional,
		selected.Kind == manifest.KindReserved && other.Kind == manifest.KindReserved,
		selected.Kind == manifest.KindIgnored && other.Kind == manifest.KindIgnored:
		selected.Value = enrichNodeMetadataPointer(selected.Value, other.Value)
	case selected.Kind == manifest.KindMap && other.Kind == manifest.KindMap:
		selected.Key = enrichNodeMetadataPointer(selected.Key, other.Key)
		selected.Value = enrichNodeMetadataPointer(selected.Value, other.Value)
	case selected.Kind == manifest.KindSequence && other.Kind == manifest.KindSequence:
		selected.Elements = append([]manifest.Node(nil), selected.Elements...)
		for i := range selected.Elements {
			if i < len(other.Elements) {
				selected.Elements[i] = enrichNodeMetadata(selected.Elements[i], other.Elements[i])
			}
		}
	case selected.Kind == manifest.KindUnion && other.Kind == manifest.KindUnion:
		selected.Variants = append([]manifest.Variant(nil), selected.Variants...)
		for i := range selected.Variants {
			for j := range other.Variants {
				if selected.Variants[i].Value == other.Variants[j].Value {
					selected.Variants[i].Encode = enrichNodeMetadata(selected.Variants[i].Encode, other.Variants[j].Encode)
					break
				}
			}
		}
	}
	return selected
}

func enrichNodeMetadataPointer(selected, other *manifest.Node) *manifest.Node {
	if selected == nil || other == nil {
		return selected
	}
	result := enrichNodeMetadata(*selected, *other)
	return &result
}

func hasBinaryTextRepresentationDifference(selected, other manifest.Node) bool {
	if selected.Kind == manifest.KindBytes && other.Kind == manifest.KindString {
		return true
	}
	switch {
	case selected.Kind == manifest.KindStruct && other.Kind == manifest.KindStruct:
		for i := range selected.Fields {
			for j := range other.Fields {
				if selected.Fields[i].Ordinal == other.Fields[j].Ordinal && hasBinaryTextRepresentationDifference(selected.Fields[i].Encode, other.Fields[j].Encode) {
					return true
				}
			}
		}
	case selected.Kind == manifest.KindArray && other.Kind == manifest.KindArray,
		selected.Kind == manifest.KindFixedArray && other.Kind == manifest.KindFixedArray:
		return selected.Element != nil && other.Element != nil && hasBinaryTextRepresentationDifference(*selected.Element, *other.Element)
	case selected.Kind == manifest.KindOptional && other.Kind == manifest.KindOptional,
		selected.Kind == manifest.KindReserved && other.Kind == manifest.KindReserved,
		selected.Kind == manifest.KindIgnored && other.Kind == manifest.KindIgnored:
		return selected.Value != nil && other.Value != nil && hasBinaryTextRepresentationDifference(*selected.Value, *other.Value)
	case selected.Kind == manifest.KindMap && other.Kind == manifest.KindMap:
		return selected.Value != nil && other.Value != nil && hasBinaryTextRepresentationDifference(*selected.Value, *other.Value)
	case selected.Kind == manifest.KindSequence && other.Kind == manifest.KindSequence:
		for i := range selected.Elements {
			if i < len(other.Elements) && hasBinaryTextRepresentationDifference(selected.Elements[i], other.Elements[i]) {
				return true
			}
		}
	case selected.Kind == manifest.KindUnion && other.Kind == manifest.KindUnion:
		for i := range selected.Variants {
			for j := range other.Variants {
				if selected.Variants[i].Value == other.Variants[j].Value && hasBinaryTextRepresentationDifference(selected.Variants[i].Encode, other.Variants[j].Encode) {
					return true
				}
			}
		}
	}
	return false
}

func normalizedWireComparable(claim claims.Claim) claims.Claim {
	claim = wireComparable(claim)
	claim.Direction = ""
	if claim.Decode != nil && reflect.DeepEqual(claim.Encode, *claim.Decode) {
		claim.Decode = nil
	}
	return claim
}

func hasConcreteEvidence(node manifest.Node) bool {
	if node.Kind == manifest.KindUnresolved || node.Kind == manifest.KindOpaque {
		return false
	}
	if node.Kind != manifest.KindStruct && node.Kind != manifest.KindSequence && node.Kind != manifest.KindUnion && node.Kind != manifest.KindConditional {
		return true
	}
	for _, child := range node.Elements {
		if hasConcreteEvidence(child) {
			return true
		}
	}
	for _, field := range node.Fields {
		if hasConcreteEvidence(field.Encode) {
			return true
		}
	}
	for _, variant := range node.Variants {
		if hasConcreteEvidence(variant.Encode) {
			return true
		}
	}
	for _, oneCase := range node.Cases {
		for _, child := range oneCase.Encode {
			if hasConcreteEvidence(child) {
				return true
			}
		}
	}
	return false
}

func mergeNode(left, right manifest.Node) (manifest.Node, bool) {
	constraints, constraintsOK := mergeConstraints(left.Constraints, right.Constraints)
	if !constraintsOK {
		return manifest.Node{}, false
	}
	if hasEvidenceGap(left) && (left.Kind == manifest.KindUnresolved || left.Kind == manifest.KindOpaque) {
		right.Constraints = constraints
		return right, true
	}
	if hasEvidenceGap(right) && (right.Kind == manifest.KindUnresolved || right.Kind == manifest.KindOpaque) {
		left.Constraints = constraints
		return left, true
	}
	if left.Kind == manifest.KindEnum && right.Kind == manifest.KindPrimitive && reflect.DeepEqual(left.Primitive, right.Primitive) {
		left.Constraints = constraints
		return left, true
	}
	if right.Kind == manifest.KindEnum && left.Kind == manifest.KindPrimitive && reflect.DeepEqual(left.Primitive, right.Primitive) {
		right.Constraints = constraints
		return right, true
	}
	if left.Kind == manifest.KindOptional && left.Value != nil && hasEvidenceGap(*left.Value) && right.Kind != manifest.KindOptional {
		result := left
		value := right
		result.Value = &value
		return result, true
	}
	if right.Kind == manifest.KindOptional && right.Value != nil && hasEvidenceGap(*right.Value) && left.Kind != manifest.KindOptional {
		result := right
		value := left
		result.Value = &value
		return result, true
	}
	if reflect.DeepEqual(wireNode(left), wireNode(right)) {
		result := left
		if nodeSemanticScore(right) > nodeSemanticScore(left) {
			result = right
		}
		result.Constraints = constraints
		return result, true
	}
	if left.Kind != right.Kind {
		return manifest.Node{}, false
	}
	result := left
	result.Constraints = constraints
	if nodeSemanticScore(right) > nodeSemanticScore(left) {
		result.Semantic, result.TypeID = right.Semantic, right.TypeID
	}
	var ok bool
	switch left.Kind {
	case manifest.KindStruct:
		if len(left.Fields) != len(right.Fields) {
			return manifest.Node{}, false
		}
		result.Fields = append([]manifest.Field(nil), left.Fields...)
		for index := range left.Fields {
			if left.Fields[index].Ordinal != right.Fields[index].Ordinal {
				return manifest.Node{}, false
			}
			result.Fields[index], ok = mergeField(left.Fields[index], right.Fields[index])
			if !ok {
				return manifest.Node{}, false
			}
		}
	case manifest.KindArray, manifest.KindFixedArray:
		if left.Length != right.Length {
			return manifest.Node{}, false
		}
		result.Prefix, ok = mergeNodePointer(left.Prefix, right.Prefix)
		if !ok {
			return manifest.Node{}, false
		}
		result.Element, ok = mergeNodePointer(left.Element, right.Element)
		if !ok {
			return manifest.Node{}, false
		}
	case manifest.KindOptional, manifest.KindReserved, manifest.KindIgnored:
		result.Value, ok = mergeNodePointer(left.Value, right.Value)
		if left.Kind == manifest.KindReserved || left.Kind == manifest.KindIgnored {
			result.Element, ok = mergeNodePointer(left.Element, right.Element)
		}
		if !ok {
			return manifest.Node{}, false
		}
	case manifest.KindMap:
		result.Prefix, ok = mergeNodePointer(left.Prefix, right.Prefix)
		if !ok {
			return manifest.Node{}, false
		}
		result.Key, ok = mergeNodePointer(left.Key, right.Key)
		if !ok {
			return manifest.Node{}, false
		}
		result.Value, ok = mergeNodePointer(left.Value, right.Value)
		if !ok {
			return manifest.Node{}, false
		}
	case manifest.KindSequence:
		if len(left.Elements) != len(right.Elements) {
			return manifest.Node{}, false
		}
		result.Elements = append([]manifest.Node(nil), left.Elements...)
		for index := range left.Elements {
			result.Elements[index], ok = mergeNode(left.Elements[index], right.Elements[index])
			if !ok {
				return manifest.Node{}, false
			}
		}
	case manifest.KindUnion:
		if len(left.Variants) != len(right.Variants) {
			return manifest.Node{}, false
		}
		result.Control, ok = mergeNodePointer(left.Control, right.Control)
		if !ok {
			return manifest.Node{}, false
		}
		result.Variants = append([]manifest.Variant(nil), left.Variants...)
		for index := range left.Variants {
			if left.Variants[index].Value != right.Variants[index].Value {
				return manifest.Node{}, false
			}
			result.Variants[index].Encode, ok = mergeNode(left.Variants[index].Encode, right.Variants[index].Encode)
			if !ok {
				return manifest.Node{}, false
			}
			leftDecode, rightDecode := left.Variants[index].Encode, right.Variants[index].Encode
			if left.Variants[index].Decode != nil {
				leftDecode = *left.Variants[index].Decode
			}
			if right.Variants[index].Decode != nil {
				rightDecode = *right.Variants[index].Decode
			}
			decode, compatible := mergeNode(leftDecode, rightDecode)
			if !compatible {
				return manifest.Node{}, false
			}
			if left.Variants[index].Decode != nil || right.Variants[index].Decode != nil {
				result.Variants[index].Decode = &decode
			}
		}
	case manifest.KindString, manifest.KindBytes:
		if left.Encoding != "" && right.Encoding != "" && left.Encoding != right.Encoding {
			return manifest.Node{}, false
		}
		if result.Encoding == "" {
			result.Encoding = right.Encoding
		}
		if left.Representation != "" && right.Representation != "" && left.Representation != right.Representation {
			return manifest.Node{}, false
		}
		if result.Representation == "" {
			result.Representation = right.Representation
		}
		result.Prefix, ok = mergeNodePointer(left.Prefix, right.Prefix)
		if !ok {
			return manifest.Node{}, false
		}
	case manifest.KindEnum:
		if !reflect.DeepEqual(left.Primitive, right.Primitive) {
			return manifest.Node{}, false
		}
		if len(right.Variants) > len(left.Variants) {
			return right, true
		}
		return left, true
	default:
		return manifest.Node{}, false
	}
	return result, true
}

func mergeConstraints(left, right *manifest.Constraints) (*manifest.Constraints, bool) {
	if left == nil {
		return right, true
	}
	if right == nil {
		return left, true
	}
	result := *left
	mergeUint := func(dst **uint64, value *uint64, choose func(uint64, uint64) uint64) {
		if value == nil {
			return
		}
		if *dst == nil {
			*dst = value
			return
		}
		selected := choose(**dst, *value)
		*dst = &selected
	}
	minimum := func(left, right uint64) uint64 {
		if left < right {
			return left
		}
		return right
	}
	maximum := func(left, right uint64) uint64 {
		if left > right {
			return left
		}
		return right
	}
	mergeUint(&result.MinLength, right.MinLength, minimum)
	mergeUint(&result.MaxLength, right.MaxLength, maximum)
	mergeUint(&result.MinItems, right.MinItems, minimum)
	mergeUint(&result.MaxItems, right.MaxItems, maximum)
	mergeUint(&result.MinProperties, right.MinProperties, minimum)
	mergeUint(&result.MaxProperties, right.MaxProperties, maximum)
	mergeNumber := func(dst **json.Number, value *json.Number, minimum bool) bool {
		if value == nil {
			return true
		}
		if *dst == nil {
			*dst = value
			return true
		}
		leftValue, leftOK := new(big.Rat).SetString(string(**dst))
		rightValue, rightOK := new(big.Rat).SetString(string(*value))
		if !leftOK || !rightOK {
			return false
		}
		comparison := rightValue.Cmp(leftValue)
		if minimum && comparison < 0 || !minimum && comparison > 0 {
			*dst = value
		}
		return true
	}
	if !mergeNumber(&result.Minimum, right.Minimum, true) || !mergeNumber(&result.Maximum, right.Maximum, false) {
		return nil, false
	}
	if result.Pattern != "" && right.Pattern != "" && result.Pattern != right.Pattern {
		return nil, false
	}
	if result.Pattern == "" {
		result.Pattern = right.Pattern
	}
	return &result, true
}

func mergeField(left, right manifest.Field) (manifest.Field, bool) {
	result := left
	if semanticScoreForField(right) > semanticScoreForField(left) {
		result.Name, result.Semantic, result.TypeID = right.Name, right.Semantic, right.TypeID
	}
	leftGap, rightGap := hasEvidenceGap(left.Encode), hasEvidenceGap(right.Encode)
	encode, ok := mergeNode(left.Encode, right.Encode)
	if !ok {
		return manifest.Field{}, false
	}
	result.Encode = encode
	if !leftGap && !rightGap {
		result.Provenance.Pins = mergeStrings(left.Provenance.Pins, right.Provenance.Pins)
	} else if leftGap && !rightGap {
		result.Provenance = right.Provenance
	}
	return result, true
}

func mergeNodePointer(left, right *manifest.Node) (*manifest.Node, bool) {
	if left == nil || right == nil {
		return nil, left == nil && right == nil
	}
	merged, ok := mergeNode(*left, *right)
	return &merged, ok
}

func semanticScoreForField(field manifest.Field) int {
	score := nodeSemanticScore(field.Encode)
	if field.Semantic != "" {
		score++
	}
	if field.TypeID != "" {
		score += 2
	}
	return score
}

func mergeStrings(left, right []string) []string {
	seen := map[string]bool{}
	for _, value := range append(append([]string(nil), left...), right...) {
		seen[value] = true
	}
	result := make([]string, 0, len(seen))
	for value := range seen {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func hasEvidenceGap(node manifest.Node) bool {
	if node.Kind == manifest.KindUnresolved || node.Kind == manifest.KindOpaque {
		return true
	}
	for _, child := range []*manifest.Node{node.Prefix, node.Element, node.Value, node.Key, node.Control, node.Default} {
		if child != nil && hasEvidenceGap(*child) {
			return true
		}
	}
	for _, child := range node.Elements {
		if hasEvidenceGap(child) {
			return true
		}
	}
	for _, field := range node.Fields {
		if hasEvidenceGap(field.Encode) || field.Decode != nil && hasEvidenceGap(*field.Decode) {
			return true
		}
	}
	for _, variant := range node.Variants {
		if hasEvidenceGap(variant.Encode) || variant.Decode != nil && hasEvidenceGap(*variant.Decode) {
			return true
		}
	}
	for _, oneCase := range node.Cases {
		for _, child := range oneCase.Encode {
			if hasEvidenceGap(child) {
				return true
			}
		}
		for _, child := range oneCase.Decode {
			if hasEvidenceGap(child) {
				return true
			}
		}
	}
	return false
}

func comparable(claim claims.Claim) claims.Claim {
	claim.SourceID = ""
	claim.Locator = ""
	return claim
}

func wireComparable(claim claims.Claim) claims.Claim {
	claim = comparable(claim)
	claim.PacketName = ""
	claim.FieldPath = ""
	claim.Name = ""
	claim.Semantic = ""
	claim.TypeID = ""
	claim.Encode = wireNode(claim.Encode)
	if claim.Decode != nil {
		decode := wireNode(*claim.Decode)
		claim.Decode = &decode
	}
	return claim
}

// wireNode also treats enum labels as metadata when merging source shapes.
func wireNode(node manifest.Node) manifest.Node {
	node = claims.NormalizeWireNode(node)
	if node.Kind == manifest.KindEnum {
		return manifest.Node{Kind: manifest.KindPrimitive, Primitive: node.Primitive}
	}
	node.Prefix = wireNodePointer(node.Prefix)
	node.Element = wireNodePointer(node.Element)
	node.Value = wireNodePointer(node.Value)
	node.Key = wireNodePointer(node.Key)
	node.Control = wireNodePointer(node.Control)
	node.Default = wireNodePointer(node.Default)
	for index := range node.Elements {
		node.Elements[index] = wireNode(node.Elements[index])
	}
	for index := range node.Fields {
		field := &node.Fields[index]
		field.Name = ""
		field.Semantic = ""
		field.TypeID = ""
		field.Provenance = manifest.Provenance{}
		field.Encode = wireNode(field.Encode)
		if field.Decode != nil {
			decode := wireNode(*field.Decode)
			field.Decode = &decode
		}
	}
	for index := range node.Variants {
		node.Variants[index].Name = ""
		node.Variants[index].Encode = wireNode(node.Variants[index].Encode)
		if node.Variants[index].Decode != nil {
			decode := wireNode(*node.Variants[index].Decode)
			node.Variants[index].Decode = &decode
		}
	}
	for caseIndex := range node.Cases {
		for nodeIndex := range node.Cases[caseIndex].Encode {
			node.Cases[caseIndex].Encode[nodeIndex] = wireNode(node.Cases[caseIndex].Encode[nodeIndex])
		}
		for nodeIndex := range node.Cases[caseIndex].Decode {
			node.Cases[caseIndex].Decode[nodeIndex] = wireNode(node.Cases[caseIndex].Decode[nodeIndex])
		}
	}
	return node
}

func wireNodePointer(node *manifest.Node) *manifest.Node {
	if node == nil {
		return nil
	}
	result := wireNode(*node)
	return &result
}

func semanticScore(claim claims.Claim) int {
	score := 0
	if claim.Semantic != "" {
		score++
	}
	if claim.TypeID != "" {
		score += 2
	}
	return score + nodeSemanticScore(claim.Encode)
}

func nodeSemanticScore(node manifest.Node) int {
	score := 0
	if node.Semantic != "" {
		score++
	}
	if node.TypeID != "" {
		score += 2
	}
	if node.Kind == manifest.KindEnum {
		score += 10 + len(node.Variants)
	}
	for _, field := range node.Fields {
		score += nodeSemanticScore(field.Encode)
	}
	return score
}

func matchClaimFingerprints(adjudication manifest.Adjudication, group []claims.Claim) error {
	want := map[string]string{}
	for _, claim := range group {
		digest, err := claims.Fingerprint(claim)
		if err != nil {
			return err
		}
		want[claim.SourceID] = digest
	}
	got := map[string]string{}
	for _, claim := range adjudication.Claims {
		if _, exists := got[claim.SourceID]; exists {
			return fmt.Errorf("duplicate fingerprint for source %q", claim.SourceID)
		}
		got[claim.SourceID] = claim.Digest
	}
	if !reflect.DeepEqual(want, got) {
		return fmt.Errorf("claim fingerprints no longer match: adjudication has %v, current claims are %v", got, want)
	}
	return nil
}
