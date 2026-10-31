// Package nbtencoding applies the reviewed NBT wire-format overlay.
package nbtencoding

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"protocolgen/internal/manifest"
)

const SchemaVersion uint32 = 1

type Table struct {
	SchemaVersion uint32          `json:"schema_version"`
	Target        manifest.Target `json:"target"`
	Source        Source          `json:"source"`
	Fields        []Entry         `json:"fields"`
}

type Source struct {
	Repository string `json:"repository"`
	Revision   string `json:"revision"`
	Locator    string `json:"locator"`
	SHA256     string `json:"sha256"`
}

type Entry struct {
	PacketID     uint32               `json:"packet_id"`
	PacketName   string               `json:"packet_name"`
	FieldOrdinal int                  `json:"field_ordinal"`
	FieldName    string               `json:"field_name"`
	Path         string               `json:"path"`
	Encoding     manifest.NBTEncoding `json:"encoding"`
	Evidence     Evidence             `json:"evidence"`
}

type Evidence struct {
	Locator string `json:"locator"`
	Summary string `json:"summary"`
}

func Load(path string) (Table, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Table{}, fmt.Errorf("read NBT encoding table: %w", err)
	}
	var table Table
	if err := json.Unmarshal(data, &table); err != nil {
		return Table{}, fmt.Errorf("parse NBT encoding table: %w", err)
	}
	if err := table.validateShape(); err != nil {
		return Table{}, fmt.Errorf("validate NBT encoding table: %w", err)
	}
	return table, nil
}

func (t Table) Apply(value *manifest.Manifest) error {
	if value == nil {
		return fmt.Errorf("cannot apply NBT encoding table to nil manifest")
	}
	if err := t.Validate(value.Target, value.Packets); err != nil {
		return err
	}
	for packetIndex := range value.Packets {
		packet := &value.Packets[packetIndex]
		for fieldIndex := range packet.Fields {
			field := &packet.Fields[fieldIndex]
			for _, entry := range t.Fields {
				if entry.PacketID != packet.ID || entry.FieldOrdinal != field.Ordinal {
					continue
				}
				node, err := nbtNodeAt(field, entry.Path)
				if err != nil {
					return fmt.Errorf("NBT field %d.%d path %q: %w", packet.ID, field.Ordinal, entry.Path, err)
				}
				node.Encoding = string(entry.Encoding)
			}
		}
	}
	return manifest.Validate(*value)
}

func (t Table) Validate(target manifest.Target, packets []manifest.Packet) error {
	if err := t.validateShape(); err != nil {
		return err
	}
	if t.Target != target {
		return fmt.Errorf("NBT encoding table target does not match manifest target")
	}
	packetByID := make(map[uint32]*manifest.Packet, len(packets))
	for index := range packets {
		packetByID[packets[index].ID] = &packets[index]
	}
	seen := make(map[string]bool, len(t.Fields))
	for _, entry := range t.Fields {
		packet, ok := packetByID[entry.PacketID]
		if !ok {
			return fmt.Errorf("NBT encoding table references unknown packet id %d", entry.PacketID)
		}
		if packet.Name != entry.PacketName {
			return fmt.Errorf("NBT encoding table packet id %d names %q, manifest names %q", entry.PacketID, entry.PacketName, packet.Name)
		}
		fieldIndex := -1
		for index := range packet.Fields {
			if packet.Fields[index].Ordinal == entry.FieldOrdinal {
				fieldIndex = index
				break
			}
		}
		if fieldIndex < 0 {
			return fmt.Errorf("NBT encoding table references unknown field %d.%d", entry.PacketID, entry.FieldOrdinal)
		}
		field := packet.Fields[fieldIndex]
		if field.Name != entry.FieldName {
			return fmt.Errorf("NBT encoding table field %d.%d names %q, manifest names %q", entry.PacketID, entry.FieldOrdinal, entry.FieldName, field.Name)
		}
		node, err := nbtNodeAt(&field, entry.Path)
		if err != nil {
			return fmt.Errorf("NBT encoding table field %d.%d path %q: %w", entry.PacketID, entry.FieldOrdinal, entry.Path, err)
		}
		if !manifest.IsNBT(*node) {
			return fmt.Errorf("NBT encoding table field %d.%d path %q does not name an NBT node", entry.PacketID, entry.FieldOrdinal, entry.Path)
		}
		key := fmt.Sprintf("%d/%d/%s", entry.PacketID, entry.FieldOrdinal, entry.Path)
		if seen[key] {
			return fmt.Errorf("duplicate NBT encoding entry %s", key)
		}
		seen[key] = true
	}

	for _, packet := range packets {
		for _, field := range packet.Fields {
			for path := range nbtNodes(&field) {
				key := fmt.Sprintf("%d/%d/%s", packet.ID, field.Ordinal, path)
				if !seen[key] {
					return fmt.Errorf("missing NBT encoding entry for packet %d field %d path %q", packet.ID, field.Ordinal, path)
				}
			}
		}
	}
	return nil
}

func (t Table) validateShape() error {
	if t.SchemaVersion != SchemaVersion {
		return fmt.Errorf("schema_version %d is not v%d", t.SchemaVersion, SchemaVersion)
	}
	if t.Target.MinecraftVersion == "" || t.Target.ProtocolVersion <= 0 {
		return fmt.Errorf("target must pin minecraft_version and positive protocol_version")
	}
	if t.Source.Repository == "" || t.Source.Revision == "" || t.Source.Locator == "" || t.Source.SHA256 == "" {
		return fmt.Errorf("source is missing repository, revision, locator, or sha256")
	}
	if len(t.Source.Revision) != 40 || !isHex(t.Source.Revision) {
		return fmt.Errorf("source revision is not a full SHA-1")
	}
	if len(t.Source.SHA256) != len("sha256:")+64 || !strings.HasPrefix(t.Source.SHA256, "sha256:") || !isHex(strings.TrimPrefix(t.Source.SHA256, "sha256:")) {
		return fmt.Errorf("source sha256 is not a SHA-256 fingerprint")
	}
	for index, entry := range t.Fields {
		if entry.PacketID == 0 || entry.PacketName == "" || entry.FieldOrdinal < 0 || entry.FieldName == "" || entry.Path == "" {
			return fmt.Errorf("fields[%d] is incomplete", index)
		}
		if !manifest.ValidNBTEncoding(string(entry.Encoding)) {
			return fmt.Errorf("fields[%d] has invalid NBT encoding %q", index, entry.Encoding)
		}
		if entry.Evidence.Locator == "" || entry.Evidence.Summary == "" {
			return fmt.Errorf("fields[%d] has incomplete evidence", index)
		}
	}
	return nil
}

// nbtNodeAt accepts only canonical paths produced by the coverage traversal.
func nbtNodeAt(field *manifest.Field, path string) (*manifest.Node, error) {
	node, ok := nbtNodes(field)[path]
	if !ok {
		return nil, fmt.Errorf("path %q does not name an NBT node", path)
	}
	return node, nil
}

// nbtNodes indexes encode and explicit decode nodes with the same path grammar.
func nbtNodes(field *manifest.Field) map[string]*manifest.Node {
	result := map[string]*manifest.Node{}
	var walk func(*manifest.Node, string)
	walk = func(node *manifest.Node, path string) {
		if node == nil {
			return
		}
		if manifest.IsNBT(*node) {
			result[path] = node
		}
		for _, child := range []struct {
			name string
			node *manifest.Node
		}{
			{"prefix", node.Prefix}, {"element", node.Element}, {"value", node.Value}, {"key", node.Key}, {"control", node.Control}, {"default", node.Default},
		} {
			walk(child.node, path+"."+child.name)
		}
		for i := range node.Elements {
			walk(&node.Elements[i], fmt.Sprintf("%s.elements[%d]", path, i))
		}
		for i := range node.Fields {
			walk(&node.Fields[i].Encode, fmt.Sprintf("%s.fields[%d].encode", path, i))
			walk(node.Fields[i].Decode, fmt.Sprintf("%s.fields[%d].decode", path, i))
		}
		for i := range node.Variants {
			walk(&node.Variants[i].Encode, fmt.Sprintf("%s.variants[%d].encode", path, i))
			walk(node.Variants[i].Decode, fmt.Sprintf("%s.variants[%d].decode", path, i))
		}
		for i := range node.Cases {
			for j := range node.Cases[i].Encode {
				walk(&node.Cases[i].Encode[j], fmt.Sprintf("%s.cases[%d].encode[%d]", path, i, j))
			}
			for j := range node.Cases[i].Decode {
				walk(&node.Cases[i].Decode[j], fmt.Sprintf("%s.cases[%d].decode[%d]", path, i, j))
			}
		}
	}
	walk(&field.Encode, "encode")
	walk(field.Decode, "decode")
	return result
}

func isHex(value string) bool {
	_, err := hex.DecodeString(value)
	return err == nil
}
