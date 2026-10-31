// Package parity adapts Axolotl's public v1 wire-manifest JSON into the same
// byte-equivalence normalization used by the v2 manifest. It does not inspect
// generated Rust or private Rust types.
package parity

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"

	"protocolgen/internal/manifest"
)

type normalizedPacket struct {
	ID         uint32         `json:"id"`
	Name       string         `json:"name"`
	Operations []normalizedOp `json:"operations"`
}

type normalizedOp struct {
	Kind      string              `json:"kind"`
	Path      string              `json:"path"`
	Code      string              `json:"code,omitempty"`
	Prefix    string              `json:"prefix,omitempty"`
	Encoding  string              `json:"encoding,omitempty"`
	Length    uint64              `json:"length,omitempty"`
	Presence  string              `json:"presence,omitempty"`
	CompareTo string              `json:"compare_to,omitempty"`
	Target    string              `json:"target,omitempty"`
	Children  []normalizedOp      `json:"children,omitempty"`
	Variants  []normalizedVariant `json:"variants,omitempty"`
	Cases     []normalizedCase    `json:"cases,omitempty"`
	Default   []normalizedOp      `json:"default,omitempty"`
}

type normalizedVariant struct {
	Value int64          `json:"value"`
	Name  string         `json:"name"`
	Ops   []normalizedOp `json:"operations"`
}

type normalizedCase struct {
	Value string         `json:"value"`
	Ops   []normalizedOp `json:"operations"`
}

type axolotlManifest struct {
	SchemaVersion    uint32          `json:"schema_version"`
	Source           string          `json:"source"`
	MinecraftVersion string          `json:"minecraft_version"`
	ProtocolVersion  int             `json:"protocol_version"`
	Packets          []axolotlPacket `json:"packets"`
}

type axolotlPacket struct {
	ID         uint32      `json:"id"`
	Name       string      `json:"name"`
	Operations []axolotlOp `json:"operations"`
}

type axolotlOp struct {
	Kind       string           `json:"kind"`
	Field      string           `json:"field"`
	Op         string           `json:"op"`
	Control    string           `json:"control"`
	Prefix     string           `json:"prefix"`
	Encoding   string           `json:"encoding"`
	Length     uint64           `json:"length"`
	Presence   string           `json:"presence"`
	CompareTo  string           `json:"compare_to"`
	TypeName   string           `json:"type_name"`
	Operations []axolotlOp      `json:"operations"`
	Element    []axolotlOp      `json:"element"`
	Value      []axolotlOp      `json:"value"`
	Variants   []axolotlVariant `json:"variants"`
	Cases      []axolotlCase    `json:"cases"`
	Default    []axolotlOp      `json:"default"`
}

type axolotlVariant struct {
	Value      int64       `json:"value"`
	Name       string      `json:"name"`
	Operations []axolotlOp `json:"operations"`
}

type axolotlCase struct {
	Value      string      `json:"value"`
	Operations []axolotlOp `json:"operations"`
}

func CompareFile(canonical manifest.Manifest, axolotlPath string) error {
	data, err := os.ReadFile(axolotlPath)
	if err != nil {
		return fmt.Errorf("read Axolotl manifest: %w", err)
	}
	return Compare(canonical, data)
}

func Compare(canonical manifest.Manifest, data []byte) error {
	if err := manifest.Validate(canonical); err != nil {
		return err
	}
	var source axolotlManifest
	if err := json.Unmarshal(data, &source); err != nil {
		return fmt.Errorf("parse Axolotl wire manifest: %w", err)
	}
	if source.SchemaVersion != 1 {
		return fmt.Errorf("Axolotl parity expects public wire manifest schema v1, got %d", source.SchemaVersion)
	}
	if source.ProtocolVersion != 0 && source.ProtocolVersion != canonical.Target.ProtocolVersion {
		return fmt.Errorf("mixes Axolotl protocol %d with canonical protocol %d", source.ProtocolVersion, canonical.Target.ProtocolVersion)
	}
	if source.MinecraftVersion != "" && source.MinecraftVersion != canonical.Target.MinecraftVersion {
		return fmt.Errorf("mixes Axolotl Minecraft %s with canonical Minecraft %s", source.MinecraftVersion, canonical.Target.MinecraftVersion)
	}
	want, err := normalizeCanonical(canonical)
	if err != nil {
		return err
	}
	got, err := normalizeAxolotl(source)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(want, got) {
		wantJSON, _ := json.Marshal(want)
		gotJSON, _ := json.Marshal(got)
		return fmt.Errorf("Axolotl parity mismatch after byte-equivalence normalization\ncanonical: %s\naxolotl: %s", wantJSON, gotJSON)
	}
	return nil
}

func normalizeCanonical(m manifest.Manifest) ([]normalizedPacket, error) {
	packets := make([]normalizedPacket, 0, len(m.Packets))
	for _, packet := range m.Packets {
		var operations []normalizedOp
		for _, field := range packet.Fields {
			if field.Decode != nil && field.Symmetry == manifest.Asymmetric {
				return nil, fmt.Errorf("Axolotl v1 parity cannot represent asymmetric field %s.%s", packet.Name, field.Name)
			}
			children, err := normalizeNode(field.Name, field.Encode)
			if err != nil {
				return nil, fmt.Errorf("canonical %s.%s: %w", packet.Name, field.Name, err)
			}
			operations = append(operations, children...)
		}
		packets = append(packets, normalizedPacket{ID: packet.ID, Name: packet.Name, Operations: operations})
	}
	sort.Slice(packets, func(i, j int) bool { return packets[i].ID < packets[j].ID })
	return packets, nil
}

func normalizeNode(path string, node manifest.Node) ([]normalizedOp, error) {
	switch node.Kind {
	case manifest.KindVoid:
		return nil, nil
	case manifest.KindPrimitive:
		if node.Primitive == nil {
			return nil, fmt.Errorf("primitive has no code")
		}
		return []normalizedOp{{Kind: "primitive", Path: path, Code: node.Primitive.Code, Encoding: node.Encoding}}, nil
	case manifest.KindString:
		prefix, err := prefixCode(node.Prefix)
		if err != nil {
			return nil, err
		}
		return []normalizedOp{{Kind: "string", Path: path, Prefix: prefix, Encoding: node.Encoding}}, nil
	case manifest.KindBytes:
		prefix, err := prefixCode(node.Prefix)
		if err != nil {
			return nil, err
		}
		return []normalizedOp{{Kind: "bytes", Path: path, Prefix: prefix}}, nil
	case manifest.KindArray, manifest.KindFixedArray:
		var children []normalizedOp
		if node.Element != nil {
			var err error
			children, err = normalizeNode(path+"[]", *node.Element)
			if err != nil {
				return nil, err
			}
		}
		op := normalizedOp{Kind: string(node.Kind), Path: path, Children: children}
		if node.Kind == manifest.KindArray {
			prefix, err := prefixCode(node.Prefix)
			if err != nil {
				return nil, err
			}
			op.Prefix = prefix
		} else {
			op.Length = node.Length
		}
		return []normalizedOp{normalizeByteArray(op)}, nil
	case manifest.KindSequence:
		var result []normalizedOp
		for _, element := range node.Elements {
			children, err := normalizeNode(path, element)
			if err != nil {
				return nil, err
			}
			result = append(result, children...)
		}
		return result, nil
	case manifest.KindOptional:
		if node.Value == nil {
			return nil, fmt.Errorf("optional has no value")
		}
		children, err := normalizeNode(path, *node.Value)
		if err != nil {
			return nil, err
		}
		return []normalizedOp{{Kind: "option", Path: path, Presence: "bool", Children: children}}, nil
	case manifest.KindStruct:
		var result []normalizedOp
		for _, field := range node.Fields {
			if field.Decode != nil {
				return nil, fmt.Errorf("Axolotl v1 parity cannot represent decode override at %s.%s", path, field.Name)
			}
			children, err := normalizeNode(path+"."+field.Name, field.Encode)
			if err != nil {
				return nil, err
			}
			result = append(result, children...)
		}
		return result, nil
	case manifest.KindMap:
		return nil, fmt.Errorf("map normalization requires an explicit map adapter")
	case manifest.KindUnion:
		control, err := primitiveCode(node.Control)
		if err != nil {
			return nil, err
		}
		variants := make([]normalizedVariant, 0, len(node.Variants))
		for _, variant := range node.Variants {
			if variant.Decode != nil {
				return nil, fmt.Errorf("Axolotl v1 parity cannot represent variant decode override at %s", path)
			}
			children, err := normalizeNode(path, variant.Encode)
			if err != nil {
				return nil, err
			}
			variants = append(variants, normalizedVariant{Value: variant.Value, Name: variant.Name, Ops: children})
		}
		return []normalizedOp{{Kind: "union", Path: path, Code: control, Variants: variants}}, nil
	case manifest.KindConditional:
		cases := make([]normalizedCase, 0, len(node.Cases))
		for _, oneCase := range node.Cases {
			if oneCase.Decode != nil {
				return nil, fmt.Errorf("Axolotl v1 parity cannot represent case decode override at %s", path)
			}
			var children []normalizedOp
			for _, child := range oneCase.Encode {
				ops, err := normalizeNode(path, child)
				if err != nil {
					return nil, err
				}
				children = append(children, ops...)
			}
			cases = append(cases, normalizedCase{Value: oneCase.Value, Ops: children})
		}
		var defaultOps []normalizedOp
		if node.Default != nil {
			var err error
			defaultOps, err = normalizeNode(path, *node.Default)
			if err != nil {
				return nil, err
			}
		}
		return []normalizedOp{{Kind: "conditional", Path: path, CompareTo: node.CompareTo, Cases: cases, Default: defaultOps}}, nil
	case manifest.KindEnum:
		return []normalizedOp{{Kind: "primitive", Path: path, Code: node.Primitive.Code}}, nil
	case manifest.KindReserved, manifest.KindIgnored:
		if node.Element == nil {
			return nil, fmt.Errorf("compatibility node has no element")
		}
		return normalizeNode(path, *node.Element)
	case manifest.KindRecursive:
		return []normalizedOp{{Kind: "recursive", Path: path, Target: node.Target}}, nil
	case manifest.KindOpaque, manifest.KindUnresolved:
		return nil, fmt.Errorf("%s node blocks parity: %s", node.Kind, node.Reason)
	default:
		return nil, fmt.Errorf("unsupported node %q", node.Kind)
	}
}

func normalizeAxolotl(source axolotlManifest) ([]normalizedPacket, error) {
	packets := make([]normalizedPacket, 0, len(source.Packets))
	for _, packet := range source.Packets {
		operations, err := normalizeAxolotlOps(packet.Operations)
		if err != nil {
			return nil, fmt.Errorf("Axolotl packet %s: %w", packet.Name, err)
		}
		packets = append(packets, normalizedPacket{ID: packet.ID, Name: packet.Name, Operations: operations})
	}
	sort.Slice(packets, func(i, j int) bool { return packets[i].ID < packets[j].ID })
	return packets, nil
}

func normalizeAxolotlOps(input []axolotlOp) ([]normalizedOp, error) {
	var result []normalizedOp
	for _, op := range input {
		switch op.Kind {
		case "primitive":
			encoding := ""
			if op.Op == "Nbt" {
				encoding = string(manifest.NBTNetwork)
			}
			result = append(result, normalizedOp{Kind: "primitive", Path: op.Field, Code: primitiveCodeName(op.Op), Encoding: encoding})
		case "string":
			result = append(result, normalizedOp{Kind: "string", Path: op.Field, Prefix: prefixCodeName(op.Prefix), Encoding: op.Encoding})
		case "encapsulated":
			children, err := normalizeAxolotlOps(op.Operations)
			if err != nil {
				return nil, err
			}
			result = append(result, normalizedOp{Kind: "encapsulated", Path: op.Field, Prefix: prefixCodeName(op.Prefix), Children: children})
		case "array":
			children, err := normalizeAxolotlOps(op.Element)
			if err != nil {
				return nil, err
			}
			result = append(result, normalizeByteArray(normalizedOp{Kind: "array", Path: op.Field, Prefix: prefixCodeName(op.Prefix), Children: children}))
		case "fixed_array":
			children, err := normalizeAxolotlOps(op.Element)
			if err != nil {
				return nil, err
			}
			result = append(result, normalizedOp{Kind: "fixed_array", Path: op.Field, Length: op.Length, Children: children})
		case "option":
			children, err := normalizeAxolotlOps(op.Value)
			if err != nil {
				return nil, err
			}
			result = append(result, normalizedOp{Kind: "option", Path: op.Field, Presence: primitiveCodeName(op.Presence), Children: children})
		case "union":
			variants := make([]normalizedVariant, 0, len(op.Variants))
			for _, variant := range op.Variants {
				children, err := normalizeAxolotlOps(variant.Operations)
				if err != nil {
					return nil, err
				}
				variants = append(variants, normalizedVariant{Value: variant.Value, Name: variant.Name, Ops: children})
			}
			control := op.Control
			if control == "" {
				control = op.Op
			}
			result = append(result, normalizedOp{Kind: "union", Path: op.Field, Code: primitiveCodeName(control), Variants: variants})
		case "conditional":
			cases := make([]normalizedCase, 0, len(op.Cases))
			for _, oneCase := range op.Cases {
				children, err := normalizeAxolotlOps(oneCase.Operations)
				if err != nil {
					return nil, err
				}
				cases = append(cases, normalizedCase{Value: oneCase.Value, Ops: children})
			}
			defaultOps, err := normalizeAxolotlOps(op.Default)
			if err != nil {
				return nil, err
			}
			result = append(result, normalizedOp{Kind: "conditional", Path: op.Field, CompareTo: op.CompareTo, Cases: cases, Default: defaultOps})
		case "recursive_reference":
			result = append(result, normalizedOp{Kind: "recursive", Path: op.Field, Target: op.TypeName})
		default:
			return nil, fmt.Errorf("unsupported Axolotl operation kind %q", op.Kind)
		}
	}
	return result, nil
}

// normalizeByteArray uses one shape for counted raw bytes and plain u8 arrays.
func normalizeByteArray(op normalizedOp) normalizedOp {
	if op.Kind == "array" && len(op.Children) == 1 && reflect.DeepEqual(op.Children[0], normalizedOp{Kind: "primitive", Path: op.Path + "[]", Code: "u8"}) {
		return normalizedOp{Kind: "bytes", Path: op.Path, Prefix: op.Prefix}
	}
	return op
}

func prefixCode(node *manifest.Node) (string, error) {
	if node == nil || node.Kind != manifest.KindPrimitive || node.Primitive == nil {
		return "", fmt.Errorf("missing explicit prefix")
	}
	return node.Primitive.Code, nil
}

func primitiveCode(node *manifest.Node) (string, error) {
	if node == nil || node.Kind != manifest.KindPrimitive || node.Primitive == nil {
		return "", fmt.Errorf("missing primitive control")
	}
	return node.Primitive.Code, nil
}

func primitiveCodeName(name string) string {
	mapName := map[string]string{
		"Bool": "bool", "U8": "u8", "I8": "i8", "U16BE": "u16be", "U16LE": "u16le", "I16BE": "i16be", "I16LE": "i16le",
		"U32BE": "u32be", "U32LE": "u32le", "I32BE": "i32be", "I32LE": "i32le", "U64BE": "u64be", "U64LE": "u64le", "I64BE": "i64be", "I64LE": "i64le",
		"F32BE": "f32be", "F32LE": "f32le", "F64BE": "f64be", "F64LE": "f64le", "VarInt": "var_i32", "VarLong": "var_i64", "ZigZag32": "zigzag_i32", "ZigZag64": "zigzag_i64", "Uuid": "uuid", "Nbt": "nbt_le", "RawBytes": "bytes",
	}
	if mapped, ok := mapName[name]; ok {
		return mapped
	}
	return name
}

func prefixCodeName(name string) string {
	switch name {
	case "VarInt":
		return "var_u32"
	case "VarLong":
		return "var_u64"
	default:
		return primitiveCodeName(name)
	}
}
