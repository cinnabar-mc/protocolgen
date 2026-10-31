// Package changelog compares two corrected Mojang schema directories and
// renders a deterministic, human-readable protocol changelog.
package changelog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Config struct {
	FromDir, ToDir           string
	FromBranch, ToBranch     string
	FromUpstream, ToUpstream string
	FromFixer, ToFixer       string
}

type schema struct {
	Name       string
	Version    string                 `json:"x-minecraft-version"`
	Protocol   int                    `json:"x-protocol-version"`
	Ref        string                 `json:"$ref"`
	Properties map[string]property    `json:"properties"`
	Required   []string               `json:"required"`
	Enum       []string               `json:"enum"`
	EnumValues []interface{}          `json:"x-enum-binary-value"`
	Underlying string                 `json:"x-underlying-type"`
	Meta       map[string]interface{} `json:"$metaProperties"`
	Raw        map[string]interface{}
}

type property struct {
	Type       string                 `json:"type"`
	Ref        string                 `json:"$ref"`
	Underlying string                 `json:"x-underlying-type"`
	Ordinal    int                    `json:"x-ordinal-index"`
	Items      map[string]interface{} `json:"items"`
	Raw        map[string]interface{}
}

type set map[string]bool

type change struct {
	Name       string
	Kind       string
	PacketID   int
	Bullets    []string
	Wire       bool
	Affected   []string
	BeforeRows []field
	AfterRows  []field
}

type field struct {
	Name, Type string
	Ordinal    int
	Optional   bool
	Raw        map[string]interface{}
}

type definition struct {
	Name     string
	Kind     string
	PacketID int
}

// Generate compares the configured corrected-schema directories.
func Generate(cfg Config) ([]byte, error) {
	if cfg.FromDir == "" || cfg.ToDir == "" {
		return nil, fmt.Errorf("from and to schema directories are required")
	}
	from, fromVersion, fromProtocol, err := loadDir(cfg.FromDir)
	if err != nil {
		return nil, fmt.Errorf("load from schemas: %w", err)
	}
	to, toVersion, toProtocol, err := loadDir(cfg.ToDir)
	if err != nil {
		return nil, fmt.Errorf("load to schemas: %w", err)
	}

	fromPackets := packets(from)
	toPackets := packets(to)
	packetPayloads := set{}
	for _, all := range []map[string]*schema{fromPackets, toPackets} {
		for _, p := range all {
			packetPayloads[refName(p.Ref)] = true
		}
	}

	renames := packetRenames(fromPackets, toPackets)
	added, removed := addedRemovedDefinitions(from, to, fromPackets, toPackets, packetPayloads, renames)
	addedForGuide, removedForGuide := append([]definition{}, added...), append([]definition{}, removed...)
	reverseRenames := map[string]string{}
	for old, name := range renames {
		reverseRenames[name] = old
		addedForGuide = append(addedForGuide, definition{Name: name, Kind: "packet", PacketID: packetID(toPackets[name])})
		removedForGuide = append(removedForGuide, definition{Name: old, Kind: "packet", PacketID: packetID(fromPackets[old])})
	}
	sortDefinitions(addedForGuide)
	sortDefinitions(removedForGuide)
	var packetChanges, typeChanges, enumChanges []change
	modifiedPackets := set{}
	for name, after := range toPackets {
		before := fromPackets[name]
		if before == nil && reverseRenames[name] != "" {
			before = fromPackets[reverseRenames[name]]
		}
		if before == nil {
			continue
		}
		b := resolveRoot(before, from)
		a := resolveRoot(after, to)
		idChanged := packetID(before) != packetID(after)
		if equalWire(b.Raw, a.Raw) && !idChanged {
			continue
		}
		c := compareObjects(name, "packet", b, a)
		c.PacketID = packetID(after)
		if idChanged {
			c.Bullets = append([]string{fmt.Sprintf("packet id changed from `%d` to `%d`", packetID(before), packetID(after))}, c.Bullets...)
			c.Wire = true
		}
		if len(c.Bullets) == 0 {
			c.Bullets = []string{"schema changed"}
		}
		packetChanges = append(packetChanges, c)
		modifiedPackets[name] = true
	}
	for name, after := range to {
		before := from[name]
		if before == nil || toPackets[name] != nil || packetPayloads[name] {
			continue
		}
		if equalWire(before.Raw, after.Raw) {
			continue
		}
		var c change
		if len(after.Enum) != 0 || len(before.Enum) != 0 {
			c = compareEnums(name, before, after)
			if len(c.Bullets) == 0 {
				c.Bullets = []string{"schema changed"}
			}
			enumChanges = append(enumChanges, c)
		} else {
			c = compareObjects(name, "struct", before, after)
			if len(c.Bullets) == 0 {
				c.Bullets = []string{"schema changed"}
			}
			typeChanges = append(typeChanges, c)
		}
	}

	deps := packetDependencies(toPackets, to)
	for groupIndex := range [][]change{typeChanges, enumChanges} {
		group := &typeChanges
		if groupIndex == 1 {
			group = &enumChanges
		}
		for i := range *group {
			for packet, references := range deps {
				if references[(*group)[i].Name] {
					(*group)[i].Affected = append((*group)[i].Affected, packet)
					modifiedPackets[packet] = true
				}
			}
			sort.Strings((*group)[i].Affected)
		}
	}
	sort.Slice(packetChanges, func(i, j int) bool {
		if packetChanges[i].PacketID != packetChanges[j].PacketID {
			return packetChanges[i].PacketID < packetChanges[j].PacketID
		}
		return packetChanges[i].Name < packetChanges[j].Name
	})
	sort.Slice(typeChanges, func(i, j int) bool { return typeChanges[i].Name < typeChanges[j].Name })
	sort.Slice(enumChanges, func(i, j int) bool { return enumChanges[i].Name < enumChanges[j].Name })

	newCount, removedCount := len(added), len(removed)
	modified := len(packetChanges) + len(typeChanges) + len(enumChanges)
	wireTypes, individual := 0, 0
	for _, c := range typeChanges {
		if c.Wire {
			wireTypes++
		}
		individual += len(c.Bullets)
	}
	for _, c := range packetChanges {
		if c.Wire {
			wireTypes++
		}
	}
	for _, c := range enumChanges {
		if c.Wire {
			wireTypes++
		}
	}
	for _, c := range packetChanges {
		individual += len(c.Bullets)
	}
	for _, c := range enumChanges {
		individual += len(c.Bullets)
	}
	individual += newCount + removedCount + len(renames)
	for _, d := range added {
		if d.Kind == "packet" {
			modifiedPackets[d.Name] = true
		}
	}
	for _, d := range removed {
		if d.Kind == "packet" {
			modifiedPackets[d.Name] = true
		}
	}
	for old, name := range renames {
		modifiedPackets[old+"→"+name] = true
	}

	var out bytes.Buffer
	fmt.Fprintf(&out, "# Bedrock protocol changes — %d to %d\n\n%s → %s\n\n", fromProtocol, toProtocol, fromVersion, toVersion)
	out.WriteString("| | From | To |\n| --- | --- | --- |\n")
	fmt.Fprintf(&out, "| Branch | `%s` | `%s` |\n", cfg.FromBranch, cfg.ToBranch)
	fmt.Fprintf(&out, "| Protocol | %d | %d |\n", fromProtocol, toProtocol)
	fmt.Fprintf(&out, "| Game version | %s | %s |\n", fromVersion, toVersion)
	fmt.Fprintf(&out, "| Upstream commit | `%s` | `%s` |\n", cfg.FromUpstream, cfg.ToUpstream)
	fmt.Fprintf(&out, "| Fixer commit | `%s` | `%s` |\n\n", cfg.FromFixer, cfg.ToFixer)
	fmt.Fprintf(&out, "- **%d** new, **%d** removed, **%d** renamed, **%d** modified\n", newCount, removedCount, len(renames), modified)
	fmt.Fprintf(&out, "- **%d** type(s) change the bytes on the wire\n", wireTypes)
	fmt.Fprintf(&out, "- **%d** packet(s) touched, counting those that only embed a changed type\n", len(modifiedPackets))
	fmt.Fprintf(&out, "- **%d** individual changes\n\n", individual)
	out.WriteString("_Derived by diffing bpd-fixer's corrected schemas for the two versions. \"Wire break\" means the byte layout moves; it says nothing about whether your code needs a version gate — see the per-project guides for that._\n")
	renderDefinitionSections(&out, "Added", addedForGuide)
	renderDefinitionSections(&out, "Removed", removedForGuide)
	if len(renames) != 0 {
		out.WriteString("\n## Renames\n")
		oldNames := make([]string, 0, len(renames))
		for old := range renames {
			oldNames = append(oldNames, old)
		}
		sort.Strings(oldNames)
		for _, old := range oldNames {
			name := renames[old]
			fmt.Fprintf(&out, "\n- `%s` → `%s` (packet id `%d`)\n", old, name, packetID(toPackets[name]))
		}
	}
	if len(packetChanges) != 0 {
		out.WriteString("\n## Modified Packets\n")
		for _, c := range packetChanges {
			renderChange(&out, c)
		}
	}
	if len(typeChanges) != 0 {
		out.WriteString("\n## Modified Types\n")
		for _, c := range typeChanges {
			renderChange(&out, c)
		}
	}
	if len(enumChanges) != 0 {
		out.WriteString("\n## Modified Enums\n")
		for _, c := range enumChanges {
			renderChange(&out, c)
		}
	}
	renderUntyped(&out, to)
	return out.Bytes(), nil
}

func loadDir(dir string) (map[string]*schema, string, int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, "", 0, err
	}
	all := map[string]*schema{}
	version := ""
	protocol := 0
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" || strings.HasPrefix(entry.Name(), "__") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, "", 0, err
		}
		var raw map[string]interface{}
		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.UseNumber()
		if err := decoder.Decode(&raw); err != nil {
			return nil, "", 0, fmt.Errorf("%s: %w", entry.Name(), err)
		}
		var s schema
		decoder = json.NewDecoder(bytes.NewReader(data))
		decoder.UseNumber()
		if err := decoder.Decode(&s); err != nil {
			return nil, "", 0, fmt.Errorf("%s: %w", entry.Name(), err)
		}
		s.Name = strings.TrimSuffix(entry.Name(), ".json")
		s.Raw = raw
		if _, present := raw["x-enum-binary-value"]; present && (s.EnumValues == nil || len(s.EnumValues) != len(s.Enum)) {
			return nil, "", 0, fmt.Errorf("%s: x-enum-binary-value must contain one value for every enum name", entry.Name())
		}
		for name, p := range s.Properties {
			if node, ok := raw["properties"].(map[string]interface{})[name].(map[string]interface{}); ok {
				p.Raw = node
				s.Properties[name] = p
			}
		}
		if s.Protocol == 0 || s.Version == "" {
			return nil, "", 0, fmt.Errorf("%s: missing x-minecraft-version or x-protocol-version", entry.Name())
		}
		if protocol != 0 && protocol != s.Protocol {
			return nil, "", 0, fmt.Errorf("inconsistent protocol version: %d and %d", protocol, s.Protocol)
		}
		if version != "" && version != s.Version {
			return nil, "", 0, fmt.Errorf("inconsistent game version: %q and %q", version, s.Version)
		}
		protocol, version, all[s.Name] = s.Protocol, s.Version, &s
	}
	if len(all) == 0 {
		return nil, "", 0, fmt.Errorf("no versioned schemas found in %s", dir)
	}
	packetNamesByID := map[int]string{}
	allPackets := packets(all)
	packetNames := make([]string, 0, len(allPackets))
	for name := range allPackets {
		packetNames = append(packetNames, name)
	}
	sort.Strings(packetNames)
	for _, name := range packetNames {
		packet := allPackets[name]
		id, ok := packetIDValue(packet)
		if !ok {
			return nil, "", 0, fmt.Errorf("%s.json: [cereal:packet] must be a numeric packet id", name)
		}
		if previous := packetNamesByID[id]; previous != "" {
			return nil, "", 0, fmt.Errorf("duplicate packet id %d in %s.json and %s.json", id, previous, name)
		}
		packetNamesByID[id] = name
	}
	return all, version, protocol, nil
}

func packets(all map[string]*schema) map[string]*schema {
	out := map[string]*schema{}
	for n, s := range all {
		if _, ok := s.Meta["[cereal:packet]"]; ok {
			out[n] = s
		}
	}
	return out
}
func packetID(s *schema) int {
	id, _ := packetIDValue(s)
	return id
}
func packetIDValue(s *schema) (int, bool) {
	switch v := s.Meta["[cereal:packet]"].(type) {
	case json.Number:
		id, err := strconv.Atoi(v.String())
		return id, err == nil && id >= 0
	case float64:
		return int(v), v == float64(int(v)) && v >= 0
	default:
		return 0, false
	}
}
func resolveRoot(s *schema, all map[string]*schema) *schema {
	if target := all[refName(s.Ref)]; target != nil {
		return target
	}
	return s
}
func refName(ref string) string { return strings.TrimSuffix(filepath.Base(ref), ".json") }

func normalized(v interface{}) interface{} { return normalizeSchema(v, true) }
func normalizeSchema(v interface{}, schemaKeywords bool) interface{} {
	switch x := v.(type) {
	case map[string]interface{}:
		out := map[string]interface{}{}
		for k, child := range x {
			if schemaKeywords && k == "default" {
				// Default presence affects wire optionality; its value does not change the codec.
				out[k] = true
				continue
			}
			if schemaKeywords && (k == "x-minecraft-version" || k == "x-protocol-version" || k == "x-format-version" || k == "$schema" || k == "$id" || k == "title" || k == "description" || k == "x-runtime-constraint-description") {
				continue
			}
			if k == "properties" {
				out[k] = normalizeSchema(child, false)
			} else {
				out[k] = normalizeSchema(child, true)
			}
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(x))
		for i := range x {
			out[i] = normalizeSchema(x[i], true)
		}
		return out
	default:
		return x
	}
}
func equalWire(a, b map[string]interface{}) bool {
	aa, _ := json.Marshal(normalized(a))
	bb, _ := json.Marshal(normalized(b))
	return bytes.Equal(aa, bb)
}

func fields(s *schema) []field {
	required := set{}
	for _, n := range s.Required {
		required[n] = true
	}
	out := make([]field, 0, len(s.Properties))
	for name, p := range s.Properties {
		_, defaulted := p.Raw["default"]
		out = append(out, field{Name: name, Type: fieldType(p), Ordinal: p.Ordinal, Optional: !required[name] && !defaulted, Raw: p.Raw})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Ordinal != out[j].Ordinal {
			return out[i].Ordinal < out[j].Ordinal
		}
		return out[i].Name < out[j].Name
	})
	return out
}
func fieldType(p property) string {
	base := p.Underlying
	if base == "" {
		base = p.Type
	}
	if p.Ref != "" {
		base = refName(p.Ref)
	}
	if base == "number" {
		base = "float"
	}
	if p.Type == "array" {
		if r, ok := p.Items["$ref"].(string); ok {
			base = "array of " + refName(r)
		} else {
			itemType, _ := p.Items["x-underlying-type"].(string)
			if itemType == "" {
				itemType, _ = p.Items["type"].(string)
			}
			base = "array"
			if itemType != "" {
				base += " of " + itemType
			}
		}
	}
	if base == "" {
		base = "untyped"
	}
	if opts, ok := p.Raw["x-serialization-options"].([]interface{}); ok {
		var names []string
		for _, v := range opts {
			names = append(names, fmt.Sprint(v))
		}
		if len(names) != 0 {
			base += " (" + strings.Join(names, ", ") + ")"
		}
	}
	return base
}

func compareObjects(name, kind string, before, after *schema) change {
	b, a := fields(before), fields(after)
	c := change{Name: name, Kind: kind, BeforeRows: b, AfterRows: a}
	bm, am := map[string]field{}, map[string]field{}
	for _, f := range b {
		bm[f.Name] = f
	}
	for _, f := range a {
		am[f.Name] = f
	}
	for _, af := range a {
		bf, existed := bm[af.Name]
		if !existed {
			suffix := af.Type
			if af.Optional {
				suffix += ", optional"
			}
			c.Bullets = append(c.Bullets, fmt.Sprintf("added `%s` at ordinal %d (%s)", af.Name, af.Ordinal, suffix))
			c.Wire = true
			continue
		}
		if bf.Ordinal != af.Ordinal {
			c.Bullets = append(c.Bullets, fmt.Sprintf("`%s` moved from ordinal %d to %d", af.Name, bf.Ordinal, af.Ordinal))
			c.Wire = true
		}
		if bf.Type != af.Type {
			c.Bullets = append(c.Bullets, fmt.Sprintf("`%s` type changed from %s to %s", af.Name, bf.Type, af.Type))
			c.Wire = true
		}
		if bf.Optional != af.Optional {
			c.Bullets = append(c.Bullets, fmt.Sprintf("`%s` changed from %s to %s", af.Name, requirement(bf.Optional), requirement(af.Optional)))
			c.Wire = true
		}
		if constraints(bf.Raw) != constraints(af.Raw) {
			c.Bullets = append(c.Bullets, fmt.Sprintf("`%s` constraints changed from %s to %s", af.Name, constraints(bf.Raw), constraints(af.Raw)))
		}
	}
	for _, bf := range b {
		if _, ok := am[bf.Name]; !ok {
			c.Bullets = append(c.Bullets, fmt.Sprintf("removed `%s` from ordinal %d", bf.Name, bf.Ordinal))
			c.Wire = true
		}
	}
	return c
}
func requirement(optional bool) string {
	if optional {
		return "optional"
	}
	return "required"
}
func constraints(raw map[string]interface{}) string {
	var parts []string
	for _, k := range []string{"maximum", "minimum", "maxItems", "minItems", "maxLength", "minLength"} {
		if v, ok := raw[k]; ok {
			parts = append(parts, k+"="+number(v))
		}
	}
	return strings.Join(parts, ", ")
}
func number(v interface{}) string {
	if f, ok := v.(float64); ok {
		return strconv.FormatFloat(f, 'f', -1, 64)
	}
	return fmt.Sprint(v)
}

func compareEnums(name string, before, after *schema) change {
	c := change{Name: name, Kind: "enum"}
	b := set{}
	for _, n := range before.Enum {
		b[n] = true
	}
	a := set{}
	for _, n := range after.Enum {
		a[n] = true
	}
	for i, n := range after.Enum {
		if !b[n] {
			c.Bullets = append(c.Bullets, fmt.Sprintf("added value `%s` = %s", n, enumValue(after, i)))
		}
	}
	for i, n := range before.Enum {
		if !a[n] {
			c.Bullets = append(c.Bullets, fmt.Sprintf("removed value `%s` = %s", n, enumValue(before, i)))
		}
	}
	if len(c.Bullets) == 0 && !equalStringSlices(before.Enum, after.Enum) {
		c.Bullets = append(c.Bullets, fmt.Sprintf("value order changed from %s to %s", quotedList(before.Enum), quotedList(after.Enum)))
	}
	if before.Underlying != after.Underlying {
		c.Bullets = append(c.Bullets, fmt.Sprintf("underlying type changed from %s to %s", before.Underlying, after.Underlying))
		c.Wire = true
	}
	if len(before.EnumValues) == 0 && len(after.EnumValues) != 0 {
		c.Bullets = append(c.Bullets, "explicit binary enum values added")
	} else if len(before.EnumValues) != 0 && len(after.EnumValues) == 0 {
		c.Bullets = append(c.Bullets, "explicit binary enum values removed")
	}
	beforeIndexes := enumIndexes(before.Enum)
	for afterIndex, name := range after.Enum {
		beforeIndex, exists := beforeIndexes[name]
		if !exists {
			continue
		}
		oldValue, newValue := enumValue(before, beforeIndex), enumValue(after, afterIndex)
		if oldValue != newValue {
			c.Bullets = append(c.Bullets, fmt.Sprintf("value `%s` changed from %s to %s", name, oldValue, newValue))
			c.Wire = true
		}
	}
	return c
}

func enumValue(s *schema, index int) string {
	if index < len(s.EnumValues) {
		return number(s.EnumValues[index])
	}
	return strconv.Itoa(index)
}
func equalStringSlices(a, b []string) bool { return strings.Join(a, "\x00") == strings.Join(b, "\x00") }
func enumIndexes(values []string) map[string]int {
	indexes := make(map[string]int, len(values))
	for index, value := range values {
		indexes[value] = index
	}
	return indexes
}
func quotedList(values []string) string {
	quoted := make([]string, len(values))
	for i, value := range values {
		quoted[i] = "`" + value + "`"
	}
	return strings.Join(quoted, ", ")
}

func packetDependencies(ps, all map[string]*schema) map[string]set {
	out := map[string]set{}
	for name, p := range ps {
		seen := set{}
		walkRefs(refName(p.Ref), all, seen)
		out[name] = seen
	}
	return out
}
func walkRefs(name string, all map[string]*schema, seen set) {
	if name == "" || seen[name] {
		return
	}
	seen[name] = true
	s := all[name]
	if s == nil {
		return
	}
	walkRawRefs(s.Raw, func(ref string) { walkRefs(refName(ref), all, seen) })
}
func walkRawRefs(v interface{}, fn func(string)) {
	switch x := v.(type) {
	case map[string]interface{}:
		for k, child := range x {
			if k == "$ref" {
				if s, ok := child.(string); ok {
					fn(s)
				}
			} else {
				walkRawRefs(child, fn)
			}
		}
	case []interface{}:
		for _, child := range x {
			walkRawRefs(child, fn)
		}
	}
}

func renderChange(out *bytes.Buffer, c change) {
	fmt.Fprintf(out, "\n### %s\n", c.Name)
	if c.Kind == "packet" {
		fmt.Fprintf(out, "packet id `%d` · struct", c.PacketID)
		if c.Wire {
			out.WriteString(" · **wire break**")
		}
		out.WriteString("\n")
	} else if c.Kind == "enum" {
		out.WriteString("enum")
		if c.Wire {
			out.WriteString(" · **wire break**")
		}
		out.WriteString("\n")
	} else if c.Wire {
		out.WriteString("struct · **wire break**\n")
	} else {
		out.WriteString("struct\n")
	}
	for _, bullet := range c.Bullets {
		fmt.Fprintf(out, "\n- %s", bullet)
	}
	out.WriteString("\n")
	if c.Kind != "enum" && (len(c.BeforeRows) != 0 || len(c.AfterRows) != 0) {
		renderTable(out, c)
	}
	if len(c.Affected) != 0 {
		out.WriteString("\n**Also affects:** ")
		for i, n := range c.Affected {
			if i > 0 {
				out.WriteString(", ")
			}
			fmt.Fprintf(out, "`%s`", n)
		}
		out.WriteString("\n")
	}
}
func renderTable(out *bytes.Buffer, c change) {
	max := len(c.BeforeRows)
	if len(c.AfterRows) > max {
		max = len(c.AfterRows)
	}
	fmt.Fprintf(out, "\n| # | Before (%s) | After |  |\n| --- | --- | --- | --- |\n", c.Name)
	for i := 0; i < max; i++ {
		b, a := "—", "—"
		changed := false
		if i < len(c.BeforeRows) {
			b = fieldLabel(c.BeforeRows[i])
		}
		if i < len(c.AfterRows) {
			a = fieldLabel(c.AfterRows[i])
		}
		changed = b != a
		marker := " "
		if changed {
			marker = "**changed**"
		}
		fmt.Fprintf(out, "| %d | %s | %s | %s |\n", i, b, a, marker)
	}
}
func fieldLabel(f field) string {
	typ := f.Type
	if f.Optional {
		typ += ", optional"
	}
	return fmt.Sprintf("`%s` — %s", f.Name, typ)
}

func packetRenames(from, to map[string]*schema) map[string]string {
	byID := map[int]string{}
	for n, p := range from {
		if to[n] == nil {
			byID[packetID(p)] = n
		}
	}
	out := map[string]string{}
	for n, p := range to {
		if old := byID[packetID(p)]; old != "" && from[n] == nil {
			out[old] = n
		}
	}
	return out
}
func addedRemovedDefinitions(from, to, fromPackets, toPackets map[string]*schema, payloads set, renames map[string]string) ([]definition, []definition) {
	logical := func(all, ps map[string]*schema) map[string]definition {
		out := map[string]definition{}
		for name, s := range all {
			if payloads[name] {
				continue
			}
			kind := "type"
			if len(s.Enum) != 0 {
				kind = "enum"
			}
			out[name] = definition{Name: name, Kind: kind}
		}
		for name, packet := range ps {
			out[name] = definition{Name: name, Kind: "packet", PacketID: packetID(packet)}
		}
		return out
	}
	a, b := logical(from, fromPackets), logical(to, toPackets)
	renamedTargets := set{}
	for _, name := range renames {
		renamedTargets[name] = true
	}
	var added, removed []definition
	for name, d := range b {
		if _, existed := a[name]; !existed && !renamedTargets[name] {
			added = append(added, d)
		}
	}
	for name, d := range a {
		if _, exists := b[name]; !exists && renames[name] == "" {
			removed = append(removed, d)
		}
	}
	sortDefinitions(added)
	sortDefinitions(removed)
	return added, removed
}

func sortDefinitions(definitions []definition) {
	sort.Slice(definitions, func(i, j int) bool {
		if definitions[i].Kind != definitions[j].Kind {
			return definitions[i].Kind < definitions[j].Kind
		}
		if definitions[i].PacketID != definitions[j].PacketID {
			return definitions[i].PacketID < definitions[j].PacketID
		}
		return definitions[i].Name < definitions[j].Name
	})
}

func renderDefinitionSections(out *bytes.Buffer, verb string, definitions []definition) {
	for _, group := range []struct{ kind, heading string }{{"packet", "Packets"}, {"type", "Types"}, {"enum", "Enums"}} {
		var selected []definition
		for _, d := range definitions {
			if d.Kind == group.kind {
				selected = append(selected, d)
			}
		}
		if len(selected) == 0 {
			continue
		}
		fmt.Fprintf(out, "\n## %s %s\n", verb, group.heading)
		for _, d := range selected {
			fmt.Fprintf(out, "\n### %s\n", d.Name)
			switch d.Kind {
			case "packet":
				fmt.Fprintf(out, "packet id `%d` · struct\n", d.PacketID)
			case "enum":
				out.WriteString("enum\n")
			default:
				out.WriteString("struct\n")
			}
		}
	}
}

func renderUntyped(out *bytes.Buffer, all map[string]*schema) {
	var rows []string
	for name, s := range all {
		for field, p := range s.Properties {
			if p.Type == "" && p.Ref == "" && p.Raw["oneOf"] == nil && p.Raw["anyOf"] == nil && p.Raw["allOf"] == nil && p.Raw["const"] == nil {
				rows = append(rows, fmt.Sprintf("- `%s` · `%s`", name, field))
			}
		}
	}
	if len(rows) == 0 {
		return
	}
	sort.Strings(rows)
	out.WriteString("\n## Fields the docs leave untyped\n\nThese carry a raw NBT compound with no schema type. Listed for completeness — their contents are not tracked by this diff.\n\n")
	for _, row := range rows {
		out.WriteString(row + "\n")
	}
}
