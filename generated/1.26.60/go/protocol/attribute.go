// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

// AttributeModifier temporarily buffs/debuffs a given attribute until the modifier is used. In vanilla, these
// are mainly used for effects.
type AttributeModifier struct {
	// ID is the unique ID of the modifier. It is used to identify the modifier in the packet.
	ID string
	// Name is the name of the attribute that is modified.
	Name string
	// Amount is the amount of difference between the current value of the attribute and the new value.
	Amount float32
	// Operation is the operation that is performed on the attribute. It can be addition, multiply base, multiply
	// total or cap.
	Operation int32
	// Operand ... TODO: Figure out what this field is used for.
	Operand int32
	// IsSerializable ... TODO: Figure out what this field is used for.
	IsSerializable bool
}

// Marshal reads or writes AttributeModifier using its canonical wire layout.
func (x *AttributeModifier) Marshal(io IO) {
	io.String(&x.ID)
	io.String(&x.Name)
	io.Float32(&x.Amount)
	io.Int32(&x.Operation)
	io.Int32(&x.Operand)
	io.Bool(&x.IsSerializable)
}
