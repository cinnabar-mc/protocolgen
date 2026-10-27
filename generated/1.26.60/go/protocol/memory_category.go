// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

// MemoryCategoryCounter represents a memory usage counter for a specific category.
type MemoryCategoryCounter struct {
	// Category is the memory category. It is one of the MemoryCategory constants above.
	Category MemoryCategory
	// Bytes is the number of bytes used by this category.
	CurrentBytes uint64
}

// Marshal reads or writes MemoryCategoryCounter using its canonical wire layout.
func (x *MemoryCategoryCounter) Marshal(io IO) {
	x.Category.Marshal(io)
	io.Uint64(&x.CurrentBytes)
}
