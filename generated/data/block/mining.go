package block

import "sort"

// ToolFamilies is a set of tools that can mine a block or qualify for its drops.
type ToolFamilies uint8

const (
	ToolPickaxe ToolFamilies = 1 << iota
	ToolAxe
	ToolShovel
	ToolHoe
	ToolSword
	ToolShears
)

// Has reports whether every family in tools is in the set.
func (families ToolFamilies) Has(tools ToolFamilies) bool {
	return families&tools == tools
}

// MiningEvidenceStatus describes how strongly a mining classification is known.
type MiningEvidenceStatus uint8

const (
	// MiningUnknown means the pinned source provided no usable classification.
	MiningUnknown MiningEvidenceStatus = iota
	// MiningProvisional means some speed classifications use an older Bedrock release.
	MiningProvisional
	// MiningNative means ordinary tools and harvest requirements use the target native dump.
	MiningNative
)

// MiningFacts combines native harvest requirements and ordinary-tool speed tags.
// Sword and shears speed targets retain older compatibility evidence; their
// complete native speed rules are not represented by the block tag dump.
// HarvestLevel is -1 when a correct tool is not required for drops.
type MiningFacts struct {
	EffectiveTools  ToolFamilies
	HarvestTools    ToolFamilies
	HarvestLevel    int8
	Evidence        MiningEvidenceStatus
	SpecialEvidence MiningEvidenceStatus
}

// MiningByName finds facts for a native block identifier. A false result means
// the identifier has no measured harvest requirement, including modded blocks.
func MiningByName(name string) (MiningFacts, bool) {
	index := sort.Search(len(miningRows), func(i int) bool { return miningRows[i].Name >= name })
	if index == len(miningRows) || miningRows[index].Name != name {
		return MiningFacts{}, false
	}
	facts := miningRows[index].Facts
	facts.Evidence = MiningNative
	facts.SpecialEvidence = MiningProvisional
	return facts, true
}

type miningRow struct {
	Name  string
	Facts MiningFacts
}
