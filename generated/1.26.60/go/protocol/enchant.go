// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

type EnchantType uint8

const (
	EnchantTypeProtection           EnchantType = 0
	EnchantTypeFireProtection       EnchantType = 1
	EnchantTypeFeatherFalling       EnchantType = 2
	EnchantTypeBlastProtection      EnchantType = 3
	EnchantTypeProjectileProtection EnchantType = 4
	EnchantTypeThorns               EnchantType = 5
	EnchantTypeRespiration          EnchantType = 6
	EnchantTypeDepthStrider         EnchantType = 7
	EnchantTypeAquaAffinity         EnchantType = 8
	EnchantTypeSharpness            EnchantType = 9
	EnchantTypeSmite                EnchantType = 10
	EnchantTypeBaneOfArthropods     EnchantType = 11
	EnchantTypeKnockback            EnchantType = 12
	EnchantTypeFireAspect           EnchantType = 13
	EnchantTypeLooting              EnchantType = 14
	EnchantTypeEfficiency           EnchantType = 15
	EnchantTypeSilkTouch            EnchantType = 16
	EnchantTypeUnbreaking           EnchantType = 17
	EnchantTypeFortune              EnchantType = 18
	EnchantTypePower                EnchantType = 19
	EnchantTypePunch                EnchantType = 20
	EnchantTypeFlame                EnchantType = 21
	EnchantTypeInfinity             EnchantType = 22
	EnchantTypeLuckOfTheSea         EnchantType = 23
	EnchantTypeLure                 EnchantType = 24
	EnchantTypeFrostWalker          EnchantType = 25
	EnchantTypeMending              EnchantType = 26
	EnchantTypeCurseOfBinding       EnchantType = 27
	EnchantTypeCurseOfVanishing     EnchantType = 28
	EnchantTypeImpaling             EnchantType = 29
	EnchantTypeRiptide              EnchantType = 30
	EnchantTypeLoyalty              EnchantType = 31
	EnchantTypeChanneling           EnchantType = 32
	EnchantTypeMultishot            EnchantType = 33
	EnchantTypePiercing             EnchantType = 34
	EnchantTypeQuickCharge          EnchantType = 35
	EnchantTypeSoulSpeed            EnchantType = 36
	EnchantTypeSwiftSneak           EnchantType = 37
	EnchantTypeWindBurst            EnchantType = 38
	EnchantTypeDensity              EnchantType = 39
	EnchantTypeBreach               EnchantType = 40
	EnchantTypeLunge                EnchantType = 41
	EnchantTypeNumEnchantments      EnchantType = 42
	EnchantTypeInvalidEnchantment   EnchantType = 43
)

// Marshal reads or writes EnchantType through its uint8 wire encoding.
func (x *EnchantType) Marshal(io IO) { io.Uint8((*uint8)(x)) }

// EnchantmentInstance represents a single enchantment instance with the type of the enchantment and its
// level.
type EnchantmentInstance struct {
	EnchantType  EnchantType
	EnchantLevel uint8
}

// Marshal reads or writes EnchantmentInstance using its canonical wire layout.
func (x *EnchantmentInstance) Marshal(io IO) {
	x.EnchantType.Marshal(io)
	io.Uint8(&x.EnchantLevel)
}
