package ingest

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"protocolgen/internal/claims"
	"protocolgen/internal/manifest"
)

// decodeJSONNumbers retains exact integer values and rejects trailing JSON values.
func decodeJSONNumbers(data []byte, value any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err != nil {
			return err
		}
		return fmt.Errorf("unexpected trailing JSON value")
	}
	return nil
}

func loadJSONFiles(root string) (map[string]any, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") && !strings.HasPrefix(entry.Name(), "__") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)
	documents := make(map[string]any, len(files))
	for _, name := range files {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", name, err)
		}
		var value any
		if err := decodeJSONNumbers(data, &value); err != nil {
			return nil, fmt.Errorf("parse %s: %w", name, err)
		}
		documents[name] = value
	}
	return documents, nil
}

func loadNamedJSONFiles(root string) (map[string]any, error) {
	documents, err := loadJSONFiles(root)
	if err != nil {
		return nil, err
	}
	return documents, nil
}

func asMap(value any) (map[string]any, bool) {
	object, ok := value.(map[string]any)
	return object, ok
}

func asArray(value any) ([]any, bool) {
	array, ok := value.([]any)
	return array, ok
}

func asInt(value any) (int64, bool) {
	switch number := value.(type) {
	case int:
		return int64(number), true
	case int64:
		return number, true
	case uint64:
		return int64(number), number <= uint64(^uint64(0)>>1)
	case float64:
		return int64(number), number == float64(int64(number))
	case json.Number:
		if parsed, err := number.Int64(); err == nil {
			return parsed, true
		}
		rational, ok := new(big.Rat).SetString(string(number))
		if !ok || !rational.IsInt() || !rational.Num().IsInt64() {
			return 0, false
		}
		return rational.Num().Int64(), true
	case string:
		var parsed int64
		_, err := fmt.Sscanf(number, "%d", &parsed)
		return parsed, err == nil
	default:
		return 0, false
	}
}

func asString(value any) string {
	valueString, _ := value.(string)
	return valueString
}

func valueAt(root any, pointer string) (any, bool) {
	if pointer == "" || pointer == "#" {
		return root, true
	}
	pointer = strings.TrimPrefix(pointer, "#")
	if !strings.HasPrefix(pointer, "/") {
		return nil, false
	}
	current := root
	for _, rawPart := range strings.Split(strings.TrimPrefix(pointer, "/"), "/") {
		part := strings.ReplaceAll(strings.ReplaceAll(rawPart, "~1", "/"), "~0", "~")
		switch typed := current.(type) {
		case map[string]any:
			var ok bool
			current, ok = typed[part]
			if !ok {
				return nil, false
			}
		case []any:
			var index int
			if _, err := fmt.Sscanf(part, "%d", &index); err != nil || index < 0 || index >= len(typed) {
				return nil, false
			}
			current = typed[index]
		default:
			return nil, false
		}
	}
	return current, true
}

func replaceAt(root any, pointer string, replacement any) error {
	if pointer == "" || pointer == "#" {
		return fmt.Errorf("root replacement is not supported; replace a named schema node")
	}
	parts := strings.Split(strings.TrimPrefix(strings.TrimPrefix(pointer, "#"), "/"), "/")
	current := root
	for index, rawPart := range parts {
		part := strings.ReplaceAll(strings.ReplaceAll(rawPart, "~1", "/"), "~0", "~")
		last := index == len(parts)-1
		switch typed := current.(type) {
		case map[string]any:
			if last {
				if _, ok := typed[part]; !ok {
					return fmt.Errorf("pointer %s does not identify an existing property", pointer)
				}
				typed[part] = replacement
				return nil
			}
			next, ok := typed[part]
			if !ok {
				return fmt.Errorf("pointer %s is missing %s", pointer, part)
			}
			current = next
		case []any:
			var arrayIndex int
			if _, err := fmt.Sscanf(part, "%d", &arrayIndex); err != nil || arrayIndex < 0 || arrayIndex >= len(typed) {
				return fmt.Errorf("pointer %s has invalid array index %s", pointer, part)
			}
			if last {
				typed[arrayIndex] = replacement
				return nil
			}
			current = typed[arrayIndex]
		default:
			return fmt.Errorf("pointer %s traverses a scalar", pointer)
		}
	}
	return fmt.Errorf("invalid empty pointer %s", pointer)
}

// canonicalNumbers stabilizes numeric spellings without rounding integers.
func canonicalNumbers(value any) any {
	switch typed := value.(type) {
	case json.Number:
		if rational, ok := new(big.Rat).SetString(string(typed)); ok && rational.IsInt() {
			return json.Number(rational.Num().String())
		}
		if number, err := typed.Float64(); err == nil {
			return number
		}
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, child := range typed {
			result[key] = canonicalNumbers(child)
		}
		return result
	case []any:
		result := make([]any, len(typed))
		for index, child := range typed {
			result[index] = canonicalNumbers(child)
		}
		return result
	}
	return value
}

func canonicalDigest(value any) (string, error) {
	data, err := json.Marshal(canonicalNumbers(value))
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func correctionContextDigest(pin manifest.SourcePin, target string, node any) (string, error) {
	return canonicalDigest(struct {
		Pin    manifest.SourcePin `json:"source_pin"`
		Target string             `json:"target"`
		Node   any                `json:"node"`
	}{Pin: pin, Target: target, Node: node})
}

func applyCorrections(documents map[string]any, directory string, pin manifest.SourcePin) ([]manifest.OverrideProof, error) {
	if directory == "" {
		return nil, nil
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("read corrections %s: %w", directory, err)
	}
	var paths []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			paths = append(paths, entry.Name())
		}
	}
	sort.Strings(paths)
	var proofs []manifest.OverrideProof
	for _, name := range paths {
		data, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			return nil, err
		}
		var patchFile map[string]any
		if err := decodeJSONNumbers(data, &patchFile); err != nil {
			return nil, fmt.Errorf("parse correction %s: %w", name, err)
		}
		operations, ok := asArray(patchFile["operations"])
		if !ok || len(operations) == 0 {
			return nil, fmt.Errorf("correction %s has no operations", name)
		}
		for index, rawOperation := range operations {
			operation, ok := asMap(rawOperation)
			if !ok {
				return nil, fmt.Errorf("correction %s operation %d is not an object", name, index)
			}
			id := asString(operation["id"])
			file := asString(operation["file"])
			pointer := asString(operation["pointer"])
			pre := asString(operation["pre_patch_sha256"])
			reason := asString(operation["why"])
			if id == "" || file == "" || pointer == "" || pre == "" || reason == "" {
				return nil, fmt.Errorf("correction %s operation %d is missing id/file/pointer/pre_patch_sha256/why", name, index)
			}
			document, ok := documents[file]
			if !ok {
				return nil, fmt.Errorf("correction %s operation %s references missing document %s", name, id, file)
			}
			node, ok := valueAt(document, pointer)
			if !ok {
				return nil, fmt.Errorf("correction %s operation %s references missing node %s#%s", name, id, file, pointer)
			}
			before, err := canonicalDigest(node)
			if err != nil {
				return nil, err
			}
			if before != pre {
				return nil, fmt.Errorf("stale correction %s: %s#%s pre-patch fingerprint is %s, current node is %s", id, file, pointer, pre, before)
			}
			context, err := correctionContextDigest(pin, file+"#"+pointer, node)
			if err != nil {
				return nil, err
			}
			if replacement, ok := operation["replace"]; ok {
				if err := replaceAt(document, pointer, replacement); err != nil {
					return nil, fmt.Errorf("apply correction %s: %w", id, err)
				}
			} else if additions, ok := asMap(operation["merge"]); ok {
				target, targetOK := asMap(node)
				if !targetOK {
					return nil, fmt.Errorf("correction %s operation %s cannot merge into a non-object", name, id)
				}
				for key, value := range additions {
					if _, exists := target[key]; exists {
						return nil, fmt.Errorf("correction %s operation %s merge would overwrite %s; use replace", name, id, key)
					}
					target[key] = value
				}
			} else {
				return nil, fmt.Errorf("correction %s operation %s has neither replace nor object merge", name, id)
			}
			afterNode, _ := valueAt(document, pointer)
			after, err := canonicalDigest(afterNode)
			if err != nil {
				return nil, err
			}
			evidence, err := parseEvidence(operation["evidence"], pin, file+pointer)
			if err != nil {
				return nil, fmt.Errorf("correction %s: %w", id, err)
			}
			proofs = append(proofs, manifest.OverrideProof{ID: id, Target: file + "#" + pointer, PrePatchNodeSHA256: pre, PrePatchContextSHA256: context, PostPatchNodeSHA256: after, Evidence: evidence, Reason: reason})
		}
	}
	return proofs, nil
}

func parseEvidence(raw any, pin manifest.SourcePin, fallbackLocator string) ([]manifest.Evidence, error) {
	values, ok := asArray(raw)
	if !ok || len(values) == 0 {
		return nil, fmt.Errorf("missing cited evidence")
	}
	result := make([]manifest.Evidence, 0, len(values))
	for index, rawValue := range values {
		object, ok := asMap(rawValue)
		if !ok {
			return nil, fmt.Errorf("evidence %d is not an object", index)
		}
		sourceID := asString(object["source_id"])
		if sourceID == "" {
			sourceID = pin.ID
		}
		locator := asString(object["locator"])
		if locator == "" {
			locator = fallbackLocator
		}
		if locator == "" {
			return nil, fmt.Errorf("evidence %d has no locator", index)
		}
		result = append(result, manifest.Evidence{SourceID: sourceID, Locator: locator, Summary: asString(object["summary"])})
	}
	return result, nil
}

func primitive(underlying string, options []string, jsonType string) manifest.Node {
	normalized := strings.ToLower(underlying)
	normalized = strings.NewReplacer("_", "", "-", "", " ", "").Replace(normalized)
	for _, option := range options {
		lower := strings.ToLower(option)
		if strings.Contains(lower, "varlong") || strings.Contains(lower, "varint64") {
			if normalized == "int64" || normalized == "long" {
				return manifest.Primitive("var_i64")
			}
			return manifest.Primitive("var_u64")
		}
		if strings.Contains(lower, "varint") {
			if normalized == "int32" || normalized == "int" || normalized == "byte" || normalized == "short" {
				return manifest.Primitive("var_i32")
			}
			return manifest.Primitive("var_u32")
		}
		if strings.Contains(lower, "zigzag64") {
			return manifest.Primitive("zigzag_i64")
		}
		if strings.Contains(lower, "zigzag") {
			return manifest.Primitive("zigzag_i32")
		}
	}
	compressed := false
	for _, option := range options {
		compressed = compressed || strings.EqualFold(option, "Compression")
	}
	if compressed {
		switch normalized {
		case "long", "int64":
			return manifest.Primitive("zigzag_i64")
		case "ulong", "uint64":
			return manifest.Primitive("var_u64")
		case "int", "int32":
			return manifest.Primitive("zigzag_i32")
		case "uint", "uint32":
			return manifest.Primitive("var_u32")
		}
	}
	if underlying == "" {
		underlying = jsonType
		normalized = strings.ToLower(underlying)
	}
	if normalized == "string" || normalized == "pstring" || normalized == "mcpe_string" {
		return manifest.String(manifest.Primitive("var_u32"))
	}
	if normalized == "bytearray" || normalized == "buffer" || normalized == "restbuffer" || normalized == "bytes" {
		return manifest.Bytes(manifest.Primitive("var_u32"))
	}
	if normalized == "bool" || normalized == "boolean" {
		return manifest.Primitive("bool")
	}
	if normalized == "uuid" || normalized == "mcpeuuid" {
		return manifest.Primitive("uuid")
	}
	code := map[string]string{
		"byte": "i8", "int8": "i8", "sbyte": "i8", "ubyte": "u8", "uint8": "u8",
		"short": "i16le", "int16": "i16le", "ushort": "u16le", "uint16": "u16le",
		"int": "i32le", "int32": "i32le", "uint": "u32le", "uint32": "u32le",
		"long": "i64le", "int64": "i64le", "ulong": "u64le", "uint64": "u64le",
		"float": "f32le", "float32": "f32le", "double": "f64le", "float64": "f64le",
		"varint": "var_i32", "varint32": "var_i32", "varlong": "var_i64", "varint64": "var_i64",
		"zigzag32": "zigzag_i32", "zigzag64": "zigzag_i64", "nbt": "nbt_le", "lnbt": "nbt_le",
	}
	if selected, ok := code[normalized]; ok {
		for _, option := range options {
			if strings.Contains(strings.ToLower(strings.ReplaceAll(option, " ", "")), "bigendian") {
				selected = strings.Replace(selected, "le", "be", 1)
			}
		}
		return manifest.Primitive(selected)
	}
	return manifest.Unresolved("unresolved primitive "+underlying, true)
}

func textOrBytes(underlying string, options []string, jsonType string, explicitText bool) manifest.Node {
	if explicitText || strings.EqualFold(underlying, "string") || strings.EqualFold(jsonType, "string") {
		return manifest.String(manifest.Primitive("var_u32"))
	}
	return primitive(underlying, options, jsonType)
}

func parseDirection(object map[string]any) manifest.Direction {
	value := strings.ToLower(asString(object["x-direction"]))
	if value == "" {
		value = strings.ToLower(asString(object["direction"]))
	}
	switch value {
	case "clientbound", "client_to_server", "serverbound":
		if value == "client_to_server" {
			return manifest.DirectionServerbound
		}
		if value == "serverbound" {
			return manifest.DirectionServerbound
		}
		return manifest.DirectionClientbound
	case "bidirectional":
		return manifest.DirectionBidirectional
	default:
		return manifest.DirectionUnknown
	}
}

func fieldCompatibility(object map[string]any) (reserved, ignored bool, compatibility []string) {
	reserved, _ = object["x-reserved"].(bool)
	ignored, _ = object["x-ignored"].(bool)
	if values, ok := asArray(object["x-compatibility"]); ok {
		for _, value := range values {
			if text := asString(value); text != "" {
				compatibility = append(compatibility, text)
			}
		}
	}
	if reserved {
		compatibility = append(compatibility, "reserved")
	}
	if ignored {
		compatibility = append(compatibility, "ignored")
	}
	sort.Strings(compatibility)
	return
}

func makeClaim(pin manifest.SourcePin, packetID uint32, packetName string, direction manifest.Direction, ordinal int, name string, object map[string]any, node manifest.Node, locator string) claims.Claim {
	semantic := asString(object["title"])
	if semantic == "" {
		semantic = name
	}
	typeID := node.TypeID
	if semantic == name && node.Semantic != "" {
		semantic = node.Semantic
	}
	if typeID == "" {
		semantic, typeID = unwrapIdentity(node, semantic, typeID)
	}
	return claims.Claim{
		SourceID: pin.ID, Locator: locator, PacketID: packetID, PacketName: packetName, Direction: direction,
		FieldPath: packetName + "." + name, Ordinal: ordinal, Name: name, Semantic: semantic, TypeID: typeID,
		Encode: node, Symmetry: manifest.Symmetric,
	}
}

func unwrapIdentity(node manifest.Node, semantic, typeID string) (string, string) {
	if node.Semantic != "" {
		semantic = node.Semantic
	}
	if node.TypeID != "" {
		typeID = node.TypeID
	}
	switch node.Kind {
	case manifest.KindOptional:
		if node.Value != nil {
			return unwrapIdentity(*node.Value, semantic, typeID)
		}
	case manifest.KindReserved, manifest.KindIgnored:
		if node.Element != nil {
			return unwrapIdentity(*node.Element, semantic, typeID)
		}
	}
	return semantic, typeID
}

func safeIdentifierName(name string) string {
	var builder strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			builder.WriteRune(r)
		} else {
			builder.WriteByte('_')
		}
	}
	return builder.String()
}
