// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

// TrimMaterial represents a material that can be used when applying an armour trim.
type TrimMaterial struct {
	// MaterialID is the identifier of the material, for example 'netherite'.
	MaterialID string
	// Color is the colour code used for text formatting, for example '§j'.
	Color string
	// ItemName is the identifier of the item that represents the material, for example,
	// 'minecraft:netherite_ingot'.
	ItemName string
}

// Marshal reads or writes TrimMaterial using its canonical wire layout.
func (x *TrimMaterial) Marshal(io IO) {
	io.String(&x.MaterialID)
	io.String(&x.Color)
	io.String(&x.ItemName)
}

// TrimPattern represents a pattern that can be applied to an armour piece in combination with a TrimMaterial.
type TrimPattern struct {
	// ItemName is the identifier of the item that represents the pattern, for example
	// 'minecraft:wayfinder_armor_trim_smithing_template'.
	ItemName string
	// PatternID is the identifier of the pattern, for example, 'wayfinder'.
	PatternID string
}

// Marshal reads or writes TrimPattern using its canonical wire layout.
func (x *TrimPattern) Marshal(io IO) {
	io.String(&x.ItemName)
	io.String(&x.PatternID)
}
