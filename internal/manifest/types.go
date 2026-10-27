// Package manifest contains the profile-neutral canonical Bedrock wire
// vocabulary. Emitters are deliberately downstream of this package: they may
// choose names and APIs, but they cannot invent or recover wire shape.
package manifest

import (
	"encoding/json"
	"fmt"
	"sort"
)

const SchemaVersion uint32 = 2

type Direction string

const (
	DirectionClientbound   Direction = "clientbound"
	DirectionServerbound   Direction = "serverbound"
	DirectionBidirectional Direction = "bidirectional"
	DirectionUnknown       Direction = "unknown"
)

type Symmetry string

const (
	Symmetric  Symmetry = "symmetric"
	Asymmetric Symmetry = "asymmetric"
)

type NBTEncoding string

const (
	NBTNetwork    NBTEncoding = "network"
	NBTPersistent NBTEncoding = "persistent"
	// NBTNetworkCompound omits the root tag ID and name, retaining the compound end tag.
	NBTNetworkCompound NBTEncoding = "network_compound"
)

type NodeKind string

const (
	KindPrimitive   NodeKind = "primitive"
	KindString      NodeKind = "string"
	KindBytes       NodeKind = "bytes"
	KindBitset      NodeKind = "bitset"
	KindArray       NodeKind = "array"
	KindFixedArray  NodeKind = "fixed_array"
	KindSequence    NodeKind = "sequence"
	KindOptional    NodeKind = "optional"
	KindStruct      NodeKind = "struct"
	KindMap         NodeKind = "map"
	KindUnion       NodeKind = "union"
	KindConditional NodeKind = "conditional"
	KindEnum        NodeKind = "enum"
	KindReserved    NodeKind = "reserved"
	KindIgnored     NodeKind = "ignored"
	KindRecursive   NodeKind = "recursive"
	KindOpaque      NodeKind = "opaque"
	KindUnresolved  NodeKind = "unresolved"
	KindVoid        NodeKind = "void"
)

// Manifest is the only wire-shape input accepted by v2 emitters.
type Manifest struct {
	SchemaVersion uint32           `json:"schema_version"`
	Target        Target           `json:"target"`
	Sources       []SourcePin      `json:"sources"`
	Packets       []Packet         `json:"packets"`
	Adjudications []Adjudication   `json:"adjudications,omitempty"`
	Overrides     []OverrideProof  `json:"overrides,omitempty"`
	Derivation    *DerivationProof `json:"derivation,omitempty"`
}

type Target struct {
	MinecraftVersion string `json:"minecraft_version"`
	ProtocolVersion  int    `json:"protocol_version"`
	Build            string `json:"build,omitempty"`
}

// SourcePin is copied from the source lock. It is intentionally a pin, not a
// confidence score: an uncontested canonical node inherits only these stable
// identifiers and does not carry verbose evidence.
type SourcePin struct {
	ID               string `json:"id"`
	Kind             string `json:"kind"`
	Revision         string `json:"revision"`
	Digest           string `json:"digest"`
	Locator          string `json:"locator,omitempty"`
	MinecraftVersion string `json:"minecraft_version,omitempty"`
	ProtocolVersion  int    `json:"protocol_version,omitempty"`
}

type Packet struct {
	ID        uint32    `json:"id"`
	Name      string    `json:"name"`
	Direction Direction `json:"direction"`
	Fields    []Field   `json:"fields"`
}

type Field struct {
	Ordinal       int        `json:"ordinal"`
	Name          string     `json:"name"`
	Semantic      string     `json:"semantic,omitempty"`
	TypeID        string     `json:"type_id,omitempty"`
	Encode        Node       `json:"encode"`
	Decode        *Node      `json:"decode,omitempty"`
	Symmetry      Symmetry   `json:"symmetry"`
	Reserved      bool       `json:"reserved,omitempty"`
	Ignored       bool       `json:"ignored,omitempty"`
	Compatibility []string   `json:"compatibility,omitempty"`
	Provenance    Provenance `json:"provenance"`
}

// Node is a recursive wire-shape node. The fields are intentionally explicit
// rather than a free-form map so validation can reject missing lengths,
// invented primitive encodings, and unresolved reachable shapes.
type Node struct {
	Kind           NodeKind        `json:"kind"`
	Semantic       string          `json:"semantic,omitempty"`
	TypeID         string          `json:"type_id,omitempty"`
	Primitive      *PrimitiveShape `json:"primitive,omitempty"`
	Prefix         *Node           `json:"prefix,omitempty"`
	Encoding       string          `json:"encoding,omitempty"`
	Representation string          `json:"representation,omitempty"`
	Element        *Node           `json:"element,omitempty"`
	Length         uint64          `json:"length,omitempty"`
	Value          *Node           `json:"value,omitempty"`
	Elements       []Node          `json:"elements,omitempty"`
	Fields         []Field         `json:"fields,omitempty"`
	Key            *Node           `json:"key,omitempty"`
	Variants       []Variant       `json:"variants,omitempty"`
	Control        *Node           `json:"control,omitempty"`
	CompareTo      string          `json:"compare_to,omitempty"`
	Cases          []Case          `json:"cases,omitempty"`
	Default        *Node           `json:"default,omitempty"`
	Target         string          `json:"target,omitempty"`
	Reason         string          `json:"reason,omitempty"`
	Reachable      bool            `json:"reachable,omitempty"`
	Constraints    *Constraints    `json:"constraints,omitempty"`
}

// Constraints retains validation bounds published by a protocol source. They
// do not change the bytes on the wire, but generated readers and writers must
// reject values outside them.
type Constraints struct {
	MinLength     *uint64  `json:"min_length,omitempty"`
	MaxLength     *uint64  `json:"max_length,omitempty"`
	MinItems      *uint64  `json:"min_items,omitempty"`
	MaxItems      *uint64  `json:"max_items,omitempty"`
	MinProperties *uint64  `json:"min_properties,omitempty"`
	MaxProperties *uint64  `json:"max_properties,omitempty"`
	Minimum       *float64 `json:"minimum,omitempty"`
	Maximum       *float64 `json:"maximum,omitempty"`
	Pattern       string   `json:"pattern,omitempty"`
}

type PrimitiveShape struct {
	Code       string `json:"code"`
	Width      uint8  `json:"width"`
	Signed     bool   `json:"signed"`
	ZigZag     bool   `json:"zigzag"`
	Endianness string `json:"endianness"`
}

type Variant struct {
	Value  int64  `json:"value"`
	Name   string `json:"name"`
	Encode Node   `json:"encode"`
	Decode *Node  `json:"decode,omitempty"`
}

type Case struct {
	Value  string `json:"value"`
	Encode []Node `json:"encode"`
	Decode []Node `json:"decode,omitempty"`
}

// EnumValue names a wire value. Distinct names may alias the same numeric value;
// the first declared name is canonical when decoding to a named variant.
type EnumValue struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

// Provenance carries only source pins for agreement. Evidence is populated for
// disagreements and corrections, where a maintainer needs an audit trail.
type Provenance struct {
	Pins     []string   `json:"pins"`
	Evidence []Evidence `json:"evidence,omitempty"`
}

type Evidence struct {
	SourceID         string `json:"source_id"`
	Locator          string `json:"locator"`
	ClaimFingerprint string `json:"claim_fingerprint,omitempty"`
	Summary          string `json:"summary,omitempty"`
}

type ClaimFingerprint struct {
	SourceID string `json:"source_id"`
	Digest   string `json:"digest"`
}

type Adjudication struct {
	ID                    string             `json:"id"`
	Target                string             `json:"target"`
	PrePatchContextSHA256 string             `json:"pre_patch_context_sha256"`
	Claims                []ClaimFingerprint `json:"claims"`
	SelectedSource        string             `json:"selected_source"`
	Evidence              []Evidence         `json:"evidence"`
	Reason                string             `json:"reason"`
}

type OverrideProof struct {
	ID                    string     `json:"id"`
	Target                string     `json:"target"`
	PrePatchNodeSHA256    string     `json:"pre_patch_node_sha256"`
	PrePatchContextSHA256 string     `json:"pre_patch_context_sha256"`
	PostPatchNodeSHA256   string     `json:"post_patch_node_sha256"`
	Evidence              []Evidence `json:"evidence"`
	Reason                string     `json:"reason"`
}

// DerivationProof records a same-protocol hotfix derived from an already
// reconciled manifest when no second complete source snapshot exists.
type DerivationProof struct {
	BaseTarget         Target                     `json:"base_target"`
	BaseManifestSHA256 string                     `json:"base_manifest_sha256"`
	Operations         []DerivationOperationProof `json:"operations"`
}

type DerivationOperationProof struct {
	ID                  string     `json:"id"`
	PacketID            uint32     `json:"packet_id"`
	FieldOrdinal        int        `json:"field_ordinal"`
	Path                string     `json:"path"`
	Operation           string     `json:"operation"`
	PrePatchNodeSHA256  string     `json:"pre_patch_node_sha256"`
	PostPatchNodeSHA256 string     `json:"post_patch_node_sha256"`
	Evidence            []Evidence `json:"evidence"`
	Reason              string     `json:"reason"`
}

func Primitive(code string) Node {
	if code == "bytes" {
		return Bytes(Primitive("var_u32"))
	}
	shape, err := PrimitiveForCode(code)
	if err != nil {
		return Node{Kind: KindUnresolved, Reason: err.Error(), Reachable: true}
	}
	return Node{Kind: KindPrimitive, Primitive: &shape}
}

// NBT returns an NBT node with its wire encoding explicitly recorded.
func NBT(encoding NBTEncoding) Node {
	node := Primitive("nbt_le")
	node.Encoding = string(encoding)
	return node
}

func IsNBT(node Node) bool {
	return node.Kind == KindPrimitive && node.Primitive != nil && node.Primitive.Code == "nbt_le"
}

func ValidNBTEncoding(value string) bool {
	switch NBTEncoding(value) {
	case NBTNetwork, NBTPersistent, NBTNetworkCompound:
		return true
	default:
		return false
	}
}

func PrimitiveForCode(code string) (PrimitiveShape, error) {
	shape, ok := primitiveTable[code]
	if !ok {
		return PrimitiveShape{}, fmt.Errorf("unknown primitive code %q", code)
	}
	return shape, nil
}

func String(prefix Node) Node {
	return Node{Kind: KindString, Prefix: &prefix, Encoding: "utf8", Representation: "text"}
}

func Bytes(prefix Node) Node {
	return Node{Kind: KindBytes, Prefix: &prefix, Representation: "bytes"}
}

func Bitset(bits uint64) Node { return Node{Kind: KindBitset, Length: bits, Representation: "bitset"} }

func Array(prefix, element Node) Node {
	return Node{Kind: KindArray, Prefix: &prefix, Element: &element}
}

// ByteRun returns the Bytes node wire-identical to n when n is a
// var_u32-counted array of plain u8, so emitters can copy it in bulk.
func ByteRun(n Node) (Node, bool) {
	if n.Kind != KindArray || !isPrimitive(n.Prefix, "var_u32") || !isPrimitive(n.Element, "u8") {
		return Node{}, false
	}
	if element := n.Element; element.Constraints != nil || element.Semantic != "" || element.TypeID != "" {
		return Node{}, false
	}
	bytes := Bytes(*n.Prefix)
	bytes.Semantic, bytes.TypeID = n.Semantic, n.TypeID
	if c := n.Constraints; c != nil {
		if c.MinLength != nil || c.MaxLength != nil || c.MinProperties != nil || c.MaxProperties != nil ||
			c.Minimum != nil || c.Maximum != nil || c.Pattern != "" {
			return Node{}, false
		}
		if c.MinItems != nil || c.MaxItems != nil {
			bytes.Constraints = &Constraints{MinLength: c.MinItems, MaxLength: c.MaxItems}
		}
	}
	return bytes, true
}

func isPrimitive(n *Node, code string) bool {
	return n != nil && n.Kind == KindPrimitive && n.Primitive != nil && n.Primitive.Code == code
}

func FixedArray(length uint64, element Node) Node {
	return Node{Kind: KindFixedArray, Length: length, Element: &element}
}

// Sequence preserves an ordered list of wire nodes without pretending that
// the nodes form a named struct. It is used for inline groups and conditional
// branches whose field identity is still carried by each child node.
func Sequence(elements ...Node) Node { return Node{Kind: KindSequence, Elements: elements} }

func Optional(value Node) Node { return Node{Kind: KindOptional, Value: &value} }

func Struct(fields ...Field) Node { return Node{Kind: KindStruct, Fields: fields} }

func Map(prefix, key, value Node) Node {
	return Node{Kind: KindMap, Prefix: &prefix, Key: &key, Value: &value, Representation: "ordered_entries"}
}

func Union(control Node, variants ...Variant) Node {
	return Node{Kind: KindUnion, Control: &control, Variants: variants}
}

func Enum(underlying string, values ...EnumValue) Node {
	node := Node{Kind: KindEnum, Primitive: &PrimitiveShape{}}
	shape, err := PrimitiveForCode(underlying)
	if err != nil {
		return Node{Kind: KindUnresolved, Reason: err.Error(), Reachable: true}
	}
	node.Primitive = &shape
	for _, value := range values {
		node.Variants = append(node.Variants, Variant{Value: value.Value, Name: value.Name, Encode: Void()})
	}
	return node
}

// MergeEnumVariants returns known plus the value/name pairs of more it lacks,
// sorted by value once any are added; sites of one enum type may each list a subset.
// A name already bound to a different value is an error; aliases of one value are kept.
func MergeEnumVariants(enum string, known, more []Variant) ([]Variant, error) {
	values := make(map[string]int64, len(known))
	for _, variant := range known {
		if _, ok := values[variant.Name]; !ok {
			values[variant.Name] = variant.Value
		}
	}
	merged := append([]Variant(nil), known...)
	for _, variant := range more {
		value, ok := values[variant.Name]
		if ok && value != variant.Value {
			return nil, fmt.Errorf("enum %s names %q both %d and %d", enum, variant.Name, value, variant.Value)
		}
		if !ok {
			values[variant.Name] = variant.Value
			merged = append(merged, variant)
		}
	}
	if len(merged) > len(known) {
		sort.SliceStable(merged, func(i, j int) bool { return merged[i].Value < merged[j].Value })
	}
	return merged, nil
}

func Reserved(value Node) Node { return Node{Kind: KindReserved, Element: &value} }

func Ignored(value Node) Node { return Node{Kind: KindIgnored, Element: &value} }

func Recursive(typeID string) Node { return Node{Kind: KindRecursive, TypeID: typeID, Target: typeID} }

func Opaque(name, reason string) Node {
	return Node{Kind: KindOpaque, Semantic: name, Reason: reason, Reachable: true}
}

func Unresolved(reason string, reachable bool) Node {
	return Node{Kind: KindUnresolved, Reason: reason, Reachable: reachable}
}

func Void() Node { return Node{Kind: KindVoid} }

func (n Node) MarshalJSON() ([]byte, error) {
	type alias Node
	return json.Marshal(alias(n))
}

var primitiveTable = map[string]PrimitiveShape{
	"bool":       {Code: "bool", Width: 1, Endianness: "none"},
	"i8":         {Code: "i8", Width: 8, Signed: true, Endianness: "none"},
	"u8":         {Code: "u8", Width: 8, Endianness: "none"},
	"i16le":      {Code: "i16le", Width: 16, Signed: true, Endianness: "little"},
	"u16le":      {Code: "u16le", Width: 16, Endianness: "little"},
	"i16be":      {Code: "i16be", Width: 16, Signed: true, Endianness: "big"},
	"u16be":      {Code: "u16be", Width: 16, Endianness: "big"},
	"i32le":      {Code: "i32le", Width: 32, Signed: true, Endianness: "little"},
	"u32le":      {Code: "u32le", Width: 32, Endianness: "little"},
	"i32be":      {Code: "i32be", Width: 32, Signed: true, Endianness: "big"},
	"u32be":      {Code: "u32be", Width: 32, Endianness: "big"},
	"i64le":      {Code: "i64le", Width: 64, Signed: true, Endianness: "little"},
	"u64le":      {Code: "u64le", Width: 64, Endianness: "little"},
	"i64be":      {Code: "i64be", Width: 64, Signed: true, Endianness: "big"},
	"u64be":      {Code: "u64be", Width: 64, Endianness: "big"},
	"f32le":      {Code: "f32le", Width: 32, Endianness: "little"},
	"f32be":      {Code: "f32be", Width: 32, Endianness: "big"},
	"f64le":      {Code: "f64le", Width: 64, Endianness: "little"},
	"f64be":      {Code: "f64be", Width: 64, Endianness: "big"},
	"var_i32":    {Code: "var_i32", Width: 32, Signed: true, Endianness: "none"},
	"var_u32":    {Code: "var_u32", Width: 32, Endianness: "none"},
	"var_i64":    {Code: "var_i64", Width: 64, Signed: true, Endianness: "none"},
	"var_u64":    {Code: "var_u64", Width: 64, Endianness: "none"},
	"zigzag_i32": {Code: "zigzag_i32", Width: 32, Signed: true, ZigZag: true, Endianness: "none"},
	"zigzag_i64": {Code: "zigzag_i64", Width: 64, Signed: true, ZigZag: true, Endianness: "none"},
	"uuid":       {Code: "uuid", Width: 128, Endianness: "none"},
	"nbt_le":     {Code: "nbt_le", Width: 0, Endianness: "little"},
}
