// Package naming contains the language-neutral public-name pipeline.
package naming

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	"protocolgen/internal/manifest"
)

// Overlay is the reviewed TypeID-to-name mapping used by both emitters.
type Overlay struct {
	Names map[string]string
}

// Entry is one reviewed public-name decision.
type Entry struct {
	TypeID    string `json:"type_id"`
	Name      string `json:"name"`
	Rationale string `json:"rationale"`
}

// Document is the on-disk reviewed naming overlay.
type Document struct {
	SchemaVersion uint32          `json:"schema_version"`
	Target        manifest.Target `json:"target"`
	Entries       []Entry         `json:"entries"`
}

// NewResolver creates a resolver whose names are collision-checked in the
// target language after the supplied casing function is applied.
func NewResolver(overlay Overlay) *Resolver {
	return &Resolver{
		overlay:   overlay,
		identity:  map[string]string{},
		usedNames: map[string]nameOwner{},
	}
}

// Resolver assigns stable names to manifest identities.
type Resolver struct {
	overlay   Overlay
	identity  map[string]string
	usedNames map[string]nameOwner
}

// nameOwner separates manifest identities from reserved symbols even when their text matches.
type nameOwner struct {
	identity string
	reserved bool
}

// Resolve returns a reviewed or heuristically derived name. A collision is an
// error; callers must add a reviewed overlay entry to disambiguate it.
func (r *Resolver) Resolve(node manifest.Node, hint string, casing func(string) string) (string, error) {
	key := IdentityKeyFor(node, hint)
	if name, ok := r.identity[key]; ok {
		return name, nil
	}
	base := hint
	inferred := InferredTypeName(node)
	if node.TypeID != "" {
		base = node.TypeID
	} else if node.Semantic != "" {
		base = node.Semantic
	} else if inferred != "" {
		base = inferred
	}

	neutral := r.overlay.Names[node.TypeID]
	if neutral == "" && inferred != "" {
		neutral = r.overlay.Names[inferred]
	}
	if neutral == "" {
		neutral = commonOverlaySuffix(node, r.overlay)
	}
	if neutral == "" {
		if node.TypeID != "" && LooksLikeArtifact(node.TypeID) {
			return "", fmt.Errorf("type ID %q requires a naming overlay entry", node.TypeID)
		}
		neutral = PublicTypeName(base)
	}
	name := casing(neutral)
	if name == "" {
		return "", fmt.Errorf("type ID %q resolved to an empty public name", node.TypeID)
	}
	if owner, ok := r.usedNames[name]; ok && owner != (nameOwner{identity: key}) {
		if node.TypeID == "" && inferred == "" {
			for _, suffix := range []string{"PacketData", "Data"} {
				candidate := casing(neutral + suffix)
				if _, used := r.usedNames[candidate]; !used {
					name = candidate
					break
				}
			}
			if name == casing(neutral) {
				return "", fmt.Errorf("public name %q collides between %s and %s; add naming overlay entries", name, owner.identity, key)
			}
		} else {
			return "", fmt.Errorf("public name %q collides between %s and %s; add naming overlay entries", name, owner.identity, key)
		}
	}
	r.identity[key] = name
	r.usedNames[name] = nameOwner{identity: key}
	return name, nil
}

// Reserve claims a public name for a packet or another non-node identity.
func (r *Resolver) Reserve(source, neutral string, casing func(string) string) error {
	name := casing(neutral)
	owner := nameOwner{identity: source, reserved: true}
	if previous, ok := r.usedNames[name]; ok && previous != owner {
		return fmt.Errorf("public name %q collides between %s and %s; add naming overlay entries", name, previous.identity, owner.identity)
	}
	r.usedNames[name] = owner
	return nil
}

// IdentityKey returns the stable identity key used for unnamed nodes.
func IdentityKey(node manifest.Node) string {
	return IdentityKeyFor(node, "")
}

// IdentityKeyFor returns the stable identity key, using hint for an unnamed
// node whose manifest carries no semantic or inferred identity.
func IdentityKeyFor(node manifest.Node, hint string) string {
	key := node.TypeID
	if key != "" {
		return key
	}
	inferred := InferredTypeName(node)
	identityHint := node.Semantic
	if identityHint == "" {
		identityHint = inferred
	}
	if identityHint == "" {
		identityHint = hint
	}
	if identityHint == "" {
		identityHint = string(node.Kind)
	}
	if node.Kind == manifest.KindUnion {
		identityHint += "/" + UnionIdentity(node)
	}
	return fmt.Sprintf("%s/%s/%s", node.Kind, node.Semantic, identityHint)
}

// PublicTypeName applies the shared normalization before target-language
// casing and keyword handling.
func PublicTypeName(value string) string {
	filenameTypeID := strings.HasSuffix(value, ".json#")
	value = strings.TrimSuffix(value, ".json#")
	value = strings.TrimSuffix(value, ".json")
	value = strings.TrimPrefix(value, "enums/")
	value = stripSharedTypeVersion(value)
	value = stripPacketPrefix(value, filenameTypeID)
	value = strings.ReplaceAll(value, "Packet::", "::")
	value = strings.ReplaceAll(value, "PacketPayload::", "::")
	value = strings.ReplaceAll(value, "PacketPayload", "")
	value = collapseRedundantGroup(value)
	if strings.HasSuffix(value, "PayloadUnion") {
		value = strings.TrimSuffix(value, "PayloadUnion") + "Value"
	}
	value = strings.TrimSuffix(value, "Union")
	value = strings.TrimSuffix(value, "Payload")
	return collapseImmediateDuplicates(typeName(value))
}

// PacketTypeName strips the transport suffix from a packet name.
func PacketTypeName(value string) string {
	name := typeName(value)
	name = strings.TrimSuffix(name, "Packet")
	if name == "" {
		return "Packet"
	}
	return name
}

// InferredTypeName identifies a shared union from the common variant prefix.
func InferredTypeName(node manifest.Node) string {
	if node.Kind != manifest.KindUnion || len(node.Variants) == 0 {
		return ""
	}
	prefix := ""
	for _, variant := range node.Variants {
		qualified := variant.Name
		if !strings.Contains(qualified, "::") && variant.Encode.TypeID != "" {
			qualified = variant.Encode.TypeID
		}
		position := strings.LastIndex(qualified, "::")
		if position < 0 {
			return ""
		}
		candidate := qualified[:position]
		if prefix == "" {
			prefix = candidate
		} else if prefix != candidate {
			return ""
		}
	}
	return prefix
}

// UnionIdentity returns a deterministic structural identity for an unnamed
// union.
func UnionIdentity(node manifest.Node) string {
	type variantIdentity struct {
		Value  int64  `json:"value"`
		Name   string `json:"name"`
		Encode any    `json:"encode"`
		Decode any    `json:"decode,omitempty"`
	}
	identity := struct {
		Control  string            `json:"control"`
		Variants []variantIdentity `json:"variants"`
	}{}
	if node.Control != nil && node.Control.Primitive != nil {
		identity.Control = node.Control.Primitive.Code
	}
	for _, variant := range node.Variants {
		payload := variantIdentity{Value: variant.Value, Name: variant.Name, Encode: payloadIdentity(variant.Encode)}
		if variant.Decode != nil {
			payload.Decode = payloadIdentity(*variant.Decode)
		}
		identity.Variants = append(identity.Variants, payload)
	}
	encoded, _ := json.Marshal(identity)
	return fmt.Sprintf("%x", sha256.Sum256(encoded))
}

// LayoutIdentity compares a definition's wire shape while retaining nested named references.
func LayoutIdentity(node manifest.Node) string {
	node.TypeID = ""
	encoded, _ := json.Marshal(inlineIdentity(node, true))
	return fmt.Sprintf("%x", sha256.Sum256(encoded))
}

// payloadIdentity distinguishes inline layouts without tying them to source pins.
func payloadIdentity(node manifest.Node) any {
	if node.TypeID != "" {
		return node.TypeID
	}
	return inlineIdentity(node, false)
}

// inlineIdentity decodes exact wire numbers before removing non-layout metadata.
func inlineIdentity(node manifest.Node, wireOnly bool) any {
	data, _ := json.Marshal(node)
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var identity any
	_ = decoder.Decode(&identity)
	return stripIdentityMetadata(identity, wireOnly)
}

// stripIdentityMetadata retains named references and exact inline wire numbers.
func stripIdentityMetadata(value any, wireOnly bool) any {
	switch value := value.(type) {
	case map[string]any:
		if _, node := value["kind"]; node {
			if id, ok := value["type_id"].(string); ok && id != "" {
				return id
			}
		}
		delete(value, "provenance")
		if wireOnly {
			delete(value, "semantic")
			delete(value, "constraints")
			if _, node := value["kind"]; !node {
				delete(value, "type_id")
			}
		}
		for key, child := range value {
			value[key] = stripIdentityMetadata(child, wireOnly)
		}
		return value
	case []any:
		for index, child := range value {
			value[index] = stripIdentityMetadata(child, wireOnly)
		}
		return value
	default:
		return value
	}
}

// LooksLikeArtifact identifies names that must be explicitly reviewed.
func LooksLikeArtifact(value string) bool {
	if strings.ContainsAny(value, "<>") {
		return true
	}
	for _, token := range strings.FieldsFunc(value, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) }) {
		for _, word := range camelWords(token) {
			if strings.HasPrefix(word, "Empty") && allDigits(word[len("Empty"):]) {
				return true
			}
			if strings.HasPrefix(word, "T") && len(word) > 1 && allDigits(word[1:]) {
				return true
			}
		}
	}
	return false
}

// PublicVariantName removes numeric placeholders after they have been
// explicitly reviewed through their owning TypeID.
func PublicVariantName(value string) string {
	if strings.HasPrefix(value, "Empty") && allDigits(value[len("Empty"):]) {
		return "Empty"
	}
	return value
}

func commonOverlaySuffix(node manifest.Node, overlay Overlay) string {
	var names []string
	for _, variant := range node.Variants {
		if name := overlay.Names[variant.Encode.TypeID]; name != "" {
			names = append(names, name)
		}
	}
	if len(names) < 2 {
		return ""
	}
	common := camelWords(names[0])
	for _, name := range names[1:] {
		common = commonSuffixWords(common, camelWords(name))
		if len(common) == 0 {
			return ""
		}
	}
	if len(common) == 0 {
		return ""
	}
	var b strings.Builder
	for _, word := range common {
		b.WriteString(word)
	}
	return b.String()
}

func commonSuffixWords(first, second []string) []string {
	count := len(first)
	if len(second) < count {
		count = len(second)
	}
	for count > 0 {
		match := true
		for index := 0; index < count; index++ {
			if !strings.EqualFold(first[len(first)-count+index], second[len(second)-count+index]) {
				match = false
				break
			}
		}
		if match {
			break
		}
		count--
	}
	if count == 0 {
		return nil
	}
	return first[len(first)-count:]
}

func stripPacketPrefix(value string, filenameTypeID bool) string {
	if parts := strings.Split(value, "::"); len(parts) > 1 {
		prefix := parts[0]
		if strings.HasSuffix(prefix, "PacketPayload") || strings.HasSuffix(prefix, "Packet") {
			return strings.Join(parts[1:], "::")
		}
	}
	if !filenameTypeID {
		return value
	}
	for _, token := range []string{"PacketPayload", "Packet"} {
		if position := strings.Index(value, token); position > 0 && position+len(token) < len(value) {
			return value[position+len(token):]
		}
	}
	return value
}

func stripSharedTypeVersion(value string) string {
	const prefix = "SharedTypes::"
	if !strings.HasPrefix(value, prefix) {
		return value
	}
	rest := strings.TrimPrefix(value, prefix)
	if separator := strings.Index(rest, "::"); separator >= 0 && strings.HasPrefix(rest[:separator], "v") {
		return rest[separator+2:]
	}
	return rest
}

func collapseRedundantGroup(value string) string {
	parts := strings.Split(value, "::")
	if len(parts) == 2 && parts[0] == parts[1]+"Group" {
		return parts[1]
	}
	return value
}

func typeName(value string) string {
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
	result := b.String()
	if result == "" {
		return "Generated"
	}
	if unicode.IsDigit([]rune(result)[0]) {
		return "Generated" + result
	}
	return result
}

func collapseImmediateDuplicates(value string) string {
	words := camelWords(value)
	if len(words) < 2 {
		return value
	}
	var b strings.Builder
	previous := ""
	for _, word := range words {
		if strings.EqualFold(word, previous) {
			continue
		}
		b.WriteString(word)
		previous = word
	}
	return b.String()
}

func camelWords(value string) []string {
	runes := []rune(value)
	if len(runes) == 0 {
		return nil
	}
	start := 0
	words := make([]string, 0, 4)
	for index := 1; index < len(runes); index++ {
		previous, current := runes[index-1], runes[index]
		nextLower := index+1 < len(runes) && unicode.IsLower(runes[index+1])
		boundary := unicode.IsUpper(current) && (unicode.IsLower(previous) || unicode.IsDigit(previous) || unicode.IsUpper(previous) && nextLower)
		if boundary {
			words = append(words, string(runes[start:index]))
			start = index
		}
	}
	return append(words, string(runes[start:]))
}

func allDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
