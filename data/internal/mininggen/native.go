package mininggen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"slices"
	"sort"
	"strings"

	"github.com/bedrock-mc/protocolgen/data/internal/nativeblock"
)

// Generate emits every native block, including verified no-tool requirements.
// Older special-tool evidence is limited to speed classifications that the
// native tag dump does not measure; it never decides harvest eligibility.
func Generate(native nativeblock.Catalog, special []byte, nativeVersion, nativeRevision, specialRevision, specialSHA256, sourceLockSHA256 string) ([]byte, int, error) {
	if len(native.Properties) == 0 || nativeVersion == "" || nativeRevision == "" || specialRevision == "" || specialSHA256 == "" || sourceLockSHA256 == "" {
		return nil, 0, fmt.Errorf("mining inputs or source identity is incomplete")
	}
	var old []struct {
		Name     string `json:"name"`
		Material string `json:"material"`
	}
	if err := json.Unmarshal(special, &old); err != nil {
		return nil, 0, fmt.Errorf("decode special-tool evidence: %w", err)
	}
	if len(old) == 0 {
		return nil, 0, fmt.Errorf("special-tool evidence is empty")
	}
	specialTools := make(map[string]uint8, len(old))
	for _, row := range old {
		name := "minecraft:" + row.Name
		if _, duplicate := specialTools[name]; row.Name == "" || duplicate {
			return nil, 0, fmt.Errorf("special-tool evidence has an empty or duplicate name %q", row.Name)
		}
		var tools uint8
		for _, material := range strings.Split(row.Material, ";") {
			switch material {
			case "leaves", "wool":
				tools |= shears
			case "coweb":
				tools |= sword | shears
			case "sword_instantly_mines":
				tools |= sword
			}
		}
		specialTools[name] = tools
	}
	names := make([]string, 0, len(native.Properties))
	for name := range native.Properties {
		names = append(names, name)
	}
	sort.Strings(names)
	var output bytes.Buffer
	fmt.Fprintf(&output, `%s

package block

// MiningEvidenceVersion identifies the native block-tag snapshot.
const MiningEvidenceVersion = %q

// MiningEvidenceRevision pins the native measurements.
const MiningEvidenceRevision = %q

// MiningSpecialRevision pins the retained special-tool speed evidence.
const MiningSpecialRevision = %q

// MiningSpecialSHA256 authenticates the retained special-tool source.
const MiningSpecialSHA256 = %q

// MiningSourceLockSHA256 identifies every input used for these facts.
const MiningSourceLockSHA256 = %q

var miningRows = [...]miningRow{
`, GeneratedMarker, nativeVersion, nativeRevision, specialRevision, specialSHA256, sourceLockSHA256)
	for _, name := range names {
		fact, err := classify(native.Properties[name], native.Tags[name], specialTools[name])
		if err != nil {
			return nil, 0, fmt.Errorf("native block %s: %w", name, err)
		}
		fmt.Fprintf(&output, "{Name:%q, Facts:MiningFacts{EffectiveTools:%s, HarvestTools:%s, HarvestLevel:%d}},\n", name, toolExpression(fact.effective), toolExpression(fact.harvest), fact.level)
	}
	output.WriteString("}\n")
	formatted, err := format.Source(output.Bytes())
	return formatted, len(names), err
}

// miningFact keeps native harvest eligibility separate from special-tool speed evidence.
type miningFact struct {
	effective, harvest uint8
	level              int
}

// classify resolves native tool tags and harvest tiers without guessing from block names.
func classify(properties nativeblock.Properties, members []string, special uint8) (miningFact, error) {
	if properties.RequiresCorrectToolForDrops == nil {
		return miningFact{}, fmt.Errorf("missing tool requirement")
	}
	var effective, harvest uint8
	for _, family := range []struct {
		tag  string
		mask uint8
	}{
		{"minecraft:is_pickaxe_item_destructible", pickaxe},
		{"minecraft:is_axe_item_destructible", axe},
		{"minecraft:is_shovel_item_destructible", shovel},
		{"minecraft:is_hoe_item_destructible", hoe},
		{"minecraft:is_sword_item_destructible", sword},
		{"minecraft:is_shears_item_destructible", shears},
	} {
		if slices.Contains(members, family.tag) {
			harvest |= family.mask
			if family.mask&(sword|shears) == 0 {
				effective |= family.mask
			}
		}
	}
	effective |= special
	level := -1
	if *properties.RequiresCorrectToolForDrops {
		level = 0
		for i, tag := range []string{"minecraft:stone_tier_destructible", "minecraft:iron_tier_destructible", "minecraft:diamond_tier_destructible", "minecraft:netherite_tier_destructible"} {
			if slices.Contains(members, tag) {
				level = i + 1
				break
			}
		}
	} else {
		harvest = 0
	}
	return miningFact{effective, harvest, level}, nil
}
