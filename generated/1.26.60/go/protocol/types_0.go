// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

type CommandData struct {
	Name                                string
	Description                         string
	Flags                               uint16
	PermissionLevel                     string
	AliasEnum                           int32
	CommandDataChainedSubcommandIndexes []uint32
	Overloads                           []CommandOverload
}

// Marshal reads or writes CommandData using its canonical wire layout.
func (x *CommandData) Marshal(io IO) {
	io.StringLimits(&x.Name, 0, 512)
	io.StringLimits(&x.Description, 0, 1000)
	io.Uint16(&x.Flags)
	io.String(&x.PermissionLevel)
	io.Int32(&x.AliasEnum)
	FuncSliceLimits(io, &x.CommandDataChainedSubcommandIndexes, io.Varuint32, 0, 250, io.Uint32)
	SliceLimits(io, &x.Overloads, 0, 250)
}

type EditorLevelMigrationVersion int8

const (
	EditorLevelMigrationVersionLegacy             EditorLevelMigrationVersion = 0
	EditorLevelMigrationVersionEditorAchievements EditorLevelMigrationVersion = 1
)

// Marshal reads or writes EditorLevelMigrationVersion through its int8 wire encoding.
func (x *EditorLevelMigrationVersion) Marshal(io IO) { io.Int8((*int8)(x)) }

type EditorWorldType int32

const (
	EditorWorldTypeNonEditor          EditorWorldType = 0
	EditorWorldTypeEditorProject      EditorWorldType = 1
	EditorWorldTypeEditorTestLevel    EditorWorldType = 2
	EditorWorldTypeEditorRealmsUpload EditorWorldType = 3
)

// Marshal reads or writes EditorWorldType through its int32 wire encoding.
func (x *EditorWorldType) Marshal(io IO) { io.Varint32((*int32)(x)) }

type SerializedSkin struct {
	ID                           string
	ResourcePatch                string
	ImageData                    SerializedSkinImage
	AnimatedImageData            []SkinAnimatedImage
	CapeImageData                SerializedSkinImage
	GeometryData                 string
	GeometryDataMinEngineVersion string
	AnimationData                string
	CapeID                       string
	FullID                       string
	ArmSize                      SkinArmSize
	SkinColor                    Color
	PersonaPieces                []SkinPersonaPiece
	PieceTintColors              []OrderedEntry[string, SkinTintColor]
	IsPremium                    bool
	IsPersona                    bool
	IsPersonaCapeOnClassicSkin   bool
	IsPrimaryUser                bool
	OverridesPlayerAppearance    bool
	TrustedSkinFlag              string
	ProfileHash                  string
}

// Marshal reads or writes SerializedSkin using its canonical wire layout.
func (x *SerializedSkin) Marshal(io IO) {
	io.String(&x.ID)
	io.String(&x.ResourcePatch)
	x.ImageData.Marshal(io)
	Slice(io, &x.AnimatedImageData)
	x.CapeImageData.Marshal(io)
	io.String(&x.GeometryData)
	io.String(&x.GeometryDataMinEngineVersion)
	io.String(&x.AnimationData)
	io.String(&x.CapeID)
	io.String(&x.FullID)
	x.ArmSize.Marshal(io)
	x.SkinColor.Marshal(io)
	Slice(io, &x.PersonaPieces)
	OrderedMap(io, &x.PieceTintColors, io.Varuint32, io.String, func(value *SkinTintColor) {
		value.Marshal(io)
	})
	io.Bool(&x.IsPremium)
	io.Bool(&x.IsPersona)
	io.Bool(&x.IsPersonaCapeOnClassicSkin)
	io.Bool(&x.IsPrimaryUser)
	io.Bool(&x.OverridesPlayerAppearance)
	io.String(&x.TrustedSkinFlag)
	io.String(&x.ProfileHash)
}

type SkinAnimatedTextureType uint32

const (
	SkinAnimatedTextureTypeNone        SkinAnimatedTextureType = 0
	SkinAnimatedTextureTypeFace        SkinAnimatedTextureType = 1
	SkinAnimatedTextureTypeBody32x32   SkinAnimatedTextureType = 2
	SkinAnimatedTextureTypeBody128x128 SkinAnimatedTextureType = 3
)

// Marshal reads or writes SkinAnimatedTextureType through its uint32 wire encoding.
func (x *SkinAnimatedTextureType) Marshal(io IO) { io.Varuint32((*uint32)(x)) }

type SkinAnimationExpression uint32

const (
	SkinAnimationExpressionLinear   SkinAnimationExpression = 0
	SkinAnimationExpressionBlinking SkinAnimationExpression = 1
)

// Marshal reads or writes SkinAnimationExpression through its uint32 wire encoding.
func (x *SkinAnimationExpression) Marshal(io IO) { io.Varuint32((*uint32)(x)) }

type SkinArmSize uint8

const (
	SkinArmSizeSlim SkinArmSize = 0
	SkinArmSizeWide SkinArmSize = 1
)

// Marshal reads or writes SkinArmSize through its uint8 wire encoding.
func (x *SkinArmSize) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type SkinPieceType uint32

const (
	SkinPieceTypeUnknown       SkinPieceType = 0
	SkinPieceTypeSkeleton      SkinPieceType = 1
	SkinPieceTypeBody          SkinPieceType = 2
	SkinPieceTypeSkin          SkinPieceType = 3
	SkinPieceTypeBottom        SkinPieceType = 4
	SkinPieceTypeFeet          SkinPieceType = 5
	SkinPieceTypeDress         SkinPieceType = 6
	SkinPieceTypeTop           SkinPieceType = 7
	SkinPieceTypeHighPants     SkinPieceType = 8
	SkinPieceTypeHands         SkinPieceType = 9
	SkinPieceTypeOuterwear     SkinPieceType = 10
	SkinPieceTypeFacialHair    SkinPieceType = 11
	SkinPieceTypeMouth         SkinPieceType = 12
	SkinPieceTypeEyes          SkinPieceType = 13
	SkinPieceTypeHair          SkinPieceType = 14
	SkinPieceTypeHood          SkinPieceType = 15
	SkinPieceTypeBack          SkinPieceType = 16
	SkinPieceTypeFaceAccessory SkinPieceType = 17
	SkinPieceTypeHead          SkinPieceType = 18
	SkinPieceTypeLegs          SkinPieceType = 19
	SkinPieceTypeLeftLeg       SkinPieceType = 20
	SkinPieceTypeRightLeg      SkinPieceType = 21
	SkinPieceTypeArms          SkinPieceType = 22
	SkinPieceTypeLeftArm       SkinPieceType = 23
	SkinPieceTypeRightArm      SkinPieceType = 24
	SkinPieceTypeCapes         SkinPieceType = 25
	SkinPieceTypeClassicSkin   SkinPieceType = 26
	SkinPieceTypeEmote         SkinPieceType = 27
	SkinPieceTypeCoCo          SkinPieceType = 28
	SkinPieceTypeUnsupported   SkinPieceType = 29
)

// Marshal reads or writes SkinPieceType through its uint32 wire encoding.
func (x *SkinPieceType) Marshal(io IO) { io.Uint32((*uint32)(x)) }
