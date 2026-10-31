package flatten

import (
	"protocolgen/internal/manifest"
	"testing"
)

func TestRecursivePayloadKeepsDefinition(t *testing.T) {
	for _, targetOnly := range []bool{false, true} {
		ref := manifest.Recursive("Tree")
		if targetOnly {
			ref.TypeID = ""
		}
		tree := manifest.Node{Kind: manifest.KindStruct, TypeID: "Tree", Fields: []manifest.Field{{Name: "Children", Encode: manifest.Array(manifest.Primitive("var_u32"), ref)}}}
		m := manifest.Manifest{Packets: []manifest.Packet{{Name: "TreePacket", Fields: []manifest.Field{{Name: "Payload", Encode: tree}}}}}
		usage := Count(m)
		if usage["Tree"] != 2 || usage.Inlined(m.Packets[0].Fields[0], func(manifest.Node) bool { return false }) {
			t.Errorf("recursive definition flattened: %#v", usage)
		}
		tree.Fields = nil
		m.Packets[0].Fields[0].Encode = tree
		m.Packets = append(m.Packets, manifest.Packet{Name: "ReferencePacket", Fields: []manifest.Field{{Name: "Reference", Encode: ref}}})
		if got := Count(m)["Tree"]; got != 2 {
			t.Errorf("cross-packet count = %d", got)
		}
	}
}
