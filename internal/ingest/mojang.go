package ingest

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"protocolgen/internal/claims"
	"protocolgen/internal/manifest"
)

type mojangLowerer struct {
	documents map[string]any
	active    map[string]bool
	source    manifest.SourcePin
}

func ParseMojang(root string, pin manifest.SourcePin, corrections string) (claims.Result, error) {
	docsRoot := root
	if filepath.Base(filepath.Clean(root)) != "json" {
		candidate := filepath.Join(root, "json")
		if _, statErr := os.Stat(candidate); statErr == nil {
			docsRoot = candidate
		}
	}
	documents, err := loadJSONFiles(docsRoot)
	if err != nil {
		return claims.Result{}, fmt.Errorf("load Mojang JSON: %w", err)
	}
	proofs, err := applyCorrections(documents, corrections, pin)
	if err != nil {
		return claims.Result{}, err
	}
	target, err := mojangTarget(documents, pin)
	if err != nil {
		return claims.Result{}, err
	}
	lowerer := &mojangLowerer{documents: documents, active: map[string]bool{}, source: pin}
	type packetDoc struct {
		file string
		id   uint32
	}
	var packets []packetDoc
	for file, rawDocument := range documents {
		document, ok := asMap(rawDocument)
		if !ok {
			continue
		}
		meta, ok := asMap(document["$metaProperties"])
		if !ok {
			continue
		}
		idValue, ok := asInt(meta["[cereal:packet]"])
		if !ok || idValue < 0 {
			return claims.Result{}, fmt.Errorf("Mojang packet %s has invalid cereal packet id", file)
		}
		packets = append(packets, packetDoc{file: file, id: uint32(idValue)})
	}
	sort.Slice(packets, func(i, j int) bool {
		if packets[i].id != packets[j].id {
			return packets[i].id < packets[j].id
		}
		return packets[i].file < packets[j].file
	})
	if len(packets) == 0 {
		return claims.Result{}, fmt.Errorf("Mojang source contains no cereal packet schemas")
	}
	result := claims.Result{Pin: pin, Target: target, Overrides: proofs}
	for _, packetDoc := range packets {
		document := documents[packetDoc.file].(map[string]any)
		name := asString(document["title"])
		if name == "" {
			name = strings.TrimSuffix(packetDoc.file, filepath.Ext(packetDoc.file))
		}
		direction := parseDirection(document)
		result.Packets = append(result.Packets, claims.PacketClaim{SourceID: pin.ID, Locator: packetDoc.file, ID: packetDoc.id, Name: name, Direction: direction})
		body, bodyFile, err := lowerer.resolvePacketBody(document, packetDoc.file)
		if err != nil {
			return claims.Result{}, fmt.Errorf("Mojang packet %s: %w", packetDoc.file, err)
		}
		properties, ok := asMap(body["properties"])
		if !ok {
			if asString(body["type"]) != "object" {
				return claims.Result{}, fmt.Errorf("Mojang packet %s has no properties object", packetDoc.file)
			}
			properties = map[string]any{}
		}
		required := requiredNames(body["required"])
		fields, err := lowerMojangFields(lowerer, packetDoc.id, name, direction, bodyFile, properties, required, "")
		if err != nil {
			return claims.Result{}, err
		}
		result.Claims = append(result.Claims, fields...)
	}
	sort.Slice(result.Claims, func(i, j int) bool {
		if result.Claims[i].PacketID != result.Claims[j].PacketID {
			return result.Claims[i].PacketID < result.Claims[j].PacketID
		}
		return result.Claims[i].Ordinal < result.Claims[j].Ordinal
	})
	return result, nil
}

func (l *mojangLowerer) resolvePacketBody(document map[string]any, file string) (map[string]any, string, error) {
	seen := map[string]bool{}
	for reference := asString(document["$ref"]); reference != ""; reference = asString(document["$ref"]) {
		fileName, pointer := splitReference(reference, file)
		key := fileName + pointer
		if seen[key] {
			return nil, "", fmt.Errorf("cyclic root reference %s", reference)
		}
		seen[key] = true
		targetDocument, ok := l.documents[fileName]
		if !ok {
			return nil, "", fmt.Errorf("missing root reference %s", reference)
		}
		target, ok := valueAt(targetDocument, pointer)
		if !ok {
			return nil, "", fmt.Errorf("missing root reference target %s", reference)
		}
		document, ok = asMap(target)
		if !ok {
			return nil, "", fmt.Errorf("root reference %s is not an object", reference)
		}
		file = fileName
	}
	return document, file, nil
}

func mojangTarget(documents map[string]any, pin manifest.SourcePin) (manifest.Target, error) {
	target := manifest.Target{MinecraftVersion: pin.MinecraftVersion, ProtocolVersion: pin.ProtocolVersion}
	seenVersion, seenProtocol := "", 0
	for file, rawDocument := range documents {
		document, ok := asMap(rawDocument)
		if !ok {
			continue
		}
		version := asString(document["x-minecraft-version"])
		if version != "" {
			version = strings.Split(version, "-")[0]
			if seenVersion == "" {
				seenVersion = version
			} else if seenVersion != version {
				return manifest.Target{}, fmt.Errorf("Mojang source mixes Minecraft versions at %s", file)
			}
		}
		if protocol, ok := asInt(document["x-protocol-version"]); ok {
			if seenProtocol == 0 {
				seenProtocol = int(protocol)
			} else if seenProtocol != int(protocol) {
				return manifest.Target{}, fmt.Errorf("Mojang source mixes protocol versions at %s", file)
			}
		}
	}
	if seenVersion != "" && target.MinecraftVersion != "" && seenVersion != target.MinecraftVersion {
		return manifest.Target{}, fmt.Errorf("Mojang source version %s does not match locked target %s", seenVersion, target.MinecraftVersion)
	}
	if seenProtocol != 0 && target.ProtocolVersion != 0 && seenProtocol != target.ProtocolVersion {
		return manifest.Target{}, fmt.Errorf("Mojang source protocol %d does not match locked target %d", seenProtocol, target.ProtocolVersion)
	}
	if target.MinecraftVersion == "" {
		target.MinecraftVersion = seenVersion
	}
	if target.ProtocolVersion == 0 {
		target.ProtocolVersion = seenProtocol
	}
	if target.MinecraftVersion == "" || target.ProtocolVersion == 0 {
		return manifest.Target{}, fmt.Errorf("Mojang source has no complete version metadata")
	}
	return target, nil
}

func lowerMojangFields(lowerer *mojangLowerer, packetID uint32, packetName string, direction manifest.Direction, file string, properties map[string]any, required map[string]bool, parent string) ([]claims.Claim, error) {
	type property struct {
		name    string
		object  map[string]any
		ordinal int
	}
	propertiesSorted := make([]property, 0, len(properties))
	names := make([]string, 0, len(properties))
	for name := range properties {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		rawProperty := properties[name]
		object, ok := asMap(rawProperty)
		if !ok {
			return nil, fmt.Errorf("Mojang %s property %s is not an object", file, name)
		}
		ordinalValue, ok := asInt(object["x-ordinal-index"])
		if !ok || ordinalValue < 0 {
			return nil, fmt.Errorf("Mojang %s property %s has no explicit ordinal", file, name)
		}
		propertiesSorted = append(propertiesSorted, property{name: name, object: object, ordinal: int(ordinalValue)})
	}
	sort.Slice(propertiesSorted, func(i, j int) bool {
		if propertiesSorted[i].ordinal != propertiesSorted[j].ordinal {
			return propertiesSorted[i].ordinal < propertiesSorted[j].ordinal
		}
		return propertiesSorted[i].name < propertiesSorted[j].name
	})
	result := make([]claims.Claim, 0, len(propertiesSorted))
	for _, item := range propertiesSorted {
		fieldPath := item.name
		if parent != "" {
			fieldPath = parent + "." + item.name
		}
		node := lowerer.lowerSchema(item.object, file, packetName+safeIdentifierName(item.name))
		if !mojangAlwaysPresent(item.object, required[item.name]) {
			node = manifest.Optional(node)
			if hasOption(item.object, "+double-optional") {
				node = manifest.Optional(node)
			}
		} else if hasOption(item.object, "+double-optional") {
			node = manifest.Optional(manifest.Optional(node))
		}
		reserved, ignored, compatibility := fieldCompatibility(item.object)
		if reserved {
			node = manifest.Reserved(node)
		}
		if ignored {
			node = manifest.Ignored(node)
		}
		claim := makeClaim(lowerer.pin(), packetID, packetName, direction, item.ordinal, item.name, item.object, node, file+"#/properties/"+escapeJSONPointer(item.name))
		claim.FieldPath = packetName + "." + fieldPath
		claim.Reserved, claim.Ignored, claim.Compatibility = reserved, ignored, compatibility
		claim.Semantic = asString(item.object["title"])
		if claim.Semantic == "" {
			claim.Semantic = item.name
		}
		result = append(result, claim)
	}
	return result, nil
}

func (l *mojangLowerer) pin() manifest.SourcePin {
	return l.source
}

func (l *mojangLowerer) lowerSchema(schema map[string]any, file, hint string) manifest.Node {
	if strings.EqualFold(asString(schema["x-wire-kind"]), "nbt") {
		return manifest.Primitive("nbt_le")
	}
	if reference := asString(schema["$ref"]); reference != "" {
		return l.lowerReference(reference, file, hint, schema)
	}
	if branches, ok := asArray(schema["oneOf"]); ok {
		return l.lowerUnion(schema, branches, file, hint)
	}
	if allOf, ok := asArray(schema["allOf"]); ok && len(allOf) == 1 {
		if branch, ok := asMap(allOf[0]); ok {
			return l.lowerSchema(branch, file, hint)
		}
	}
	if _, ok := schema["enum"]; ok {
		return withMojangConstraints(l.lowerEnum(schema, file, hint), schema)
	}
	schemaType := asString(schema["type"])
	switch schemaType {
	case "object", "":
		if properties, ok := asMap(schema["properties"]); ok {
			fields, err := lowerMojangFields(l, 0, hint, manifest.DirectionUnknown, file, properties, requiredNames(schema["required"]), hint)
			if err != nil {
				return manifest.Unresolved(err.Error(), true)
			}
			return manifest.Node{Kind: manifest.KindStruct, Semantic: asString(schema["title"]), TypeID: hint, Fields: claimsToFields(fields)}
		}
		if additional, ok := schema["additionalProperties"]; ok {
			if valueSchema, ok := asMap(additional); ok {
				if entryProperties, ok := asMap(valueSchema["properties"]); ok {
					keySchema, keyOK := asMap(entryProperties["key"])
					entryValueSchema, valueOK := asMap(entryProperties["value"])
					if keyOK && valueOK {
						return withMojangConstraints(manifest.Map(
							manifest.Primitive("var_u32"),
							l.lowerSchema(keySchema, file, hint+"Key"),
							l.lowerSchema(entryValueSchema, file, hint+"Value"),
						), schema)
					}
				}
				return withMojangConstraints(manifest.Map(manifest.Primitive("var_u32"), manifest.String(manifest.Primitive("var_u32")), l.lowerSchema(valueSchema, file, hint+"Value")), schema)
			}
		}
		if schemaType == "" {
			return manifest.Unresolved("untyped Mojang schema "+hint, true)
		}
		return manifest.Struct()
	case "array":
		items, ok := asMap(schema["items"])
		if !ok {
			return manifest.Unresolved("array has no items "+hint, true)
		}
		inner := l.lowerSchema(items, file, hint+"Item")
		if min, okMin := asInt(schema["minItems"]); okMin {
			if max, okMax := asInt(schema["maxItems"]); okMax && min == max && min >= 0 {
				return withMojangConstraints(manifest.FixedArray(uint64(min), inner), schema)
			}
		}
		prefixCode := "var_u32"
		if explicit := asString(schema["x-valentine-array-count-type"]); explicit != "" {
			prefixCode = explicit
		} else if hasOption(schema, "No size compression") {
			prefixCode = "u32le"
		}
		return withMojangConstraints(manifest.Array(manifest.Primitive(prefixCode), inner), schema)
	case "string":
		if strings.EqualFold(asString(schema["x-wire-kind"]), "bytes") || strings.EqualFold(asString(schema["x-underlying-type"]), "bytearray") {
			return withMojangConstraints(manifest.Bytes(manifest.Primitive("var_u32")), schema)
		}
		return withMojangConstraints(manifest.String(manifest.Primitive("var_u32")), schema)
	case "boolean":
		return manifest.Primitive("bool")
	case "integer", "number":
		return withMojangConstraints(primitive(asString(schema["x-underlying-type"]), options(schema), schemaType), schema)
	case "null":
		return manifest.Void()
	default:
		return manifest.Unresolved("unsupported Mojang schema type "+schemaType, true)
	}
}

func withMojangConstraints(node manifest.Node, schema map[string]any) manifest.Node {
	constraints := manifest.Constraints{}
	setUint := func(key string, target **uint64) {
		if value, ok := asInt(schema[key]); ok && value >= 0 {
			converted := uint64(value)
			*target = &converted
		}
	}
	setFloat := func(key string, target **float64) {
		if value, ok := asFloat(schema[key]); ok {
			*target = &value
		}
	}
	setUint("minLength", &constraints.MinLength)
	setUint("maxLength", &constraints.MaxLength)
	setUint("minItems", &constraints.MinItems)
	setUint("maxItems", &constraints.MaxItems)
	// JSON Schema property bounds do not constrain arrays. Some published
	// packet schemas put maxProperties on an array; do not invent a wire bound.
	if kind := asString(schema["type"]); kind == "object" || kind == "" {
		setUint("minProperties", &constraints.MinProperties)
		setUint("maxProperties", &constraints.MaxProperties)
	}
	setFloat("minimum", &constraints.Minimum)
	setFloat("maximum", &constraints.Maximum)
	// Both sources also put prose descriptions in "pattern"; only a whole-string
	// regular expression constrains the wire value.
	if pattern := asString(schema["pattern"]); manifest.IsSchemaPattern(pattern) {
		constraints.Pattern = pattern
	}
	if constraints != (manifest.Constraints{}) {
		node.Constraints = &constraints
	}
	return node
}

func (l *mojangLowerer) lowerReference(reference, file, hint string, context map[string]any) manifest.Node {
	fileName, pointer := splitReference(reference, file)
	if strings.HasSuffix(pointer, "/3172631924") {
		return manifest.Primitive("nbt_le")
	}
	document, ok := l.documents[fileName]
	if !ok {
		return manifest.Unresolved("missing Mojang reference "+reference, true)
	}
	target, ok := valueAt(document, pointer)
	if !ok {
		return manifest.Unresolved("missing Mojang reference target "+reference, true)
	}
	typeID := fileName + pointer
	if l.active[typeID] {
		return manifest.Recursive(typeID)
	}
	targetObject, ok := asMap(target)
	if !ok {
		return manifest.Unresolved("Mojang reference is not a schema "+reference, true)
	}
	if strings.EqualFold(asString(targetObject["title"]), "mce::UUID") {
		result := manifest.Primitive("uuid")
		result.Semantic = "mce::UUID"
		result.TypeID = typeID
		return withMojangConstraints(result, context)
	}
	if nested := asString(targetObject["$ref"]); nested != "" {
		nestedFile, nestedPointer := splitReference(nested, fileName)
		if nestedFile == fileName && nestedPointer == pointer && targetObject["type"] == nil && targetObject["properties"] == nil && targetObject["oneOf"] == nil && targetObject["allOf"] == nil && targetObject["enum"] == nil {
			return manifest.Unresolved("bare self-referencing Mojang schema "+reference, true)
		}
	}
	targetType := asString(targetObject["type"])
	if targetType == "integer" || targetType == "number" || hasOption(context, "Enum-as-Value") || hasOption(targetObject, "Enum-as-Value") {
		copyOfTarget := cloneMap(targetObject)
		if underlying := asString(context["x-underlying-type"]); underlying != "" {
			copyOfTarget["x-underlying-type"] = underlying
		}
		// Serialization flags on the definition and this use both apply.
		values, _ := asArray(targetObject["x-serialization-options"])
		contextValues, _ := asArray(context["x-serialization-options"])
		copyOfTarget["x-serialization-options"] = append(append([]any(nil), values...), contextValues...)
		targetObject = copyOfTarget
	}
	l.active[typeID] = true
	result := l.lowerSchema(targetObject, fileName, hint)
	delete(l.active, typeID)
	if result.Semantic == "" {
		result.Semantic = asString(targetObject["title"])
	}
	result.TypeID = typeID
	return mergeMojangConstraints(result, context)
}

func mergeMojangConstraints(node manifest.Node, schema map[string]any) manifest.Node {
	overlay := withMojangConstraints(manifest.Node{}, schema).Constraints
	if overlay == nil {
		return node
	}
	if node.Constraints == nil {
		node.Constraints = overlay
		return node
	}
	merged := *node.Constraints
	for _, pair := range []struct{ dst, src **uint64 }{
		{&merged.MinLength, &overlay.MinLength}, {&merged.MaxLength, &overlay.MaxLength},
		{&merged.MinItems, &overlay.MinItems}, {&merged.MaxItems, &overlay.MaxItems},
		{&merged.MinProperties, &overlay.MinProperties}, {&merged.MaxProperties, &overlay.MaxProperties},
	} {
		if *pair.src != nil {
			*pair.dst = *pair.src
		}
	}
	if overlay.Minimum != nil {
		merged.Minimum = overlay.Minimum
	}
	if overlay.Maximum != nil {
		merged.Maximum = overlay.Maximum
	}
	if overlay.Pattern != "" {
		merged.Pattern = overlay.Pattern
	}
	node.Constraints = &merged
	return node
}

// lowerUnion writes a wire variant's zero-based index as a varuint32 before its fields.
func (l *mojangLowerer) lowerUnion(schema map[string]any, branches []any, file, hint string) manifest.Node {
	// A plain JSON oneOf can describe alternate representations of the same value,
	// such as hexadecimal and array colours. Only marked wire variants use this rule.
	if asString(schema["x-control-value-type"]) != "uint32" {
		return manifest.Unresolved("Mojang oneOf lacks a supported wire selector type "+hint, true)
	}
	if len(branches) == 0 {
		return manifest.Unresolved("Mojang oneOf has no alternatives "+hint, true)
	}
	variants := make([]manifest.Variant, 0, len(branches))
	for index, rawBranch := range branches {
		branch, ok := asMap(rawBranch)
		if !ok {
			return manifest.Unresolved(fmt.Sprintf("Mojang oneOf branch %d is not an object", index), true)
		}
		value := int64(index)
		name := asString(branch["title"])
		if name == "" {
			name = fmt.Sprintf("Variant%d", value)
		}
		variants = append(variants, manifest.Variant{Value: value, Name: name, Encode: l.lowerSchema(branch, file, hint+name)})
	}
	return manifest.Union(manifest.Primitive("var_u32"), variants...)
}

// lowerEnum reads the explicit numeric mapping or the string encoding selected at this use.
func (l *mojangLowerer) lowerEnum(schema map[string]any, file, hint string) manifest.Node {
	values, ok := asArray(schema["enum"])
	if !ok || len(values) == 0 {
		return manifest.Unresolved("Mojang enum has no values "+hint, true)
	}
	explicit, hasExplicit := asArray(schema["x-enum-binary-value"])
	if raw, exists := schema["x-enum-binary-value"]; exists && (!hasExplicit || len(explicit) != len(values)) {
		return manifest.Unresolved(fmt.Sprintf("Mojang enum has invalid binary value mapping %s: %v", hint, raw), true)
	}
	// U6 mappings distinguish named enums from numeric values. Older unmapped enums
	// remain unresolved, preserving their version-pinned adjudications.
	if hasExplicit && asString(schema["type"]) == "string" && !hasOption(schema, "Enum-as-Value") {
		return manifest.String(manifest.Primitive("var_u32"))
	}
	variants := make([]manifest.Variant, 0, len(values))
	seenNames := map[string]bool{}
	for index, rawValue := range values {
		value, numeric := mojangEnumInteger(rawValue)
		if hasExplicit {
			value, numeric = mojangEnumInteger(explicit[index])
		}
		if !numeric {
			return manifest.Unresolved("Mojang enum lacks unique explicit ordinals "+hint, true)
		}
		name := asString(rawValue)
		if name == "" {
			name = fmt.Sprintf("Value%d", value)
		}
		if seenNames[name] {
			return manifest.Unresolved("Mojang enum has duplicate names "+hint, true)
		}
		seenNames[name] = true
		variants = append(variants, manifest.Variant{Value: value, Name: name, Encode: manifest.Void()})
	}
	underlying := primitive(asString(schema["x-underlying-type"]), options(schema), asString(schema["type"]))
	if underlying.Kind != manifest.KindPrimitive || underlying.Primitive == nil {
		return manifest.Unresolved("Mojang enum has unresolved underlying codec "+hint, true)
	}
	return manifest.Node{Kind: manifest.KindEnum, Primitive: underlying.Primitive, Semantic: asString(schema["title"]), TypeID: file + "#" + hint, Variants: variants}
}

// mojangEnumInteger accepts exact JSON integers, never names or rounded floating-point values.
func mojangEnumInteger(raw any) (int64, bool) {
	if _, text := raw.(string); text {
		return 0, false
	}
	if value, floating := raw.(float64); floating && (value > 1<<53-1 || value < -(1<<53-1)) {
		return 0, false
	}
	return asInt(raw)
}

func claimsToFields(input []claims.Claim) []manifest.Field {
	fields := make([]manifest.Field, 0, len(input))
	for _, claim := range input {
		fields = append(fields, manifest.Field{Ordinal: claim.Ordinal, Name: claim.Name, Semantic: claim.Semantic, TypeID: claim.TypeID, Encode: claim.Encode, Decode: claim.Decode, Symmetry: claim.Symmetry, Reserved: claim.Reserved, Ignored: claim.Ignored, Compatibility: claim.Compatibility, Provenance: manifest.Provenance{Pins: []string{claim.SourceID}}})
	}
	return fields
}

// mojangAlwaysPresent reports whether a property is written unconditionally. Mojang omits a property
// from "required" when it declares a default, because the default is written in place of an unset value.
func mojangAlwaysPresent(schema map[string]any, listed bool) bool {
	if listed {
		return true
	}
	_, defaulted := schema["default"]
	return defaulted
}

func requiredNames(raw any) map[string]bool {
	result := map[string]bool{}
	if values, ok := asArray(raw); ok {
		for _, value := range values {
			if name := asString(value); name != "" {
				result[name] = true
			}
		}
	}
	return result
}

func options(schema map[string]any) []string {
	values, _ := asArray(schema["x-serialization-options"])
	result := make([]string, 0, len(values))
	for _, value := range values {
		if text := asString(value); text != "" {
			result = append(result, text)
		}
	}
	return result
}

func hasOption(schema map[string]any, want string) bool {
	for _, option := range options(schema) {
		if strings.EqualFold(option, want) {
			return true
		}
	}
	return false
}

func splitReference(reference, currentFile string) (string, string) {
	parts := strings.SplitN(reference, "#", 2)
	file := currentFile
	pointer := ""
	if len(parts) == 2 {
		pointer = parts[1]
	}
	if parts[0] != "" {
		file = filepath.Base(parts[0])
	}
	if pointer == "" {
		pointer = "#"
	}
	if !strings.HasPrefix(pointer, "#") {
		pointer = "#" + pointer
	}
	return file, pointer
}

func escapeJSONPointer(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(value, "~", "~0"), "/", "~1")
}

func cloneMap(input map[string]any) map[string]any {
	output := make(map[string]any, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}
