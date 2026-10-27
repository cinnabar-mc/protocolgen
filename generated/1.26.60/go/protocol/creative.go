// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

// CreativeGroup represents a group of items in the creative inventory. Each group has a category, name and an
// icon that represents the group.
type CreativeGroupInfo struct {
	// CreativeCategory is the category the group falls under. It is one of the constants above.
	CreativeCategory CreativeItemCategory
	// Name is the locale name of the group, i.e. "itemGroup.name.planks".
	Name string
	// GroupIconItem is the item that represents the group in the creative inventory.
	GroupIconItem NetworkItemInstanceDescriptorSerializedData
}

// Marshal reads or writes CreativeGroupInfo using its canonical wire layout.
func (x *CreativeGroupInfo) Marshal(io IO) {
	x.CreativeCategory.Marshal(io)
	io.String(&x.Name)
	x.GroupIconItem.Marshal(io)
}

type CreativeItemCategory uint8

const (
	CreativeItemCategoryConstruction    CreativeItemCategory = 1
	CreativeItemCategoryNature          CreativeItemCategory = 2
	CreativeItemCategoryEquipment       CreativeItemCategory = 3
	CreativeItemCategoryItems           CreativeItemCategory = 4
	CreativeItemCategoryItemCommandOnly CreativeItemCategory = 5
)

// Marshal reads or writes CreativeItemCategory through its uint8 wire encoding.
func (x *CreativeItemCategory) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type CreativeItemEntry struct {
	CreativeNetID CreativeItemNetID
	ItemInstance  NetworkItemInstanceDescriptorSerializedData
	GroupIndex    uint32
}

// Marshal reads or writes CreativeItemEntry using its canonical wire layout.
func (x *CreativeItemEntry) Marshal(io IO) {
	x.CreativeNetID.Marshal(io)
	x.ItemInstance.Marshal(io)
	io.Varuint32(&x.GroupIndex)
}

type CreativeItemNetID struct {
	ID uint32
}

// Marshal reads or writes CreativeItemNetID using its canonical wire layout.
func (x *CreativeItemNetID) Marshal(io IO) {
	io.Varuint32(&x.ID)
}
