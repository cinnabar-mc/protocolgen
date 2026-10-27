// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

type BedrockSafetyRedactableString struct {
	Unredacted string
	Redacted   Optional[string]
}

// Marshal reads or writes BedrockSafetyRedactableString using its canonical wire layout.
func (x *BedrockSafetyRedactableString) Marshal(io IO) {
	io.String(&x.Unredacted)
	OptionalFunc(io, &x.Redacted, io.String)
}
