// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

type BedrockDDUIDataStoreUpdateData interface {
	Marshaler
	tagBedrockDDUIDataStoreUpdateData() uint32
}

// MarshalBedrockDDUIDataStoreUpdateData reads or writes the BedrockDDUIDataStoreUpdateData union using its canonical wire layout.
func MarshalBedrockDDUIDataStoreUpdateData(io IO, x *BedrockDDUIDataStoreUpdateData) {
	Union(io, x, io.Varuint32, BedrockDDUIDataStoreUpdateData.tagBedrockDDUIDataStoreUpdateData, func(tag uint32) BedrockDDUIDataStoreUpdateData {
		switch tag {
		case 0:
			return new(BedrockDDUIDataStoreUpdateDataDouble)
		case 1:
			return new(BedrockDDUIDataStoreUpdateDataBool)
		case 2:
			return new(BedrockDDUIDataStoreUpdateDataString)
		}
		return nil
	})
}

type BedrockDDUIDataStoreUpdateDataBool struct {
	Value bool
}

func (*BedrockDDUIDataStoreUpdateDataBool) tagBedrockDDUIDataStoreUpdateData() uint32 { return 1 }

// Marshal reads or writes BedrockDDUIDataStoreUpdateDataBool using its canonical wire layout.
func (x *BedrockDDUIDataStoreUpdateDataBool) Marshal(io IO) {
	io.Bool(&x.Value)
}

type BedrockDDUIDataStoreUpdateDataDouble struct {
	Value float64
}

func (*BedrockDDUIDataStoreUpdateDataDouble) tagBedrockDDUIDataStoreUpdateData() uint32 { return 0 }

// Marshal reads or writes BedrockDDUIDataStoreUpdateDataDouble using its canonical wire layout.
func (x *BedrockDDUIDataStoreUpdateDataDouble) Marshal(io IO) {
	io.Float64(&x.Value)
}

type BedrockDDUIDataStoreUpdateDataString struct {
	Value string
}

func (*BedrockDDUIDataStoreUpdateDataString) tagBedrockDDUIDataStoreUpdateData() uint32 { return 2 }

// Marshal reads or writes BedrockDDUIDataStoreUpdateDataString using its canonical wire layout.
func (x *BedrockDDUIDataStoreUpdateDataString) Marshal(io IO) {
	io.StringLimits(&x.Value, 0, 5000)
}

// Bitset131 stores the 131-bit value used by the wire bitset encoding.
type Bitset131 [3]uint64

const Bitset131Length = 131

// Set marks bit index i. It panics when i is outside [0, 131).
func (b *Bitset131) Set(i int) {
	if i < 0 || i >= Bitset131Length {
		panic("index out of bounds")
	}
	b[i/64] |= uint64(1) << uint(i%64)
}

// Unset clears bit index i. It panics when i is outside [0, 131).
func (b *Bitset131) Unset(i int) {
	if i < 0 || i >= Bitset131Length {
		panic("index out of bounds")
	}
	b[i/64] &^= uint64(1) << uint(i%64)
}

// Load reports whether bit index i is set. It panics when i is outside [0, 131).
func (b Bitset131) Load(i int) bool {
	if i < 0 || i >= Bitset131Length {
		panic("index out of bounds")
	}
	return b[i/64]&(uint64(1)<<uint(i%64)) != 0
}

// Len returns the number of bits in the bitset.
func (b Bitset131) Len() int { return Bitset131Length }

type DataItemEntryValue interface {
	Marshaler
	tagDataItemEntryValue() uint32
}

// MarshalDataItemEntryValue reads or writes the DataItemEntryValue union using its canonical wire layout.
func MarshalDataItemEntryValue(io IO, x *DataItemEntryValue) {
	Union(io, x, io.Varuint32, DataItemEntryValue.tagDataItemEntryValue, func(tag uint32) DataItemEntryValue {
		switch tag {
		case 0:
			return new(DataItemByte)
		case 1:
			return new(DataItemShort)
		case 2:
			return new(DataItemInt)
		case 3:
			return new(DataItemFloat)
		case 4:
			return new(DataItemString)
		case 5:
			return new(DataItemCompoundTag)
		case 6:
			return new(DataItemPos)
		case 7:
			return new(DataItemInt64)
		case 8:
			return new(DataItemVec3)
		}
		return nil
	})
}

type DisconnectMessages interface {
	Marshaler
	tagDisconnectMessages() uint32
}

// MarshalDisconnectMessages reads or writes the DisconnectMessages union using its canonical wire layout.
func MarshalDisconnectMessages(io IO, x *DisconnectMessages) {
	Union(io, x, io.Varuint32, DisconnectMessages.tagDisconnectMessages, func(tag uint32) DisconnectMessages {
		switch tag {
		case 0:
			return new(DisconnectMessagesData)
		case 1:
			return new(DisconnectMessagesEmpty)
		}
		return nil
	})
}

type DisconnectMessagesEmpty struct {
}

func (*DisconnectMessagesEmpty) tagDisconnectMessages() uint32 { return 1 }

// Marshal reads or writes DisconnectMessagesEmpty using its canonical wire layout.
func (x *DisconnectMessagesEmpty) Marshal(io IO) {
}

type DynamicValueBool struct {
	Value bool
}

func (*DynamicValueBool) tagDynamicValue() int32 { return 1 }

// Marshal reads or writes DynamicValueBool using its canonical wire layout.
func (x *DynamicValueBool) Marshal(io IO) {
	io.Bool(&x.Value)
}

type DynamicValueDouble struct {
	Value float64
}

func (*DynamicValueDouble) tagDynamicValue() int32 { return 3 }

// Marshal reads or writes DynamicValueDouble using its canonical wire layout.
func (x *DynamicValueDouble) Marshal(io IO) {
	io.Float64(&x.Value)
}

type DynamicValueInt64 struct {
	Value int64
}

func (*DynamicValueInt64) tagDynamicValue() int32 { return 2 }

// Marshal reads or writes DynamicValueInt64 using its canonical wire layout.
func (x *DynamicValueInt64) Marshal(io IO) {
	io.Int64(&x.Value)
}

type DynamicValueList struct {
	Value []DynamicValue
}

func (*DynamicValueList) tagDynamicValue() int32 { return 5 }

// Marshal reads or writes DynamicValueList using its canonical wire layout.
func (x *DynamicValueList) Marshal(io IO) {
	FuncSlice(io, &x.Value, io.Varuint32, func(value *DynamicValue) {
		MarshalDynamicValue(io, value)
	})
}

type DynamicValueMap struct {
	Value []OrderedEntry[string, DynamicValue]
}

func (*DynamicValueMap) tagDynamicValue() int32 { return 6 }

// Marshal reads or writes DynamicValueMap using its canonical wire layout.
func (x *DynamicValueMap) Marshal(io IO) {
	OrderedMap(io, &x.Value, io.Varuint32, io.String, func(value *DynamicValue) {
		MarshalDynamicValue(io, value)
	})
}

type DynamicValueNone struct {
}

func (*DynamicValueNone) tagDynamicValue() int32 { return 0 }

// Marshal reads or writes DynamicValueNone using its canonical wire layout.
func (x *DynamicValueNone) Marshal(io IO) {
}

type DynamicValueString struct {
	Value string
}

func (*DynamicValueString) tagDynamicValue() int32 { return 4 }

// Marshal reads or writes DynamicValueString using its canonical wire layout.
func (x *DynamicValueString) Marshal(io IO) {
	io.String(&x.Value)
}

type GameRuleValue interface {
	Marshaler
	tagGameRuleValue() uint32
}

// MarshalGameRuleValue reads or writes the GameRuleValue union using its canonical wire layout.
func MarshalGameRuleValue(io IO, x *GameRuleValue) {
	Union(io, x, io.Varuint32, GameRuleValue.tagGameRuleValue, func(tag uint32) GameRuleValue {
		switch tag {
		case 0:
			return new(GameRuleValueEmpty)
		case 1:
			return new(GameRuleValueBool)
		case 2:
			return new(GameRuleValueInt32)
		case 3:
			return new(GameRuleValueFloat)
		}
		return nil
	})
}

type GameRuleValueBool struct {
	Value bool
}

func (*GameRuleValueBool) tagGameRuleValue() uint32 { return 1 }

// Marshal reads or writes GameRuleValueBool using its canonical wire layout.
func (x *GameRuleValueBool) Marshal(io IO) {
	io.Bool(&x.Value)
}

type GameRuleValueEmpty struct {
}

func (*GameRuleValueEmpty) tagGameRuleValue() uint32 { return 0 }

// Marshal reads or writes GameRuleValueEmpty using its canonical wire layout.
func (x *GameRuleValueEmpty) Marshal(io IO) {
}

type GameRuleValueFloat struct {
	Value float32
}

func (*GameRuleValueFloat) tagGameRuleValue() uint32 { return 3 }

// Marshal reads or writes GameRuleValueFloat using its canonical wire layout.
func (x *GameRuleValueFloat) Marshal(io IO) {
	io.Float32(&x.Value)
}

type GameRuleValueInt32 struct {
	Value int32
}

func (*GameRuleValueInt32) tagGameRuleValue() uint32 { return 2 }

// Marshal reads or writes GameRuleValueInt32 using its canonical wire layout.
func (x *GameRuleValueInt32) Marshal(io IO) {
	io.Int32(&x.Value)
}

type InventoryTransactionPacketData interface {
	Marshaler
	tagInventoryTransactionPacketData() uint32
}

// MarshalInventoryTransactionPacketData reads or writes the InventoryTransactionPacketData union using its canonical wire layout.
func MarshalInventoryTransactionPacketData(io IO, x *InventoryTransactionPacketData) {
	Union(io, x, io.Varuint32, InventoryTransactionPacketData.tagInventoryTransactionPacketData, func(tag uint32) InventoryTransactionPacketData {
		switch tag {
		case 0:
			return new(NormalTransactionData)
		case 1:
			return new(InventoryMismatchData)
		case 2:
			return new(ItemUseInventoryTransaction)
		case 3:
			return new(ItemUseOnActorInventoryTransaction)
		case 4:
			return new(ItemReleaseInventoryTransaction)
		}
		return nil
	})
}

type LevelGameRuleValue interface {
	Marshaler
	tagLevelGameRuleValue() uint32
}

// MarshalLevelGameRuleValue reads or writes the LevelGameRuleValue union using its canonical wire layout.
func MarshalLevelGameRuleValue(io IO, x *LevelGameRuleValue) {
	Union(io, x, io.Varuint32, LevelGameRuleValue.tagLevelGameRuleValue, func(tag uint32) LevelGameRuleValue {
		switch tag {
		case 0:
			return new(LevelGameRuleValueVariant0)
		case 1:
			return new(LevelGameRuleValueVariant1)
		case 2:
			return new(LevelGameRuleValueVariant2)
		case 3:
			return new(LevelGameRuleValueVariant3)
		}
		return nil
	})
}

type LevelGameRuleValueVariant0 struct {
}

func (*LevelGameRuleValueVariant0) tagLevelGameRuleValue() uint32 { return 0 }

// Marshal reads or writes LevelGameRuleValueVariant0 using its canonical wire layout.
func (x *LevelGameRuleValueVariant0) Marshal(io IO) {
}

type LevelGameRuleValueVariant1 struct {
	Value bool
}

func (*LevelGameRuleValueVariant1) tagLevelGameRuleValue() uint32 { return 1 }

// Marshal reads or writes LevelGameRuleValueVariant1 using its canonical wire layout.
func (x *LevelGameRuleValueVariant1) Marshal(io IO) {
	io.Bool(&x.Value)
}

type LevelGameRuleValueVariant2 struct {
	Value int32
}

func (*LevelGameRuleValueVariant2) tagLevelGameRuleValue() uint32 { return 2 }

// Marshal reads or writes LevelGameRuleValueVariant2 using its canonical wire layout.
func (x *LevelGameRuleValueVariant2) Marshal(io IO) {
	io.Int32(&x.Value)
}

type LevelGameRuleValueVariant3 struct {
	Value float32
}

func (*LevelGameRuleValueVariant3) tagLevelGameRuleValue() uint32 { return 3 }

// Marshal reads or writes LevelGameRuleValueVariant3 using its canonical wire layout.
func (x *LevelGameRuleValueVariant3) Marshal(io IO) {
	io.Float32(&x.Value)
}

type PrimitiveShapeExtraShapeData interface {
	Marshaler
	tagPrimitiveShapeExtraShapeData() uint32
}

// MarshalPrimitiveShapeExtraShapeData reads or writes the PrimitiveShapeExtraShapeData union using its canonical wire layout.
func MarshalPrimitiveShapeExtraShapeData(io IO, x *PrimitiveShapeExtraShapeData) {
	Union(io, x, io.Varuint32, PrimitiveShapeExtraShapeData.tagPrimitiveShapeExtraShapeData, func(tag uint32) PrimitiveShapeExtraShapeData {
		switch tag {
		case 0:
			return new(PrimitiveShapeExtraShapeDataEmpty)
		case 1:
			return new(ArrowData)
		case 2:
			return new(TextShape)
		case 3:
			return new(BoxData)
		case 4:
			return new(LineData)
		case 5:
			return new(SphereData)
		case 6:
			return new(CylinderData)
		case 7:
			return new(PyramidData)
		case 8:
			return new(EllipsoidData)
		case 9:
			return new(ConeData)
		}
		return nil
	})
}

type PrimitiveShapeExtraShapeDataEmpty struct {
}

func (*PrimitiveShapeExtraShapeDataEmpty) tagPrimitiveShapeExtraShapeData() uint32 { return 0 }

// Marshal reads or writes PrimitiveShapeExtraShapeDataEmpty using its canonical wire layout.
func (x *PrimitiveShapeExtraShapeDataEmpty) Marshal(io IO) {
}

type ServerboundPackSettingChangePackSettingValue interface {
	Marshaler
	tagServerboundPackSettingChangePackSettingValue() uint32
}

// MarshalServerboundPackSettingChangePackSettingValue reads or writes the ServerboundPackSettingChangePackSettingValue union using its canonical wire layout.
func MarshalServerboundPackSettingChangePackSettingValue(io IO, x *ServerboundPackSettingChangePackSettingValue) {
	Union(io, x, io.Varuint32, ServerboundPackSettingChangePackSettingValue.tagServerboundPackSettingChangePackSettingValue, func(tag uint32) ServerboundPackSettingChangePackSettingValue {
		switch tag {
		case 0:
			return new(ServerboundPackSettingChangePackSettingValueFloat)
		case 1:
			return new(ServerboundPackSettingChangePackSettingValueBool)
		case 2:
			return new(ServerboundPackSettingChangePackSettingValueString)
		case 3:
			return new(ServerboundPackSettingChangePackSettingValueCase3)
		}
		return nil
	})
}

type ServerboundPackSettingChangePackSettingValueBool struct {
	Value bool
}

func (*ServerboundPackSettingChangePackSettingValueBool) tagServerboundPackSettingChangePackSettingValue() uint32 {
	return 1
}

// Marshal reads or writes ServerboundPackSettingChangePackSettingValueBool using its canonical wire layout.
func (x *ServerboundPackSettingChangePackSettingValueBool) Marshal(io IO) {
	io.Bool(&x.Value)
}

type ServerboundPackSettingChangePackSettingValueCase3 struct {
	Value []string
}

func (*ServerboundPackSettingChangePackSettingValueCase3) tagServerboundPackSettingChangePackSettingValue() uint32 {
	return 3
}

// Marshal reads or writes ServerboundPackSettingChangePackSettingValueCase3 using its canonical wire layout.
func (x *ServerboundPackSettingChangePackSettingValueCase3) Marshal(io IO) {
	FuncSlice(io, &x.Value, io.Varuint32, io.String)
}

type ServerboundPackSettingChangePackSettingValueFloat struct {
	Value float32
}

func (*ServerboundPackSettingChangePackSettingValueFloat) tagServerboundPackSettingChangePackSettingValue() uint32 {
	return 0
}

// Marshal reads or writes ServerboundPackSettingChangePackSettingValueFloat using its canonical wire layout.
func (x *ServerboundPackSettingChangePackSettingValueFloat) Marshal(io IO) {
	io.Float32(&x.Value)
}

type ServerboundPackSettingChangePackSettingValueString struct {
	Value string
}

func (*ServerboundPackSettingChangePackSettingValueString) tagServerboundPackSettingChangePackSettingValue() uint32 {
	return 2
}

// Marshal reads or writes ServerboundPackSettingChangePackSettingValueString using its canonical wire layout.
func (x *ServerboundPackSettingChangePackSettingValueString) Marshal(io IO) {
	io.String(&x.Value)
}

type SetScoreInfoItem interface {
	Marshaler
	tagSetScoreInfoItem() uint32
}

// MarshalSetScoreInfoItem reads or writes the SetScoreInfoItem union using its canonical wire layout.
func MarshalSetScoreInfoItem(io IO, x *SetScoreInfoItem) {
	Union(io, x, io.Varuint32, SetScoreInfoItem.tagSetScoreInfoItem, func(tag uint32) SetScoreInfoItem {
		switch tag {
		case 0:
			return new(RemoveScore)
		case 1:
			return new(ChangePlayerScore)
		case 2:
			return new(ChangeEntityScore)
		case 3:
			return new(ChangeFakePlayerScore)
		}
		return nil
	})
}
