// Package updateguide turns a human-readable protocol changelog and the target
// Mojang schema snapshot into implementation-oriented gophertunnel snippets.
package updateguide

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"maps"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"protocolgen/internal/naming"
)

// Generate reads changed schema names from changelog and renders their target
// definitions from schemasDir. The changelog remains the source of truth for
// which definitions changed; unchanged schemas are never emitted.
func Generate(changelog []byte, schemasDir string) ([]byte, error) {
	documents, err := loadDocuments(schemasDir)
	if err != nil {
		return nil, err
	}
	parsed, err := parseChangelog(string(changelog))
	if err != nil {
		return nil, err
	}
	if documents.protocolVersion == 0 {
		return nil, fmt.Errorf("target schemas do not declare x-protocol-version")
	}
	targetProtocol, err := strconv.Atoi(parsed.toProtocol)
	if err != nil {
		return nil, fmt.Errorf("changelog target protocol %q is not numeric", parsed.toProtocol)
	}
	if documents.protocolVersion != targetProtocol {
		return nil, fmt.Errorf("target schema snapshot targets protocol %d, changelog targets protocol %d", documents.protocolVersion, targetProtocol)
	}
	var output bytes.Buffer
	fmt.Fprintf(&output, "# gophertunnel update guide — protocol %s to %s\n\n", parsed.fromProtocol, parsed.toProtocol)
	if parsed.versions != "" {
		fmt.Fprintf(&output, "%s\n\n", parsed.versions)
	}
	if len(parsed.table) != 0 {
		for _, line := range parsed.table {
			fmt.Fprintln(&output, line)
		}
		fmt.Fprintln(&output)
	}
	fmt.Fprintln(&output, "Read `changelog.md` for the human-readable diff. This guide gives a target-version `Marshal` transcription for each changed definition.")
	fmt.Fprintln(&output)
	fmt.Fprintln(&output, "**The Go below is a transcription aid, not a patch.** Names come from the schema, so they may not match gophertunnel's existing names. Field comments preserve Mojang descriptions and runtime constraints; they do not implement validation.")

	emitted := map[string]bool{}
	for _, section := range parsed.sections {
		fmt.Fprintf(&output, "\n## %s\n", section.title)
		for _, item := range section.items {
			fmt.Fprintf(&output, "\n### %s\n", item.name)
			if item.summary != "" {
				fmt.Fprintf(&output, "%s\n", item.summary)
			}
			if len(item.changes) != 0 {
				fmt.Fprintln(&output)
				for _, change := range item.changes {
					fmt.Fprintln(&output, change)
				}
			}
			if len(item.affects) != 0 {
				fmt.Fprintf(&output, "\n**Also update:** %s — %s embeds this type.\n", quotedNames(item.affects), cardinality(len(item.affects)))
			}
			if section.removed {
				fmt.Fprintln(&output, "\nRemove the corresponding target definition and its registrations or references.")
				continue
			}
			key := section.kind + "\x00" + item.name
			if emitted[key] {
				fmt.Fprintln(&output, "\nUse the target definition emitted above; this section records an additional change to it.")
				continue
			}
			document, ok := documents.byTitle[item.name]
			if !ok {
				return nil, fmt.Errorf("changed schema %q is not present in %s", item.name, schemasDir)
			}
			snippet, err := renderSnippet(document, documents, section.kind, strings.HasPrefix(section.title, "New ") || strings.HasPrefix(section.title, "Added "))
			if err != nil {
				return nil, fmt.Errorf("render %s: %w", item.name, err)
			}
			snippet, err = formatSnippet(snippet)
			if err != nil {
				return nil, fmt.Errorf("format %s: %w", item.name, err)
			}
			fmt.Fprintf(&output, "\n```go\n%s\n```\n", snippet)
			emitted[key] = true
		}
	}
	return output.Bytes(), nil
}

func formatSnippet(snippet string) (string, error) {
	formatted, err := format.Source([]byte("package guide\n\n" + snippet + "\n"))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(strings.TrimPrefix(string(formatted), "package guide\n\n")), nil
}

type schemaSet struct {
	byFile          map[string]map[string]any
	byTitle         map[string]map[string]any
	byTitleFile     map[string]string
	protocolVersion int
}

func loadDocuments(directory string) (schemaSet, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return schemaSet{}, fmt.Errorf("read target schema directory: %w", err)
	}
	set := schemaSet{byFile: map[string]map[string]any{}, byTitle: map[string]map[string]any{}, byTitleFile: map[string]string{}}
	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), "__") || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(directory, entry.Name()))
		if err != nil {
			return schemaSet{}, fmt.Errorf("read target schema %s: %w", entry.Name(), err)
		}
		var document map[string]any
		if err := json.Unmarshal(data, &document); err != nil {
			return schemaSet{}, fmt.Errorf("parse target schema %s: %w", entry.Name(), err)
		}
		set.byFile[entry.Name()] = document
		if protocol, ok := document["x-protocol-version"].(float64); ok {
			if set.protocolVersion != 0 && set.protocolVersion != int(protocol) {
				return schemaSet{}, fmt.Errorf("target schemas mix protocol versions %d and %d", set.protocolVersion, int(protocol))
			}
			set.protocolVersion = int(protocol)
		}
		if err := set.addTitle(strings.TrimSuffix(entry.Name(), ".json"), entry.Name(), document); err != nil {
			return schemaSet{}, err
		}
		if title, _ := document["title"].(string); title != "" {
			if err := set.addTitle(title, entry.Name(), document); err != nil {
				return schemaSet{}, err
			}
		}
	}
	return set, nil
}

func (set *schemaSet) addTitle(key, file string, document map[string]any) error {
	if previous := set.byTitleFile[key]; previous != "" && previous != file {
		return fmt.Errorf("target schema lookup %q collides between %s and %s", key, previous, file)
	}
	set.byTitle[key] = document
	set.byTitleFile[key] = file
	return nil
}

type changelogDocument struct {
	fromProtocol string
	toProtocol   string
	versions     string
	table        []string
	sections     []changelogSection
}

type changelogSection struct {
	title   string
	kind    string
	removed bool
	items   []changelogItem
}

type changelogItem struct {
	name    string
	summary string
	changes []string
	affects []string
}

func parseChangelog(input string) (changelogDocument, error) {
	lines := strings.Split(strings.ReplaceAll(input, "\r\n", "\n"), "\n")
	document := changelogDocument{}
	for _, line := range lines {
		if strings.HasPrefix(line, "## ") {
			break
		}
		if strings.HasPrefix(line, "# ") {
			parts := strings.Fields(line)
			for index, part := range parts {
				if part == "to" && index > 0 && index+1 < len(parts) {
					document.fromProtocol = strings.Trim(parts[index-1], "` ")
					document.toProtocol = strings.Trim(parts[index+1], "` ")
				}
			}
			continue
		}
		if document.versions == "" && strings.Contains(line, "→") && !strings.HasPrefix(line, "|") {
			document.versions = strings.TrimSpace(line)
		}
		if strings.HasPrefix(line, "|") {
			document.table = append(document.table, line)
		}
	}
	if document.fromProtocol == "" || document.toProtocol == "" {
		return changelogDocument{}, fmt.Errorf("changelog heading must identify the source and target protocols")
	}

	var section *changelogSection
	var item *changelogItem
	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		switch {
		case strings.HasPrefix(line, "## "):
			title := strings.TrimPrefix(line, "## ")
			if strings.HasPrefix(title, "Renamed ") {
				return changelogDocument{}, fmt.Errorf("renamed definitions are not supported by update-guide")
			}
			kind := ""
			switch {
			case strings.Contains(title, "Packets"):
				kind = "packet"
			case strings.Contains(title, "Types"):
				kind = "type"
			case strings.Contains(title, "Enums"):
				kind = "enum"
			}
			if kind == "" {
				if startsChangeSection(title) {
					return changelogDocument{}, fmt.Errorf("unsupported changelog change section %q", title)
				}
				section, item = nil, nil
				continue
			}
			document.sections = append(document.sections, changelogSection{title: title, kind: kind, removed: strings.Contains(title, "Removed")})
			section = &document.sections[len(document.sections)-1]
			item = nil
		case strings.HasPrefix(line, "### ") && section != nil:
			section.items = append(section.items, changelogItem{name: strings.Trim(strings.TrimPrefix(line, "### "), "`")})
			item = &section.items[len(section.items)-1]
		case item != nil && strings.HasPrefix(line, "**Also affects:**"):
			item.affects = splitNames(strings.TrimSpace(strings.TrimPrefix(line, "**Also affects:**")))
		case item != nil && strings.HasPrefix(line, "- "):
			item.changes = append(item.changes, line)
		case item != nil && line != "" && item.summary == "" && !strings.HasPrefix(line, "|"):
			item.summary = line
		}
	}
	if len(document.sections) == 0 {
		return changelogDocument{}, fmt.Errorf("changelog contains no supported packet, type, or enum change sections")
	}
	for _, section := range document.sections {
		if len(section.items) == 0 {
			return changelogDocument{}, fmt.Errorf("changelog section %q contains no definitions", section.title)
		}
	}
	return document, nil
}

func startsChangeSection(title string) bool {
	for _, prefix := range []string{"New ", "Added ", "Modified ", "Removed ", "Renamed "} {
		if strings.HasPrefix(title, prefix) {
			return true
		}
	}
	return false
}

func splitNames(input string) []string {
	parts := strings.Split(input, ",")
	names := make([]string, 0, len(parts))
	for _, part := range parts {
		if name := strings.Trim(strings.TrimSpace(part), "`"); name != "" {
			names = append(names, name)
		}
	}
	return names
}

func quotedNames(names []string) string {
	quoted := make([]string, len(names))
	for index, name := range names {
		quoted[index] = "`" + name + "`"
	}
	return strings.Join(quoted, ", ")
}

func cardinality(count int) string {
	if count == 1 {
		return "it"
	}
	return "each"
}

func renderSnippet(document map[string]any, schemas schemaSet, kind string, newDefinition bool) (string, error) {
	if _, enum := document["enum"]; enum {
		return renderEnum(document)
	}
	name := schemaName(document)
	body, err := resolveRoot(document, schemas, map[string]bool{})
	if err != nil {
		return "", err
	}
	qualifier, ioType := "protocol.", "protocol.IO"
	if kind != "packet" {
		qualifier, ioType = "", "IO"
	}
	fields, err := schemaFields(body, schemas, qualifier)
	if err != nil {
		return "", err
	}
	var output strings.Builder
	if kind == "packet" && newDefinition {
		meta, ok := document["$metaProperties"].(map[string]any)
		if !ok {
			return "", fmt.Errorf("new packet has no $metaProperties")
		}
		id, err := number(meta["[cereal:packet]"])
		if err != nil || id < 0 {
			return "", fmt.Errorf("new packet has no valid cereal packet ID")
		}
		fmt.Fprintf(&output, "const ID%s uint32 = %d\n\n", name, id)
	}
	fmt.Fprintf(&output, "type %s struct {\n", name)
	for _, field := range fields {
		if field.description != "" {
			fmt.Fprintf(&output, "\t// %s %s\n", field.name, oneLine(field.description))
		}
		fmt.Fprintf(&output, "\t%s %s\n", field.name, field.goType)
	}
	fmt.Fprintln(&output, "}")
	if kind == "packet" {
		fmt.Fprintf(&output, "\nfunc (*%s) ID() uint32 {\n\treturn ID%s\n}\n", name, name)
	}
	fmt.Fprintf(&output, "\nfunc (pk *%s) Marshal(io %s) {\n", name, ioType)
	for _, field := range fields {
		fmt.Fprintf(&output, "\t%s\n", field.marshal)
	}
	fmt.Fprint(&output, "}")
	return output.String(), nil
}

func renderEnum(document map[string]any) (string, error) {
	values, ok := document["enum"].([]any)
	if !ok || len(values) == 0 {
		return "", fmt.Errorf("enum has no values")
	}
	rawExplicit, hasExplicit := document["x-enum-binary-value"]
	explicit, validExplicit := rawExplicit.([]any)
	if hasExplicit && (!validExplicit || len(explicit) != len(values)) {
		return "", fmt.Errorf("enum x-enum-binary-value must contain one value for every name")
	}
	name := schemaName(document)
	var output strings.Builder
	fmt.Fprintln(&output, "const (")
	for index, rawValue := range values {
		valueName := naming.GoExportName(fmt.Sprint(rawValue))
		value := int64(index)
		if index < len(explicit) {
			parsed, err := number(explicit[index])
			if err != nil {
				return "", fmt.Errorf("enum value %s: %w", valueName, err)
			}
			value = parsed
		}
		fmt.Fprintf(&output, "\t%s%s = %d\n", name, valueName, value)
	}
	fmt.Fprint(&output, ")")
	return output.String(), nil
}

type renderedField struct {
	ordinal     int
	name        string
	description string
	goType      string
	marshal     string
}

func schemaFields(document map[string]any, schemas schemaSet, qualifier string) ([]renderedField, error) {
	properties, ok := document["properties"].(map[string]any)
	if !ok {
		if document["type"] == "object" {
			return nil, nil
		}
		return nil, fmt.Errorf("schema is not an object")
	}
	required := map[string]bool{}
	if values, ok := document["required"].([]any); ok {
		for _, value := range values {
			required[fmt.Sprint(value)] = true
		}
	}
	fields := make([]renderedField, 0, len(properties))
	for rawName, rawSchema := range properties {
		schema, ok := rawSchema.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("field %s is not a schema", rawName)
		}
		ordinal, err := number(schema["x-ordinal-index"])
		if err != nil {
			return nil, fmt.Errorf("field %s has no ordinal", rawName)
		}
		name := naming.GoExportName(rawName)
		goType, marshal, marshaler, err := renderFieldSchema(schema, schemas, "&pk."+name, qualifier)
		if err != nil {
			return nil, fmt.Errorf("field %s: %w", rawName, err)
		}
		_, defaulted := schema["default"]
		if !required[rawName] && !defaulted {
			valueType := goType
			goType = qualifier + "Optional[" + valueType + "]"
			if strings.HasPrefix(marshal, "// ") {
				// Keep the explicit unresolved-wire warning instead of inventing an optional codec.
			} else if marshaler {
				marshal = qualifier + "OptionalMarshaler(io, &pk." + name + ")"
			} else {
				address := "&pk." + name
				funcSlicePrefix := qualifier + "FuncSlice(io, " + address + ", "
				sliceCall := qualifier + "Slice(io, " + address + ")"
				switch {
				case strings.Contains(marshal, "\n"):
					body := strings.ReplaceAll(marshal, address, "value")
					body = strings.ReplaceAll(body, strings.TrimPrefix(address, "&"), "*value")
					marshal = qualifier + "OptionalFunc(io, " + address + ", func(value *" + valueType + ") {\n" + indentLines(body, "\t") + "\n})"
				case strings.HasPrefix(marshal, funcSlicePrefix) && strings.HasSuffix(marshal, ")"):
					codec := strings.TrimSuffix(strings.TrimPrefix(marshal, funcSlicePrefix), ")")
					marshal = qualifier + "OptionalFunc(io, " + address + ", func(value *" + valueType + ") { " + qualifier + "FuncSlice(io, value, " + codec + ") })"
				case marshal == sliceCall:
					marshal = qualifier + "OptionalFunc(io, " + address + ", func(value *" + valueType + ") { " + qualifier + "Slice(io, value) })"
				default:
					marshal = strings.TrimSuffix(marshal, "("+address+")")
					marshal = qualifier + "OptionalFunc(io, " + address + ", " + marshal + ")"
				}
			}
		}
		description, _ := schema["description"].(string)
		if constraint, _ := schema["x-runtime-constraint-description"].(string); constraint != "" {
			description = strings.TrimSpace(description + " Runtime constraint: " + constraint)
		}
		fields = append(fields, renderedField{ordinal: int(ordinal), name: name, description: description, goType: goType, marshal: marshal})
	}
	sort.SliceStable(fields, func(i, j int) bool { return fields[i].ordinal < fields[j].ordinal })
	for index, field := range fields {
		if field.ordinal != index {
			if index > 0 && field.ordinal == fields[index-1].ordinal {
				return nil, fmt.Errorf("duplicate ordinal %d", field.ordinal)
			}
			return nil, fmt.Errorf("field ordinals must be contiguous from zero; got %d at position %d", field.ordinal, index)
		}
	}
	return fields, nil
}

// renderFieldSchema returns the Go type, either a complete marshal statement
// or an IO method expression for optionals, and whether the type marshals itself.
func renderFieldSchema(schema map[string]any, schemas schemaSet, address, qualifier string) (string, string, bool, error) {
	if reference, _ := schema["$ref"].(string); reference != "" {
		target, err := resolveReference(reference, schemas)
		if err != nil {
			return "", "", false, err
		}
		if _, enum := target["enum"]; enum {
			effective := maps.Clone(target)
			if options, exists := schema["x-serialization-options"]; exists {
				effective["x-serialization-options"] = options
			}
			goType, method, err := primitive(effective)
			return goType, chooseStatement(address, method), false, err
		}
		return qualifier + schemaName(target), qualifier + "Single(io, " + address + ")", true, nil
	}
	switch schema["type"] {
	case "string":
		return "string", constrainedStatement(schema, address, "io.String", "string"), false, nil
	case "boolean":
		return "bool", chooseStatement(address, "io.Bool"), false, nil
	case "integer", "number":
		goType, method, err := primitive(schema)
		if err != nil {
			return "", "", false, err
		}
		return goType, constrainedStatement(schema, address, method, "number"), false, nil
	case "array":
		if min, minOK := schema["minItems"].(float64); minOK {
			if max, maxOK := schema["maxItems"].(float64); maxOK && min == max {
				return "", "", false, fmt.Errorf("fixed-length arrays are not supported by the gophertunnel guide")
			}
		}
		if err := validateOptions(schema, map[string]bool{}); err != nil {
			return "", "", false, err
		}
		items, ok := schema["items"].(map[string]any)
		if !ok {
			return "", "", false, fmt.Errorf("array has no item schema")
		}
		if items["type"] == "array" {
			return "", "", false, fmt.Errorf("nested arrays are not supported by the gophertunnel guide")
		}
		itemType, itemMarshal, itemMarshaler, err := renderFieldSchema(items, schemas, "", qualifier)
		if err != nil {
			return "", "", false, err
		}
		if itemMarshaler {
			return "[]" + itemType, constrainedCollectionStatement(schema, address, qualifier+"Slice(io, "+address+")"), false, nil
		}
		return "[]" + itemType, constrainedCollectionStatement(schema, address, qualifier+"FuncSlice(io, "+address+", "+itemMarshal+")"), false, nil
	case "object":
		return "", "", false, fmt.Errorf("inline object fields are not supported; extract a named schema")
	default:
		if schema["type"] == nil {
			name := strings.TrimPrefix(address, "&pk.")
			return "any", "// " + name + " is untyped in the target schema; resolve its wire shape from exact-version evidence.", false, nil
		}
		return "", "", false, fmt.Errorf("unsupported schema type %q", schema["type"])
	}
}

func constrainedStatement(schema map[string]any, address, method, kind string) string {
	statement := chooseStatement(address, method)
	value := strings.TrimPrefix(address, "&")
	if address == "" {
		value = "*value"
		statement = method + "(value)"
	}
	field := constraintFieldName(value)
	var checks []string
	if kind == "string" {
		if minimum, ok := nonNegativeConstraint(schema["minLength"]); ok {
			checks = append(checks, fmt.Sprintf("if len(%s) < %d { io.InvalidValue(%s, %q, \"string too short\") }", value, minimum, value, field))
		}
		if maximum, ok := nonNegativeConstraint(schema["maxLength"]); ok {
			checks = append(checks, fmt.Sprintf("if len(%s) > %d { io.InvalidValue(%s, %q, \"string too long\") }", value, maximum, value, field))
		}
		if pattern, ok := schema["pattern"].(string); ok && pattern != "" {
			checks = append(checks, fmt.Sprintf("if matched, _ := regexp.MatchString(%q, %s); !matched { io.InvalidValue(%s, %q, \"string does not match pattern\") }", pattern, value, value, field))
		}
	} else {
		if minimum, ok := numericConstraint(schema["minimum"]); ok {
			checks = append(checks, fmt.Sprintf("if %s < %s { io.InvalidValue(%s, %q, \"value below minimum\") }", value, minimum, value, field))
		}
		if maximum, ok := numericConstraint(schema["maximum"]); ok {
			checks = append(checks, fmt.Sprintf("if %s > %s { io.InvalidValue(%s, %q, \"value above maximum\") }", value, maximum, value, field))
		}
	}
	if len(checks) == 0 {
		return chooseStatement(address, method)
	}
	if address == "" {
		return "func(value *" + constraintCallbackType(schema) + ") { " + strings.Join(append([]string{statement}, checks...), "; ") + " }"
	}
	return strings.Join(append([]string{statement}, checks...), "\n")
}

func constrainedCollectionStatement(schema map[string]any, address, statement string) string {
	value := strings.TrimPrefix(address, "&")
	field := constraintFieldName(value)
	var checks []string
	if minimum, ok := nonNegativeConstraint(schema["minItems"]); ok {
		checks = append(checks, fmt.Sprintf("if len(%s) < %d { io.InvalidValue(%s, %q, \"too few items\") }", value, minimum, value, field))
	}
	if maximum, ok := nonNegativeConstraint(schema["maxItems"]); ok {
		checks = append(checks, fmt.Sprintf("if len(%s) > %d { io.InvalidValue(%s, %q, \"too many items\") }", value, maximum, value, field))
	}
	return strings.Join(append([]string{statement}, checks...), "\n")
}

func nonNegativeConstraint(value any) (uint64, bool) {
	number, err := number(value)
	return uint64(number), err == nil && number >= 0
}

func numericConstraint(value any) (string, bool) {
	if value == nil {
		return "", false
	}
	switch value := value.(type) {
	case float64:
		return strconv.FormatFloat(value, 'g', -1, 64), true
	case json.Number:
		return value.String(), true
	case int, int64:
		return fmt.Sprint(value), true
	default:
		return "", false
	}
}

func constraintCallbackType(schema map[string]any) string {
	if schema["type"] == "string" {
		return "string"
	}
	typeName, _, err := primitive(schema)
	if err != nil {
		return "any"
	}
	return typeName
}

func constraintFieldName(value string) string {
	value = strings.TrimPrefix(value, "*")
	if index := strings.LastIndex(value, "."); index >= 0 {
		value = value[index+1:]
	}
	if value == "value" || value == "" {
		return "value"
	}
	return strings.ToLower(value[:1]) + value[1:]
}

func indentLines(value, indent string) string {
	return indent + strings.ReplaceAll(value, "\n", "\n"+indent)
}

func chooseStatement(address, method string) string {
	if address == "" {
		return method
	}
	return method + "(" + address + ")"
}

func primitive(schema map[string]any) (string, string, error) {
	underlying := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(fmt.Sprint(schema["x-underlying-type"]), "_", ""), " ", ""))
	options, err := serializationOptions(schema)
	if err != nil {
		return "", "", err
	}
	allowed := map[string]bool{"compression": true, "big endian": true}
	if err := validateOptionSet(options, allowed); err != nil {
		return "", "", err
	}
	compressed := options["compression"]
	bigEndian := options["big endian"]
	if underlying == "" || underlying == "<nil>" {
		underlying = fmt.Sprint(schema["type"])
	}
	if compressed {
		switch underlying {
		case "int", "int32", "short", "int16", "byte", "int8":
			return "int32", "io.Varint32", nil
		case "uint", "uint32", "ushort", "uint16", "ubyte", "uint8":
			return "uint32", "io.Varuint32", nil
		case "long", "int64":
			return "int64", "io.Varint64", nil
		case "ulong", "uint64":
			return "uint64", "io.Varuint64", nil
		}
	}
	mapping := map[string][2]string{
		"byte": {"int8", "io.Int8"}, "int8": {"int8", "io.Int8"},
		"ubyte": {"uint8", "io.Uint8"}, "uint8": {"uint8", "io.Uint8"},
		"short": {"int16", "io.Int16"}, "int16": {"int16", "io.Int16"},
		"ushort": {"uint16", "io.Uint16"}, "uint16": {"uint16", "io.Uint16"},
		"int": {"int32", "io.Int32"}, "int32": {"int32", "io.Int32"},
		"uint": {"uint32", "io.Uint32"}, "uint32": {"uint32", "io.Uint32"},
		"long": {"int64", "io.Int64"}, "int64": {"int64", "io.Int64"},
		"ulong": {"uint64", "io.Uint64"}, "uint64": {"uint64", "io.Uint64"},
		"float": {"float32", "io.Float32"}, "float32": {"float32", "io.Float32"},
		"double": {"float64", "io.Float64"}, "float64": {"float64", "io.Float64"},
	}
	result, ok := mapping[underlying]
	if !ok {
		return "", "", fmt.Errorf("unsupported primitive %q", underlying)
	}
	if bigEndian {
		if result[1] != "io.Int32" {
			return "", "", fmt.Errorf("unsupported big-endian primitive %q", underlying)
		}
		result[1] = "io.BEInt32"
	}
	return result[0], result[1], nil
}

func serializationOptions(schema map[string]any) (map[string]bool, error) {
	result := map[string]bool{}
	raw, exists := schema["x-serialization-options"]
	if !exists {
		return result, nil
	}
	values, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("x-serialization-options is not an array")
	}
	for _, rawValue := range values {
		value, ok := rawValue.(string)
		if !ok || strings.TrimSpace(value) == "" {
			return nil, fmt.Errorf("x-serialization-options contains a non-string value")
		}
		result[strings.ToLower(strings.TrimSpace(value))] = true
	}
	return result, nil
}

func validateOptions(schema map[string]any, allowed map[string]bool) error {
	options, err := serializationOptions(schema)
	if err != nil {
		return err
	}
	return validateOptionSet(options, allowed)
}

func validateOptionSet(options, allowed map[string]bool) error {
	for option := range options {
		if !allowed[option] {
			return fmt.Errorf("unsupported serialization option %q", option)
		}
	}
	return nil
}

func resolveRoot(document map[string]any, schemas schemaSet, active map[string]bool) (map[string]any, error) {
	reference, _ := document["$ref"].(string)
	if reference == "" {
		return document, nil
	}
	if active[reference] {
		return nil, fmt.Errorf("cyclic root reference %s", reference)
	}
	active[reference] = true
	target, err := resolveReference(reference, schemas)
	if err != nil {
		return nil, err
	}
	return resolveRoot(target, schemas, active)
}

func resolveReference(reference string, schemas schemaSet) (map[string]any, error) {
	parts := strings.SplitN(reference, "#", 2)
	if len(parts) == 2 && parts[1] != "" {
		return nil, fmt.Errorf("reference fragments are not supported: %s", reference)
	}
	file := filepath.Base(parts[0])
	target, ok := schemas.byFile[file]
	if !ok {
		return nil, fmt.Errorf("missing reference %s", reference)
	}
	return target, nil
}

func schemaName(document map[string]any) string {
	title, _ := document["title"].(string)
	return naming.GoExportName(naming.PublicTypeName(title))
}

func number(value any) (int64, error) {
	switch value := value.(type) {
	case float64:
		return int64(value), nil
	case int:
		return int64(value), nil
	case int64:
		return value, nil
	case json.Number:
		return value.Int64()
	case string:
		return strconv.ParseInt(value, 10, 64)
	default:
		return 0, fmt.Errorf("not an integer")
	}
}

func oneLine(value string) string {
	return strings.Join(strings.Fields(value), " ")
}
