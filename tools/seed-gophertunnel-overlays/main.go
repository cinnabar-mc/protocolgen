package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"protocolgen/internal/docs"
	"protocolgen/internal/domains"
	"protocolgen/internal/manifest"
	"protocolgen/internal/naming"
)

type sourceField struct {
	Name string
	Type string
	Doc  string
}

type sourceType struct {
	Name   string
	File   string
	Doc    string
	Fields []sourceField
}

type identityIndex struct {
	Names  map[string]string
	Fields map[string][]manifest.Field
}

func main() {
	manifestPath := flag.String("manifest", "", "canonical manifest")
	namingPath := flag.String("naming", "", "reviewed naming overlay")
	gopherPath := flag.String("gophertunnel", "", "gophertunnel checkout")
	domainsPath := flag.String("domains-out", "", "domains overlay output")
	docsPath := flag.String("docs-out", "", "docs overlay output")
	flag.Parse()
	if *manifestPath == "" || *namingPath == "" || *gopherPath == "" || *domainsPath == "" || *docsPath == "" {
		fail("-manifest, -naming, -gophertunnel, -domains-out, and -docs-out are required")
	}
	m, err := manifest.Load(*manifestPath)
	if err != nil {
		fail("load manifest: %v", err)
	}
	namingOverlay, err := naming.LoadOverlay(*namingPath, m)
	if err != nil {
		fail("load naming overlay: %v", err)
	}
	index, err := buildIdentityIndex(m, namingOverlay)
	if err != nil {
		fail("build identity index: %v", err)
	}
	protocolTypes, err := parseTypes(filepath.Join(*gopherPath, "minecraft", "protocol"), "protocol")
	if err != nil {
		fail("parse protocol sources: %v", err)
	}
	packetTypes, err := parseTypes(filepath.Join(*gopherPath, "minecraft", "protocol", "packet"), "packet")
	if err != nil {
		fail("parse packet sources: %v", err)
	}
	if err := writeDomains(*domainsPath, m, index, protocolTypes); err != nil {
		fail("write domains overlay: %v", err)
	}
	docOverlay, report := portDocs(m, index, protocolTypes, packetTypes)
	if err := writeDocs(*docsPath, m, docOverlay); err != nil {
		fail("write docs overlay: %v", err)
	}
	fmt.Printf("ported type docs: %d; unmatched documented source types: %d\n", report.typeDocs, report.unmatchedTypeDocs)
	fmt.Printf("ported field docs: %d; unmatched documented source fields: %d; positional matches skipped: %d\n", report.fieldDocs, report.unmatchedFieldDocs, report.positionalSkipped)
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}

func buildIdentityIndex(m manifest.Manifest, overlay naming.Overlay) (identityIndex, error) {
	resolver := naming.NewResolver(overlay)
	for _, packet := range m.Packets {
		if err := resolver.Reserve(packet.Name, naming.PacketTypeName(packet.Name), goName); err != nil {
			return identityIndex{}, err
		}
	}
	index := identityIndex{Names: map[string]string{}, Fields: map[string][]manifest.Field{}}
	var walkNode func(manifest.Node) error
	var walkField func(manifest.Field) error
	walkField = func(field manifest.Field) error {
		if err := walkNode(field.Encode); err != nil {
			return err
		}
		if field.Decode != nil {
			return walkNode(*field.Decode)
		}
		return nil
	}
	walkNode = func(node manifest.Node) error {
		typeID := node.TypeID
		if typeID == "" {
			typeID = naming.InferredTypeName(node)
		}
		if typeID != "" {
			name, err := resolver.Resolve(node, "", goName)
			if err != nil {
				return err
			}
			index.Names[typeID] = name
			if len(node.Fields) != 0 {
				index.Fields[typeID] = node.Fields
			}
		}
		for _, field := range node.Fields {
			if err := walkField(field); err != nil {
				return err
			}
		}
		for _, variant := range node.Variants {
			if err := walkNode(variant.Encode); err != nil {
				return err
			}
			if variant.Decode != nil {
				if err := walkNode(*variant.Decode); err != nil {
					return err
				}
			}
		}
		for _, child := range []*manifest.Node{node.Prefix, node.Element, node.Value, node.Key, node.Control, node.Default} {
			if child != nil {
				if err := walkNode(*child); err != nil {
					return err
				}
			}
		}
		for _, child := range node.Elements {
			if err := walkNode(child); err != nil {
				return err
			}
		}
		for _, oneCase := range node.Cases {
			for _, child := range append(oneCase.Encode, oneCase.Decode...) {
				if err := walkNode(child); err != nil {
					return err
				}
			}
		}
		return nil
	}
	for _, packet := range m.Packets {
		index.Fields[packet.Name] = packet.Fields
		for _, field := range packet.Fields {
			if err := walkField(field); err != nil {
				return identityIndex{}, err
			}
		}
	}
	for typeID := range naming.TypeIDs(m) {
		if index.Names[typeID] == "" {
			neutral := overlay.Names[typeID]
			if neutral == "" {
				neutral = naming.PublicTypeName(typeID)
			}
			index.Names[typeID] = goName(neutral)
		}
	}
	return index, nil
}

func parseTypes(directory, packageName string) (map[string]sourceType, error) {
	paths, err := filepath.Glob(filepath.Join(directory, "*.go"))
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	result := map[string]sourceType{}
	sourceFiles := 0
	fset := token.NewFileSet()
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		if file.Name.Name != packageName {
			continue
		}
		sourceFiles++
		for _, declaration := range file.Decls {
			gen, ok := declaration.(*ast.GenDecl)
			if !ok || gen.Tok.String() != "type" {
				continue
			}
			for _, spec := range gen.Specs {
				typeSpec := spec.(*ast.TypeSpec)
				doc := commentText(typeSpec.Doc)
				if doc == "" {
					doc = commentText(gen.Doc)
				}
				item := sourceType{Name: typeSpec.Name.Name, File: strings.TrimSuffix(filepath.Base(path), ".go"), Doc: doc}
				if structure, ok := typeSpec.Type.(*ast.StructType); ok {
					item.Fields = sourceFields(fset, structure)
				}
				if _, exists := result[item.Name]; !exists {
					result[item.Name] = item
				}
			}
		}
	}
	if sourceFiles == 0 {
		return nil, fmt.Errorf("no non-test %s source files in %s", packageName, directory)
	}
	return result, nil
}

func sourceFields(fset *token.FileSet, structure *ast.StructType) []sourceField {
	var result []sourceField
	for _, field := range structure.Fields.List {
		doc := commentText(field.Doc)
		if doc == "" {
			doc = commentText(field.Comment)
		}
		var typ strings.Builder
		_ = format.Node(&typ, fset, field.Type)
		for _, name := range field.Names {
			result = append(result, sourceField{Name: name.Name, Type: typ.String(), Doc: doc})
		}
	}
	return result
}

func commentText(group *ast.CommentGroup) string {
	if group == nil {
		return ""
	}
	return strings.TrimSpace(group.Text())
}

func writeDomains(path string, m manifest.Manifest, index identityIndex, sourceTypes map[string]sourceType) error {
	ids := make([]string, 0, len(index.Names))
	for typeID := range naming.TypeIDs(m) {
		ids = append(ids, typeID)
	}
	sort.Strings(ids)
	document := domains.Document{SchemaVersion: 1, Target: m.Target}
	for _, typeID := range ids {
		name := index.Names[typeID]
		stem, direct := domainForName(name, sourceTypes)
		rationale := "Grouped by the generated type-family prefix; no direct gophertunnel counterpart."
		if direct {
			rationale = "Matches the gophertunnel protocol domain file layout."
		}
		document.Entries = append(document.Entries, domains.Entry{TypeID: typeID, Domain: stem, Rationale: rationale})
	}
	return writeJSON(path, document)
}

type portReport struct {
	typeDocs, unmatchedTypeDocs   int
	fieldDocs, unmatchedFieldDocs int
	positionalSkipped             int
}

func portDocs(m manifest.Manifest, index identityIndex, protocolTypes, packetTypes map[string]sourceType) (docs.Overlay, portReport) {
	overlay := docs.Overlay{Types: map[string]string{}, Fields: map[string]string{}}
	var report portReport
	matchedProtocol := map[string]bool{}
	matchedPackets := map[string]bool{}
	for typeID, name := range index.Names {
		source, ok := sourceTypeForName(name, protocolTypes)
		if !ok {
			continue
		}
		matchedProtocol[source.Name] = true
		if source.Doc != "" {
			overlay.Types[typeID] = source.Doc
			report.typeDocs++
		}
		ported, unmatched, skipped := portFieldDocs(typeID, index.Fields[typeID], source, overlay)
		report.fieldDocs += ported
		report.unmatchedFieldDocs += unmatched
		report.positionalSkipped += skipped
	}
	for _, packet := range m.Packets {
		name := goName(naming.PacketTypeName(packet.Name))
		source, ok := packetTypes[name]
		if !ok {
			continue
		}
		matchedPackets[source.Name] = true
		if source.Doc != "" {
			overlay.Types[packet.Name] = source.Doc
			report.typeDocs++
		}
		ported, unmatched, skipped := portFieldDocs(packet.Name, packet.Fields, source, overlay)
		report.fieldDocs += ported
		report.unmatchedFieldDocs += unmatched
		report.positionalSkipped += skipped
	}
	for name, source := range protocolTypes {
		if source.Doc != "" && !matchedProtocol[name] {
			report.unmatchedTypeDocs++
		}
		if !matchedProtocol[name] {
			report.unmatchedFieldDocs += documentedFieldCount(source)
		}
	}
	for name, source := range packetTypes {
		if source.Doc != "" && !matchedPackets[name] {
			report.unmatchedTypeDocs++
		}
		if !matchedPackets[name] {
			report.unmatchedFieldDocs += documentedFieldCount(source)
		}
	}
	return overlay, report
}

func documentedFieldCount(source sourceType) int {
	count := 0
	for _, field := range source.Fields {
		if field.Doc != "" {
			count++
		}
	}
	return count
}

func portFieldDocs(owner string, fields []manifest.Field, source sourceType, overlay docs.Overlay) (ported, unmatched, positionalSkipped int) {
	if len(fields) == 0 || len(source.Fields) == 0 {
		return 0, 0, 0
	}
	used := map[string]bool{}
	for _, sourceField := range source.Fields {
		if sourceField.Doc == "" {
			continue
		}
		matches := []int{}
		for index, field := range fields {
			if !used[field.Name] && normalize(sourceField.Name) == normalize(field.Name) {
				matches = append(matches, index)
			}
		}
		if len(matches) == 1 {
			field := fields[matches[0]]
			overlay.Fields[docs.FieldKey(owner, field.Name)] = docs.LeadWith(sourceField.Doc, sourceField.Name, naming.GoExportName(field.Name))
			used[field.Name] = true
			ported++
			continue
		}
		unmatched++
	}
	if unmatched == 0 || len(fields) != len(source.Fields) || !fieldTypesAlign(fields, source.Fields) {
		if unmatched != 0 {
			positionalSkipped++
		}
		return
	}
	for index, sourceField := range source.Fields {
		if sourceField.Doc == "" {
			continue
		}
		field := fields[index]
		key := docs.FieldKey(owner, field.Name)
		if _, exists := overlay.Fields[key]; exists {
			continue
		}
		overlay.Fields[key] = docs.LeadWith(sourceField.Doc, sourceField.Name, naming.GoExportName(field.Name))
		ported++
	}
	return ported, 0, 0
}

func fieldTypesAlign(fields []manifest.Field, source []sourceField) bool {
	for index := range fields {
		if typeCategory(fields[index].Encode) != sourceTypeCategory(source[index].Type) {
			return false
		}
	}
	return true
}

func typeCategory(node manifest.Node) string {
	switch node.Kind {
	case manifest.KindPrimitive:
		if node.Primitive == nil {
			return "unknown"
		}
		switch node.Primitive.Code {
		case "bool":
			return "bool"
		case "f32le", "f32be":
			return "float32"
		case "f64le", "f64be":
			return "float64"
		case "i8", "i16le", "i16be", "i32le", "i32be", "i64le", "i64be", "var_i32", "var_i64", "zigzag_i32", "zigzag_i64":
			return "signed"
		default:
			return "unsigned"
		}
	case manifest.KindString:
		return "string"
	case manifest.KindBytes, manifest.KindArray:
		return "slice"
	case manifest.KindFixedArray:
		return "array"
	case manifest.KindOptional:
		return "optional"
	case manifest.KindMap:
		return "map"
	default:
		return "named"
	}
}

func sourceTypeCategory(value string) string {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "[]") {
		return "slice"
	}
	if strings.HasPrefix(value, "[") {
		return "array"
	}
	if strings.HasPrefix(value, "map[") {
		return "map"
	}
	if strings.HasPrefix(value, "*") {
		return "named"
	}
	if strings.Contains(value, "Optional[") {
		return "optional"
	}
	switch value {
	case "bool":
		return "bool"
	case "float32":
		return "float32"
	case "float64":
		return "float64"
	case "int8", "int16", "int32", "int64", "int":
		return "signed"
	case "uint8", "uint16", "uint32", "uint64", "uint":
		return "unsigned"
	case "string":
		return "string"
	default:
		return "named"
	}
}

func writeDocs(path string, m manifest.Manifest, overlay docs.Overlay) error {
	document := docs.Document{SchemaVersion: 1, Target: m.Target, Entries: docs.SortedEntries(overlay)}
	return writeJSON(path, document)
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func domainForName(name string, sourceTypes map[string]sourceType) (string, bool) {
	if source, ok := sourceTypes[name]; ok {
		return source.File, true
	}
	normalized := normalize(name)
	for sourceName, source := range sourceTypes {
		if normalized == normalize(sourceName+"Data") || normalized == normalize(sourceName+"Info") || normalized == normalize(sourceName+"Type") || normalize(sourceName) == normalize(name+"Data") {
			return source.File, true
		}
	}
	snake := snake(name)
	for _, prefix := range []string{
		"bedrock_ddui", "bedrock_profile", "bedrock_safety", "sync_world_clocks", "attribute_layer", "available_commands", "legacy_telemetry", "telemetry", "item_stack", "stack_request", "resource_pack", "server_join", "player", "biome", "camera", "command", "inventory", "recipe", "shape", "entity", "sub_chunk", "scoreboard", "text", "memory", "structure", "waypoint", "world", "block", "chunk", "education", "enchant", "creative", "data_store", "trim", "skin", "sound", "voxel", "actor", "attribute", "container", "experiment", "game_rule", "map", "npc", "pack", "position_tracking", "settings", "item", "events",
	} {
		if strings.HasPrefix(snake, prefix+"_") || snake == prefix {
			if prefix == "legacy_telemetry" || prefix == "telemetry" {
				return "telemetry", false
			}
			if prefix == "stack_request" {
				return "item_stack", false
			}
			return prefix, false
		}
	}
	return "misc", false
}

func sourceTypeForName(name string, sourceTypes map[string]sourceType) (sourceType, bool) {
	if source, ok := sourceTypes[name]; ok {
		return source, true
	}
	normalized := normalize(name)
	for sourceName, source := range sourceTypes {
		if normalized == normalize(sourceName+"Data") || normalized == normalize(sourceName+"Info") || normalized == normalize(sourceName+"Type") || normalize(sourceName) == normalize(name+"Data") {
			return source, true
		}
	}
	return sourceType{}, false
}

func normalize(value string) string {
	var b strings.Builder
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
		}
	}
	return b.String()
}

func snake(value string) string {
	var b strings.Builder
	for index, r := range value {
		if unicode.IsUpper(r) && index > 0 {
			b.WriteByte('_')
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteByte('_')
		}
	}
	return strings.Trim(b.String(), "_")
}

func goName(value string) string {
	var b strings.Builder
	upper := true
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if upper {
				r = unicode.ToUpper(r)
				upper = false
			}
			b.WriteRune(r)
		} else {
			upper = true
		}
	}
	name := b.String()
	if name == "" {
		return "Generated"
	}
	if unicode.IsDigit([]rune(name)[0]) {
		name = "Generated" + name
	}
	initialisms := map[string]string{"id": "ID", "uuid": "UUID", "ui": "UI", "url": "URL", "uri": "URI", "json": "JSON", "nbt": "NBT", "rgb": "RGB", "rgba": "RGBA", "api": "API", "uid": "UID", "molang": "MoLang"}
	for lower, replacement := range initialisms {
		name = replaceInitialism(name, lower, replacement)
	}
	return name
}

func replaceInitialism(value, lower, replacement string) string {
	if strings.EqualFold(value, lower) {
		return replacement
	}
	if strings.HasSuffix(strings.ToLower(value), lower) {
		return value[:len(value)-len(lower)] + replacement
	}
	return value
}
