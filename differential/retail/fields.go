package retail

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	genprotocol "protocolgen/generated/1.26.51/go/protocol"
)

// The Rust and Go codecs come from one manifest, so a decoded entry must hold
// the same values in both. Neither renders the other's syntax, so both sides
// become a value tree: type names, newtype wrappers, boxes and pointers drop
// out, an absent optional is none, and leaves keep the text or value each
// language printed. Rust enum variants resolve through the generated crate's
// own From<raw> tables.

type nodeKind int

const (
	nodeNone   nodeKind = iota
	nodeStruct          // fields in declaration order
	nodeList
	nodeLeaf   // a Rust token (number, identifier, uuid) or a Go scalar
	nodeString // decoded string bytes
	nodeBytes
)

type node struct {
	kind   nodeKind
	names  []string // struct field names
	items  []*node  // struct fields or list elements
	text   string   // leaf token, string or bytes
	goVal  reflect.Value
	goEnum string // Go named integer type, for resolving Rust variants
}

// ParseRustDebug parses the compact `{:?}` rendering of a generated Rust
// packet into a value tree.
func ParseRustDebug(text string) (*node, error) {
	p := &debugParser{s: text}
	n, err := p.value()
	if err != nil {
		return nil, fmt.Errorf("at byte %d: %w", p.i, err)
	}
	p.space()
	if p.i != len(p.s) {
		return nil, fmt.Errorf("trailing text at byte %d", p.i)
	}
	return n, nil
}

type debugParser struct {
	s string
	i int
}

func (p *debugParser) space() {
	for p.i < len(p.s) && p.s[p.i] == ' ' {
		p.i++
	}
}

func (p *debugParser) eat(c byte) bool {
	p.space()
	if p.i < len(p.s) && p.s[p.i] == c {
		p.i++
		return true
	}
	return false
}

func isToken(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '.' || c == '-' || c == '+'
}

func (p *debugParser) value() (*node, error) {
	p.space()
	if p.i >= len(p.s) {
		return nil, errors.New("unexpected end")
	}
	switch c := p.s[p.i]; {
	case c == '"':
		text, err := p.quoted(false)
		return &node{kind: nodeString, text: text}, err
	case c == 'b' && p.i+1 < len(p.s) && p.s[p.i+1] == '"':
		p.i++
		text, err := p.quoted(true)
		return &node{kind: nodeBytes, text: text}, err
	case c == '[':
		p.i++
		items, err := p.sequence(']')
		return &node{kind: nodeList, items: items}, err
	case c == '(':
		p.i++
		items, err := p.sequence(')')
		return tuple("", items), err
	case isToken(c):
		start := p.i
		for p.i < len(p.s) && isToken(p.s[p.i]) {
			p.i++
		}
		token := p.s[start:p.i]
		switch {
		case p.i < len(p.s) && p.s[p.i] == '(':
			p.i++
			items, err := p.sequence(')')
			return tuple(token, items), err
		case strings.HasPrefix(p.s[p.i:], " {"):
			p.i += 2
			return p.structure()
		case token == "None":
			// Also an enum variant name; diffTrees resolves it against Go.
			return &node{kind: nodeNone, text: token}, nil
		}
		return &node{kind: nodeLeaf, text: token}, nil
	default:
		return nil, fmt.Errorf("unexpected %q", c)
	}
}

// tuple unwraps Some, newtypes and one-field tuple variants, and keeps other
// tuples, such as glam vectors, as lists.
func tuple(_ string, items []*node) *node {
	if len(items) == 1 {
		return items[0]
	}
	return &node{kind: nodeList, items: items}
}

func (p *debugParser) sequence(end byte) ([]*node, error) {
	var items []*node
	for {
		if p.eat(end) {
			return items, nil
		}
		item, err := p.value()
		if err != nil {
			return nil, err
		}
		items = append(items, item)
		if !p.eat(',') && !p.eat(end) {
			return nil, fmt.Errorf("want , or %q", end)
		} else if p.s[p.i-1] == end {
			return items, nil
		}
	}
}

func (p *debugParser) structure() (*node, error) {
	n := &node{kind: nodeStruct}
	for {
		if p.eat('}') {
			return n, nil
		}
		p.space()
		start := p.i
		for p.i < len(p.s) && isToken(p.s[p.i]) {
			p.i++
		}
		name := p.s[start:p.i]
		if name == "" || !p.eat(':') {
			return nil, errors.New("want a field name and :")
		}
		value, err := p.value()
		if err != nil {
			return nil, err
		}
		n.names, n.items = append(n.names, name), append(n.items, value)
		if !p.eat(',') && !p.eat('}') {
			return nil, errors.New("want , or }")
		} else if p.s[p.i-1] == '}' {
			return n, nil
		}
	}
}

// quoted reads a Rust string or byte-string literal starting at its quote.
func (p *debugParser) quoted(bytesLiteral bool) (string, error) {
	p.i++ // opening quote
	var b strings.Builder
	for p.i < len(p.s) {
		c := p.s[p.i]
		p.i++
		switch c {
		case '"':
			return b.String(), nil
		case '\\':
			if p.i >= len(p.s) {
				return "", errors.New("truncated escape")
			}
			e := p.s[p.i]
			p.i++
			switch e {
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			case 't':
				b.WriteByte('\t')
			case '0':
				b.WriteByte(0)
			case '\\', '"', '\'':
				b.WriteByte(e)
			case 'x':
				if p.i+2 > len(p.s) {
					return "", errors.New("truncated \\x escape")
				}
				v, err := strconv.ParseUint(p.s[p.i:p.i+2], 16, 8)
				if err != nil {
					return "", err
				}
				b.WriteByte(byte(v))
				p.i += 2
			case 'u':
				end := strings.IndexByte(p.s[p.i:], '}')
				if !strings.HasPrefix(p.s[p.i:], "{") || end < 0 {
					return "", errors.New("malformed \\u escape")
				}
				v, err := strconv.ParseUint(p.s[p.i+1:p.i+end], 16, 32)
				if err != nil {
					return "", err
				}
				b.WriteRune(rune(v))
				p.i += end + 1
			default:
				return "", fmt.Errorf("unknown escape \\%c", e)
			}
		default:
			b.WriteByte(c)
		}
	}
	if bytesLiteral {
		return "", errors.New("unterminated byte string")
	}
	return "", errors.New("unterminated string")
}

var (
	uuidType     = reflect.TypeOf(uuid.UUID{})
	protocolPath = reflect.TypeOf(genprotocol.BlockPos{}).PkgPath()
)

// goTree turns a decoded generated-Go value into a value tree.
func goTree(v reflect.Value) (*node, error) {
	switch v.Kind() {
	case reflect.Invalid:
		return &node{kind: nodeNone}, nil
	case reflect.Pointer, reflect.Interface:
		if v.IsNil() {
			return &node{kind: nodeNone}, nil
		}
		return goTree(v.Elem())
	case reflect.Struct:
		t := v.Type()
		if t == uuidType {
			return &node{kind: nodeLeaf, text: v.Interface().(uuid.UUID).String()}, nil
		}
		if t.PkgPath() == protocolPath && strings.HasPrefix(t.Name(), "Optional[") {
			if !v.FieldByName("set").Bool() {
				return &node{kind: nodeNone}, nil
			}
			return goTree(v.FieldByName("val"))
		}
		n := &node{kind: nodeStruct}
		for i := 0; i < v.NumField(); i++ {
			child, err := goTree(v.Field(i))
			if err != nil {
				return nil, fmt.Errorf("%s.%s: %w", t.Name(), t.Field(i).Name, err)
			}
			n.names, n.items = append(n.names, t.Field(i).Name), append(n.items, child)
		}
		return n, nil
	case reflect.Slice, reflect.Array:
		if v.Kind() == reflect.Slice && v.Type().Elem().Kind() == reflect.Uint8 {
			return &node{kind: nodeBytes, text: string(v.Bytes())}, nil
		}
		n := &node{kind: nodeList}
		for i := 0; i < v.Len(); i++ {
			child, err := goTree(v.Index(i))
			if err != nil {
				return nil, fmt.Errorf("[%d]: %w", i, err)
			}
			n.items = append(n.items, child)
		}
		return n, nil
	case reflect.String:
		return &node{kind: nodeString, text: v.String()}, nil
	case reflect.Bool, reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n := &node{kind: nodeLeaf, goVal: v}
		if v.Type().PkgPath() == protocolPath {
			n.goEnum = v.Type().Name()
		}
		return n, nil
	default:
		return nil, fmt.Errorf("unsupported %s", v.Type())
	}
}

// RustEnums maps each generated Rust enum to its variants' raw values, read
// from the crate's From<raw> implementations.
type RustEnums map[string]map[string]string

var (
	enumFromImpl = regexp.MustCompile(`(?s)impl From<[a-z0-9]+> for (\w+) \{\s*fn from\(value: [a-z0-9]+\) -> Self \{\s*match value \{(.*?)\n\s*\}\n\s*\}\n\}`)
	enumArm      = regexp.MustCompile(`(-?\d+) => Self::(\w+),`)
)

// LoadRustEnums reads every enum table in the generated crate's sources.
func LoadRustEnums(srcDir string) (RustEnums, error) {
	files, err := filepath.Glob(filepath.Join(srcDir, "*.rs"))
	if err != nil || len(files) == 0 {
		return nil, fmt.Errorf("no Rust sources in %s: %v", srcDir, err)
	}
	enums := RustEnums{}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		for _, m := range enumFromImpl.FindAllStringSubmatch(string(data), -1) {
			variants := map[string]string{}
			for _, arm := range enumArm.FindAllStringSubmatch(m[2], -1) {
				variants[arm[2]] = arm[1]
			}
			enums[m[1]] = variants
		}
	}
	if len(enums) == 0 {
		return nil, fmt.Errorf("no enum tables in %s", srcDir)
	}
	return enums, nil
}

// diffTrees returns the first difference between a Rust and a Go tree, or "".
func diffTrees(r, g *node, path string, enums RustEnums) string {
	// A one-field Go struct can hold what Rust prints as its bare value, and
	// a field-less one what Rust prints as a unit variant.
	if g.kind == nodeStruct && r.kind != nodeStruct {
		switch {
		case len(g.items) == 1:
			return diffTrees(r, g.items[0], path+"."+g.names[0], enums)
		case len(g.items) == 0 && r.kind == nodeLeaf:
			return ""
		}
	}
	if r.kind == nodeStruct && len(r.items) == 1 && (g.kind != nodeStruct || len(g.items) != 1) {
		return diffTrees(r.items[0], g, path+"."+r.names[0], enums)
	}
	if r.kind == nodeNone && r.text != "" && g.kind == nodeLeaf {
		r = &node{kind: nodeLeaf, text: r.text}
	}
	// Rust prints map entries as (key, value) tuples, Go holds them as
	// two-field structs.
	if g.kind == nodeStruct && r.kind == nodeList && len(r.items) == len(g.items) {
		r = &node{kind: nodeStruct, names: g.names, items: r.items}
	}
	// Rust prints a uuid as its hyphenated text, Go holds the sixteen bytes.
	if g.kind == nodeList && len(g.items) == 16 && r.kind == nodeLeaf {
		id, err := uuid.Parse(r.text)
		if err != nil {
			return fmt.Sprintf("%s: go 16 bytes, rust %s", path, r.text)
		}
		return diffBytesList(string(id[:]), g, path)
	}
	switch g.kind {
	case nodeNone:
		if r.kind != nodeNone {
			return path + ": rust has a value, go has none"
		}
	case nodeStruct:
		if r.kind != nodeStruct || len(r.items) != len(g.items) {
			return fmt.Sprintf("%s: go struct with %d fields, rust %s", path, len(g.items), describeNode(r))
		}
		for i := range g.items {
			if d := diffTrees(r.items[i], g.items[i], path+"."+g.names[i], enums); d != "" {
				return d
			}
		}
	case nodeList:
		if r.kind == nodeBytes {
			return diffBytesList(r.text, g, path)
		}
		if r.kind != nodeList || len(r.items) != len(g.items) {
			return fmt.Sprintf("%s: go list of %d, rust %s", path, len(g.items), describeNode(r))
		}
		for i := range g.items {
			if d := diffTrees(r.items[i], g.items[i], fmt.Sprintf("%s[%d]", path, i), enums); d != "" {
				return d
			}
		}
	case nodeBytes:
		switch r.kind {
		case nodeBytes:
			if r.text != g.text {
				return fmt.Sprintf("%s: bytes differ (rust %d, go %d)", path, len(r.text), len(g.text))
			}
		case nodeList:
			return diffBytesList(g.text, r, path)
		default:
			return fmt.Sprintf("%s: go bytes, rust %s", path, describeNode(r))
		}
	case nodeString:
		// Rust renders strings lossily, so compare the lossy form.
		if r.kind != nodeString || r.text != lossy(g.text) {
			return fmt.Sprintf("%s: go %q, rust %s", path, g.text, describeNode(r))
		}
	case nodeLeaf:
		if r.kind != nodeLeaf || !leafEqual(r.text, g, enums) {
			return fmt.Sprintf("%s: go %v, rust %s", path, g.goVal, describeNode(r))
		}
	}
	return ""
}

func diffBytesList(data string, list *node, path string) string {
	if list.kind != nodeList || len(list.items) != len(data) {
		return fmt.Sprintf("%s: %d bytes against %s", path, len(data), describeNode(list))
	}
	for i, item := range list.items {
		if item.kind != nodeLeaf || item.text != strconv.Itoa(int(data[i])) && !(item.goVal.IsValid() && item.goVal.CanUint() && item.goVal.Uint() == uint64(data[i])) {
			return fmt.Sprintf("%s[%d]: byte %d against %s", path, i, data[i], describeNode(item))
		}
	}
	return ""
}

func leafEqual(token string, g *node, enums RustEnums) bool {
	v := g.goVal
	switch v.Kind() {
	case reflect.Bool:
		return token == strconv.FormatBool(v.Bool())
	case reflect.Float32, reflect.Float64:
		bits := 64
		if v.Kind() == reflect.Float32 {
			bits = 32
		}
		f, err := strconv.ParseFloat(strings.ReplaceAll(token, "inf", "Inf"), bits)
		if err != nil {
			return false
		}
		if math.IsNaN(f) || math.IsNaN(v.Float()) {
			return math.IsNaN(f) && math.IsNaN(v.Float())
		}
		if bits == 32 {
			return math.Float32bits(float32(f)) == math.Float32bits(float32(v.Float()))
		}
		return math.Float64bits(f) == math.Float64bits(v.Float())
	}
	want := ""
	if v.CanInt() {
		want = strconv.FormatInt(v.Int(), 10)
	} else {
		want = strconv.FormatUint(v.Uint(), 10)
	}
	if token == want {
		return true
	}
	if raw, ok := enums[g.goEnum][token]; ok {
		return raw == want
	}
	return false
}

func lossy(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	return strings.ToValidUTF8(s, "�")
}

func describeNode(n *node) string {
	switch n.kind {
	case nodeNone:
		return "none"
	case nodeStruct:
		return fmt.Sprintf("struct with %d fields", len(n.items))
	case nodeList:
		return fmt.Sprintf("list of %d", len(n.items))
	case nodeString:
		return strconv.Quote(n.text)
	case nodeBytes:
		return fmt.Sprintf("%d bytes", len(n.text))
	}
	return n.text
}

// FieldCheck is one decoded entry whose Rust and Go values were compared.
type FieldCheck struct {
	Fixture, Case string
	Entry         int
	Diff          string // "" when the values agree
}

// CompareFields checks every entry the Rust runner rendered against the
// generated Go decoding of the same input entry.
func CompareFields(out string, inputs []Input, enums RustEnums) ([]FieldCheck, error) {
	byKey := map[[2]string]Input{}
	for _, in := range inputs {
		byKey[[2]string{in.Case, in.Fixture}] = in
	}
	file, err := os.Open(filepath.Join(out, "fields", DecoderRust+".tsv"))
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(nil, 256<<20)
	var checks []FieldCheck
	for scanner.Scan() {
		cols := strings.SplitN(scanner.Text(), "\t", 4)
		if len(cols) != 4 {
			return nil, errors.New("fields/rust.tsv: want 4 columns")
		}
		entry, err := strconv.Atoi(cols[2])
		if err != nil {
			return nil, err
		}
		check := FieldCheck{Fixture: cols[0], Case: cols[1], Entry: entry}
		check.Diff = compareEntryFields(byKey[[2]string{cols[1], cols[0]}], entry, cols[3], enums)
		checks = append(checks, check)
	}
	return checks, scanner.Err()
}

func compareEntryFields(in Input, index int, debug string, enums RustEnums) string {
	entries, _ := SplitBatch(in.Batch)
	if index >= len(entries) {
		return "no such input entry"
	}
	e := entries[index]
	pk, ok := newGeneratedPacket(in.Sender, e.Header&idMask)
	if !ok {
		return "go has no such packet"
	}
	if err := decodeGenerated(pk, e.Body, generatedReader(in.Sender)); err != nil {
		return "go does not decode: " + err.Error()
	}
	rust, err := ParseRustDebug(debug)
	if err != nil {
		return "rust rendering does not parse: " + err.Error()
	}
	g, err := goTree(reflect.ValueOf(pk))
	if err != nil {
		return err.Error()
	}
	// Rust renders the packet enum's variant around the packet struct.
	return diffTrees(rust, g, reflect.TypeOf(pk).Elem().Name(), enums)
}
