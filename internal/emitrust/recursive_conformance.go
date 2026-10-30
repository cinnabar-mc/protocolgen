package emitrust

import (
	"fmt"
	"strings"

	"protocolgen/internal/manifest"
)

type recursiveValueFixture struct {
	name, control, listPrefix, mapPrefix, keyPrefix string
	leaf, list, dictionary, unknown                 int64
}

// recursiveValueShape derives a public test fixture from self-recursive list/map union arms.
func recursiveValueShape(item definition) (recursiveValueFixture, bool) {
	fixture := recursiveValueFixture{name: item.Name, control: wireCodecType(item.ControlCode)}
	control, err := manifest.PrimitiveForCode(item.ControlCode)
	if item.Kind != manifest.KindUnion || item.TypeID == "" || fixture.control == "" || err != nil || strings.HasPrefix(control.Code, "f") {
		return fixture, false
	}
	var leaf, list, dictionary bool
	used := map[int64]bool{}
	for _, variant := range item.Union {
		used[variant.Discriminant] = true
		node := variant.Node
		switch {
		case node.Kind == manifest.KindVoid:
			fixture.leaf, leaf = variant.Discriminant, true
		case node.Kind == manifest.KindArray && node.Element != nil && node.Element.Kind == manifest.KindRecursive && node.Element.Target == item.TypeID:
			if min, max, bounded := arrayBounds(node); bounded && (min > 1 || max < 1) {
				continue
			}
			code, err := prefixCode(node)
			if err == nil {
				fixture.listPrefix = wireCodecType(code)
				fixture.list, list = variant.Discriminant, fixture.listPrefix != ""
			}
		case node.Kind == manifest.KindMap && node.Key != nil && node.Value != nil && node.Key.Kind == manifest.KindString && node.Value.Kind == manifest.KindRecursive && node.Value.Target == item.TypeID:
			if min, max, bounded := mapBounds(node); bounded && (min > 1 || max < 1) {
				continue
			}
			if min, max, bounded := stringBounds(*node.Key); bounded && (min > 1 || max < 1) {
				continue
			}
			if node.Key.Constraints != nil && node.Key.Constraints.Pattern != "" {
				continue
			}
			code, mapErr := prefixCode(node)
			key, keyErr := prefixCode(*node.Key)
			if mapErr == nil && keyErr == nil {
				fixture.mapPrefix, fixture.keyPrefix = wireCodecType(code), wireCodecType(key)
				fixture.dictionary, dictionary = variant.Discriminant, fixture.mapPrefix != "" && fixture.keyPrefix != ""
			}
		}
	}
	for used[fixture.unknown] {
		fixture.unknown++
	}
	// Skip a saturated tag domain rather than truncate the unknown-tag fixture.
	width := control.Width
	if control.Signed {
		width--
	}
	if width < 63 && uint64(fixture.unknown) >= uint64(1)<<width {
		return fixture, false
	}
	return fixture, leaf && list && dictionary
}

// emitRecursiveValueTests exercises public generated values, with wire tags derived from schema.
func emitRecursiveValueTests(m manifest.Manifest, definitions []definition, contract int) string {
	crate := strings.ReplaceAll(rustPackageName(m, contract), "-", "_")
	var b strings.Builder
	unwrap := ""
	if contract > 0 {
		unwrap = ".unwrap()"
	}
	for index, item := range definitions {
		fixture, ok := recursiveValueShape(item)
		if !ok {
			continue
		}
		if b.Len() == 0 {
			b.WriteString("// Code generated from canonical protocol manifest v2. DO NOT EDIT.\n\n")
			fmt.Fprintf(&b, "use %s::wire::{self, Decode, Encode};\n", crate)
		}
		fmt.Fprintf(&b, "\nmod recursive_value_%d {\n    use super::*;\n    use %s::types::%s as Value;\n", index, crate, fixture.name)
		fmt.Fprintf(&b, "    const LEAF: i64 = %d;\n    const LIST: i64 = %d;\n    const MAP: i64 = %d;\n    const UNKNOWN: i64 = %d;\n", fixture.leaf, fixture.list, fixture.dictionary, fixture.unknown)
		fmt.Fprintf(&b, `
    // Encode the manifest's union selector without constructing a recursive value.
    fn tag(out: &mut Vec<u8>, value: i64) {
        out.extend(wire::%s(value as _).encode_to_vec()%s);
    }

    #[derive(Clone, Copy)]
    enum Shape { List, Map, Mixed }

    // Construct bytes directly so an over-budget fixture never creates an overdeep owned tree.
    fn encoded(shape: Shape, depth: usize, leaf: i64) -> Vec<u8> {
        let mut out = Vec::new();
        for level in 0..depth {
            let map = match shape {
                Shape::List => false,
                Shape::Map => true,
                Shape::Mixed => level %% 2 != 0,
            };
            tag(&mut out, if map { MAP } else { LIST });
            if map {
                out.extend(wire::%s(1).encode_to_vec()%s);
                out.extend(wire::%s(1).encode_to_vec()%s);
                out.push(b'k');
            } else {
                out.extend(wire::%s(1).encode_to_vec()%s);
            }
        }
        tag(&mut out, leaf);
        out
    }

    #[test]
    fn recursive_lists_maps_and_mixed_values_obey_the_edge_budget() {
        for shape in [Shape::List, Shape::Map, Shape::Mixed] {
            for depth in [0, 1, wire::MAX_RECURSIVE_VALUE_DEPTH] {
                let bytes = encoded(shape, depth, LEAF);
                let value = Value::decode_exact(&bytes).unwrap();
                assert_eq!(value.encode_to_vec()%s, bytes);
                let cloned = value.clone();
                assert_eq!(cloned, value);
                drop(value);
                drop(cloned);
            }
            let bytes = encoded(shape, wire::MAX_RECURSIVE_VALUE_DEPTH + 1, LEAF);
            assert_eq!(Value::decode_exact(&bytes), Err(wire::DecodeError::RecursiveDepthExceeded));
        }
    }

    #[test]
    fn recursive_budget_returns_after_success_and_child_decode_errors() {
        for shape in [Shape::List, Shape::Map, Shape::Mixed] {
            let valid = encoded(shape, wire::MAX_RECURSIVE_VALUE_DEPTH, LEAF);
            let mut bytes = valid.clone();
            bytes.extend_from_slice(&valid);
            let mut reader = wire::Reader::new(&bytes);
            Value::decode(&mut reader).unwrap();
            Value::decode(&mut reader).unwrap();
            reader.expect_consumed().unwrap();

            let mut bytes = encoded(shape, wire::MAX_RECURSIVE_VALUE_DEPTH, UNKNOWN);
            bytes.extend_from_slice(&valid);
            let mut reader = wire::Reader::new(&bytes);
            assert!(matches!(Value::decode(&mut reader), Err(wire::DecodeError::UnknownVariant { .. })));
            Value::decode(&mut reader).unwrap();
            reader.expect_consumed().unwrap();

            let mut bytes = encoded(shape, wire::MAX_RECURSIVE_VALUE_DEPTH + 1, LEAF);
            let refused_end = bytes.len();
            bytes.extend_from_slice(&valid);
            let mut reader = wire::Reader::new(&bytes);
            assert_eq!(Value::decode(&mut reader), Err(wire::DecodeError::RecursiveDepthExceeded));
            // Framing owns recovery: discard the refused value's known remaining bytes.
            let remaining = refused_end - reader.position();
            reader.take(remaining).unwrap();
            Value::decode(&mut reader).unwrap();
            reader.expect_consumed().unwrap();
        }
    }
}
`, fixture.control, unwrap, fixture.mapPrefix, unwrap, fixture.keyPrefix, unwrap, fixture.listPrefix, unwrap, unwrap)
	}
	return b.String()
}
