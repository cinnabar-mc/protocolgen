package gophertunneloracle

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const extractionDepthLimit = 40

// reviewedHelperRevisions are the gophertunnel commits whose hand-written
// helper bodies, private writer methods and interface variant tables were
// reviewed; the extractor expands them only at these commits.
var reviewedHelperRevisions = map[string]bool{
	"77149b94f8024826e34efda6f625c366596118d2": true, // protocol 2225
}

type typeRef struct {
	Kind string // named, primitive, slice, array, optional, pointer, unknown
	Name string
	Elem *typeRef
	Len  int
	Path string
	// WireCode records an explicit primitive write used as an external
	// length. It is deliberately data-flow metadata, not a guessed type.
	WireCode string
	Function *ast.FuncLit
}

type typeInfo struct {
	Key      string
	Name     string
	Pkg      string
	Fields   map[string]typeRef
	Embedded []typeRef
}

type sourceFile struct {
	Path    string
	Rel     string
	Pkg     string
	AST     *ast.File
	Imports map[string]string
}

type marshalInfo struct {
	Key  string
	Recv string
	IO   string
	Decl *ast.FuncDecl
	File *sourceFile
}

type extractor struct {
	root        string
	fset        *token.FileSet
	files       []*sourceFile
	types       map[string]*typeInfo
	arrays      map[string]typeRef // named fixed-size array types, by package-qualified name
	marshals    map[string]*marshalInfo
	functions   map[string]*marshalInfo
	ioHelpers   map[string]*marshalInfo
	consts      map[string]int
	recursion   map[string]int // re-entries per local helper, bounded by recursionUnroll
	diagnostics []diagnostic
	packet      string
	revision    string
}

var sourcePrimitive = map[string]string{
	"Uint8":                   "u8",
	"Int8":                    "i8",
	"Uint16":                  "u16le",
	"Int16":                   "i16le",
	"Uint32":                  "u32le",
	"Int32":                   "i32le",
	"BEInt32":                 "i32be",
	"BEARGB":                  "argb32", // channel-swapped big-endian write; same bytes as a little-endian ARGB int
	"Uint64":                  "u64le",
	"Int64":                   "i64le",
	"Float32":                 "f32le",
	"Float64":                 "f64le",
	"Bool":                    "bool",
	"Varint32":                "zigzag_i32",
	"Varuint32":               "var_u32",
	"Varint64":                "zigzag_i64",
	"Varuint64":               "var_u64",
	"ActorRuntimeID":          "var_u64",
	"ActorRuntimeIDVarint64":  "zigzag_i64",
	"ActorRuntimeIDVaruint32": "var_u32",
	"ActorUniqueID":           "zigzag_i64",
	"ActorUniqueIDInt64":      "i64le",
	"ActorUniqueIDUint64":     "u64le",
	"ActorUniqueIDVaruint64":  "var_u64",
	"PlayerInputTick":         "var_u64",
}

var sourceBuiltinPrimitive = map[string]string{
	"byte": "u8", "uint8": "u8", "int8": "i8",
	"uint16": "u16le", "int16": "i16le",
	"uint32": "u32le", "int32": "i32le",
	"uint64": "u64le", "int64": "i64le",
	"float32": "f32le", "float64": "f64le", "bool": "bool",
}

var sourceArrayPrefixes = map[string]string{
	"Slice":                   "var_u32",
	"FuncSlice":               "var_u32",
	"FuncIOSlice":             "var_u32",
	"SliceOfFunc":             "var_u32",
	"SliceUint8Length":        "u8",
	"SliceUint16Length":       "u16le",
	"SliceUint32Length":       "u32le",
	"SliceVarint32Length":     "zigzag_i32",
	"FuncSliceUint8Length":    "u8",
	"FuncSliceUint16Length":   "u16le",
	"FuncSliceUint32Length":   "u32le",
	"FuncSliceVarint32Length": "zigzag_i32",
}

// These are the only handwritten IO helpers whose bodies are admitted into
// the oracle. The allowlist is deliberately explicit: a newly added upstream
// helper remains unresolved until its wire summary is reviewed and pinned to
// the source revision used by the lock file.
var reviewedIOHelpers = map[string]bool{
	"AbilityValue":          true,
	"EntityMetadata":        true,
	"EventOrdinal":          true,
	"EventType":             true,
	"GameRule":              true,
	"Item":                  true,
	"ItemDescriptorCount":   true,
	"ItemInstance":          true,
	"PackSetting":           true,
	"PlayerInventoryAction": true,
	"RGB":                   true,
	"RGBA":                  true,
	"ShapeData":             true,
	"StackRequestAction":    true,
	"StackRequestItem":      true,
	"MaterialReducer":       true,
	"TransactionDataType":   true,
	"InputFlagList":         true,
	"CommandOriginData":     true,
	"PlayerMoveSettings":    true,
}

var reviewedProtocolHelpers = map[string]bool{
	"CommandOriginData":               true,
	"InputFlagList":                   true,
	"PlayerMoveSettings":              true,
	"StackRequestItemDescriptorCount": true,
}

// reviewedHelperMethods are private Writer/Reader methods whose callers are
// part of an admitted helper summary. They are kept separate from the public
// IO surface so an arbitrary implementation detail cannot silently become a
// new oracle rule.
var reviewedHelperMethods = map[string]bool{
	"itemUserData": true,
}

type reviewedInterfaceVariant struct {
	Value int64
	Type  string
}

// These dynamic interfaces are selected by a reviewed discriminator helper
// before Marshal is invoked. The values are the on-wire variant values at the
// pinned revision, not Go's internal type IDs.
var reviewedInterfaceVariants = map[string][]reviewedInterfaceVariant{
	"protocol.InventoryTransactionData": {
		{Value: 0, Type: "protocol.NormalTransactionData"},
		{Value: 1, Type: "protocol.MismatchTransactionData"},
		{Value: 2, Type: "protocol.UseItemTransactionData"},
		{Value: 3, Type: "protocol.UseItemOnEntityTransactionData"},
		{Value: 4, Type: "protocol.ReleaseItemTransactionData"},
	},
	"protocol.StackRequestAction": {
		{Value: 0, Type: "protocol.TakeStackRequestAction"},
		{Value: 1, Type: "protocol.PlaceStackRequestAction"},
		{Value: 2, Type: "protocol.SwapStackRequestAction"},
		{Value: 3, Type: "protocol.DropStackRequestAction"},
		{Value: 4, Type: "protocol.DestroyStackRequestAction"},
		{Value: 5, Type: "protocol.ConsumeStackRequestAction"},
		{Value: 6, Type: "protocol.CreateStackRequestAction"},
		{Value: 7, Type: "protocol.LabTableCombineStackRequestAction"},
		{Value: 8, Type: "protocol.BeaconPaymentStackRequestAction"},
		{Value: 9, Type: "protocol.MineBlockStackRequestAction"},
		{Value: 10, Type: "protocol.CraftRecipeStackRequestAction"},
		{Value: 11, Type: "protocol.AutoCraftRecipeStackRequestAction"},
		{Value: 12, Type: "protocol.CraftCreativeStackRequestAction"},
		{Value: 13, Type: "protocol.CraftRecipeOptionalStackRequestAction"},
		{Value: 14, Type: "protocol.CraftGrindstoneRecipeStackRequestAction"},
		{Value: 15, Type: "protocol.CraftLoomRecipeStackRequestAction"},
		{Value: 16, Type: "protocol.CraftReservedStackRequestAction"},
		{Value: 17, Type: "protocol.CraftNonImplementedStackRequestAction"},
		{Value: 18, Type: "protocol.CraftResultsDeprecatedStackRequestAction"},
	},
	"protocol.Event": {
		{Value: 0, Type: "protocol.AchievementAwardedEvent"},
		{Value: 1, Type: "protocol.EntityInteractEvent"},
		{Value: 2, Type: "protocol.PortalBuiltEvent"},
		{Value: 3, Type: "protocol.PortalUsedEvent"},
		{Value: 4, Type: "protocol.MobKilledEvent"},
		{Value: 5, Type: "protocol.CauldronUsedEvent"},
		{Value: 6, Type: "protocol.PlayerDiedEvent"},
		{Value: 7, Type: "protocol.BossKilledEvent"},
		{Value: 8, Type: "protocol.AgentCommandEvent"},
		{Value: 9, Type: "protocol.AgentCreatedEvent"},
		{Value: 10, Type: "protocol.PatternRemovedEvent"},
		{Value: 11, Type: "protocol.SlashCommandExecutedEvent"},
		{Value: 12, Type: "protocol.FishBucketedEvent"},
		{Value: 13, Type: "protocol.MobBornEvent"},
		{Value: 14, Type: "protocol.PetDiedEvent"},
		{Value: 15, Type: "protocol.CauldronInteractEvent"},
		{Value: 16, Type: "protocol.ComposterInteractEvent"},
		{Value: 17, Type: "protocol.BellUsedEvent"},
		{Value: 18, Type: "protocol.EntityDefinitionTriggerEvent"},
		{Value: 19, Type: "protocol.RaidUpdateEvent"},
		{Value: 20, Type: "protocol.MovementAnomalyEvent"},
		{Value: 21, Type: "protocol.MovementCorrectedEvent"},
		{Value: 22, Type: "protocol.ExtractHoneyEvent"},
		{Value: 23, Type: "protocol.TargetBlockHitEvent"},
		{Value: 24, Type: "protocol.PiglinBarterEvent"},
		{Value: 25, Type: "protocol.WaxedOrUnwaxedCopperEvent"},
		{Value: 26, Type: "protocol.CodeBuilderRuntimeActionEvent"},
		{Value: 27, Type: "protocol.CodeBuilderScoreboardEvent"},
		{Value: 28, Type: "protocol.StriderRiddenInLavaInOverworldEvent"},
		{Value: 29, Type: "protocol.SneakCloseToSculkSensorEvent"},
		{Value: 30, Type: "protocol.CarefulRestorationEvent"},
		{Value: 31, Type: "protocol.ItemUsedEvent"},
	},
	"protocol.ItemDescriptor": {
		{Value: 0, Type: "protocol.InvalidItemDescriptor"},
		{Value: 1, Type: "protocol.DefaultItemDescriptor"},
		{Value: 2, Type: "protocol.MoLangItemDescriptor"},
		{Value: 3, Type: "protocol.ItemTagItemDescriptor"},
	},
	"protocol.ShapeData": {
		{Value: 0, Type: "protocol.LastShape"},
		{Value: 1, Type: "protocol.ArrowShape"},
		{Value: 2, Type: "protocol.TextShape"},
		{Value: 3, Type: "protocol.BoxShape"},
		{Value: 4, Type: "protocol.LineShape"},
		{Value: 5, Type: "protocol.SphereShape"},
		{Value: 6, Type: "protocol.CylinderShape"},
		{Value: 7, Type: "protocol.PyramidShape"},
		{Value: 8, Type: "protocol.EllipsoidShape"},
		{Value: 9, Type: "protocol.ConeShape"},
	},
}

func Extract(root string) (extraction, error) {
	return ExtractAtRevision(root, "")
}

// ExtractAtRevision enables only helper summaries reviewed against the exact
// pinned source revision. Plain Extract remains useful for synthetic fixtures
// and intentionally does not enable revision-bound summaries.
func ExtractAtRevision(root, revision string) (extraction, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return extraction{}, fmt.Errorf("resolve gophertunnel root: %w", err)
	}
	e := &extractor{
		root:      abs,
		fset:      token.NewFileSet(),
		types:     map[string]*typeInfo{},
		arrays:    map[string]typeRef{},
		marshals:  map[string]*marshalInfo{},
		functions: map[string]*marshalInfo{},
		ioHelpers: map[string]*marshalInfo{},
		consts:    map[string]int{},
		recursion: map[string]int{},
		revision:  revision,
	}
	if err := e.load(); err != nil {
		return extraction{}, err
	}
	ids, packetTypes := e.packetIDs()
	keys := make([]string, 0, len(packetTypes))
	for key := range packetTypes {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	packets := make([]sourcePacket, 0, len(keys))
	seenIDs := map[uint32]string{}
	for _, key := range keys {
		idName := packetTypes[key]
		id, ok := ids[idName]
		if !ok {
			e.packet = shortType(key)
			e.addDiagnostic(nil, key, "", "packet ID constant could not be evaluated", idName)
			continue
		}
		if id < 0 || uint64(id) > 1<<32-1 {
			return extraction{}, fmt.Errorf("invalid packet ID %d for %s", id, key)
		}
		if previous, ok := seenIDs[uint32(id)]; ok {
			return extraction{}, fmt.Errorf("duplicate packet ID %d for %s and %s", id, previous, key)
		}
		seenIDs[uint32(id)] = key
		e.packet = shortType(key)
		ops := e.expandType(key, e.packet, 0, map[string]bool{})
		packets = append(packets, sourcePacket{ID: uint32(id), Name: e.packet, Operations: ops})
	}
	sort.Slice(packets, func(i, j int) bool {
		if packets[i].ID != packets[j].ID {
			return packets[i].ID < packets[j].ID
		}
		return packets[i].Name < packets[j].Name
	})
	return extraction{Packets: packets, Diagnostics: e.diagnostics}, nil
}

func (e *extractor) load() error {
	base := filepath.Join(e.root, "minecraft", "protocol")
	if info, err := os.Stat(base); err != nil || !info.IsDir() {
		return fmt.Errorf("gophertunnel root has no minecraft/protocol directory: %s", e.root)
	}
	err := filepath.WalkDir(base, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(e.fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		rel, err := filepath.Rel(base, path)
		if err != nil {
			return err
		}
		sf := &sourceFile{Path: path, Rel: filepath.ToSlash(rel), Pkg: file.Name.Name, AST: file, Imports: map[string]string{}}
		for _, imp := range file.Imports {
			importPath, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				continue
			}
			alias := filepath.Base(importPath)
			if imp.Name != nil {
				alias = imp.Name.Name
			}
			sf.Imports[alias] = filepath.Base(importPath)
		}
		e.files = append(e.files, sf)
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk gophertunnel protocol source: %w", err)
	}
	sort.Slice(e.files, func(i, j int) bool { return e.files[i].Rel < e.files[j].Rel })
	for _, sf := range e.files {
		e.collectConstants(sf)
	}
	for _, sf := range e.files {
		e.collectTypes(sf)
	}
	for _, sf := range e.files {
		e.collectMethods(sf)
	}
	return nil
}

func (e *extractor) collectConstants(sf *sourceFile) {
	for _, decl := range sf.AST.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		var previous []ast.Expr
		for i, spec := range gen.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			if len(valueSpec.Values) > 0 {
				previous = valueSpec.Values
			}
			for index, name := range valueSpec.Names {
				if name.Name == "_" || index >= len(previous) {
					continue
				}
				if value, ok := e.evalConst(previous[index], i, sf.Pkg); ok {
					e.consts[sf.Pkg+"."+name.Name] = value
				}
			}
		}
	}
}

func (e *extractor) collectTypes(sf *sourceFile) {
	for _, decl := range sf.AST.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if array, ok := ts.Type.(*ast.ArrayType); ok && array.Len != nil {
				e.arrays[sf.Pkg+"."+ts.Name.Name] = e.parseType(array, sf)
				continue
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				continue
			}
			key := sf.Pkg + "." + ts.Name.Name
			info := &typeInfo{Key: key, Name: ts.Name.Name, Pkg: sf.Pkg, Fields: map[string]typeRef{}}
			for _, field := range st.Fields.List {
				ref := e.parseType(field.Type, sf)
				if len(field.Names) == 0 {
					info.Embedded = append(info.Embedded, ref)
					info.Fields[shortType(ref.Name)] = ref
					continue
				}
				for _, name := range field.Names {
					info.Fields[name.Name] = ref
				}
			}
			e.types[key] = info
		}
	}
}

func (e *extractor) collectMethods(sf *sourceFile) {
	for _, decl := range sf.AST.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		if fn.Recv == nil {
			info := &marshalInfo{Key: sf.Pkg + "." + fn.Name.Name, Decl: fn, File: sf}
			info.IO = ioParameter(fn, sf)
			e.functions[info.Key] = info
			continue
		}
		recvType, recvVar := receiver(fn)
		if recvType == "" {
			continue
		}
		if fn.Name.Name != "Marshal" {
			if sf.Pkg == "protocol" && (recvType == "Writer" || recvType == "Reader") {
				key := sf.Pkg + "." + recvType + "." + fn.Name.Name
				e.ioHelpers[key] = &marshalInfo{Key: key, Recv: recvVar, IO: recvVar, Decl: fn, File: sf}
			}
			continue
		}
		ioName := ""
		if fn.Type.Params != nil && len(fn.Type.Params.List) > 0 && len(fn.Type.Params.List[0].Names) > 0 {
			ioName = fn.Type.Params.List[0].Names[0].Name
		}
		key := sf.Pkg + "." + recvType
		e.marshals[key] = &marshalInfo{Key: key, Recv: recvVar, IO: ioName, Decl: fn, File: sf}
	}
}

func ioParameter(fn *ast.FuncDecl, sf *sourceFile) string {
	if fn.Type.Params == nil {
		return ""
	}
	for _, param := range fn.Type.Params.List {
		ref := parseTypeForParameter(param.Type, sf)
		if isIORef(ref) && len(param.Names) > 0 {
			return param.Names[0].Name
		}
	}
	return ""
}

func parseTypeForParameter(expr ast.Expr, sf *sourceFile) typeRef {
	// The extractor has not been constructed when this helper is used during
	// method collection, so handle only the IO reference needed to identify
	// local helper functions. Full type parsing happens through extractor.parseType.
	if selector, ok := expr.(*ast.SelectorExpr); ok {
		return typeRef{Kind: "named", Name: sf.Imports[identName(selector.X)] + "." + selector.Sel.Name}
	}
	if ident, ok := expr.(*ast.Ident); ok {
		return typeRef{Kind: "named", Name: sf.Pkg + "." + ident.Name}
	}
	return typeRef{Kind: "unknown"}
}

func identName(expr ast.Expr) string {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name
	}
	return ""
}

func (e *extractor) packetIDs() (map[string]int, map[string]string) {
	ids := map[string]int{}
	for name, value := range e.consts {
		if strings.HasPrefix(name, "packet.ID") {
			ids[strings.TrimPrefix(name, "packet.")] = value
		}
	}
	packetTypes := map[string]string{}
	for _, sf := range e.files {
		if sf.Pkg != "packet" {
			continue
		}
		for _, decl := range sf.AST.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "ID" || fn.Recv == nil || fn.Body == nil {
				continue
			}
			recv, _ := receiver(fn)
			idName := idConstFromReturn(fn)
			if idName != "" {
				packetTypes["packet."+recv] = idName
			}
		}
	}
	return ids, packetTypes
}

func receiver(fn *ast.FuncDecl) (string, string) {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return "", ""
	}
	field := fn.Recv.List[0]
	name := ""
	if len(field.Names) > 0 {
		name = field.Names[0].Name
	}
	switch typ := field.Type.(type) {
	case *ast.Ident:
		return typ.Name, name
	case *ast.StarExpr:
		if ident, ok := typ.X.(*ast.Ident); ok {
			return ident.Name, name
		}
	}
	return "", name
}

func idConstFromReturn(fn *ast.FuncDecl) string {
	for _, stmt := range fn.Body.List {
		ret, ok := stmt.(*ast.ReturnStmt)
		if !ok || len(ret.Results) != 1 {
			continue
		}
		if ident, ok := ret.Results[0].(*ast.Ident); ok {
			return ident.Name
		}
	}
	return ""
}

func (e *extractor) evalConst(expr ast.Expr, iota int, pkg string) (int, bool) {
	switch value := expr.(type) {
	case nil:
		return iota, true
	case *ast.BasicLit:
		if value.Kind != token.INT {
			return 0, false
		}
		n, err := strconv.ParseInt(value.Value, 0, 64)
		return int(n), err == nil
	case *ast.Ident:
		if value.Name == "iota" {
			return iota, true
		}
		n, ok := e.consts[pkg+"."+value.Name]
		return n, ok
	case *ast.SelectorExpr:
		base, ok := value.X.(*ast.Ident)
		if !ok {
			return 0, false
		}
		n, ok := e.consts[base.Name+"."+value.Sel.Name]
		return n, ok
	case *ast.ParenExpr:
		return e.evalConst(value.X, iota, pkg)
	case *ast.BinaryExpr:
		left, ok := e.evalConst(value.X, iota, pkg)
		if !ok {
			return 0, false
		}
		right, ok := e.evalConst(value.Y, iota, pkg)
		if !ok {
			return 0, false
		}
		switch value.Op {
		case token.ADD:
			return left + right, true
		case token.SUB:
			return left - right, true
		case token.MUL:
			return left * right, true
		case token.SHL:
			return left << right, true
		case token.OR:
			return left | right, true
		}
	}
	return 0, false
}

func (e *extractor) expandType(key, field string, depth int, stack map[string]bool) []sourceOperation {
	if depth > extractionDepthLimit {
		return []sourceOperation{e.unresolved(nil, key, field, "recursion depth limit exceeded", key)}
	}
	if stack[key] {
		if e.recursion[key] >= recursionUnroll {
			return []sourceOperation{{Kind: "recursive", Field: field, TypeName: key, Reason: "recursive reference"}}
		}
		e.recursion[key]++
		defer func() { e.recursion[key]-- }()
	}
	method := e.marshalForType(key, map[string]bool{})
	if method == nil {
		return []sourceOperation{e.unresolved(nil, key, field, "target type has no Marshal method", key)}
	}
	next := cloneBoolMap(stack)
	next[key] = true
	env := map[string]typeRef{}
	if method.Recv != "" {
		env[method.Recv] = typeRef{Kind: "named", Name: key}
	}
	return e.extractBlock(method.Decl.Body.List, method, env, field, depth, next)
}

// marshalForType follows promoted methods on embedded structs. A number of
// protocol unions intentionally use small concrete types that inherit the
// Marshal method of an embedded base action; treating those as opaque would
// lose a reachable variant even though the Go method set is unambiguous.
func (e *extractor) marshalForType(key string, seen map[string]bool) *marshalInfo {
	if seen[key] {
		return nil
	}
	seen[key] = true
	if method := e.marshals[key]; method != nil {
		return method
	}
	info := e.types[key]
	if info == nil {
		return nil
	}
	for _, embedded := range info.Embedded {
		for embedded.Kind == "pointer" && embedded.Elem != nil {
			embedded = *embedded.Elem
		}
		if embedded.Kind == "named" {
			if method := e.marshalForType(embedded.Name, seen); method != nil {
				return method
			}
		}
	}
	return nil
}

func cloneBoolMap(input map[string]bool) map[string]bool {
	result := make(map[string]bool, len(input))
	for key, value := range input {
		result[key] = value
	}
	return result
}

func (e *extractor) extractBlock(stmts []ast.Stmt, method *marshalInfo, env map[string]typeRef, base string, depth int, stack map[string]bool) []sourceOperation {
	var result []sourceOperation
	for index, stmt := range stmts {
		switch current := stmt.(type) {
		case *ast.ExprStmt:
			if call, ok := current.X.(*ast.CallExpr); ok {
				result = append(result, e.extractCall(call, method, env, base, depth, stack)...)
			}
		case *ast.AssignStmt:
			for index, rhs := range current.Rhs {
				if index >= len(current.Lhs) {
					break
				}
				if ident, ok := current.Lhs[index].(*ast.Ident); ok && ident.Name != "_" {
					e.bindLocal(env, ident.Name, rhs, method, base)
				}
			}
			for _, rhs := range current.Rhs {
				if call, ok := rhs.(*ast.CallExpr); ok && e.isWireCall(call, method.IO, method.Key) {
					result = append(result, e.extractCall(call, method, env, base, depth, stack)...)
				}
			}
		case *ast.DeclStmt:
			if gen, ok := current.Decl.(*ast.GenDecl); ok {
				for _, spec := range gen.Specs {
					valueSpec, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}
					for _, name := range valueSpec.Names {
						ref := e.parseType(valueSpec.Type, method.File)
						if len(valueSpec.Values) > 0 {
							inferred := e.inferExprType(valueSpec.Values[0], method, env)
							if inferred.Kind != "unknown" {
								ref = inferred
							}
						}
						if ref.Kind != "unknown" {
							if len(valueSpec.Values) > 0 {
								ref.Path = e.fieldPath(valueSpec.Values[0], method, env, base)
							}
							env[name.Name] = ref
						}
					}
					for _, value := range valueSpec.Values {
						if call, ok := value.(*ast.CallExpr); ok && e.isWireCall(call, method.IO, method.Key) {
							result = append(result, e.extractCall(call, method, env, base, depth, stack)...)
						}
					}
				}
			}
		case *ast.IfStmt:
			returns := endsWithReturn(current.Body)
			if returns && isInvalidInputSink(current.Body) {
				// Rejecting bad input is not a wire path; the block continues.
				continue
			}
			if e.statementHasWire(current, method.IO, method.Key) || returns && e.statementsHaveWire(stmts[index+1:], method) {
				thenOps := e.extractBlock(current.Body.List, method, cloneTypeEnv(env), base, depth+1, stack)
				elseOps := []sourceOperation(nil)
				switch alternate := current.Else.(type) {
				case *ast.BlockStmt:
					elseOps = e.extractBlock(alternate.List, method, cloneTypeEnv(env), base, depth+1, stack)
				case *ast.IfStmt:
					elseOps = e.extractBlock([]ast.Stmt{alternate}, method, cloneTypeEnv(env), base, depth+1, stack)
				}
				if returns {
					// The body leaves the function, so the rest of this block is
					// the else path.
					elseOps = append(elseOps, e.extractBlock(stmts[index+1:], method, env, base, depth, stack)...)
				}
				condition := e.conditionVariant(current.Cond, method, env, base)
				result = append(result, sourceOperation{Kind: "conditional", Field: base, CompareTo: condition.CompareTo, Predicate: condition.Predicate, Variants: []sourceVariant{{
					Value: condition.Value, Values: []int64{condition.Value}, Name: condition.Name, Constraint: condition.Constraint, Discriminant: condition.Discriminant, Negated: condition.Negated, Site: e.nodeSite(current), Ops: thenOps,
				}}, Default: elseOps, HasDefault: true, Site: e.nodeSite(current)})
				if returns {
					return result
				}
			}
		case *ast.SwitchStmt:
			if e.statementHasWire(current, method.IO, method.Key) {
				result = append(result, e.extractSwitch(current, method, env, base, depth, stack))
			}
		case *ast.TypeSwitchStmt:
			if e.statementHasWire(current, method.IO, method.Key) {
				result = append(result, e.extractTypeSwitch(current, method, env, base, depth, stack))
			}
		case *ast.ForStmt:
			if e.statementHasWire(current, method.IO, method.Key) {
				result = append(result, e.extractFor(current, method, env, base, depth, stack))
			}
		case *ast.RangeStmt:
			if e.statementHasWire(current, method.IO, method.Key) {
				result = append(result, e.extractRange(current, method, env, base, depth, stack))
			}
		case *ast.DeferStmt:
			if e.statementHasWire(current, method.IO, method.Key) {
				result = append(result, e.unresolved(current, method.Key, base, "deferred wire call in Marshal", e.nodeString(current.Call)))
			}
		case *ast.BlockStmt:
			result = append(result, e.extractBlock(current.List, method, env, base, depth, stack)...)
		}
	}
	return result
}

// isInvalidInputSink reports whether a returning block only reports invalid
// input, such as `io.UnknownEnumOption(...); return`.
func isInvalidInputSink(block *ast.BlockStmt) bool {
	found := false
	ast.Inspect(block, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		if selector, ok := call.Fun.(*ast.SelectorExpr); ok {
			switch selector.Sel.Name {
			case "InvalidValue", "UnknownEnumOption":
				found = true
			}
		}
		return true
	})
	return found
}

func endsWithReturn(block *ast.BlockStmt) bool {
	if block == nil || len(block.List) == 0 {
		return false
	}
	_, ok := block.List[len(block.List)-1].(*ast.ReturnStmt)
	return ok
}

func (e *extractor) statementsHaveWire(stmts []ast.Stmt, method *marshalInfo) bool {
	for _, stmt := range stmts {
		if e.statementHasWire(stmt, method.IO, method.Key) {
			return true
		}
	}
	return false
}

type conditionVariant struct {
	Value        int64
	Name         string
	CompareTo    string
	Predicate    string
	Constraint   string
	Discriminant bool
	Negated      bool
}

func (e *extractor) conditionVariant(expr ast.Expr, method *marshalInfo, env map[string]typeRef, base string) conditionVariant {
	result := conditionVariant{Predicate: e.nodeString(expr), Constraint: e.nodeString(expr)}
	if binary, ok := expr.(*ast.BinaryExpr); ok && (binary.Op == token.EQL || binary.Op == token.NEQ) {
		result.CompareTo = e.fieldPath(binary.X, method, env, base)
		if value, ok := e.literalInt(binary.Y, method.File); ok {
			result.Value = int64(value)
			result.Discriminant = true
			result.Negated = binary.Op == token.NEQ
			if binary.Op == token.NEQ {
				result.Constraint = fmt.Sprintf("%s != %d", result.CompareTo, value)
			} else {
				result.Constraint = fmt.Sprintf("%s = %d", result.CompareTo, value)
			}
		}
		if literal, ok := binary.Y.(*ast.Ident); ok && (literal.Name == "true" || literal.Name == "false") {
			result.Value = 0
			if literal.Name == "true" {
				result.Value = 1
			}
			result.Negated = binary.Op == token.NEQ
		} else if !result.Discriminant {
			result.CompareTo = ""
		}
		return result
	}
	if unary, ok := expr.(*ast.UnaryExpr); ok && unary.Op == token.NOT {
		result.CompareTo = e.fieldPath(unary.X, method, env, base)
		result.Value = 0
		result.Constraint = result.CompareTo + " = false"
		return result
	}
	switch expr.(type) {
	case *ast.Ident, *ast.SelectorExpr:
	default:
		return result
	}
	result.CompareTo = e.fieldPath(expr, method, env, base)
	result.Value = 1
	if result.CompareTo != "" {
		result.Constraint = result.CompareTo + " = true"
	}
	return result
}

func (e *extractor) extractSwitch(stmt *ast.SwitchStmt, method *marshalInfo, env map[string]typeRef, base string, depth int, stack map[string]bool) sourceOperation {
	compareTo := base
	if stmt.Tag != nil {
		compareTo = e.fieldPath(stmt.Tag, method, env, base)
	}
	operation := sourceOperation{Kind: "switch", Field: base, CompareTo: compareTo, Predicate: e.nodeString(stmt.Tag), Site: e.nodeSite(stmt)}
	for _, stmt := range stmt.Body.List {
		clause, ok := stmt.(*ast.CaseClause)
		if !ok {
			continue
		}
		ops := e.extractBlock(clause.Body, method, cloneTypeEnv(env), base, depth+1, stack)
		if len(clause.List) == 0 {
			// A default containing only InvalidValue/UnknownEnumOption is an
			// invalid-input sink, not a reachable wire path. Keeping it as an
			// empty path would manufacture a discriminator-only variant that
			// the canonical manifest never promises.
			if len(ops) > 0 {
				operation.Default = ops
				operation.HasDefault = true
			}
			continue
		}
		variant := sourceVariant{Ops: ops, Site: e.nodeSite(clause)}
		for _, expr := range clause.List {
			value, ok := e.literalInt(expr, method.File)
			if !ok {
				variant.Constraint = e.nodeString(expr)
				variant.Values = nil
				operation.Variants = append(operation.Variants, variant)
				continue
			}
			variant.Values = append(variant.Values, int64(value))
		}
		if len(variant.Values) == 0 && variant.Constraint == "" {
			variant.Constraint = e.nodeString(clause)
		}
		operation.Variants = append(operation.Variants, variant)
	}
	return operation
}

func (e *extractor) extractTypeSwitch(stmt *ast.TypeSwitchStmt, method *marshalInfo, env map[string]typeRef, base string, depth int, stack map[string]bool) sourceOperation {
	operation := sourceOperation{Kind: "type_switch", Field: base, Predicate: e.nodeString(stmt.Assign), Site: e.nodeSite(stmt)}
	// switch v := x.(type) gives v the case's type in a single-type clause.
	var bound *ast.Ident
	var subject ast.Expr
	if assign, ok := stmt.Assign.(*ast.AssignStmt); ok && len(assign.Lhs) == 1 && len(assign.Rhs) == 1 {
		if assertion, ok := assign.Rhs[0].(*ast.TypeAssertExpr); ok {
			bound, _ = assign.Lhs[0].(*ast.Ident)
			subject = assertion.X
		}
	}
	for _, stmt := range stmt.Body.List {
		clause, ok := stmt.(*ast.CaseClause)
		if !ok {
			continue
		}
		clauseEnv := cloneTypeEnv(env)
		if bound != nil && bound.Name != "_" && len(clause.List) == 1 {
			ref := e.parseType(clause.List[0], method.File)
			ref.Path = e.fieldPath(subject, method, env, base)
			clauseEnv[bound.Name] = ref
		}
		ops := e.extractBlock(clause.Body, method, clauseEnv, base, depth+1, stack)
		if len(clause.List) == 0 {
			if len(ops) > 0 {
				operation.Default = ops
				operation.HasDefault = true
			}
			continue
		}
		value, ok := e.caseDiscriminant(clause, method, env, base, &ops)
		for _, expr := range clause.List {
			name := e.nodeString(expr)
			variant := sourceVariant{Name: name, Constraint: "type=" + name, Site: e.nodeSite(clause), Ops: ops}
			if ok {
				variant.Value, variant.Values, variant.Discriminant = value, []int64{value}, true
			}
			operation.Variants = append(operation.Variants, variant)
		}
	}
	return operation
}

// caseDiscriminant recognises a case body that assigns a constant to a local
// and writes that local as its first wire operation, which is how hand-written
// type switches emit a union discriminant. The variant marker is inserted
// after that write so the language matches a manifest union.
func (e *extractor) caseDiscriminant(clause *ast.CaseClause, method *marshalInfo, env map[string]typeRef, base string, ops *[]sourceOperation) (int64, bool) {
	for _, stmt := range clause.Body {
		assign, ok := stmt.(*ast.AssignStmt)
		if !ok || len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
			continue
		}
		ident, ok := assign.Lhs[0].(*ast.Ident)
		if !ok {
			continue
		}
		expr := assign.Rhs[0]
		if call, isCall := expr.(*ast.CallExpr); isCall && len(call.Args) == 1 {
			expr = call.Args[0]
		}
		value, ok := e.literalInt(expr, method.File)
		if !ok {
			continue
		}
		field := e.fieldPath(ident, method, env, base)
		for index, operation := range *ops {
			if operation.Kind != "primitive" {
				continue
			}
			if operation.Field != field && operation.Field != base+"."+ident.Name && operation.Field != ident.Name {
				break
			}
			marker := sourceOperation{Kind: "variant_marker", Field: operation.Field, VariantValue: int64(value), Site: operation.Site}
			*ops = append((*ops)[:index+1], append([]sourceOperation{marker}, (*ops)[index+1:]...)...)
			return int64(value), true
		}
		return 0, false
	}
	return 0, false
}

func (e *extractor) extractFor(stmt *ast.ForStmt, method *marshalInfo, env map[string]typeRef, base string, depth int, stack map[string]bool) sourceOperation {
	count, variable, ok := e.staticForCount(stmt, method.File)
	if !ok {
		return e.unresolved(stmt, method.Key, base, "runtime loop bound is not a static positive integer", e.nodeString(stmt))
	}
	loopEnv := cloneTypeEnv(env)
	if variable != "" {
		loopEnv[variable] = typeRef{Kind: "primitive", Name: "int"}
	}
	children := e.extractBlock(stmt.Body.List, method, loopEnv, base+"[]", depth+1, stack)
	return sourceOperation{Kind: "fixed_array", Field: base, Length: uint64(count), Element: children, Site: e.nodeSite(stmt)}
}

func (e *extractor) extractRange(stmt *ast.RangeStmt, method *marshalInfo, env map[string]typeRef, base string, depth int, stack map[string]bool) sourceOperation {
	field := e.fieldPath(stmt.X, method, env, base)
	ref := e.underlyingArray(e.resolveExprType(stmt.X, method, env))
	if ref.Kind != "array" || ref.Elem == nil || ref.Len <= 0 {
		if method.Key == "protocol.Writer.EntityMetadata" {
			loopEnv := cloneTypeEnv(env)
			if ident, ok := stmt.Key.(*ast.Ident); ok && ident.Name != "_" {
				loopEnv[ident.Name] = typeRef{Kind: "primitive", Name: "uint32", Path: field + ".<key>"}
			}
			if ident, ok := stmt.Value.(*ast.Ident); ok && ident.Name != "_" {
				loopEnv[ident.Name] = typeRef{Kind: "named", Name: "protocol.EntityMetadataValue", Path: field + ".<value>"}
			}
			children := e.extractBlock(stmt.Body.List, method, loopEnv, field+"[]", depth+1, stack)
			return sourceOperation{Kind: "array", Field: field, Prefix: "var_u32", Element: children, ConsumesPrefix: true, Site: e.nodeSite(stmt)}
		}
		return e.unresolved(stmt, method.Key, field, "runtime loop outside a statically sized array", e.nodeString(stmt.X))
	}
	loopEnv := cloneTypeEnv(env)
	if ident, ok := stmt.Key.(*ast.Ident); ok && ident.Name != "_" {
		loopEnv[ident.Name] = typeRef{Kind: "primitive", Name: "int"}
	}
	if ident, ok := stmt.Value.(*ast.Ident); ok && ident.Name != "_" {
		loopEnv[ident.Name] = *ref.Elem
	}
	children := e.extractBlock(stmt.Body.List, method, loopEnv, field+"[]", depth+1, stack)
	return sourceOperation{Kind: "fixed_array", Field: field, Length: uint64(ref.Len), Element: children, Site: e.nodeSite(stmt)}
}

func (e *extractor) staticForCount(stmt *ast.ForStmt, sf *sourceFile) (int, string, bool) {
	variable := ""
	start := 0
	if assign, ok := stmt.Init.(*ast.AssignStmt); ok && len(assign.Lhs) == 1 && len(assign.Rhs) == 1 {
		if ident, yes := assign.Lhs[0].(*ast.Ident); yes {
			variable = ident.Name
		}
		if value, yes := e.literalInt(assign.Rhs[0], sf); yes {
			start = value
		}
	}
	if variable == "" || stmt.Cond == nil {
		return 0, variable, false
	}
	binary, ok := stmt.Cond.(*ast.BinaryExpr)
	if !ok || (binary.Op != token.LSS && binary.Op != token.LEQ) {
		return 0, variable, false
	}
	ident, ok := binary.X.(*ast.Ident)
	if !ok || ident.Name != variable {
		return 0, variable, false
	}
	limit, ok := e.literalInt(binary.Y, sf)
	if !ok {
		return 0, variable, false
	}
	if _, ok := stmt.Post.(*ast.IncDecStmt); !ok {
		return 0, variable, false
	}
	count := limit - start
	if binary.Op == token.LEQ {
		count++
	}
	return count, variable, count > 0
}

func cloneTypeEnv(input map[string]typeRef) map[string]typeRef {
	result := make(map[string]typeRef, len(input))
	for key, value := range input {
		result[key] = value
	}
	return result
}

func (e *extractor) bindLocal(env map[string]typeRef, name string, expr ast.Expr, method *marshalInfo, base string) {
	if literal, ok := expr.(*ast.FuncLit); ok {
		env[name] = typeRef{Kind: "function", Name: name, Path: base, Function: literal}
		return
	}
	ref := e.inferExprType(expr, method, env)
	if ref.Kind == "unknown" {
		return
	}
	ref.Path = e.fieldPath(expr, method, env, base)
	env[name] = ref
}

func (e *extractor) inferExprType(expr ast.Expr, method *marshalInfo, env map[string]typeRef) typeRef {
	if literal, ok := expr.(*ast.FuncLit); ok {
		return typeRef{Kind: "function", Function: literal}
	}
	if ref := e.resolveExprType(expr, method, env); ref.Kind != "unknown" {
		return ref
	}
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return typeRef{Kind: "unknown"}
	}
	fun := unwrapIndex(call.Fun)
	if ident, ok := fun.(*ast.Ident); ok {
		if _, primitive := sourceBuiltinPrimitive[ident.Name]; primitive {
			return typeRef{Kind: "primitive", Name: ident.Name}
		}
		if ident.Name == "len" {
			return typeRef{Kind: "primitive", Name: "int"}
		}
	}
	return typeRef{Kind: "unknown"}
}

func (e *extractor) expandFunctionLiteral(call *ast.CallExpr, literal *ast.FuncLit, caller *marshalInfo, env map[string]typeRef, base string, depth int, stack map[string]bool) []sourceOperation {
	if depth >= extractionDepthLimit {
		return []sourceOperation{e.unresolved(call, caller.Key, base, "nested function expansion depth limit exceeded", e.nodeString(call))}
	}
	site := e.nodeSite(literal)
	stackKey := "literal:" + site
	if stack[stackKey] {
		return []sourceOperation{e.unresolved(call, caller.Key, base, "recursive nested function call", site)}
	}
	nextStack := cloneBoolMap(stack)
	nextStack[stackKey] = true
	nextEnv := cloneTypeEnv(env)
	context := *caller
	argIndex := 0
	if params := literal.Type.Params; params != nil {
		for _, param := range params.List {
			ref := e.parseType(param.Type, caller.File)
			for _, name := range param.Names {
				if isIORef(ref) {
					context.IO = name.Name
					nextEnv[name.Name] = ref
				} else if argIndex < len(call.Args) {
					bound := ref
					bound.Path = e.fieldPath(call.Args[argIndex], caller, env, base)
					nextEnv[name.Name] = bound
				} else {
					nextEnv[name.Name] = ref
				}
				argIndex++
			}
		}
	}
	return e.extractBlock(literal.Body.List, &context, nextEnv, base, depth+1, nextStack)
}

func (e *extractor) nodeSite(node ast.Node) string {
	if node == nil {
		return ""
	}
	position := e.fset.Position(node.Pos())
	return filepath.ToSlash(position.Filename) + ":" + strconv.Itoa(position.Line)
}

func (e *extractor) statementHasWire(node ast.Node, ioName, methodKey string) bool {
	found := false
	ast.Inspect(node, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if ok && e.isWireCall(call, ioName, methodKey) {
			found = true
			return false
		}
		return !found
	})
	return found
}

func (e *extractor) isWireCall(call *ast.CallExpr, ioName, methodKey string) bool {
	fun := unwrapIndex(call.Fun)
	switch current := fun.(type) {
	case *ast.SelectorExpr:
		if base, ok := current.X.(*ast.Ident); ok {
			if base.Name == ioName {
				return current.Sel.Name != "InvalidValue" && current.Sel.Name != "UnknownEnumOption" && current.Sel.Name != "ShieldID"
			}
			if base.Name == "protocol" {
				return current.Sel.Name != "Option"
			}
		}
		return current.Sel.Name == "Marshal" || referencesIO(call, ioName)
	case *ast.Ident:
		if current.Name == "writeType" && methodKey == "protocol.Writer.EntityMetadata" {
			// The closure is a pinned part of Writer.EntityMetadata. Its
			// invocation emits the discriminator and legacy discriminator.
			return true
		}
		return referencesIO(call, ioName)
	default:
		return false
	}
}

func (e *extractor) extractCall(call *ast.CallExpr, method *marshalInfo, env map[string]typeRef, base string, depth int, stack map[string]bool) []sourceOperation {
	fun := unwrapIndex(call.Fun)
	selector, isSelector := fun.(*ast.SelectorExpr)
	if !isSelector {
		if ident, ok := fun.(*ast.Ident); ok && method.File.Pkg == "protocol" && isProtocolHelper(ident.Name) {
			return e.expandProtocol(call, ident.Name, method, env, base, depth, stack)
		}
		if ident, ok := fun.(*ast.Ident); ok && ident.Name == "writeType" && method.Key == "protocol.Writer.EntityMetadata" {
			return []sourceOperation{
				{Kind: "primitive", Field: base, Code: "var_u32"},
				{Kind: "primitive", Field: base, Code: "u8"},
			}
		}
		if ident, ok := fun.(*ast.Ident); ok {
			if ref, ok := env[ident.Name]; ok && ref.Kind == "function" && ref.Function != nil {
				return e.expandFunctionLiteral(call, ref.Function, method, env, base, depth, stack)
			}
			if target := e.functions[method.File.Pkg+"."+ident.Name]; target != nil && target.IO != "" {
				return e.expandFunctionCall(call, target, method, env, base, depth, stack)
			}
		}
		if referencesIO(call, method.IO) && !strings.HasSuffix(method.Key, ".itemUserData") {
			return []sourceOperation{e.unresolved(call, method.Key, base, "unrecognized function call in Marshal", e.nodeString(call))}
		}
		return nil
	}
	name := selector.Sel.Name
	if ident, ok := selector.X.(*ast.Ident); ok && ident.Name == method.IO && reviewedHelperRevisions[e.revision] && reviewedHelperMethods[name] {
		return e.expandReviewedHelperMethod(call, name, method, env, base, depth, stack)
	}
	if ident, ok := selector.X.(*ast.Ident); ok && ident.Name == method.IO {
		if name == "InvalidValue" || name == "UnknownEnumOption" || name == "ShieldID" {
			return nil
		}
		field := base
		if len(call.Args) > 0 {
			field = e.fieldPath(call.Args[0], method, env, base)
		}
		if code, ok := sourcePrimitive[name]; ok {
			e.recordWireCode(call.Args, code, env)
			return []sourceOperation{{Kind: "primitive", Field: field, Code: code, Site: e.nodeSite(call)}}
		}
		operations := e.expandIOHelper(call, name, field, method, depth, stack)
		stampOperationSites(operations, e.nodeSite(call))
		return operations
	}
	if ident, ok := selector.X.(*ast.Ident); ok && ident.Name == "protocol" {
		if name == "Option" {
			return nil
		}
		operations := e.expandProtocol(call, name, method, env, base, depth, stack)
		stampOperationSites(operations, e.nodeSite(call))
		return operations
	}
	if name == "Marshal" {
		ref := e.resolveExprType(selector.X, method, env)
		for ref.Kind == "pointer" && ref.Elem != nil {
			ref = *ref.Elem
		}
		field := e.fieldPath(selector.X, method, env, base)
		if ref.Kind != "named" {
			return []sourceOperation{e.unresolved(call, method.Key, field, "Marshal receiver type could not be resolved", e.nodeString(selector.X))}
		}
		return e.expandRef(ref, field, depth+1, stack, call, method)
	}
	if ident, ok := selector.X.(*ast.Ident); ok && ident.Name == method.File.Pkg {
		if target := e.functions[method.File.Pkg+"."+name]; target != nil && target.IO != "" {
			return e.expandFunctionCall(call, target, method, env, base, depth, stack)
		}
	}
	if referencesIO(call, method.IO) {
		return []sourceOperation{e.unresolved(call, method.Key, base, "opaque method/helper call in Marshal", e.nodeString(call))}
	}
	return nil
}

func stampOperationSites(operations []sourceOperation, site string) {
	if site == "" {
		return
	}
	for index := range operations {
		if operations[index].Site == "" {
			operations[index].Site = site
		}
	}
}

func (e *extractor) recordWireCode(args []ast.Expr, code string, env map[string]typeRef) {
	if len(args) == 0 {
		return
	}
	expr := args[0]
	for {
		switch current := expr.(type) {
		case *ast.UnaryExpr:
			expr = current.X
		case *ast.ParenExpr:
			expr = current.X
		default:
			ident, ok := expr.(*ast.Ident)
			if !ok {
				return
			}
			ref, ok := env[ident.Name]
			if !ok {
				return
			}
			ref.WireCode = code
			env[ident.Name] = ref
			return
		}
	}
}

func (e *extractor) expandFunctionCall(call *ast.CallExpr, target, caller *marshalInfo, env map[string]typeRef, base string, depth int, stack map[string]bool) []sourceOperation {
	stackKey := "function:" + target.Key
	if stack[stackKey] {
		if e.recursion[target.Key] >= recursionUnroll {
			return []sourceOperation{{Kind: "recursive", Field: base, Reason: target.Key, Site: e.nodeSite(call)}}
		}
		e.recursion[target.Key]++
		defer func() { e.recursion[target.Key]-- }()
	}
	if depth >= extractionDepthLimit {
		return []sourceOperation{e.unresolved(call, caller.Key, base, "local helper expansion depth limit exceeded", target.Key)}
	}
	nextStack := cloneBoolMap(stack)
	nextStack[stackKey] = true
	nextEnv := cloneTypeEnv(env)
	params := target.Decl.Type.Params
	argIndex := 0
	if params != nil {
		for _, param := range params.List {
			ref := e.parseType(param.Type, target.File)
			for _, name := range param.Names {
				var arg ast.Expr
				if argIndex < len(call.Args) {
					arg = call.Args[argIndex]
				}
				if isIORef(ref) {
					nextEnv[name.Name] = ref
				} else if arg != nil {
					bound := ref
					bound.Path = e.fieldPath(arg, caller, env, base)
					nextEnv[name.Name] = bound
				}
				argIndex++
			}
			if len(param.Names) == 0 {
				argIndex++
			}
		}
	}
	context := *target
	return e.extractBlock(target.Decl.Body.List, &context, nextEnv, base, depth+1, nextStack)
}

func isProtocolHelper(name string) bool {
	if _, ok := sourceArrayPrefixes[name]; ok {
		return true
	}
	switch name {
	case "SliceOfLen", "FuncSliceOfLen", "FuncIOSliceOfLen", "Optional", "OptionalFunc", "DoubleOptionalFunc", "OptionalMarshaler", "Single", "IntegerFunc", "StackRequestItemDescriptorCount":
		return true
	default:
		return false
	}
}

func (e *extractor) expandIOHelper(call *ast.CallExpr, name, field string, method *marshalInfo, depth int, stack map[string]bool) []sourceOperation {
	primN := func(code string, count int) []sourceOperation {
		result := make([]sourceOperation, count)
		axis := []string{"X", "Y", "Z"}
		for i := range result {
			path := field
			if count > 1 {
				path += "." + axis[i]
			}
			result[i] = sourceOperation{Kind: "primitive", Field: path, Code: code}
		}
		return result
	}
	switch name {
	case "Vec3":
		return primN("f32le", 3)
	case "Vec2":
		return primN("f32le", 2)
	case "BlockPos":
		return primN("zigzag_i32", 3)
	case "ChunkPos":
		return primN("zigzag_i32", 2)
	case "SubChunkPos":
		return primN("i32le", 3)
	case "SoundPos":
		return primN("zigzag_i32", 3)
	case "ByteFloat":
		return primN("u8", 1)
	case "UUID":
		return []sourceOperation{{Kind: "uuid", Field: field, Length: 16}}
	case "String":
		return []sourceOperation{{Kind: "string", Field: field, Prefix: "var_u32", Encoding: "utf8"}}
	case "StringUTF":
		return []sourceOperation{{Kind: "string", Field: field, Prefix: "i16le", Encoding: "utf8"}}
	case "ByteSlice":
		return []sourceOperation{{Kind: "bytes", Field: field, Prefix: "var_u32"}}
	case "NBT":
		encoding := ""
		if len(call.Args) > 1 {
			if selector, ok := call.Args[1].(*ast.SelectorExpr); ok {
				if qualifier, ok := selector.X.(*ast.Ident); ok && method.File.Imports[qualifier.Name] == "nbt" {
					switch selector.Sel.Name {
					case "NetworkLittleEndian":
						encoding = "network"
					case "LittleEndian":
						encoding = "persistent"
					}
				}
			}
		}
		return []sourceOperation{{Kind: "primitive", Field: field, Code: "nbt_le", Encoding: encoding}}
	case "NBTList", "RawNBT":
		return []sourceOperation{e.unresolved(call, method.Key, field, "NBT helper framing is not established", name)}
	case "Bytes":
		return []sourceOperation{{Kind: "primitive", Field: field, Code: "raw_bytes"}}
	case "Bitset":
		if len(call.Args) < 2 {
			return []sourceOperation{e.unresolved(call, method.Key, field, "bitset size argument is missing", name)}
		}
		length, ok := e.literalInt(call.Args[1], method.File)
		if !ok || length <= 0 {
			return []sourceOperation{e.unresolved(call, method.Key, field, "bitset size is not a static integer", e.nodeString(call.Args[1]))}
		}
		return []sourceOperation{{Kind: "bitset", Field: field, Length: uint64(length)}}
	default:
		if reviewedHelperRevisions[e.revision] && reviewedIOHelpers[name] {
			return e.expandReviewedIOHelper(call, name, field, method, depth, stack)
		}
		return []sourceOperation{e.unresolved(call, method.Key, field, "opaque or unsupported IO helper", name)}
	}
}

func (e *extractor) expandReviewedIOHelper(call *ast.CallExpr, name, field string, caller *marshalInfo, depth int, stack map[string]bool) []sourceOperation {
	var target *marshalInfo
	for _, receiver := range []string{"Writer", "Reader"} {
		if candidate := e.ioHelpers["protocol."+receiver+"."+name]; candidate != nil {
			target = candidate
			break
		}
	}
	if target == nil {
		return []sourceOperation{e.unresolved(call, caller.Key, field, "reviewed helper has no pinned implementation", name)}
	}
	stackKey := "helper:" + target.Key
	if stack[stackKey] {
		return []sourceOperation{e.unresolved(call, caller.Key, field, "recursive reviewed helper", target.Key)}
	}
	if depth >= extractionDepthLimit {
		return []sourceOperation{e.unresolved(call, caller.Key, field, "reviewed helper expansion depth limit exceeded", name)}
	}
	nextStack := cloneBoolMap(stack)
	nextStack[stackKey] = true
	env := map[string]typeRef{target.Recv: {Kind: "named", Name: target.Key}}
	params := target.Decl.Type.Params
	argIndex := 0
	if params != nil {
		for _, param := range params.List {
			ref := e.parseType(param.Type, target.File)
			for _, paramName := range param.Names {
				if isIORef(ref) {
					continue
				}
				if argIndex == 0 {
					ref.Path = field
				}
				env[paramName.Name] = ref
				argIndex++
			}
		}
	}
	context := *target
	return e.extractBlock(target.Decl.Body.List, &context, env, field, depth+1, nextStack)
}

func (e *extractor) expandProtocol(call *ast.CallExpr, name string, method *marshalInfo, env map[string]typeRef, base string, depth int, stack map[string]bool) []sourceOperation {
	if prefix, ok := sourceArrayPrefixes[name]; ok {
		if len(call.Args) < 2 {
			return []sourceOperation{e.unresolved(call, method.Key, base, "slice helper has too few arguments", name)}
		}
		field := e.fieldPath(call.Args[1], method, env, base)
		ref := e.resolveExprType(call.Args[1], method, env)
		for ref.Kind == "pointer" && ref.Elem != nil {
			ref = *ref.Elem
		}
		if (ref.Kind != "slice" && ref.Kind != "array") || ref.Elem == nil {
			return []sourceOperation{e.unresolved(call, method.Key, field, "slice field type could not be resolved", e.nodeString(call.Args[1]))}
		}
		element := *ref.Elem
		var children []sourceOperation
		if strings.HasPrefix(name, "Func") || name == "SliceOfFunc" {
			if len(call.Args) < 3 {
				children = []sourceOperation{e.unresolved(call, method.Key, field+"[]", "slice callback is missing", name)}
			} else {
				children = e.callbackOps(call.Args[2], element, method, env, field+"[]", depth, stack)
			}
		} else {
			children = e.expandRef(element, field+"[]", depth+1, stack, call, method)
		}
		return []sourceOperation{{Kind: "array", Field: field, Prefix: prefix, Element: children}}
	}
	if name == "SliceOfLen" || name == "FuncSliceOfLen" || name == "FuncIOSliceOfLen" {
		if len(call.Args) < 3 {
			return []sourceOperation{e.unresolved(call, method.Key, base, "slice-of-length helper has too few arguments", name)}
		}
		field := e.fieldPath(call.Args[2], method, env, base)
		ref := e.resolveExprType(call.Args[2], method, env)
		for ref.Kind == "pointer" && ref.Elem != nil {
			ref = *ref.Elem
		}
		if (ref.Kind != "slice" && ref.Kind != "array") || ref.Elem == nil {
			return []sourceOperation{e.unresolved(call, method.Key, field, "slice field type could not be resolved", e.nodeString(call.Args[2]))}
		}
		length, ok := e.literalInt(call.Args[1], method.File)
		prefix := ""
		consumesPrefix := false
		if !ok || length <= 0 {
			lengthRef := e.resolveExprType(call.Args[1], method, env)
			if lengthRef.WireCode == "" {
				return []sourceOperation{e.unresolved(call, method.Key, field, "slice helper uses an external runtime length that cannot be represented", e.nodeString(call.Args[1]))}
			}
			prefix = lengthRef.WireCode
			consumesPrefix = true
			length = 0
		}
		var children []sourceOperation
		if name == "SliceOfLen" {
			children = e.expandRef(*ref.Elem, field+"[]", depth+1, stack, call, method)
		} else if len(call.Args) < 4 {
			children = []sourceOperation{e.unresolved(call, method.Key, field+"[]", "slice callback is missing", name)}
		} else {
			children = e.callbackOps(call.Args[3], *ref.Elem, method, env, field+"[]", depth, stack)
		}
		if consumesPrefix {
			return []sourceOperation{{Kind: "array", Field: field, Prefix: prefix, Element: children, ConsumesPrefix: true}}
		}
		return []sourceOperation{{Kind: "fixed_array", Field: field, Length: uint64(length), Element: children}}
	}
	if name == "Optional" || name == "OptionalFunc" || name == "DoubleOptionalFunc" || name == "OptionalMarshaler" {
		if len(call.Args) < 2 {
			return []sourceOperation{e.unresolved(call, method.Key, base, "optional helper has too few arguments", name)}
		}
		field := e.fieldPath(call.Args[1], method, env, base)
		ref := e.resolveExprType(call.Args[1], method, env)
		for ref.Kind == "pointer" && ref.Elem != nil {
			ref = *ref.Elem
		}
		if ref.Kind != "optional" || ref.Elem == nil {
			return []sourceOperation{e.unresolved(call, method.Key, field, "optional field type could not be resolved", e.nodeString(call.Args[1]))}
		}
		var children []sourceOperation
		if (name == "OptionalFunc" || name == "DoubleOptionalFunc") && len(call.Args) >= 3 {
			children = e.callbackOps(call.Args[2], *ref.Elem, method, env, field, depth, stack)
		} else {
			children = e.expandRef(*ref.Elem, field, depth+1, stack, call, method)
		}
		inner := sourceOperation{Kind: "optional", Field: field, Presence: "bool", Value: children}
		if name == "DoubleOptionalFunc" {
			return []sourceOperation{{Kind: "optional", Field: field, Presence: "bool", Value: []sourceOperation{inner}}}
		}
		return []sourceOperation{inner}
	}
	if name == "IntegerFunc" {
		if len(call.Args) < 2 {
			return []sourceOperation{e.unresolved(call, method.Key, base, "IntegerFunc has too few arguments", name)}
		}
		field := e.fieldPath(call.Args[0], method, env, base)
		callback, ok := unwrapIndex(call.Args[len(call.Args)-1]).(*ast.SelectorExpr)
		if !ok {
			return []sourceOperation{e.unresolved(call, method.Key, field, "IntegerFunc callback could not be resolved", e.nodeString(call))}
		}
		ioBase, ok := callback.X.(*ast.Ident)
		if !ok || ioBase.Name != method.IO {
			return []sourceOperation{e.unresolved(call, method.Key, field, "IntegerFunc callback is not an IO method", e.nodeString(call))}
		}
		code, ok := sourcePrimitive[callback.Sel.Name]
		if !ok {
			return []sourceOperation{e.unresolved(call, method.Key, field, "IntegerFunc callback is unsupported", callback.Sel.Name)}
		}
		return []sourceOperation{{Kind: "primitive", Field: field, Code: code}}
	}
	if name == "Single" {
		if len(call.Args) < 2 {
			return []sourceOperation{e.unresolved(call, method.Key, base, "Single helper has too few arguments", name)}
		}
		field := e.fieldPath(call.Args[1], method, env, base)
		ref := e.resolveExprType(call.Args[1], method, env)
		for ref.Kind == "pointer" && ref.Elem != nil {
			ref = *ref.Elem
		}
		return e.expandRef(ref, field, depth+1, stack, call, method)
	}
	if reviewedHelperRevisions[e.revision] && reviewedProtocolHelpers[name] {
		if target := e.functions["protocol."+name]; target != nil && target.IO != "" {
			return e.expandFunctionCall(call, target, method, env, base, depth, stack)
		}
		return []sourceOperation{e.unresolved(call, method.Key, base, "reviewed protocol helper has no pinned implementation", name)}
	}
	return []sourceOperation{e.unresolved(call, method.Key, base, "unsupported protocol helper", name)}
}

func (e *extractor) expandReviewedHelperMethod(call *ast.CallExpr, name string, caller *marshalInfo, callerEnv map[string]typeRef, base string, depth int, stack map[string]bool) []sourceOperation {
	target := e.ioHelpers["protocol."+shortType(caller.IO)+"."+name]
	if target == nil {
		// The receiver name is normally Writer/Reader, but caller.IO is the
		// local receiver variable (for example "w"). Search both reviewed
		// implementations so the summary stays independent of variable names.
		for _, receiver := range []string{"Writer", "Reader"} {
			if candidate := e.ioHelpers["protocol."+receiver+"."+name]; candidate != nil {
				target = candidate
				break
			}
		}
	}
	if target == nil {
		return []sourceOperation{e.unresolved(call, caller.Key, base, "reviewed helper method has no pinned implementation", name)}
	}
	stackKey := "helper:" + target.Key
	if stack[stackKey] {
		return []sourceOperation{e.unresolved(call, caller.Key, base, "recursive reviewed helper method", target.Key)}
	}
	if depth >= extractionDepthLimit {
		return []sourceOperation{e.unresolved(call, caller.Key, base, "reviewed helper method expansion depth limit exceeded", name)}
	}
	nextStack := cloneBoolMap(stack)
	nextStack[stackKey] = true
	receiverType := ""
	keyParts := strings.Split(target.Key, ".")
	if len(keyParts) >= 3 {
		receiverType = keyParts[len(keyParts)-2]
	}
	nextEnv := map[string]typeRef{target.Recv: {Kind: "named", Name: "protocol." + receiverType}}
	argIndex := 0
	if params := target.Decl.Type.Params; params != nil {
		for _, param := range params.List {
			ref := e.parseType(param.Type, target.File)
			for _, paramName := range param.Names {
				if argIndex < len(call.Args) {
					bound := ref
					bound.Path = e.fieldPath(call.Args[argIndex], caller, callerEnv, base)
					nextEnv[paramName.Name] = bound
				} else {
					nextEnv[paramName.Name] = ref
				}
				argIndex++
			}
		}
	}
	context := *target
	return e.extractBlock(target.Decl.Body.List, &context, nextEnv, base, depth+1, nextStack)
}

func fieldOrBase(field, base string) string {
	if field != "" {
		return field
	}
	return base
}

func (e *extractor) callbackOps(fn ast.Expr, element typeRef, method *marshalInfo, env map[string]typeRef, field string, depth int, stack map[string]bool) []sourceOperation {
	fn = unwrapIndex(fn)
	if strings.HasSuffix(method.Key, ".itemUserData") {
		// itemUserData serializes its private buffer and emits it once through
		// the enclosing ByteSlice. The buffer's callbacks are not additional
		// outer wire operations.
		return nil
	}
	if selector, ok := fn.(*ast.SelectorExpr); ok {
		if base, yes := selector.X.(*ast.Ident); yes && base.Name == method.IO {
			fake := &ast.CallExpr{Fun: selector, Args: []ast.Expr{ast.NewIdent(field)}}
			if code, yes := sourcePrimitive[selector.Sel.Name]; yes {
				return []sourceOperation{{Kind: "primitive", Field: field, Code: code}}
			}
			return e.expandIOHelper(fake, selector.Sel.Name, field, method, depth, stack)
		}
		if selector.Sel.Name == "Marshal" {
			ref := e.resolveExprType(selector.X, method, env)
			for ref.Kind == "pointer" && ref.Elem != nil {
				ref = *ref.Elem
			}
			if ref.Kind == "named" && e.marshals[ref.Name] != nil {
				return e.callbackDecl(e.marshals[ref.Name], element, field, depth, stack)
			}
		}
	}
	if literal, ok := fn.(*ast.FuncLit); ok {
		nextEnv := map[string]typeRef{}
		for key, value := range env {
			nextEnv[key] = value
		}
		nextMethod := *method
		if literal.Type.Params != nil {
			for index, param := range literal.Type.Params.List {
				for _, name := range param.Names {
					ref := e.parseType(param.Type, method.File)
					if isIORef(ref) {
						nextMethod.IO = name.Name
						nextEnv[name.Name] = ref
					} else if index == len(literal.Type.Params.List)-1 {
						nextEnv[name.Name] = element
					} else {
						nextEnv[name.Name] = ref
					}
				}
			}
		}
		return e.extractBlock(literal.Body.List, &nextMethod, nextEnv, field, depth+1, stack)
	}
	if ident, ok := fn.(*ast.Ident); ok && method.File.Pkg == "protocol" {
		if target := e.functions["protocol."+ident.Name]; target != nil {
			return e.callbackDecl(target, element, field, depth, stack)
		}
	}
	return []sourceOperation{e.unresolved(fn, method.Key, field, "slice/optional callback could not be resolved", e.nodeString(fn))}
}

func (e *extractor) callbackDecl(target *marshalInfo, element typeRef, field string, depth int, stack map[string]bool) []sourceOperation {
	context := *target
	env := map[string]typeRef{}
	if target.Recv != "" {
		env[target.Recv] = typeRef{Kind: "named", Name: target.Key}
	}
	params := target.Decl.Type.Params
	if params != nil {
		for index, param := range params.List {
			for _, name := range param.Names {
				ref := e.parseType(param.Type, target.File)
				if isIORef(ref) {
					context.IO = name.Name
					env[name.Name] = ref
				} else if index == len(params.List)-1 {
					env[name.Name] = element
				} else {
					env[name.Name] = ref
				}
			}
		}
	}
	if context.IO == "" {
		return []sourceOperation{e.unresolved(target.Decl, target.Key, field, "named callback IO parameter could not be resolved", target.Key)}
	}
	return e.extractBlock(target.Decl.Body.List, &context, env, field, depth+1, stack)
}

func (e *extractor) expandRef(ref typeRef, field string, depth int, stack map[string]bool, node ast.Node, method *marshalInfo) []sourceOperation {
	for ref.Kind == "pointer" && ref.Elem != nil {
		ref = *ref.Elem
	}
	switch ref.Kind {
	case "named":
		if reviewedHelperRevisions[e.revision] {
			if variants, ok := reviewedInterfaceVariants[ref.Name]; ok {
				result := make([]sourceVariant, 0, len(variants))
				for _, variant := range variants {
					result = append(result, sourceVariant{
						Value: variant.Value,
						Name:  shortType(variant.Type),
						Ops:   e.expandType(variant.Type, field, depth+1, stack),
					})
				}
				return []sourceOperation{{Kind: "union", Field: field, Variants: result}}
			}
		}
		return e.expandType(ref.Name, field, depth, stack)
	case "primitive":
		if code, ok := sourceBuiltinPrimitive[ref.Name]; ok {
			return []sourceOperation{{Kind: "primitive", Field: field, Code: code}}
		}
		return []sourceOperation{e.unresolved(node, method.Key, field, "element primitive could not be resolved", ref.Name)}
	case "array":
		if ref.Elem == nil || ref.Len <= 0 {
			return []sourceOperation{e.unresolved(node, method.Key, field, "fixed array type is incomplete", ref.String())}
		}
		children := e.expandRef(*ref.Elem, field+"[]", depth+1, stack, node, method)
		return []sourceOperation{{Kind: "fixed_array", Field: field, Length: uint64(ref.Len), Element: children}}
	case "slice":
		return []sourceOperation{e.unresolved(node, method.Key, field, "bare slice has no known length prefix", ref.String())}
	default:
		return []sourceOperation{e.unresolved(node, method.Key, field, "element/value type could not be resolved", ref.String())}
	}
}

func (e *extractor) parseType(expr ast.Expr, sf *sourceFile) typeRef {
	switch current := expr.(type) {
	case *ast.Ident:
		if _, ok := sourceBuiltinPrimitive[current.Name]; ok {
			return typeRef{Kind: "primitive", Name: current.Name}
		}
		return typeRef{Kind: "named", Name: sf.Pkg + "." + current.Name}
	case *ast.SelectorExpr:
		if base, ok := current.X.(*ast.Ident); ok {
			pkg := sf.Imports[base.Name]
			if pkg == "" {
				pkg = base.Name
			}
			return typeRef{Kind: "named", Name: pkg + "." + current.Sel.Name}
		}
	case *ast.StarExpr:
		inner := e.parseType(current.X, sf)
		return typeRef{Kind: "pointer", Elem: &inner}
	case *ast.ArrayType:
		inner := e.parseType(current.Elt, sf)
		if current.Len == nil {
			return typeRef{Kind: "slice", Elem: &inner}
		}
		length, _ := e.literalInt(current.Len, sf)
		return typeRef{Kind: "array", Elem: &inner, Len: length}
	case *ast.IndexExpr:
		if isOptionalType(current.X) {
			inner := e.parseType(current.Index, sf)
			return typeRef{Kind: "optional", Elem: &inner}
		}
	case *ast.IndexListExpr:
		if isOptionalType(current.X) && len(current.Indices) == 1 {
			inner := e.parseType(current.Indices[0], sf)
			return typeRef{Kind: "optional", Elem: &inner}
		}
	}
	return typeRef{Kind: "unknown"}
}

func isOptionalType(expr ast.Expr) bool {
	switch current := expr.(type) {
	case *ast.Ident:
		return current.Name == "Optional"
	case *ast.SelectorExpr:
		return current.Sel.Name == "Optional"
	default:
		return false
	}
}

func isIORef(ref typeRef) bool {
	for ref.Kind == "pointer" && ref.Elem != nil {
		ref = *ref.Elem
	}
	return ref.Kind == "named" && shortType(ref.Name) == "IO"
}

func (e *extractor) resolveExprType(expr ast.Expr, method *marshalInfo, env map[string]typeRef) typeRef {
	switch current := expr.(type) {
	case *ast.UnaryExpr:
		return e.resolveExprType(current.X, method, env)
	case *ast.ParenExpr:
		return e.resolveExprType(current.X, method, env)
	case *ast.StarExpr:
		return e.resolveExprType(current.X, method, env)
	case *ast.Ident:
		if ref, ok := env[current.Name]; ok {
			return ref
		}
		return typeRef{Kind: "unknown"}
	case *ast.IndexExpr:
		ref := e.underlyingArray(e.resolveExprType(current.X, method, env))
		if (ref.Kind == "slice" || ref.Kind == "array") && ref.Elem != nil {
			return *ref.Elem
		}
		return typeRef{Kind: "unknown"}
	case *ast.SelectorExpr:
		ref := e.resolveExprType(current.X, method, env)
		for ref.Kind == "pointer" && ref.Elem != nil {
			ref = *ref.Elem
		}
		if ref.Kind != "named" {
			return typeRef{Kind: "unknown"}
		}
		return e.lookupField(ref.Name, current.Sel.Name, map[string]bool{})
	case *ast.CompositeLit:
		return e.parseType(current.Type, method.File)
	default:
		return typeRef{Kind: "unknown"}
	}
}

// underlyingArray strips pointers and resolves a named fixed-size array type
// (type HeightMap [16][16]int8) to its array type.
func (e *extractor) underlyingArray(ref typeRef) typeRef {
	for ref.Kind == "pointer" && ref.Elem != nil {
		ref = *ref.Elem
	}
	if ref.Kind == "named" {
		if array, ok := e.arrays[ref.Name]; ok {
			return array
		}
	}
	return ref
}

func (e *extractor) lookupField(key, name string, seen map[string]bool) typeRef {
	if seen[key] {
		return typeRef{Kind: "unknown"}
	}
	seen[key] = true
	info := e.types[key]
	if info == nil {
		return typeRef{Kind: "unknown"}
	}
	if ref, ok := info.Fields[name]; ok {
		return ref
	}
	for _, embedded := range info.Embedded {
		for embedded.Kind == "pointer" && embedded.Elem != nil {
			embedded = *embedded.Elem
		}
		if embedded.Kind == "named" {
			if ref := e.lookupField(embedded.Name, name, seen); ref.Kind != "unknown" {
				return ref
			}
		}
	}
	return typeRef{Kind: "unknown"}
}

func (e *extractor) fieldPath(expr ast.Expr, method *marshalInfo, env map[string]typeRef, fallback string) string {
	var parts []string
	root := ""
	var walk func(ast.Expr) bool
	walk = func(current ast.Expr) bool {
		switch value := current.(type) {
		case *ast.UnaryExpr:
			return walk(value.X)
		case *ast.ParenExpr:
			return walk(value.X)
		case *ast.IndexExpr:
			if !walk(value.X) {
				return false
			}
			parts = append(parts, "[]")
			return true
		case *ast.Ident:
			if value.Name == method.Recv {
				return true
			}
			if ref, ok := env[value.Name]; ok {
				root = ref.Path
				return true
			}
			return false
		case *ast.SelectorExpr:
			if !walk(value.X) {
				return false
			}
			parts = append(parts, value.Sel.Name)
			return true
		default:
			return false
		}
	}
	if !walk(expr) || len(parts) == 0 {
		return fallback
	}
	path := fallback
	if root != "" {
		path = root
	}
	for _, part := range parts {
		if part == "[]" {
			path += part
		} else if path == "" {
			path = part
		} else {
			path += "." + part
		}
	}
	return path
}

func hasEnv(env map[string]typeRef, name string) bool {
	_, ok := env[name]
	return ok
}

func (e *extractor) literalInt(expr ast.Expr, sf *sourceFile) (int, bool) {
	return e.evalConst(expr, 0, sf.Pkg)
}

func (t typeRef) String() string {
	if t.Kind == "named" || t.Kind == "primitive" {
		return t.Name
	}
	if t.Elem != nil {
		return t.Kind + "[" + t.Elem.String() + "]"
	}
	return t.Kind
}

func unwrapIndex(expr ast.Expr) ast.Expr {
	switch current := expr.(type) {
	case *ast.IndexExpr:
		return current.X
	case *ast.IndexListExpr:
		return current.X
	default:
		return expr
	}
}

func referencesIO(call *ast.CallExpr, ioName string) bool {
	for _, arg := range call.Args {
		if ident, ok := arg.(*ast.Ident); ok && ident.Name == ioName {
			return true
		}
		if selector, ok := arg.(*ast.SelectorExpr); ok {
			if base, ok := selector.X.(*ast.Ident); ok && base.Name == ioName {
				return true
			}
		}
	}
	return false
}

func (e *extractor) unresolved(node ast.Node, typ, field, reason, raw string) sourceOperation {
	file := ""
	line := 0
	site := ""
	if node != nil {
		position := e.fset.Position(node.Pos())
		file = filepath.ToSlash(position.Filename)
		line = position.Line
		site = file + ":" + strconv.Itoa(line)
	}
	e.diagnostics = append(e.diagnostics, diagnostic{Packet: e.packet, Type: typ, Field: field, File: file, Line: line, Reason: reason, Raw: raw})
	return sourceOperation{Kind: "unresolved", Field: field, Reason: reason, Site: site}
}

func (e *extractor) addDiagnostic(node ast.Node, typ, field, reason, raw string) {
	_ = e.unresolved(node, typ, field, reason, raw)
}

func (e *extractor) nodeString(node ast.Node) string {
	if node == nil {
		return ""
	}
	var builder strings.Builder
	_ = printer.Fprint(&builder, e.fset, node)
	result := builder.String()
	if len(result) > 300 {
		return result[:300] + "..."
	}
	return result
}

func shortType(name string) string {
	if index := strings.LastIndexByte(name, '.'); index >= 0 {
		return name[index+1:]
	}
	return name
}
