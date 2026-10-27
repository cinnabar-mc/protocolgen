// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

// AutoCraftRecipeStackRequestAction is sent by the client similarly to the CraftRecipeStackRequestAction. The
// only difference is that the recipe is automatically created and crafted by shift clicking the recipe book.
type AutoCraftRecipeStackRequestAction struct {
	ActionType              ItemStackRequestActionType
	RecipeNetID             RecipeNetID
	NumberOfRequestedCrafts uint8
	// Ingredients is a slice of ItemDescriptorCount that contains the ingredients that were used to craft the
	// recipe. It is not exactly clear what this is used for, but it is sent by the vanilla client.
	Ingredients []RecipeIngredient
}

func (*AutoCraftRecipeStackRequestAction) tagStackRequestAction() uint32 { return 11 }

// Marshal reads or writes AutoCraftRecipeStackRequestAction using its canonical wire layout.
func (x *AutoCraftRecipeStackRequestAction) Marshal(io IO) {
	x.ActionType.Marshal(io)
	x.RecipeNetID.Marshal(io)
	io.Uint8(&x.NumberOfRequestedCrafts)
	Minimum(io, &x.NumberOfRequestedCrafts, 1)
	Slice(io, &x.Ingredients)
}

// BeaconPaymentStackRequestAction is sent by the client when it submits an item to enable effects from a
// beacon. These items will have been moved into the beacon item slot in advance.
type BeaconPaymentStackRequestAction struct {
	ActionType ItemStackRequestActionType
	// PrimaryEffect and SecondaryEffect are the effects that were selected from the beacon.
	PrimaryEffectID int32
	// PrimaryEffect and SecondaryEffect are the effects that were selected from the beacon.
	SecondaryEffectID int32
}

func (*BeaconPaymentStackRequestAction) tagStackRequestAction() uint32 { return 8 }

// Marshal reads or writes BeaconPaymentStackRequestAction using its canonical wire layout.
func (x *BeaconPaymentStackRequestAction) Marshal(io IO) {
	x.ActionType.Marshal(io)
	io.Varint32(&x.PrimaryEffectID)
	Minimum(io, &x.PrimaryEffectID, 0)
	Maximum(io, &x.PrimaryEffectID, 38)
	io.Varint32(&x.SecondaryEffectID)
	Minimum(io, &x.SecondaryEffectID, 0)
	Maximum(io, &x.SecondaryEffectID, 38)
}

// ConsumeStackRequestAction is sent by the client when it uses an item to craft another item. The original
// item is 'consumed'.
type ConsumeStackRequestAction struct {
	ActionType ItemStackRequestActionType
	Amount     uint8
	Source     StackRequestSlotInfo
}

func (*ConsumeStackRequestAction) tagStackRequestAction() uint32 { return 5 }

// Marshal reads or writes ConsumeStackRequestAction using its canonical wire layout.
func (x *ConsumeStackRequestAction) Marshal(io IO) {
	x.ActionType.Marshal(io)
	io.Uint8(&x.Amount)
	Minimum(io, &x.Amount, 1)
	Maximum(io, &x.Amount, 64)
	x.Source.Marshal(io)
}

// CraftCreativeStackRequestAction is sent by the client when it takes an item out fo the creative inventory.
// The item is thus not really crafted, but instantly created.
type CraftCreativeStackRequestAction struct {
	ActionType              ItemStackRequestActionType
	CreativeItemNetID       uint32
	NumberOfRequestedCrafts uint8
}

func (*CraftCreativeStackRequestAction) tagStackRequestAction() uint32 { return 12 }

// Marshal reads or writes CraftCreativeStackRequestAction using its canonical wire layout.
func (x *CraftCreativeStackRequestAction) Marshal(io IO) {
	x.ActionType.Marshal(io)
	io.Varuint32(&x.CreativeItemNetID)
	Minimum(io, &x.CreativeItemNetID, 1)
	io.Uint8(&x.NumberOfRequestedCrafts)
	Minimum(io, &x.NumberOfRequestedCrafts, 1)
}

// CraftNonImplementedStackRequestAction is an action sent for inventory actions that aren't yet implemented
// in the new system. These include, for example, anvils.
type CraftNonImplementedStackRequestAction struct {
	ActionType ItemStackRequestActionType
}

func (*CraftNonImplementedStackRequestAction) tagStackRequestAction() uint32 { return 17 }

// Marshal reads or writes CraftNonImplementedStackRequestAction using its canonical wire layout.
func (x *CraftNonImplementedStackRequestAction) Marshal(io IO) {
	x.ActionType.Marshal(io)
}

// CraftRecipeOptionalStackRequestAction is sent when using an anvil. When this action is sent, the
// FilterStrings field in the respective stack request is non-empty and contains the name of the item created
// using the anvil or cartography table.
type CraftRecipeOptionalStackRequestAction struct {
	ActionType          ItemStackRequestActionType
	RecipeNetID         RecipeNetID
	FilteredStringIndex int32
}

func (*CraftRecipeOptionalStackRequestAction) tagStackRequestAction() uint32 { return 13 }

// Marshal reads or writes CraftRecipeOptionalStackRequestAction using its canonical wire layout.
func (x *CraftRecipeOptionalStackRequestAction) Marshal(io IO) {
	x.ActionType.Marshal(io)
	x.RecipeNetID.Marshal(io)
	io.Int32(&x.FilteredStringIndex)
}

// CraftRecipeStackRequestAction is sent by the client the moment it begins crafting an item. This is the
// first action sent, before the Consume and Create item stack request actions. This action is also sent when
// an item is enchanted. Enchanting should be treated mostly the same way as crafting, where the old item is
// consumed.
type CraftRecipeStackRequestAction struct {
	ActionType              ItemStackRequestActionType
	RecipeNetID             RecipeNetID
	NumberOfRequestedCrafts uint8
}

func (*CraftRecipeStackRequestAction) tagStackRequestAction() uint32 { return 10 }

// Marshal reads or writes CraftRecipeStackRequestAction using its canonical wire layout.
func (x *CraftRecipeStackRequestAction) Marshal(io IO) {
	x.ActionType.Marshal(io)
	x.RecipeNetID.Marshal(io)
	io.Uint8(&x.NumberOfRequestedCrafts)
	Minimum(io, &x.NumberOfRequestedCrafts, 1)
}

// CraftResultsDeprecatedStackRequestAction is an additional, deprecated packet sent by the client after
// crafting. It holds the final results and the amount of times the recipe was crafted. It shouldn't be used.
// This action is also sent when an item is enchanted. Enchanting should be treated mostly the same way as
// crafting, where the old item is consumed.
type CraftResultsDeprecatedStackRequestAction struct {
	ActionType   ItemStackRequestActionType
	CraftResults []ItemInstance
	NumCrafts    uint8
}

func (*CraftResultsDeprecatedStackRequestAction) tagStackRequestAction() uint32 { return 18 }

// Marshal reads or writes CraftResultsDeprecatedStackRequestAction using its canonical wire layout.
func (x *CraftResultsDeprecatedStackRequestAction) Marshal(io IO) {
	x.ActionType.Marshal(io)
	SliceLimits(io, &x.CraftResults, 1, 18446744073709551615)
	io.Uint8(&x.NumCrafts)
	Minimum(io, &x.NumCrafts, 1)
}

// CreateStackRequestAction is sent by the client when an item is created through being used as part of a
// recipe. For example, when milk is used to craft a cake, the buckets are leftover. The buckets are moved to
// the slot sent by the client here. Note that before this is sent, an action for consuming all items in the
// crafting table/grid is sent. Items that are not fully consumed when used for a recipe should not be
// destroyed there, but instead, should be turned into their respective resulting items.
type CreateStackRequestAction struct {
	ActionType   ItemStackRequestActionType
	ResultsIndex uint8
}

func (*CreateStackRequestAction) tagStackRequestAction() uint32 { return 6 }

// Marshal reads or writes CreateStackRequestAction using its canonical wire layout.
func (x *CreateStackRequestAction) Marshal(io IO) {
	x.ActionType.Marshal(io)
	io.Uint8(&x.ResultsIndex)
}

type CursorItemDragState uint8

const (
	CursorItemDragStateStart CursorItemDragState = 0
	CursorItemDragStateStop  CursorItemDragState = 1
)

// Marshal reads or writes CursorItemDragState through its uint8 wire encoding.
func (x *CursorItemDragState) Marshal(io IO) { io.Uint8((*uint8)(x)) }

// DestroyStackRequestAction is sent by the client when it destroys an item in creative mode by moving it back
// into the creative inventory.
type DestroyStackRequestAction struct {
	ActionType ItemStackRequestActionType
	Amount     uint8
	// Source is the source slot from which items came that were destroyed by moving them into the creative
	// inventory.
	Source StackRequestSlotInfo
}

func (*DestroyStackRequestAction) tagStackRequestAction() uint32 { return 4 }

// Marshal reads or writes DestroyStackRequestAction using its canonical wire layout.
func (x *DestroyStackRequestAction) Marshal(io IO) {
	x.ActionType.Marshal(io)
	io.Uint8(&x.Amount)
	Minimum(io, &x.Amount, 1)
	Maximum(io, &x.Amount, 64)
	x.Source.Marshal(io)
}

// DropStackRequestAction is sent by the client when it drops an item out of the inventory when it has its
// inventory opened. This action is not sent when a player drops an item out of the hotbar using the Q button
// (or the equivalent on mobile). The InventoryTransaction packet is still used for that action, regardless of
// whether the item stack network IDs are used or not.
type DropStackRequestAction struct {
	ActionType ItemStackRequestActionType
	Amount     uint8
	// Source is the source slot from which items were dropped to the ground.
	Source StackRequestSlotInfo
	// Randomly seems to be set to false in most cases. I'm not entirely sure what this does, but this is what
	// vanilla calls this field.
	Randomly bool
}

func (*DropStackRequestAction) tagStackRequestAction() uint32 { return 3 }

// Marshal reads or writes DropStackRequestAction using its canonical wire layout.
func (x *DropStackRequestAction) Marshal(io IO) {
	x.ActionType.Marshal(io)
	io.Uint8(&x.Amount)
	Minimum(io, &x.Amount, 1)
	Maximum(io, &x.Amount, 64)
	x.Source.Marshal(io)
	io.Bool(&x.Randomly)
}

type ItemDescriptor interface {
	Marshaler
	tagItemDescriptor() uint32
}

// MarshalItemDescriptor reads or writes the ItemDescriptor union using its canonical wire layout.
func MarshalItemDescriptor(io IO, x *ItemDescriptor) {
	Union(io, x, io.Varuint32, ItemDescriptor.tagItemDescriptor, func(tag uint32) ItemDescriptor {
		switch tag {
		case 0:
			return new(InvalidItemDescriptor)
		case 1:
			return new(DefaultItemDescriptor)
		case 2:
			return new(MoLangItemDescriptor)
		case 3:
			return new(ItemTagItemDescriptor)
		}
		return nil
	})
}

type ItemStackLegacyRequestID struct {
	ID int32
}

// Marshal reads or writes ItemStackLegacyRequestID using its canonical wire layout.
func (x *ItemStackLegacyRequestID) Marshal(io IO) {
	io.Varint32(&x.ID)
}

type ItemStackNetID struct {
	ID int32
}

// Marshal reads or writes ItemStackNetID using its canonical wire layout.
func (x *ItemStackNetID) Marshal(io IO) {
	io.Varint32(&x.ID)
}

type ItemStackNetResult uint8

const (
	ItemStackNetResultSuccess                                          ItemStackNetResult = 0
	ItemStackNetResultError                                            ItemStackNetResult = 1
	ItemStackNetResultInvalidRequestActionType                         ItemStackNetResult = 2
	ItemStackNetResultActionRequestNotAllowed                          ItemStackNetResult = 3
	ItemStackNetResultScreenHandlerEndRequestFailed                    ItemStackNetResult = 4
	ItemStackNetResultItemRequestActionHandlerCommitFailed             ItemStackNetResult = 5
	ItemStackNetResultInvalidRequestCraftActionType                    ItemStackNetResult = 6
	ItemStackNetResultInvalidCraftRequest                              ItemStackNetResult = 7
	ItemStackNetResultInvalidCraftRequestScreen                        ItemStackNetResult = 8
	ItemStackNetResultInvalidCraftResult                               ItemStackNetResult = 9
	ItemStackNetResultInvalidCraftResultIndex                          ItemStackNetResult = 10
	ItemStackNetResultInvalidCraftResultItem                           ItemStackNetResult = 11
	ItemStackNetResultInvalidItemNetID                                 ItemStackNetResult = 12
	ItemStackNetResultMissingCreatedOutputContainer                    ItemStackNetResult = 13
	ItemStackNetResultFailedToSetCreatedItemOutputSlot                 ItemStackNetResult = 14
	ItemStackNetResultRequestAlreadyInProgress                         ItemStackNetResult = 15
	ItemStackNetResultFailedToInitSparseContainer                      ItemStackNetResult = 16
	ItemStackNetResultResultTransferFailed                             ItemStackNetResult = 17
	ItemStackNetResultExpectedItemSlotNotFullyConsumed                 ItemStackNetResult = 18
	ItemStackNetResultExpectedAnywhereItemNotFullyConsumed             ItemStackNetResult = 19
	ItemStackNetResultItemAlreadyConsumedFromSlot                      ItemStackNetResult = 20
	ItemStackNetResultConsumedTooMuchFromSlot                          ItemStackNetResult = 21
	ItemStackNetResultMismatchSlotExpectedConsumedItem                 ItemStackNetResult = 22
	ItemStackNetResultMismatchSlotExpectedConsumedItemNetIDVariant     ItemStackNetResult = 23
	ItemStackNetResultFailedToMatchExpectedSlotConsumedItem            ItemStackNetResult = 24
	ItemStackNetResultFailedToMatchExpectedAllowedAnywhereConsumedItem ItemStackNetResult = 25
	ItemStackNetResultConsumedItemOutOfAllowedSlotRange                ItemStackNetResult = 26
	ItemStackNetResultConsumedItemNotAllowed                           ItemStackNetResult = 27
	ItemStackNetResultPlayerNotInCreativeMode                          ItemStackNetResult = 28
	ItemStackNetResultInvalidExperimentalRecipeRequest                 ItemStackNetResult = 29
	ItemStackNetResultFailedToCraftCreative                            ItemStackNetResult = 30
	ItemStackNetResultFailedToGetLevelRecipe                           ItemStackNetResult = 31
	ItemStackNetResultFailedToFindRecipeByNetID                        ItemStackNetResult = 32
	ItemStackNetResultMismatchedCraftingSize                           ItemStackNetResult = 33
	ItemStackNetResultMissingInputSparseContainer                      ItemStackNetResult = 34
	ItemStackNetResultMismatchedRecipeForInputGridItems                ItemStackNetResult = 35
	ItemStackNetResultEmptyCraftResults                                ItemStackNetResult = 36
	ItemStackNetResultFailedToEnchant                                  ItemStackNetResult = 37
	ItemStackNetResultMissingInputItem                                 ItemStackNetResult = 38
	ItemStackNetResultInsufficientPlayerLevelToEnchant                 ItemStackNetResult = 39
	ItemStackNetResultMissingMaterialItem                              ItemStackNetResult = 40
	ItemStackNetResultMissingActor                                     ItemStackNetResult = 41
	ItemStackNetResultUnknownPrimaryEffect                             ItemStackNetResult = 42
	ItemStackNetResultPrimaryEffectOutOfRange                          ItemStackNetResult = 43
	ItemStackNetResultPrimaryEffectUnavailable                         ItemStackNetResult = 44
	ItemStackNetResultSecondaryEffectOutOfRange                        ItemStackNetResult = 45
	ItemStackNetResultSecondaryEffectUnavailable                       ItemStackNetResult = 46
	ItemStackNetResultDstContainerEqualToCreatedOutputContainer        ItemStackNetResult = 47
	ItemStackNetResultDstContainerAndSlotEqualToSrcContainerAndSlot    ItemStackNetResult = 48
	ItemStackNetResultFailedToValidateSrcSlot                          ItemStackNetResult = 49
	ItemStackNetResultFailedToValidateDstSlot                          ItemStackNetResult = 50
	ItemStackNetResultInvalidAdjustedAmount                            ItemStackNetResult = 51
	ItemStackNetResultInvalidItemSetType                               ItemStackNetResult = 52
	ItemStackNetResultInvalidTransferAmount                            ItemStackNetResult = 53
	ItemStackNetResultCannotSwapItem                                   ItemStackNetResult = 54
	ItemStackNetResultCannotPlaceItem                                  ItemStackNetResult = 55
	ItemStackNetResultUnhandledItemSetType                             ItemStackNetResult = 56
	ItemStackNetResultInvalidRemovedAmount                             ItemStackNetResult = 57
	ItemStackNetResultInvalidRegion                                    ItemStackNetResult = 58
	ItemStackNetResultCannotDropItem                                   ItemStackNetResult = 59
	ItemStackNetResultCannotDestroyItem                                ItemStackNetResult = 60
	ItemStackNetResultInvalidSourceContainer                           ItemStackNetResult = 61
	ItemStackNetResultItemNotConsumed                                  ItemStackNetResult = 62
	ItemStackNetResultInvalidNumCrafts                                 ItemStackNetResult = 63
	ItemStackNetResultInvalidCraftResultStackSize                      ItemStackNetResult = 64
	ItemStackNetResultCannotRemoveItem                                 ItemStackNetResult = 65
	ItemStackNetResultCannotConsumeItem                                ItemStackNetResult = 66
	ItemStackNetResultScreenStackError                                 ItemStackNetResult = 67
)

// Marshal reads or writes ItemStackNetResult through its uint8 wire encoding.
func (x *ItemStackNetResult) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type ItemStackRequestActionType uint8

const (
	ItemStackRequestActionTypeTake                     ItemStackRequestActionType = 0
	ItemStackRequestActionTypePlace                    ItemStackRequestActionType = 1
	ItemStackRequestActionTypeSwap                     ItemStackRequestActionType = 2
	ItemStackRequestActionTypeDrop                     ItemStackRequestActionType = 3
	ItemStackRequestActionTypeDestroy                  ItemStackRequestActionType = 4
	ItemStackRequestActionTypeConsume                  ItemStackRequestActionType = 5
	ItemStackRequestActionTypeCreate                   ItemStackRequestActionType = 6
	ItemStackRequestActionTypePlaceInItemContainer     ItemStackRequestActionType = 7
	ItemStackRequestActionTypeTakeFromItemContainer    ItemStackRequestActionType = 8
	ItemStackRequestActionTypeScreenLabTableCombine    ItemStackRequestActionType = 9
	ItemStackRequestActionTypeScreenBeaconPayment      ItemStackRequestActionType = 10
	ItemStackRequestActionTypeScreenHUDMineBlock       ItemStackRequestActionType = 11
	ItemStackRequestActionTypeCraftRecipe              ItemStackRequestActionType = 12
	ItemStackRequestActionTypeCraftRecipeAuto          ItemStackRequestActionType = 13
	ItemStackRequestActionTypeCraftCreative            ItemStackRequestActionType = 14
	ItemStackRequestActionTypeCraftRecipeOptional      ItemStackRequestActionType = 15
	ItemStackRequestActionTypeCraftRepairAndDisenchant ItemStackRequestActionType = 16
	ItemStackRequestActionTypeCraftLoom                ItemStackRequestActionType = 17
	ItemStackRequestActionTypeCraftReservedAction      ItemStackRequestActionType = 18
	ItemStackRequestActionTypeCraftNonImplemented      ItemStackRequestActionType = 19
	ItemStackRequestActionTypeCraftResults             ItemStackRequestActionType = 20
)

// Marshal reads or writes ItemStackRequestActionType through its uint8 wire encoding.
func (x *ItemStackRequestActionType) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type ItemStackRequestCerealCraftDiscPressActionData struct {
	ActionType ItemStackRequestActionType
	ReservedID string
	NumCrafts  uint8
}

func (*ItemStackRequestCerealCraftDiscPressActionData) tagStackRequestAction() uint32 { return 16 }

// Marshal reads or writes ItemStackRequestCerealCraftDiscPressActionData using its canonical wire layout.
func (x *ItemStackRequestCerealCraftDiscPressActionData) Marshal(io IO) {
	x.ActionType.Marshal(io)
	io.StringLimits(&x.ReservedID, 0, 256)
	io.Uint8(&x.NumCrafts)
	Minimum(io, &x.NumCrafts, 1)
}

// ItemStackRequest represents a single request present in an ItemStackRequest packet sent by the client to
// change an item in an inventory. Item stack requests are either approved or rejected by the server using the
// ItemStackResponse packet.
type ItemStackRequestData struct {
	ClientRequestID ItemStackRequestID
	// Actions is a list of actions performed by the client. The actual type of the actions depends on which ID
	// was present, and is one of the concrete types below.
	Actions               []StackRequestAction
	StringsToFilter       []string
	StringsToFilterOrigin TextProcessingEventOrigin
}

// Marshal reads or writes ItemStackRequestData using its canonical wire layout.
func (x *ItemStackRequestData) Marshal(io IO) {
	x.ClientRequestID.Marshal(io)
	FuncSliceLimits(io, &x.Actions, io.Varuint32, 1, 100, func(value *StackRequestAction) {
		MarshalStackRequestAction(io, value)
	})
	FuncSlice(io, &x.StringsToFilter, io.Varuint32, func(value *string) {
		io.StringLimits(value, 0, 1000)
	})
	x.StringsToFilterOrigin.Marshal(io)
}

type ItemStackRequestID struct {
	ID int32
}

// Marshal reads or writes ItemStackRequestID using its canonical wire layout.
func (x *ItemStackRequestID) Marshal(io IO) {
	io.Varint32(&x.ID)
}

type ItemStackRequestPacketData struct {
	ClientRequestID       ItemStackRequestID
	Actions               []StackRequestAction
	StringsToFilter       []string
	StringsToFilterOrigin TextProcessingEventOrigin
}

// Marshal reads or writes ItemStackRequestPacketData using its canonical wire layout.
func (x *ItemStackRequestPacketData) Marshal(io IO) {
	x.ClientRequestID.Marshal(io)
	FuncSliceLimits(io, &x.Actions, io.Varuint32, 1, 100, func(value *StackRequestAction) {
		MarshalStackRequestAction(io, value)
	})
	FuncSlice(io, &x.StringsToFilter, io.Varuint32, func(value *string) {
		io.StringLimits(value, 0, 1000)
	})
	x.StringsToFilterOrigin.Marshal(io)
}

type ItemStackResponseContainerInfo struct {
	FullContainerName FullContainerName
	Slots             []ItemStackResponseSlotInfo
}

// Marshal reads or writes ItemStackResponseContainerInfo using its canonical wire layout.
func (x *ItemStackResponseContainerInfo) Marshal(io IO) {
	x.FullContainerName.Marshal(io)
	Slice(io, &x.Slots)
}

// ItemStackResponse is a response to an individual ItemStackRequest.
type ItemStackResponseInfo struct {
	Result          ItemStackNetResult
	ClientRequestID ItemStackRequestID
	Containers      Optional[[]ItemStackResponseContainerInfo]
}

// Marshal reads or writes ItemStackResponseInfo using its canonical wire layout.
func (x *ItemStackResponseInfo) Marshal(io IO) {
	x.Result.Marshal(io)
	x.ClientRequestID.Marshal(io)
	OptionalFunc(io, &x.Containers, func(value *[]ItemStackResponseContainerInfo) {
		Slice(io, value)
	})
}

type ItemStackResponseSlotInfo struct {
	RequestedSlot        uint8
	Slot                 uint8
	Amount               uint8
	ItemStackNetID       Optional[ItemStackNetID]
	CustomName           BedrockSafetyRedactableString
	DurabilityCorrection int32
}

// Marshal reads or writes ItemStackResponseSlotInfo using its canonical wire layout.
func (x *ItemStackResponseSlotInfo) Marshal(io IO) {
	io.Uint8(&x.RequestedSlot)
	io.Uint8(&x.Slot)
	io.Uint8(&x.Amount)
	OptionalMarshaler(io, &x.ItemStackNetID)
	x.CustomName.Marshal(io)
	io.Varint32(&x.DurabilityCorrection)
	Minimum(io, &x.DurabilityCorrection, -32768)
	Maximum(io, &x.DurabilityCorrection, 32767)
}

// LabTableCombineStackRequestAction is sent by the client when it uses a lab table to combine item stacks.
type LabTableCombineStackRequestAction struct {
	ActionType ItemStackRequestActionType
}

func (*LabTableCombineStackRequestAction) tagStackRequestAction() uint32 { return 7 }

// Marshal reads or writes LabTableCombineStackRequestAction using its canonical wire layout.
func (x *LabTableCombineStackRequestAction) Marshal(io IO) {
	x.ActionType.Marshal(io)
}

// MineBlockStackRequestAction is sent by the client when it breaks a block.
type MineBlockStackRequestAction struct {
	ActionType ItemStackRequestActionType
	// HotbarSlot is the slot held by the player while mining a block.
	Slot int32
	// PredictedDurability is the durability of the item that the client assumes to be present at the time.
	PredictedDurability int32
	NetIDVariant        int32
}

func (*MineBlockStackRequestAction) tagStackRequestAction() uint32 { return 9 }

// Marshal reads or writes MineBlockStackRequestAction using its canonical wire layout.
func (x *MineBlockStackRequestAction) Marshal(io IO) {
	x.ActionType.Marshal(io)
	io.Varint32(&x.Slot)
	io.Varint32(&x.PredictedDurability)
	io.Int32(&x.NetIDVariant)
}

// PlaceStackRequestAction is sent by the client to the server to place x amount of items from one slot into
// another slot, such as when shift clicking an item in the inventory to move it around or when moving an item
// in the cursor into a slot.
type PlaceStackRequestAction struct {
	ActionType  ItemStackRequestActionType
	Amount      uint8
	Source      StackRequestSlotInfo
	Destination StackRequestSlotInfo
}

func (*PlaceStackRequestAction) tagStackRequestAction() uint32 { return 1 }

// Marshal reads or writes PlaceStackRequestAction using its canonical wire layout.
func (x *PlaceStackRequestAction) Marshal(io IO) {
	x.ActionType.Marshal(io)
	io.Uint8(&x.Amount)
	Minimum(io, &x.Amount, 1)
	Maximum(io, &x.Amount, 64)
	x.Source.Marshal(io)
	x.Destination.Marshal(io)
}

type StackRequestAction interface {
	Marshaler
	tagStackRequestAction() uint32
}

// MarshalStackRequestAction reads or writes the StackRequestAction union using its canonical wire layout.
func MarshalStackRequestAction(io IO, x *StackRequestAction) {
	Union(io, x, io.Varuint32, StackRequestAction.tagStackRequestAction, func(tag uint32) StackRequestAction {
		switch tag {
		case 0:
			return new(TakeStackRequestAction)
		case 1:
			return new(PlaceStackRequestAction)
		case 2:
			return new(SwapStackRequestAction)
		case 3:
			return new(DropStackRequestAction)
		case 4:
			return new(DestroyStackRequestAction)
		case 5:
			return new(ConsumeStackRequestAction)
		case 6:
			return new(CreateStackRequestAction)
		case 7:
			return new(LabTableCombineStackRequestAction)
		case 8:
			return new(BeaconPaymentStackRequestAction)
		case 9:
			return new(MineBlockStackRequestAction)
		case 10:
			return new(CraftRecipeStackRequestAction)
		case 11:
			return new(AutoCraftRecipeStackRequestAction)
		case 12:
			return new(CraftCreativeStackRequestAction)
		case 13:
			return new(CraftRecipeOptionalStackRequestAction)
		case 14:
			return new(CraftRepairAndDisenchantStackRequestAction)
		case 15:
			return new(CraftLoomStackRequestAction)
		case 16:
			return new(ItemStackRequestCerealCraftDiscPressActionData)
		case 17:
			return new(CraftNonImplementedStackRequestAction)
		case 18:
			return new(CraftResultsDeprecatedStackRequestAction)
		}
		return nil
	})
}

// StackRequestSlotInfo holds information on a specific slot client-side.
type StackRequestSlotInfo struct {
	// Container is the FullContainerName that describes the container that the slot is in.
	FullContainerName FullContainerName
	// Slot is the index of the slot within the container with the ContainerID above.
	Slot uint8
	// StackNetworkID is the unique stack ID that the client assumes to be present in this slot. The server must
	// check if these IDs match. If they do not match, servers should reject the stack request that the action
	// holding this info was in.
	NetIDVariant int32
}

// Marshal reads or writes StackRequestSlotInfo using its canonical wire layout.
func (x *StackRequestSlotInfo) Marshal(io IO) {
	x.FullContainerName.Marshal(io)
	io.Uint8(&x.Slot)
	io.Int32(&x.NetIDVariant)
}

// SwapStackRequestAction is sent by the client to swap the item in its cursor with an item present in another
// container. The two item stacks swap places.
type SwapStackRequestAction struct {
	ActionType ItemStackRequestActionType
	// Source and Destination point to the source slot from which Count of the item stack were taken and the
	// destination slot to which this item was moved.
	Source StackRequestSlotInfo
	// Source and Destination point to the source slot from which Count of the item stack were taken and the
	// destination slot to which this item was moved.
	Destination StackRequestSlotInfo
}

func (*SwapStackRequestAction) tagStackRequestAction() uint32 { return 2 }

// Marshal reads or writes SwapStackRequestAction using its canonical wire layout.
func (x *SwapStackRequestAction) Marshal(io IO) {
	x.ActionType.Marshal(io)
	x.Source.Marshal(io)
	x.Destination.Marshal(io)
}

// TakeStackRequestAction is sent by the client to the server to take x amount of items from one slot in a
// container to the cursor.
type TakeStackRequestAction struct {
	ActionType  ItemStackRequestActionType
	Amount      uint8
	Source      StackRequestSlotInfo
	Destination StackRequestSlotInfo
}

func (*TakeStackRequestAction) tagStackRequestAction() uint32 { return 0 }

// Marshal reads or writes TakeStackRequestAction using its canonical wire layout.
func (x *TakeStackRequestAction) Marshal(io IO) {
	x.ActionType.Marshal(io)
	io.Uint8(&x.Amount)
	Minimum(io, &x.Amount, 1)
	Maximum(io, &x.Amount, 64)
	x.Source.Marshal(io)
	x.Destination.Marshal(io)
}
