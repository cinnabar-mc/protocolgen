// Package emitgo emits a standalone Go view from the canonical v2 manifest.
// It receives no source documents; target-specific conveniences are selected
// explicitly through Options.
package emitgo

import (
	"embed"
	"fmt"
	"go/format"
	"math"
	"slices"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"protocolgen/internal/docs"
	"protocolgen/internal/domains"
	"protocolgen/internal/flatten"
	"protocolgen/internal/layout"
	"protocolgen/internal/licensing"
	"protocolgen/internal/manifest"
	"protocolgen/internal/naming"
	"protocolgen/internal/semantics"
)

type typeDefinition struct {
	Name       string
	TypeID     string
	Kind       manifest.NodeKind
	Fields     []typedField
	EntryKey   string
	EntryValue string
	Underlying string
	Primitive  string
	Variants   []manifest.Variant
	Union      []goUnionMember
	Implements []unionMembership
	BitLength  uint64
}

// unionMembership records the tag a struct carries when written as a variant
// of one union; a struct may belong to several unions with different tags.
type unionMembership struct {
	Union   string
	Tag     int64
	TagType string
}

type typedField struct {
	Name     string
	Hint     string // type-name hint for anonymous nodes: the exported wire name, never a layout name
	WireName string
	Type     string
	Node     manifest.Node
}

type goUnionMember struct {
	Name  string
	Value int64
	Node  manifest.Node
}

type generator struct {
	definitions        map[string]typeDefinition
	identity           map[string]string
	usedNames          map[string]bool
	resolver           *naming.Resolver
	protocolImportPath string
	nativeTypes        bool
	emitPacketRuntime  bool
	emitPacketPools    bool
	domains            domains.Overlay
	docs               docs.Overlay
	layout             layout.Overlay
	semantics          semantics.Overlay
	usage              flatten.Usage
	enumIdentities     map[string]int      // enum definitions per type ID; layout constants apply only to unique ones
	packetConstants    map[string][]string // packet file stem -> const blocks relocated there by the layout overlay
}

// isNative reports whether a node is emitted as an established Go type
// rather than a generated struct.
func (g *generator) isNative(node manifest.Node) bool {
	_, matched, _ := g.nativeGoType(node)
	return matched
}

// packetFields is a packet's effective field list: single-use payload structs
// are inlined, so their fields belong to the packet as in the hand-written
// packages.
func (g *generator) packetFields(packet manifest.Packet) []flatten.Field {
	return g.usage.PacketFields(packet, g.isNative)
}

//go:embed runtime/codec.go runtime/helpers.go runtime/reader.go runtime/types.go runtime/writer.go
var runtimeSource embed.FS

func emitRuntimeFiles() (map[string]string, error) {
	files := make(map[string]string, 5)
	for _, name := range []string{"codec.go", "helpers.go", "reader.go", "types.go", "writer.go"} {
		data, err := runtimeSource.ReadFile("runtime/" + name)
		if err != nil {
			return nil, fmt.Errorf("read embedded runtime %s: %w", name, err)
		}
		files["protocol/"+name] = string(data)
	}
	return files, nil
}

// Options controls the target-specific conveniences emitted alongside the
// canonical wire representation. NativeTypes enables established Go types
// such as uuid.UUID and mgl32 vectors; disabling it keeps every named shape in
// generated protocol structs. Packet runtime and pool emission are enabled by
// Generate and can be disabled by callers that only need definitions.
type Options struct {
	ProtocolImportPath string
	Naming             naming.Overlay
	Domains            domains.Overlay
	Docs               docs.Overlay
	Layout             layout.Overlay
	Semantics          semantics.Overlay
	NativeTypes        bool
	EmitPacketRuntime  bool
	EmitPacketPools    bool
}

// Generate emits a protocol package and its packet subpackage. The protocol
// import path is embedded in packet files so the generated tree can be used as
// an ordinary Go package from any parent module.
func Generate(m manifest.Manifest, protocolImportPath string) (map[string]string, error) {
	return GenerateWithOptions(m, Options{
		ProtocolImportPath: protocolImportPath,
		NativeTypes:        true,
		EmitPacketRuntime:  true,
		EmitPacketPools:    true,
	})
}

// GenerateWithOptions emits a protocol package using explicit target options.
// The canonical manifest remains the source of truth; options only affect the
// generated API surface and Go type mappings.
func GenerateWithOptions(m manifest.Manifest, options Options) (map[string]string, error) {
	if err := manifest.Validate(m); err != nil {
		return nil, err
	}
	if err := naming.ValidateRequiredEntries(m, options.Naming); err != nil {
		return nil, err
	}
	if options.Domains.Domains != nil {
		if err := domains.ValidateAssignments(m, options.Domains); err != nil {
			return nil, err
		}
	}
	if options.ProtocolImportPath == "" || strings.ContainsAny(options.ProtocolImportPath, " \t\r\n") {
		return nil, fmt.Errorf("invalid protocol import path %q", options.ProtocolImportPath)
	}
	// Layout type names override the naming overlay for shared types; packets
	// are named directly below.
	if len(options.Layout.Types) > 0 {
		names := make(map[string]string, len(options.Naming.Names)+len(options.Layout.Types))
		for typeID, name := range options.Naming.Names {
			names[typeID] = name
		}
		for typeID, name := range options.Layout.Types {
			names[typeID] = name
		}
		options.Naming = naming.Overlay{Names: names}
	}
	g := &generator{
		definitions:        map[string]typeDefinition{},
		identity:           map[string]string{},
		usedNames:          map[string]bool{},
		resolver:           naming.NewResolver(options.Naming),
		protocolImportPath: options.ProtocolImportPath,
		nativeTypes:        options.NativeTypes,
		emitPacketRuntime:  options.EmitPacketRuntime,
		emitPacketPools:    options.EmitPacketPools,
		domains:            options.Domains,
		docs:               options.Docs,
		layout:             options.Layout,
		semantics:          options.Semantics,
		usage:              flatten.Count(m),
		packetConstants:    map[string][]string{},
	}
	packets := append([]manifest.Packet(nil), m.Packets...)
	sort.Slice(packets, func(i, j int) bool { return packets[i].ID < packets[j].ID })
	packetNames := map[uint32]string{}
	for _, packet := range packets {
		neutral := naming.PacketTypeName(packet.Name)
		if reviewed := g.layout.TypeName(packet.Name); reviewed != "" {
			neutral = reviewed
		}
		name := exportName(neutral)
		if err := g.resolver.Reserve(packet.Name, neutral, exportName); err != nil {
			return nil, fmt.Errorf("packet %s: %w", packet.Name, err)
		}
		g.usedNames[name] = true
		packetNames[packet.ID] = name
		for _, effective := range g.packetFields(packet) {
			field := effective.Field
			if err := ensureCodecSymmetric(field); err != nil {
				return nil, fmt.Errorf("packet %s field %s: %w", packet.Name, field.Name, err)
			}
			if _, err := g.goType(g.semantics.Apply(effective.Owner, field), name+exportName(field.Name)); err != nil {
				return nil, fmt.Errorf("packet %s field %s: %w", packet.Name, field.Name, err)
			}
		}
	}
	files, err := g.emitFiles(m, packets, packetNames)
	if err != nil {
		return nil, err
	}
	return files, nil
}

func ensureCodecSymmetric(field manifest.Field) error {
	if field.Decode != nil || field.Symmetry != manifest.Symmetric {
		return fmt.Errorf("asymmetric encode/decode layouts require separate codec methods")
	}
	return ensureNodeCodecSymmetric(field.Encode)
}

func ensureNodeCodecSymmetric(node manifest.Node) error {
	for _, field := range node.Fields {
		if err := ensureCodecSymmetric(field); err != nil {
			return fmt.Errorf("nested field %s: %w", field.Name, err)
		}
	}
	for _, variant := range node.Variants {
		if variant.Decode != nil {
			return fmt.Errorf("union variant %s has an asymmetric decode layout", variant.Name)
		}
		if err := ensureNodeCodecSymmetric(variant.Encode); err != nil {
			return err
		}
	}
	for _, child := range []*manifest.Node{node.Prefix, node.Element, node.Value, node.Key, node.Control, node.Default} {
		if child != nil {
			if err := ensureNodeCodecSymmetric(*child); err != nil {
				return err
			}
		}
	}
	for _, child := range node.Elements {
		if err := ensureNodeCodecSymmetric(child); err != nil {
			return err
		}
	}
	for _, oneCase := range node.Cases {
		if len(oneCase.Decode) != 0 {
			return fmt.Errorf("conditional case %s has an asymmetric decode layout", oneCase.Value)
		}
		for _, child := range oneCase.Encode {
			if err := ensureNodeCodecSymmetric(child); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *generator) goType(node manifest.Node, hint string) (string, error) {
	if typ, matched, err := g.nativeGoType(node); matched || err != nil {
		return typ, err
	}
	switch node.Kind {
	case manifest.KindPrimitive:
		if node.Primitive == nil {
			return "", fmt.Errorf("primitive has no shape")
		}
		if node.Primitive.Code == "uuid" {
			return "[16]byte", nil
		}
		return primitiveGoType(node.Primitive.Code)
	case manifest.KindString:
		return "string", nil
	case manifest.KindBytes:
		return "[]byte", nil
	case manifest.KindBitset:
		name := fmt.Sprintf("Bitset%d", node.Length)
		if _, exists := g.definitions[name]; !exists {
			g.definitions[name] = typeDefinition{Name: name, Kind: manifest.KindBitset, BitLength: node.Length}
		}
		return name, nil
	case manifest.KindArray:
		if node.Element == nil {
			return "", fmt.Errorf("array has no element")
		}
		element, err := g.goType(*node.Element, hint+"Item")
		if err != nil {
			return "", err
		}
		return "[]" + element, nil
	case manifest.KindFixedArray:
		if node.Element == nil {
			return "", fmt.Errorf("fixed array has no element")
		}
		element, err := g.goType(*node.Element, hint+"Item")
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("[%d]%s", node.Length, element), nil
	case manifest.KindSequence:
		return "", fmt.Errorf("sequence nodes require a target-specific representation")
	case manifest.KindOptional:
		if node.Value == nil {
			return "", fmt.Errorf("optional has no value")
		}
		valueNode := *node.Value
		// Cereal double optionals carry an always-present outer marker. Keep both
		// nodes in the manifest for codec generation, but expose only the
		// meaningful inner presence state, matching gophertunnel's Optional[T].
		if valueNode.Kind == manifest.KindOptional {
			if valueNode.Value == nil {
				return "", fmt.Errorf("nested optional has no value")
			}
			valueNode = *valueNode.Value
		}
		value, err := g.goType(valueNode, hint+"Value")
		if err != nil {
			return "", err
		}
		return "Optional[" + value + "]", nil
	case manifest.KindStruct:
		return g.registerStruct(node, hint)
	case manifest.KindMap:
		if node.Key == nil || node.Value == nil {
			return "", fmt.Errorf("map has no key/value")
		}
		key, err := g.goType(*node.Key, hint+"Key")
		if err != nil {
			return "", err
		}
		value, err := g.goType(*node.Value, hint+"Value")
		if err != nil {
			return "", err
		}
		return "[]OrderedEntry[" + key + ", " + value + "]", nil
	case manifest.KindUnion:
		name, err := g.registerIdentity(node, hint+"Union")
		if err != nil {
			return "", err
		}
		if _, exists := g.definitions[name]; !exists {
			if node.Control == nil || node.Control.Primitive == nil {
				return "", fmt.Errorf("union has no primitive discriminator")
			}
			tagType, err := primitiveGoType(node.Control.Primitive.Code)
			if err != nil {
				return "", err
			}
			g.definitions[name] = typeDefinition{Name: name, TypeID: nodeTypeID(node), Kind: manifest.KindUnion, Underlying: tagType, Primitive: node.Control.Primitive.Code}
			members := make([]goUnionMember, 0, len(node.Variants))
			usedMembers := map[string]bool{}
			for _, variant := range node.Variants {
				if !fitsGoInteger(variant.Value, tagType) {
					return "", fmt.Errorf("union %s variant %s tag %d does not fit %s", name, variant.Name, variant.Value, tagType)
				}
				membership := unionMembership{Union: name, Tag: variant.Value, TagType: tagType}
				member, err := g.registerUnionMember(membership, variant)
				if err != nil {
					return "", err
				}
				if usedMembers[member] {
					wrapper := g.unique(name + exportName(naming.PublicVariantName(shortTypeName(variant.Name))))
					g.definitions[wrapper] = typeDefinition{
						Name:       wrapper,
						TypeID:     nodeTypeID(variant.Encode),
						Kind:       manifest.KindStruct,
						Fields:     []typedField{{Name: "Value", WireName: "Value", Type: member, Node: variant.Encode}},
						Implements: []unionMembership{membership},
					}
					member = wrapper
				}
				usedMembers[member] = true
				members = append(members, goUnionMember{Name: member, Value: variant.Value, Node: variant.Encode})
			}
			definition := g.definitions[name]
			definition.Union = members
			g.definitions[name] = definition
		}
		return name, nil
	case manifest.KindEnum:
		if node.Primitive == nil {
			return "", fmt.Errorf("enum has no underlying primitive")
		}
		if len(node.Variants) == 0 {
			return "", fmt.Errorf("enum has no variants")
		}
		underlying, err := primitiveGoType(node.Primitive.Code)
		if err != nil {
			return "", err
		}
		name, err := g.registerIdentity(node, hint+"Enum")
		if err != nil {
			return "", err
		}
		if existing, exists := g.definitions[name]; !exists {
			g.definitions[name] = typeDefinition{Name: name, TypeID: nodeTypeID(node), Kind: manifest.KindEnum, Underlying: underlying, Primitive: node.Primitive.Code, Variants: append([]manifest.Variant(nil), node.Variants...)}
		} else if existing.Primitive != node.Primitive.Code {
			return "", fmt.Errorf("enum %s is encoded as both %s and %s", name, existing.Primitive, node.Primitive.Code)
		} else {
			if existing.Variants, err = manifest.MergeEnumVariants(name, existing.Variants, node.Variants); err != nil {
				return "", err
			}
			g.definitions[name] = existing
		}
		return name, nil
	case manifest.KindReserved, manifest.KindIgnored:
		return "", fmt.Errorf("%s nodes require explicit write/discard codec semantics", node.Kind)
	case manifest.KindRecursive:
		name, ok := g.identity[node.Target]
		if !ok {
			return "", fmt.Errorf("recursive target %q is not a registered named type", node.Target)
		}
		return name, nil
	case manifest.KindVoid:
		return "struct{}", nil
	case manifest.KindOpaque, manifest.KindUnresolved:
		return "", fmt.Errorf("reachable %s node cannot be emitted: %s", node.Kind, node.Reason)
	default:
		return "", fmt.Errorf("unsupported node kind %q", node.Kind)
	}
}

// nativeGoType maps canonical wire semantics to established Go ecosystem types.
func nativeGoType(node manifest.Node) (string, bool, error) {
	return nativeGoTypeWithOptions(node, true)
}

func (g *generator) nativeGoType(node manifest.Node) (string, bool, error) {
	return nativeGoTypeWithOptions(node, g.nativeTypes)
}

func nativeGoTypeWithOptions(node manifest.Node, enabled bool) (string, bool, error) {
	if !enabled {
		return "", false, nil
	}
	if node.Kind == manifest.KindPrimitive && node.Primitive != nil {
		switch node.Primitive.Code {
		case "uuid":
			return "uuid.UUID", true, nil
		}
	}
	if node.Kind != manifest.KindStruct {
		return "", false, nil
	}
	switch node.TypeID {
	case "ActorRuntimeID":
		switch code, ok := primitiveStructCode(node); {
		case ok && code == "var_u64":
			return "uint64", true, nil
		case ok && code == "zigzag_i64":
			return "int64", true, nil
		case ok && code == "var_u32":
			return "uint32", true, nil
		case ok && code == "var_i64":
			return "int64", true, nil
		default:
			return "", true, fmt.Errorf("native ActorRuntimeID mapping requires a supported variable-width ID field")
		}
	case "ActorUniqueID":
		switch code, ok := primitiveStructCode(node); {
		case ok && code == "zigzag_i64":
			return "int64", true, nil
		case ok && code == "i64le":
			return "int64", true, nil
		case ok && code == "u64le":
			return "uint64", true, nil
		case ok && code == "var_u64":
			return "uint64", true, nil
		case ok && code == "var_i64":
			return "int64", true, nil
		default:
			return "", true, fmt.Errorf("native ActorUniqueID mapping requires a supported signed, unsigned, or fixed-width ID field")
		}
	case "PlayerInputTick":
		if !isPrimitiveStruct(node, "var_u64") {
			return "", true, fmt.Errorf("native PlayerInputTick mapping requires exactly one var_u64 field")
		}
		return "uint64", true, nil
	case "Vec2":
		if !isPrimitiveStruct(node, "f32le", "f32le") {
			return "", true, fmt.Errorf("native Vec2 mapping requires exactly two f32le fields")
		}
		return "mgl32.Vec2", true, nil
	case "Vec3":
		if !isPrimitiveStruct(node, "f32le", "f32le", "f32le") {
			return "", true, fmt.Errorf("native Vec3 mapping requires exactly three f32le fields")
		}
		return "mgl32.Vec3", true, nil
	case "mce::Color":
		if !isPrimitiveStruct(node, "i32le") {
			return "", true, fmt.Errorf("native colour mapping requires exactly one i32le field")
		}
		return "color.RGBA", true, nil
	default:
		return "", false, nil
	}
}

func isPrimitiveStruct(node manifest.Node, codes ...string) bool {
	if len(node.Fields) != len(codes) {
		return false
	}
	for index, code := range codes {
		field := node.Fields[index].Encode
		if field.Kind != manifest.KindPrimitive || field.Primitive == nil || field.Primitive.Code != code {
			return false
		}
	}
	return true
}

func primitiveStructCode(node manifest.Node) (string, bool) {
	if len(node.Fields) != 1 {
		return "", false
	}
	field := node.Fields[0].Encode
	if field.Kind != manifest.KindPrimitive || field.Primitive == nil {
		return "", false
	}
	return field.Primitive.Code, true
}

func (g *generator) registerUnionMember(membership unionMembership, variant manifest.Variant) (string, error) {
	union := membership.Union
	member, err := g.goType(variant.Encode, union+exportName(naming.PublicVariantName(shortTypeName(variant.Name))))
	if err != nil {
		return "", err
	}
	if variant.Encode.Kind == manifest.KindVoid {
		member = g.unique(union + exportName(naming.PublicVariantName(shortTypeName(variant.Name))))
		g.definitions[member] = typeDefinition{Name: member, TypeID: nodeTypeID(variant.Encode), Kind: manifest.KindStruct}
	}
	definition, ok := g.definitions[member]
	if !ok || definition.Kind != manifest.KindStruct {
		wrapper := g.unique(union + exportName(naming.PublicVariantName(shortTypeName(variant.Name))))
		g.definitions[wrapper] = typeDefinition{Name: wrapper, TypeID: nodeTypeID(variant.Encode), Kind: manifest.KindStruct, Fields: []typedField{{Name: "Value", WireName: "Value", Type: member, Node: variant.Encode}}, Implements: []unionMembership{membership}}
		return wrapper, nil
	}
	if !slices.ContainsFunc(definition.Implements, func(existing unionMembership) bool { return existing.Union == union }) {
		definition.Implements = append(definition.Implements, membership)
	}
	g.definitions[member] = definition
	return member, nil
}

// fitsGoInteger reports whether value is representable by the named Go integer type.
func fitsGoInteger(value int64, typ string) bool {
	switch typ {
	case "int8":
		return value >= math.MinInt8 && value <= math.MaxInt8
	case "int16":
		return value >= math.MinInt16 && value <= math.MaxInt16
	case "int32":
		return value >= math.MinInt32 && value <= math.MaxInt32
	case "int64":
		return true
	case "uint8":
		return value >= 0 && value <= math.MaxUint8
	case "uint16":
		return value >= 0 && value <= math.MaxUint16
	case "uint32":
		return value >= 0 && value <= math.MaxUint32
	case "uint64":
		return value >= 0
	default:
		return false
	}
}

func nodeTypeID(node manifest.Node) string {
	if node.TypeID != "" {
		return node.TypeID
	}
	return naming.InferredTypeName(node)
}

func (g *generator) registerStruct(node manifest.Node, hint string) (string, error) {
	name, err := g.registerIdentity(node, hint+"Struct")
	if err != nil {
		return "", err
	}
	if _, exists := g.definitions[name]; exists {
		return name, nil
	}
	g.definitions[name] = typeDefinition{Name: name, TypeID: nodeTypeID(node), Kind: manifest.KindStruct}
	used, usedHints := map[string]bool{}, map[string]bool{}
	var fields []typedField
	for _, field := range node.Fields {
		fieldName := uniqueFieldName(g.fieldName(nodeTypeID(node), field.Name), used)
		hint := uniqueFieldName(exportName(field.Name), usedHints)
		encode := g.semantics.Apply(nodeTypeID(node), field)
		fieldType, err := g.goType(encode, name+hint)
		if err != nil {
			return "", err
		}
		fields = append(fields, typedField{Name: fieldName, Hint: hint, WireName: field.Name, Type: fieldType, Node: encode})
	}
	definition := g.definitions[name]
	definition.Fields = fields
	g.definitions[name] = definition
	return name, nil
}

// fieldName is the reviewed layout name for a field, else its exported wire name.
func (g *generator) fieldName(owner, wire string) string {
	if name := g.layout.FieldName(owner, wire); name != "" {
		return name
	}
	return exportName(wire)
}

func (g *generator) registerIdentity(node manifest.Node, hint string) (string, error) {
	name, err := g.resolver.Resolve(node, hint, exportName)
	if err != nil {
		return "", err
	}
	key := naming.IdentityKeyFor(node, hint)
	g.identity[key] = name
	g.usedNames[name] = true
	return name, nil
}

func (g *generator) unique(base string) string {
	if base == "" {
		base = "GeneratedType"
	}
	if !g.usedNames[base] {
		g.usedNames[base] = true
		return base
	}
	for index := 2; ; index++ {
		candidate := fmt.Sprintf("%s%d", base, index)
		if !g.usedNames[candidate] {
			g.usedNames[candidate] = true
			return candidate
		}
	}
}

func emitVersion(m manifest.Manifest) string {
	return fmt.Sprintf(`// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

const (
	GAME_VERSION     = %q
	PROTOCOL_VERSION = %d
)
`, m.Target.MinecraftVersion, m.Target.ProtocolVersion)
}

func (g *generator) emitFiles(m manifest.Manifest, packets []manifest.Packet, packetNames map[uint32]string) (map[string]string, error) {
	definitions := make([]typeDefinition, 0, len(g.definitions))
	g.enumIdentities = map[string]int{}
	for _, definition := range g.definitions {
		definitions = append(definitions, definition)
		if definition.Kind == manifest.KindEnum {
			g.enumIdentities[definition.TypeID]++
		}
	}
	sort.Slice(definitions, func(i, j int) bool { return definitions[i].Name < definitions[j].Name })

	files, err := emitRuntimeFiles()
	if err != nil {
		return nil, err
	}
	files["LICENSE-MIT"] = licensing.MIT
	files["NOTICE"] = licensing.Notice
	files["protocol/version.go"] = emitVersion(m)
	usedFiles := map[string]bool{
		"types.go": true, "codec.go": true, "helpers.go": true, "reader.go": true, "writer.go": true, "version.go": true,
	}
	if g.domains.Domains == nil {
		for _, definition := range definitions {
			source, err := emitDefinitionFile(g, definition)
			if err != nil {
				return nil, err
			}
			stem := g.layout.File(definition.TypeID)
			if stem == "" {
				stem = snakeName(definition.Name)
			}
			name := uniqueFileName(stem+".go", 0, usedFiles)
			files["protocol/"+name] = source
		}
	} else {
		byDomain := map[string][]typeDefinition{}
		for _, definition := range definitions {
			domain := g.domainFor(definition)
			byDomain[domain] = append(byDomain[domain], definition)
		}
		domainsByName := make([]string, 0, len(byDomain))
		for domain := range byDomain {
			domainsByName = append(domainsByName, domain)
		}
		sort.Strings(domainsByName)
		for _, domain := range domainsByName {
			source, err := emitDomainFile(g, byDomain[domain])
			if err != nil {
				return nil, err
			}
			name := uniqueFileName(domain+".go", 0, usedFiles)
			files["protocol/"+name] = source
		}
	}
	packetFiles := map[string]string{}
	packetUsed := map[string]bool{"ids.go": true}
	if g.emitPacketRuntime {
		packetUsed["packet.go"] = true
	}
	if g.emitPacketPools {
		packetUsed["pool.go"] = true
	}
	packetFiles["ids.go"] = emitPacketIDs(packets, packetNames)
	if g.emitPacketRuntime {
		packetFiles["packet.go"] = emitPacketRuntime(g.protocolImportPath)
	}
	if g.emitPacketPools {
		packetFiles["pool.go"] = emitPacketPools(packets, packetNames)
	}
	for _, packet := range packets {
		packetName := packetNames[packet.ID]
		stem := snakeName(packetName)
		if reviewed := g.layout.File(packet.Name); reviewed != "" {
			stem = reviewed
		}
		name := uniqueFileName(stem+".go", packet.ID, packetUsed)
		// Constants relocated to this file go above the packet, where the
		// hand-written packages keep them.
		constants := g.packetConstants[stem]
		delete(g.packetConstants, stem)
		source, err := g.emitPacket(packet, packetName, constants)
		if err != nil {
			return nil, err
		}
		packetFiles[name] = source
	}
	for stem, blocks := range g.packetConstants {
		name := stem + ".go"
		source := fmt.Sprintf("// Code generated from canonical protocol manifest v2. DO NOT EDIT.\n\npackage packet\n\nimport %q\n\n%s", g.protocolImportPath, strings.Join(blocks, "\n"))
		formatted, err := formatGoSource(source)
		if err != nil {
			return nil, fmt.Errorf("packet constants in %s: %w", name, err)
		}
		packetFiles[name] = formatted
	}
	for name, source := range packetFiles {
		files["protocol/packet/"+name] = source
	}
	return files, nil
}

func (g *generator) domainFor(definition typeDefinition) string {
	if file := g.layout.File(definition.TypeID); file != "" {
		return file
	}
	if definition.TypeID == "" {
		return "generated"
	}
	return g.domains.Domain(definition.TypeID)
}

func emitDefinitionFile(g *generator, definition typeDefinition) (string, error) {
	source, err := emitDefinitionBody(g, definition)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString("// Code generated from canonical protocol manifest v2. DO NOT EDIT.\n\npackage protocol\n\n")
	writeGoImports(&b, goImportsForTypes(definitionTypes(definition)))
	b.WriteString(source)
	return formatGoSource(b.String())
}

func emitDomainFile(g *generator, definitions []typeDefinition) (string, error) {
	sort.Slice(definitions, func(i, j int) bool { return definitions[i].Name < definitions[j].Name })
	var b strings.Builder
	b.WriteString("// Code generated from canonical protocol manifest v2. DO NOT EDIT.\n\npackage protocol\n\n")
	var types []string
	for _, definition := range definitions {
		types = append(types, definitionTypes(definition)...)
	}
	writeGoImports(&b, goImportsForTypes(types))
	for index, definition := range definitions {
		if index != 0 {
			b.WriteString("\n")
		}
		source, err := emitDefinitionBody(g, definition)
		if err != nil {
			return "", err
		}
		b.WriteString(source)
	}
	return formatGoSource(b.String())
}

func definitionTypes(definition typeDefinition) []string {
	types := make([]string, 0, len(definition.Fields))
	for _, field := range definition.Fields {
		types = append(types, field.Type)
	}
	return types
}

func emitDefinitionBody(g *generator, definition typeDefinition) (string, error) {
	var b strings.Builder
	for _, line := range docs.GoComments(g.docs.Type(definition.TypeID)) {
		b.WriteString(line)
		b.WriteByte('\n')
	}
	switch definition.Kind {
	case manifest.KindStruct:
		fmt.Fprintf(&b, "type %s struct {\n", definition.Name)
		for _, field := range definition.Fields {
			for _, line := range docs.GoComments(g.docs.Field(definition.TypeID, field.WireName)) {
				fmt.Fprintf(&b, "\t%s\n", line)
			}
			fmt.Fprintf(&b, "\t%s %s\n", field.Name, field.Type)
		}
		b.WriteString("}\n\n")
		for _, membership := range definition.Implements {
			fmt.Fprintf(&b, "func (*%s) tag%s() %s { return %d }\n\n", definition.Name, membership.Union, membership.TagType, membership.Tag)
		}
		fmt.Fprintf(&b, "// Marshal reads or writes %s using its canonical wire layout.\n", definition.Name)
		fmt.Fprintf(&b, "func (x *%s) Marshal(io IO) {\n", definition.Name)
		emitter := marshalEmitter{g: g}
		for _, field := range definition.Fields {
			hint := field.Hint
			if hint == "" {
				hint = field.Name
			}
			if err := emitter.node(&b, field.Node, "x."+field.Name, definition.Name+hint, "\t", addressStrategy{}); err != nil {
				return "", fmt.Errorf("type %s field %s marshal: %w", definition.Name, field.Name, err)
			}
		}
		b.WriteString("}\n")
	case manifest.KindUnion:
		fmt.Fprintf(&b, "type %s interface {\n\tMarshaler\n\ttag%s() %s\n}\n\n", definition.Name, definition.Name, definition.Underlying)
		emitter := marshalEmitter{g: g}
		if err := emitter.union(&b, definition); err != nil {
			return "", err
		}
	case manifest.KindEnum:
		fmt.Fprintf(&b, "type %s %s\n\n", definition.Name, definition.Underlying)
		placement, placed := g.layout.Constants[definition.TypeID]
		if g.enumIdentities[definition.TypeID] != 1 {
			// Anonymous enums can share an inferred identity; a reviewed
			// placement cannot tell them apart, so neither gets it.
			placement, placed = layout.Placement{}, false
		}
		constants, err := enumConstants(definition, placement)
		if err != nil {
			return "", err
		}
		if placed && placement.Package == "packet" {
			g.packetConstants[placement.File] = append(g.packetConstants[placement.File], constants)
		} else {
			b.WriteString(constants)
			b.WriteString("\n")
		}
		method, err := primitiveIOMethod(definition.Primitive)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&b, "// Marshal reads or writes %s through its %s wire encoding.\n", definition.Name, definition.Underlying)
		fmt.Fprintf(&b, "func (x *%s) Marshal(io IO) { io.%s((*%s)(x)) }\n", definition.Name, method, definition.Underlying)
	case manifest.KindBitset:
		fmt.Fprintf(&b, "// %s stores the %d-bit value used by the wire bitset encoding.\n", definition.Name, definition.BitLength)
		fmt.Fprintf(&b, "type %s [%d]uint64\n", definition.Name, (definition.BitLength+63)/64)
		fmt.Fprintf(&b, "\nconst %sLength = %d\n\n", definition.Name, definition.BitLength)
		fmt.Fprintf(&b, "// Set marks bit index i. It panics when i is outside [0, %d).\n", definition.BitLength)
		fmt.Fprintf(&b, "func (b *%s) Set(i int) {\n", definition.Name)
		fmt.Fprintf(&b, "\tif i < 0 || i >= %sLength {\n\t\tpanic(\"index out of bounds\")\n\t}\n", definition.Name)
		fmt.Fprintf(&b, "\tb[i/64] |= uint64(1) << uint(i%%64)\n}\n\n")
		fmt.Fprintf(&b, "// Unset clears bit index i. It panics when i is outside [0, %d).\n", definition.BitLength)
		fmt.Fprintf(&b, "func (b *%s) Unset(i int) {\n", definition.Name)
		fmt.Fprintf(&b, "\tif i < 0 || i >= %sLength {\n\t\tpanic(\"index out of bounds\")\n\t}\n", definition.Name)
		fmt.Fprintf(&b, "\tb[i/64] &^= uint64(1) << uint(i%%64)\n}\n\n")
		fmt.Fprintf(&b, "// Load reports whether bit index i is set. It panics when i is outside [0, %d).\n", definition.BitLength)
		fmt.Fprintf(&b, "func (b %s) Load(i int) bool {\n", definition.Name)
		fmt.Fprintf(&b, "\tif i < 0 || i >= %sLength {\n\t\tpanic(\"index out of bounds\")\n\t}\n", definition.Name)
		fmt.Fprintf(&b, "\treturn b[i/64]&(uint64(1)<<uint(i%%64)) != 0\n}\n\n")
		fmt.Fprintf(&b, "// Len returns the number of bits in the bitset.\nfunc (b %s) Len() int { return %sLength }\n", definition.Name, definition.Name)
	default:
		return "", fmt.Errorf("unsupported definition kind %q", definition.Kind)
	}
	return b.String(), nil
}

// enumConstants renders an enum's constant block. A layout placement renames
// variants and, for the packet package, qualifies the enum type.
func enumConstants(definition typeDefinition, placement layout.Placement) (string, error) {
	qualifier := ""
	if placement.Package == "packet" {
		qualifier = "protocol."
	}
	var b strings.Builder
	b.WriteString("const (\n")
	used := map[string]string{}
	for _, variant := range definition.Variants {
		name := placement.Names[variant.Name]
		if name == "" {
			name = definition.Name + naming.EnumVariantName(variant.Name)
		}
		if previous, exists := used[name]; exists {
			return "", fmt.Errorf("enum %s variants %q and %q both map to %s", definition.Name, previous, variant.Name, name)
		}
		used[name] = variant.Name
		fmt.Fprintf(&b, "\t%s %s%s = %d\n", name, qualifier, definition.Name, variant.Value)
	}
	b.WriteString(")\n")
	return b.String(), nil
}

type packetField struct {
	name     string
	hint     string // type-name hint: the exported wire name, never a layout name
	wireName string
	owner    string // type ID of the struct declaring the field; the packet name for top-level fields
	typ      string
	node     manifest.Node
}

func (g *generator) emitPacket(packet manifest.Packet, packetName string, constants []string) (string, error) {
	var b strings.Builder
	b.WriteString("// Code generated from canonical protocol manifest v2. DO NOT EDIT.\n\npackage packet\n\n")
	used, usedHints := map[string]bool{}, map[string]bool{}
	effective := g.packetFields(packet)
	fields := make([]packetField, 0, len(effective))
	for _, item := range effective {
		field := item.Field
		baseName, hintBase := g.fieldName(item.Owner, field.Name), exportName(field.Name)
		if g.emitPacketRuntime && baseName == "ID" {
			// ID is reserved by the generated packet runtime method. Keep the
			// wire field explicit without making the struct fail to compile.
			baseName = "IDValue"
		}
		if g.emitPacketRuntime && hintBase == "ID" {
			hintBase = "IDValue"
		}
		name, hint := uniqueFieldName(baseName, used), uniqueFieldName(hintBase, usedHints)
		encode := g.semantics.Apply(item.Owner, field)
		typ := mustGoType(g, encode, packetName+hint)
		fields = append(fields, packetField{name: name, hint: hint, wireName: field.Name, owner: item.Owner, typ: qualifyGoType(typ, g.definitions), node: encode})
	}
	imports := append([]string{g.protocolImportPath}, goImportsForFields(fields)...)
	writeGoImports(&b, imports)
	for _, block := range constants {
		b.WriteString(block)
		b.WriteString("\n")
	}
	for _, line := range docs.GoComments(g.docs.Type(packet.Name)) {
		b.WriteString(line)
		b.WriteByte('\n')
	}
	fmt.Fprintf(&b, "type %s struct {\n", packetName)
	for _, field := range fields {
		for _, line := range docs.GoComments(g.docs.Field(field.owner, field.wireName)) {
			fmt.Fprintf(&b, "\t%s\n", line)
		}
		fmt.Fprintf(&b, "\t%s %s\n", field.name, field.typ)
	}
	b.WriteString("}\n\n")
	if g.emitPacketRuntime {
		fmt.Fprintf(&b, "// ID ...\nfunc (*%s) ID() uint32 {\n\treturn ID%s\n}\n\n", packetName, packetName)
	}
	fmt.Fprintf(&b, "func (pk *%s) Marshal(io protocol.IO) {\n", packetName)
	emitter := marshalEmitter{g: g, qualifier: "protocol."}
	for _, field := range fields {
		if err := emitter.node(&b, field.node, "pk."+field.name, packetName+field.hint, "\t", addressStrategy{}); err != nil {
			return "", fmt.Errorf("packet %s field %s marshal: %w", packet.Name, field.name, err)
		}
	}
	b.WriteString("}\n")
	return formatGoSource(b.String())
}

func emitPacketIDs(packets []manifest.Packet, packetNames map[uint32]string) string {
	var b strings.Builder
	b.WriteString("// Code generated from canonical protocol manifest v2. DO NOT EDIT.\n\npackage packet\n\nconst (\n")
	for _, packet := range packets {
		fmt.Fprintf(&b, "\tID%s uint32 = %d\n", packetNames[packet.ID], packet.ID)
	}
	b.WriteString(")\n")
	return mustFormatGoSource(b.String())
}

func emitPacketRuntime(protocolImportPath string) string {
	return mustFormatGoSource(fmt.Sprintf(`// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"fmt"

	%q
)

// Packet is the common runtime contract for every generated Bedrock packet.
// Marshal reads from or writes to the supplied protocol IO implementation.
type Packet interface {
	ID() uint32
	Marshal(protocol.IO)
}

// Decode unmarshals one packet and rejects malformed or trailing input.
func Decode(data []byte, pk Packet) error {
	if pk == nil {
		return fmt.Errorf("decode packet <nil>")
	}
	reader := protocol.NewReader(data)
	pk.Marshal(reader)
	if err := reader.Err(); err != nil {
		return fmt.Errorf("decode packet %%T: %%w", pk, err)
	}
	if remaining := reader.Remaining(); remaining != 0 {
		return fmt.Errorf("decode packet %%T: %%d trailing bytes", pk, remaining)
	}
	return nil
}

// Encode marshals one packet and reports codec errors.
func Encode(pk Packet) ([]byte, error) {
	if pk == nil {
		return nil, fmt.Errorf("encode packet <nil>")
	}
	writer := protocol.NewWriter()
	pk.Marshal(writer)
	if err := writer.Err(); err != nil {
		return nil, fmt.Errorf("encode packet %%T: %%w", pk, err)
	}
	return writer.Data(), nil
}
`, protocolImportPath))
}

func emitPacketPools(packets []manifest.Packet, packetNames map[uint32]string) string {
	var b strings.Builder
	b.WriteString(`// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

// Factory constructs a fresh packet value for decoding.
type Factory func() Packet

// Pool maps protocol IDs to packet constructors.
type Pool map[uint32]Factory

`)
	emit := func(name string, include func(manifest.Direction) bool) {
		fmt.Fprintf(&b, "var %s = Pool{\n", name)
		for _, packet := range packets {
			if !include(packet.Direction) {
				continue
			}
			packetName := packetNames[packet.ID]
			fmt.Fprintf(&b, "\tID%s: func() Packet { return &%s{} },\n", packetName, packetName)
		}
		b.WriteString("}\n\n")
	}
	emit("allPacketFactories", func(manifest.Direction) bool { return true })
	emit("clientPacketFactories", func(direction manifest.Direction) bool {
		return direction == manifest.DirectionServerbound || direction == manifest.DirectionBidirectional
	})
	emit("serverPacketFactories", func(direction manifest.Direction) bool {
		return direction == manifest.DirectionClientbound || direction == manifest.DirectionBidirectional
	})
	b.WriteString(`// NewPool returns a copy of the complete packet factory table.
func NewPool() Pool { return clonePool(allPacketFactories) }

// NewClientPool returns factories for packets sent by a client, including
// packets explicitly marked bidirectional in the manifest.
func NewClientPool() Pool { return clonePool(clientPacketFactories) }

// NewServerPool returns factories for packets sent by a server, including
// packets explicitly marked bidirectional in the manifest.
func NewServerPool() Pool { return clonePool(serverPacketFactories) }

// NewPacket constructs the packet with id, if it is known.
func NewPacket(id uint32) (Packet, bool) { return newFromPool(allPacketFactories, id) }

// NewClientPacket constructs a known packet sent by a client.
func NewClientPacket(id uint32) (Packet, bool) { return newFromPool(clientPacketFactories, id) }

// NewServerPacket constructs a known packet sent by a server.
func NewServerPacket(id uint32) (Packet, bool) { return newFromPool(serverPacketFactories, id) }

func clonePool(source Pool) Pool {
	pool := make(Pool, len(source))
	for id, factory := range source { pool[id] = factory }
	return pool
}

func newFromPool(pool Pool, id uint32) (Packet, bool) {
	factory, ok := pool[id]
	if !ok { return nil, false }
	return factory(), true
}
`)
	return mustFormatGoSource(b.String())
}

func mustFormatGoSource(source string) string {
	formatted, err := formatGoSource(source)
	if err != nil {
		panic(err)
	}
	return formatted
}

type marshalEmitter struct {
	g         *generator
	qualifier string
	counter   int
}

func (e *marshalEmitter) mustGoType(node manifest.Node, hint string) string {
	typ := mustGoType(e.g, node, hint)
	if e.qualifier == "" {
		return typ
	}
	return qualifyGoType(typ, e.g.definitions)
}

func (e *marshalEmitter) runtime(name string) string {
	return e.qualifier + name
}

func (e *marshalEmitter) ioType() string {
	return e.qualifier + "IO"
}

func (e *marshalEmitter) temporary(prefix string) string {
	e.counter++
	return fmt.Sprintf("%s%d", prefix, e.counter)
}

type addressStrategy struct {
	pointer bool
}

func (a addressStrategy) address(expression string) string {
	if a.pointer {
		return expression
	}
	return "&" + expression
}

func (a addressStrategy) container(expression string) string {
	if a.pointer {
		return "*" + expression
	}
	return expression
}

func (a addressStrategy) element(expression, index string) string {
	if a.pointer {
		return "(*" + expression + ")[" + index + "]"
	}
	return expression + "[" + index + "]"
}

func (a addressStrategy) bits(expression string) string {
	if a.pointer {
		return "(*" + expression + ")[:]"
	}
	return expression + "[:]"
}

func (e *marshalEmitter) node(b *strings.Builder, node manifest.Node, expression, hint, indent string, address addressStrategy) error {
	if method, ok := e.semanticIOCall(node); ok {
		fmt.Fprintf(b, "%sio.%s(%s)\n", indent, method, address.address(expression))
		return e.numberConstraints(b, node, expression, indent, address)
	}
	if native, matched, err := e.g.nativeGoType(node); matched || err != nil {
		if err != nil {
			return err
		}
		fmt.Fprintf(b, "%sio.%s(%s)\n", indent, nativeMethod(native), address.address(expression))
		return e.numberConstraints(b, node, expression, indent, address)
	}
	switch node.Kind {
	case manifest.KindVoid:
		return nil
	case manifest.KindPrimitive:
		if node.Primitive == nil {
			return fmt.Errorf("primitive has no shape")
		}
		if node.Primitive.Code == "uuid" {
			fmt.Fprintf(b, "%sio.UUIDBytes(%s)\n", indent, address.address(expression))
			return nil
		}
		if node.Primitive.Code == "nbt_le" {
			encoding, err := nbtIOConstant(node, e.qualifier)
			if err != nil {
				return err
			}
			fmt.Fprintf(b, "%sio.NBT(%s, %s)\n", indent, address.address(expression), encoding)
			return nil
		}
		method, err := primitiveIOMethod(node.Primitive.Code)
		if err != nil {
			return err
		}
		fmt.Fprintf(b, "%sio.%s(%s)\n", indent, method, address.address(expression))
		return e.numberConstraints(b, node, expression, indent, address)
	case manifest.KindString:
		if !varuint32Prefix(node) {
			return fmt.Errorf("string has unsupported length prefix")
		}
		if node.Constraints != nil && (node.Constraints.MinLength != nil || node.Constraints.MaxLength != nil) {
			min, max := lengthBounds(node.Constraints.MinLength, node.Constraints.MaxLength)
			fmt.Fprintf(b, "%sio.StringLimits(%s, %d, %d)\n", indent, address.address(expression), min, max)
		} else {
			fmt.Fprintf(b, "%sio.String(%s)\n", indent, address.address(expression))
		}
		if node.Constraints != nil && node.Constraints.Pattern != "" {
			fmt.Fprintf(b, "%s%s(io, %s, %q)\n", indent, e.runtime("Pattern"), address.address(expression), node.Constraints.Pattern)
		}
		return nil
	case manifest.KindBytes:
		if !varuint32Prefix(node) {
			return fmt.Errorf("bytes have unsupported length prefix")
		}
		if node.Constraints != nil && (node.Constraints.MinLength != nil || node.Constraints.MaxLength != nil) {
			min, max := lengthBounds(node.Constraints.MinLength, node.Constraints.MaxLength)
			fmt.Fprintf(b, "%sio.ByteSliceLimits(%s, %d, %d)\n", indent, address.address(expression), min, max)
		} else {
			fmt.Fprintf(b, "%sio.ByteSlice(%s)\n", indent, address.address(expression))
		}
		return nil
	case manifest.KindBitset:
		fmt.Fprintf(b, "%sio.Bitset(%s, %d)\n", indent, address.bits(expression), node.Length)
		return nil
	case manifest.KindStruct:
		fmt.Fprintf(b, "%s%s.Marshal(io)\n", indent, expression)
		return nil
	case manifest.KindEnum:
		fmt.Fprintf(b, "%s%s.Marshal(io)\n", indent, expression)
		return e.numberConstraints(b, node, expression, indent, address)
	case manifest.KindRecursive:
		fmt.Fprintf(b, "%s%s(io, %s)\n", indent, e.runtime("Marshal"+e.g.identity[node.Target]), address.address(expression))
		return nil
	case manifest.KindOptional:
		if node.Value == nil {
			return fmt.Errorf("optional has no value")
		}
		value := *node.Value
		if value.Kind == manifest.KindOptional {
			if value.Value == nil {
				return fmt.Errorf("nested optional has no value")
			}
			return e.optionalCall(b, "DoubleOptionalFunc", *value.Value, expression, hint+"Value", indent, address)
		}
		return e.optionalCall(b, "OptionalFunc", value, expression, hint+"Value", indent, address)
	case manifest.KindArray:
		if node.Element == nil || node.Prefix == nil {
			return fmt.Errorf("array has no element or prefix")
		}
		if bytes, ok := manifest.ByteRun(node); ok && e.mustGoType(node, hint) == "[]uint8" {
			return e.node(b, bytes, expression, hint, indent, address)
		}
		return e.collection(b, node, expression, hint+"Item", indent, address)
	case manifest.KindFixedArray:
		if node.Element == nil {
			return fmt.Errorf("fixed array has no element")
		}
		index := e.temporary("index")
		fmt.Fprintf(b, "%sfor %s := range %s {\n", indent, index, address.container(expression))
		if err := e.node(b, *node.Element, address.element(expression, index), hint+"Item", indent+"\t", addressStrategy{}); err != nil {
			return err
		}
		fmt.Fprintf(b, "%s}\n", indent)
		return nil
	case manifest.KindMap:
		if node.Key == nil || node.Value == nil || node.Prefix == nil {
			return fmt.Errorf("map has no key, value, or prefix")
		}
		return e.mapEntries(b, node, expression, hint, indent, address)
	case manifest.KindUnion:
		name := mustGoType(e.g, node, hint)
		fmt.Fprintf(b, "%s%s(io, %s)\n", indent, e.runtime("Marshal"+name), address.address(expression))
		return nil
	case manifest.KindReserved, manifest.KindIgnored:
		return fmt.Errorf("%s nodes require explicit write/discard codec semantics", node.Kind)
	case manifest.KindSequence, manifest.KindConditional:
		return fmt.Errorf("%s nodes do not yet have a generated codec", node.Kind)
	case manifest.KindOpaque, manifest.KindUnresolved:
		return fmt.Errorf("%s node blocks codec generation: %s", node.Kind, node.Reason)
	default:
		return fmt.Errorf("unsupported node kind %q", node.Kind)
	}
}

func nativeMethod(native string) string {
	switch native {
	case "uuid.UUID":
		return "UUID"
	case "mgl32.Vec2":
		return "Vec2"
	case "mgl32.Vec3":
		return "Vec3"
	case "color.RGBA":
		return "RGBA"
	default:
		panic("native type " + native + " has no codec operation")
	}
}

func (e *marshalEmitter) optionalCall(b *strings.Builder, helper string, value manifest.Node, expression, hint, indent string, address addressStrategy) error {
	if helper == "OptionalFunc" && e.marshalable(value, hint) {
		fmt.Fprintf(b, "%s%s(io, %s)\n", indent, e.runtime("OptionalMarshaler"), address.address(expression))
		return nil
	}
	fmt.Fprintf(b, "%s%s(io, %s, ", indent, e.runtime(helper), address.address(expression))
	if method, ok := e.directIOCall(value); ok {
		fmt.Fprintf(b, "io.%s)\n", method)
		return nil
	}
	typ := e.mustGoType(value, hint)
	fmt.Fprintf(b, "func(value *%s) {\n", typ)
	if err := e.node(b, value, "value", hint, indent+"\t", addressStrategy{pointer: true}); err != nil {
		return err
	}
	fmt.Fprintf(b, "%s})\n", indent)
	return nil
}

func (e *marshalEmitter) collection(b *strings.Builder, node manifest.Node, expression, hint, indent string, address addressStrategy) error {
	prefix, element := *node.Prefix, *node.Element
	if prefix.Kind != manifest.KindPrimitive || prefix.Primitive == nil {
		return fmt.Errorf("collection prefix must be a primitive")
	}
	countMethod, err := primitiveIOMethod(prefix.Primitive.Code)
	if err != nil {
		return err
	}
	min, max, constrained := itemBounds(node.Constraints)
	if prefix.Primitive.Code == "var_u32" && e.marshalable(element, hint) {
		helper := "Slice"
		if constrained {
			helper = "SliceLimits"
		}
		fmt.Fprintf(b, "%s%s(io, %s", indent, e.runtime(helper), address.address(expression))
		if constrained {
			fmt.Fprintf(b, ", %d, %d", min, max)
		}
		b.WriteString(")\n")
		return nil
	}
	helper := "FuncSlice"
	if constrained {
		helper = "FuncSliceLimits"
	}
	fmt.Fprintf(b, "%s%s(io, %s, io.%s, ", indent, e.runtime(helper), address.address(expression), countMethod)
	if constrained {
		fmt.Fprintf(b, "%d, %d, ", min, max)
	}
	if method, ok := e.directIOCall(element); ok {
		fmt.Fprintf(b, "io.%s)\n", method)
		return nil
	}
	typ := e.mustGoType(element, hint)
	fmt.Fprintf(b, "func(value *%s) {\n", typ)
	if err := e.node(b, element, "value", hint, indent+"\t", addressStrategy{pointer: true}); err != nil {
		return err
	}
	fmt.Fprintf(b, "%s})\n", indent)
	return nil
}

// marshalable reports whether a node is emitted as a generated struct or enum
// whose pointer marshals itself without further per-site constraints.
func (e *marshalEmitter) marshalable(node manifest.Node, hint string) bool {
	if node.Constraints != nil {
		return false
	}
	if node.Kind != manifest.KindStruct && node.Kind != manifest.KindEnum && node.Kind != manifest.KindRecursive {
		return false
	}
	if _, native, _ := e.g.nativeGoType(node); native {
		return false
	}
	typ := mustGoType(e.g, node, hint)
	definition, ok := e.g.definitions[typ]
	return ok && (definition.Kind == manifest.KindStruct || definition.Kind == manifest.KindEnum)
}

func (e *marshalEmitter) mapEntries(b *strings.Builder, node manifest.Node, expression, hint, indent string, address addressStrategy) error {
	if node.Prefix.Kind != manifest.KindPrimitive || node.Prefix.Primitive == nil {
		return fmt.Errorf("map prefix must be a primitive")
	}
	countMethod, err := primitiveIOMethod(node.Prefix.Primitive.Code)
	if err != nil {
		return err
	}
	min, max, constrained := propertyBounds(node.Constraints)
	helper := "OrderedMap"
	if constrained {
		helper = "OrderedMapLimits"
	}
	fmt.Fprintf(b, "%s%s(io, %s, io.%s, ", indent, e.runtime(helper), address.address(expression), countMethod)
	if constrained {
		fmt.Fprintf(b, "%d, %d, ", min, max)
	}
	if method, ok := e.directIOCall(*node.Key); ok {
		fmt.Fprintf(b, "io.%s, ", method)
	} else {
		typ := e.mustGoType(*node.Key, hint+"Key")
		fmt.Fprintf(b, "func(value *%s) {\n", typ)
		if err := e.node(b, *node.Key, "value", hint+"Key", indent+"\t", addressStrategy{pointer: true}); err != nil {
			return err
		}
		fmt.Fprintf(b, "%s}, ", indent)
	}
	if method, ok := e.directIOCall(*node.Value); ok {
		fmt.Fprintf(b, "io.%s)\n", method)
		return nil
	}
	typ := e.mustGoType(*node.Value, hint+"Value")
	fmt.Fprintf(b, "func(value *%s) {\n", typ)
	if err := e.node(b, *node.Value, "value", hint+"Value", indent+"\t", addressStrategy{pointer: true}); err != nil {
		return err
	}
	fmt.Fprintf(b, "%s})\n", indent)
	return nil
}

// numberConstraints emits schema bounds that the wire type cannot already
// guarantee, so an unsigned field never checks a minimum of zero.
func (e *marshalEmitter) numberConstraints(b *strings.Builder, node manifest.Node, expression, indent string, address addressStrategy) error {
	if node.Constraints == nil {
		return nil
	}
	low, high, bounded := wireRange(node)
	if minimum := node.Constraints.Minimum; minimum != nil && !(bounded && *minimum <= low) {
		fmt.Fprintf(b, "%s%s(io, %s, %s)\n", indent, e.runtime("Minimum"), address.address(expression), strconv.FormatFloat(*minimum, 'g', -1, 64))
	}
	if maximum := node.Constraints.Maximum; maximum != nil && !(bounded && *maximum >= high) {
		fmt.Fprintf(b, "%s%s(io, %s, %s)\n", indent, e.runtime("Maximum"), address.address(expression), strconv.FormatFloat(*maximum, 'g', -1, 64))
	}
	return nil
}

// wireRange reports the value range an integer wire encoding can carry.
func wireRange(node manifest.Node) (low, high float64, ok bool) {
	code := ""
	switch {
	case node.Primitive != nil:
		code = node.Primitive.Code
	case node.Kind == manifest.KindStruct:
		code, _ = primitiveStructCode(node)
	}
	typ, err := primitiveGoType(code)
	if err != nil {
		return 0, 0, false
	}
	switch typ {
	case "int8":
		return math.MinInt8, math.MaxInt8, true
	case "int16":
		return math.MinInt16, math.MaxInt16, true
	case "int32":
		return math.MinInt32, math.MaxInt32, true
	case "int64":
		return math.MinInt64, math.MaxInt64, true
	case "uint8":
		return 0, math.MaxUint8, true
	case "uint16":
		return 0, math.MaxUint16, true
	case "uint32":
		return 0, math.MaxUint32, true
	case "uint64":
		return 0, math.MaxUint64, true
	default:
		return 0, 0, false
	}
}

func lengthBounds(minimum, maximum *uint64) (uint64, uint64) {
	min, max := uint64(0), ^uint64(0)
	if minimum != nil {
		min = *minimum
	}
	if maximum != nil {
		max = *maximum
	}
	return min, max
}

func itemBounds(constraints *manifest.Constraints) (uint64, uint64, bool) {
	if constraints == nil || constraints.MinItems == nil && constraints.MaxItems == nil {
		return 0, 0, false
	}
	min, max := lengthBounds(constraints.MinItems, constraints.MaxItems)
	return min, max, true
}

func propertyBounds(constraints *manifest.Constraints) (uint64, uint64, bool) {
	if constraints == nil || constraints.MinProperties == nil && constraints.MaxProperties == nil {
		return 0, 0, false
	}
	min, max := lengthBounds(constraints.MinProperties, constraints.MaxProperties)
	return min, max, true
}

func (e *marshalEmitter) directIOCall(node manifest.Node) (string, bool) {
	if node.Constraints != nil {
		return "", false
	}
	if method, ok := e.semanticIOCall(node); ok {
		return method, true
	}
	if native, matched, err := e.g.nativeGoType(node); matched && err == nil {
		switch native {
		case "uuid.UUID":
			return "UUID", true
		case "mgl32.Vec2":
			return "Vec2", true
		case "mgl32.Vec3":
			return "Vec3", true
		case "color.RGBA":
			return "RGBA", true
		}
	}
	switch node.Kind {
	case manifest.KindPrimitive:
		if node.Primitive == nil {
			return "", false
		}
		if node.Primitive.Code == "uuid" {
			return "UUIDBytes", true
		}
		method, err := primitiveIOMethod(node.Primitive.Code)
		return method, err == nil
	case manifest.KindString:
		if varuint32Prefix(node) {
			return "String", true
		}
	case manifest.KindBytes:
		if varuint32Prefix(node) {
			return "ByteSlice", true
		}
	}
	return "", false
}

func nbtIOConstant(node manifest.Node, qualifier string) (string, error) {
	switch manifest.NBTEncoding(node.Encoding) {
	case manifest.NBTNetwork:
		return qualifier + "NBTNetwork", nil
	case manifest.NBTPersistent:
		return qualifier + "NBTPersistent", nil
	default:
		return "", fmt.Errorf("NBT node has invalid encoding %q", node.Encoding)
	}
}

func (e *marshalEmitter) semanticIOCall(node manifest.Node) (string, bool) {
	native, matched, err := e.g.nativeGoType(node)
	if err != nil || !matched {
		return "", false
	}
	code, hasCode := primitiveStructCode(node)
	switch node.TypeID {
	case "ActorRuntimeID":
		switch {
		case native == "uint64" && hasCode && code == "var_u64":
			return "ActorRuntimeID", true
		case native == "int64" && hasCode && code == "zigzag_i64":
			return "ActorRuntimeIDVarint64", true
		case native == "uint32" && hasCode && code == "var_u32":
			return "ActorRuntimeIDVaruint32", true
		case native == "int64" && hasCode && code == "var_i64":
			return "SignedVarint64", true
		}
	case "ActorUniqueID":
		switch {
		case native == "int64" && hasCode && code == "zigzag_i64":
			return "ActorUniqueID", true
		case native == "int64" && hasCode && code == "i64le":
			return "ActorUniqueIDInt64", true
		case native == "uint64" && hasCode && code == "u64le":
			return "ActorUniqueIDUint64", true
		case native == "uint64" && hasCode && code == "var_u64":
			return "ActorUniqueIDVaruint64", true
		case native == "int64" && hasCode && code == "var_i64":
			return "SignedVarint64", true
		}
	case "PlayerInputTick":
		if native == "uint64" {
			return "PlayerInputTick", true
		}
	default:
		return "", false
	}
	return "", false
}

// semanticIOCall retains the package-local helper used by emitter tests and
// older integrations; generated code uses the option-aware method above.
func semanticIOCall(node manifest.Node) (string, bool) {
	return (&marshalEmitter{g: &generator{nativeTypes: true}}).semanticIOCall(node)
}

func (e *marshalEmitter) union(b *strings.Builder, definition typeDefinition) error {
	fmt.Fprintf(b, "// Marshal%s reads or writes the %s union using its canonical wire layout.\n", definition.Name, definition.Name)
	fmt.Fprintf(b, "func Marshal%s(io %s, x *%s) {\n", definition.Name, e.ioType(), definition.Name)
	if len(definition.Union) == 0 {
		b.WriteString("\tio.InvalidValue(nil, \"union has no variants\")\n}\n\n")
		return nil
	}
	tagMethod, err := primitiveIOMethod(definition.Primitive)
	if err != nil {
		return err
	}
	fmt.Fprintf(b, "\t%s(io, x, io.%s, %s.tag%s, func(tag %s) %s {\n\t\tswitch tag {\n", e.runtime("Union"), tagMethod, definition.Name, definition.Name, definition.Underlying, definition.Name)
	for _, member := range definition.Union {
		fmt.Fprintf(b, "\t\tcase %d:\n\t\t\treturn new(%s)\n", member.Value, member.Name)
	}
	b.WriteString("\t\t}\n\t\treturn nil\n\t})\n}\n\n")
	return nil
}

func mustGoType(g *generator, node manifest.Node, hint string) string {
	typ, err := g.goType(node, hint)
	if err != nil {
		panic(err)
	}
	return typ
}

func varuint32Prefix(node manifest.Node) bool {
	return node.Prefix != nil && node.Prefix.Kind == manifest.KindPrimitive && node.Prefix.Primitive != nil && node.Prefix.Primitive.Code == "var_u32"
}

func primitiveIOMethod(code string) (string, error) {
	switch code {
	case "bool":
		return "Bool", nil
	case "i8":
		return "Int8", nil
	case "u8":
		return "Uint8", nil
	case "i16le":
		return "Int16", nil
	case "u16le":
		return "Uint16", nil
	case "i16be":
		return "BEInt16", nil
	case "u16be":
		return "BEUint16", nil
	case "i32le":
		return "Int32", nil
	case "u32le":
		return "Uint32", nil
	case "i32be":
		return "BEInt32", nil
	case "u32be":
		return "BEUint32", nil
	case "i64le":
		return "Int64", nil
	case "u64le":
		return "Uint64", nil
	case "i64be":
		return "BEInt64", nil
	case "u64be":
		return "BEUint64", nil
	case "f32le":
		return "Float32", nil
	case "f64le":
		return "Float64", nil
	case "f32be":
		return "BEFloat32", nil
	case "f64be":
		return "BEFloat64", nil
	case "zigzag_i32":
		return "Varint32", nil
	case "zigzag_i64":
		return "Varint64", nil
	case "var_u32":
		return "Varuint32", nil
	case "var_u64":
		return "Varuint64", nil
	case "var_i32":
		return "SignedVarint32", nil
	case "var_i64":
		return "SignedVarint64", nil
	default:
		return "", fmt.Errorf("primitive %q has no IO method", code)
	}
}

func goImportsForFields(fields []packetField) []string {
	types := make([]string, 0, len(fields))
	for _, field := range fields {
		types = append(types, field.typ)
	}
	return goImportsForTypes(types)
}

func qualifyGoType(typ string, definitions map[string]typeDefinition) string {
	var b strings.Builder
	for index := 0; index < len(typ); {
		if !isGoIdentifierByte(typ[index]) {
			b.WriteByte(typ[index])
			index++
			continue
		}
		end := index + 1
		for end < len(typ) && isGoIdentifierByte(typ[end]) {
			end++
		}
		name := typ[index:end]
		_, generated := definitions[name]
		if (generated || name == "Optional" || name == "OrderedEntry") && (index == 0 || typ[index-1] != '.') {
			b.WriteString("protocol.")
		}
		b.WriteString(name)
		index = end
	}
	return b.String()
}

func isGoIdentifierByte(value byte) bool {
	return value == '_' || value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z' || value >= '0' && value <= '9'
}

func goImportsForTypes(types []string) []string {
	used := map[string]bool{}
	for _, typ := range types {
		if strings.Contains(typ, "color.") {
			used["image/color"] = true
		}
		if strings.Contains(typ, "mgl32.") {
			used["github.com/go-gl/mathgl/mgl32"] = true
		}
		if strings.Contains(typ, "uuid.") {
			used["github.com/google/uuid"] = true
		}
	}
	imports := make([]string, 0, len(used))
	for path := range used {
		imports = append(imports, path)
	}
	sort.Strings(imports)
	return imports
}

func writeGoImports(b *strings.Builder, imports []string) {
	if len(imports) == 0 {
		return
	}
	sort.Strings(imports)
	var standard, external []string
	for _, path := range imports {
		first := strings.SplitN(path, "/", 2)[0]
		if strings.Contains(first, ".") {
			external = append(external, path)
		} else {
			standard = append(standard, path)
		}
	}
	b.WriteString("import (\n")
	for _, path := range standard {
		fmt.Fprintf(b, "\t%q\n", path)
	}
	if len(standard) != 0 && len(external) != 0 {
		b.WriteByte('\n')
	}
	for _, path := range external {
		fmt.Fprintf(b, "\t%q\n", path)
	}
	b.WriteString(")\n\n")
}

func formatGoSource(source string) (string, error) {
	formatted, err := format.Source([]byte(source))
	if err != nil {
		return source, fmt.Errorf("format generated Go: %w", err)
	}
	return string(formatted), nil
}

func shortTypeName(name string) string {
	if position := strings.LastIndex(name, "::"); position >= 0 {
		return name[position+2:]
	}
	return name
}

func primitiveGoType(code string) (string, error) {
	switch code {
	case "bool":
		return "bool", nil
	case "uuid":
		return "[16]byte", nil
	case "i8":
		return "int8", nil
	case "u8":
		return "uint8", nil
	case "i16le", "i16be":
		return "int16", nil
	case "u16le", "u16be":
		return "uint16", nil
	case "i32le", "i32be", "var_i32", "zigzag_i32":
		return "int32", nil
	case "u32le", "u32be", "var_u32":
		return "uint32", nil
	case "i64le", "i64be", "var_i64", "zigzag_i64":
		return "int64", nil
	case "u64le", "u64be", "var_u64":
		return "uint64", nil
	case "f32le", "f32be":
		return "float32", nil
	case "f64le", "f64be":
		return "float64", nil
	case "nbt_le":
		return "[]byte", nil
	default:
		return "", fmt.Errorf("unsupported primitive code %q", code)
	}
}

func publicTypeName(value string) string {
	return strings.ReplaceAll(exportName(naming.PublicTypeName(value)), "Molang", "MoLang")
}

func snakeName(value string) string {
	var b strings.Builder
	runes := []rune(value)
	for index, r := range runes {
		if unicode.IsUpper(r) && index > 0 {
			previousLower := unicode.IsLower(runes[index-1]) || unicode.IsDigit(runes[index-1])
			nextLower := index+1 < len(runes) && unicode.IsLower(runes[index+1])
			if previousLower || nextLower {
				b.WriteByte('_')
			}
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

func uniqueFileName(base string, packetID uint32, used map[string]bool) string {
	if !used[base] {
		used[base] = true
		return base
	}
	name := strings.TrimSuffix(base, ".go") + fmt.Sprintf("_%d.go", packetID)
	used[name] = true
	return name
}

func exportName(value string) string { return naming.GoExportName(value) }

func normalizeGoInitialisms(value string) string { return naming.NormalizeGoInitialisms(value) }

func uniqueFieldName(base string, used map[string]bool) string {
	if base == "" {
		base = "Field"
	}
	if !used[base] {
		used[base] = true
		return base
	}
	for index := 2; ; index++ {
		candidate := fmt.Sprintf("%s%d", base, index)
		if !used[candidate] {
			used[candidate] = true
			return candidate
		}
	}
}
