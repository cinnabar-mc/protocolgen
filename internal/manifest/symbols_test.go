package manifest

import (
	"math"
	"testing"
)

func symbolManifestFixture() Manifest {
	return Manifest{
		SchemaVersion: SchemaVersion,
		Target:        Target{MinecraftVersion: "1.26.51", ProtocolVersion: 2193},
		Sources:       []SourcePin{{ID: "wire", Kind: "synthetic", Revision: "fixture", Digest: "fixture:wire"}},
		Packets:       []Packet{{ID: 1, Name: "Fixture", Direction: DirectionClientbound}},
		Symbols: []Symbol{{ID: "symbols/ActorDataIDs", Name: "ActorDataIDs", Representation: "u32",
			Values:     []EnumValue{{Name: "RESERVED_0", Value: 0}, {Name: "ALIAS", Value: 0}},
			Source:     SourcePin{ID: "reference", Kind: "mojang-enum-catalog", Revision: "fixture", Digest: "fixture:reference", Locator: "https://example.invalid/catalog", MinecraftVersion: "1.26.40-beta.0", ProtocolVersion: 2168},
			Provenance: Provenance{Pins: []string{"reference"}},
		}},
	}
}

func TestReferenceCatalogCannotAuthorizePacketWireShape(t *testing.T) {
	m := symbolManifestFixture()
	if err := Validate(m); err != nil {
		t.Fatal(err)
	}
	m.Packets[0].Fields = []Field{{Ordinal: 0, Name: "ID", Encode: Primitive("var_u32"), Symmetry: Symmetric, Provenance: Provenance{Pins: []string{"reference"}}}}
	if err := Validate(m); err == nil {
		t.Fatal("older non-wire vocabulary was accepted as packet wire evidence")
	}
	m.Packets[0].Fields[0].Provenance.Pins[0] = "wire"
	m.Sources = append(m.Sources, m.Symbols[0].Source)
	if err := Validate(m); err == nil {
		t.Fatal("older catalog was accepted as a target-version wire source")
	}
}

func TestManifestRejectsInvalidSymbols(t *testing.T) {
	for name, mutate := range map[string]func(*Manifest){
		"representation":            func(m *Manifest) { m.Symbols[0].Representation = "var_u32" },
		"empty values":              func(m *Manifest) { m.Symbols[0].Values = nil },
		"duplicate groups":          func(m *Manifest) { m.Symbols = append(m.Symbols, m.Symbols[0]) },
		"duplicate names":           func(m *Manifest) { m.Symbols[0].Values[1].Name = "RESERVED_0" },
		"mismatched ID":             func(m *Manifest) { m.Symbols[0].ID = "ActorDataIDs" },
		"empty name":                func(m *Manifest) { m.Symbols[0].Name = "" },
		"invalid value name":        func(m *Manifest) { m.Symbols[0].Values[0].Name = "_" },
		"negative value":            func(m *Manifest) { m.Symbols[0].Values[0].Value = -1 },
		"overflow":                  func(m *Manifest) { m.Symbols[0].Values[0].Value = int64(math.MaxUint32) + 1 },
		"missing provenance":        func(m *Manifest) { m.Symbols[0].Provenance.Pins = nil },
		"unknown pin":               func(m *Manifest) { m.Symbols[0].Provenance.Pins[0] = "wire" },
		"missing reference version": func(m *Manifest) { m.Symbols[0].Source.ProtocolVersion = 0 },
	} {
		t.Run(name, func(t *testing.T) {
			m := symbolManifestFixture()
			mutate(&m)
			if err := Validate(m); err == nil {
				t.Fatal("accepted invalid non-wire symbols")
			}
		})
	}
}
