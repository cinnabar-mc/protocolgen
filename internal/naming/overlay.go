package naming

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"protocolgen/internal/manifest"
)

// LoadOverlay reads and validates a reviewed naming document for a manifest.
func LoadOverlay(path string, m manifest.Manifest) (Overlay, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Overlay{}, fmt.Errorf("read naming overlay: %w", err)
	}
	var document Document
	if err := json.Unmarshal(data, &document); err != nil {
		return Overlay{}, fmt.Errorf("parse naming overlay: %w", err)
	}
	overlay := Overlay{Names: make(map[string]string, len(document.Entries))}
	if err := ValidateOverlay(m, document); err != nil {
		return Overlay{}, err
	}
	for _, entry := range document.Entries {
		overlay.Names[entry.TypeID] = entry.Name
	}
	return overlay, nil
}

// ValidateOverlay enforces target matching, fail-closed TypeIDs, and reviewed
// canonical PascalCase names.
func ValidateOverlay(m manifest.Manifest, document Document) error {
	if document.SchemaVersion != 1 {
		return fmt.Errorf("naming overlay schema_version %d is not v1", document.SchemaVersion)
	}
	if document.Target.MinecraftVersion != m.Target.MinecraftVersion || document.Target.ProtocolVersion != m.Target.ProtocolVersion {
		return fmt.Errorf("naming overlay target does not match manifest target")
	}
	existing := TypeIDs(m)
	seen := make(map[string]bool, len(document.Entries))
	usedNames := make(map[string]string, len(document.Entries))
	for index, entry := range document.Entries {
		if entry.TypeID == "" || entry.Name == "" || entry.Rationale == "" {
			return fmt.Errorf("naming overlay entry[%d] requires type_id, name, and rationale", index)
		}
		if !existing[entry.TypeID] {
			return fmt.Errorf("naming overlay entry[%d] refers to stale TypeID %q", index, entry.TypeID)
		}
		if seen[entry.TypeID] {
			return fmt.Errorf("naming overlay contains duplicate TypeID %q", entry.TypeID)
		}
		seen[entry.TypeID] = true
		if PublicTypeName(entry.Name) != entry.Name {
			return fmt.Errorf("naming overlay entry %q name %q is not canonical PascalCase", entry.TypeID, entry.Name)
		}
		if previous, ok := usedNames[entry.Name]; ok {
			return fmt.Errorf("naming overlay names %q for both %q and %q", entry.Name, previous, entry.TypeID)
		}
		usedNames[entry.Name] = entry.TypeID
	}
	return nil
}

// ValidateRequiredEntries reports all artifact TypeIDs that lack review.
func ValidateRequiredEntries(m manifest.Manifest, overlay Overlay) error {
	var offenders []string
	seen := map[string]bool{}
	for typeID := range TypeIDs(m) {
		if LooksLikeArtifact(typeID) && overlay.Names[typeID] == "" {
			seen[typeID] = true
		}
	}
	for _, packet := range m.Packets {
		for _, field := range packet.Fields {
			collectVariantArtifacts(field.Encode, "", overlay, seen)
			if field.Decode != nil {
				collectVariantArtifacts(*field.Decode, "", overlay, seen)
			}
		}
	}
	for typeID := range seen {
		offenders = append(offenders, typeID)
	}
	if len(offenders) == 0 {
		return nil
	}
	sort.Strings(offenders)
	return fmt.Errorf("naming overlay required for artifact TypeIDs: %s", strings.Join(offenders, ", "))
}

func collectVariantArtifacts(node manifest.Node, owner string, overlay Overlay, offenders map[string]bool) {
	if node.TypeID != "" {
		owner = node.TypeID
	}
	for _, variant := range node.Variants {
		if LooksLikeArtifact(variant.Name) && owner != "" && overlay.Names[owner] == "" {
			offenders[owner+" (variant "+variant.Name+")"] = true
		}
		collectVariantArtifacts(variant.Encode, owner, overlay, offenders)
		if variant.Decode != nil {
			collectVariantArtifacts(*variant.Decode, owner, overlay, offenders)
		}
	}
	for _, field := range node.Fields {
		collectVariantArtifacts(field.Encode, owner, overlay, offenders)
		if field.Decode != nil {
			collectVariantArtifacts(*field.Decode, owner, overlay, offenders)
		}
	}
	for _, child := range []*manifest.Node{node.Prefix, node.Element, node.Value, node.Key, node.Control, node.Default} {
		if child != nil {
			collectVariantArtifacts(*child, owner, overlay, offenders)
		}
	}
	for _, child := range node.Elements {
		collectVariantArtifacts(child, owner, overlay, offenders)
	}
	for _, oneCase := range node.Cases {
		for _, child := range append(oneCase.Encode, oneCase.Decode...) {
			collectVariantArtifacts(child, owner, overlay, offenders)
		}
	}
}

// TypeIDs returns every TypeID occurring anywhere in a manifest node tree.
func TypeIDs(m manifest.Manifest) map[string]bool {
	result := map[string]bool{}
	for _, packet := range m.Packets {
		for _, field := range packet.Fields {
			collectFieldTypeIDs(field, result)
		}
	}
	return result
}

func collectFieldTypeIDs(field manifest.Field, result map[string]bool) {
	if field.TypeID != "" {
		result[field.TypeID] = true
	}
	collectNodeTypeIDs(field.Encode, result)
	if field.Decode != nil {
		collectNodeTypeIDs(*field.Decode, result)
	}
}

func collectNodeTypeIDs(node manifest.Node, result map[string]bool) {
	if node.TypeID != "" {
		result[node.TypeID] = true
	}
	if inferred := InferredTypeName(node); inferred != "" {
		result[inferred] = true
	}
	for _, field := range node.Fields {
		collectFieldTypeIDs(field, result)
	}
	for _, variant := range node.Variants {
		if variant.Encode.TypeID != "" {
			result[variant.Encode.TypeID] = true
		}
		collectNodeTypeIDs(variant.Encode, result)
		if variant.Decode != nil {
			collectNodeTypeIDs(*variant.Decode, result)
		}
	}
	for _, child := range []*manifest.Node{node.Prefix, node.Element, node.Value, node.Key, node.Control, node.Default} {
		if child != nil {
			collectNodeTypeIDs(*child, result)
		}
	}
	for _, child := range node.Elements {
		collectNodeTypeIDs(child, result)
	}
	for _, oneCase := range node.Cases {
		for _, child := range oneCase.Encode {
			collectNodeTypeIDs(child, result)
		}
		for _, child := range oneCase.Decode {
			collectNodeTypeIDs(child, result)
		}
	}
}
