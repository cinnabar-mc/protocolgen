// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

import (
	"image/color"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
)

type Achievement struct {
	AchievementID MinecraftEventingAchievementIds
}

func (*Achievement) tagEventData() uint32 { return 0 }

// Marshal reads or writes Achievement using its canonical wire layout.
func (x *Achievement) Marshal(io IO) {
	x.AchievementID.Marshal(io)
}

type AddEntry struct {
	Action           PlayerListPacketType
	UUID             uuid.UUID
	ActorUniqueID    int64
	PlayerName       string
	XBLXUID          string
	PlayFabID        string
	PlatformOnlineID string
	BuildPlatform    BuildPlatform
	SerializedSkin   SerializedSkinRef
	IsTeacher        bool
	IsHost           bool
	IsSubClient      bool
	PlayerColor      color.RGBA
}

func (*AddEntry) tagPlayerListData() uint32 { return 1 }

// Marshal reads or writes AddEntry using its canonical wire layout.
func (x *AddEntry) Marshal(io IO) {
	x.Action.Marshal(io)
	io.UUID(&x.UUID)
	io.ActorUniqueID(&x.ActorUniqueID)
	io.String(&x.PlayerName)
	io.String(&x.XBLXUID)
	io.String(&x.PlayFabID)
	io.String(&x.PlatformOnlineID)
	x.BuildPlatform.Marshal(io)
	x.SerializedSkin.Marshal(io)
	io.Bool(&x.IsTeacher)
	io.Bool(&x.IsHost)
	io.Bool(&x.IsSubClient)
	io.RGBA(&x.PlayerColor)
}

type AddTimeMarkerData struct {
	ClockID     uint64
	TimeMarkers []TimeMarkerData
}

func (*AddTimeMarkerData) tagSyncWorldClocksData() uint32 { return 2 }

// Marshal reads or writes AddTimeMarkerData using its canonical wire layout.
func (x *AddTimeMarkerData) Marshal(io IO) {
	io.Varuint64(&x.ClockID)
	SliceLimits(io, &x.TimeMarkers, 0, 256)
}

type AgentActionType int32

const (
	AgentActionTypeAttack            AgentActionType = 1
	AgentActionTypeCollect           AgentActionType = 2
	AgentActionTypeDestroy           AgentActionType = 3
	AgentActionTypeDetectRedstone    AgentActionType = 4
	AgentActionTypeDetectObstacle    AgentActionType = 5
	AgentActionTypeDrop              AgentActionType = 6
	AgentActionTypeDropAll           AgentActionType = 7
	AgentActionTypeInspect           AgentActionType = 8
	AgentActionTypeInspectData       AgentActionType = 9
	AgentActionTypeInspectItemCount  AgentActionType = 10
	AgentActionTypeInspectItemDetail AgentActionType = 11
	AgentActionTypeInspectItemSpace  AgentActionType = 12
	AgentActionTypeInteract          AgentActionType = 13
	AgentActionTypeMove              AgentActionType = 14
	AgentActionTypePlaceBlock        AgentActionType = 15
	AgentActionTypeTill              AgentActionType = 16
	AgentActionTypeTransferItemTo    AgentActionType = 17
	AgentActionTypeTurn              AgentActionType = 18
)

// Marshal reads or writes AgentActionType through its int32 wire encoding.
func (x *AgentActionType) Marshal(io IO) { io.Int32((*int32)(x)) }

type AgentAnimationType uint8

const (
	AgentAnimationTypeArmSwing AgentAnimationType = 0
	AgentAnimationTypeShrug    AgentAnimationType = 1
)

// Marshal reads or writes AgentAnimationType through its uint8 wire encoding.
func (x *AgentAnimationType) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type AnimateAction uint8

const (
	AnimateActionNoAction         AnimateAction = 0
	AnimateActionSwing            AnimateAction = 1
	AnimateActionWakeUp           AnimateAction = 3
	AnimateActionCriticalHit      AnimateAction = 4
	AnimateActionMagicCriticalHit AnimateAction = 5
)

// Marshal reads or writes AnimateAction through its uint8 wire encoding.
func (x *AnimateAction) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type AnimatedImageData struct {
	SkinImage           SkinImage
	AnimatedTextureType PersonaAnimatedTextureType
	Frames              float32
	AnimationExpression PersonaAnimationExpression
}

// Marshal reads or writes AnimatedImageData using its canonical wire layout.
func (x *AnimatedImageData) Marshal(io IO) {
	x.SkinImage.Marshal(io)
	x.AnimatedTextureType.Marshal(io)
	io.Float32(&x.Frames)
	x.AnimationExpression.Marshal(io)
}

// PlayerArmourDamageEntry represents an entry for a single piece of armour that should be damaged.
type ArmorSlotAndDamagePair struct {
	// ArmourSlot is the index of the armour slot to damage.
	ArmorSlot LegacyArmorSlot
	// Damage is the amount of damage to apply to the armour in the specified slot.
	Damage int16
}

// Marshal reads or writes ArmorSlotAndDamagePair using its canonical wire layout.
func (x *ArmorSlotAndDamagePair) Marshal(io IO) {
	x.ArmorSlot.Marshal(io)
	io.Int16(&x.Damage)
}

// ArrowShape represents an arrow debug shape.
type ArrowData struct {
	// ArrowEndLocation is the arrow end location of the shape.
	ArrowEndLocation Optional[mgl32.Vec3]
	// ArrowHeadLength is the arrow head length of the shape.
	ArrowHeadLength Optional[float32]
	// ArrowHeadRadius is the arrow head radius of the shape.
	ArrowHeadRadius Optional[float32]
	// Segments is the segments that used for the debug arrow's head.
	NumSegments Optional[uint8]
}

func (*ArrowData) tagPrimitiveShapeExtraShapeData() uint32 { return 1 }

// Marshal reads or writes ArrowData using its canonical wire layout.
func (x *ArrowData) Marshal(io IO) {
	OptionalFunc(io, &x.ArrowEndLocation, io.Vec3)
	OptionalFunc(io, &x.ArrowHeadLength, io.Float32)
	OptionalFunc(io, &x.ArrowHeadRadius, io.Float32)
	OptionalFunc(io, &x.NumSegments, io.Uint8)
}

type AudioSoundHandle struct {
	ServerSoundHandle uint64
}

// Marshal reads or writes AudioSoundHandle using its canonical wire layout.
func (x *AudioSoundHandle) Marshal(io IO) {
	io.Uint64(&x.ServerSoundHandle)
}

type AuthorAndMessage struct {
	MessageType TextPacketType
	PlayerName  string
	Message     string
}

func (*AuthorAndMessage) tagTextData() uint32 { return 1 }

// Marshal reads or writes AuthorAndMessage using its canonical wire layout.
func (x *AuthorAndMessage) Marshal(io IO) {
	x.MessageType.Marshal(io)
	io.StringLimits(&x.PlayerName, 0, 256)
	io.StringLimits(&x.Message, 1, 65536)
}

type BedrockDDUI interface {
	Marshaler
	tagBedrockDDUI() uint32
}

// MarshalBedrockDDUI reads or writes the BedrockDDUI union using its canonical wire layout.
func MarshalBedrockDDUI(io IO, x *BedrockDDUI) {
	Union(io, x, io.Varuint32, BedrockDDUI.tagBedrockDDUI, func(tag uint32) BedrockDDUI {
		switch tag {
		case 0:
			return new(BedrockDDUIDataStoreUpdate)
		case 1:
			return new(BedrockDDUIDataStoreChange)
		case 2:
			return new(BedrockDDUIDataStoreRemoval)
		}
		return nil
	})
}

// DataStoreChange represents a change to a data store property value.
type BedrockDDUIDataStoreChange struct {
	// DataStoreName is the name of the data store.
	DataStoreName string
	// Property is the property that changed.
	Property string
	// UpdateCount is the update count.
	UpdateCount uint32
	// NewValue is the new property value.
	TheNewPropertyValue DynamicValue
}

func (*BedrockDDUIDataStoreChange) tagBedrockDDUI() uint32 { return 1 }

// Marshal reads or writes BedrockDDUIDataStoreChange using its canonical wire layout.
func (x *BedrockDDUIDataStoreChange) Marshal(io IO) {
	io.StringLimits(&x.DataStoreName, 1, 1000)
	io.StringLimits(&x.Property, 1, 1000)
	io.Uint32(&x.UpdateCount)
	Maximum(io, &x.UpdateCount, 4.294967294e+09)
	MarshalDynamicValue(io, &x.TheNewPropertyValue)
}

type BedrockDDUIDataStoreRemoval struct {
	DataStoreName string
}

func (*BedrockDDUIDataStoreRemoval) tagBedrockDDUI() uint32 { return 2 }

// Marshal reads or writes BedrockDDUIDataStoreRemoval using its canonical wire layout.
func (x *BedrockDDUIDataStoreRemoval) Marshal(io IO) {
	io.StringLimits(&x.DataStoreName, 1, 1000)
}

type BedrockDDUIDataStoreUpdate struct {
	DataStoreName       string
	Property            string
	Path                string
	Data                BedrockDDUIDataStoreUpdateData
	PropertyUpdateCount uint32
	PathUpdateCount     uint32
}

func (*BedrockDDUIDataStoreUpdate) tagBedrockDDUI() uint32 { return 0 }

// Marshal reads or writes BedrockDDUIDataStoreUpdate using its canonical wire layout.
func (x *BedrockDDUIDataStoreUpdate) Marshal(io IO) {
	io.StringLimits(&x.DataStoreName, 1, 1000)
	io.StringLimits(&x.Property, 1, 1000)
	io.StringLimits(&x.Path, 0, 1000)
	MarshalBedrockDDUIDataStoreUpdateData(io, &x.Data)
	io.Uint32(&x.PropertyUpdateCount)
	Maximum(io, &x.PropertyUpdateCount, 4.294967294e+09)
	io.Uint32(&x.PathUpdateCount)
	Maximum(io, &x.PathUpdateCount, 4.294967294e+09)
}

// BellUsedEvent is the event data sent when a bell is used.
type BellUsed struct {
	// ItemID ...
	ItemID int16
}

func (*BellUsed) tagEventData() uint32 { return 12 }

// Marshal reads or writes BellUsed using its canonical wire layout.
func (x *BellUsed) Marshal(io IO) {
	io.Int16(&x.ItemID)
}

type BookEditActionAddPage struct {
	PageIndex int32
	PageText  string
	PhotoName string
}

func (*BookEditActionAddPage) tagBookEditAction() uint32 { return 1 }

// Marshal reads or writes BookEditActionAddPage using its canonical wire layout.
func (x *BookEditActionAddPage) Marshal(io IO) {
	io.Varint32(&x.PageIndex)
	io.StringLimits(&x.PageText, 0, 768)
	io.StringLimits(&x.PhotoName, 0, 768)
}

type BookEditActionDeletePage struct {
	PageIndex int32
}

func (*BookEditActionDeletePage) tagBookEditAction() uint32 { return 2 }

// Marshal reads or writes BookEditActionDeletePage using its canonical wire layout.
func (x *BookEditActionDeletePage) Marshal(io IO) {
	io.Varint32(&x.PageIndex)
}

type BookEditActionFinalize struct {
	Title  string
	Author string
	XUID   string
}

func (*BookEditActionFinalize) tagBookEditAction() uint32 { return 4 }

// Marshal reads or writes BookEditActionFinalize using its canonical wire layout.
func (x *BookEditActionFinalize) Marshal(io IO) {
	io.StringLimits(&x.Title, 0, 768)
	io.StringLimits(&x.Author, 0, 768)
	io.StringLimits(&x.XUID, 0, 768)
}

type BookEditActionReplacePage struct {
	PageIndex int32
	PageText  string
	PhotoName string
}

func (*BookEditActionReplacePage) tagBookEditAction() uint32 { return 0 }

// Marshal reads or writes BookEditActionReplacePage using its canonical wire layout.
func (x *BookEditActionReplacePage) Marshal(io IO) {
	io.Varint32(&x.PageIndex)
	io.StringLimits(&x.PageText, 0, 768)
	io.StringLimits(&x.PhotoName, 0, 768)
}

type BookEditActionSwapPages struct {
	PageIndex     int32
	SwapWithIndex int32
}

func (*BookEditActionSwapPages) tagBookEditAction() uint32 { return 3 }

// Marshal reads or writes BookEditActionSwapPages using its canonical wire layout.
func (x *BookEditActionSwapPages) Marshal(io IO) {
	io.Varint32(&x.PageIndex)
	io.Varint32(&x.SwapWithIndex)
}

type BooleanAttributeValue struct {
	Value     bool
	Operation string
}

func (*BooleanAttributeValue) tagAttributeValue() uint32 { return 0 }

// Marshal reads or writes BooleanAttributeValue using its canonical wire layout.
func (x *BooleanAttributeValue) Marshal(io IO) {
	io.Bool(&x.Value)
	io.String(&x.Operation)
}

type BossBarColor uint8

const (
	BossBarColorPink          BossBarColor = 0
	BossBarColorBlue          BossBarColor = 1
	BossBarColorRed           BossBarColor = 2
	BossBarColorGreen         BossBarColor = 3
	BossBarColorYellow        BossBarColor = 4
	BossBarColorPurple        BossBarColor = 5
	BossBarColorRebeccaPurple BossBarColor = 6
	BossBarColorWhite         BossBarColor = 7
)

// Marshal reads or writes BossBarColor through its uint8 wire encoding.
func (x *BossBarColor) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type BossBarOverlay uint8

const (
	BossBarOverlayProgress           BossBarOverlay = 0
	BossBarOverlayNotchedGenerated6  BossBarOverlay = 1
	BossBarOverlayNotchedGenerated10 BossBarOverlay = 2
	BossBarOverlayNotchedGenerated12 BossBarOverlay = 3
	BossBarOverlayNotchedGenerated20 BossBarOverlay = 4
)

// Marshal reads or writes BossBarOverlay through its uint8 wire encoding.
func (x *BossBarOverlay) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type BossEventUpdateType uint8

const (
	BossEventUpdateTypeAdd              BossEventUpdateType = 0
	BossEventUpdateTypePlayerAdded      BossEventUpdateType = 1
	BossEventUpdateTypeRemove           BossEventUpdateType = 2
	BossEventUpdateTypePlayerRemoved    BossEventUpdateType = 3
	BossEventUpdateTypeUpdatePercent    BossEventUpdateType = 4
	BossEventUpdateTypeUpdateName       BossEventUpdateType = 5
	BossEventUpdateTypeUpdateProperties BossEventUpdateType = 6
	BossEventUpdateTypeUpdateStyle      BossEventUpdateType = 7
	BossEventUpdateTypeQuery            BossEventUpdateType = 8
)

// Marshal reads or writes BossEventUpdateType through its uint8 wire encoding.
func (x *BossEventUpdateType) Marshal(io IO) { io.Uint8((*uint8)(x)) }

// BossKilledEvent is the event data sent when a boss dies.
type BossKilled struct {
	// BossEntityUniqueID ...
	BossActorID int64
	// PlayerPartySize ...
	PartySize int32
	// InteractionEntityType ...
	BossType int32
}

func (*BossKilled) tagEventData() uint32 { return 7 }

// Marshal reads or writes BossKilled using its canonical wire layout.
func (x *BossKilled) Marshal(io IO) {
	io.Varint64(&x.BossActorID)
	io.Varint32(&x.PartySize)
	io.Varint32(&x.BossType)
}

type BoxData struct {
	BoxBound mgl32.Vec3
}

func (*BoxData) tagPrimitiveShapeExtraShapeData() uint32 { return 3 }

// Marshal reads or writes BoxData using its canonical wire layout.
func (x *BoxData) Marshal(io IO) {
	io.Vec3(&x.BoxBound)
}

type BuildPlatform int32

const (
	BuildPlatformUnknown      BuildPlatform = -1
	BuildPlatformGoogle       BuildPlatform = 1
	BuildPlatformIOS          BuildPlatform = 2
	BuildPlatformOSX          BuildPlatform = 3
	BuildPlatformAmazon       BuildPlatform = 4
	BuildPlatformGearVR       BuildPlatform = 5
	BuildPlatformUWP          BuildPlatform = 7
	BuildPlatformWin32        BuildPlatform = 8
	BuildPlatformDedicated    BuildPlatform = 9
	BuildPlatformTvOS         BuildPlatform = 10
	BuildPlatformSony         BuildPlatform = 11
	BuildPlatformNintendo     BuildPlatform = 12
	BuildPlatformXbox         BuildPlatform = 13
	BuildPlatformWindowsPhone BuildPlatform = 14
	BuildPlatformLinux        BuildPlatform = 15
)

// Marshal reads or writes BuildPlatform through its int32 wire encoding.
func (x *BuildPlatform) Marshal(io IO) { io.Int32((*int32)(x)) }

// CauldronUsedEvent is the event data sent when a cauldron is used.
type CauldronUsed struct {
	// Colour ...
	ContentsColor uint32
	// PotionID ...
	ContentsType int16
	// FillLevel ...
	FillLevel int16
}

func (*CauldronUsed) tagEventData() uint32 { return 5 }

// Marshal reads or writes CauldronUsed using its canonical wire layout.
func (x *CauldronUsed) Marshal(io IO) {
	io.Varuint32(&x.ContentsColor)
	io.Int16(&x.ContentsType)
	io.Int16(&x.FillLevel)
}

type ChangeEntityScore struct {
	Action        string
	ScoreboardID  ScoreboardID
	ObjectiveName string
	ScoreValue    int32
	ActorID       int64
}

func (*ChangeEntityScore) tagSetScoreInfoItem() uint32 { return 2 }

// Marshal reads or writes ChangeEntityScore using its canonical wire layout.
func (x *ChangeEntityScore) Marshal(io IO) {
	io.String(&x.Action)
	x.ScoreboardID.Marshal(io)
	io.String(&x.ObjectiveName)
	io.Int32(&x.ScoreValue)
	io.ActorUniqueID(&x.ActorID)
}

type ChangeFakePlayerScore struct {
	Action         string
	ScoreboardID   ScoreboardID
	ObjectiveName  string
	ScoreValue     int32
	FakePlayerName string
}

func (*ChangeFakePlayerScore) tagSetScoreInfoItem() uint32 { return 3 }

// Marshal reads or writes ChangeFakePlayerScore using its canonical wire layout.
func (x *ChangeFakePlayerScore) Marshal(io IO) {
	io.String(&x.Action)
	x.ScoreboardID.Marshal(io)
	io.String(&x.ObjectiveName)
	io.Int32(&x.ScoreValue)
	io.String(&x.FakePlayerName)
}

type ChangePlayerScore struct {
	Action         string
	ScoreboardID   ScoreboardID
	ObjectiveName  string
	ScoreValue     int32
	PlayerUniqueID PlayerScoreboardID
}

func (*ChangePlayerScore) tagSetScoreInfoItem() uint32 { return 1 }

// Marshal reads or writes ChangePlayerScore using its canonical wire layout.
func (x *ChangePlayerScore) Marshal(io IO) {
	io.String(&x.Action)
	x.ScoreboardID.Marshal(io)
	io.String(&x.ObjectiveName)
	io.Int32(&x.ScoreValue)
	x.PlayerUniqueID.Marshal(io)
}

type ChatRestrictionLevel uint8

const (
	ChatRestrictionLevelNone     ChatRestrictionLevel = 0
	ChatRestrictionLevelDropped  ChatRestrictionLevel = 1
	ChatRestrictionLevelDisabled ChatRestrictionLevel = 2
)

// Marshal reads or writes ChatRestrictionLevel through its uint8 wire encoding.
func (x *ChatRestrictionLevel) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type ClearOverride struct {
	Type string
}

func (*ClearOverride) tagPlayerUpdateEntityOverridesData() uint32 { return 0 }

// Marshal reads or writes ClearOverride using its canonical wire layout.
func (x *ClearOverride) Marshal(io IO) {
	io.String(&x.Type)
}

type ClientCameraAimAssistAction uint8

const (
	ClientCameraAimAssistActionSetFromCameraPreset ClientCameraAimAssistAction = 0
	ClientCameraAimAssistActionClear               ClientCameraAimAssistAction = 1
)

// Marshal reads or writes ClientCameraAimAssistAction through its uint8 wire encoding.
func (x *ClientCameraAimAssistAction) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type ClientPlayMode uint32

const (
	ClientPlayModeNormal              ClientPlayMode = 0
	ClientPlayModeTeaser              ClientPlayMode = 1
	ClientPlayModeScreen              ClientPlayMode = 2
	ClientPlayModeViewer              ClientPlayMode = 3
	ClientPlayModeReality             ClientPlayMode = 4
	ClientPlayModePlacement           ClientPlayMode = 5
	ClientPlayModeLivingRoom          ClientPlayMode = 6
	ClientPlayModeExitLevel           ClientPlayMode = 7
	ClientPlayModeExitLevelLivingRoom ClientPlayMode = 8
	ClientPlayModeNumModes            ClientPlayMode = 9
)

// Marshal reads or writes ClientPlayMode through its uint32 wire encoding.
func (x *ClientPlayMode) Marshal(io IO) { io.Varuint32((*uint32)(x)) }

type ClientboundTextureShiftAction uint8

const (
	ClientboundTextureShiftActionInvalid    ClientboundTextureShiftAction = 0
	ClientboundTextureShiftActionInitialize ClientboundTextureShiftAction = 1
	ClientboundTextureShiftActionStart      ClientboundTextureShiftAction = 2
	ClientboundTextureShiftActionSetEnabled ClientboundTextureShiftAction = 3
	ClientboundTextureShiftActionSync       ClientboundTextureShiftAction = 4
)

// Marshal reads or writes ClientboundTextureShiftAction through its uint8 wire encoding.
func (x *ClientboundTextureShiftAction) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type CodeBuilderExecutionStateCodeStatus uint8

const (
	CodeBuilderExecutionStateCodeStatusNone       CodeBuilderExecutionStateCodeStatus = 0
	CodeBuilderExecutionStateCodeStatusNotStarted CodeBuilderExecutionStateCodeStatus = 1
	CodeBuilderExecutionStateCodeStatusInProgress CodeBuilderExecutionStateCodeStatus = 2
	CodeBuilderExecutionStateCodeStatusPaused     CodeBuilderExecutionStateCodeStatus = 3
	CodeBuilderExecutionStateCodeStatusError      CodeBuilderExecutionStateCodeStatus = 4
	CodeBuilderExecutionStateCodeStatusSucceeded  CodeBuilderExecutionStateCodeStatus = 5
)

// Marshal reads or writes CodeBuilderExecutionStateCodeStatus through its uint8 wire encoding.
func (x *CodeBuilderExecutionStateCodeStatus) Marshal(io IO) { io.Uint8((*uint8)(x)) }

// CodeBuilderRuntimeActionEvent is an event sent by the server when a code builder runtime action is
// performed.
type CodeBuilderRuntimeAction struct {
	// Action ...
	CodeBuilderRuntimeAction string
}

func (*CodeBuilderRuntimeAction) tagEventData() uint32 { return 18 }

// Marshal reads or writes CodeBuilderRuntimeAction using its canonical wire layout.
func (x *CodeBuilderRuntimeAction) Marshal(io IO) {
	io.StringLimits(&x.CodeBuilderRuntimeAction, 0, 16)
}

// CodeBuilderScoreboardEvent is an event sent by the server when a code builder scoreboard is updated.
type CodeBuilderScoreboard struct {
	// ObjectiveName ...
	ObjectiveName string
	// Score ...
	Score int32
}

func (*CodeBuilderScoreboard) tagEventData() uint32 { return 19 }

// Marshal reads or writes CodeBuilderScoreboard using its canonical wire layout.
func (x *CodeBuilderScoreboard) Marshal(io IO) {
	io.StringLimits(&x.ObjectiveName, 0, 256)
	io.Varint32(&x.Score)
}

type CodeBuilderStorageQueryOptionsCategory uint8

const (
	CodeBuilderStorageQueryOptionsCategoryNone          CodeBuilderStorageQueryOptionsCategory = 0
	CodeBuilderStorageQueryOptionsCategoryCodeStatus    CodeBuilderStorageQueryOptionsCategory = 1
	CodeBuilderStorageQueryOptionsCategoryInstantiation CodeBuilderStorageQueryOptionsCategory = 2
)

// Marshal reads or writes CodeBuilderStorageQueryOptionsCategory through its uint8 wire encoding.
func (x *CodeBuilderStorageQueryOptionsCategory) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type CodeBuilderStorageQueryOptionsOperation uint8

const (
	CodeBuilderStorageQueryOptionsOperationNone  CodeBuilderStorageQueryOptionsOperation = 0
	CodeBuilderStorageQueryOptionsOperationGet   CodeBuilderStorageQueryOptionsOperation = 1
	CodeBuilderStorageQueryOptionsOperationSet   CodeBuilderStorageQueryOptionsOperation = 2
	CodeBuilderStorageQueryOptionsOperationReset CodeBuilderStorageQueryOptionsOperation = 3
)

// Marshal reads or writes CodeBuilderStorageQueryOptionsOperation through its uint8 wire encoding.
func (x *CodeBuilderStorageQueryOptionsOperation) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type Color struct {
	Color int32
}

// Marshal reads or writes Color using its canonical wire layout.
func (x *Color) Marshal(io IO) {
	io.Int32(&x.Color)
}

type ColourAttributeValue struct {
	Value     [4]int32
	Operation string
}

func (*ColourAttributeValue) tagAttributeValue() uint32 { return 2 }

// Marshal reads or writes ColourAttributeValue using its canonical wire layout.
func (x *ColourAttributeValue) Marshal(io IO) {
	for index1 := range x.Value {
		io.Int32(&x.Value[index1])
	}
	io.String(&x.Operation)
}

// ComposterInteractEvent is the event data sent when a composter is interacted with.
type ComposterUsed struct {
	// BlockInteractionType ...
	BlockInteractionType MinecraftEventingPOIBlockInteractionType
	// ItemID ...
	ItemID int16
}

func (*ComposterUsed) tagEventData() uint32 { return 11 }

// Marshal reads or writes ComposterUsed using its canonical wire layout.
func (x *ComposterUsed) Marshal(io IO) {
	x.BlockInteractionType.Marshal(io)
	io.Int16(&x.ItemID)
}

// ConeShape represents a cone debug shape.
type ConeData struct {
	// Radii are the radii along the X/Z axes of the cone base.
	Radii mgl32.Vec2
	// Height is the height of the cone.
	Height float32
	// NumSegments is the number of segments used for the cone.
	NumSegments uint8
}

func (*ConeData) tagPrimitiveShapeExtraShapeData() uint32 { return 9 }

// Marshal reads or writes ConeData using its canonical wire layout.
func (x *ConeData) Marshal(io IO) {
	io.Vec2(&x.Radii)
	io.Float32(&x.Height)
	io.Uint8(&x.NumSegments)
}

type ConnectionDisconnectFailReason int32

const (
	ConnectionDisconnectFailReasonUnknown                                       ConnectionDisconnectFailReason = 0
	ConnectionDisconnectFailReasonCantConnectNoInternet                         ConnectionDisconnectFailReason = 1
	ConnectionDisconnectFailReasonNoPermissions                                 ConnectionDisconnectFailReason = 2
	ConnectionDisconnectFailReasonUnrecoverableError                            ConnectionDisconnectFailReason = 3
	ConnectionDisconnectFailReasonThirdPartyBlocked                             ConnectionDisconnectFailReason = 4
	ConnectionDisconnectFailReasonThirdPartyNoInternet                          ConnectionDisconnectFailReason = 5
	ConnectionDisconnectFailReasonThirdPartyBadIP                               ConnectionDisconnectFailReason = 6
	ConnectionDisconnectFailReasonThirdPartyNoServerOrServerLocked              ConnectionDisconnectFailReason = 7
	ConnectionDisconnectFailReasonVersionMismatch                               ConnectionDisconnectFailReason = 8
	ConnectionDisconnectFailReasonSkinIssue                                     ConnectionDisconnectFailReason = 9
	ConnectionDisconnectFailReasonInviteSessionNotFound                         ConnectionDisconnectFailReason = 10
	ConnectionDisconnectFailReasonEduLevelSettingsMissing                       ConnectionDisconnectFailReason = 11
	ConnectionDisconnectFailReasonLocalServerNotFound                           ConnectionDisconnectFailReason = 12
	ConnectionDisconnectFailReasonLegacyDisconnect                              ConnectionDisconnectFailReason = 13
	ConnectionDisconnectFailReasonINTERNALUserLeaveGameAttempted                ConnectionDisconnectFailReason = 14
	ConnectionDisconnectFailReasonPlatformLockedSkinsError                      ConnectionDisconnectFailReason = 15
	ConnectionDisconnectFailReasonRealmsWorldUnassigned                         ConnectionDisconnectFailReason = 16
	ConnectionDisconnectFailReasonRealmsServerCantConnect                       ConnectionDisconnectFailReason = 17
	ConnectionDisconnectFailReasonRealmsServerHidden                            ConnectionDisconnectFailReason = 18
	ConnectionDisconnectFailReasonRealmsServerDisabledBeta                      ConnectionDisconnectFailReason = 19
	ConnectionDisconnectFailReasonRealmsServerDisabled                          ConnectionDisconnectFailReason = 20
	ConnectionDisconnectFailReasonCrossPlatformDisabled                         ConnectionDisconnectFailReason = 21
	ConnectionDisconnectFailReasonTESTONLYCantConnect                           ConnectionDisconnectFailReason = 22
	ConnectionDisconnectFailReasonSessionNotFound                               ConnectionDisconnectFailReason = 23
	ConnectionDisconnectFailReasonClientSettingsIncompatibleWithServer          ConnectionDisconnectFailReason = 24
	ConnectionDisconnectFailReasonServerFull                                    ConnectionDisconnectFailReason = 25
	ConnectionDisconnectFailReasonInvalidPlatformSkin                           ConnectionDisconnectFailReason = 26
	ConnectionDisconnectFailReasonEditionVersionMismatch                        ConnectionDisconnectFailReason = 27
	ConnectionDisconnectFailReasonEditionMismatch                               ConnectionDisconnectFailReason = 28
	ConnectionDisconnectFailReasonLevelNewerThanExeVersion                      ConnectionDisconnectFailReason = 29
	ConnectionDisconnectFailReasonINTERNALNoFailOccurred                        ConnectionDisconnectFailReason = 30
	ConnectionDisconnectFailReasonBannedSkin                                    ConnectionDisconnectFailReason = 31
	ConnectionDisconnectFailReasonTimeout                                       ConnectionDisconnectFailReason = 32
	ConnectionDisconnectFailReasonServerNotFound                                ConnectionDisconnectFailReason = 33
	ConnectionDisconnectFailReasonOutdatedServer                                ConnectionDisconnectFailReason = 34
	ConnectionDisconnectFailReasonOutdatedClient                                ConnectionDisconnectFailReason = 35
	ConnectionDisconnectFailReasonNoPremiumPlatform                             ConnectionDisconnectFailReason = 36
	ConnectionDisconnectFailReasonMultiplayerDisabled                           ConnectionDisconnectFailReason = 37
	ConnectionDisconnectFailReasonNoWiFi                                        ConnectionDisconnectFailReason = 38
	ConnectionDisconnectFailReasonWorldCorruption                               ConnectionDisconnectFailReason = 39
	ConnectionDisconnectFailReasonNoReason                                      ConnectionDisconnectFailReason = 40
	ConnectionDisconnectFailReasonDisconnected                                  ConnectionDisconnectFailReason = 41
	ConnectionDisconnectFailReasonInvalidPlayer                                 ConnectionDisconnectFailReason = 42
	ConnectionDisconnectFailReasonLoggedInOtherLocation                         ConnectionDisconnectFailReason = 43
	ConnectionDisconnectFailReasonServerIDConflict                              ConnectionDisconnectFailReason = 44
	ConnectionDisconnectFailReasonNotAllowed                                    ConnectionDisconnectFailReason = 45
	ConnectionDisconnectFailReasonNotAuthenticated                              ConnectionDisconnectFailReason = 46
	ConnectionDisconnectFailReasonInvalidTenant                                 ConnectionDisconnectFailReason = 47
	ConnectionDisconnectFailReasonUnknownPacket                                 ConnectionDisconnectFailReason = 48
	ConnectionDisconnectFailReasonUnexpectedPacket                              ConnectionDisconnectFailReason = 49
	ConnectionDisconnectFailReasonInvalidCommandRequestPacket                   ConnectionDisconnectFailReason = 50
	ConnectionDisconnectFailReasonHostSuspended                                 ConnectionDisconnectFailReason = 51
	ConnectionDisconnectFailReasonLoginPacketNoRequest                          ConnectionDisconnectFailReason = 52
	ConnectionDisconnectFailReasonLoginPacketNoCert                             ConnectionDisconnectFailReason = 53
	ConnectionDisconnectFailReasonMissingClient                                 ConnectionDisconnectFailReason = 54
	ConnectionDisconnectFailReasonKicked                                        ConnectionDisconnectFailReason = 55
	ConnectionDisconnectFailReasonKickedForExploit                              ConnectionDisconnectFailReason = 56
	ConnectionDisconnectFailReasonKickedForIdle                                 ConnectionDisconnectFailReason = 57
	ConnectionDisconnectFailReasonResourcePackProblem                           ConnectionDisconnectFailReason = 58
	ConnectionDisconnectFailReasonIncompatiblePack                              ConnectionDisconnectFailReason = 59
	ConnectionDisconnectFailReasonOutOfStorage                                  ConnectionDisconnectFailReason = 60
	ConnectionDisconnectFailReasonInvalidLevel                                  ConnectionDisconnectFailReason = 61
	ConnectionDisconnectFailReasonDisconnectPacket                              ConnectionDisconnectFailReason = 62
	ConnectionDisconnectFailReasonBlockMismatch                                 ConnectionDisconnectFailReason = 63
	ConnectionDisconnectFailReasonInvalidHeights                                ConnectionDisconnectFailReason = 64
	ConnectionDisconnectFailReasonInvalidWidths                                 ConnectionDisconnectFailReason = 65
	ConnectionDisconnectFailReasonConnectionLost                                ConnectionDisconnectFailReason = 66
	ConnectionDisconnectFailReasonZombieConnection                              ConnectionDisconnectFailReason = 67
	ConnectionDisconnectFailReasonShutdown                                      ConnectionDisconnectFailReason = 68
	ConnectionDisconnectFailReasonReasonNotSet                                  ConnectionDisconnectFailReason = 69
	ConnectionDisconnectFailReasonLoadingStateTimeout                           ConnectionDisconnectFailReason = 70
	ConnectionDisconnectFailReasonResourcePackLoadingFailed                     ConnectionDisconnectFailReason = 71
	ConnectionDisconnectFailReasonSearchingForSessionLoadingScreenFailed        ConnectionDisconnectFailReason = 72
	ConnectionDisconnectFailReasonNetherNetProtocolVersion                      ConnectionDisconnectFailReason = 73
	ConnectionDisconnectFailReasonSubsystemStatusError                          ConnectionDisconnectFailReason = 74
	ConnectionDisconnectFailReasonEmptyAuthFromDiscovery                        ConnectionDisconnectFailReason = 75
	ConnectionDisconnectFailReasonEmptyURLFromDiscovery                         ConnectionDisconnectFailReason = 76
	ConnectionDisconnectFailReasonExpiredAuthFromDiscovery                      ConnectionDisconnectFailReason = 77
	ConnectionDisconnectFailReasonUnknownSignalServiceSignInFailure             ConnectionDisconnectFailReason = 78
	ConnectionDisconnectFailReasonXBLJoinLobbyFailure                           ConnectionDisconnectFailReason = 79
	ConnectionDisconnectFailReasonUnspecifiedClientInstanceDisconnection        ConnectionDisconnectFailReason = 80
	ConnectionDisconnectFailReasonNetherNetSessionNotFound                      ConnectionDisconnectFailReason = 81
	ConnectionDisconnectFailReasonNetherNetCreatePeerConnection                 ConnectionDisconnectFailReason = 82
	ConnectionDisconnectFailReasonNetherNetICE                                  ConnectionDisconnectFailReason = 83
	ConnectionDisconnectFailReasonNetherNetConnectRequest                       ConnectionDisconnectFailReason = 84
	ConnectionDisconnectFailReasonNetherNetConnectResponse                      ConnectionDisconnectFailReason = 85
	ConnectionDisconnectFailReasonNetherNetNegotiationTimeout                   ConnectionDisconnectFailReason = 86
	ConnectionDisconnectFailReasonNetherNetInactivityTimeout                    ConnectionDisconnectFailReason = 87
	ConnectionDisconnectFailReasonStaleConnectionBeingReplaced                  ConnectionDisconnectFailReason = 88
	ConnectionDisconnectFailReasonRealmsSessionNotFound                         ConnectionDisconnectFailReason = 89
	ConnectionDisconnectFailReasonBadPacket                                     ConnectionDisconnectFailReason = 90
	ConnectionDisconnectFailReasonNetherNetFailedToCreateOffer                  ConnectionDisconnectFailReason = 91
	ConnectionDisconnectFailReasonNetherNetFailedToCreateAnswer                 ConnectionDisconnectFailReason = 92
	ConnectionDisconnectFailReasonNetherNetFailedToSetLocalDescription          ConnectionDisconnectFailReason = 93
	ConnectionDisconnectFailReasonNetherNetFailedToSetRemoteDescription         ConnectionDisconnectFailReason = 94
	ConnectionDisconnectFailReasonNetherNetNegotiationTimeoutWaitingForResponse ConnectionDisconnectFailReason = 95
	ConnectionDisconnectFailReasonNetherNetNegotiationTimeoutWaitingForAccept   ConnectionDisconnectFailReason = 96
	ConnectionDisconnectFailReasonNetherNetIncomingConnectionIgnored            ConnectionDisconnectFailReason = 97
	ConnectionDisconnectFailReasonNetherNetSignalingParsingFailure              ConnectionDisconnectFailReason = 98
	ConnectionDisconnectFailReasonNetherNetSignalingUnknownError                ConnectionDisconnectFailReason = 99
	ConnectionDisconnectFailReasonNetherNetSignalingUnicastDeliveryFailed       ConnectionDisconnectFailReason = 100
	ConnectionDisconnectFailReasonNetherNetSignalingBroadcastDeliveryFailed     ConnectionDisconnectFailReason = 101
	ConnectionDisconnectFailReasonNetherNetSignalingGenericDeliveryFailed       ConnectionDisconnectFailReason = 102
	ConnectionDisconnectFailReasonEditorMismatchEditorWorld                     ConnectionDisconnectFailReason = 103
	ConnectionDisconnectFailReasonEditorMismatchVanillaWorld                    ConnectionDisconnectFailReason = 104
	ConnectionDisconnectFailReasonWorldTransferNotPrimaryClient                 ConnectionDisconnectFailReason = 105
	ConnectionDisconnectFailReasonINTERNALRequestServerShutdown                 ConnectionDisconnectFailReason = 106
	ConnectionDisconnectFailReasonClientGameSetupCancelled                      ConnectionDisconnectFailReason = 107
	ConnectionDisconnectFailReasonClientGameSetupFailed                         ConnectionDisconnectFailReason = 108
	ConnectionDisconnectFailReasonNoVenue                                       ConnectionDisconnectFailReason = 109
	ConnectionDisconnectFailReasonNetherNetSignalingSigninFailed                ConnectionDisconnectFailReason = 110
	ConnectionDisconnectFailReasonSessionAccessDenied                           ConnectionDisconnectFailReason = 111
	ConnectionDisconnectFailReasonServiceSigninIssue                            ConnectionDisconnectFailReason = 112
	ConnectionDisconnectFailReasonNetherNetNoSignalingChannel                   ConnectionDisconnectFailReason = 113
	ConnectionDisconnectFailReasonNetherNetNotLoggedIn                          ConnectionDisconnectFailReason = 114
	ConnectionDisconnectFailReasonNetherNetClientSignalingError                 ConnectionDisconnectFailReason = 115
	ConnectionDisconnectFailReasonSubClientLoginDisabled                        ConnectionDisconnectFailReason = 116
	ConnectionDisconnectFailReasonDeepLinkTryingToOpenDemoWorldWhileSignedIn    ConnectionDisconnectFailReason = 117
	ConnectionDisconnectFailReasonAsyncJoinTaskDenied                           ConnectionDisconnectFailReason = 118
	ConnectionDisconnectFailReasonRealmsTimelineRequired                        ConnectionDisconnectFailReason = 119
	ConnectionDisconnectFailReasonGuestWithoutHost                              ConnectionDisconnectFailReason = 120
	ConnectionDisconnectFailReasonFailedToJoinExperience                        ConnectionDisconnectFailReason = 121
	ConnectionDisconnectFailReasonNetherNetDataChannelClosed                    ConnectionDisconnectFailReason = 122
	ConnectionDisconnectFailReasonDiscoveryEnvironmentMismatch                  ConnectionDisconnectFailReason = 123
	ConnectionDisconnectFailReasonHostWithoutKeys                               ConnectionDisconnectFailReason = 124
	ConnectionDisconnectFailReasonHostSignedOut                                 ConnectionDisconnectFailReason = 125
	ConnectionDisconnectFailReasonScriptWatchdogException                       ConnectionDisconnectFailReason = 126
	ConnectionDisconnectFailReasonScriptMemoryLimitExceeded                     ConnectionDisconnectFailReason = 127
	ConnectionDisconnectFailReasonStorageLowDuringGameplay                      ConnectionDisconnectFailReason = 128
	ConnectionDisconnectFailReasonStorageFullDuringGameplay                     ConnectionDisconnectFailReason = 129
	ConnectionDisconnectFailReasonLevelStorageCorruption                        ConnectionDisconnectFailReason = 130
	ConnectionDisconnectFailReasonEditionMismatchVanillaToEdu                   ConnectionDisconnectFailReason = 131
	ConnectionDisconnectFailReasonEditionMismatchEduToVanilla                   ConnectionDisconnectFailReason = 132
	ConnectionDisconnectFailReasonEditorMismatchEditorToVanilla                 ConnectionDisconnectFailReason = 133
	ConnectionDisconnectFailReasonEditorMismatchVanillaToEditor                 ConnectionDisconnectFailReason = 134
	ConnectionDisconnectFailReasonDenyListed                                    ConnectionDisconnectFailReason = 135
	ConnectionDisconnectFailReasonNonceMissing                                  ConnectionDisconnectFailReason = 136
	ConnectionDisconnectFailReasonNonceNotFound                                 ConnectionDisconnectFailReason = 137
	ConnectionDisconnectFailReasonNonceExpired                                  ConnectionDisconnectFailReason = 138
	ConnectionDisconnectFailReasonNonceNotValid                                 ConnectionDisconnectFailReason = 139
	ConnectionDisconnectFailReasonHostDisconnected                              ConnectionDisconnectFailReason = 140
	ConnectionDisconnectFailReasonEditorJoinIntentPolicyFailure                 ConnectionDisconnectFailReason = 141
	ConnectionDisconnectFailReasonNetherNetIdentityNotAllowed                   ConnectionDisconnectFailReason = 142
	ConnectionDisconnectFailReasonInvalidName                                   ConnectionDisconnectFailReason = 143
	ConnectionDisconnectFailReasonExpiredToken                                  ConnectionDisconnectFailReason = 144
	ConnectionDisconnectFailReasonHostAcceptsNoTypeOfAuth                       ConnectionDisconnectFailReason = 145
	ConnectionDisconnectFailReasonNotAuthenticatedFastFail                      ConnectionDisconnectFailReason = 146
	ConnectionDisconnectFailReasonEditorNotAllowed                              ConnectionDisconnectFailReason = 147
	ConnectionDisconnectFailReasonMissingStructureData                          ConnectionDisconnectFailReason = 148
	ConnectionDisconnectFailReasonUnsupportedTransport                          ConnectionDisconnectFailReason = 149
	ConnectionDisconnectFailReasonSubClientLoginNotBound                        ConnectionDisconnectFailReason = 150
)

// Marshal reads or writes ConnectionDisconnectFailReason through its int32 wire encoding.
func (x *ConnectionDisconnectFailReason) Marshal(io IO) { io.Varint32((*int32)(x)) }

type ContentIdentity struct {
	Identity string
}

// Marshal reads or writes ContentIdentity using its canonical wire layout.
func (x *ContentIdentity) Marshal(io IO) {
	io.String(&x.Identity)
}

type ControlScheme uint8

const (
	ControlSchemeLockedPlayerRelativeStrafe ControlScheme = 0
	ControlSchemeCameraRelative             ControlScheme = 1
	ControlSchemeCameraRelativeStrafe       ControlScheme = 2
	ControlSchemePlayerRelative             ControlScheme = 3
	ControlSchemePlayerRelativeStrafe       ControlScheme = 4
)

// Marshal reads or writes ControlScheme through its uint8 wire encoding.
func (x *ControlScheme) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type CoordinateEvaluationOrder int32

const (
	CoordinateEvaluationOrderXyz CoordinateEvaluationOrder = 0
	CoordinateEvaluationOrderXzy CoordinateEvaluationOrder = 1
	CoordinateEvaluationOrderYxz CoordinateEvaluationOrder = 2
	CoordinateEvaluationOrderYzx CoordinateEvaluationOrder = 3
	CoordinateEvaluationOrderZxy CoordinateEvaluationOrder = 4
	CoordinateEvaluationOrderZyx CoordinateEvaluationOrder = 5
)

// Marshal reads or writes CoordinateEvaluationOrder through its int32 wire encoding.
func (x *CoordinateEvaluationOrder) Marshal(io IO) { io.Varint32((*int32)(x)) }

type CoordinatesLocation struct {
	PacketType PlayerLocationType
	Position   mgl32.Vec3
}

func (*CoordinatesLocation) tagPlayerLocationData() uint32 { return 0 }

// Marshal reads or writes CoordinatesLocation using its canonical wire layout.
func (x *CoordinatesLocation) Marshal(io IO) {
	x.PacketType.Marshal(io)
	io.Vec3(&x.Position)
}

type CraftLoomStackRequestAction struct {
	ActionType    ItemStackRequestActionType
	PatternNameID string
	NumCrafts     uint8
}

func (*CraftLoomStackRequestAction) tagStackRequestAction() uint32 { return 15 }

// Marshal reads or writes CraftLoomStackRequestAction using its canonical wire layout.
func (x *CraftLoomStackRequestAction) Marshal(io IO) {
	x.ActionType.Marshal(io)
	io.String(&x.PatternNameID)
	io.Uint8(&x.NumCrafts)
	Minimum(io, &x.NumCrafts, 1)
}

type CraftRepairAndDisenchantStackRequestAction struct {
	ActionType              ItemStackRequestActionType
	RecipeNetID             int32
	NumberOfRequestedCrafts uint8
	RepairCost              int32
}

func (*CraftRepairAndDisenchantStackRequestAction) tagStackRequestAction() uint32 { return 14 }

// Marshal reads or writes CraftRepairAndDisenchantStackRequestAction using its canonical wire layout.
func (x *CraftRepairAndDisenchantStackRequestAction) Marshal(io IO) {
	x.ActionType.Marshal(io)
	io.Int32(&x.RecipeNetID)
	io.Uint8(&x.NumberOfRequestedCrafts)
	Minimum(io, &x.NumberOfRequestedCrafts, 1)
	io.Varint32(&x.RepairCost)
	Minimum(io, &x.RepairCost, 0)
}

// CylinderShape represents a cylinder debug shape.
type CylinderData struct {
	// RadiusX is the radius of the cylinder along the X axis.
	RadiusX mgl32.Vec2
	// RadiusZ is the radius of the cylinder along the Z axis.
	RadiusZ mgl32.Vec2
	// Height is the height of the cylinder.
	Height float32
	// NumSegments is the number of segments used for the cylinder.
	NumSegments uint8
}

func (*CylinderData) tagPrimitiveShapeExtraShapeData() uint32 { return 6 }

// Marshal reads or writes CylinderData using its canonical wire layout.
func (x *CylinderData) Marshal(io IO) {
	io.Vec2(&x.RadiusX)
	io.Vec2(&x.RadiusZ)
	io.Float32(&x.Height)
	io.Uint8(&x.NumSegments)
}

type DataItemByte struct {
	Type  DataItemType
	Value int8
}

func (*DataItemByte) tagDataItemEntryValue() uint32 { return 0 }

// Marshal reads or writes DataItemByte using its canonical wire layout.
func (x *DataItemByte) Marshal(io IO) {
	x.Type.Marshal(io)
	io.Int8(&x.Value)
}

type DataItemCompoundTag struct {
	Type  DataItemType
	Value []byte
}

func (*DataItemCompoundTag) tagDataItemEntryValue() uint32 { return 5 }

// Marshal reads or writes DataItemCompoundTag using its canonical wire layout.
func (x *DataItemCompoundTag) Marshal(io IO) {
	x.Type.Marshal(io)
	io.NBT(&x.Value, NBTNetwork)
}

type DataItemEntry struct {
	ID      uint32
	Payload DataItemEntryValue
}

// Marshal reads or writes DataItemEntry using its canonical wire layout.
func (x *DataItemEntry) Marshal(io IO) {
	io.Varuint32(&x.ID)
	MarshalDataItemEntryValue(io, &x.Payload)
}

type DataItemFloat struct {
	Type  DataItemType
	Value float32
}

func (*DataItemFloat) tagDataItemEntryValue() uint32 { return 3 }

// Marshal reads or writes DataItemFloat using its canonical wire layout.
func (x *DataItemFloat) Marshal(io IO) {
	x.Type.Marshal(io)
	io.Float32(&x.Value)
}

type DataItemInt struct {
	Type  DataItemType
	Value int32
}

func (*DataItemInt) tagDataItemEntryValue() uint32 { return 2 }

// Marshal reads or writes DataItemInt using its canonical wire layout.
func (x *DataItemInt) Marshal(io IO) {
	x.Type.Marshal(io)
	io.Varint32(&x.Value)
}

type DataItemInt64 struct {
	Type  DataItemType
	Value int64
}

func (*DataItemInt64) tagDataItemEntryValue() uint32 { return 7 }

// Marshal reads or writes DataItemInt64 using its canonical wire layout.
func (x *DataItemInt64) Marshal(io IO) {
	x.Type.Marshal(io)
	io.Varint64(&x.Value)
}

type DataItemPos struct {
	Type  DataItemType
	Value BlockPos
}

func (*DataItemPos) tagDataItemEntryValue() uint32 { return 6 }

// Marshal reads or writes DataItemPos using its canonical wire layout.
func (x *DataItemPos) Marshal(io IO) {
	x.Type.Marshal(io)
	x.Value.Marshal(io)
}

type DataItemShort struct {
	Type  DataItemType
	Value int16
}

func (*DataItemShort) tagDataItemEntryValue() uint32 { return 1 }

// Marshal reads or writes DataItemShort using its canonical wire layout.
func (x *DataItemShort) Marshal(io IO) {
	x.Type.Marshal(io)
	io.Int16(&x.Value)
}

type DataItemString struct {
	Type  DataItemType
	Value string
}

func (*DataItemString) tagDataItemEntryValue() uint32 { return 4 }

// Marshal reads or writes DataItemString using its canonical wire layout.
func (x *DataItemString) Marshal(io IO) {
	x.Type.Marshal(io)
	io.String(&x.Value)
}

type DataItemType uint8

const (
	DataItemTypeByte        DataItemType = 0
	DataItemTypeShort       DataItemType = 1
	DataItemTypeInt         DataItemType = 2
	DataItemTypeFloat       DataItemType = 3
	DataItemTypeString      DataItemType = 4
	DataItemTypeCompoundTag DataItemType = 5
	DataItemTypePos         DataItemType = 6
	DataItemTypeInt64       DataItemType = 7
	DataItemTypeVec3        DataItemType = 8
)

// Marshal reads or writes DataItemType through its uint8 wire encoding.
func (x *DataItemType) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type DataItemVec3 struct {
	Type  DataItemType
	Value mgl32.Vec3
}

func (*DataItemVec3) tagDataItemEntryValue() uint32 { return 8 }

// Marshal reads or writes DataItemVec3 using its canonical wire layout.
func (x *DataItemVec3) Marshal(io IO) {
	x.Type.Marshal(io)
	io.Vec3(&x.Value)
}

type DebugMarkerData struct {
	Text     string
	Position mgl32.Vec3
	Color    color.RGBA
	Duration uint64
}

// Marshal reads or writes DebugMarkerData using its canonical wire layout.
func (x *DebugMarkerData) Marshal(io IO) {
	io.StringLimits(&x.Text, 0, 4096)
	io.Vec3(&x.Position)
	io.RGBA(&x.Color)
	io.Uint64(&x.Duration)
}

type Difficulty int32

const (
	DifficultyPeaceful Difficulty = 0
	DifficultyEasy     Difficulty = 1
	DifficultyNormal   Difficulty = 2
	DifficultyHard     Difficulty = 3
	DifficultyCount    Difficulty = 4
	DifficultyUnknown  Difficulty = 5
)

// Marshal reads or writes Difficulty through its int32 wire encoding.
func (x *Difficulty) Marshal(io IO) { io.Varint32((*int32)(x)) }

type DimensionType struct {
	Value int32
}

// Marshal reads or writes DimensionType using its canonical wire layout.
func (x *DimensionType) Marshal(io IO) {
	io.Varint32(&x.Value)
}

type DisconnectMessagesData struct {
	Message         string
	FilteredMessage string
}

func (*DisconnectMessagesData) tagDisconnectMessages() uint32 { return 0 }

// Marshal reads or writes DisconnectMessagesData using its canonical wire layout.
func (x *DisconnectMessagesData) Marshal(io IO) {
	io.String(&x.Message)
	io.String(&x.FilteredMessage)
}

type DynamicValue interface {
	Marshaler
	tagDynamicValue() int32
}

// MarshalDynamicValue reads or writes the DynamicValue union using its canonical wire layout.
func MarshalDynamicValue(io IO, x *DynamicValue) {
	Union(io, x, io.Int32, DynamicValue.tagDynamicValue, func(tag int32) DynamicValue {
		switch tag {
		case 0:
			return new(DynamicValueNone)
		case 1:
			return new(DynamicValueBool)
		case 2:
			return new(DynamicValueInt64)
		case 3:
			return new(DynamicValueDouble)
		case 4:
			return new(DynamicValueString)
		case 5:
			return new(DynamicValueList)
		case 6:
			return new(DynamicValueMap)
		}
		return nil
	})
}

type EASAttributeLayerData struct {
	Name       string
	Dimension  DimensionType
	Settings   EASAttributeLayerSettings
	Attributes []EASEnvironmentAttributeData
}

// Marshal reads or writes EASAttributeLayerData using its canonical wire layout.
func (x *EASAttributeLayerData) Marshal(io IO) {
	io.StringLimits(&x.Name, 0, 128)
	x.Dimension.Marshal(io)
	x.Settings.Marshal(io)
	SliceLimits(io, &x.Attributes, 0, 1024)
}

type EASAttributeLayerSettings struct {
	Priority          int32
	Weight            float32
	Enabled           bool
	TransitionsPaused bool
}

// Marshal reads or writes EASAttributeLayerSettings using its canonical wire layout.
func (x *EASAttributeLayerSettings) Marshal(io IO) {
	io.Int32(&x.Priority)
	io.Float32(&x.Weight)
	io.Bool(&x.Enabled)
	io.Bool(&x.TransitionsPaused)
}

type EASEnvironmentAttributeData struct {
	AttributeName string
	Payload       EnvironmentAttribute
}

// Marshal reads or writes EASEnvironmentAttributeData using its canonical wire layout.
func (x *EASEnvironmentAttributeData) Marshal(io IO) {
	io.StringLimits(&x.AttributeName, 0, 128)
	MarshalEnvironmentAttribute(io, &x.Payload)
}

// EntityDiagnosticTimingInfo represents diagnostics for a specific entity type.
type ECSProfilingDiagnosticsEntityDiagnosticTimingInfo struct {
	// DisplayName is the name to display for this timing entry.
	DisplayName string
	// Entity is the identifier of the entity that is being timed.
	Entity string
	// DurationNanos is whole long the timing entry has lasted, in nanoseconds.
	TimeInNS uint64
	// PercentOfTotal is the percentage of time that this timing entry has used compared to others.
	PercentOfTotal uint8
	// Position is the position of the entity that is being timed.
	Position Optional[mgl32.Vec3]
	// Dimension is the name of the dimension that the entity being timed is in.
	Dimension Optional[string]
}

// Marshal reads or writes ECSProfilingDiagnosticsEntityDiagnosticTimingInfo using its canonical wire layout.
func (x *ECSProfilingDiagnosticsEntityDiagnosticTimingInfo) Marshal(io IO) {
	io.String(&x.DisplayName)
	io.String(&x.Entity)
	io.Uint64(&x.TimeInNS)
	io.Uint8(&x.PercentOfTotal)
	OptionalFunc(io, &x.Position, io.Vec3)
	OptionalFunc(io, &x.Dimension, io.String)
}

// SystemCategory maps a diagnostics category name to a system index.
type ECSProfilingDiagnosticsSystemCategory struct {
	CategoryName string
	SystemIndex  uint64
}

// Marshal reads or writes ECSProfilingDiagnosticsSystemCategory using its canonical wire layout.
func (x *ECSProfilingDiagnosticsSystemCategory) Marshal(io IO) {
	io.String(&x.CategoryName)
	io.Uint64(&x.SystemIndex)
}

// SystemDiagnosticTimingInfo represents diagnostics for a specific system index.
type ECSProfilingDiagnosticsSystemDiagnosticTimingInfo struct {
	// DisplayName is the name to display for this timing entry.
	DisplayName string
	// SystemIndex is the index of the system that is being timed.
	SystemIndex uint64
	// DurationNanos is whole long the timing entry has lasted, in nanoseconds.
	TimeInNS uint64
	// PercentOfTotal is the percentage of time that this timing entry has used compared to others.
	PercentOfTotal uint8
}

// Marshal reads or writes ECSProfilingDiagnosticsSystemDiagnosticTimingInfo using its canonical wire layout.
func (x *ECSProfilingDiagnosticsSystemDiagnosticTimingInfo) Marshal(io IO) {
	io.String(&x.DisplayName)
	io.Uint64(&x.SystemIndex)
	io.Uint64(&x.TimeInNS)
	io.Uint8(&x.PercentOfTotal)
}

type EducationSharedResourceURI struct {
	ButtonName string
	LinkURI    string
}

// Marshal reads or writes EducationSharedResourceURI using its canonical wire layout.
func (x *EducationSharedResourceURI) Marshal(io IO) {
	io.String(&x.ButtonName)
	io.String(&x.LinkURI)
}

// EllipsoidShape represents an ellipsoid debug shape.
type EllipsoidData struct {
	// Radii are the radii of the ellipsoid along the X, Y and Z axes.
	Radii mgl32.Vec3
	// SegmentsPerAxis is the number of segments used per axis for the ellipsoid.
	SegmentsPerAxis uint8
}

func (*EllipsoidData) tagPrimitiveShapeExtraShapeData() uint32 { return 8 }

// Marshal reads or writes EllipsoidData using its canonical wire layout.
func (x *EllipsoidData) Marshal(io IO) {
	io.Vec3(&x.Radii)
	io.Uint8(&x.SegmentsPerAxis)
}

type Empty struct {
}

func (*Empty) tagEventData() uint32 { return 21 }

// Marshal reads or writes Empty using its canonical wire layout.
func (x *Empty) Marshal(io IO) {
}

type Experiments struct {
	Toggles                []ExperimentToggle
	ExperimentsEverToggled bool
}

// Marshal reads or writes Experiments using its canonical wire layout.
func (x *Experiments) Marshal(io IO) {
	FuncSlice(io, &x.Toggles, io.Uint32, func(value *ExperimentToggle) {
		value.Marshal(io)
	})
	io.Bool(&x.ExperimentsEverToggled)
}

// EducationExternalLinkSettings ...
type ExternalLinkSettings struct {
	// URL is the external link URL.
	URL string
	// DisplayName is the display name in game.
	DisplayName string
}

// Marshal reads or writes ExternalLinkSettings using its canonical wire layout.
func (x *ExternalLinkSettings) Marshal(io IO) {
	io.String(&x.URL)
	io.String(&x.DisplayName)
}

type FeatureRegistryFeatureBinaryJSONFormat struct {
	FeatureName      string
	BinaryJSONOutput []byte
}

// Marshal reads or writes FeatureRegistryFeatureBinaryJSONFormat using its canonical wire layout.
func (x *FeatureRegistryFeatureBinaryJSONFormat) Marshal(io IO) {
	io.String(&x.FeatureName)
	io.ByteSlice(&x.BinaryJSONOutput)
}

type FloatAttributeValue struct {
	Value         float32
	Operation     string
	ConstraintMin Optional[float32]
	ConstraintMax Optional[float32]
}

func (*FloatAttributeValue) tagAttributeValue() uint32 { return 1 }

// Marshal reads or writes FloatAttributeValue using its canonical wire layout.
func (x *FloatAttributeValue) Marshal(io IO) {
	io.Float32(&x.Value)
	io.String(&x.Operation)
	OptionalFunc(io, &x.ConstraintMin, io.Float32)
	OptionalFunc(io, &x.ConstraintMax, io.Float32)
}

type FloatOverride struct {
	Type  string
	Value float32
}

func (*FloatOverride) tagPlayerUpdateEntityOverridesData() uint32 { return 3 }

// Marshal reads or writes FloatOverride using its canonical wire layout.
func (x *FloatOverride) Marshal(io IO) {
	io.String(&x.Type)
	io.Float32(&x.Value)
}

type GamePublishSetting int32

const (
	GamePublishSettingNoMultiPlay      GamePublishSetting = 0
	GamePublishSettingInviteOnly       GamePublishSetting = 1
	GamePublishSettingFriendsOnly      GamePublishSetting = 2
	GamePublishSettingFriendsOfFriends GamePublishSetting = 3
	GamePublishSettingPublic           GamePublishSetting = 4
)

// Marshal reads or writes GamePublishSetting through its int32 wire encoding.
func (x *GamePublishSetting) Marshal(io IO) { io.Varint32((*int32)(x)) }

type GameType int32

const (
	GameTypeSurvival  GameType = 0
	GameTypeCreative  GameType = 1
	GameTypeAdventure GameType = 2
	GameTypeDefault   GameType = 5
	GameTypeSpectator GameType = 6
)

// Marshal reads or writes GameType through its int32 wire encoding.
func (x *GameType) Marshal(io IO) { io.Varint32((*int32)(x)) }

type GeneratorType int32

const (
	GeneratorTypeLegacy    GeneratorType = 0
	GeneratorTypeOverworld GeneratorType = 1
	GeneratorTypeFlat      GeneratorType = 2
	GeneratorTypeNether    GeneratorType = 3
	GeneratorTypeTheEnd    GeneratorType = 4
	GeneratorTypeVoid      GeneratorType = 5
	GeneratorTypeUndefined GeneratorType = 6
)

// Marshal reads or writes GeneratorType through its int32 wire encoding.
func (x *GeneratorType) Marshal(io IO) { io.Varint32((*int32)(x)) }

type GraphicsMode uint8

const (
	GraphicsModeSimple    GraphicsMode = 0
	GraphicsModeFancy     GraphicsMode = 1
	GraphicsModeAdvanced  GraphicsMode = 2
	GraphicsModeRayTraced GraphicsMode = 3
)

// Marshal reads or writes GraphicsMode through its uint8 wire encoding.
func (x *GraphicsMode) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type GraphicsOverrideParameterType uint8

const (
	GraphicsOverrideParameterTypeSkyZenithColor          GraphicsOverrideParameterType = 0
	GraphicsOverrideParameterTypeSkyHorizonColor         GraphicsOverrideParameterType = 1
	GraphicsOverrideParameterTypeHorizonBlendMin         GraphicsOverrideParameterType = 2
	GraphicsOverrideParameterTypeHorizonBlendMax         GraphicsOverrideParameterType = 3
	GraphicsOverrideParameterTypeHorizonBlendStart       GraphicsOverrideParameterType = 4
	GraphicsOverrideParameterTypeHorizonBlendMieStart    GraphicsOverrideParameterType = 5
	GraphicsOverrideParameterTypeRayleighStrength        GraphicsOverrideParameterType = 6
	GraphicsOverrideParameterTypeSunMieStrength          GraphicsOverrideParameterType = 7
	GraphicsOverrideParameterTypeMoonMieStrength         GraphicsOverrideParameterType = 8
	GraphicsOverrideParameterTypeSunGlareShape           GraphicsOverrideParameterType = 9
	GraphicsOverrideParameterTypeChlorophyll             GraphicsOverrideParameterType = 10
	GraphicsOverrideParameterTypeCdom                    GraphicsOverrideParameterType = 11
	GraphicsOverrideParameterTypeSuspendedSediment       GraphicsOverrideParameterType = 12
	GraphicsOverrideParameterTypeWavesDepth              GraphicsOverrideParameterType = 13
	GraphicsOverrideParameterTypeWavesFrequency          GraphicsOverrideParameterType = 14
	GraphicsOverrideParameterTypeWavesFrequencyScaling   GraphicsOverrideParameterType = 15
	GraphicsOverrideParameterTypeWavesSpeed              GraphicsOverrideParameterType = 16
	GraphicsOverrideParameterTypeWavesSpeedScaling       GraphicsOverrideParameterType = 17
	GraphicsOverrideParameterTypeWavesShape              GraphicsOverrideParameterType = 18
	GraphicsOverrideParameterTypeWavesOctaves            GraphicsOverrideParameterType = 19
	GraphicsOverrideParameterTypeWavesMix                GraphicsOverrideParameterType = 20
	GraphicsOverrideParameterTypeWavesPull               GraphicsOverrideParameterType = 21
	GraphicsOverrideParameterTypeWavesDirectionIncrement GraphicsOverrideParameterType = 22
	GraphicsOverrideParameterTypeMidtonesContrast        GraphicsOverrideParameterType = 23
	GraphicsOverrideParameterTypeHighlightsContrast      GraphicsOverrideParameterType = 24
	GraphicsOverrideParameterTypeShadowsContrast         GraphicsOverrideParameterType = 25
	GraphicsOverrideParameterTypeHighlightsGain          GraphicsOverrideParameterType = 26
	GraphicsOverrideParameterTypeHighlightsGamma         GraphicsOverrideParameterType = 27
	GraphicsOverrideParameterTypeHighlightsOffset        GraphicsOverrideParameterType = 28
	GraphicsOverrideParameterTypeHighlightsSaturation    GraphicsOverrideParameterType = 29
	GraphicsOverrideParameterTypeMidtonesGain            GraphicsOverrideParameterType = 30
	GraphicsOverrideParameterTypeMidtonesGamma           GraphicsOverrideParameterType = 31
	GraphicsOverrideParameterTypeMidtonesOffset          GraphicsOverrideParameterType = 32
	GraphicsOverrideParameterTypeMidtonesSaturation      GraphicsOverrideParameterType = 33
	GraphicsOverrideParameterTypeShadowsGain             GraphicsOverrideParameterType = 34
	GraphicsOverrideParameterTypeShadowsGamma            GraphicsOverrideParameterType = 35
	GraphicsOverrideParameterTypeShadowsOffset           GraphicsOverrideParameterType = 36
	GraphicsOverrideParameterTypeShadowsSaturation       GraphicsOverrideParameterType = 37
	GraphicsOverrideParameterTypeHighlightsMin           GraphicsOverrideParameterType = 38
	GraphicsOverrideParameterTypeShadowsMax              GraphicsOverrideParameterType = 39
	GraphicsOverrideParameterTypeTemperature             GraphicsOverrideParameterType = 40
	GraphicsOverrideParameterTypeSunColor                GraphicsOverrideParameterType = 41
	GraphicsOverrideParameterTypeSunIlluminance          GraphicsOverrideParameterType = 42
	GraphicsOverrideParameterTypeMoonColor               GraphicsOverrideParameterType = 43
	GraphicsOverrideParameterTypeMoonIlluminance         GraphicsOverrideParameterType = 44
	GraphicsOverrideParameterTypeFlashColor              GraphicsOverrideParameterType = 45
	GraphicsOverrideParameterTypeFlashIlluminance        GraphicsOverrideParameterType = 46
	GraphicsOverrideParameterTypeAmbientColor            GraphicsOverrideParameterType = 47
	GraphicsOverrideParameterTypeAmbientIlluminance      GraphicsOverrideParameterType = 48
	GraphicsOverrideParameterTypeEmissiveDesaturation    GraphicsOverrideParameterType = 49
	GraphicsOverrideParameterTypeSkyIntensity            GraphicsOverrideParameterType = 50
	GraphicsOverrideParameterTypeOrbitalOffsetDegrees    GraphicsOverrideParameterType = 51
)

// Marshal reads or writes GraphicsOverrideParameterType through its uint8 wire encoding.
func (x *GraphicsOverrideParameterType) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type HeightMapDataType uint8

const (
	HeightMapDataTypeNoData     HeightMapDataType = 0
	HeightMapDataTypeHasData    HeightMapDataType = 1
	HeightMapDataTypeAllTooHigh HeightMapDataType = 2
	HeightMapDataTypeAllTooLow  HeightMapDataType = 3
	HeightMapDataTypeAllCopied  HeightMapDataType = 4
)

// Marshal reads or writes HeightMapDataType through its uint8 wire encoding.
func (x *HeightMapDataType) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type HeightmapData struct {
	HeightMapType           HeightMapDataType
	SubchunkHeightMap       Optional[[16][]int8]
	RenderHeightMapType     HeightMapDataType
	SubchunkRenderHeightMap Optional[[16][]int8]
}

// Marshal reads or writes HeightmapData using its canonical wire layout.
func (x *HeightmapData) Marshal(io IO) {
	x.HeightMapType.Marshal(io)
	OptionalFunc(io, &x.SubchunkHeightMap, func(value *[16][]int8) {
		for index1 := range *value {
			FuncSliceLimits(io, &(*value)[index1], io.Varuint32, 16, 16, io.Int8)
		}
	})
	x.RenderHeightMapType.Marshal(io)
	OptionalFunc(io, &x.SubchunkRenderHeightMap, func(value *[16][]int8) {
		for index2 := range *value {
			FuncSliceLimits(io, &(*value)[index2], io.Varuint32, 16, 16, io.Int8)
		}
	})
}

type HiddenLocation struct {
	PacketType PlayerLocationType
}

func (*HiddenLocation) tagPlayerLocationData() uint32 { return 1 }

// Marshal reads or writes HiddenLocation using its canonical wire layout.
func (x *HiddenLocation) Marshal(io IO) {
	x.PacketType.Marshal(io)
}

type HudElement int32

const (
	HudElementPaperDoll     HudElement = 0
	HudElementArmor         HudElement = 1
	HudElementToolTips      HudElement = 2
	HudElementTouchControls HudElement = 3
	HudElementCrosshair     HudElement = 4
	HudElementHotBar        HudElement = 5
	HudElementHealth        HudElement = 6
	HudElementProgressBar   HudElement = 7
	HudElementHunger        HudElement = 8
	HudElementAirBubbles    HudElement = 9
	HudElementHorseHealth   HudElement = 10
	HudElementStatusEffects HudElement = 11
	HudElementItemText      HudElement = 12
)

// Marshal reads or writes HudElement through its int32 wire encoding.
func (x *HudElement) Marshal(io IO) { io.Varint32((*int32)(x)) }

type HudVisibility int32

const (
	HudVisibilityHide  HudVisibility = 0
	HudVisibilityReset HudVisibility = 1
)

// Marshal reads or writes HudVisibility through its int32 wire encoding.
func (x *HudVisibility) Marshal(io IO) { io.Varint32((*int32)(x)) }

type InitializeRegistryData struct {
	ClockData []WorldClockData
}

func (*InitializeRegistryData) tagSyncWorldClocksData() uint32 { return 1 }

// Marshal reads or writes InitializeRegistryData using its canonical wire layout.
func (x *InitializeRegistryData) Marshal(io IO) {
	SliceLimits(io, &x.ClockData, 0, 256)
}

type InputData int32

const (
	InputDataAscend                          InputData = 0
	InputDataDescend                         InputData = 1
	InputDataNorthJump                       InputData = 2
	InputDataJumpDown                        InputData = 3
	InputDataSprintDown                      InputData = 4
	InputDataChangeHeight                    InputData = 5
	InputDataJumping                         InputData = 6
	InputDataAutoJumpingInWater              InputData = 7
	InputDataSneaking                        InputData = 8
	InputDataSneakDown                       InputData = 9
	InputDataUp                              InputData = 10
	InputDataDown                            InputData = 11
	InputDataLeft                            InputData = 12
	InputDataRight                           InputData = 13
	InputDataUpLeft                          InputData = 14
	InputDataUpRight                         InputData = 15
	InputDataWantUp                          InputData = 16
	InputDataWantDown                        InputData = 17
	InputDataWantDownSlow                    InputData = 18
	InputDataWantUpSlow                      InputData = 19
	InputDataSprinting                       InputData = 20
	InputDataAscendBlock                     InputData = 21
	InputDataDescendBlock                    InputData = 22
	InputDataSneakToggleDown                 InputData = 23
	InputDataPersistSneak                    InputData = 24
	InputDataStartSprinting                  InputData = 25
	InputDataStopSprinting                   InputData = 26
	InputDataStartSneaking                   InputData = 27
	InputDataStopSneaking                    InputData = 28
	InputDataStartSwimming                   InputData = 29
	InputDataStopSwimming                    InputData = 30
	InputDataStartJumping                    InputData = 31
	InputDataStartGliding                    InputData = 32
	InputDataStopGliding                     InputData = 33
	InputDataPerformItemInteraction          InputData = 34
	InputDataPerformBlockActions             InputData = 35
	InputDataPerformItemStackRequest         InputData = 36
	InputDataHandledTeleport                 InputData = 37
	InputDataEmoting                         InputData = 38
	InputDataMissedSwing                     InputData = 39
	InputDataStartCrawling                   InputData = 40
	InputDataStopCrawling                    InputData = 41
	InputDataStartFlying                     InputData = 42
	InputDataStopFlying                      InputData = 43
	InputDataClientAckServerData             InputData = 44
	InputDataIsInClientPredictedVehicle      InputData = 45
	InputDataPaddlingLeft                    InputData = 46
	InputDataPaddlingRight                   InputData = 47
	InputDataBlockBreakingDelayEnabled       InputData = 48
	InputDataHorizontalCollision             InputData = 49
	InputDataVerticalCollision               InputData = 50
	InputDataDownLeft                        InputData = 51
	InputDataDownRight                       InputData = 52
	InputDataStartUsingItem                  InputData = 53
	InputDataIsCameraRelativeMovementEnabled InputData = 54
	InputDataIsRotControlledByMoveDirection  InputData = 55
	InputDataStartSpinAttack                 InputData = 56
	InputDataStopSpinAttack                  InputData = 57
	InputDataIsHotbarOnlyTouch               InputData = 58
	InputDataJumpReleasedRaw                 InputData = 59
	InputDataJumpPressedRaw                  InputData = 60
	InputDataJumpCurrentRaw                  InputData = 61
	InputDataSneakReleasedRaw                InputData = 62
	InputDataSneakPressedRaw                 InputData = 63
	InputDataSneakCurrentRaw                 InputData = 64
	InputDataInternalUpdate                  InputData = 65
)

// Marshal reads or writes InputData through its int32 wire encoding.
func (x *InputData) Marshal(io IO) { io.Varint32((*int32)(x)) }

type InputMode uint32

const (
	InputModeUndefined        InputMode = 0
	InputModeMouse            InputMode = 1
	InputModeTouch            InputMode = 2
	InputModeGamePad          InputMode = 3
	InputModeMotionController InputMode = 4
	InputModeCount            InputMode = 5
)

// Marshal reads or writes InputMode through its uint32 wire encoding.
func (x *InputMode) Marshal(io IO) { io.Varuint32((*uint32)(x)) }

type IntOverride struct {
	Type  string
	Value int32
}

func (*IntOverride) tagPlayerUpdateEntityOverridesData() uint32 { return 2 }

// Marshal reads or writes IntOverride using its canonical wire layout.
func (x *IntOverride) Marshal(io IO) {
	io.String(&x.Type)
	io.Int32(&x.Value)
}

type InteractAction uint8

const (
	InteractActionInvalid        InteractAction = 0
	InteractActionStopRiding     InteractAction = 3
	InteractActionInteractUpdate InteractAction = 4
	InteractActionNpcOpen        InteractAction = 5
	InteractActionOpenInventory  InteractAction = 6
)

// Marshal reads or writes InteractAction through its uint8 wire encoding.
func (x *InteractAction) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type Interaction struct {
	InteractedEntityID      int64
	InteractionType         MinecraftEventingInteractionType
	InteractionActorType    int32
	InteractionActorVariant int32
	InteractionActorColor   uint8
}

func (*Interaction) tagEventData() uint32 { return 1 }

// Marshal reads or writes Interaction using its canonical wire layout.
func (x *Interaction) Marshal(io IO) {
	io.Varint64(&x.InteractedEntityID)
	x.InteractionType.Marshal(io)
	io.Varint32(&x.InteractionActorType)
	io.Varint32(&x.InteractionActorVariant)
	io.Uint8(&x.InteractionActorColor)
}

type LabTableReactionType uint8

const (
	LabTableReactionTypeNone               LabTableReactionType = 0
	LabTableReactionTypeIceBomb            LabTableReactionType = 1
	LabTableReactionTypeBleach             LabTableReactionType = 2
	LabTableReactionTypeElephantToothpaste LabTableReactionType = 3
	LabTableReactionTypeFertilizer         LabTableReactionType = 4
	LabTableReactionTypeHeatBlock          LabTableReactionType = 5
	LabTableReactionTypeMagnesiumSalts     LabTableReactionType = 6
	LabTableReactionTypeMiscFire           LabTableReactionType = 7
	LabTableReactionTypeMiscExplosion      LabTableReactionType = 8
	LabTableReactionTypeMiscLava           LabTableReactionType = 9
	LabTableReactionTypeMiscMystical       LabTableReactionType = 10
	LabTableReactionTypeMiscSmoke          LabTableReactionType = 11
	LabTableReactionTypeMiscLargeSmoke     LabTableReactionType = 12
)

// Marshal reads or writes LabTableReactionType through its uint8 wire encoding.
func (x *LabTableReactionType) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type LabTableType uint8

const (
	LabTableTypeStartCombine  LabTableType = 0
	LabTableTypeStartReaction LabTableType = 1
	LabTableTypeReset         LabTableType = 2
)

// Marshal reads or writes LabTableType through its uint8 wire encoding.
func (x *LabTableType) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type LegacyArmorSlot int32

const (
	LegacyArmorSlotHead  LegacyArmorSlot = 0
	LegacyArmorSlotTorso LegacyArmorSlot = 1
	LegacyArmorSlotLegs  LegacyArmorSlot = 2
	LegacyArmorSlotFeet  LegacyArmorSlot = 3
	LegacyArmorSlotBody  LegacyArmorSlot = 4
)

// Marshal reads or writes LegacyArmorSlot through its int32 wire encoding.
func (x *LegacyArmorSlot) Marshal(io IO) { io.Varint32((*int32)(x)) }

type LegacySetSlot struct {
	ContainerEnum ContainerEnumName
	Slots         []uint8
}

// Marshal reads or writes LegacySetSlot using its canonical wire layout.
func (x *LegacySetSlot) Marshal(io IO) {
	x.ContainerEnum.Marshal(io)
	io.ByteSlice(&x.Slots)
}

type LevelExperiments struct {
	Toggles                []LevelExperimentToggle
	ExperimentsEverToggled bool
}

// Marshal reads or writes LevelExperiments using its canonical wire layout.
func (x *LevelExperiments) Marshal(io IO) {
	FuncSlice(io, &x.Toggles, io.Uint32, func(value *LevelExperimentToggle) {
		value.Marshal(io)
	})
	io.Bool(&x.ExperimentsEverToggled)
}

type LevelGameType int32

const (
	LevelGameTypeUndefined    LevelGameType = -1
	LevelGameTypeSurvival     LevelGameType = 0
	LevelGameTypeWorldDefault LevelGameType = 0
	LevelGameTypeCreative     LevelGameType = 1
	LevelGameTypeAdventure    LevelGameType = 2
	LevelGameTypeDefault      LevelGameType = 5
	LevelGameTypeSpectator    LevelGameType = 6
)

// Marshal reads or writes LevelGameType through its int32 wire encoding.
func (x *LevelGameType) Marshal(io IO) { io.Varint32((*int32)(x)) }

type LevelGeneratorType int32

const (
	LevelGeneratorTypeLegacy    LevelGeneratorType = 0
	LevelGeneratorTypeOverworld LevelGeneratorType = 1
	LevelGeneratorTypeFlat      LevelGeneratorType = 2
	LevelGeneratorTypeNether    LevelGeneratorType = 3
	LevelGeneratorTypeTheEnd    LevelGeneratorType = 4
	LevelGeneratorTypeVoid      LevelGeneratorType = 5
	LevelGeneratorTypeUndefined LevelGeneratorType = 6
)

// Marshal reads or writes LevelGeneratorType through its int32 wire encoding.
func (x *LevelGeneratorType) Marshal(io IO) { io.Varint32((*int32)(x)) }

type LevelSettings struct {
	Seed                                   uint64
	SpawnSettings                          SpawnSettings
	GeneratorType                          LevelGeneratorType
	GameType                               LevelGameType
	IsHardcore                             bool
	GameDifficulty                         Difficulty
	DefaultSpawnBlockPosition              BlockPosition
	AchievementsDisabled                   bool
	EditorWorldType                        EditorWorldType
	IsCreatedInEditor                      bool
	IsExportedFromEditor                   bool
	EditorLevelMigrationVersion            EditorLevelMigrationVersion
	DayCycleStopTime                       int32
	EducationEditionOffer                  EducationEditionOffer
	EducationFeaturesEnabled               bool
	EducationProductID                     string
	RainLevel                              float32
	LightningLevel                         float32
	HasConfirmedPlatformLockedContent      bool
	MultiplayerGameIntent                  bool
	LANBroadcastIntent                     bool
	XboxLiveBroadcastSetting               GamePublishSetting
	PlatformBroadcastSetting               GamePublishSetting
	CommandsEnabled                        bool
	TexturePacksRequired                   bool
	RuleData                               GameRuleData
	Experiments                            LevelExperiments
	HasBonusChestEnabled                   bool
	StartWithMapEnabled                    bool
	PlayerPermissions                      PlayerPermission
	ServerChunkTickRange                   int32
	HasLockedBehaviorPack                  bool
	HasLockedResourcePack                  bool
	IsFromLockedTemplate                   bool
	UseMsaGamertagsOnly                    bool
	IsFromWorldTemplate                    bool
	IsWorldTemplateOptionLocked            bool
	OnlySpawnV1Villagers                   bool
	PersonaDisabled                        bool
	CustomSkinsDisabled                    bool
	EmoteChatMuted                         bool
	BaseGameVersion                        string
	LimitedWorldWidth                      int32
	LimitedWorldDepth                      int32
	NetherType                             bool
	EduSharedURIResource                   EducationSharedResourceURI
	OverrideForceExperimentalGameplay      Optional[bool]
	ChatRestrictionLevel                   ChatRestrictionLevel
	DisablePlayerInteractions              bool
	ServerEditorConnectionPolicy           ServerEditorConnectionPolicy
	AllowAnonymousBlockDropsInEditorWorlds bool
}

// Marshal reads or writes LevelSettings using its canonical wire layout.
func (x *LevelSettings) Marshal(io IO) {
	io.Uint64(&x.Seed)
	x.SpawnSettings.Marshal(io)
	x.GeneratorType.Marshal(io)
	x.GameType.Marshal(io)
	io.Bool(&x.IsHardcore)
	x.GameDifficulty.Marshal(io)
	x.DefaultSpawnBlockPosition.Marshal(io)
	io.Bool(&x.AchievementsDisabled)
	x.EditorWorldType.Marshal(io)
	io.Bool(&x.IsCreatedInEditor)
	io.Bool(&x.IsExportedFromEditor)
	x.EditorLevelMigrationVersion.Marshal(io)
	io.Varint32(&x.DayCycleStopTime)
	x.EducationEditionOffer.Marshal(io)
	io.Bool(&x.EducationFeaturesEnabled)
	io.String(&x.EducationProductID)
	io.Float32(&x.RainLevel)
	io.Float32(&x.LightningLevel)
	io.Bool(&x.HasConfirmedPlatformLockedContent)
	io.Bool(&x.MultiplayerGameIntent)
	io.Bool(&x.LANBroadcastIntent)
	x.XboxLiveBroadcastSetting.Marshal(io)
	x.PlatformBroadcastSetting.Marshal(io)
	io.Bool(&x.CommandsEnabled)
	io.Bool(&x.TexturePacksRequired)
	x.RuleData.Marshal(io)
	x.Experiments.Marshal(io)
	io.Bool(&x.HasBonusChestEnabled)
	io.Bool(&x.StartWithMapEnabled)
	x.PlayerPermissions.Marshal(io)
	io.Int32(&x.ServerChunkTickRange)
	io.Bool(&x.HasLockedBehaviorPack)
	io.Bool(&x.HasLockedResourcePack)
	io.Bool(&x.IsFromLockedTemplate)
	io.Bool(&x.UseMsaGamertagsOnly)
	io.Bool(&x.IsFromWorldTemplate)
	io.Bool(&x.IsWorldTemplateOptionLocked)
	io.Bool(&x.OnlySpawnV1Villagers)
	io.Bool(&x.PersonaDisabled)
	io.Bool(&x.CustomSkinsDisabled)
	io.Bool(&x.EmoteChatMuted)
	io.String(&x.BaseGameVersion)
	io.Int32(&x.LimitedWorldWidth)
	io.Int32(&x.LimitedWorldDepth)
	io.Bool(&x.NetherType)
	x.EduSharedURIResource.Marshal(io)
	OptionalFunc(io, &x.OverrideForceExperimentalGameplay, io.Bool)
	x.ChatRestrictionLevel.Marshal(io)
	io.Bool(&x.DisablePlayerInteractions)
	x.ServerEditorConnectionPolicy.Marshal(io)
	io.Bool(&x.AllowAnonymousBlockDropsInEditorWorlds)
}

type LineData struct {
	LineEndLocation mgl32.Vec3
}

func (*LineData) tagPrimitiveShapeExtraShapeData() uint32 { return 4 }

// Marshal reads or writes LineData using its canonical wire layout.
func (x *LineData) Marshal(io IO) {
	io.Vec3(&x.LineEndLocation)
}

type MaterialReducerDataEntry struct {
	FromItemKey      int32
	ItemIdsAndCounts []MaterialReducerEntryOutput
}

// Marshal reads or writes MaterialReducerDataEntry using its canonical wire layout.
func (x *MaterialReducerDataEntry) Marshal(io IO) {
	io.Varint32(&x.FromItemKey)
	Slice(io, &x.ItemIdsAndCounts)
}

type MaterialReducerEntryOutput struct {
	ItemID    int32
	ItemCount int32
}

// Marshal reads or writes MaterialReducerEntryOutput using its canonical wire layout.
func (x *MaterialReducerEntryOutput) Marshal(io IO) {
	io.Varint32(&x.ItemID)
	io.Varint32(&x.ItemCount)
}

type MessageAndParams struct {
	MessageType   TextPacketType
	Message       string
	ParameterList []string
}

func (*MessageAndParams) tagTextData() uint32 { return 2 }

// Marshal reads or writes MessageAndParams using its canonical wire layout.
func (x *MessageAndParams) Marshal(io IO) {
	x.MessageType.Marshal(io)
	io.StringLimits(&x.Message, 1, 65536)
	FuncSliceLimits(io, &x.ParameterList, io.Varuint32, 0, 4, io.String)
}

type MessageOnly struct {
	MessageType TextPacketType
	Message     string
}

func (*MessageOnly) tagTextData() uint32 { return 0 }

// Marshal reads or writes MessageOnly using its canonical wire layout.
func (x *MessageOnly) Marshal(io IO) {
	x.MessageType.Marshal(io)
	io.StringLimits(&x.Message, 1, 65536)
}

type MinecraftEventingAchievementIds uint8

const (
	MinecraftEventingAchievementIdsChestFullOfCobblestone          MinecraftEventingAchievementIds = 7
	MinecraftEventingAchievementIdsDiamondForYou                   MinecraftEventingAchievementIds = 10
	MinecraftEventingAchievementIdsIronBelly                       MinecraftEventingAchievementIds = 20
	MinecraftEventingAchievementIdsIronMan                         MinecraftEventingAchievementIds = 21
	MinecraftEventingAchievementIdsOnARail                         MinecraftEventingAchievementIds = 29
	MinecraftEventingAchievementIdsOverkill                        MinecraftEventingAchievementIds = 30
	MinecraftEventingAchievementIdsReturnToSender                  MinecraftEventingAchievementIds = 37
	MinecraftEventingAchievementIdsSniperDuel                      MinecraftEventingAchievementIds = 38
	MinecraftEventingAchievementIdsStayinFrosty                    MinecraftEventingAchievementIds = 39
	MinecraftEventingAchievementIdsTakeInventory                   MinecraftEventingAchievementIds = 40
	MinecraftEventingAchievementIdsMapRoom                         MinecraftEventingAchievementIds = 50
	MinecraftEventingAchievementIdsFreightStation                  MinecraftEventingAchievementIds = 52
	MinecraftEventingAchievementIdsSmeltEverything                 MinecraftEventingAchievementIds = 53
	MinecraftEventingAchievementIdsTasteOfYourOwnMedicine          MinecraftEventingAchievementIds = 54
	MinecraftEventingAchievementIdsWhenPigsFly                     MinecraftEventingAchievementIds = 56
	MinecraftEventingAchievementIdsInception                       MinecraftEventingAchievementIds = 58
	MinecraftEventingAchievementIdsArtificialSelection             MinecraftEventingAchievementIds = 60
	MinecraftEventingAchievementIdsFreeDiver                       MinecraftEventingAchievementIds = 61
	MinecraftEventingAchievementIdsSpawnTheWither                  MinecraftEventingAchievementIds = 62
	MinecraftEventingAchievementIdsBeaconator                      MinecraftEventingAchievementIds = 63
	MinecraftEventingAchievementIdsGreatView                       MinecraftEventingAchievementIds = 64
	MinecraftEventingAchievementIdsSuperSonic                      MinecraftEventingAchievementIds = 65
	MinecraftEventingAchievementIdsTheEndAgain                     MinecraftEventingAchievementIds = 66
	MinecraftEventingAchievementIdsTreasureHunter                  MinecraftEventingAchievementIds = 67
	MinecraftEventingAchievementIdsShootingStar                    MinecraftEventingAchievementIds = 68
	MinecraftEventingAchievementIdsFashionShow                     MinecraftEventingAchievementIds = 69
	MinecraftEventingAchievementIdsSelfPublishedAuthor             MinecraftEventingAchievementIds = 71
	MinecraftEventingAchievementIdsAlternativeFuel                 MinecraftEventingAchievementIds = 72
	MinecraftEventingAchievementIdsSleepWithTheFishes              MinecraftEventingAchievementIds = 73
	MinecraftEventingAchievementIdsCastaway                        MinecraftEventingAchievementIds = 74
	MinecraftEventingAchievementIdsImAMarineBiologist              MinecraftEventingAchievementIds = 75
	MinecraftEventingAchievementIdsSailThe7Seas                    MinecraftEventingAchievementIds = 76
	MinecraftEventingAchievementIdsMeGold                          MinecraftEventingAchievementIds = 77
	MinecraftEventingAchievementIdsAhoy                            MinecraftEventingAchievementIds = 78
	MinecraftEventingAchievementIdsAtlantis                        MinecraftEventingAchievementIds = 79
	MinecraftEventingAchievementIdsOnePickleTwoPickleSeaPickleFour MinecraftEventingAchievementIds = 80
	MinecraftEventingAchievementIdsDoaBarrelRoll                   MinecraftEventingAchievementIds = 81
	MinecraftEventingAchievementIdsMoskstraumen                    MinecraftEventingAchievementIds = 82
	MinecraftEventingAchievementIdsEcholocation                    MinecraftEventingAchievementIds = 83
	MinecraftEventingAchievementIdsWhereHaveYouBeen                MinecraftEventingAchievementIds = 84
	MinecraftEventingAchievementIdsTopOfTheWorld                   MinecraftEventingAchievementIds = 85
	MinecraftEventingAchievementIdsFruitOnTheLoom                  MinecraftEventingAchievementIds = 86
	MinecraftEventingAchievementIdsSoundTheAlarm                   MinecraftEventingAchievementIds = 87
	MinecraftEventingAchievementIdsBuyLowSellHigh                  MinecraftEventingAchievementIds = 88
	MinecraftEventingAchievementIdsDisenchanted                    MinecraftEventingAchievementIds = 89
	MinecraftEventingAchievementIdsTimeForStew                     MinecraftEventingAchievementIds = 90
	MinecraftEventingAchievementIdsBeeOurGuest                     MinecraftEventingAchievementIds = 91
	MinecraftEventingAchievementIdsTotalBeeLocation                MinecraftEventingAchievementIds = 92
	MinecraftEventingAchievementIdsStickySituation                 MinecraftEventingAchievementIds = 93
	MinecraftEventingAchievementIdsCoverMeInDebris                 MinecraftEventingAchievementIds = 94
	MinecraftEventingAchievementIdsFloatYourGoat                   MinecraftEventingAchievementIds = 95
	MinecraftEventingAchievementIdsFriend                          MinecraftEventingAchievementIds = 96
	MinecraftEventingAchievementIdsWaxOnWaxOff                     MinecraftEventingAchievementIds = 97
	MinecraftEventingAchievementIdsStriderRiddenInLavaInOverworld  MinecraftEventingAchievementIds = 98
	MinecraftEventingAchievementIdsGoatHornAcquired                MinecraftEventingAchievementIds = 99
	MinecraftEventingAchievementIdsJukeboxUsedInMeadows            MinecraftEventingAchievementIds = 100
	MinecraftEventingAchievementIdsTradedAtWorldHeight             MinecraftEventingAchievementIds = 101
	MinecraftEventingAchievementIdsSurvivedFallFromWorldHeight     MinecraftEventingAchievementIds = 102
	MinecraftEventingAchievementIdsSneakCloseToSculkSensor         MinecraftEventingAchievementIds = 103
	MinecraftEventingAchievementIdsItSpreads                       MinecraftEventingAchievementIds = 104
	MinecraftEventingAchievementIdsBirthdaySong                    MinecraftEventingAchievementIds = 105
	MinecraftEventingAchievementIdsWithOurPowersCombined           MinecraftEventingAchievementIds = 106
	MinecraftEventingAchievementIdsPlantingThePast                 MinecraftEventingAchievementIds = 107
	MinecraftEventingAchievementIdsCarefulRestoration              MinecraftEventingAchievementIds = 108
	MinecraftEventingAchievementIdsRevaulting                      MinecraftEventingAchievementIds = 109
	MinecraftEventingAchievementIdsCraftersCraftingCrafters        MinecraftEventingAchievementIds = 110
	MinecraftEventingAchievementIdsWhoNeedsRockets                 MinecraftEventingAchievementIds = 111
	MinecraftEventingAchievementIdsOverOverkill                    MinecraftEventingAchievementIds = 112
	MinecraftEventingAchievementIdsHeartTransplanter               MinecraftEventingAchievementIds = 113
	MinecraftEventingAchievementIdsStayHydrated                    MinecraftEventingAchievementIds = 114
	MinecraftEventingAchievementIdsMobKabob                        MinecraftEventingAchievementIds = 115
	MinecraftEventingAchievementIdsAdventuringTime                 MinecraftEventingAchievementIds = 116
	MinecraftEventingAchievementIdsUhOh                            MinecraftEventingAchievementIds = 117
	MinecraftEventingAchievementIdsGettingWood                     MinecraftEventingAchievementIds = 118
	MinecraftEventingAchievementIdsBenchMaking                     MinecraftEventingAchievementIds = 119
	MinecraftEventingAchievementIdsTimeToMine                      MinecraftEventingAchievementIds = 120
	MinecraftEventingAchievementIdsHotTopic                        MinecraftEventingAchievementIds = 121
	MinecraftEventingAchievementIdsAcquireHardware                 MinecraftEventingAchievementIds = 122
	MinecraftEventingAchievementIdsGettingAnUpgrade                MinecraftEventingAchievementIds = 123
	MinecraftEventingAchievementIdsMonsterHunter                   MinecraftEventingAchievementIds = 124
	MinecraftEventingAchievementIdsDiamonds                        MinecraftEventingAchievementIds = 125
	MinecraftEventingAchievementIdsPlethoraOfCats                  MinecraftEventingAchievementIds = 126
	MinecraftEventingAchievementIdsWhatGoesAroundComesAround       MinecraftEventingAchievementIds = 127
)

// Marshal reads or writes MinecraftEventingAchievementIds through its uint8 wire encoding.
func (x *MinecraftEventingAchievementIds) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type MinecraftEventingInteractionType uint8

const (
	MinecraftEventingInteractionTypeBreeding   MinecraftEventingInteractionType = 1
	MinecraftEventingInteractionTypeTaming     MinecraftEventingInteractionType = 2
	MinecraftEventingInteractionTypeCuring     MinecraftEventingInteractionType = 3
	MinecraftEventingInteractionTypeCrafted    MinecraftEventingInteractionType = 4
	MinecraftEventingInteractionTypeShearing   MinecraftEventingInteractionType = 5
	MinecraftEventingInteractionTypeMilking    MinecraftEventingInteractionType = 6
	MinecraftEventingInteractionTypeTrading    MinecraftEventingInteractionType = 7
	MinecraftEventingInteractionTypeFeeding    MinecraftEventingInteractionType = 8
	MinecraftEventingInteractionTypeIgniting   MinecraftEventingInteractionType = 9
	MinecraftEventingInteractionTypeColoring   MinecraftEventingInteractionType = 10
	MinecraftEventingInteractionTypeNaming     MinecraftEventingInteractionType = 11
	MinecraftEventingInteractionTypeLeashing   MinecraftEventingInteractionType = 12
	MinecraftEventingInteractionTypeUnleashing MinecraftEventingInteractionType = 13
	MinecraftEventingInteractionTypePetSleep   MinecraftEventingInteractionType = 14
	MinecraftEventingInteractionTypeTrusting   MinecraftEventingInteractionType = 15
	MinecraftEventingInteractionTypeCommanding MinecraftEventingInteractionType = 16
	MinecraftEventingInteractionTypeEquipping  MinecraftEventingInteractionType = 17
)

// Marshal reads or writes MinecraftEventingInteractionType through its uint8 wire encoding.
func (x *MinecraftEventingInteractionType) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type MinecraftEventingPOIBlockInteractionType uint8

const (
	MinecraftEventingPOIBlockInteractionTypeNone                MinecraftEventingPOIBlockInteractionType = 0
	MinecraftEventingPOIBlockInteractionTypeExtend              MinecraftEventingPOIBlockInteractionType = 1
	MinecraftEventingPOIBlockInteractionTypeClone               MinecraftEventingPOIBlockInteractionType = 2
	MinecraftEventingPOIBlockInteractionTypeLock                MinecraftEventingPOIBlockInteractionType = 3
	MinecraftEventingPOIBlockInteractionTypeCreate              MinecraftEventingPOIBlockInteractionType = 4
	MinecraftEventingPOIBlockInteractionTypeCreateLocator       MinecraftEventingPOIBlockInteractionType = 5
	MinecraftEventingPOIBlockInteractionTypeRename              MinecraftEventingPOIBlockInteractionType = 6
	MinecraftEventingPOIBlockInteractionTypeItemPlaced          MinecraftEventingPOIBlockInteractionType = 7
	MinecraftEventingPOIBlockInteractionTypeItemRemoved         MinecraftEventingPOIBlockInteractionType = 8
	MinecraftEventingPOIBlockInteractionTypeCooking             MinecraftEventingPOIBlockInteractionType = 9
	MinecraftEventingPOIBlockInteractionTypeDousing             MinecraftEventingPOIBlockInteractionType = 10
	MinecraftEventingPOIBlockInteractionTypeLighting            MinecraftEventingPOIBlockInteractionType = 11
	MinecraftEventingPOIBlockInteractionTypeHaystack            MinecraftEventingPOIBlockInteractionType = 12
	MinecraftEventingPOIBlockInteractionTypeFilled              MinecraftEventingPOIBlockInteractionType = 13
	MinecraftEventingPOIBlockInteractionTypeEmptied             MinecraftEventingPOIBlockInteractionType = 14
	MinecraftEventingPOIBlockInteractionTypeAddDye              MinecraftEventingPOIBlockInteractionType = 15
	MinecraftEventingPOIBlockInteractionTypeDyeItem             MinecraftEventingPOIBlockInteractionType = 16
	MinecraftEventingPOIBlockInteractionTypeClearItem           MinecraftEventingPOIBlockInteractionType = 17
	MinecraftEventingPOIBlockInteractionTypeEnchantArrow        MinecraftEventingPOIBlockInteractionType = 18
	MinecraftEventingPOIBlockInteractionTypeCompostItemPlaced   MinecraftEventingPOIBlockInteractionType = 19
	MinecraftEventingPOIBlockInteractionTypeRecoveredBonemeal   MinecraftEventingPOIBlockInteractionType = 20
	MinecraftEventingPOIBlockInteractionTypeBookPlaced          MinecraftEventingPOIBlockInteractionType = 21
	MinecraftEventingPOIBlockInteractionTypeBookOpened          MinecraftEventingPOIBlockInteractionType = 22
	MinecraftEventingPOIBlockInteractionTypeDisenchant          MinecraftEventingPOIBlockInteractionType = 23
	MinecraftEventingPOIBlockInteractionTypeRepair              MinecraftEventingPOIBlockInteractionType = 24
	MinecraftEventingPOIBlockInteractionTypeDisenchantAndRepair MinecraftEventingPOIBlockInteractionType = 25
)

// Marshal reads or writes MinecraftEventingPOIBlockInteractionType through its uint8 wire encoding.
func (x *MinecraftEventingPOIBlockInteractionType) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type MissingBlobData struct {
	BlobID   uint64
	BlobData []byte
}

// Marshal reads or writes MissingBlobData using its canonical wire layout.
func (x *MissingBlobData) Marshal(io IO) {
	io.Uint64(&x.BlobID)
	io.ByteSlice(&x.BlobData)
}

type MoLangVersion int16

const (
	MoLangVersionInvalid                                MoLangVersion = -1
	MoLangVersionBeforeVersioning                       MoLangVersion = 0
	MoLangVersionInitial                                MoLangVersion = 1
	MoLangVersionFixedItemRemainingUseDurationQuery     MoLangVersion = 2
	MoLangVersionExpressionErrorMessages                MoLangVersion = 3
	MoLangVersionUnexpectedOperatorErrors               MoLangVersion = 4
	MoLangVersionConditionalOperatorAssociativity       MoLangVersion = 5
	MoLangVersionComparisonAndLogicalOperatorPrecedence MoLangVersion = 6
	MoLangVersionDivideByNegativeValue                  MoLangVersion = 7
	MoLangVersionFixedCapeFlapAmountQuery               MoLangVersion = 8
	MoLangVersionQueryBlockPropertyRenamedToState       MoLangVersion = 9
	MoLangVersionDeprecateOldBlockQueryNames            MoLangVersion = 10
	MoLangVersionDeprecatedSnifferAndCamelQueries       MoLangVersion = 11
	MoLangVersionLeafSupportingInFirstSolidBlockBelow   MoLangVersion = 12
	MoLangVersionLatest                                 MoLangVersion = 13
	MoLangVersionNumValidVersions                       MoLangVersion = 14
)

// Marshal reads or writes MoLangVersion through its int16 wire encoding.
func (x *MoLangVersion) Marshal(io IO) { io.Int16((*int16)(x)) }

// MobBornEvent is the event data sent when a mob is born.
type MobBorn struct {
	// EntityType ...
	BornBabyEntityType int32
	// Variant ...
	BornBabyEntityVariant int32
	// Colour ...
	BornBabyColor uint8
}

func (*MobBorn) tagEventData() uint32 { return 9 }

// Marshal reads or writes MobBorn using its canonical wire layout.
func (x *MobBorn) Marshal(io IO) {
	io.Varint32(&x.BornBabyEntityType)
	io.Varint32(&x.BornBabyEntityVariant)
	io.Uint8(&x.BornBabyColor)
}

type MobEffectEvent uint8

const (
	MobEffectEventInvalid MobEffectEvent = 0
	MobEffectEventAdd     MobEffectEvent = 1
	MobEffectEventUpdate  MobEffectEvent = 2
	MobEffectEventRemove  MobEffectEvent = 3
)

// Marshal reads or writes MobEffectEvent through its uint8 wire encoding.
func (x *MobEffectEvent) Marshal(io IO) { io.Uint8((*uint8)(x)) }

// MobKilledEvent is the event data sent when a mob is killed.
type MobKilled struct {
	// KillerEntityUniqueID ...
	InstigatorActorID int64
	// VictimEntityUniqueID ...
	TargetActorID int64
	// KillerEntityType ...
	InstigatorSChildActorType ActorType
	// EntityDamageCause ...
	DamageSource int32
	// VillagerTradeTier -1 if not a trading actor.
	TradeTier int32
	// VillagerDisplayName Empty if not a trading actor.
	TraderName string
}

func (*MobKilled) tagEventData() uint32 { return 4 }

// Marshal reads or writes MobKilled using its canonical wire layout.
func (x *MobKilled) Marshal(io IO) {
	io.Varint64(&x.InstigatorActorID)
	io.Varint64(&x.TargetActorID)
	x.InstigatorSChildActorType.Marshal(io)
	io.Varint32(&x.DamageSource)
	io.Varint32(&x.TradeTier)
	io.StringLimits(&x.TraderName, 0, 128)
}

type ModalFormCancelReason uint8

const (
	ModalFormCancelReasonUserClosed ModalFormCancelReason = 0
	ModalFormCancelReasonUserBusy   ModalFormCancelReason = 1
)

// Marshal reads or writes ModalFormCancelReason through its uint8 wire encoding.
func (x *ModalFormCancelReason) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type MovePlayerTeleportData struct {
	TeleportationCause int32
	SourceActorType    int32
}

// Marshal reads or writes MovePlayerTeleportData using its canonical wire layout.
func (x *MovePlayerTeleportData) Marshal(io IO) {
	io.Int32(&x.TeleportationCause)
	io.Int32(&x.SourceActorType)
}

type MovementEffectType int32

const (
	MovementEffectTypeGlideBoost   MovementEffectType = 0
	MovementEffectTypeDolphinBoost MovementEffectType = 1
	MovementEffectTypeGeyserBoost  MovementEffectType = 2
)

// Marshal reads or writes MovementEffectType through its int32 wire encoding.
func (x *MovementEffectType) Marshal(io IO) { io.Varint32((*int32)(x)) }

type MultiplayerSettingsType int32

const (
	MultiplayerSettingsTypeEnableMultiplayer  MultiplayerSettingsType = 0
	MultiplayerSettingsTypeDisableMultiplayer MultiplayerSettingsType = 1
	MultiplayerSettingsTypeRefreshJoincode    MultiplayerSettingsType = 2
)

// Marshal reads or writes MultiplayerSettingsType through its int32 wire encoding.
func (x *MultiplayerSettingsType) Marshal(io IO) { io.Varint32((*int32)(x)) }

type NetworkItemInstanceDescriptorSerializedData struct {
	ID             int32
	StackSize      uint16
	AuxValue       uint32
	BlockRuntimeID int32
	UserDataBuffer []byte
}

// Marshal reads or writes NetworkItemInstanceDescriptorSerializedData using its canonical wire layout.
func (x *NetworkItemInstanceDescriptorSerializedData) Marshal(io IO) {
	io.Varint32(&x.ID)
	Minimum(io, &x.ID, -32768)
	Maximum(io, &x.ID, 32767)
	io.Uint16(&x.StackSize)
	io.Varuint32(&x.AuxValue)
	Maximum(io, &x.AuxValue, 32767)
	io.Varint32(&x.BlockRuntimeID)
	io.ByteSlice(&x.UserDataBuffer)
}

type NetworkItemStackDescriptorSerializedData struct {
	ID             int16
	StackSize      uint16
	AuxValue       uint32
	NetIDVariant   Optional[int32]
	BlockRuntimeID uint32
	UserDataBuffer []byte
}

// Marshal reads or writes NetworkItemStackDescriptorSerializedData using its canonical wire layout.
func (x *NetworkItemStackDescriptorSerializedData) Marshal(io IO) {
	io.Int16(&x.ID)
	io.Uint16(&x.StackSize)
	io.Varuint32(&x.AuxValue)
	Maximum(io, &x.AuxValue, 32767)
	OptionalFunc(io, &x.NetIDVariant, io.Varint32)
	io.Varuint32(&x.BlockRuntimeID)
	io.ByteSlice(&x.UserDataBuffer)
}

type NetworkPermissions struct {
	ServerAuthSoundEnabled bool
}

// Marshal reads or writes NetworkPermissions using its canonical wire layout.
func (x *NetworkPermissions) Marshal(io IO) {
	io.Bool(&x.ServerAuthSoundEnabled)
}

type NewInteractionModel int32

const (
	NewInteractionModelTouch     NewInteractionModel = 0
	NewInteractionModelCrosshair NewInteractionModel = 1
	NewInteractionModelClassic   NewInteractionModel = 2
	NewInteractionModelCount     NewInteractionModel = 3
)

// Marshal reads or writes NewInteractionModel through its int32 wire encoding.
func (x *NewInteractionModel) Marshal(io IO) { io.Varint32((*int32)(x)) }

// CauldronInteractEvent is the event data sent when a cauldron is interacted with.
type POICauldronUsed struct {
	// BlockInteractionType ...
	BlockInteractionType MinecraftEventingPOIBlockInteractionType
	// ItemID ...
	ItemID int16
}

func (*POICauldronUsed) tagEventData() uint32 { return 10 }

// Marshal reads or writes POICauldronUsed using its canonical wire layout.
func (x *POICauldronUsed) Marshal(io IO) {
	x.BlockInteractionType.Marshal(io)
	io.Int16(&x.ItemID)
}

type PackedItemUseLegacyInventoryTransaction struct {
	LegacyRequestID    ItemStackLegacyRequestID
	LegacySetItemSlots Optional[[]LegacySetSlot]
	ItemUseTransaction ItemUseInventoryTransaction
}

// Marshal reads or writes PackedItemUseLegacyInventoryTransaction using its canonical wire layout.
func (x *PackedItemUseLegacyInventoryTransaction) Marshal(io IO) {
	x.LegacyRequestID.Marshal(io)
	OptionalFunc(io, &x.LegacySetItemSlots, func(value *[]LegacySetSlot) {
		Slice(io, value)
	})
	x.ItemUseTransaction.Marshal(io)
}

type PacketCompressionAlgorithm uint16

const (
	PacketCompressionAlgorithmZLib   PacketCompressionAlgorithm = 0
	PacketCompressionAlgorithmSnappy PacketCompressionAlgorithm = 1
	PacketCompressionAlgorithmNone   PacketCompressionAlgorithm = 65535
)

// Marshal reads or writes PacketCompressionAlgorithm through its uint16 wire encoding.
func (x *PacketCompressionAlgorithm) Marshal(io IO) { io.Uint16((*uint16)(x)) }

type PacketType uint32

const (
	PacketTypeEmpty                    PacketType = 0
	PacketTypeInitiallyUnlockedRecipes PacketType = 1
	PacketTypeNewlyUnlockedRecipes     PacketType = 2
	PacketTypeRemoveUnlockedRecipes    PacketType = 3
	PacketTypeRemoveAllUnlockedRecipes PacketType = 4
)

// Marshal reads or writes PacketType through its uint32 wire encoding.
func (x *PacketType) Marshal(io IO) { io.Uint32((*uint32)(x)) }

type PacketViolationSeverity int32

const (
	PacketViolationSeverityUnknown               PacketViolationSeverity = -1
	PacketViolationSeverityWarning               PacketViolationSeverity = 0
	PacketViolationSeverityFinalWarning          PacketViolationSeverity = 1
	PacketViolationSeverityTerminatingConnection PacketViolationSeverity = 2
)

// Marshal reads or writes PacketViolationSeverity through its int32 wire encoding.
func (x *PacketViolationSeverity) Marshal(io IO) { io.Varint32((*int32)(x)) }

type PacketViolationType int32

const (
	PacketViolationTypeUnknown         PacketViolationType = -1
	PacketViolationTypePacketMalformed PacketViolationType = 0
)

// Marshal reads or writes PacketViolationType through its int32 wire encoding.
func (x *PacketViolationType) Marshal(io IO) { io.Varint32((*int32)(x)) }

type PersonaAnimatedTextureType uint32

const (
	PersonaAnimatedTextureTypeNone        PersonaAnimatedTextureType = 0
	PersonaAnimatedTextureTypeFace        PersonaAnimatedTextureType = 1
	PersonaAnimatedTextureTypeBody32x32   PersonaAnimatedTextureType = 2
	PersonaAnimatedTextureTypeBody128x128 PersonaAnimatedTextureType = 3
)

// Marshal reads or writes PersonaAnimatedTextureType through its uint32 wire encoding.
func (x *PersonaAnimatedTextureType) Marshal(io IO) { io.Varuint32((*uint32)(x)) }

type PersonaAnimationExpression uint32

const (
	PersonaAnimationExpressionLinear   PersonaAnimationExpression = 0
	PersonaAnimationExpressionBlinking PersonaAnimationExpression = 1
)

// Marshal reads or writes PersonaAnimationExpression through its uint32 wire encoding.
func (x *PersonaAnimationExpression) Marshal(io IO) { io.Varuint32((*uint32)(x)) }

type PersonaArmSizeType uint8

const (
	PersonaArmSizeTypeSlim PersonaArmSizeType = 0
	PersonaArmSizeTypeWide PersonaArmSizeType = 1
)

// Marshal reads or writes PersonaArmSizeType through its uint8 wire encoding.
func (x *PersonaArmSizeType) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type PhotoType uint8

const (
	PhotoTypePortfolio PhotoType = 0
	PhotoTypePhotoItem PhotoType = 1
	PhotoTypeBook      PhotoType = 2
)

// Marshal reads or writes PhotoType through its uint8 wire encoding.
func (x *PhotoType) Marshal(io IO) { io.Uint8((*uint8)(x)) }

// PiglinBarterEvent is called when a player drops gold ingots to a piglin to initiate a trade for an item.
type PiglinBarter struct {
	// ItemID ...
	ItemID int32
	// WasTargetingBarteringPlayer ...
	WasTargetingBarteringPlayer bool
}

func (*PiglinBarter) tagEventData() uint32 { return 16 }

// Marshal reads or writes PiglinBarter using its canonical wire layout.
func (x *PiglinBarter) Marshal(io IO) {
	io.Varint32(&x.ItemID)
	io.Bool(&x.WasTargetingBarteringPlayer)
}

type PlayStatusType int32

const (
	PlayStatusTypeLoginSuccess                             PlayStatusType = 0
	PlayStatusTypeLoginFailedClientOld                     PlayStatusType = 1
	PlayStatusTypeLoginFailedServerOld                     PlayStatusType = 2
	PlayStatusTypePlayerSpawn                              PlayStatusType = 3
	PlayStatusTypeLoginFailedInvalidTenant                 PlayStatusType = 4
	PlayStatusTypeLoginFailedEditionMismatchEduToVanilla   PlayStatusType = 5
	PlayStatusTypeLoginFailedEditionMismatchVanillaToEdu   PlayStatusType = 6
	PlayStatusTypeLoginFailedServerFullSubClient           PlayStatusType = 7
	PlayStatusTypeLoginFailedEditorMismatchEditorToVanilla PlayStatusType = 8
	PlayStatusTypeLoginFailedEditorMismatchVanillaToEditor PlayStatusType = 9
)

// Marshal reads or writes PlayStatusType through its int32 wire encoding.
func (x *PlayStatusType) Marshal(io IO) { io.BEInt32((*int32)(x)) }

type PortalCreated struct {
	DimensionID int32
}

func (*PortalCreated) tagEventData() uint32 { return 2 }

// Marshal reads or writes PortalCreated using its canonical wire layout.
func (x *PortalCreated) Marshal(io IO) {
	io.Varint32(&x.DimensionID)
}

// PortalUsedEvent is the event data sent when a portal is used.
type PortalUsed struct {
	// FromDimensionID ...
	SourceDimensionID int32
	// ToDimensionID ...
	TargetDimensionID int32
}

func (*PortalUsed) tagEventData() uint32 { return 3 }

// Marshal reads or writes PortalUsed using its canonical wire layout.
func (x *PortalUsed) Marshal(io IO) {
	io.Varint32(&x.SourceDimensionID)
	io.Varint32(&x.TargetDimensionID)
}

type PotionMixDataEntry struct {
	FromPotionID   int32
	FromItemAux    int32
	ReagentItemID  int32
	ReagentItemAux int32
	ToPotionID     int32
	ToItemAux      int32
}

// Marshal reads or writes PotionMixDataEntry using its canonical wire layout.
func (x *PotionMixDataEntry) Marshal(io IO) {
	io.Varint32(&x.FromPotionID)
	io.Varint32(&x.FromItemAux)
	io.Varint32(&x.ReagentItemID)
	io.Varint32(&x.ReagentItemAux)
	io.Varint32(&x.ToPotionID)
	io.Varint32(&x.ToItemAux)
}

type PropertySyncData struct {
	IntEntriesList   []PropertySyncDataPropertySyncIntEntry
	FloatEntriesList []PropertySyncDataPropertySyncFloatEntry
}

// Marshal reads or writes PropertySyncData using its canonical wire layout.
func (x *PropertySyncData) Marshal(io IO) {
	Slice(io, &x.IntEntriesList)
	Slice(io, &x.FloatEntriesList)
}

type PropertySyncDataPropertySyncFloatEntry struct {
	PropertyIndex uint32
	Data          float32
}

// Marshal reads or writes PropertySyncDataPropertySyncFloatEntry using its canonical wire layout.
func (x *PropertySyncDataPropertySyncFloatEntry) Marshal(io IO) {
	io.Varuint32(&x.PropertyIndex)
	io.Float32(&x.Data)
}

type PropertySyncDataPropertySyncIntEntry struct {
	PropertyIndex uint32
	Data          int32
}

// Marshal reads or writes PropertySyncDataPropertySyncIntEntry using its canonical wire layout.
func (x *PropertySyncDataPropertySyncIntEntry) Marshal(io IO) {
	io.Varuint32(&x.PropertyIndex)
	io.Varint32(&x.Data)
}

// PyramidShape represents a pyramid debug shape.
type PyramidData struct {
	// Width is the width along the X axis of the pyramid base.
	Width float32
	// Depth is the optional depth along the Z axis of the pyramid base. It defaults to Width if unset.
	Depth Optional[float32]
	// Height is the height of the pyramid.
	Height float32
}

func (*PyramidData) tagPrimitiveShapeExtraShapeData() uint32 { return 7 }

// Marshal reads or writes PyramidData using its canonical wire layout.
func (x *PyramidData) Marshal(io IO) {
	io.Float32(&x.Width)
	OptionalFunc(io, &x.Depth, io.Float32)
	io.Float32(&x.Height)
}

// RaidUpdateEvent is an event used to update a raids progress client side.
type RaidUpdate struct {
	// CurrentRaidWave ...
	CurrentWave int32
	// TotalRaidWaves ...
	TotalWaves int32
	// WonRaid ...
	Success bool
}

func (*RaidUpdate) tagEventData() uint32 { return 14 }

// Marshal reads or writes RaidUpdate using its canonical wire layout.
func (x *RaidUpdate) Marshal(io IO) {
	io.Varint32(&x.CurrentWave)
	io.Varint32(&x.TotalWaves)
	io.Bool(&x.Success)
}

type RandomDistributionType int32

const (
	RandomDistributionTypeSingleValued    RandomDistributionType = 0
	RandomDistributionTypeUniform         RandomDistributionType = 1
	RandomDistributionTypeGaussian        RandomDistributionType = 2
	RandomDistributionTypeInverseGaussian RandomDistributionType = 3
	RandomDistributionTypeFixedGrid       RandomDistributionType = 4
	RandomDistributionTypeJitteredGrid    RandomDistributionType = 5
	RandomDistributionTypeTriangle        RandomDistributionType = 6
)

// Marshal reads or writes RandomDistributionType through its int32 wire encoding.
func (x *RandomDistributionType) Marshal(io IO) { io.Varint32((*int32)(x)) }

type RemoveEntry struct {
	Action PlayerListPacketType
	UUID   uuid.UUID
}

func (*RemoveEntry) tagPlayerListData() uint32 { return 0 }

// Marshal reads or writes RemoveEntry using its canonical wire layout.
func (x *RemoveEntry) Marshal(io IO) {
	x.Action.Marshal(io)
	io.UUID(&x.UUID)
}

type RemoveEnvironmentAttributes struct {
	AttributeLayerName      string
	AttributeLayerDimension DimensionType
	Attributes              []string
}

func (*RemoveEnvironmentAttributes) tagAttributeLayerSyncData() uint32 { return 3 }

// Marshal reads or writes RemoveEnvironmentAttributes using its canonical wire layout.
func (x *RemoveEnvironmentAttributes) Marshal(io IO) {
	io.StringLimits(&x.AttributeLayerName, 0, 128)
	x.AttributeLayerDimension.Marshal(io)
	FuncSliceLimits(io, &x.Attributes, io.Varuint32, 0, 1024, func(value *string) {
		io.StringLimits(value, 0, 128)
	})
}

type RemoveOverride struct {
	Type string
}

func (*RemoveOverride) tagPlayerUpdateEntityOverridesData() uint32 { return 1 }

// Marshal reads or writes RemoveOverride using its canonical wire layout.
func (x *RemoveOverride) Marshal(io IO) {
	io.String(&x.Type)
}

type RemoveScore struct {
	Action        string
	ScoreboardID  ScoreboardID
	ObjectiveName Optional[string]
}

func (*RemoveScore) tagSetScoreInfoItem() uint32 { return 0 }

// Marshal reads or writes RemoveScore using its canonical wire layout.
func (x *RemoveScore) Marshal(io IO) {
	io.String(&x.Action)
	x.ScoreboardID.Marshal(io)
	OptionalFunc(io, &x.ObjectiveName, io.String)
}

type RemoveTimeMarkerData struct {
	ClockID       uint64
	TimeMarkerIds []uint64
}

func (*RemoveTimeMarkerData) tagSyncWorldClocksData() uint32 { return 3 }

// Marshal reads or writes RemoveTimeMarkerData using its canonical wire layout.
func (x *RemoveTimeMarkerData) Marshal(io IO) {
	io.Varuint64(&x.ClockID)
	FuncSliceLimits(io, &x.TimeMarkerIds, io.Varuint32, 0, 256, io.Varuint64)
}

type RequestAbilityType uint8

const (
	RequestAbilityTypeUnset RequestAbilityType = 0
	RequestAbilityTypeBool  RequestAbilityType = 1
	RequestAbilityTypeFloat RequestAbilityType = 2
)

// Marshal reads or writes RequestAbilityType through its uint8 wire encoding.
func (x *RequestAbilityType) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type RequestType uint8

const (
	RequestTypeSetActions             RequestType = 0
	RequestTypeExecuteAction          RequestType = 1
	RequestTypeExecuteClosingCommands RequestType = 2
	RequestTypeSetName                RequestType = 3
	RequestTypeSetSkin                RequestType = 4
	RequestTypeSetInteractText        RequestType = 5
	RequestTypeExecuteOpeningCommands RequestType = 6
)

// Marshal reads or writes RequestType through its uint8 wire encoding.
func (x *RequestType) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type RewindType uint8

const (
	RewindTypePlayer  RewindType = 0
	RewindTypeVehicle RewindType = 1
)

// Marshal reads or writes RewindType through its uint8 wire encoding.
func (x *RewindType) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type ScriptModuleMinecraftScriptPrimitiveShapeType uint8

const (
	ScriptModuleMinecraftScriptPrimitiveShapeTypeLine      ScriptModuleMinecraftScriptPrimitiveShapeType = 0
	ScriptModuleMinecraftScriptPrimitiveShapeTypeBox       ScriptModuleMinecraftScriptPrimitiveShapeType = 1
	ScriptModuleMinecraftScriptPrimitiveShapeTypeSphere    ScriptModuleMinecraftScriptPrimitiveShapeType = 2
	ScriptModuleMinecraftScriptPrimitiveShapeTypeCircle    ScriptModuleMinecraftScriptPrimitiveShapeType = 3
	ScriptModuleMinecraftScriptPrimitiveShapeTypeText      ScriptModuleMinecraftScriptPrimitiveShapeType = 4
	ScriptModuleMinecraftScriptPrimitiveShapeTypeArrow     ScriptModuleMinecraftScriptPrimitiveShapeType = 5
	ScriptModuleMinecraftScriptPrimitiveShapeTypeCylinder  ScriptModuleMinecraftScriptPrimitiveShapeType = 6
	ScriptModuleMinecraftScriptPrimitiveShapeTypePyramid   ScriptModuleMinecraftScriptPrimitiveShapeType = 7
	ScriptModuleMinecraftScriptPrimitiveShapeTypeEllipsoid ScriptModuleMinecraftScriptPrimitiveShapeType = 8
	ScriptModuleMinecraftScriptPrimitiveShapeTypeCone      ScriptModuleMinecraftScriptPrimitiveShapeType = 9
)

// Marshal reads or writes ScriptModuleMinecraftScriptPrimitiveShapeType through its uint8 wire encoding.
func (x *ScriptModuleMinecraftScriptPrimitiveShapeType) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type SemVersion struct {
	Version string
}

// Marshal reads or writes SemVersion using its canonical wire layout.
func (x *SemVersion) Marshal(io IO) {
	io.String(&x.Version)
}

type SemVersionData struct {
	Version string
}

// Marshal reads or writes SemVersionData using its canonical wire layout.
func (x *SemVersionData) Marshal(io IO) {
	io.String(&x.Version)
}

// AbilityData represents various data about the abilities of a player, such as ability layers or permissions.
type SerializedAbilitiesData struct {
	// EntityUniqueID is a unique identifier of the player. It appears it is not required to fill this field out
	// with a correct value. Simply writing 0 seems to work.
	TargetPlayerRawID int64
	// PlayerPermissions is the permission level of the player as it shows up in the player list built up using
	// the PlayerList packet.
	PlayerPermissions PlayerPermissionLevel
	// CommandPermissions is a set of permissions that specify what commands a player is allowed to execute.
	CommandPermissions CommandPermissionLevel
	// Layers contains all ability layers and their potential values. This should at least have one entry, being
	// the base layer.
	Layers []SerializedAbilitiesDataSerializedLayer
}

// Marshal reads or writes SerializedAbilitiesData using its canonical wire layout.
func (x *SerializedAbilitiesData) Marshal(io IO) {
	io.Int64(&x.TargetPlayerRawID)
	x.PlayerPermissions.Marshal(io)
	x.CommandPermissions.Marshal(io)
	Slice(io, &x.Layers)
}

type SerializedAbilitiesDataSerializedLayer struct {
	SerializedLayer  uint16
	AbilitiesSet     uint32
	AbilityValues    uint32
	FlySpeed         float32
	VerticalFlySpeed float32
	WalkSpeed        float32
}

// Marshal reads or writes SerializedAbilitiesDataSerializedLayer using its canonical wire layout.
func (x *SerializedAbilitiesDataSerializedLayer) Marshal(io IO) {
	io.Uint16(&x.SerializedLayer)
	io.Uint32(&x.AbilitiesSet)
	io.Uint32(&x.AbilityValues)
	io.Float32(&x.FlySpeed)
	io.Float32(&x.VerticalFlySpeed)
	io.Float32(&x.WalkSpeed)
}

// NoiseBlockSpecifier specifies a block placed by the gradient noise based on a threshold and range.
type SerializedNoiseBlockSpecifier struct {
	// Noise is the noise name.
	Noise string
	// Threshold is the noise threshold above which the block is placed.
	Threshold float32
	// Range is the noise range within which the block is placed.
	Range FloatRange
	// Block is the block runtime ID placed by this specifier.
	Block uint32
}

// Marshal reads or writes SerializedNoiseBlockSpecifier using its canonical wire layout.
func (x *SerializedNoiseBlockSpecifier) Marshal(io IO) {
	io.String(&x.Noise)
	io.Float32(&x.Threshold)
	x.Range.Marshal(io)
	io.Uint32(&x.Block)
}

// PersonaPiece represents a piece of a persona skin. All pieces are sent separately.
type SerializedPersonaPieceHandle struct {
	// PieceId is a UUID that identifies the piece itself, which is unique for each separate piece.
	PieceID string
	// PieceType holds the type of the piece. This is one of the PieceType constants above.
	PieceType PersonaPieceType
	// PackID is a UUID that identifies the pack that the persona piece belongs to.
	PackID uuid.UUID
	// Default specifies if the piece is one of the default pieces. This is true when the piece is one of those
	// that a Steve or Alex skin have.
	IsDefaultPiece bool
	// ProductID is a UUID that identifies the piece when it comes to purchases. It is empty for pieces that have
	// the 'Default' field set to true.
	ProductID string
}

// Marshal reads or writes SerializedPersonaPieceHandle using its canonical wire layout.
func (x *SerializedPersonaPieceHandle) Marshal(io IO) {
	io.String(&x.PieceID)
	x.PieceType.Marshal(io)
	io.UUID(&x.PackID)
	io.Bool(&x.IsDefaultPiece)
	io.String(&x.ProductID)
}

type SerializedSkinRef struct {
	ID                           string
	ResourcePatch                string
	ImageData                    SkinImage
	AnimatedImageData            []AnimatedImageData
	CapeImageData                SkinImage
	GeometryData                 string
	GeometryDataMinEngineVersion string
	AnimationData                string
	CapeID                       string
	FullID                       string
	ArmSize                      PersonaArmSizeType
	SkinColor                    color.RGBA
	PersonaPieces                []SerializedPersonaPieceHandle
	PieceTintColors              []OrderedEntry[string, TintMapColor]
	IsPremium                    bool
	IsPersona                    bool
	IsPersonaCapeOnClassicSkin   bool
	IsPrimaryUser                bool
	OverridesPlayerAppearance    bool
	TrustedSkinFlag              string
	ProfileHash                  string
}

// Marshal reads or writes SerializedSkinRef using its canonical wire layout.
func (x *SerializedSkinRef) Marshal(io IO) {
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
	io.RGBA(&x.SkinColor)
	Slice(io, &x.PersonaPieces)
	OrderedMap(io, &x.PieceTintColors, io.Varuint32, io.String, func(value *TintMapColor) {
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

type ServerBlockProperty struct {
	BlockName       string
	BlockDefinition []byte
}

// Marshal reads or writes ServerBlockProperty using its canonical wire layout.
func (x *ServerBlockProperty) Marshal(io IO) {
	io.String(&x.BlockName)
	io.NBT(&x.BlockDefinition, NBTNetwork)
}

// StoreEntryPointInfo contains information about the store entry point.
type ServerConfigurationClientStoreEntryPointConfiguration struct {
	// StoreID is the store identifier.
	StoreID string
	// StoreName is the store name.
	StoreName string
}

// Marshal reads or writes ServerConfigurationClientStoreEntryPointConfiguration using its canonical wire layout.
func (x *ServerConfigurationClientStoreEntryPointConfiguration) Marshal(io IO) {
	io.String(&x.StoreID)
	io.String(&x.StoreName)
}

// GatheringJoinInfo contains information about the gathering (experience) the player is joining.
type ServerConfigurationGatheringsConfigurationJoinInfo struct {
	// ExperienceID is the UUID of the experience.
	ExperienceID uuid.UUID
	// ExperienceName is the name of the experience.
	ExperienceName string
	// ExperienceWorldID is the UUID of the experience world.
	WorldID Optional[uuid.UUID]
	// ExperienceWorldName is the world name of the experience.
	WorldName Optional[string]
	// CreatorID is the ID of the creator.
	CreatorID string
	// TargetID is the session ID of the experience.
	TargetID Optional[uuid.UUID]
	// ScenarioID is the scenario ID of experience.
	ScenarioID Optional[string]
	// ServerID is the server identifier.
	ServerID Optional[string]
}

// Marshal reads or writes ServerConfigurationGatheringsConfigurationJoinInfo using its canonical wire layout.
func (x *ServerConfigurationGatheringsConfigurationJoinInfo) Marshal(io IO) {
	io.UUID(&x.ExperienceID)
	io.StringLimits(&x.ExperienceName, 1, 29)
	OptionalFunc(io, &x.WorldID, io.UUID)
	OptionalFunc(io, &x.WorldName, func(value *string) {
		io.StringLimits(value, 1, 29)
	})
	io.StringLimits(&x.CreatorID, 1, 60)
	OptionalFunc(io, &x.TargetID, io.UUID)
	OptionalFunc(io, &x.ScenarioID, func(value *string) {
		io.StringLimits(value, 1, 100)
	})
	OptionalFunc(io, &x.ServerID, func(value *string) {
		io.StringLimits(value, 1, 100)
	})
}

type ServerConfigurationPresenceConfiguration struct {
	RichPresenceID Optional[string]
}

// Marshal reads or writes ServerConfigurationPresenceConfiguration using its canonical wire layout.
func (x *ServerConfigurationPresenceConfiguration) Marshal(io IO) {
	OptionalFunc(io, &x.RichPresenceID, func(value *string) {
		io.StringLimits(value, 0, 50)
	})
}

type ServerConfigurationServerConfigurationJoinInfo struct {
	Gathering             Optional[ServerConfigurationGatheringsConfigurationJoinInfo]
	ClientStoreEntryPoint Optional[ServerConfigurationClientStoreEntryPointConfiguration]
	Presence              Optional[ServerConfigurationPresenceConfiguration]
}

// Marshal reads or writes ServerConfigurationServerConfigurationJoinInfo using its canonical wire layout.
func (x *ServerConfigurationServerConfigurationJoinInfo) Marshal(io IO) {
	OptionalMarshaler(io, &x.Gathering)
	OptionalMarshaler(io, &x.ClientStoreEntryPoint)
	OptionalMarshaler(io, &x.Presence)
}

type ServerEditorConnectionPolicy int32

const (
	ServerEditorConnectionPolicyMatchWorldType ServerEditorConnectionPolicy = 0
	ServerEditorConnectionPolicyEditorOnly     ServerEditorConnectionPolicy = 1
	ServerEditorConnectionPolicyVanillaOnly    ServerEditorConnectionPolicy = 2
	ServerEditorConnectionPolicyMixed          ServerEditorConnectionPolicy = 3
)

// Marshal reads or writes ServerEditorConnectionPolicy through its int32 wire encoding.
func (x *ServerEditorConnectionPolicy) Marshal(io IO) { io.Varint32((*int32)(x)) }

type ServerSoundHandle struct {
	ServerSoundHandle uint64
}

// Marshal reads or writes ServerSoundHandle using its canonical wire layout.
func (x *ServerSoundHandle) Marshal(io IO) {
	io.Uint64(&x.ServerSoundHandle)
}

// Waypoint holds optional data for a locator bar waypoint.
type ServerWaypoint struct {
	// UpdateFlag is a bitmask indicating which optional fields are set.
	UpdateFlag uint32
	// Visible determines whether the waypoint is shown.
	IsVisible Optional[bool]
	// WorldPosition is the position and dimension of the waypoint.
	WorldPosition Optional[WorldPosition]
	// TexturePath is the resource path for the waypoint icon texture.
	TexturePath Optional[string]
	// IconSize is the size of the waypoint icon.
	IconSize Optional[mgl32.Vec2]
	// Colour is the RGB colour used to tint the waypoint icon.
	Color Optional[color.RGBA]
	// ClientPositionAuthority determines whether the client has authority over the waypoint position.
	ClientPositionAuthority Optional[bool]
	// ActorUniqueID is the unique ID of the entity the waypoint tracks.
	ActorUniqueID Optional[int64]
}

// Marshal reads or writes ServerWaypoint using its canonical wire layout.
func (x *ServerWaypoint) Marshal(io IO) {
	io.Uint32(&x.UpdateFlag)
	OptionalFunc(io, &x.IsVisible, io.Bool)
	OptionalMarshaler(io, &x.WorldPosition)
	OptionalFunc(io, &x.TexturePath, io.String)
	OptionalFunc(io, &x.IconSize, io.Vec2)
	OptionalFunc(io, &x.Color, io.RGBA)
	OptionalFunc(io, &x.ClientPositionAuthority, io.Bool)
	OptionalFunc(io, &x.ActorUniqueID, io.ActorUniqueID)
}

type ServerWaypointGroupAction uint8

const (
	ServerWaypointGroupActionNone   ServerWaypointGroupAction = 0
	ServerWaypointGroupActionAdd    ServerWaypointGroupAction = 1
	ServerWaypointGroupActionRemove ServerWaypointGroupAction = 2
	ServerWaypointGroupActionUpdate ServerWaypointGroupAction = 3
)

// Marshal reads or writes ServerWaypointGroupAction through its uint8 wire encoding.
func (x *ServerWaypointGroupAction) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type ServerboundLoadingScreenType int32

const (
	ServerboundLoadingScreenTypeStartLoadingScreen ServerboundLoadingScreenType = 1
	ServerboundLoadingScreenTypeEndLoadingScreen   ServerboundLoadingScreenType = 2
)

// Marshal reads or writes ServerboundLoadingScreenType through its int32 wire encoding.
func (x *ServerboundLoadingScreenType) Marshal(io IO) { io.Varint32((*int32)(x)) }

type ShowStoreOfferRedirectType uint8

const (
	ShowStoreOfferRedirectTypeMarketplaceOffer     ShowStoreOfferRedirectType = 0
	ShowStoreOfferRedirectTypeDressingRoomOffer    ShowStoreOfferRedirectType = 1
	ShowStoreOfferRedirectTypeThirdPartyServerPage ShowStoreOfferRedirectType = 2
)

// Marshal reads or writes ShowStoreOfferRedirectType through its uint8 wire encoding.
func (x *ShowStoreOfferRedirectType) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type SimulationTypeEnum uint8

const (
	SimulationTypeEnumGame    SimulationTypeEnum = 0
	SimulationTypeEnumEditor  SimulationTypeEnum = 1
	SimulationTypeEnumTest    SimulationTypeEnum = 2
	SimulationTypeEnumInvalid SimulationTypeEnum = 3
)

// Marshal reads or writes SimulationTypeEnum through its uint8 wire encoding.
func (x *SimulationTypeEnum) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type SkinAnimatedImage struct {
	SkinImage           SerializedSkinImage
	AnimatedTextureType SkinAnimatedTextureType
	Frames              float32
	AnimationExpression SkinAnimationExpression
}

// Marshal reads or writes SkinAnimatedImage using its canonical wire layout.
func (x *SkinAnimatedImage) Marshal(io IO) {
	x.SkinImage.Marshal(io)
	x.AnimatedTextureType.Marshal(io)
	io.Float32(&x.Frames)
	x.AnimationExpression.Marshal(io)
}

type SkinAnimationMode uint8

const (
	SkinAnimationModeNone   SkinAnimationMode = 0
	SkinAnimationModeLayers SkinAnimationMode = 1
	SkinAnimationModeBlocks SkinAnimationMode = 2
)

// Marshal reads or writes SkinAnimationMode through its uint8 wire encoding.
func (x *SkinAnimationMode) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type SkinPersonaPiece struct {
	PieceID        string
	PieceType      SkinPieceType
	PackID         uuid.UUID
	IsDefaultPiece bool
	ProductID      string
}

// Marshal reads or writes SkinPersonaPiece using its canonical wire layout.
func (x *SkinPersonaPiece) Marshal(io IO) {
	io.String(&x.PieceID)
	x.PieceType.Marshal(io)
	io.UUID(&x.PackID)
	io.Bool(&x.IsDefaultPiece)
	io.String(&x.ProductID)
}

type SkinTintColor struct {
	Colors [4]Color
}

// Marshal reads or writes SkinTintColor using its canonical wire layout.
func (x *SkinTintColor) Marshal(io IO) {
	for index1 := range x.Colors {
		x.Colors[index1].Marshal(io)
	}
}

type SlashCommand struct {
	SuccessCount int32
	ErrorCount   int32
	CommandName  string
	ErrorList    string
}

func (*SlashCommand) tagEventData() uint32 { return 8 }

// Marshal reads or writes SlashCommand using its canonical wire layout.
func (x *SlashCommand) Marshal(io IO) {
	io.Varint32(&x.SuccessCount)
	io.Varint32(&x.ErrorCount)
	io.StringLimits(&x.CommandName, 0, 512)
	io.StringLimits(&x.ErrorList, 0, 2048)
}

type SocialEventsServerTelemetryData struct {
	ServerID   string
	ScenarioID string
	WorldID    string
	OwnerID    string
}

// Marshal reads or writes SocialEventsServerTelemetryData using its canonical wire layout.
func (x *SocialEventsServerTelemetryData) Marshal(io IO) {
	io.String(&x.ServerID)
	io.String(&x.ScenarioID)
	io.String(&x.WorldID)
	io.String(&x.OwnerID)
}

type SoftEnumUpdateType uint8

const (
	SoftEnumUpdateTypeAdd     SoftEnumUpdateType = 0
	SoftEnumUpdateTypeRemove  SoftEnumUpdateType = 1
	SoftEnumUpdateTypeReplace SoftEnumUpdateType = 2
)

// Marshal reads or writes SoftEnumUpdateType through its uint8 wire encoding.
func (x *SoftEnumUpdateType) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type SpawnBiomeType int16

const (
	SpawnBiomeTypeDefault     SpawnBiomeType = 0
	SpawnBiomeTypeUserDefined SpawnBiomeType = 1
)

// Marshal reads or writes SpawnBiomeType through its int16 wire encoding.
func (x *SpawnBiomeType) Marshal(io IO) { io.Int16((*int16)(x)) }

type SpawnPositionType int32

const (
	SpawnPositionTypePlayerRespawn SpawnPositionType = 0
	SpawnPositionTypeWorldSpawn    SpawnPositionType = 1
)

// Marshal reads or writes SpawnPositionType through its int32 wire encoding.
func (x *SpawnPositionType) Marshal(io IO) { io.Varint32((*int32)(x)) }

type SpawnSettings struct {
	SpawnBiomeType       SpawnBiomeType
	UserDefinedBiomeName string
	Dimension            int32
}

// Marshal reads or writes SpawnSettings using its canonical wire layout.
func (x *SpawnSettings) Marshal(io IO) {
	x.SpawnBiomeType.Marshal(io)
	io.String(&x.UserDefinedBiomeName)
	io.Varint32(&x.Dimension)
}

type SphereData struct {
	NumSegments uint8
}

func (*SphereData) tagPrimitiveShapeExtraShapeData() uint32 { return 5 }

// Marshal reads or writes SphereData using its canonical wire layout.
func (x *SphereData) Marshal(io IO) {
	io.Uint8(&x.NumSegments)
}

type StartVideoCapture struct {
	FrameRate  uint32
	FilePrefix string
}

func (*StartVideoCapture) tagPlayerVideoCaptureData() uint32 { return 0 }

// Marshal reads or writes StartVideoCapture using its canonical wire layout.
func (x *StartVideoCapture) Marshal(io IO) {
	io.Uint32(&x.FrameRate)
	io.String(&x.FilePrefix)
}

type StopVideoCapture struct {
}

func (*StopVideoCapture) tagPlayerVideoCaptureData() uint32 { return 1 }

// Marshal reads or writes StopVideoCapture using its canonical wire layout.
func (x *StopVideoCapture) Marshal(io IO) {
}

type StructureMirror uint8

const (
	StructureMirrorNone StructureMirror = 0
	StructureMirrorX    StructureMirror = 1
	StructureMirrorZ    StructureMirror = 2
	StructureMirrorXZ   StructureMirror = 3
)

// Marshal reads or writes StructureMirror through its uint8 wire encoding.
func (x *StructureMirror) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type StructureRotation uint8

const (
	StructureRotationNone               StructureRotation = 0
	StructureRotationRotate90           StructureRotation = 1
	StructureRotationClockwise90        StructureRotation = 1
	StructureRotationRotate180          StructureRotation = 2
	StructureRotationClockwise180       StructureRotation = 2
	StructureRotationRotate270          StructureRotation = 3
	StructureRotationCounterClockwise90 StructureRotation = 3
)

// Marshal reads or writes StructureRotation through its uint8 wire encoding.
func (x *StructureRotation) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type Subtype uint16

const (
	SubtypeUninitializedSubtype        Subtype = 0
	SubtypeEnableCommands              Subtype = 1
	SubtypeDisableCommands             Subtype = 2
	SubtypeUnlockWorldTemplateSettings Subtype = 3
)

// Marshal reads or writes Subtype through its uint16 wire encoding.
func (x *Subtype) Marshal(io IO) { io.Uint16((*uint16)(x)) }

type SyncStateData struct {
	ClockData []SyncWorldClockStateData
}

func (*SyncStateData) tagSyncWorldClocksData() uint32 { return 0 }

// Marshal reads or writes SyncStateData using its canonical wire layout.
func (x *SyncStateData) Marshal(io IO) {
	SliceLimits(io, &x.ClockData, 0, 256)
}

type SyncedAttribute struct {
	AttributeName string
	MinValue      float32
	CurrentValue  float32
	MaxValue      float32
}

// Marshal reads or writes SyncedAttribute using its canonical wire layout.
func (x *SyncedAttribute) Marshal(io IO) {
	io.String(&x.AttributeName)
	io.Float32(&x.MinValue)
	io.Float32(&x.CurrentValue)
	io.Float32(&x.MaxValue)
}

// PlayerMovementSettings represents the different server authoritative movement settings. These control how
// the client will provide input to the server.
type SyncedPlayerMovementSettings struct {
	// RewindHistorySize is the amount of history to keep at maximum.
	RewindHistorySize int32
	// ServerAuthoritativeBlockBreaking specifies if block breaking should be sent through packet.PlayerAuthInput
	// or not.
	ServerAuthoritativeBlockBreaking bool
}

// Marshal reads or writes SyncedPlayerMovementSettings using its canonical wire layout.
func (x *SyncedPlayerMovementSettings) Marshal(io IO) {
	io.Varint32(&x.RewindHistorySize)
	io.Bool(&x.ServerAuthoritativeBlockBreaking)
}

type SynchedActorDataCopyableDataList struct {
	Data []DataItemEntry
}

// Marshal reads or writes SynchedActorDataCopyableDataList using its canonical wire layout.
func (x *SynchedActorDataCopyableDataList) Marshal(io IO) {
	Slice(io, &x.Data)
}

// TargetBlockHitEvent is an event used when a target block is hit by a arrow.
type TargetBlockHit struct {
	// RedstoneLevel ...
	RedstoneLevel int32
}

func (*TargetBlockHit) tagEventData() uint32 { return 15 }

// Marshal reads or writes TargetBlockHit using its canonical wire layout.
func (x *TargetBlockHit) Marshal(io IO) {
	io.Varint32(&x.RedstoneLevel)
}

type TargetMode uint8

const (
	TargetModeAngle    TargetMode = 0
	TargetModeDistance TargetMode = 1
)

// Marshal reads or writes TargetMode through its uint8 wire encoding.
func (x *TargetMode) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type TintMapColor struct {
	Colors [4]color.RGBA
}

// Marshal reads or writes TintMapColor using its canonical wire layout.
func (x *TintMapColor) Marshal(io IO) {
	for index1 := range x.Colors {
		io.RGBA(&x.Colors[index1])
	}
}

type TitleType int32

const (
	TitleTypeClear               TitleType = 0
	TitleTypeReset               TitleType = 1
	TitleTypeTitle               TitleType = 2
	TitleTypeSubtitle            TitleType = 3
	TitleTypeActionbar           TitleType = 4
	TitleTypeTimes               TitleType = 5
	TitleTypeTitleTextObject     TitleType = 6
	TitleTypeSubtitleTextObject  TitleType = 7
	TitleTypeActionbarTextObject TitleType = 8
)

// Marshal reads or writes TitleType through its int32 wire encoding.
func (x *TitleType) Marshal(io IO) { io.Varint32((*int32)(x)) }

type UpdateSubChunkBlocksChangedInfo struct {
	BlocksChangedStandards []UpdateSubChunkNetworkBlockInfo
	BlocksChangedExtras    []UpdateSubChunkNetworkBlockInfo
}

// Marshal reads or writes UpdateSubChunkBlocksChangedInfo using its canonical wire layout.
func (x *UpdateSubChunkBlocksChangedInfo) Marshal(io IO) {
	Slice(io, &x.BlocksChangedStandards)
	Slice(io, &x.BlocksChangedExtras)
}

type UpdateSubChunkNetworkBlockInfo struct {
	Pos                       BlockPos
	RuntimeID                 uint32
	UpdateFlags               uint32
	SyncMessageEntityUniqueID uint64
	SyncMessageMessage        uint32
}

// Marshal reads or writes UpdateSubChunkNetworkBlockInfo using its canonical wire layout.
func (x *UpdateSubChunkNetworkBlockInfo) Marshal(io IO) {
	x.Pos.Marshal(io)
	io.Varuint32(&x.RuntimeID)
	io.Varuint32(&x.UpdateFlags)
	io.Varuint64(&x.SyncMessageEntityUniqueID)
	io.Varuint32(&x.SyncMessageMessage)
}

type Vec3 struct {
	X float32
	Y float32
	Z float32
}

// Marshal reads or writes Vec3 using its canonical wire layout.
func (x *Vec3) Marshal(io IO) {
	io.Float32(&x.X)
	io.Float32(&x.Y)
	io.Float32(&x.Z)
}

type VillageType uint8

const (
	VillageTypeDesert  VillageType = 0
	VillageTypeIce     VillageType = 1
	VillageTypeSavanna VillageType = 2
	VillageTypeTaiga   VillageType = 3
	VillageTypeDefault VillageType = 4
)

// Marshal reads or writes VillageType through its uint8 wire encoding.
func (x *VillageType) Marshal(io IO) { io.Uint8((*uint8)(x)) }
