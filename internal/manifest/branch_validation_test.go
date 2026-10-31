package manifest

import "testing"

func TestConditionalDecodeIsValidated(t *testing.T) {
	malformed := Primitive("u8")
	malformed.Primitive.Width = 64
	for _, child := range []Node{Unresolved("missing", true), {Kind: KindOpaque}, malformed} {
		node := Node{Kind: KindConditional, CompareTo: "selector", Cases: []Case{{Value: "0", Encode: []Node{Void()}, Decode: []Node{child}}}}
		if err := validateNode(node, "root", nil); err == nil {
			t.Errorf("accepted invalid decode node %s", child.Kind)
		}
	}
	valid := Node{Kind: KindConditional, CompareTo: "selector", Cases: []Case{{Value: "0", Encode: []Node{Void()}, Decode: []Node{Primitive("u8")}}}}
	if err := validateNode(valid, "root", nil); err != nil {
		t.Fatal(err)
	}
}

func TestPrefixesAndControlsRequireExactPrimitives(t *testing.T) {
	for _, mutate := range []func(*Node){
		func(n *Node) { n.Primitive.Width = 64 },
		func(n *Node) { n.Primitive.Signed = true },
		func(n *Node) { n.Primitive.Endianness = "big" },
	} {
		prefix := Primitive("u8")
		mutate(&prefix)
		for _, node := range []Node{Bytes(prefix), Union(prefix, Variant{Value: 256, Name: "Overflow", Encode: Void()})} {
			if err := validateNode(node, "root", nil); err == nil {
				t.Errorf("accepted invented %s primitive", node.Kind)
			}
		}
	}
}
