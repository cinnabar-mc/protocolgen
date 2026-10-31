package ingest

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestMojangArrayReferenceUsesResolvedConstraintsAndOptions(t *testing.T) {
	l := &mojangLowerer{documents: map[string]any{"List.json": map[string]any{"type": "array", "items": map[string]any{"type": "string"}}}, active: map[string]bool{}}
	node := l.lowerSchema(map[string]any{"$ref": "List.json", "maxItems": 100, "maxProperties": 65535, "x-serialization-options": []any{"No size compression"}}, "Packet.json", "List")
	if node.Constraints == nil || node.Constraints.MaxItems == nil || *node.Constraints.MaxItems != 100 || node.Constraints.MaxProperties != nil {
		t.Fatalf("invalid reference constraints: %+v", node.Constraints)
	}
	if node.Prefix.Primitive.Code != "u32le" {
		t.Fatalf("array prefix = %s", node.Prefix.Primitive.Code)
	}
}

func TestParsersRejectPacketIDOverflow(t *testing.T) {
	for _, id := range []uint64{4294967295, 4294967297} {
		for _, kind := range []string{"mojang", "endstone"} {
			t.Run(kind+"/"+fmt.Sprint(id), func(t *testing.T) {
				root := t.TempDir()
				var err error
				if kind == "mojang" {
					writeJSON(t, filepath.Join(root, "Packet.json"), map[string]any{"title": "Packet", "type": "object", "$metaProperties": map[string]any{"[cereal:packet]": id}, "x-minecraft-version": "1.26.40", "x-protocol-version": 2168})
					_, err = ParseMojang(root, fixturePin(kind), "")
				} else {
					for _, directory := range []string{"packets", "types", "enums"} {
						if err := os.MkdirAll(filepath.Join(root, directory), 0755); err != nil {
							t.Fatal(err)
						}
					}
					writeJSON(t, filepath.Join(root, "packets", "Packet.json"), map[string]any{"name": "Packet", "id": id, "fields": []any{}})
					_, err = ParseEndstone(root, fixturePin(kind), "")
				}
				if (err != nil) != (id > 4294967295) {
					t.Fatalf("ID %d: %v", id, err)
				}
			})
		}
	}
}
