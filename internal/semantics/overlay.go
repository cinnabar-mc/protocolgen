// Package semantics contains the reviewed list of integer fields that carry
// an actor identifier. The sources describe many of them as plain integers;
// marking them lets the emitters use the identifier IO operations so a
// consumer can translate every actor ID in one place.
package semantics

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"protocolgen/internal/manifest"
	"protocolgen/internal/naming"
)

const (
	ActorUniqueID  = "ActorUniqueID"
	ActorRuntimeID = "ActorRuntimeID"
)

// Entry marks one field. TypeID is the owning type ID or packet name; Field
// is the wire field name.
type Entry struct {
	TypeID    string `json:"type_id"`
	Field     string `json:"field"`
	Semantic  string `json:"semantic"`
	Rationale string `json:"rationale"`
}

type Document struct {
	SchemaVersion uint32          `json:"schema_version"`
	Target        manifest.Target `json:"target"`
	Entries       []Entry         `json:"entries"`
}

type Overlay struct {
	Fields map[string]string // FieldKey -> semantic
}

func FieldKey(typeID, field string) string { return typeID + "\x00" + field }

// Semantic returns the reviewed identifier semantic of a field, or "".
func (o Overlay) Semantic(typeID, field string) string {
	if o.Fields == nil {
		return ""
	}
	return o.Fields[FieldKey(typeID, field)]
}

// Apply wraps a plain integer node in the canonical identifier struct the
// emitters already map to identifier IO operations. Nodes that are not plain
// integers, or already identifiers, are returned unchanged.
func (o Overlay) Apply(typeID string, field manifest.Field) manifest.Node {
	semantic := o.Semantic(typeID, field.Name)
	node := field.Encode
	if semantic == "" || node.Kind != manifest.KindPrimitive || node.Primitive == nil || !Carries(semantic, node.Primitive.Code) {
		return node
	}
	return manifest.Node{
		Kind:   manifest.KindStruct,
		TypeID: semantic,
		Fields: []manifest.Field{{Ordinal: 0, Name: "ID", Encode: node, Symmetry: manifest.Symmetric}},
	}
}

// Carries reports whether a wire encoding can carry the identifier.
func Carries(semantic, code string) bool {
	switch semantic {
	case ActorUniqueID:
		return code == "zigzag_i64" || code == "i64le" || code == "u64le" || code == "var_u64" || code == "var_i64"
	case ActorRuntimeID:
		return code == "var_u64" || code == "zigzag_i64" || code == "var_u32" || code == "var_i64"
	}
	return false
}

func LoadOverlay(path string, m manifest.Manifest) (Overlay, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Overlay{}, fmt.Errorf("read semantics overlay: %w", err)
	}
	var document Document
	if err := json.Unmarshal(data, &document); err != nil {
		return Overlay{}, fmt.Errorf("parse semantics overlay: %w", err)
	}
	if err := ValidateOverlay(m, document); err != nil {
		return Overlay{}, err
	}
	overlay := Overlay{Fields: make(map[string]string, len(document.Entries))}
	for _, entry := range document.Entries {
		overlay.Fields[FieldKey(entry.TypeID, entry.Field)] = entry.Semantic
	}
	return overlay, nil
}

// ValidateOverlay rejects entries whose field does not exist or whose wire
// encoding cannot carry the identifier, so a stale entry fails instead of
// silently doing nothing.
func ValidateOverlay(m manifest.Manifest, document Document) error {
	if document.SchemaVersion != 1 {
		return fmt.Errorf("semantics overlay schema_version %d is not v1", document.SchemaVersion)
	}
	if document.Target.MinecraftVersion != m.Target.MinecraftVersion || document.Target.ProtocolVersion != m.Target.ProtocolVersion {
		return fmt.Errorf("semantics overlay target does not match manifest target")
	}
	fields := knownFields(m)
	seen := map[string]bool{}
	for _, entry := range document.Entries {
		key := FieldKey(entry.TypeID, entry.Field)
		if seen[key] {
			return fmt.Errorf("semantics overlay repeats %s.%s", entry.TypeID, entry.Field)
		}
		seen[key] = true
		nodes, ok := fields[key]
		if !ok {
			return fmt.Errorf("semantics overlay field %s.%s does not exist in the manifest", entry.TypeID, entry.Field)
		}
		for _, node := range nodes {
			if node.Kind != manifest.KindPrimitive || node.Primitive == nil || !Carries(entry.Semantic, node.Primitive.Code) {
				return fmt.Errorf("semantics overlay field %s.%s cannot carry %s", entry.TypeID, entry.Field, entry.Semantic)
			}
		}
	}
	return nil
}

// knownFields indexes every encode and decode occurrence without changing the manifest.
func knownFields(m manifest.Manifest) map[string][]manifest.Node {
	fields := map[string][]manifest.Node{}
	add := func(owner string, field manifest.Field) {
		key := FieldKey(owner, field.Name)
		fields[key] = append(fields[key], field.Encode)
		if field.Decode != nil {
			fields[key] = append(fields[key], *field.Decode)
		}
	}
	var walk func(node manifest.Node)
	walk = func(node manifest.Node) {
		if node.Kind == manifest.KindStruct {
			typeID := node.TypeID
			if typeID == "" {
				typeID = naming.InferredTypeName(node)
			}
			if typeID != "" {
				for _, field := range node.Fields {
					add(typeID, field)
				}
			}
		}
		for _, field := range node.Fields {
			walk(field.Encode)
			if field.Decode != nil {
				walk(*field.Decode)
			}
		}
		for _, variant := range node.Variants {
			walk(variant.Encode)
			if variant.Decode != nil {
				walk(*variant.Decode)
			}
		}
		for _, child := range []*manifest.Node{node.Prefix, node.Element, node.Value, node.Key, node.Control, node.Default} {
			if child != nil {
				walk(*child)
			}
		}
		for _, child := range node.Elements {
			walk(child)
		}
		for _, oneCase := range node.Cases {
			for _, child := range oneCase.Encode {
				walk(child)
			}
			for _, child := range oneCase.Decode {
				walk(child)
			}
		}
	}
	for _, packet := range m.Packets {
		for _, field := range packet.Fields {
			add(packet.Name, field)
			walk(field.Encode)
			if field.Decode != nil {
				walk(*field.Decode)
			}
		}
	}
	return fields
}

// SortedDocument orders entries so a regenerated overlay diffs cleanly.
func SortedDocument(document Document) Document {
	sort.Slice(document.Entries, func(i, j int) bool {
		if document.Entries[i].TypeID != document.Entries[j].TypeID {
			return document.Entries[i].TypeID < document.Entries[j].TypeID
		}
		return document.Entries[i].Field < document.Entries[j].Field
	})
	return document
}
