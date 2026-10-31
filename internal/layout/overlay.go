// Package layout contains the reviewed mapping that makes the emitted Go tree
// overlay a hand-written gophertunnel checkout: where each enum's constants
// live and what they are called, and the Go name of each struct field.
package layout

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"protocolgen/internal/manifest"
	"protocolgen/internal/naming"
)

// ConstantEntry places one enum's constants in a package and file under
// reviewed names. Variants without a name keep their generated name.
type ConstantEntry struct {
	TypeID    string            `json:"type_id"`
	Package   string            `json:"package"`
	File      string            `json:"file"`
	Names     map[string]string `json:"names"`
	Rationale string            `json:"rationale"`
}

// FieldEntry renames one struct or packet field. TypeID is the owning type ID
// or packet name; Field is the wire field name.
type FieldEntry struct {
	TypeID    string `json:"type_id"`
	Field     string `json:"field"`
	Name      string `json:"name"`
	Rationale string `json:"rationale"`
}

// FileEntry places one type or packet in a file (a stem without .go) of its
// package, so generated code lands where the fork keeps the hand-written
// counterpart. TypeID is the type ID or packet name.
type FileEntry struct {
	TypeID    string `json:"type_id"`
	Package   string `json:"package"`
	File      string `json:"file"`
	Rationale string `json:"rationale"`
}

// TypeEntry gives one type or packet the fork's Go name.
type TypeEntry struct {
	TypeID    string `json:"type_id"`
	Name      string `json:"name"`
	Rationale string `json:"rationale"`
}

type Document struct {
	SchemaVersion uint32          `json:"schema_version"`
	Target        manifest.Target `json:"target"`
	Types         []TypeEntry     `json:"types"`
	Constants     []ConstantEntry `json:"constants"`
	Fields        []FieldEntry    `json:"fields"`
	Files         []FileEntry     `json:"files"`
}

// Placement is where an enum's constants are emitted.
type Placement struct {
	Package string
	File    string
	Names   map[string]string
}

type Overlay struct {
	Types     map[string]string // type ID or packet name -> Go name
	Constants map[string]Placement
	Fields    map[string]string // FieldKey -> Go name
	Files     map[string]string // type ID or packet name -> file stem
}

// TypeName returns the reviewed Go name for a type or packet, or "".
func (o Overlay) TypeName(typeID string) string {
	if o.Types == nil {
		return ""
	}
	return o.Types[typeID]
}

// File returns the reviewed file stem for a type or packet, or "".
func (o Overlay) File(typeID string) string {
	if o.Files == nil {
		return ""
	}
	return o.Files[typeID]
}

func FieldKey(typeID, field string) string { return typeID + "\x00" + field }

// FieldName returns the reviewed Go name for a field, or "" when it keeps the
// generated one.
func (o Overlay) FieldName(typeID, field string) string {
	if o.Fields == nil {
		return ""
	}
	return o.Fields[FieldKey(typeID, field)]
}

func LoadOverlay(path string, m manifest.Manifest) (Overlay, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Overlay{}, fmt.Errorf("read layout overlay: %w", err)
	}
	var document Document
	if err := json.Unmarshal(data, &document); err != nil {
		return Overlay{}, fmt.Errorf("parse layout overlay: %w", err)
	}
	if err := ValidateOverlay(m, document); err != nil {
		return Overlay{}, err
	}
	overlay := Overlay{Types: map[string]string{}, Constants: map[string]Placement{}, Fields: map[string]string{}, Files: map[string]string{}}
	for _, entry := range document.Types {
		overlay.Types[entry.TypeID] = entry.Name
	}
	for _, entry := range document.Constants {
		overlay.Constants[entry.TypeID] = Placement{Package: entry.Package, File: entry.File, Names: entry.Names}
	}
	for _, entry := range document.Fields {
		overlay.Fields[FieldKey(entry.TypeID, entry.Field)] = entry.Name
	}
	for _, entry := range document.Files {
		overlay.Files[entry.TypeID] = entry.File
	}
	return overlay, nil
}

// ValidateOverlay rejects entries that name nothing in the manifest, so a
// stale mapping fails instead of silently doing nothing.
func ValidateOverlay(m manifest.Manifest, document Document) error {
	if document.SchemaVersion != 1 {
		return fmt.Errorf("layout overlay schema_version %d is not v1", document.SchemaVersion)
	}
	if document.Target.MinecraftVersion != m.Target.MinecraftVersion || document.Target.ProtocolVersion != m.Target.ProtocolVersion {
		return fmt.Errorf("layout overlay target does not match manifest target")
	}
	enums, fields := knownEnumsAndFields(m)
	packets := map[string]bool{}
	for _, packet := range m.Packets {
		packets[packet.Name] = true
	}
	known := naming.TypeIDs(m)
	for typeID := range enums {
		known[typeID] = true
	}
	for typeID := range fields {
		known[typeID] = true
	}
	seenType := map[string]bool{}
	for _, entry := range document.Types {
		if seenType[entry.TypeID] {
			return fmt.Errorf("layout overlay repeats type name for %q", entry.TypeID)
		}
		seenType[entry.TypeID] = true
		if !known[entry.TypeID] && !packets[entry.TypeID] {
			return fmt.Errorf("layout overlay type name %q is not a manifest type or packet", entry.TypeID)
		}
		if !naming.IsExportedGoIdentifier(entry.Name) {
			return fmt.Errorf("layout overlay type %q maps to invalid identifier %q", entry.TypeID, entry.Name)
		}
	}
	seenFile := map[string]bool{}
	for _, entry := range document.Files {
		if seenFile[entry.TypeID] {
			return fmt.Errorf("layout overlay repeats file placement for %q", entry.TypeID)
		}
		seenFile[entry.TypeID] = true
		if entry.File == "" {
			return fmt.Errorf("layout overlay file placement for %q has no file", entry.TypeID)
		}
		switch entry.Package {
		case "packet":
			if !packets[entry.TypeID] {
				return fmt.Errorf("layout overlay file placement %q is not a packet", entry.TypeID)
			}
		case "protocol":
			if !known[entry.TypeID] || packets[entry.TypeID] {
				return fmt.Errorf("layout overlay file placement %q is not a manifest type", entry.TypeID)
			}
		default:
			return fmt.Errorf("layout overlay file placement %q has package %q; want protocol or packet", entry.TypeID, entry.Package)
		}
	}
	seen := map[string]bool{}
	for _, entry := range document.Constants {
		if seen[entry.TypeID] {
			return fmt.Errorf("layout overlay repeats constants for %q", entry.TypeID)
		}
		seen[entry.TypeID] = true
		variants, ok := enums[entry.TypeID]
		if !ok {
			return fmt.Errorf("layout overlay constants entry %q is not a manifest enum", entry.TypeID)
		}
		if entry.Package != "protocol" && entry.Package != "packet" {
			return fmt.Errorf("layout overlay constants entry %q has package %q; want protocol or packet", entry.TypeID, entry.Package)
		}
		if entry.File == "" {
			return fmt.Errorf("layout overlay constants entry %q has no file", entry.TypeID)
		}
		for variant, name := range entry.Names {
			if !variants[variant] {
				return fmt.Errorf("layout overlay constants entry %q names unknown variant %q", entry.TypeID, variant)
			}
			if !naming.IsExportedGoIdentifier(name) {
				return fmt.Errorf("layout overlay constants entry %q maps %q to invalid identifier %q", entry.TypeID, variant, name)
			}
		}
	}
	seenField := map[string]bool{}
	for _, entry := range document.Fields {
		key := FieldKey(entry.TypeID, entry.Field)
		if seenField[key] {
			return fmt.Errorf("layout overlay repeats field %s.%s", entry.TypeID, entry.Field)
		}
		seenField[key] = true
		owner, ok := fields[entry.TypeID]
		if !ok || !owner[entry.Field] {
			return fmt.Errorf("layout overlay field %s.%s does not exist in the manifest", entry.TypeID, entry.Field)
		}
		if !naming.IsExportedGoIdentifier(entry.Name) {
			return fmt.Errorf("layout overlay field %s.%s maps to invalid identifier %q", entry.TypeID, entry.Field, entry.Name)
		}
	}
	return nil
}

// knownEnumsAndFields indexes every enum's variant names and every owner's
// field names, keyed the way the emitters key them: enum and struct nodes by
// TypeID (or inferred name), packets by packet name.
func knownEnumsAndFields(m manifest.Manifest) (map[string]map[string]bool, map[string]map[string]bool) {
	enums := map[string]map[string]bool{}
	fields := map[string]map[string]bool{}
	var walk func(node manifest.Node)
	walk = func(node manifest.Node) {
		typeID := node.TypeID
		if typeID == "" {
			typeID = naming.InferredTypeName(node)
		}
		if typeID != "" {
			switch node.Kind {
			case manifest.KindEnum:
				// One enum can appear at several sites with different variant
				// subsets; the overlay may name any of them.
				variants := enums[typeID]
				if variants == nil {
					variants = map[string]bool{}
					enums[typeID] = variants
				}
				for _, variant := range node.Variants {
					variants[variant.Name] = true
				}
			case manifest.KindStruct:
				owner := fields[typeID]
				if owner == nil {
					owner = map[string]bool{}
					fields[typeID] = owner
				}
				for _, field := range node.Fields {
					owner[field.Name] = true
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
			for _, child := range append(oneCase.Encode, oneCase.Decode...) {
				walk(child)
			}
		}
	}
	for _, packet := range m.Packets {
		owner := map[string]bool{}
		for _, field := range packet.Fields {
			owner[field.Name] = true
			walk(field.Encode)
			if field.Decode != nil {
				walk(*field.Decode)
			}
		}
		fields[packet.Name] = owner
	}
	return enums, fields
}

// SortedDocument orders entries so a regenerated overlay diffs cleanly.
func SortedDocument(document Document) Document {
	sort.Slice(document.Types, func(i, j int) bool { return document.Types[i].TypeID < document.Types[j].TypeID })
	sort.Slice(document.Constants, func(i, j int) bool { return document.Constants[i].TypeID < document.Constants[j].TypeID })
	sort.Slice(document.Files, func(i, j int) bool { return document.Files[i].TypeID < document.Files[j].TypeID })
	sort.Slice(document.Fields, func(i, j int) bool {
		if document.Fields[i].TypeID != document.Fields[j].TypeID {
			return document.Fields[i].TypeID < document.Fields[j].TypeID
		}
		return document.Fields[i].Field < document.Fields[j].Field
	})
	return document
}
