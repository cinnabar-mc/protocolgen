package ingest

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"protocolgen/internal/claims"
	"protocolgen/internal/manifest"
)

type endstoneLowerer struct {
	types  map[string]any
	enums  map[string]any
	active map[string]bool
	source manifest.SourcePin
}

type endstoneField struct {
	object        map[string]any
	sourceIndex   int
	outerOptional bool
}

func ParseEndstone(root string, pin manifest.SourcePin, corrections string) (claims.Result, error) {
	packetDocs, err := loadJSONFiles(filepath.Join(root, "packets"))
	if err != nil {
		return claims.Result{}, fmt.Errorf("load Endstone packets: %w", err)
	}
	typeDocs, err := loadJSONFiles(filepath.Join(root, "types"))
	if err != nil {
		return claims.Result{}, fmt.Errorf("load Endstone types: %w", err)
	}
	enumDocs, err := loadJSONFiles(filepath.Join(root, "enums"))
	if err != nil {
		return claims.Result{}, fmt.Errorf("load Endstone enums: %w", err)
	}
	combined := map[string]any{}
	for name, value := range packetDocs {
		combined["packets/"+name] = value
	}
	for name, value := range typeDocs {
		combined["types/"+name] = value
	}
	for name, value := range enumDocs {
		combined["enums/"+name] = value
	}
	proofs, err := applyCorrections(combined, corrections, pin)
	if err != nil {
		return claims.Result{}, err
	}
	for key, value := range combined {
		switch {
		case strings.HasPrefix(key, "packets/"):
			packetDocs[strings.TrimPrefix(key, "packets/")] = value
		case strings.HasPrefix(key, "types/"):
			typeDocs[strings.TrimPrefix(key, "types/")] = value
		case strings.HasPrefix(key, "enums/"):
			enumDocs[strings.TrimPrefix(key, "enums/")] = value
		}
	}
	target, err := endstoneTarget(root, pin)
	if err != nil {
		return claims.Result{}, err
	}
	lowerer := &endstoneLowerer{types: typeDocs, enums: enumDocs, active: map[string]bool{}, source: pin}
	type packetDoc struct {
		file string
		name string
		id   uint32
	}
	var packets []packetDoc
	for file, rawDocument := range packetDocs {
		document, ok := asMap(rawDocument)
		if !ok {
			return claims.Result{}, fmt.Errorf("Endstone packet %s is not an object", file)
		}
		idValue, ok := asInt(document["id"])
		if !ok || idValue < 0 {
			return claims.Result{}, fmt.Errorf("Endstone packet %s has invalid id", file)
		}
		name := asString(document["name"])
		if name == "" {
			name = strings.TrimSuffix(file, filepath.Ext(file))
		}
		packets = append(packets, packetDoc{file: file, name: name, id: uint32(idValue)})
	}
	sort.Slice(packets, func(i, j int) bool {
		if packets[i].id != packets[j].id {
			return packets[i].id < packets[j].id
		}
		return packets[i].name < packets[j].name
	})
	if len(packets) == 0 {
		return claims.Result{}, fmt.Errorf("Endstone source contains no packets")
	}
	result := claims.Result{Pin: pin, Target: target, Overrides: proofs}
	for _, packet := range packets {
		document := packetDocs[packet.file]
		object, ok := asMap(document)
		if !ok {
			return claims.Result{}, fmt.Errorf("Endstone packet %s document disappeared", packet.name)
		}
		rawFields, ok := asArray(object["fields"])
		if !ok {
			return claims.Result{}, fmt.Errorf("Endstone packet %s has no fields array", packet.name)
		}
		fields, err := canonicalEndstoneFields(rawFields, packet.name)
		if err != nil {
			return claims.Result{}, err
		}
		direction := parseDirection(object)
		result.Packets = append(result.Packets, claims.PacketClaim{SourceID: pin.ID, Locator: "packets/" + packet.file, ID: packet.id, Name: packet.name, Direction: direction})
		for index, canonicalField := range fields {
			field := canonicalField.object
			name := asString(field["name"])
			typeValue, ok := field["type"]
			if !ok {
				return claims.Result{}, fmt.Errorf("Endstone packet %s field %s has no type", packet.name, name)
			}
			node := lowerer.lowerTypeValue(typeValue, packet.name+safeIdentifierName(name), packet.name+"."+name)
			node = lowerer.applyFieldWrappers(node, field, packet.name+"."+name)
			if canonicalField.outerOptional {
				node = manifest.Optional(node)
			}
			reserved, ignored, compatibility := fieldCompatibility(field)
			if reserved {
				node = manifest.Reserved(node)
			}
			if ignored {
				node = manifest.Ignored(node)
			}
			claim := makeClaim(pin, packet.id, packet.name, direction, index, name, field, node, "packets/"+packet.file+"#/fields/"+fmt.Sprint(canonicalField.sourceIndex))
			claim.Reserved, claim.Ignored, claim.Compatibility = reserved, ignored, compatibility
			result.Claims = append(result.Claims, claim)
		}
	}
	return result, nil
}

func endstoneTarget(root string, pin manifest.SourcePin) (manifest.Target, error) {
	target := manifest.Target{MinecraftVersion: pin.MinecraftVersion, ProtocolVersion: pin.ProtocolVersion}
	readme, err := os.ReadFile(filepath.Join(root, "README.md"))
	if err == nil {
		for _, line := range strings.Split(string(readme), "\n") {
			if strings.Contains(line, "Minecraft Version:") {
				value := strings.TrimSpace(strings.Trim(strings.SplitN(line, "Minecraft Version:", 2)[1], " *`"))
				target.MinecraftVersion = strings.Join(strings.Split(value, ".")[:minInt(3, len(strings.Split(value, ".")))], ".")
			}
			if strings.Contains(line, "Network Version:") {
				value := strings.TrimSpace(strings.Trim(strings.SplitN(line, "Network Version:", 2)[1], " *`"))
				if parsed, ok := asInt(value); ok {
					target.ProtocolVersion = int(parsed)
				}
			}
		}
	}
	if target.MinecraftVersion != pin.MinecraftVersion || target.ProtocolVersion != pin.ProtocolVersion {
		return manifest.Target{}, fmt.Errorf("Endstone source target %s/%d does not match locked target %s/%d", target.MinecraftVersion, target.ProtocolVersion, pin.MinecraftVersion, pin.ProtocolVersion)
	}
	if target.MinecraftVersion == "" || target.ProtocolVersion == 0 {
		return manifest.Target{}, fmt.Errorf("Endstone source has no complete version metadata")
	}
	return target, nil
}

func (l *endstoneLowerer) lowerNamed(name, context string) manifest.Node {
	if name == "cereal::DynamicValue" {
		return cerealDynamicValue()
	}
	if primitiveNode := endstoneScalar(name); primitiveNode.Kind != manifest.KindUnresolved {
		return primitiveNode
	}
	document, ok := namedDocument(l.types, name)
	if !ok {
		return manifest.Unresolved("unknown Endstone type "+name+" at "+context, true)
	}
	if l.active[name] {
		return manifest.Recursive(name)
	}
	object, ok := asMap(document)
	if !ok {
		return manifest.Unresolved("Endstone type "+name+" is not an object", true)
	}
	l.active[name] = true
	result := l.lowerContainer(object, name, context)
	delete(l.active, name)
	result.Semantic = name
	result.TypeID = name
	return result
}

func cerealDynamicValue() manifest.Node {
	const typeID = "cereal::DynamicValue"
	recursive := manifest.Recursive(typeID)
	node := manifest.Union(
		manifest.Primitive("i32le"),
		manifest.Variant{Value: 0, Name: "None", Encode: manifest.Void()},
		manifest.Variant{Value: 1, Name: "Bool", Encode: manifest.Primitive("bool")},
		manifest.Variant{Value: 2, Name: "Int64", Encode: manifest.Primitive("i64le")},
		manifest.Variant{Value: 3, Name: "Double", Encode: manifest.Primitive("f64le")},
		manifest.Variant{Value: 4, Name: "String", Encode: manifest.String(manifest.Primitive("var_u32"))},
		manifest.Variant{Value: 5, Name: "List", Encode: manifest.Array(manifest.Primitive("var_u32"), recursive)},
		manifest.Variant{Value: 6, Name: "Map", Encode: manifest.Map(manifest.Primitive("var_u32"), manifest.String(manifest.Primitive("var_u32")), recursive)},
	)
	node.Semantic = typeID
	node.TypeID = typeID
	return node
}

func (l *endstoneLowerer) lowerContainer(object map[string]any, name, context string) manifest.Node {
	rawValues, ok := asArray(object["fields"])
	if !ok {
		return manifest.Unresolved("Endstone type "+name+" has no fields", true)
	}
	values, err := canonicalEndstoneFields(rawValues, name)
	if err != nil {
		return manifest.Unresolved(err.Error(), true)
	}
	fields := make([]manifest.Field, 0, len(values))
	for index, canonicalField := range values {
		field := canonicalField.object
		fieldName := asString(field["name"])
		typeValue := field["type"]
		node := l.lowerTypeValue(typeValue, name+safeIdentifierName(fieldName), context+"."+fieldName)
		node = l.applyFieldWrappers(node, field, context+"."+fieldName)
		if canonicalField.outerOptional {
			node = manifest.Optional(node)
		}
		fields = append(fields, manifest.Field{Ordinal: index, Name: fieldName, Semantic: fieldName, Encode: node, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{l.source.ID}}})
	}
	return manifest.Node{Kind: manifest.KindStruct, Semantic: name, TypeID: name, Fields: fields}
}

func canonicalEndstoneFields(values []any, context string) ([]endstoneField, error) {
	result := make([]endstoneField, 0, len(values))
	for index := 0; index < len(values); index++ {
		field, ok := asMap(values[index])
		if !ok {
			return nil, fmt.Errorf("Endstone %s field %d is not an object", context, index)
		}
		if asString(field["name"]) != "" {
			result = append(result, endstoneField{object: field, sourceIndex: index})
			continue
		}
		if !isAlwaysPresentOptionalMarker(field) || index+1 >= len(values) {
			return nil, fmt.Errorf("Endstone %s has unsupported anonymous field %d", context, index)
		}
		next, ok := asMap(values[index+1])
		if !ok || asString(next["name"]) == "" {
			return nil, fmt.Errorf("Endstone %s optional marker %d does not precede a named field", context, index)
		}
		result = append(result, endstoneField{object: next, sourceIndex: index + 1, outerOptional: true})
		index++
	}
	return result, nil
}

func isAlwaysPresentOptionalMarker(field map[string]any) bool {
	value, ok := field["value"].(bool)
	return ok && value && strings.EqualFold(asString(field["type"]), "bool")
}

func (l *endstoneLowerer) applyFieldWrappers(node manifest.Node, field map[string]any, context string) manifest.Node {
	if enumName := asString(field["enum"]); enumName != "" && node.Kind == manifest.KindPrimitive {
		node = l.lowerEnum(enumName, node, context, endstoneEnumConstraints(field))
	}
	if repeat, ok := field["repeat"]; ok {
		node = lowerEndstoneRepeat(repeat, node, context)
	}
	node = withEndstoneConstraints(node, field)
	if field["optional"] == true {
		node = manifest.Optional(node)
	}
	if field["double_optional"] == true {
		node = manifest.Optional(manifest.Optional(node))
	}
	return node
}

func withEndstoneConstraints(node manifest.Node, field map[string]any) manifest.Node {
	constraints, ok := asMap(field["constraints"])
	if !ok {
		return node
	}
	return withEndstoneConstraintMap(node, constraints)
}

func withEndstoneConstraintMap(node manifest.Node, constraints map[string]any) manifest.Node {
	translated := map[string]any{}
	switch node.Kind {
	case manifest.KindString:
		translated["minLength"], translated["maxLength"] = constraints["min_length"], constraints["max_length"]
		translated["pattern"] = constraints["pattern"]
	case manifest.KindBytes:
		translated["minLength"], translated["maxLength"] = constraints["min_length"], constraints["max_length"]
	case manifest.KindArray, manifest.KindFixedArray:
		translated["minItems"] = looserEndstoneCount(constraints["min_items"], constraints["min_properties"], true)
		translated["maxItems"] = looserEndstoneCount(constraints["max_items"], constraints["max_properties"], false)
		if items, ok := asMap(constraints["items"]); ok && node.Element != nil {
			element := withEndstoneConstraintMap(*node.Element, items)
			node.Element = &element
		}
	case manifest.KindMap:
		translated["minProperties"], translated["maxProperties"] = constraints["min_properties"], constraints["max_properties"]
		if propertyNames, ok := asMap(constraints["property_names"]); ok && node.Key != nil {
			key := withEndstoneConstraintMap(*node.Key, propertyNames)
			node.Key = &key
		}
		if additionalProperties, ok := asMap(constraints["additional_properties"]); ok && node.Value != nil {
			value := withEndstoneConstraintMap(*node.Value, additionalProperties)
			node.Value = &value
		}
	case manifest.KindPrimitive, manifest.KindEnum:
		translated["minimum"], translated["maximum"] = constraints["minimum"], constraints["maximum"]
	case manifest.KindUnion:
		if variantConstraints, ok := asArray(constraints["variant_types"]); ok {
			node.Variants = append([]manifest.Variant(nil), node.Variants...)
			for index := range node.Variants {
				if index >= len(variantConstraints) {
					break
				}
				if constraint, ok := asMap(variantConstraints[index]); ok {
					node.Variants[index].Encode = withEndstoneConstraintMap(node.Variants[index].Encode, constraint)
				}
			}
		}
	}
	return mergeMojangConstraints(node, translated)
}

func looserEndstoneCount(primary, alias any, minimum bool) any {
	left, leftOK := asInt(primary)
	right, rightOK := asInt(alias)
	if !leftOK {
		return alias
	}
	if !rightOK {
		return primary
	}
	if minimum && right < left || !minimum && right > left {
		return alias
	}
	return primary
}

func (l *endstoneLowerer) lowerTypeValue(value any, hint, context string) manifest.Node {
	if name, ok := value.(string); ok {
		return l.lowerNamed(name, context)
	}
	object, ok := asMap(value)
	if !ok {
		return manifest.Unresolved("Endstone type is not a string or object at "+context, true)
	}
	// U6 emits primitive cereal discriminators directly as strings. The
	// existing object form also carries legacy enum/name metadata.
	switchObject, hasSwitch := asMap(object["switch"])
	compactControl, compactSwitch := object["switch"].(string)
	if compactSwitch {
		switchObject = map[string]any{"type": compactControl}
		hasSwitch = true
	}
	if hasSwitch {
		controlName := asString(switchObject["type"])
		control := endstoneScalar(controlName)
		enumName := asString(switchObject["enum"])
		discriminants := l.enumValues(enumName)
		cases, ok := asArray(object["cases"])
		// Compact switches are cereal variants. Their outer index is separate
		// from any constrained legacy tag inside the selected branch.
		if ok && !compactSwitch {
			if variants, found := l.constrainedSwitchVariants(cases, control, asString(switchObject["name"]), enumName, context); found {
				return manifest.Union(control, variants...)
			}
		}
		if ok && enumName == "" {
			discriminants = make([]int64, len(cases))
			for index := range cases {
				discriminants[index] = int64(index)
			}
		}
		if !ok || len(cases) != len(discriminants) {
			return manifest.Unresolved("Endstone switch cases do not have explicit enum ordinals at "+context, true)
		}
		variants := make([]manifest.Variant, 0, len(cases))
		for index, rawCase := range cases {
			value := discriminants[index]
			if rawCase == nil {
				variants = append(variants, manifest.Variant{Value: value, Name: fmt.Sprintf("Empty%d", index), Encode: manifest.Void()})
				continue
			}
			caseName, ok := rawCase.(string)
			if !ok {
				caseName = fmt.Sprintf("Case%d", index)
			}
			variants = append(variants, manifest.Variant{Value: value, Name: caseName, Encode: l.lowerTypeValue(rawCase, hint+caseName, context)})
		}
		return manifest.Union(control, variants...)
	}
	if key, hasKey := object["key"]; hasKey {
		valueType, hasValue := object["value"]
		if !hasValue {
			return manifest.Unresolved("Endstone map has no value type at "+context, true)
		}
		return manifest.Map(manifest.Primitive("var_u32"), l.lowerTypeValue(key, hint+"Key", context+".key"), l.lowerTypeValue(valueType, hint+"Value", context+".value"))
	}
	if inner, hasType := object["type"]; hasType {
		var result manifest.Node
		if encoding := asString(object["encoding"]); encoding != "" && strings.EqualFold(asString(inner), "string") {
			result = manifest.String(manifest.Primitive("var_u32"))
		} else {
			result = l.lowerTypeValue(inner, hint, context)
		}
		if enumName := asString(object["enum"]); enumName != "" && result.Kind == manifest.KindPrimitive {
			result = l.lowerEnum(enumName, result, context, endstoneEnumConstraints(object))
		}
		if repeat, ok := object["repeat"]; ok {
			result = lowerEndstoneRepeat(repeat, result, context)
		}
		return result
	}
	return manifest.Unresolved("unsupported Endstone type at "+context, true)
}

func (l *endstoneLowerer) constrainedSwitchVariants(cases []any, control manifest.Node, controlField, enumName, context string) ([]manifest.Variant, bool) {
	names := map[int64]string{}
	for _, value := range l.enumValuesWithNames(enumName) {
		names[value.value] = value.name
	}
	seen := map[int64]bool{}
	var variants []manifest.Variant
	for _, rawCase := range cases {
		caseName, ok := rawCase.(string)
		if !ok {
			return nil, false
		}
		document, ok := namedDocument(l.types, caseName)
		if !ok {
			return nil, false
		}
		object, ok := asMap(document)
		if !ok {
			return nil, false
		}
		fields, ok := asArray(object["fields"])
		if !ok || len(fields) == 0 {
			return nil, false
		}
		discriminator, ok := asMap(fields[0])
		if !ok || controlField != "" && asString(discriminator["name"]) != controlField {
			return nil, false
		}
		constraints, ok := asMap(discriminator["constraints"])
		if !ok {
			return nil, false
		}
		values, ok := asArray(constraints["enum_values"])
		if !ok || len(values) == 0 {
			return nil, false
		}
		payload := l.lowerNamed(caseName, context)
		if payload.Kind != manifest.KindStruct || len(payload.Fields) == 0 {
			return nil, false
		}
		discriminatorNode := l.lowerTypeValue(discriminator["type"], caseName+"Discriminator", context+"."+asString(discriminator["name"]))
		discriminatorNode = l.applyFieldWrappers(discriminatorNode, discriminator, context+"."+asString(discriminator["name"]))
		if samePrimitiveEncoding(control, discriminatorNode) {
			payload.Fields = append([]manifest.Field(nil), payload.Fields[1:]...)
		}
		for _, rawValue := range values {
			value, ok := asInt(rawValue)
			if !ok || seen[value] {
				return nil, false
			}
			seen[value] = true
			name := names[value]
			if name == "" {
				name = fmt.Sprintf("%s%d", safeIdentifierName(caseName), value)
			}
			variants = append(variants, manifest.Variant{Value: value, Name: name, Encode: payload})
		}
	}
	sort.Slice(variants, func(i, j int) bool { return variants[i].Value < variants[j].Value })
	return variants, len(variants) != 0
}

func samePrimitiveEncoding(left, right manifest.Node) bool {
	if left.Kind == manifest.KindEnum {
		left.Kind = manifest.KindPrimitive
	}
	if right.Kind == manifest.KindEnum {
		right.Kind = manifest.KindPrimitive
	}
	return left.Kind == manifest.KindPrimitive && right.Kind == manifest.KindPrimitive && left.Primitive != nil && right.Primitive != nil && *left.Primitive == *right.Primitive
}

func (l *endstoneLowerer) lowerEnum(name string, underlying manifest.Node, context string, allowed map[int64]bool) manifest.Node {
	values := l.enumValuesWithNames(name)
	if len(values) == 0 || underlying.Primitive == nil {
		return manifest.Unresolved("Endstone enum "+name+" is incomplete at "+context, true)
	}
	variants := make([]manifest.Variant, 0, len(values))
	seen := map[int64]bool{}
	for _, value := range values {
		if len(allowed) != 0 && !allowed[value.value] || seen[value.value] {
			continue
		}
		seen[value.value] = true
		variants = append(variants, manifest.Variant{Value: value.value, Name: value.name, Encode: manifest.Void()})
	}
	if len(allowed) != 0 && len(seen) != len(allowed) {
		return manifest.Unresolved("Endstone enum "+name+" constraints reference unknown values at "+context, true)
	}
	return manifest.Node{Kind: manifest.KindEnum, Primitive: underlying.Primitive, Semantic: name, TypeID: "enums/" + name, Variants: variants}
}

func endstoneEnumConstraints(field map[string]any) map[int64]bool {
	constraints, ok := asMap(field["constraints"])
	if !ok {
		return nil
	}
	values, ok := asArray(constraints["enum_values"])
	if !ok {
		return nil
	}
	result := make(map[int64]bool, len(values))
	for _, rawValue := range values {
		value, ok := asInt(rawValue)
		if !ok {
			return nil
		}
		result[value] = true
	}
	return result
}

func (l *endstoneLowerer) enumValues(name string) []int64 {
	values := l.enumValuesWithNames(name)
	result := make([]int64, 0, len(values))
	for _, value := range values {
		result = append(result, value.value)
	}
	return result
}

type namedEnumValue struct {
	name  string
	value int64
}

func (l *endstoneLowerer) enumValuesWithNames(name string) []namedEnumValue {
	document, ok := namedDocument(l.enums, name)
	if !ok {
		return nil
	}
	object, ok := asMap(document)
	if !ok {
		return nil
	}
	values, ok := asArray(object["values"])
	if !ok {
		return nil
	}
	result := make([]namedEnumValue, 0, len(values))
	for _, rawValue := range values {
		value, ok := asMap(rawValue)
		if !ok {
			return nil
		}
		number, numberOK := asInt(value["value"])
		nameValue := asString(value["name"])
		if !numberOK || nameValue == "" {
			return nil
		}
		result = append(result, namedEnumValue{name: nameValue, value: number})
	}
	return result
}

func namedDocument(documents map[string]any, name string) (any, bool) {
	if document, ok := documents[name]; ok {
		return document, true
	}
	if document, ok := documents[name+".json"]; ok {
		return document, true
	}
	for file, document := range documents {
		if strings.TrimSuffix(file, filepath.Ext(file)) == name {
			return document, true
		}
		if object, ok := asMap(document); ok && asString(object["name"]) == name {
			return document, true
		}
	}
	return nil, false
}

func endstoneScalar(name string) manifest.Node {
	lowerName := strings.ToLower(name)
	for _, prefix := range []string{"brstd::bitset<", "std::bitset<"} {
		if strings.HasPrefix(lowerName, prefix) && strings.HasSuffix(lowerName, ">") {
			bits, err := strconv.ParseUint(strings.TrimSuffix(strings.TrimPrefix(lowerName, prefix), ">"), 10, 64)
			if err == nil && bits > 0 {
				return manifest.Bitset(bits)
			}
		}
	}
	switch strings.ToLower(name) {
	case "string", "pstring", "mcpe_string":
		return manifest.String(manifest.Primitive("var_u32"))
	case "restbuffer", "bytearray", "buffer", "bytes":
		return manifest.Bytes(manifest.Primitive("var_u32"))
	case "nbt", "lnbt", "compoundtag":
		return manifest.Primitive("nbt_le")
	case "bool", "boolean":
		return manifest.Primitive("bool")
	}
	known := map[string]string{"byte": "i8", "int8": "i8", "ubyte": "u8", "uint8": "u8", "short": "i16le", "int16": "i16le", "ushort": "u16le", "uint16": "u16le", "int": "i32le", "int32": "i32le", "int32_be": "i32be", "uint": "u32le", "uint32": "u32le", "long": "i64le", "int64": "i64le", "ulong": "u64le", "uint64": "u64le", "float": "f32le", "float32": "f32le", "double": "f64le", "float64": "f64le", "varint": "zigzag_i32", "varint32": "zigzag_i32", "varlong": "zigzag_i64", "varint64": "zigzag_i64", "uvarint32": "var_u32", "uvarint64": "var_u64", "uuid": "uuid", "mce::uuid": "uuid"}
	if code, ok := known[strings.ToLower(name)]; ok {
		return manifest.Primitive(code)
	}
	return manifest.Unresolved("unknown Endstone scalar "+name, false)
}

// lowerEndstoneRepeat preserves both legacy and compact U6 array framing.
func lowerEndstoneRepeat(raw any, inner manifest.Node, context string) manifest.Node {
	repeat, ok := asMap(raw)
	if !ok {
		if prefix, ok := raw.(string); ok {
			repeat = map[string]any{"prefix": prefix}
		} else if count, ok := asInt(raw); ok {
			repeat = map[string]any{"count": count}
		} else {
			return manifest.Unresolved("unsupported Endstone repeat at "+context, true)
		}
	}
	if prefix := asString(repeat["prefix"]); prefix != "" {
		prefixNode := endstoneScalar(prefix)
		if prefixNode.Kind == manifest.KindUnresolved {
			return manifest.Unresolved("unknown Endstone repeat prefix "+prefix+" at "+context, true)
		}
		return manifest.Array(prefixNode, inner)
	}
	if count, ok := asInt(repeat["count"]); ok && count > 0 {
		return manifest.FixedArray(uint64(count), inner)
	}
	return manifest.Unresolved("Endstone repeat lacks explicit prefix/count at "+context, true)
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}
