// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

type FurnaceLayout int32

const (
	FurnaceLayoutNone          FurnaceLayout = 0
	FurnaceLayoutInventoryOnly FurnaceLayout = 1
	FurnaceLayoutDefault       FurnaceLayout = 2
)

// Marshal reads or writes FurnaceLayout through its int32 wire encoding.
func (x *FurnaceLayout) Marshal(io IO) { io.Varint32((*int32)(x)) }

type FurnaceLeftTabIndex int32

const (
	FurnaceLeftTabIndexNone         FurnaceLeftTabIndex = 0
	FurnaceLeftTabIndexRecipeFood   FurnaceLeftTabIndex = 1
	FurnaceLeftTabIndexRecipeItems  FurnaceLeftTabIndex = 2
	FurnaceLeftTabIndexRecipeBlocks FurnaceLeftTabIndex = 3
	FurnaceLeftTabIndexRecipeSearch FurnaceLeftTabIndex = 4
	FurnaceLeftTabIndexInventory    FurnaceLeftTabIndex = 5
)

// Marshal reads or writes FurnaceLeftTabIndex through its int32 wire encoding.
func (x *FurnaceLeftTabIndex) Marshal(io IO) { io.Varint32((*int32)(x)) }

// FurnaceOptions holds the options that a player has selected in a furnace-like container's UI.
type FurnaceOptions struct {
	// LeftFurnaceTab is the tab that is selected on the left side of the furnace UI. It is one of the
	// FurnaceLeftTab constants above.
	LeftFurnaceTab FurnaceLeftTabIndex
	// Filtering is whether the player has enabled the filtering between recipes they have unlocked or not.
	Filtering bool
	// Layout is the layout of the furnace UI. It is one of the FurnaceLayout constants above.
	Layout FurnaceLayout
}

// Marshal reads or writes FurnaceOptions using its canonical wire layout.
func (x *FurnaceOptions) Marshal(io IO) {
	x.LeftFurnaceTab.Marshal(io)
	io.Bool(&x.Filtering)
	x.Layout.Marshal(io)
}

type FurnaceType uint8

const (
	FurnaceTypeNone         FurnaceType = 0
	FurnaceTypeFurnace      FurnaceType = 1
	FurnaceTypeBlastFurnace FurnaceType = 2
	FurnaceTypeSmoker       FurnaceType = 3
)

// Marshal reads or writes FurnaceType through its uint8 wire encoding.
func (x *FurnaceType) Marshal(io IO) { io.Uint8((*uint8)(x)) }
