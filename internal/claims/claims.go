// Package claims is the source-facing side of reconciliation. A claim is a
// complete machine-derived assertion about one packet field, including its
// semantic identity and locator. It is never a confidence vote.
package claims

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"protocolgen/internal/manifest"
)

type Claim struct {
	SourceID      string             `json:"source_id"`
	Locator       string             `json:"locator"`
	PacketID      uint32             `json:"packet_id"`
	PacketName    string             `json:"packet_name"`
	Direction     manifest.Direction `json:"direction"`
	FieldPath     string             `json:"field_path"`
	Ordinal       int                `json:"ordinal"`
	Name          string             `json:"name"`
	Semantic      string             `json:"semantic,omitempty"`
	TypeID        string             `json:"type_id,omitempty"`
	Encode        manifest.Node      `json:"encode"`
	Decode        *manifest.Node     `json:"decode,omitempty"`
	Symmetry      manifest.Symmetry  `json:"symmetry"`
	Reserved      bool               `json:"reserved,omitempty"`
	Ignored       bool               `json:"ignored,omitempty"`
	Compatibility []string           `json:"compatibility,omitempty"`
}

// PacketClaim records packet identity independently of its fields so empty
// packets remain part of the canonical protocol inventory.
type PacketClaim struct {
	SourceID  string             `json:"source_id"`
	Locator   string             `json:"locator"`
	ID        uint32             `json:"id"`
	Name      string             `json:"name"`
	Direction manifest.Direction `json:"direction"`
}

type Result struct {
	Pin       manifest.SourcePin
	Target    manifest.Target
	Packets   []PacketClaim
	Claims    []Claim
	Overrides []manifest.OverrideProof
}

func Fingerprint(claim Claim) (string, error) {
	return digest(withoutValidationConstraints(claim))
}

// WireFingerprint identifies only a claim's encoded contract, excluding source, naming, and validation metadata.
func WireFingerprint(claim Claim) (string, error) {
	claim.SourceID = ""
	claim.Locator = ""
	claim.PacketName = ""
	claim.Direction = ""
	claim.FieldPath = ""
	claim.Name = ""
	claim.Semantic = ""
	claim.TypeID = ""
	claim.Encode = NormalizeWireNode(claim.Encode)
	if claim.Decode != nil {
		decoded := NormalizeWireNode(*claim.Decode)
		claim.Decode = &decoded
	}
	return digest(claim)
}

// FieldWireFingerprint identifies the same encoded contract when it is already stored in a manifest field.
func FieldWireFingerprint(packetID uint32, direction manifest.Direction, field manifest.Field) (string, error) {
	return WireFingerprint(Claim{
		PacketID: packetID, Direction: direction, Ordinal: field.Ordinal,
		Encode: field.Encode, Decode: field.Decode, Symmetry: field.Symmetry,
		Reserved: field.Reserved, Ignored: field.Ignored, Compatibility: field.Compatibility,
	})
}

func ContextFingerprint(target manifest.Target, input []Claim) (string, error) {
	claims := append([]Claim(nil), input...)
	for index := range claims {
		claims[index] = withoutValidationConstraints(claims[index])
	}
	sort.SliceStable(claims, func(i, j int) bool {
		if claims[i].SourceID != claims[j].SourceID {
			return claims[i].SourceID < claims[j].SourceID
		}
		if claims[i].FieldPath != claims[j].FieldPath {
			return claims[i].FieldPath < claims[j].FieldPath
		}
		return claims[i].Ordinal < claims[j].Ordinal
	})
	return digest(struct {
		Target manifest.Target `json:"target"`
		Claims []Claim         `json:"claims"`
	}{Target: target, Claims: claims})
}

// Validation constraints do not alter the wire layout selected by an
// adjudication. Keep existing wire-evidence fingerprints stable when a source
// begins publishing bounds; constraint disagreements are resolved by fixing a
// source claim rather than silently changing a wire adjudication's scope.
func withoutValidationConstraints(claim Claim) Claim {
	claim.Encode = nodeWithoutValidationConstraints(claim.Encode)
	if claim.Decode != nil {
		decoded := nodeWithoutValidationConstraints(*claim.Decode)
		claim.Decode = &decoded
	}
	return claim
}

func nodeWithoutValidationConstraints(node manifest.Node) manifest.Node {
	node.Constraints = nil
	node.Prefix = nodePointerWithoutValidationConstraints(node.Prefix)
	node.Element = nodePointerWithoutValidationConstraints(node.Element)
	node.Value = nodePointerWithoutValidationConstraints(node.Value)
	node.Key = nodePointerWithoutValidationConstraints(node.Key)
	node.Control = nodePointerWithoutValidationConstraints(node.Control)
	node.Default = nodePointerWithoutValidationConstraints(node.Default)
	node.Elements = append([]manifest.Node(nil), node.Elements...)
	for index := range node.Elements {
		node.Elements[index] = nodeWithoutValidationConstraints(node.Elements[index])
	}
	node.Fields = append([]manifest.Field(nil), node.Fields...)
	for index := range node.Fields {
		node.Fields[index].Encode = nodeWithoutValidationConstraints(node.Fields[index].Encode)
		if node.Fields[index].Decode != nil {
			updated := nodeWithoutValidationConstraints(*node.Fields[index].Decode)
			node.Fields[index].Decode = &updated
		}
	}
	node.Variants = append([]manifest.Variant(nil), node.Variants...)
	for index := range node.Variants {
		node.Variants[index].Encode = nodeWithoutValidationConstraints(node.Variants[index].Encode)
		if node.Variants[index].Decode != nil {
			updated := nodeWithoutValidationConstraints(*node.Variants[index].Decode)
			node.Variants[index].Decode = &updated
		}
	}
	node.Cases = append([]manifest.Case(nil), node.Cases...)
	for caseIndex := range node.Cases {
		node.Cases[caseIndex].Encode = append([]manifest.Node(nil), node.Cases[caseIndex].Encode...)
		node.Cases[caseIndex].Decode = append([]manifest.Node(nil), node.Cases[caseIndex].Decode...)
		for index := range node.Cases[caseIndex].Encode {
			node.Cases[caseIndex].Encode[index] = nodeWithoutValidationConstraints(node.Cases[caseIndex].Encode[index])
		}
		for index := range node.Cases[caseIndex].Decode {
			node.Cases[caseIndex].Decode[index] = nodeWithoutValidationConstraints(node.Cases[caseIndex].Decode[index])
		}
	}
	return node
}

func nodePointerWithoutValidationConstraints(node *manifest.Node) *manifest.Node {
	if node == nil {
		return nil
	}
	updated := nodeWithoutValidationConstraints(*node)
	return &updated
}

// NormalizeWireNode removes source metadata while preserving wire properties and selector bindings.
func NormalizeWireNode(node manifest.Node) manifest.Node { return wireNodeInScope(node, nil, "") }

// wireNodeInScope binds conditional selectors to sibling ordinals before removing names.
func wireNodeInScope(node manifest.Node, bindings map[string]string, path string) manifest.Node {
	node = nodeWithoutValidationConstraints(node)
	if node.Kind == manifest.KindStruct {
		scope := make(map[string]string, len(bindings)+len(node.Fields))
		for name, ordinal := range bindings {
			scope[name] = ordinal
		}
		for _, field := range node.Fields {
			scope[field.Name] = "#" + path + "/" + strconv.Itoa(field.Ordinal)
		}
		bindings = scope
	}
	if ordinal, ok := bindings[node.CompareTo]; ok && node.Kind == manifest.KindConditional {
		node.CompareTo = ordinal
	}
	node.Semantic = ""
	if node.Kind == manifest.KindRecursive && node.Target == "" {
		node.Target = node.TypeID
	}
	node.TypeID = ""
	node.Reason = ""
	node.Prefix = wireNodePointerInScope(node.Prefix, bindings, path)
	node.Element = wireNodePointerInScope(node.Element, bindings, path)
	node.Value = wireNodePointerInScope(node.Value, bindings, path)
	node.Key = wireNodePointerInScope(node.Key, bindings, path)
	node.Control = wireNodePointerInScope(node.Control, bindings, path)
	node.Default = wireNodePointerInScope(node.Default, bindings, path)
	for index := range node.Elements {
		node.Elements[index] = wireNodeInScope(node.Elements[index], bindings, path)
	}
	for index := range node.Fields {
		field := &node.Fields[index]
		field.Name = ""
		field.Semantic = ""
		field.TypeID = ""
		field.Provenance = manifest.Provenance{}
		field.Encode = wireNodeInScope(field.Encode, bindings, path+"/"+strconv.Itoa(field.Ordinal))
		field.Decode = wireNodePointerInScope(field.Decode, bindings, path+"/"+strconv.Itoa(field.Ordinal))
	}
	for index := range node.Variants {
		node.Variants[index].Name = ""
		node.Variants[index].Encode = wireNodeInScope(node.Variants[index].Encode, bindings, path)
		node.Variants[index].Decode = wireNodePointerInScope(node.Variants[index].Decode, bindings, path)
	}
	for caseIndex := range node.Cases {
		for index := range node.Cases[caseIndex].Encode {
			node.Cases[caseIndex].Encode[index] = wireNodeInScope(node.Cases[caseIndex].Encode[index], bindings, path)
		}
		for index := range node.Cases[caseIndex].Decode {
			node.Cases[caseIndex].Decode[index] = wireNodeInScope(node.Cases[caseIndex].Decode[index], bindings, path)
		}
	}
	return node
}

// wireNodePointerInScope strips metadata while retaining selector bindings.
func wireNodePointerInScope(node *manifest.Node, bindings map[string]string, path string) *manifest.Node {
	if node == nil {
		return nil
	}
	updated := wireNodeInScope(*node, bindings, path)
	return &updated
}

func digest(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("fingerprint JSON: %w", err)
	}
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}
