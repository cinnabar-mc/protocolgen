// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

import (
	"github.com/go-gl/mathgl/mgl32"
)

type ItemData struct {
	ItemName          string
	ItemID            int16
	IsComponentBased  bool
	ItemVersion       ItemVersion
	ItemComponentData []byte
}

// Marshal reads or writes ItemData using its canonical wire layout.
func (x *ItemData) Marshal(io IO) {
	io.String(&x.ItemName)
	io.Int16(&x.ItemID)
	io.Bool(&x.IsComponentBased)
	x.ItemVersion.Marshal(io)
	io.NBT(&x.ItemComponentData, NBTNetwork)
}

type ItemEnchantOption struct {
	Cost         uint8
	Enchants     ItemEnchants
	EnchantName  string
	EnchantNetID RecipeNetID
}

// Marshal reads or writes ItemEnchantOption using its canonical wire layout.
func (x *ItemEnchantOption) Marshal(io IO) {
	io.Uint8(&x.Cost)
	x.Enchants.Marshal(io)
	io.StringLimits(&x.EnchantName, 1, 256)
	x.EnchantNetID.Marshal(io)
}

type ItemEnchants struct {
	Slot         int32
	ItemEnchants [3][]EnchantmentInstance
}

// Marshal reads or writes ItemEnchants using its canonical wire layout.
func (x *ItemEnchants) Marshal(io IO) {
	io.Int32(&x.Slot)
	for index1 := range x.ItemEnchants {
		Slice(io, &x.ItemEnchants[index1])
	}
}

// ItemInstance represents a unique instance of an item stack. These instances carry a specific network ID
// that is persistent for the stack.
type ItemInstance struct {
	ItemDescriptor ItemDescriptor
	StackSize      uint16
	BlockRuntimeID uint32
	UserDataBuffer []byte
}

// Marshal reads or writes ItemInstance using its canonical wire layout.
func (x *ItemInstance) Marshal(io IO) {
	MarshalItemDescriptor(io, &x.ItemDescriptor)
	io.Uint16(&x.StackSize)
	Minimum(io, &x.StackSize, 1)
	Maximum(io, &x.StackSize, 64)
	io.Varuint32(&x.BlockRuntimeID)
	io.ByteSlice(&x.UserDataBuffer)
}

type ItemReleaseInventoryTransaction struct {
	Actions      InventoryTransactionData
	ActionType   ItemReleaseInventoryTransactionActionType
	Slot         int32
	Item         NetworkItemStackDescriptorSerializedData
	FromPosition mgl32.Vec3
	Hand         HandSlot
}

func (*ItemReleaseInventoryTransaction) tagInventoryTransactionPacketData() uint32 { return 4 }

// Marshal reads or writes ItemReleaseInventoryTransaction using its canonical wire layout.
func (x *ItemReleaseInventoryTransaction) Marshal(io IO) {
	x.Actions.Marshal(io)
	x.ActionType.Marshal(io)
	io.Varint32(&x.Slot)
	x.Item.Marshal(io)
	io.Vec3(&x.FromPosition)
	x.Hand.Marshal(io)
}

type ItemReleaseInventoryTransactionActionType int32

const (
	ItemReleaseInventoryTransactionActionTypeRelease ItemReleaseInventoryTransactionActionType = 0
	ItemReleaseInventoryTransactionActionTypeUse     ItemReleaseInventoryTransactionActionType = 1
)

// Marshal reads or writes ItemReleaseInventoryTransactionActionType through its int32 wire encoding.
func (x *ItemReleaseInventoryTransactionActionType) Marshal(io IO) { io.Varint32((*int32)(x)) }

type ItemUseInventoryTransaction struct {
	Actions                  InventoryTransactionData
	ActionType               ItemUseInventoryTransactionActionType
	TriggerType              ItemUseInventoryTransactionTriggerType
	Position                 BlockPos
	Face                     uint8
	Slot                     int32
	Hand                     HandSlot
	Item                     NetworkItemStackDescriptorSerializedData
	FromPosition             mgl32.Vec3
	ClickPosition            mgl32.Vec3
	TargetBlockID            uint32
	ClientInteractPrediction ItemUseInventoryTransactionPredictedResult
	ClientCooldownState      ItemUseInventoryTransactionClientCooldownState
}

func (*ItemUseInventoryTransaction) tagInventoryTransactionPacketData() uint32 { return 2 }

// Marshal reads or writes ItemUseInventoryTransaction using its canonical wire layout.
func (x *ItemUseInventoryTransaction) Marshal(io IO) {
	x.Actions.Marshal(io)
	x.ActionType.Marshal(io)
	x.TriggerType.Marshal(io)
	x.Position.Marshal(io)
	io.Uint8(&x.Face)
	io.Varint32(&x.Slot)
	x.Hand.Marshal(io)
	x.Item.Marshal(io)
	io.Vec3(&x.FromPosition)
	io.Vec3(&x.ClickPosition)
	io.Varuint32(&x.TargetBlockID)
	x.ClientInteractPrediction.Marshal(io)
	x.ClientCooldownState.Marshal(io)
}

type ItemUseInventoryTransactionActionType int32

const (
	ItemUseInventoryTransactionActionTypePlace       ItemUseInventoryTransactionActionType = 0
	ItemUseInventoryTransactionActionTypeUse         ItemUseInventoryTransactionActionType = 1
	ItemUseInventoryTransactionActionTypeDestroy     ItemUseInventoryTransactionActionType = 2
	ItemUseInventoryTransactionActionTypeUseAsAttack ItemUseInventoryTransactionActionType = 3
)

// Marshal reads or writes ItemUseInventoryTransactionActionType through its int32 wire encoding.
func (x *ItemUseInventoryTransactionActionType) Marshal(io IO) { io.Varint32((*int32)(x)) }

type ItemUseInventoryTransactionClientCooldownState uint8

const (
	ItemUseInventoryTransactionClientCooldownStateOff ItemUseInventoryTransactionClientCooldownState = 0
	ItemUseInventoryTransactionClientCooldownStateOn  ItemUseInventoryTransactionClientCooldownState = 1
)

// Marshal reads or writes ItemUseInventoryTransactionClientCooldownState through its uint8 wire encoding.
func (x *ItemUseInventoryTransactionClientCooldownState) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type ItemUseInventoryTransactionPredictedResult uint8

const (
	ItemUseInventoryTransactionPredictedResultFailure ItemUseInventoryTransactionPredictedResult = 0
	ItemUseInventoryTransactionPredictedResultSuccess ItemUseInventoryTransactionPredictedResult = 1
)

// Marshal reads or writes ItemUseInventoryTransactionPredictedResult through its uint8 wire encoding.
func (x *ItemUseInventoryTransactionPredictedResult) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type ItemUseInventoryTransactionTriggerType uint8

const (
	ItemUseInventoryTransactionTriggerTypeUnknown        ItemUseInventoryTransactionTriggerType = 0
	ItemUseInventoryTransactionTriggerTypePlayerInput    ItemUseInventoryTransactionTriggerType = 1
	ItemUseInventoryTransactionTriggerTypeSimulationTick ItemUseInventoryTransactionTriggerType = 2
)

// Marshal reads or writes ItemUseInventoryTransactionTriggerType through its uint8 wire encoding.
func (x *ItemUseInventoryTransactionTriggerType) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type ItemUseOnActorInventoryTransaction struct {
	Actions      InventoryTransactionData
	RuntimeID    uint64
	ActionType   ItemUseOnActorInventoryTransactionActionType
	Slot         int32
	Hand         HandSlot
	Item         NetworkItemStackDescriptorSerializedData
	FromPosition mgl32.Vec3
	HitPosition  mgl32.Vec3
}

func (*ItemUseOnActorInventoryTransaction) tagInventoryTransactionPacketData() uint32 { return 3 }

// Marshal reads or writes ItemUseOnActorInventoryTransaction using its canonical wire layout.
func (x *ItemUseOnActorInventoryTransaction) Marshal(io IO) {
	x.Actions.Marshal(io)
	io.ActorRuntimeID(&x.RuntimeID)
	x.ActionType.Marshal(io)
	io.Varint32(&x.Slot)
	x.Hand.Marshal(io)
	x.Item.Marshal(io)
	io.Vec3(&x.FromPosition)
	io.Vec3(&x.HitPosition)
}

type ItemUseOnActorInventoryTransactionActionType int32

const (
	ItemUseOnActorInventoryTransactionActionTypeInteract     ItemUseOnActorInventoryTransactionActionType = 0
	ItemUseOnActorInventoryTransactionActionTypeAttack       ItemUseOnActorInventoryTransactionActionType = 1
	ItemUseOnActorInventoryTransactionActionTypeItemInteract ItemUseOnActorInventoryTransactionActionType = 2
)

// Marshal reads or writes ItemUseOnActorInventoryTransactionActionType through its int32 wire encoding.
func (x *ItemUseOnActorInventoryTransactionActionType) Marshal(io IO) { io.Varint32((*int32)(x)) }

// ItemUsedEvent is when a player right clicks a item.
type ItemUsed struct {
	ItemID    int16
	ItemAux   int32
	UseMethod int32
	Count     int32
}

func (*ItemUsed) tagEventData() uint32 { return 20 }

// Marshal reads or writes ItemUsed using its canonical wire layout.
func (x *ItemUsed) Marshal(io IO) {
	io.Int16(&x.ItemID)
	io.Int32(&x.ItemAux)
	io.Int32(&x.UseMethod)
	io.Int32(&x.Count)
}

type ItemVersion int32

const (
	ItemVersionLegacy     ItemVersion = 0
	ItemVersionDataDriven ItemVersion = 1
	ItemVersionNone       ItemVersion = 2
)

// Marshal reads or writes ItemVersion through its int32 wire encoding.
func (x *ItemVersion) Marshal(io IO) { io.Varint32((*int32)(x)) }
