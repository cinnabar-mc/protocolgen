// Package flatten decides which packet payload structs are inlined into
// their packet. Endstone models most packets as one payload struct; when a
// packet's only field is a struct used nowhere else, that struct carries no
// information of its own, so its fields become the packet's fields, as the
// hand-written packages do. Sub-structs of multi-field packets keep their
// name because inlining them would strip the context their field names need.
package flatten

import (
	"protocolgen/internal/manifest"
	"protocolgen/internal/naming"
)

// Field is one effective packet field: the manifest field and the type ID of
// the struct that declares it, which is the packet name for top-level fields.
type Field struct {
	Owner string
	Field manifest.Field
}

// Usage counts how many places reference each named struct.
type Usage map[string]int

// Count indexes struct references across the manifest by type ID or inferred
// name, so a payload struct used by two packets is never inlined.
func Count(m manifest.Manifest) Usage {
	usage := Usage{}
	var walk func(node manifest.Node)
	walk = func(node manifest.Node) {
		switch node.Kind {
		case manifest.KindStruct:
			if id := TypeID(node); id != "" {
				usage[id]++
			}
		case manifest.KindRecursive:
			id := node.TypeID
			if id == "" {
				id = node.Target
			}
			if id != "" {
				usage[id]++
			}
		}
		for _, field := range node.Fields {
			walk(field.Encode)
			if field.Decode != nil {
				walk(*field.Decode)
			}
		}
		for _, variant := range node.Variants {
			walk(variant.Encode)
			if variant.Decode != nil {
				walk(*variant.Decode)
			}
		}
		for _, child := range []*manifest.Node{node.Prefix, node.Element, node.Value, node.Key, node.Control, node.Default} {
			if child != nil {
				walk(*child)
			}
		}
		for _, child := range node.Elements {
			walk(child)
		}
		for _, oneCase := range node.Cases {
			for _, child := range append(oneCase.Encode, oneCase.Decode...) {
				walk(child)
			}
		}
	}
	for _, packet := range m.Packets {
		for _, field := range packet.Fields {
			walk(field.Encode)
			if field.Decode != nil {
				walk(*field.Decode)
			}
		}
	}
	return usage
}

// TypeID is the identity the emitters key a node by.
func TypeID(node manifest.Node) string {
	if node.TypeID != "" {
		return node.TypeID
	}
	return naming.InferredTypeName(node)
}

// Inlined reports whether a field's struct is used only here and so is
// flattened into its packet. Native shapes (vectors, actor IDs) are never
// structs in Go and are left alone.
func (u Usage) Inlined(field manifest.Field, native func(manifest.Node) bool) bool {
	node := field.Encode
	if node.Kind != manifest.KindStruct || field.Decode != nil || len(node.Fields) == 0 || native(node) {
		return false
	}
	id := TypeID(node)
	return id == "" || u[id] == 1
}

// PacketFields returns a packet's effective fields: while the field list is
// one single-use payload struct, that struct's fields replace it.
func (u Usage) PacketFields(packet manifest.Packet, native func(manifest.Node) bool) []Field {
	owner, fields := packet.Name, packet.Fields
	for len(fields) == 1 && u.Inlined(fields[0], native) {
		owner, fields = TypeID(fields[0].Encode), fields[0].Encode.Fields
	}
	result := make([]Field, 0, len(fields))
	for _, field := range fields {
		result = append(result, Field{Owner: owner, Field: field})
	}
	return result
}
