package manifest

import (
	"fmt"
	"go/token"
	"math"
)

func validateSymbols(symbols []Symbol) error {
	ids, names := map[string]bool{}, map[string]bool{}
	for _, symbol := range symbols {
		if !token.IsIdentifier(symbol.Name) || symbol.ID != "symbols/"+symbol.Name {
			return fmt.Errorf("symbol %q must have an identifier name and matching symbols/ ID", symbol.ID)
		}
		if ids[symbol.ID] || names[symbol.Name] {
			return fmt.Errorf("duplicate symbol %q", symbol.ID)
		}
		ids[symbol.ID], names[symbol.Name] = true, true
		if symbol.Representation != "u32" || len(symbol.Values) == 0 {
			return fmt.Errorf("symbol %s requires u32 representation and nonempty values", symbol.ID)
		}
		values := map[string]bool{}
		for _, value := range symbol.Values {
			if !token.IsIdentifier(value.Name) || value.Name == "_" || values[value.Name] {
				return fmt.Errorf("symbol %s has invalid or duplicate value name %q", symbol.ID, value.Name)
			}
			values[value.Name] = true
			if value.Value < 0 || value.Value > math.MaxUint32 {
				return fmt.Errorf("symbol %s value %s does not fit u32", symbol.ID, value.Name)
			}
		}
		source := symbol.Source
		if source.ID == "" || source.Kind == "" || source.Revision == "" || source.Digest == "" ||
			source.Locator == "" || source.MinecraftVersion == "" || source.ProtocolVersion <= 0 {
			return fmt.Errorf("symbol %s requires a complete, versioned reference source", symbol.ID)
		}
		if len(symbol.Provenance.Pins) == 0 {
			return fmt.Errorf("symbol %s has no machine-derived provenance pins", symbol.ID)
		}
		if err := validateProvenance(symbol.Provenance, map[string]bool{source.ID: true}, symbol.ID); err != nil {
			return err
		}
	}
	return nil
}
