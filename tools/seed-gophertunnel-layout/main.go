// Command seed-gophertunnel-layout derives a layout overlay from a gophertunnel
// checkout and writes a gap report: which fork types, fields, and constants
// the generated tree already covers, and which it does not.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/constant"
	"go/format"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"protocolgen/internal/docs"
	"protocolgen/internal/flatten"
	"protocolgen/internal/layout"
	"protocolgen/internal/manifest"
	"protocolgen/internal/naming"
	"protocolgen/internal/semantics"
)

func main() {
	manifestPath := flag.String("manifest", "", "canonical manifest")
	namingPath := flag.String("naming", "", "reviewed naming overlay")
	gopherPath := flag.String("gophertunnel", "", "gophertunnel checkout")
	outPath := flag.String("out", "", "layout overlay output")
	reportPath := flag.String("report", "", "gap report output (Markdown)")
	docsPath := flag.String("docs", "", "reviewed docs overlay to extend with fork comments for matched types and fields (optional)")
	semanticsPath := flag.String("semantics-out", "", "semantics overlay output marking actor identifier fields (optional)")
	flag.Parse()
	if *manifestPath == "" || *namingPath == "" || *gopherPath == "" || *outPath == "" || *reportPath == "" {
		fail("-manifest, -naming, -gophertunnel, -out and -report are required")
	}
	m, err := manifest.Load(*manifestPath)
	if err != nil {
		fail("%v", err)
	}
	overlay, err := naming.LoadOverlay(*namingPath, m)
	if err != nil {
		fail("%v", err)
	}
	index, err := buildIndex(m, overlay)
	if err != nil {
		fail("%v", err)
	}
	fork, err := parseFork(*gopherPath)
	if err != nil {
		fail("%v", err)
	}
	var docOverlay docs.Overlay
	if *docsPath != "" {
		docOverlay, err = docs.LoadOverlay(*docsPath, m)
		if err != nil {
			fail("%v", err)
		}
	}
	document, report := seed(m, index, fork, docOverlay)
	document.Target = m.Target
	if *docsPath != "" {
		if err := writeJSON(*docsPath, docs.Document{SchemaVersion: 1, Target: m.Target, Entries: docs.SortedEntries(docOverlay)}); err != nil {
			fail("%v", err)
		}
		fmt.Printf("docs: %d fork comments ported -> %s\n", report.portedDocs, *docsPath)
	}
	if *semanticsPath != "" {
		semantic := semantics.SortedDocument(semantics.Document{SchemaVersion: 1, Target: m.Target, Entries: report.semantics})
		if err := semantics.ValidateOverlay(m, semantic); err != nil {
			fail("seeded semantics are invalid: %v", err)
		}
		if err := writeJSON(*semanticsPath, semantic); err != nil {
			fail("%v", err)
		}
		fmt.Printf("semantics: %d actor identifier fields (%d by name only) -> %s\n", len(semantic.Entries), report.heuristicSemantics, *semanticsPath)
	}
	if err := layout.ValidateOverlay(m, document); err != nil {
		fail("seeded overlay is invalid: %v", err)
	}
	if err := writeJSON(*outPath, layout.SortedDocument(document)); err != nil {
		fail("%v", err)
	}
	if err := os.WriteFile(*reportPath, []byte(report.render(m, fork)), 0o644); err != nil {
		fail("%v", err)
	}
	fmt.Printf("layout: %d constant placements (%d variants named), %d field names -> %s\n", len(document.Constants), report.namedVariants, len(document.Fields), *outPath)
	fmt.Printf("gaps: %d fork types unmatched, %d generated types unmatched, %d const groups unmatched -> %s\n", len(report.forkOnlyTypes), len(report.generatedOnlyTypes), len(report.unmatchedGroups), *reportPath)
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

// --- manifest side ---------------------------------------------------------

type enumInfo struct {
	TypeID    string
	Name      string
	Variants  []manifest.Variant
	Inferred  bool // identity inferred from the node, so other enums may share it
	Ambiguous bool // two enums share the identity with different variants; no placement can tell them apart
}

type ownerInfo struct {
	TypeID  string // type ID that owns the fields: the inlined payload struct for a flattened packet
	FileKey string // key for file placement: the packet name for packets, else TypeID
	Name    string // Go name
	Fields  []manifest.Field
	Packet  bool
}

type index struct {
	enums  []enumInfo
	owners []ownerInfo
}

func buildIndex(m manifest.Manifest, overlay naming.Overlay) (index, error) {
	resolver := naming.NewResolver(overlay)
	for _, packet := range m.Packets {
		if err := resolver.Reserve(packet.Name, naming.PacketTypeName(packet.Name), naming.GoExportName); err != nil {
			return index{}, err
		}
	}
	var result index
	seenEnum := map[string]int{}
	seenOwner := map[string]bool{}
	// A payload struct inlined into its packet is indexed as that packet, not
	// as a shared type.
	usage := flatten.Count(m)
	native := func(node manifest.Node) bool { return nativeTypeIDs[node.TypeID] }
	for _, packet := range m.Packets {
		fields := packet.Fields
		for len(fields) == 1 && usage.Inlined(fields[0], native) {
			seenOwner[flatten.TypeID(fields[0].Encode)] = true
			fields = fields[0].Encode.Fields
		}
	}
	var walk func(node manifest.Node) error
	walk = func(node manifest.Node) error {
		typeID := node.TypeID
		if typeID == "" {
			typeID = naming.InferredTypeName(node)
		}
		if typeID != "" {
			switch node.Kind {
			case manifest.KindEnum:
				position, seen := seenEnum[typeID]
				if !seen {
					name, err := resolver.Resolve(node, "", naming.GoExportName)
					if err != nil {
						return err
					}
					position = len(result.enums)
					seenEnum[typeID] = position
					result.enums = append(result.enums, enumInfo{TypeID: typeID, Name: name, Inferred: node.TypeID == ""})
				}
				// Union the variants of every site so partial claims still map; an
				// inferred identity whose sites disagree on a value is ambiguous.
				known := map[string]int64{}
				for _, variant := range result.enums[position].Variants {
					known[variant.Name] = variant.Value
				}
				for _, variant := range node.Variants {
					if value, ok := known[variant.Name]; ok {
						if value != variant.Value && result.enums[position].Inferred {
							result.enums[position].Ambiguous = true
						}
						continue
					}
					result.enums[position].Variants = append(result.enums[position].Variants, variant)
				}
			case manifest.KindStruct:
				if !seenOwner[typeID] && len(node.Fields) > 0 {
					name, err := resolver.Resolve(node, "", naming.GoExportName)
					if err != nil {
						return err
					}
					seenOwner[typeID] = true
					result.owners = append(result.owners, ownerInfo{TypeID: typeID, FileKey: typeID, Name: name, Fields: node.Fields})
				}
			}
		}
		for _, field := range node.Fields {
			if err := walk(field.Encode); err != nil {
				return err
			}
			if field.Decode != nil {
				if err := walk(*field.Decode); err != nil {
					return err
				}
			}
		}
		for _, variant := range node.Variants {
			if err := walk(variant.Encode); err != nil {
				return err
			}
			if variant.Decode != nil {
				if err := walk(*variant.Decode); err != nil {
					return err
				}
			}
		}
		for _, child := range []*manifest.Node{node.Prefix, node.Element, node.Value, node.Key, node.Control, node.Default} {
			if child != nil {
				if err := walk(*child); err != nil {
					return err
				}
			}
		}
		for _, child := range node.Elements {
			if err := walk(child); err != nil {
				return err
			}
		}
		for _, oneCase := range node.Cases {
			for _, child := range append(oneCase.Encode, oneCase.Decode...) {
				if err := walk(child); err != nil {
					return err
				}
			}
		}
		return nil
	}
	for _, packet := range m.Packets {
		// The emitter inlines a sole single-use payload struct, so the packet's
		// effective fields and their owning type ID must match what it keys
		// docs and layout names by.
		effective := usage.PacketFields(packet, native)
		owner := packet.Name
		fields := make([]manifest.Field, 0, len(effective))
		for _, item := range effective {
			owner = item.Owner
			fields = append(fields, item.Field)
		}
		result.owners = append(result.owners, ownerInfo{TypeID: owner, FileKey: packet.Name, Name: naming.GoExportName(naming.PacketTypeName(packet.Name)), Fields: fields, Packet: true})
		for _, field := range packet.Fields {
			if err := walk(field.Encode); err != nil {
				return index{}, err
			}
			if field.Decode != nil {
				if err := walk(*field.Decode); err != nil {
					return index{}, err
				}
			}
		}
	}
	return result, nil
}

// --- fork side --------------------------------------------------------------

type forkField struct {
	Name     string
	Type     string
	Doc      string
	Semantic string // ActorUniqueID or ActorRuntimeID when the fork marshals the field with an identifier operation
}

type forkType struct {
	Name     string
	Package  string
	File     string
	Doc      string
	Fields   []forkField
	Struct   bool
	Marshals bool // has a Marshal method, so it is a wire type rather than runtime plumbing
}

type forkConst struct {
	Name  string
	Value int64
}

type forkGroup struct {
	Package string
	File    string
	Consts  []forkConst
	prefix  string
}

type forkIndex struct {
	commit string
	types  map[string]forkType // key: package + "." + name
	groups []forkGroup
}

func parseFork(root string) (forkIndex, error) {
	result := forkIndex{types: map[string]forkType{}}
	marshals := map[string]bool{}
	fieldSemantics := map[string]string{} // pkg.Type.Field -> semantic
	if out, err := exec.Command("git", "-C", root, "rev-parse", "HEAD").Output(); err == nil {
		result.commit = strings.TrimSpace(string(out))
	}
	for _, pkg := range []string{"protocol", "packet"} {
		directory := filepath.Join(root, "minecraft", "protocol")
		if pkg == "packet" {
			directory = filepath.Join(directory, "packet")
		}
		paths, err := filepath.Glob(filepath.Join(directory, "*.go"))
		if err != nil {
			return forkIndex{}, err
		}
		sort.Strings(paths)
		fset := token.NewFileSet()
		var files []*ast.File
		stems := map[*ast.File]string{}
		for _, path := range paths {
			if strings.HasSuffix(path, "_test.go") {
				continue
			}
			file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
			if err != nil {
				return forkIndex{}, fmt.Errorf("parse %s: %w", path, err)
			}
			if file.Name.Name != pkg {
				continue
			}
			files = append(files, file)
			stems[file] = strings.TrimSuffix(filepath.Base(path), ".go")
		}
		if len(files) == 0 {
			return forkIndex{}, fmt.Errorf("no non-test %s source files in %s", pkg, directory)
		}
		evaluator := newConstEvaluator(fset, files)
		for _, file := range files {
			for _, declaration := range file.Decls {
				if fn, ok := declaration.(*ast.FuncDecl); ok && fn.Name.Name == "Marshal" && fn.Recv != nil && len(fn.Recv.List) == 1 {
					receiver := fn.Recv.List[0].Type
					if star, ok := receiver.(*ast.StarExpr); ok {
						receiver = star.X
					}
					if ident, ok := receiver.(*ast.Ident); ok {
						marshals[pkg+"."+ident.Name] = true
					}
				}
				if fn, ok := declaration.(*ast.FuncDecl); ok {
					for key, semantic := range identifierFields(fn) {
						fieldSemantics[pkg+"."+key] = semantic
					}
					continue
				}
				gen, ok := declaration.(*ast.GenDecl)
				if !ok {
					continue
				}
				switch gen.Tok {
				case token.TYPE:
					for _, spec := range gen.Specs {
						typeSpec := spec.(*ast.TypeSpec)
						item := forkType{Name: typeSpec.Name.Name, Package: pkg, File: stems[file], Doc: commentText(typeSpec.Doc)}
						if item.Doc == "" {
							item.Doc = commentText(gen.Doc)
						}
						if structure, ok := typeSpec.Type.(*ast.StructType); ok {
							item.Struct = true
							for _, field := range structure.Fields.List {
								var typ strings.Builder
								_ = format.Node(&typ, fset, field.Type)
								doc := commentText(field.Doc)
								if doc == "" {
									doc = commentText(field.Comment)
								}
								for _, name := range field.Names {
									item.Fields = append(item.Fields, forkField{Name: name.Name, Type: typ.String(), Doc: doc})
								}
							}
						}
						key := pkg + "." + item.Name
						if _, exists := result.types[key]; !exists {
							result.types[key] = item
						}
					}
				case token.CONST:
					group := forkGroup{Package: pkg, File: stems[file]}
					for _, c := range evaluator.group(gen) {
						group.Consts = append(group.Consts, c)
					}
					if len(group.Consts) > 0 {
						group.prefix = commonPrefix(group.Consts)
						// The fork sometimes continues one enum in a second block in
						// the same file (ActorEvent restarts at 57); treat those as one.
						if last := len(result.groups) - 1; last >= 0 && result.groups[last].File == group.File && result.groups[last].Package == group.Package && result.groups[last].prefix != "" && result.groups[last].prefix == group.prefix {
							result.groups[last].Consts = append(result.groups[last].Consts, group.Consts...)
						} else {
							result.groups = append(result.groups, group)
						}
					}
				}
			}
		}
	}
	for key, item := range result.types {
		item.Marshals = marshals[key]
		for index := range item.Fields {
			item.Fields[index].Semantic = fieldSemantics[key+"."+item.Fields[index].Name]
		}
		result.types[key] = item
	}
	return result, nil
}

// identifierFields finds the struct fields a function passes to an actor
// identifier IO operation, keyed "Type.Field", through the function's
// receiver and pointer parameters (Marshal methods and the Reader/Writer
// helpers alike), including as an optional's callback.
func identifierFields(fn *ast.FuncDecl) map[string]string {
	result := map[string]string{}
	if fn.Body == nil {
		return result
	}
	variables := map[string]string{}
	bind := func(names []*ast.Ident, typ ast.Expr) {
		if star, ok := typ.(*ast.StarExpr); ok {
			typ = star.X
		}
		ident, ok := typ.(*ast.Ident)
		if !ok {
			return
		}
		for _, name := range names {
			variables[name.Name] = ident.Name
		}
	}
	if fn.Recv != nil {
		for _, field := range fn.Recv.List {
			bind(field.Names, field.Type)
		}
	}
	for _, field := range fn.Type.Params.List {
		bind(field.Names, field.Type)
	}
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		semantic := ""
		operation := func(expr ast.Expr) {
			if selector, ok := expr.(*ast.SelectorExpr); ok {
				switch selector.Sel.Name {
				case semantics.ActorUniqueID, semantics.ActorRuntimeID:
					semantic = selector.Sel.Name
				}
			}
		}
		operation(call.Fun)
		for _, argument := range call.Args {
			operation(argument)
		}
		if semantic == "" {
			return true
		}
		for _, argument := range call.Args {
			address, ok := argument.(*ast.UnaryExpr)
			if !ok || address.Op != token.AND {
				continue
			}
			selector, ok := address.X.(*ast.SelectorExpr)
			if !ok {
				continue
			}
			variable, ok := selector.X.(*ast.Ident)
			if !ok {
				continue
			}
			if typ, ok := variables[variable.Name]; ok {
				result[typ+"."+selector.Sel.Name] = semantic
			}
		}
		return true
	})
	return result
}

// constEvaluator stores the exact integer constants Go can resolve in a package.
type constEvaluator struct {
	values map[string]int64
}

// newConstEvaluator uses Go's type checker so conversions and typed arithmetic
// keep their declared widths. Unsupported constants are omitted.
func newConstEvaluator(fset *token.FileSet, files []*ast.File) *constEvaluator {
	e := &constEvaluator{values: map[string]int64{}}
	info := &types.Info{Defs: map[*ast.Ident]types.Object{}}
	cfg := types.Config{Importer: importer.Default(), IgnoreFuncBodies: true, Error: func(error) {}}
	// Unrelated unresolved types must not hide successfully checked constants.
	_, _ = cfg.Check(files[0].Name.Name, fset, files, info)
	for ident, object := range info.Defs {
		value, ok := object.(*types.Const)
		if !ok || ident.Name == "_" || value.Val().Kind() != constant.Int {
			continue
		}
		if integer, exact := constant.Int64Val(value.Val()); exact {
			e.values[ident.Name] = integer
		}
	}
	return e
}

// group returns the supported integer constants declared in one const block.
func (e *constEvaluator) group(gen *ast.GenDecl) []forkConst {
	var result []forkConst
	for _, spec := range gen.Specs {
		for _, name := range spec.(*ast.ValueSpec).Names {
			if value, ok := e.values[name.Name]; ok {
				result = append(result, forkConst{Name: name.Name, Value: value})
			}
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

func commonPrefix(consts []forkConst) string {
	if len(consts) == 0 {
		return ""
	}
	prefix := consts[0].Name
	for _, c := range consts[1:] {
		for !strings.HasPrefix(c.Name, prefix) {
			prefix = prefix[:len(prefix)-1]
		}
	}
	// Cut back to a word boundary so "GameTypeS" becomes "GameType" while
	// "IDLogin"/"IDPlayStatus" keeps "ID".
	for k := len(prefix); k > 0; k-- {
		boundary := true
		for _, c := range consts {
			if len(c.Name) > k && !unicode.IsUpper(rune(c.Name[k])) && !unicode.IsDigit(rune(c.Name[k])) {
				boundary = false
				break
			}
		}
		if boundary {
			return prefix[:k]
		}
	}
	return ""
}

// --- matching -----------------------------------------------------------------

type gapReport struct {
	namedVariants      int
	portedDocs         int
	heuristicSemantics int
	semantics          []semantics.Entry
	ambiguousEnums     []enumInfo
	skippedTypeNames   []string
	placements         []placementNote
	unmatchedEnums     []enumInfo
	unmatchedGroups    []forkGroup
	forkOnlyTypes      []forkType
	generatedOnlyTypes []ownerInfo
	fieldGaps          []fieldGap
}

type placementNote struct {
	Enum     enumInfo
	Group    forkGroup
	Unnamed  []string
	Foreign  []forkConst // fork constants with no manifest variant
	Mismatch []string    // same name, different value
	Score    int
}

type fieldGap struct {
	Owner        ownerInfo
	Fork         forkType
	Positional   bool
	OnlyManifest []string
	OnlyFork     []string
	TypeMismatch []string
	Renamed      int
}

func seed(m manifest.Manifest, idx index, fork forkIndex, docOverlay docs.Overlay) (layout.Document, *gapReport) {
	document := layout.Document{SchemaVersion: 1}
	report := &gapReport{}
	usedGroups := map[int]bool{}
	for i, group := range fork.groups {
		// Packet IDs are emitted by the generator itself; never relocate them.
		if group.prefix == "ID" {
			usedGroups[i] = true
		}
	}
	// Every enum scores every block, and the best pairs claim first, so a
	// block is not taken by a weaker enum that merely sorts earlier.
	type candidate struct {
		enum, group, score int
	}
	var candidates []candidate
	for e, enum := range idx.enums {
		if enum.Ambiguous {
			report.ambiguousEnums = append(report.ambiguousEnums, enum)
			continue
		}
		for i, group := range fork.groups {
			if usedGroups[i] {
				continue
			}
			if score := matchScore(enum, group); score >= 2 && score*2 >= len(enum.Variants) {
				candidates = append(candidates, candidate{e, i, score})
			}
		}
	}
	sort.SliceStable(candidates, func(i, j int) bool { return candidates[i].score > candidates[j].score })
	chosen := map[int]candidate{}
	for _, c := range candidates {
		if _, taken := chosen[c.enum]; taken || usedGroups[c.group] {
			continue
		}
		chosen[c.enum], usedGroups[c.group] = c, true
	}
	for e, enum := range idx.enums {
		if enum.Ambiguous {
			continue
		}
		c, ok := chosen[e]
		if !ok {
			report.unmatchedEnums = append(report.unmatchedEnums, enum)
			continue
		}
		bestScore := c.score
		group := fork.groups[c.group]
		if group.Package == "protocol" {
			document.Files = append(document.Files, layout.FileEntry{TypeID: enum.TypeID, Package: "protocol", File: group.File, Rationale: fmt.Sprintf("gophertunnel keeps the %s constants in protocol/%s.go.", group.prefix, group.File)})
		}
		names, note := variantNames(enum, group)
		note.Score = bestScore
		report.placements = append(report.placements, note)
		report.namedVariants += len(names)
		document.Constants = append(document.Constants, layout.ConstantEntry{
			TypeID:    enum.TypeID,
			Package:   group.Package,
			File:      group.File,
			Names:     names,
			Rationale: fmt.Sprintf("Matches the %s const block in gophertunnel %s/%s.go (%d of %d variants named by value and name).", group.prefix, group.Package, group.File, len(names), len(enum.Variants)),
		})
	}
	for i, group := range fork.groups {
		if !usedGroups[i] {
			report.unmatchedGroups = append(report.unmatchedGroups, group)
		}
	}

	matchedFork := map[string]bool{}
	var unmatchedOwners []ownerInfo
	// Generated names are unique across both packages, so a fork name that
	// another type or packet already carries cannot be taken.
	taken := map[string]string{}
	for _, owner := range idx.owners {
		taken[owner.Name] = owner.FileKey
	}
	pair := func(owner ownerInfo, source forkType) {
		matchedFork[source.Package+"."+source.Name] = true
		finalName := owner.Name
		if source.Name != owner.Name {
			if holder, exists := taken[source.Name]; exists && holder != owner.FileKey {
				report.skippedTypeNames = append(report.skippedTypeNames, fmt.Sprintf("%s -> %s (name held by %s)", owner.Name, source.Name, holder))
			} else {
				taken[source.Name] = owner.FileKey
				finalName = source.Name
				document.Types = append(document.Types, layout.TypeEntry{TypeID: owner.FileKey, Name: source.Name, Rationale: fmt.Sprintf("gophertunnel names this %s.", source.Name)})
			}
		}
		document.Files = append(document.Files, layout.FileEntry{TypeID: owner.FileKey, Package: source.Package, File: source.File, Rationale: fmt.Sprintf("gophertunnel keeps %s in %s/%s.go.", source.Name, source.Package, source.File)})
		if docOverlay.Types != nil && source.Doc != "" {
			// The doc leads with the type's final name; an earlier port under the
			// generated name is re-led rather than duplicated.
			if existing := docOverlay.Types[owner.FileKey]; existing == "" {
				docOverlay.Types[owner.FileKey] = docs.LeadWith(source.Doc, source.Name, finalName)
				report.portedDocs++
			} else if led := docs.LeadWith(existing, owner.Name, finalName); led != existing {
				docOverlay.Types[owner.FileKey] = led
			}
		}
		gap := fieldGap{Owner: owner, Fork: source}
		entries := matchFields(owner, source, &gap, docOverlay, report)
		document.Fields = append(document.Fields, entries...)
		gap.Renamed = len(entries)
		if len(gap.OnlyManifest)+len(gap.OnlyFork)+len(gap.TypeMismatch) > 0 || gap.Renamed > 0 {
			report.fieldGaps = append(report.fieldGaps, gap)
		}
	}
	for _, owner := range idx.owners {
		if nativeTypeIDs[owner.TypeID] {
			continue
		}
		if source, ok := forkTypeFor(owner.Name, ownerPackage(owner), fork); ok {
			pair(owner, source)
		} else {
			unmatchedOwners = append(unmatchedOwners, owner)
		}
	}
	// Second pass: types whose names differ but whose field names mostly agree.
	for _, owner := range unmatchedOwners {
		source, ok := forkTypeByFields(owner, fork, matchedFork)
		if ok {
			pair(owner, source)
		} else {
			report.generatedOnlyTypes = append(report.generatedOnlyTypes, owner)
			report.semanticsByName(owner, nil)
			document.Fields = append(document.Fields, report.conventionEntries(owner, nil)...)
		}
	}
	for key, item := range fork.types {
		if item.Struct && item.Marshals && !matchedFork[key] {
			report.forkOnlyTypes = append(report.forkOnlyTypes, item)
		}
	}
	sort.Slice(report.forkOnlyTypes, func(i, j int) bool {
		return report.forkOnlyTypes[i].Package+report.forkOnlyTypes[i].Name < report.forkOnlyTypes[j].Package+report.forkOnlyTypes[j].Name
	})
	return document, report
}

// matchScore counts variants a const group explains: two points when a
// constant shares the value and the name, one when only the value matches an
// otherwise plausible group.
// genericVariants are variant names most enums carry, so matching one says
// nothing about which enum a const block belongs to.
var genericVariants = map[string]bool{"unknown": true, "none": true, "default": true, "count": true, "invalid": true, "all": true, "other": true, "custom": true, "undefined": true}

// matchScore rates a fork const block for an enum: three per variant matched
// by value and name, one per value alone, plus a bonus when the names are
// related. A block whose name shares no word with the enum's needs two
// specific names to match, or Unknown=0 and a lone value would attach any
// small enum to any block.
func matchScore(enum enumInfo, group forkGroup) int {
	byValue := map[int64][]forkConst{}
	for _, c := range group.Consts {
		byValue[c.Value] = append(byValue[c.Value], c)
	}
	score, named, specific := 0, 0, 0
	for _, variant := range enum.Variants {
		candidates := byValue[variant.Value]
		if len(candidates) == 0 {
			continue
		}
		if _, ok := nameMatch(variant.Name, group.prefix, candidates); ok {
			score += 3
			named++
			if !genericVariants[normalize(variant.Name)] {
				specific++
			}
		} else if len(candidates) == 1 {
			score++
		}
	}
	// The enum's own words come from its type ID as well as its Go name: an
	// anonymous Subtype under SimpleEventPacketPayload is a SimpleEvent.
	words := nameWords(enum.TypeID)
	for word := range nameWords(enum.Name) {
		words[word] = true
	}
	shared := sharedWords(words, nameWords(group.prefix))
	strong := similar(enum.Name, group.prefix) || similar(strings.TrimPrefix(enum.TypeID, "enums/"), group.prefix) || (shared >= 2 && named > 0)
	if named == 0 && !strong {
		return 0
	}
	if shared == 0 && !strong && specific < 2 {
		return 0
	}
	if strong {
		score += 2
	} else if shared > 0 {
		score++
	}
	return score
}

func nameMatch(variant, prefix string, candidates []forkConst) (forkConst, bool) {
	want := normalize(variant)
	for _, c := range candidates {
		suffix := normalize(strings.TrimPrefix(c.Name, prefix))
		if suffix == want || strings.HasSuffix(normalize(c.Name), want) || (want != "" && strings.HasSuffix(want, suffix) && suffix != "") {
			return c, true
		}
	}
	return forkConst{}, false
}

func variantNames(enum enumInfo, group forkGroup) (map[string]string, placementNote) {
	note := placementNote{Enum: enum, Group: group}
	byValue := map[int64][]forkConst{}
	for _, c := range group.Consts {
		byValue[c.Value] = append(byValue[c.Value], c)
	}
	names := map[string]string{}
	taken := map[string]bool{}
	for _, variant := range enum.Variants {
		candidates := byValue[variant.Value]
		var chosen forkConst
		ok := false
		if c, matched := nameMatch(variant.Name, group.prefix, candidates); matched {
			chosen, ok = c, true
		} else if len(candidates) == 1 {
			chosen, ok = candidates[0], true
		}
		if !ok || taken[chosen.Name] {
			// Keep the family naming for a variant the fork does not have.
			if styled := group.prefix + naming.EnumVariantName(variant.Name); group.prefix != "" && !taken[styled] && naming.IsExportedGoIdentifier(styled) {
				taken[styled] = true
				names[variant.Name] = styled
			}
			note.Unnamed = append(note.Unnamed, fmt.Sprintf("%s=%d", variant.Name, variant.Value))
			continue
		}
		taken[chosen.Name] = true
		names[variant.Name] = chosen.Name
	}
	for _, c := range group.Consts {
		if !taken[c.Name] {
			note.Foreign = append(note.Foreign, c)
		}
	}
	return names, note
}

func ownerPackage(owner ownerInfo) string {
	if owner.Packet {
		return "packet"
	}
	return "protocol"
}

// forkTypeByFields matches a type to an unclaimed fork struct in the same
// package whose name shares a word with the type's, when field names agree
// covering two thirds of both sides and no other fork struct scores the
// same. Fields alone are not evidence: NoiseAlignment and a byte metadata
// item are both Type and Value.
func forkTypeByFields(owner ownerInfo, fork forkIndex, claimed map[string]bool) (forkType, bool) {
	if len(owner.Fields) < 2 {
		return forkType{}, false
	}
	words := nameWords(owner.TypeID)
	want := map[string]bool{}
	for _, field := range owner.Fields {
		want[normalize(field.Name)] = true
	}
	var best forkType
	bestShared, ties := 0, 0
	for key, item := range fork.types {
		if !item.Struct || claimed[key] || item.Package != ownerPackage(owner) || len(item.Fields) < 2 || !sharesWord(words, nameWords(item.Name)) {
			continue
		}
		shared := 0
		for _, field := range item.Fields {
			if want[normalize(field.Name)] {
				shared++
			}
		}
		if shared < 2 || shared*3 < len(owner.Fields)*2 || shared*3 < len(item.Fields)*2 {
			continue
		}
		// Two shared names (Type and Value) prove little on their own; then
		// every other field must also pair up by category in order.
		if shared < 3 && !leftoversPair(owner.Fields, item.Fields, want) {
			continue
		}
		switch {
		case shared > bestShared:
			best, bestShared, ties = item, shared, 0
		case shared == bestShared:
			ties++
		}
	}
	return best, bestShared > 0 && ties == 0
}

func forkTypeFor(name, pkg string, fork forkIndex) (forkType, bool) {
	if item, ok := fork.types[pkg+"."+name]; ok && item.Struct {
		return item, true
	}
	want := normalize(name)
	var candidates []forkType
	for _, item := range fork.types {
		if !item.Struct || item.Package != pkg {
			continue
		}
		have := normalize(item.Name)
		if want == have {
			return item, true
		}
		if want == have+"data" || want == have+"info" || want == have+"type" || have == want+"data" || have == want+"event" || want == have+"event" {
			candidates = append(candidates, item)
		}
	}
	if len(candidates) == 1 {
		return candidates[0], true
	}
	return forkType{}, false
}

// nativeTypeIDs are emitted as mgl32, uuid, or plain integers, so they never
// need a fork struct.
var nativeTypeIDs = map[string]bool{"Vec2": true, "Vec3": true, "BlockPos": true, "ActorUniqueID": true, "ActorRuntimeID": true, "PlayerInputTick": true, "mce::Color": true, "mce::UUID": true}

// generic words that most type names carry and so prove nothing shared.
var genericWords = map[string]bool{"data": true, "info": true, "type": true, "payload": true, "packet": true, "definition": true, "entry": true, "config": true, "configuration": true, "setting": true, "shared": true, "anon": true, "json": true, "action": true, "status": true, "state": true, "mode": true, "operation": true, "event": true, "flag": true, "kind": true, "category": true, "source": true, "request": true, "response": true, "result": true, "group": true, "option": true, "value": true, "item": true, "id": true, "enum": true, "cereal": true, "cerealizer": true}

// nameWords splits a type identifier (Go or manifest, namespaces included)
// into lower-case words, dropping generic and version words.
func nameWords(name string) map[string]bool {
	words := map[string]bool{}
	var current []rune
	flush := func() {
		word := normalize(string(current))
		switch {
		case strings.HasSuffix(word, "ies"):
			word = strings.TrimSuffix(word, "ies") + "y"
		case len(word) > 3 && strings.HasSuffix(word, "s") && !strings.HasSuffix(word, "ss"):
			word = strings.TrimSuffix(word, "s")
		}
		if len(word) >= 3 && !genericWords[word] && !strings.HasPrefix(word, "v1") {
			words[word] = true
		}
		current = current[:0]
	}
	runes := []rune(name)
	for i, r := range runes {
		switch {
		case !unicode.IsLetter(r) && !unicode.IsDigit(r):
			flush()
			continue
		case unicode.IsUpper(r) && i > 0 && (unicode.IsLower(runes[i-1]) || (i+1 < len(runes) && unicode.IsLower(runes[i+1]) && unicode.IsUpper(runes[i-1]))):
			flush()
		}
		current = append(current, r)
	}
	flush()
	return words
}

func sharesWord(a, b map[string]bool) bool { return sharedWords(a, b) > 0 }

func sharedWords(a, b map[string]bool) int {
	shared := 0
	for word := range a {
		if b[word] {
			shared++
		}
	}
	return shared
}

// leftoversPair reports whether the fields not shared by name pair up by
// category in order, with none left over on either side.
func leftoversPair(fields []manifest.Field, forkFields []forkField, shared map[string]bool) bool {
	var left []string
	for _, field := range fields {
		if _, ok := forkNames(forkFields)[normalize(field.Name)]; !ok {
			left = append(left, typeCategory(field.Encode))
		}
	}
	var right []string
	for _, field := range forkFields {
		if !shared[normalize(field.Name)] {
			right = append(right, forkCategory(field.Type))
		}
	}
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func forkNames(fields []forkField) map[string]bool {
	names := map[string]bool{}
	for _, field := range fields {
		names[normalize(field.Name)] = true
	}
	return names
}

// forkSpelling applies the fork's naming conventions to a generated field
// name: entities rather than actors, and British spellings.
func forkSpelling(name string) string {
	if !strings.Contains(name, "Entity") {
		name = strings.ReplaceAll(name, "Actor", "Entity")
	}
	for _, pair := range [][2]string{{"Armor", "Armour"}, {"Color", "Colour"}, {"Behavior", "Behaviour"}} {
		name = strings.ReplaceAll(name, pair[0], pair[1])
	}
	return name
}

// identifierName gives an actor identifier field the fork's convention:
// EntityUniqueID and EntityRuntimeID, keeping any leading qualifier.
func identifierName(name, semantic string) string {
	suffix := "EntityUniqueID"
	if semantic == semantics.ActorRuntimeID {
		suffix = "EntityRuntimeID"
	}
	if strings.HasSuffix(name, suffix) {
		return name
	}
	for _, trailing := range []string{"ActorUniqueID", "EntityUniqueID", "ActorRuntimeID", "EntityRuntimeID", "UniqueID", "RuntimeID", "ActorID", "EntityID", "ID"} {
		if strings.HasSuffix(name, trailing) {
			prefix := strings.TrimSuffix(name, trailing)
			prefix = strings.TrimSuffix(strings.TrimSuffix(prefix, "Entity"), "Actor")
			return prefix + suffix
		}
	}
	return name
}

// fieldSemantic reports the actor identifier a field carries, from its node
// or from a recorded semantics entry.
func (r *gapReport) fieldSemantic(owner ownerInfo, field manifest.Field) string {
	if field.Encode.Kind == manifest.KindStruct {
		switch field.Encode.TypeID {
		case semantics.ActorUniqueID, semantics.ActorRuntimeID:
			return field.Encode.TypeID
		}
	}
	for _, entry := range r.semantics {
		if entry.TypeID == owner.TypeID && entry.Field == field.Name {
			return entry.Semantic
		}
	}
	return ""
}

// conventionEntries renames fields the fork does not name so they still read
// like the fork's; matched is the set of wire names that already have a fork
// name.
func (r *gapReport) conventionEntries(owner ownerInfo, matched map[string]bool) []layout.FieldEntry {
	var entries []layout.FieldEntry
	for _, field := range owner.Fields {
		if matched[field.Name] {
			continue
		}
		generated := naming.GoExportName(field.Name)
		styled := forkSpelling(generated)
		if semantic := r.fieldSemantic(owner, field); semantic != "" {
			styled = identifierName(styled, semantic)
		}
		if styled != generated {
			entries = append(entries, layout.FieldEntry{TypeID: owner.TypeID, Field: field.Name, Name: styled, Rationale: "gophertunnel naming convention for a field it does not name."})
		}
	}
	return entries
}

func matchFields(owner ownerInfo, source forkType, gap *fieldGap, docOverlay docs.Overlay, report *gapReport) []layout.FieldEntry {
	var entries []layout.FieldEntry
	usedFork := map[int]bool{}
	matched := map[int]int{} // manifest index -> fork index
	// Actor identifiers pair by wire order when both sides carry the same
	// number of them (the fork's EntityUniqueID is the wire's second unique ID
	// in UpdateTrade), so they are kept out of the name passes until then.
	for _, semantic := range []string{semantics.ActorUniqueID, semantics.ActorRuntimeID} {
		var wire, fork []int
		for i, field := range owner.Fields {
			if report.fieldSemantic(owner, field) == semantic {
				wire = append(wire, i)
			}
		}
		for j, forkField := range source.Fields {
			if forkCarries(forkField, semantic) {
				fork = append(fork, j)
			}
		}
		if len(wire) == len(fork) {
			for k := range wire {
				matched[wire[k]], usedFork[fork[k]] = fork[k], true
			}
		}
	}
	for i, field := range owner.Fields {
		if _, ok := matched[i]; ok {
			continue
		}
		for j, forkField := range source.Fields {
			if usedFork[j] {
				continue
			}
			if normalize(field.Name) == normalize(forkField.Name) || normalize(naming.GoExportName(field.Name)) == normalize(forkField.Name) {
				matched[i], usedFork[j] = j, true
				break
			}
		}
	}
	// Names that agree once a trailing Data/List/Info or plural is dropped
	// (Enum Data for Enums), then names one of which starts or ends with the
	// other (ShouldTrackOutput for Track Output); each only when the pair is
	// the sole candidate for both sides and the categories agree.
	for _, pass := range []func(want, have string) bool{
		func(want, have string) bool { return stem(want) == stem(have) },
		contains,
	} {
		pairUnique(owner, source, matched, usedFork, func(field manifest.Field, forkField forkField) bool {
			return pass(normalize(field.Name), normalize(forkField.Name)) && typeCategory(field.Encode) == forkCategory(forkField.Type)
		})
	}
	// Leftover fields are aligned in wire order, but only when every leftover
	// on both sides pairs up by category; a partial alignment shifts names
	// onto the wrong fields.
	if len(matched) < len(owner.Fields) {
		var left, right []int
		for i := range owner.Fields {
			if _, ok := matched[i]; !ok {
				left = append(left, i)
			}
		}
		for j := range source.Fields {
			if !usedFork[j] {
				right = append(right, j)
			}
		}
		pairs := alignByCategory(owner.Fields, source.Fields, left, right)
		if len(pairs) == len(left) && len(pairs) == len(right) {
			for _, pairIndex := range pairs {
				matched[pairIndex[0]], usedFork[pairIndex[1]] = pairIndex[1], true
			}
			gap.Positional = true
		}
	}
	named := map[string]bool{}
	for i, field := range owner.Fields {
		j, ok := matched[i]
		if !ok {
			gap.OnlyManifest = append(gap.OnlyManifest, field.Name)
			continue
		}
		forkField := source.Fields[j]
		if typeCategory(field.Encode) != forkCategory(forkField.Type) {
			gap.TypeMismatch = append(gap.TypeMismatch, fmt.Sprintf("%s (%s vs %s)", forkField.Name, describe(field.Encode), forkField.Type))
		}
		goName := naming.GoExportName(field.Name)
		if goName != forkField.Name && naming.IsExportedGoIdentifier(forkField.Name) {
			entries = append(entries, layout.FieldEntry{TypeID: owner.TypeID, Field: field.Name, Name: forkField.Name, Rationale: fmt.Sprintf("gophertunnel %s.%s field name.", source.Name, forkField.Name)})
			goName = forkField.Name
		}
		named[field.Name] = true
		if forkField.Semantic != "" {
			report.semantic(owner, field, forkField.Semantic, fmt.Sprintf("gophertunnel marshals %s.%s with an %s operation.", source.Name, forkField.Name, forkField.Semantic))
		}
		if docOverlay.Fields != nil && forkField.Doc != "" {
			key := docs.FieldKey(owner.TypeID, field.Name)
			if docOverlay.Fields[key] == "" {
				docOverlay.Fields[key] = docs.LeadWith(forkField.Doc, forkField.Name, goName)
				report.portedDocs++
			}
		}
	}
	report.semanticsByName(owner, named)
	entries = append(entries, report.conventionEntries(owner, named)...)
	for j, forkField := range source.Fields {
		if !usedFork[j] {
			gap.OnlyFork = append(gap.OnlyFork, forkField.Name+" "+forkField.Type)
		}
	}
	return entries
}

// alignByCategory pairs leftover manifest and fork fields by longest common
// subsequence over their type categories, preserving wire order.
func alignByCategory(fields []manifest.Field, source []forkField, left, right []int) [][2]int {
	n, m := len(left), len(right)
	if n == 0 || m == 0 {
		return nil
	}
	lcs := make([][]int, n+1)
	for i := range lcs {
		lcs[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if typeCategory(fields[left[i]].Encode) == forkCategory(source[right[j]].Type) {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else if lcs[i+1][j] >= lcs[i][j+1] {
				lcs[i][j] = lcs[i+1][j]
			} else {
				lcs[i][j] = lcs[i][j+1]
			}
		}
	}
	var pairs [][2]int
	for i, j := 0, 0; i < n && j < m; {
		switch {
		case typeCategory(fields[left[i]].Encode) == forkCategory(source[right[j]].Type) && lcs[i][j] == lcs[i+1][j+1]+1:
			pairs = append(pairs, [2]int{left[i], right[j]})
			i++
			j++
		case lcs[i+1][j] >= lcs[i][j+1]:
			i++
		default:
			j++
		}
	}
	return pairs
}

// pairUnique matches every unmatched wire field to an unused fork field when
// the predicate holds for that pair alone on both sides.
func pairUnique(owner ownerInfo, source forkType, matched map[int]int, usedFork map[int]bool, accept func(manifest.Field, forkField) bool) {
	wireCandidates := map[int][]int{}
	forkCandidates := map[int][]int{}
	for i, field := range owner.Fields {
		if _, ok := matched[i]; ok {
			continue
		}
		for j, forkField := range source.Fields {
			if !usedFork[j] && accept(field, forkField) {
				wireCandidates[i] = append(wireCandidates[i], j)
				forkCandidates[j] = append(forkCandidates[j], i)
			}
		}
	}
	for i, candidates := range wireCandidates {
		if len(candidates) == 1 && len(forkCandidates[candidates[0]]) == 1 {
			matched[i], usedFork[candidates[0]] = candidates[0], true
		}
	}
}

// contains reports names one of which ends with the other, or starts with
// the other and continues only with a generic word (Primary Effect Id for
// PrimaryEffect, but not Attribute Layer Dimension for Attribute).
func contains(want, have string) bool {
	if len(want) < 4 || len(have) < 4 {
		return false
	}
	if strings.HasSuffix(have, want) || strings.HasSuffix(want, have) {
		return true
	}
	for _, pair := range [][2]string{{want, have}, {have, want}} {
		if strings.HasPrefix(pair[0], pair[1]) {
			switch pair[0][len(pair[1]):] {
			case "id", "ids", "data", "list", "info", "type", "value", "values", "s", "es":
				return true
			}
		}
	}
	return false
}

// forkCarries reports whether a fork field holds an actor identifier, by the
// IO operation that marshals it or, failing that, by its name.
func forkCarries(field forkField, semantic string) bool {
	if field.Semantic != "" {
		return field.Semantic == semantic
	}
	if forkCategory(field.Type) != "integer" && forkCategory(field.Type) != "optional" {
		return false
	}
	switch semantic {
	case semantics.ActorUniqueID:
		return strings.HasSuffix(field.Name, "UniqueID") || strings.HasSuffix(field.Name, "EntityID")
	case semantics.ActorRuntimeID:
		return strings.HasSuffix(field.Name, "RuntimeID")
	}
	return false
}

// stem drops the trailing Data/List/Info and plural markers a name may carry
// on one side only; a stem shorter than three letters is not compared.
func stem(name string) string {
	for changed := true; changed; {
		changed = false
		for _, suffix := range []string{"data", "list", "info", "s"} {
			if trimmed := strings.TrimSuffix(name, suffix); trimmed != name && len(trimmed) >= 3 {
				name, changed = trimmed, true
			}
		}
	}
	if len(name) < 3 {
		return "\x00" + name
	}
	return name
}

// semantic records an actor identifier field when its wire encoding is a
// plain integer the identifier operations can carry.
func (r *gapReport) semantic(owner ownerInfo, field manifest.Field, semantic, rationale string) {
	typeID, name, node := owner.TypeID, field.Name, field.Encode
	// The fork marks the value; the manifest may wrap it in an optional or a
	// one-field struct, so the entry keys the integer leaf.
	for {
		if node.Kind == manifest.KindStruct && (node.TypeID == semantics.ActorUniqueID || node.TypeID == semantics.ActorRuntimeID) {
			return // already an identifier
		}
		if node.Kind == manifest.KindOptional && node.Value != nil {
			node = *node.Value
			continue
		}
		if node.Kind == manifest.KindStruct && len(node.Fields) == 1 && node.TypeID != "" {
			typeID, name, node = flatten.TypeID(node), node.Fields[0].Name, node.Fields[0].Encode
			continue
		}
		break
	}
	if node.Kind != manifest.KindPrimitive || node.Primitive == nil || !semantics.Carries(semantic, node.Primitive.Code) {
		return
	}
	for _, entry := range r.semantics {
		if entry.TypeID == typeID && entry.Field == name {
			return
		}
	}
	r.semantics = append(r.semantics, semantics.Entry{TypeID: typeID, Field: name, Semantic: semantic, Rationale: rationale})
}

// semanticsByName marks unmatched fields whose wire name says they are actor
// identifiers; block, item, and pack IDs are excluded.
func (r *gapReport) semanticsByName(owner ownerInfo, matched map[string]bool) {
	for _, field := range owner.Fields {
		if matched[field.Name] || field.Encode.Kind != manifest.KindPrimitive {
			continue
		}
		name := normalize(field.Name)
		if nonActorName(name) {
			continue
		}
		if !strings.Contains(name, "actor") && !strings.Contains(name, "player") && nonActorName(normalize(owner.Name)) {
			continue
		}
		semantic := ""
		switch {
		case strings.HasSuffix(name, "uniqueid"):
			semantic = semantics.ActorUniqueID
		case strings.HasSuffix(name, "runtimeid"):
			semantic = semantics.ActorRuntimeID
		}
		if semantic == "" {
			continue
		}
		before := len(r.semantics)
		r.semantic(owner, field, semantic, "Wire field name names an actor identifier; not confirmed by the fork.")
		if len(r.semantics) > before {
			r.heuristicSemantics++
		}
	}
}

// nonActorName reports a name whose "runtime ID" or "unique ID" is not an
// actor's (block runtime IDs, item network IDs, map IDs).
func nonActorName(name string) bool {
	for _, word := range []string{"block", "item", "recipe", "pack", "biome", "map"} {
		if strings.Contains(name, word) {
			return true
		}
	}
	return false
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
		case "uuid", "nbt_le":
			return "named"
		default:
			return "integer"
		}
	case manifest.KindEnum:
		// gophertunnel keeps enums as plain integers.
		return "integer"
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
	case manifest.KindStruct:
		// Actor identifiers and input ticks are plain integers in Go.
		switch node.TypeID {
		case "ActorUniqueID", "ActorRuntimeID", "PlayerInputTick":
			return "integer"
		}
		return "named"
	default:
		return "named"
	}
}

func forkCategory(value string) string {
	value = strings.TrimSpace(value)
	switch {
	case strings.HasPrefix(value, "[]"):
		return "slice"
	case strings.HasPrefix(value, "["):
		return "array"
	case strings.HasPrefix(value, "map["):
		return "map"
	case strings.Contains(value, "Optional["):
		return "optional"
	}
	switch value {
	case "bool":
		return "bool"
	case "float32":
		return "float32"
	case "float64":
		return "float64"
	case "int8", "int16", "int32", "int64", "int", "uint8", "uint16", "uint32", "uint64", "uint", "byte":
		return "integer"
	case "string":
		return "string"
	default:
		return "named"
	}
}

func describe(node manifest.Node) string {
	switch node.Kind {
	case manifest.KindPrimitive:
		if node.Primitive != nil {
			return node.Primitive.Code
		}
	case manifest.KindEnum:
		if node.Primitive != nil {
			return "enum " + node.Primitive.Code
		}
	case manifest.KindOptional:
		if node.Value != nil {
			return "optional " + describe(*node.Value)
		}
	case manifest.KindArray:
		if node.Element != nil {
			return "[]" + describe(*node.Element)
		}
	}
	if node.TypeID != "" {
		return node.TypeID
	}
	return string(node.Kind)
}

func similar(a, b string) bool {
	x, y := normalize(a), normalize(b)
	return x != "" && y != "" && (strings.Contains(x, y) || strings.Contains(y, x))
}

// normalize folds case, punctuation, and the British spellings gophertunnel
// uses so names compare by meaning.
func normalize(value string) string {
	var b strings.Builder
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
		}
	}
	result := b.String()
	for _, pair := range [][2]string{{"armour", "armor"}, {"colour", "color"}, {"behaviour", "behavior"}, {"centre", "center"}, {"entity", "actor"}, {"initialised", "initialized"}, {"serialised", "serialized"}, {"synchronised", "synchronized"}} {
		result = strings.ReplaceAll(result, pair[0], pair[1])
	}
	return result
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// --- report -------------------------------------------------------------------

func (r *gapReport) render(m manifest.Manifest, fork forkIndex) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# gophertunnel layout gaps for Minecraft %s / protocol %d\n\n", m.Target.MinecraftVersion, m.Target.ProtocolVersion)
	fmt.Fprintf(&b, "Fork checkout `%s`. Generated by `tools/seed-gophertunnel-layout`; the numbers describe what the layout overlay could map automatically, not wire correctness.\n\n", fork.commit)
	fmt.Fprintf(&b, "| | count |\n|---|---:|\n")
	fmt.Fprintf(&b, "| enums placed onto a fork const block | %d |\n", len(r.placements))
	fmt.Fprintf(&b, "| enums with no fork const block | %d |\n", len(r.unmatchedEnums))
	fmt.Fprintf(&b, "| fork const blocks with no enum | %d |\n", len(r.unmatchedGroups))
	fmt.Fprintf(&b, "| fork struct types with no generated type | %d |\n", len(r.forkOnlyTypes))
	fmt.Fprintf(&b, "| generated types with no fork type | %d |\n", len(r.generatedOnlyTypes))
	fmt.Fprintf(&b, "| matched types with field differences | %d |\n\n", len(r.fieldGaps))

	b.WriteString("## Fork struct types with no generated counterpart\n\nHand-written decoding or a name the overlay does not map yet.\n\n")
	for _, item := range r.forkOnlyTypes {
		fmt.Fprintf(&b, "- `%s.%s` (%s.go, %d fields)\n", item.Package, item.Name, item.File, len(item.Fields))
	}
	b.WriteString("\n## Generated types with no fork counterpart\n\n")
	for _, owner := range r.generatedOnlyTypes {
		fmt.Fprintf(&b, "- `%s` (%s, %d fields)\n", owner.Name, owner.TypeID, len(owner.Fields))
	}
	b.WriteString("\n## Field differences on matched types\n\n")
	for _, gap := range r.fieldGaps {
		fmt.Fprintf(&b, "### `%s` vs fork `%s.%s`", gap.Owner.Name, gap.Fork.Package, gap.Fork.Name)
		if gap.Positional {
			b.WriteString(" (some fields matched by position)")
		}
		fmt.Fprintf(&b, "\n\n- renamed to fork names: %d\n", gap.Renamed)
		if len(gap.OnlyManifest) > 0 {
			fmt.Fprintf(&b, "- only in manifest: %s\n", strings.Join(quote(gap.OnlyManifest), ", "))
		}
		if len(gap.OnlyFork) > 0 {
			fmt.Fprintf(&b, "- only in fork: %s\n", strings.Join(quote(gap.OnlyFork), ", "))
		}
		if len(gap.TypeMismatch) > 0 {
			fmt.Fprintf(&b, "- type category differs: %s\n", strings.Join(quote(gap.TypeMismatch), ", "))
		}
		b.WriteString("\n")
	}
	b.WriteString("## Constant placements needing review\n\nVariants the fork block does not name keep their generated name; fork constants with no variant are listed as foreign.\n\n")
	for _, note := range r.placements {
		if len(note.Unnamed) == 0 && len(note.Foreign) == 0 {
			continue
		}
		fmt.Fprintf(&b, "- `%s` -> `%s.%s` block `%s*` (score %d)", note.Enum.Name, note.Group.Package, note.Group.File, note.Group.prefix, note.Score)
		if len(note.Unnamed) > 0 {
			fmt.Fprintf(&b, "; unnamed: %s", strings.Join(note.Unnamed, ", "))
		}
		if len(note.Foreign) > 0 {
			names := make([]string, 0, len(note.Foreign))
			for _, c := range note.Foreign {
				names = append(names, fmt.Sprintf("%s=%d", c.Name, c.Value))
			}
			fmt.Fprintf(&b, "; foreign: %s", strings.Join(names, ", "))
		}
		b.WriteString("\n")
	}
	b.WriteString("\n## Fork type names that could not be taken\n\nThe name is already carried by another generated type or packet.\n\n")
	for _, note := range r.skippedTypeNames {
		fmt.Fprintf(&b, "- %s\n", note)
	}
	b.WriteString("\n## Enums whose inferred identity is shared\n\nTwo anonymous enums infer the same name with different values; give them type IDs before placing constants.\n\n")
	for _, enum := range r.ambiguousEnums {
		fmt.Fprintf(&b, "- `%s` (%s)\n", enum.Name, enum.TypeID)
	}
	b.WriteString("\n## Enums with no fork const block\n\n")
	for _, enum := range r.unmatchedEnums {
		fmt.Fprintf(&b, "- `%s` (%s, %d variants)\n", enum.Name, enum.TypeID, len(enum.Variants))
	}
	b.WriteString("\n## Fork const blocks with no enum\n\n")
	for _, group := range r.unmatchedGroups {
		fmt.Fprintf(&b, "- `%s.%s` `%s*` (%d constants)\n", group.Package, group.File, group.prefix, len(group.Consts))
	}
	return b.String()
}

func quote(values []string) []string {
	result := make([]string, len(values))
	for i, value := range values {
		result[i] = "`" + value + "`"
	}
	return result
}
