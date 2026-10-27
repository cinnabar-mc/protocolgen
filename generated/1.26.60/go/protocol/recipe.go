// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

import (
	"github.com/google/uuid"
)

// MultiRecipe serves as an 'enable' switch for multi-shape recipes.
type MultiRecipe struct {
	// UUID is a UUID identifying the recipe. Since the CraftingEvent packet no longer exists, this can always be
	// empty.
	MultiRecipeUUID uuid.UUID
	NetID           RecipeNetID
}

// Marshal reads or writes MultiRecipe using its canonical wire layout.
func (x *MultiRecipe) Marshal(io IO) {
	io.UUID(&x.MultiRecipeUUID)
	x.NetID.Marshal(io)
}

type RecipeIngredient struct {
	ItemDescriptor ItemDescriptor
	StackSize      uint16
}

// Marshal reads or writes RecipeIngredient using its canonical wire layout.
func (x *RecipeIngredient) Marshal(io IO) {
	MarshalItemDescriptor(io, &x.ItemDescriptor)
	io.Uint16(&x.StackSize)
	Minimum(io, &x.StackSize, 1)
}

type RecipeIngredientSerializedData struct {
	Descriptor []OrderedEntry[string, string]
	AuxValue   int32
	StackSize  int32
}

// Marshal reads or writes RecipeIngredientSerializedData using its canonical wire layout.
func (x *RecipeIngredientSerializedData) Marshal(io IO) {
	OrderedMap(io, &x.Descriptor, io.Varuint32, io.String, io.String)
	io.Varint32(&x.AuxValue)
	io.Varint32(&x.StackSize)
	Minimum(io, &x.StackSize, 0)
	Maximum(io, &x.StackSize, 64)
}

type RecipeNetID struct {
	RawID uint32
}

// Marshal reads or writes RecipeNetID using its canonical wire layout.
func (x *RecipeNetID) Marshal(io IO) {
	io.Varuint32(&x.RawID)
}

type RecipeUnlockRequirementSerializedData struct {
	UnlockingContext     RecipeUnlockingRequirementUnlockingContext
	UnlockingIngredients Optional[[]RecipeIngredientSerializedData]
}

// Marshal reads or writes RecipeUnlockRequirementSerializedData using its canonical wire layout.
func (x *RecipeUnlockRequirementSerializedData) Marshal(io IO) {
	x.UnlockingContext.Marshal(io)
	OptionalFunc(io, &x.UnlockingIngredients, func(value *[]RecipeIngredientSerializedData) {
		SliceLimits(io, value, 0, 128)
	})
}

type RecipeUnlockingRequirementUnlockingContext int32

const (
	RecipeUnlockingRequirementUnlockingContextNone               RecipeUnlockingRequirementUnlockingContext = 0
	RecipeUnlockingRequirementUnlockingContextAlwaysUnlocked     RecipeUnlockingRequirementUnlockingContext = 1
	RecipeUnlockingRequirementUnlockingContextPlayerInWater      RecipeUnlockingRequirementUnlockingContext = 2
	RecipeUnlockingRequirementUnlockingContextPlayerHasManyItems RecipeUnlockingRequirementUnlockingContext = 3
)

// Marshal reads or writes RecipeUnlockingRequirementUnlockingContext through its int32 wire encoding.
func (x *RecipeUnlockingRequirementUnlockingContext) Marshal(io IO) { io.Varint32((*int32)(x)) }

// ShapedRecipe is a recipe that has a specific shape that must be used to craft the output of the recipe.
// Trying to craft the item in any other shape will not work. The ShapedRecipe is of the same structure as the
// ShapedChemistryRecipe.
type ShapedRecipe struct {
	// RecipeID is a unique ID of the recipe. This ID must be unique amongst all other types of recipes too, but
	// its functionality is not exactly known.
	RecipeID string
	// Width is the width of the recipe's shape.
	Width int32
	// Height is the height of the recipe's shape.
	Height      int32
	Ingredients []RecipeIngredientSerializedData
	Results     []NetworkItemInstanceDescriptorSerializedData
	// UUID is a UUID identifying the recipe. Since the CraftingEvent packet no longer exists, this can always be
	// empty.
	UUID uuid.UUID
	Tag  string
	// Priority ...
	Priority int32
	// AssumeSymmetry specifies if the recipe is symmetrical. If this is set to true, the recipe will be mirrored
	// along the diagonal axis. This means that the recipe will be the same if rotated 180 degrees.
	AssumeSymmetry       bool
	UnlockingRequirement Optional[RecipeUnlockRequirementSerializedData]
	NetID                RecipeNetID
}

// Marshal reads or writes ShapedRecipe using its canonical wire layout.
func (x *ShapedRecipe) Marshal(io IO) {
	io.String(&x.RecipeID)
	io.Varint32(&x.Width)
	io.Varint32(&x.Height)
	SliceLimits(io, &x.Ingredients, 0, 128)
	Slice(io, &x.Results)
	io.UUID(&x.UUID)
	io.String(&x.Tag)
	io.Varint32(&x.Priority)
	io.Bool(&x.AssumeSymmetry)
	OptionalMarshaler(io, &x.UnlockingRequirement)
	x.NetID.Marshal(io)
}

// ShapelessRecipe is a recipe that has no particular shape. Its functionality is shared with the
// RecipeShulkerBox and RecipeShapelessChemistry types.
type ShapelessRecipe struct {
	// RecipeID is a unique ID of the recipe. This ID must be unique amongst all other types of recipes too, but
	// its functionality is not exactly known.
	RecipeID    string
	Ingredients []RecipeIngredientSerializedData
	Results     []NetworkItemInstanceDescriptorSerializedData
	// UUID is a UUID identifying the recipe. Since the CraftingEvent packet no longer exists, this can always be
	// empty.
	UUID uuid.UUID
	Tag  string
	// Priority ...
	Priority             int32
	UnlockingRequirement Optional[RecipeUnlockRequirementSerializedData]
	NetID                RecipeNetID
}

// Marshal reads or writes ShapelessRecipe using its canonical wire layout.
func (x *ShapelessRecipe) Marshal(io IO) {
	io.String(&x.RecipeID)
	SliceLimits(io, &x.Ingredients, 0, 128)
	Slice(io, &x.Results)
	io.UUID(&x.UUID)
	io.String(&x.Tag)
	io.Varint32(&x.Priority)
	OptionalMarshaler(io, &x.UnlockingRequirement)
	x.NetID.Marshal(io)
}

// SmithingTransformRecipe is a recipe specifically used for smithing tables. It has three input items and
// adds them together, resulting in a new item.
type SmithingTransformRecipe struct {
	// RecipeID is a unique ID of the recipe. This ID must be unique amongst all other types of recipes too, but
	// its functionality is not exactly known.
	RecipeID           string
	TemplateIngredient RecipeIngredientSerializedData
	BaseIngredient     RecipeIngredientSerializedData
	AdditionIngredient RecipeIngredientSerializedData
	// Result is the resulting item from the two items being added together.
	Result NetworkItemInstanceDescriptorSerializedData
	Tag    string
	NetID  RecipeNetID
}

// Marshal reads or writes SmithingTransformRecipe using its canonical wire layout.
func (x *SmithingTransformRecipe) Marshal(io IO) {
	io.String(&x.RecipeID)
	x.TemplateIngredient.Marshal(io)
	x.BaseIngredient.Marshal(io)
	x.AdditionIngredient.Marshal(io)
	x.Result.Marshal(io)
	io.String(&x.Tag)
	x.NetID.Marshal(io)
}

// SmithingTrimRecipe is a recipe specifically used for applying armour trims to an armour piece inside a
// smithing table.
type SmithingTrimRecipe struct {
	// RecipeID is a unique ID of the recipe. This ID must be unique amongst all other types of recipes too, but
	// its functionality is not exactly known.
	RecipeID           string
	TemplateIngredient RecipeIngredientSerializedData
	BaseIngredient     RecipeIngredientSerializedData
	AdditionIngredient RecipeIngredientSerializedData
	Tag                string
	NetID              RecipeNetID
}

// Marshal reads or writes SmithingTrimRecipe using its canonical wire layout.
func (x *SmithingTrimRecipe) Marshal(io IO) {
	io.String(&x.RecipeID)
	x.TemplateIngredient.Marshal(io)
	x.BaseIngredient.Marshal(io)
	x.AdditionIngredient.Marshal(io)
	io.String(&x.Tag)
	x.NetID.Marshal(io)
}
