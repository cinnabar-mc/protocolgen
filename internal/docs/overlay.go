// Package docs contains the reviewed documentation overlay shared by the Go
// and Rust emitters.
package docs

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"protocolgen/internal/manifest"
	"protocolgen/internal/naming"
)

const fieldKeySeparator = "\x00"

type Entry struct {
	TypeID string `json:"type_id"`
	Field  string `json:"field,omitempty"`
	Doc    string `json:"doc"`
}

type Document struct {
	SchemaVersion uint32          `json:"schema_version"`
	Target        manifest.Target `json:"target"`
	Entries       []Entry         `json:"entries"`
}

type Overlay struct {
	Types  map[string]string
	Fields map[string]string
}

type Coverage struct {
	TypesDocumented  int
	TypesTotal       int
	FieldsDocumented int
	FieldsTotal      int
}

func LoadOverlay(path string, m manifest.Manifest) (Overlay, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Overlay{}, fmt.Errorf("read docs overlay: %w", err)
	}
	var document Document
	if err := json.Unmarshal(data, &document); err != nil {
		return Overlay{}, fmt.Errorf("parse docs overlay: %w", err)
	}
	if err := ValidateOverlay(m, document); err != nil {
		return Overlay{}, err
	}
	overlay := Overlay{Types: map[string]string{}, Fields: map[string]string{}}
	for _, entry := range document.Entries {
		if entry.Field == "" {
			overlay.Types[entry.TypeID] = entry.Doc
		} else {
			overlay.Fields[FieldKey(entry.TypeID, entry.Field)] = entry.Doc
		}
	}
	return overlay, nil
}

func ValidateOverlay(m manifest.Manifest, document Document) error {
	if document.SchemaVersion != 1 {
		return fmt.Errorf("docs overlay schema_version %d is not v1", document.SchemaVersion)
	}
	if document.Target.MinecraftVersion != m.Target.MinecraftVersion || document.Target.ProtocolVersion != m.Target.ProtocolVersion {
		return fmt.Errorf("docs overlay target does not match manifest target")
	}
	types, fields := knownTypesAndFields(m)
	seen := map[string]bool{}
	for index, entry := range document.Entries {
		if entry.TypeID == "" || strings.TrimSpace(entry.Doc) == "" {
			return fmt.Errorf("docs overlay entry[%d] requires type_id and doc", index)
		}
		if !types[entry.TypeID] {
			return fmt.Errorf("docs overlay entry[%d] refers to stale TypeID %q", index, entry.TypeID)
		}
		key := FieldKey(entry.TypeID, entry.Field)
		if seen[key] {
			return fmt.Errorf("docs overlay contains duplicate key %q", key)
		}
		seen[key] = true
		if entry.Field != "" && !fields[entry.TypeID][entry.Field] {
			return fmt.Errorf("docs overlay entry[%d] refers to stale field %q on %q", index, entry.Field, entry.TypeID)
		}
	}
	return nil
}

func FieldKey(typeID, field string) string {
	return typeID + fieldKeySeparator + field
}

func (o Overlay) Type(typeID string) string {
	return o.Types[typeID]
}

func (o Overlay) Field(typeID, field string) string {
	return o.Fields[FieldKey(typeID, field)]
}

func CoverageOf(m manifest.Manifest, overlay Overlay) Coverage {
	types, fields := knownTypesAndFields(m)
	coverage := Coverage{TypesTotal: len(types)}
	for typeID := range types {
		if strings.TrimSpace(overlay.Type(typeID)) != "" {
			coverage.TypesDocumented++
		}
	}
	for typeID, names := range fields {
		for field := range names {
			coverage.FieldsTotal++
			if strings.TrimSpace(overlay.Field(typeID, field)) != "" {
				coverage.FieldsDocumented++
			}
		}
	}
	return coverage
}

func GoComments(text string) []string {
	return comments(text, "// ")
}

func RustComments(text string) []string {
	return comments(text, "/// ")
}

// LeadWith replaces a leading token exactly equal to from with to, so docs
// ported from a source with different member names lead with the local one.
// Exact matching keeps ordinary English openers untouched.
func LeadWith(text, from, to string) string {
	text = strings.TrimSpace(text)
	if text == "" || from == "" || from == to {
		return text
	}
	rest, ok := strings.CutPrefix(text, from)
	if !ok || (rest != "" && rest[0] != ' ' && rest[0] != ',' && rest[0] != '.' && rest[0] != '\'') {
		return text
	}
	return to + rest
}

func comments(text, prefix string) []string {
	// gophertunnel wraps its doc comments at 110 columns; matching it keeps
	// ported docs byte-identical.
	const width = 110
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	var result []string
	for _, paragraph := range strings.Split(text, "\n\n") {
		paragraph = strings.TrimSpace(strings.ReplaceAll(paragraph, "\n", " "))
		if paragraph == "" {
			continue
		}
		if len(result) > 0 {
			result = append(result, strings.TrimRight(prefix, " "))
		}
		words := strings.Fields(paragraph)
		line := prefix
		for _, word := range words {
			if len(line)+len(word)+1 > width && strings.TrimSpace(line) != "" {
				result = append(result, strings.TrimRight(line, " "))
				line = prefix + word
				continue
			}
			if line != prefix {
				line += " "
			}
			line += word
		}
		result = append(result, strings.TrimRight(line, " "))
	}
	return result
}

func knownTypesAndFields(m manifest.Manifest) (map[string]bool, map[string]map[string]bool) {
	types := naming.TypeIDs(m)
	for _, packet := range m.Packets {
		types[packet.Name] = true
	}
	fields := map[string]map[string]bool{}
	for _, packet := range m.Packets {
		addFields(fields, packet.Name, packet.Fields)
		for _, field := range packet.Fields {
			walkNodeFields(field.Encode, fields)
			if field.Decode != nil {
				walkNodeFields(*field.Decode, fields)
			}
		}
	}
	return types, fields
}

func walkNodeFields(node manifest.Node, fields map[string]map[string]bool) {
	typeID := node.TypeID
	if typeID == "" {
		typeID = naming.InferredTypeName(node)
	}
	if typeID != "" {
		addFields(fields, typeID, node.Fields)
	}
	for _, field := range node.Fields {
		walkNodeFields(field.Encode, fields)
		if field.Decode != nil {
			walkNodeFields(*field.Decode, fields)
		}
	}
	for _, variant := range node.Variants {
		walkNodeFields(variant.Encode, fields)
		if variant.Decode != nil {
			walkNodeFields(*variant.Decode, fields)
		}
	}
	for _, child := range []*manifest.Node{node.Prefix, node.Element, node.Value, node.Key, node.Control, node.Default} {
		if child != nil {
			walkNodeFields(*child, fields)
		}
	}
	for _, child := range node.Elements {
		walkNodeFields(child, fields)
	}
	for _, oneCase := range node.Cases {
		for _, child := range oneCase.Encode {
			walkNodeFields(child, fields)
		}
		for _, child := range oneCase.Decode {
			walkNodeFields(child, fields)
		}
	}
}

func addFields(fields map[string]map[string]bool, typeID string, values []manifest.Field) {
	if fields[typeID] == nil {
		fields[typeID] = map[string]bool{}
	}
	for _, field := range values {
		fields[typeID][field.Name] = true
	}
}

func SortedEntries(overlay Overlay) []Entry {
	entries := make([]Entry, 0, len(overlay.Types)+len(overlay.Fields))
	for typeID, doc := range overlay.Types {
		entries = append(entries, Entry{TypeID: typeID, Doc: doc})
	}
	for key, doc := range overlay.Fields {
		parts := strings.SplitN(key, fieldKeySeparator, 2)
		entries = append(entries, Entry{TypeID: parts[0], Field: parts[1], Doc: doc})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].TypeID != entries[j].TypeID {
			return entries[i].TypeID < entries[j].TypeID
		}
		return entries[i].Field < entries[j].Field
	})
	return entries
}
