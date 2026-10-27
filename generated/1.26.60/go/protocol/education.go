// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

type EducationEditionOffer uint32

const (
	EducationEditionOfferNone            EducationEditionOffer = 0
	EducationEditionOfferRestOfWorld     EducationEditionOffer = 1
	EducationEditionOfferChinaDeprecated EducationEditionOffer = 2
)

// Marshal reads or writes EducationEditionOffer through its uint32 wire encoding.
func (x *EducationEditionOffer) Marshal(io IO) { io.Varuint32((*uint32)(x)) }

type EducationLocalLevelSettings struct {
	CodeBuilderOverrideURI Optional[string]
}

// Marshal reads or writes EducationLocalLevelSettings using its canonical wire layout.
func (x *EducationLocalLevelSettings) Marshal(io IO) {
	OptionalFunc(io, &x.CodeBuilderOverrideURI, io.String)
}
