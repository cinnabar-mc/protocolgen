// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

type BookEditAction interface {
	Marshaler
	tagBookEditAction() uint32
}

// MarshalBookEditAction reads or writes the BookEditAction union using its canonical wire layout.
func MarshalBookEditAction(io IO, x *BookEditAction) {
	Union(io, x, io.Varuint32, BookEditAction.tagBookEditAction, func(tag uint32) BookEditAction {
		switch tag {
		case 0:
			return new(BookEditActionReplacePage)
		case 1:
			return new(BookEditActionAddPage)
		case 2:
			return new(BookEditActionDeletePage)
		case 3:
			return new(BookEditActionSwapPages)
		case 4:
			return new(BookEditActionFinalize)
		}
		return nil
	})
}
