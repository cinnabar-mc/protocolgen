package emitrust

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"

	"protocolgen/internal/manifest"
)

// hasPrefixBounds reports numeric constraints that the container codec must enforce.
func hasPrefixBounds(node manifest.Node) bool {
	return node.Prefix != nil && node.Prefix.Constraints != nil &&
		(node.Prefix.Constraints.Minimum != nil || node.Prefix.Constraints.Maximum != nil)
}

// withPrefixBounds intersects prefix constraints with the container's length bounds.
func withPrefixBounds(node manifest.Node) (manifest.Node, error) {
	if !hasPrefixBounds(node) {
		return node, nil
	}
	code, err := prefixCode(node)
	if err != nil {
		return node, err
	}
	if code != "var_u32" && code != "u32le" {
		return node, fmt.Errorf("unsupported collection prefix %q", code)
	}
	constraints := manifest.Constraints{}
	if node.Constraints != nil {
		constraints = *node.Constraints
	}
	var lower, upper **uint64
	switch node.Kind {
	case manifest.KindString, manifest.KindBytes:
		lower, upper = &constraints.MinLength, &constraints.MaxLength
	case manifest.KindArray:
		lower, upper = &constraints.MinItems, &constraints.MaxItems
	case manifest.KindMap:
		lower, upper = &constraints.MinProperties, &constraints.MaxProperties
	default:
		return node, nil
	}
	minimum, maximum := uint64(0), uint64(math.MaxUint32)
	if *lower != nil {
		minimum = **lower
	}
	if *upper != nil {
		maximum = min(maximum, **upper)
	}
	prefix := node.Prefix.Constraints
	for _, bound := range []struct {
		value   *json.Number
		minimum bool
	}{{prefix.Minimum, true}, {prefix.Maximum, false}} {
		if bound.value == nil {
			continue
		}
		number, ok := new(big.Rat).SetString(string(*bound.value))
		if !ok || !number.IsInt() {
			return node, fmt.Errorf("prefix length bound must be an integer")
		}
		if number.Sign() < 0 {
			if bound.minimum {
				continue
			}
			return node, fmt.Errorf("prefix maximum is negative")
		}
		if number.Cmp(big.NewRat(math.MaxUint32, 1)) > 0 {
			if bound.minimum {
				return node, fmt.Errorf("prefix minimum exceeds %s range", code)
			}
			continue
		}
		value := number.Num().Uint64()
		if bound.minimum {
			minimum = max(minimum, value)
		} else {
			maximum = min(maximum, value)
		}
	}
	if minimum > maximum {
		return node, fmt.Errorf("prefix and container length constraints do not overlap")
	}
	*lower, *upper = &minimum, &maximum
	node.Constraints = &constraints
	return node, nil
}
