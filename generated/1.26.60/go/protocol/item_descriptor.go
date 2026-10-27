// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

// DefaultItemDescriptor represents an item descriptor for regular items. This is used for the significant
// majority of items.
type DefaultItemDescriptor struct {
	DescriptorType ItemDescriptorType
	// Name is the identifier of the item, such as minecraft:stone.
	FullName string
	AuxValue int32
}

func (*DefaultItemDescriptor) tagItemDescriptor() uint32 { return 1 }

// Marshal reads or writes DefaultItemDescriptor using its canonical wire layout.
func (x *DefaultItemDescriptor) Marshal(io IO) {
	x.DescriptorType.Marshal(io)
	io.StringLimits(&x.FullName, 1, 18446744073709551615)
	io.Varint32(&x.AuxValue)
	Minimum(io, &x.AuxValue, 0)
	Maximum(io, &x.AuxValue, 32767)
}

// InvalidItemDescriptor represents an invalid item descriptor. This is usually sent by the vanilla server for
// empty slots or ingredients.
type InvalidItemDescriptor struct {
	DescriptorType ItemDescriptorType
}

func (*InvalidItemDescriptor) tagItemDescriptor() uint32 { return 0 }

// Marshal reads or writes InvalidItemDescriptor using its canonical wire layout.
func (x *InvalidItemDescriptor) Marshal(io IO) {
	x.DescriptorType.Marshal(io)
}

// ItemDescriptor represents a type of item descriptor. This is one of the concrete types below. It is an
// alias of Marshaler.
type ItemDescriptorType uint8

const (
	ItemDescriptorTypeEmpty    ItemDescriptorType = 0
	ItemDescriptorTypeItemName ItemDescriptorType = 1
	ItemDescriptorTypeMoLang   ItemDescriptorType = 2
	ItemDescriptorTypeItemTag  ItemDescriptorType = 3
)

// Marshal reads or writes ItemDescriptorType through its uint8 wire encoding.
func (x *ItemDescriptorType) Marshal(io IO) { io.Uint8((*uint8)(x)) }

// ItemTagItemDescriptor represents an item descriptor that uses item tagging. This should be used to reduce
// duplicative entries for items that can be grouped under a single tag.
type ItemTagItemDescriptor struct {
	DescriptorType ItemDescriptorType
	ItemTag        string
}

func (*ItemTagItemDescriptor) tagItemDescriptor() uint32 { return 3 }

// Marshal reads or writes ItemTagItemDescriptor using its canonical wire layout.
func (x *ItemTagItemDescriptor) Marshal(io IO) {
	x.DescriptorType.Marshal(io)
	io.StringLimits(&x.ItemTag, 1, 18446744073709551615)
}

// MoLangItemDescriptor represents an item descriptor for items that use MoLang (e.g. behaviour packs).
type MoLangItemDescriptor struct {
	DescriptorType ItemDescriptorType
	// Expression represents the MoLang expression used to identify the item/it's associated tag.
	TagExpression string
	// Version represents the version of MoLang to use.
	MoLangVersion MoLangVersion
}

func (*MoLangItemDescriptor) tagItemDescriptor() uint32 { return 2 }

// Marshal reads or writes MoLangItemDescriptor using its canonical wire layout.
func (x *MoLangItemDescriptor) Marshal(io IO) {
	x.DescriptorType.Marshal(io)
	io.StringLimits(&x.TagExpression, 1, 18446744073709551615)
	x.MoLangVersion.Marshal(io)
}
