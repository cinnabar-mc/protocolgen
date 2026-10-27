// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

type ContainerEnumName uint8

const (
	ContainerEnumNameAnvilInputContainer                 ContainerEnumName = 0
	ContainerEnumNameAnvilMaterialContainer              ContainerEnumName = 1
	ContainerEnumNameAnvilResultPreviewContainer         ContainerEnumName = 2
	ContainerEnumNameSmithingTableInputContainer         ContainerEnumName = 3
	ContainerEnumNameSmithingTableMaterialContainer      ContainerEnumName = 4
	ContainerEnumNameSmithingTableResultPreviewContainer ContainerEnumName = 5
	ContainerEnumNameArmorContainer                      ContainerEnumName = 6
	ContainerEnumNameLevelEntityContainer                ContainerEnumName = 7
	ContainerEnumNameBeaconPaymentContainer              ContainerEnumName = 8
	ContainerEnumNameBrewingStandInputContainer          ContainerEnumName = 9
	ContainerEnumNameBrewingStandResultContainer         ContainerEnumName = 10
	ContainerEnumNameBrewingStandFuelContainer           ContainerEnumName = 11
	ContainerEnumNameCombinedHotbarAndInventoryContainer ContainerEnumName = 12
	ContainerEnumNameCraftingInputContainer              ContainerEnumName = 13
	ContainerEnumNameCraftingOutputPreviewContainer      ContainerEnumName = 14
	ContainerEnumNameRecipeConstructionContainer         ContainerEnumName = 15
	ContainerEnumNameRecipeNatureContainer               ContainerEnumName = 16
	ContainerEnumNameRecipeItemsContainer                ContainerEnumName = 17
	ContainerEnumNameRecipeSearchContainer               ContainerEnumName = 18
	ContainerEnumNameRecipeSearchBarContainer            ContainerEnumName = 19
	ContainerEnumNameRecipeEquipmentContainer            ContainerEnumName = 20
	ContainerEnumNameRecipeBookContainer                 ContainerEnumName = 21
	ContainerEnumNameEnchantingInputContainer            ContainerEnumName = 22
	ContainerEnumNameEnchantingMaterialContainer         ContainerEnumName = 23
	ContainerEnumNameFurnaceFuelContainer                ContainerEnumName = 24
	ContainerEnumNameFurnaceIngredientContainer          ContainerEnumName = 25
	ContainerEnumNameFurnaceResultContainer              ContainerEnumName = 26
	ContainerEnumNameHorseEquipContainer                 ContainerEnumName = 27
	ContainerEnumNameHotbarContainer                     ContainerEnumName = 28
	ContainerEnumNameInventoryContainer                  ContainerEnumName = 29
	ContainerEnumNameShulkerBoxContainer                 ContainerEnumName = 30
	ContainerEnumNameTradeIngredient1Container           ContainerEnumName = 31
	ContainerEnumNameTradeIngredient2Container           ContainerEnumName = 32
	ContainerEnumNameTradeResultPreviewContainer         ContainerEnumName = 33
	ContainerEnumNameOffhandContainer                    ContainerEnumName = 34
	ContainerEnumNameCompoundCreatorInput                ContainerEnumName = 35
	ContainerEnumNameCompoundCreatorOutputPreview        ContainerEnumName = 36
	ContainerEnumNameElementConstructorOutputPreview     ContainerEnumName = 37
	ContainerEnumNameMaterialReducerInput                ContainerEnumName = 38
	ContainerEnumNameMaterialReducerOutput               ContainerEnumName = 39
	ContainerEnumNameLabTableInput                       ContainerEnumName = 40
	ContainerEnumNameLoomInputContainer                  ContainerEnumName = 41
	ContainerEnumNameLoomDyeContainer                    ContainerEnumName = 42
	ContainerEnumNameLoomMaterialContainer               ContainerEnumName = 43
	ContainerEnumNameLoomResultPreviewContainer          ContainerEnumName = 44
	ContainerEnumNameBlastFurnaceIngredientContainer     ContainerEnumName = 45
	ContainerEnumNameSmokerIngredientContainer           ContainerEnumName = 46
	ContainerEnumNameTrade2Ingredient1Container          ContainerEnumName = 47
	ContainerEnumNameTrade2Ingredient2Container          ContainerEnumName = 48
	ContainerEnumNameTrade2ResultPreviewContainer        ContainerEnumName = 49
	ContainerEnumNameGrindstoneInputContainer            ContainerEnumName = 50
	ContainerEnumNameGrindstoneAdditionalContainer       ContainerEnumName = 51
	ContainerEnumNameGrindstoneResultPreviewContainer    ContainerEnumName = 52
	ContainerEnumNameStonecutterInputContainer           ContainerEnumName = 53
	ContainerEnumNameStonecutterResultPreviewContainer   ContainerEnumName = 54
	ContainerEnumNameCartographyInputContainer           ContainerEnumName = 55
	ContainerEnumNameCartographyAdditionalContainer      ContainerEnumName = 56
	ContainerEnumNameCartographyResultPreviewContainer   ContainerEnumName = 57
	ContainerEnumNameBarrelContainer                     ContainerEnumName = 58
	ContainerEnumNameCursorContainer                     ContainerEnumName = 59
	ContainerEnumNameCreatedOutputContainer              ContainerEnumName = 60
	ContainerEnumNameSmithingTableTemplateContainer      ContainerEnumName = 61
	ContainerEnumNameCrafterLevelEntityContainer         ContainerEnumName = 62
	ContainerEnumNameDynamicContainer                    ContainerEnumName = 63
	ContainerEnumNameRecipeFoodContainer                 ContainerEnumName = 64
	ContainerEnumNameRecipeBlocksContainer               ContainerEnumName = 65
	ContainerEnumNameRecipeFurnaceItemsContainer         ContainerEnumName = 66
	ContainerEnumNameReservedContainerA                  ContainerEnumName = 67
	ContainerEnumNameReservedContainerB                  ContainerEnumName = 68
	ContainerEnumNameReservedContainerC                  ContainerEnumName = 69
	ContainerEnumNameReservedContainerD                  ContainerEnumName = 70
)

// Marshal reads or writes ContainerEnumName through its uint8 wire encoding.
func (x *ContainerEnumName) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type ContainerMixDataEntry struct {
	FromItemID    int32
	ReagentItemID int32
	ToItemID      int32
}

// Marshal reads or writes ContainerMixDataEntry using its canonical wire layout.
func (x *ContainerMixDataEntry) Marshal(io IO) {
	io.Varint32(&x.FromItemID)
	io.Varint32(&x.ReagentItemID)
	io.Varint32(&x.ToItemID)
}

// FullContainerName contains information required to identify a container in a StackRequestSlotInfo.
type FullContainerName struct {
	// ContainerName is the ID of the container that the slot was in.
	ContainerName ContainerEnumName
	// DynamicID is the ID of the container if it is dynamic. If the container is not dynamic, this field should
	// be left empty. A non-optional value of 0 is assumed to be non-empty.
	DynamicID Optional[uint32]
}

// Marshal reads or writes FullContainerName using its canonical wire layout.
func (x *FullContainerName) Marshal(io IO) {
	x.ContainerName.Marshal(io)
	OptionalFunc(io, &x.DynamicID, io.Uint32)
}
