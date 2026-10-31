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
	AgentActionTypeDetectredstone    AgentActionType = 4
	AgentActionTypeDetectobstacle    AgentActionType = 5
	AgentActionTypeDrop              AgentActionType = 6
	AgentActionTypeDropall           AgentActionType = 7
	AgentActionTypeInspect           AgentActionType = 8
	AgentActionTypeInspectdata       AgentActionType = 9
	AgentActionTypeInspectitemcount  AgentActionType = 10
	AgentActionTypeInspectitemdetail AgentActionType = 11
	AgentActionTypeInspectitemspace  AgentActionType = 12
	AgentActionTypeInteract          AgentActionType = 13
	AgentActionTypeMove              AgentActionType = 14
	AgentActionTypePlaceblock        AgentActionType = 15
	AgentActionTypeTill              AgentActionType = 16
	AgentActionTypeTransferitemto    AgentActionType = 17
	AgentActionTypeTurn              AgentActionType = 18
)

// Marshal reads or writes AgentActionType through its int32 wire encoding.
func (x *AgentActionType) Marshal(io IO) { io.Int32((*int32)(x)) }

type AgentAnimationType uint8

const (
	AgentAnimationTypeArmswing AgentAnimationType = 0
	AgentAnimationTypeShrug    AgentAnimationType = 1
)

// Marshal reads or writes AgentAnimationType through its uint8 wire encoding.
func (x *AgentAnimationType) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type AnimateAction uint8

const (
	AnimateActionNoaction         AnimateAction = 0
	AnimateActionSwing            AnimateAction = 1
	AnimateActionWakeup           AnimateAction = 3
	AnimateActionCriticalhit      AnimateAction = 4
	AnimateActionMagiccriticalhit AnimateAction = 5
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

type AnimationMode uint8

const (
	AnimationModeNone   AnimationMode = 0
	AnimationModeLayers AnimationMode = 1
	AnimationModeBlocks AnimationMode = 2
)

// Marshal reads or writes AnimationMode through its uint8 wire encoding.
func (x *AnimationMode) Marshal(io IO) { io.Uint8((*uint8)(x)) }

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
	Maximum(io, &x.UpdateCount, 4294967294)
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
	Maximum(io, &x.PropertyUpdateCount, 4294967294)
	io.Uint32(&x.PathUpdateCount)
	Maximum(io, &x.PathUpdateCount, 4294967294)
}

// BellUsedEvent is the event data sent when a bell is used.
type BellUsed struct {
	// ItemID ...
	ItemID int32
}

func (*BellUsed) tagEventData() uint32 { return 12 }

// Marshal reads or writes BellUsed using its canonical wire layout.
func (x *BellUsed) Marshal(io IO) {
	io.Varint32(&x.ItemID)
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
	BossBarOverlayProgress  BossBarOverlay = 0
	BossBarOverlayNotched6  BossBarOverlay = 1
	BossBarOverlayNotched10 BossBarOverlay = 2
	BossBarOverlayNotched12 BossBarOverlay = 3
	BossBarOverlayNotched20 BossBarOverlay = 4
)

// Marshal reads or writes BossBarOverlay through its uint8 wire encoding.
func (x *BossBarOverlay) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type BossEventUpdateType uint8

const (
	BossEventUpdateTypeAdd              BossEventUpdateType = 0
	BossEventUpdateTypePlayeradded      BossEventUpdateType = 1
	BossEventUpdateTypeRemove           BossEventUpdateType = 2
	BossEventUpdateTypePlayerremoved    BossEventUpdateType = 3
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
	io.ActorUniqueID(&x.BossActorID)
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
	BuildPlatformIos          BuildPlatform = 2
	BuildPlatformOSX          BuildPlatform = 3
	BuildPlatformAmazon       BuildPlatform = 4
	BuildPlatformGearvr       BuildPlatform = 5
	BuildPlatformUWP          BuildPlatform = 7
	BuildPlatformWin32        BuildPlatform = 8
	BuildPlatformDedicated    BuildPlatform = 9
	BuildPlatformTvos         BuildPlatform = 10
	BuildPlatformSony         BuildPlatform = 11
	BuildPlatformNintendo     BuildPlatform = 12
	BuildPlatformXbox         BuildPlatform = 13
	BuildPlatformWindowsphone BuildPlatform = 14
	BuildPlatformLinux        BuildPlatform = 15
)

// Marshal reads or writes BuildPlatform through its int32 wire encoding.
func (x *BuildPlatform) Marshal(io IO) { io.Int32((*int32)(x)) }

type Cancel struct {
	ResponseType string
}

func (*Cancel) tagResourcePackClientResponseData() uint32 { return 0 }

// Marshal reads or writes Cancel using its canonical wire layout.
func (x *Cancel) Marshal(io IO) {
	io.String(&x.ResponseType)
}

// CauldronUsedEvent is the event data sent when a cauldron is used.
type CauldronUsed struct {
	// Colour ...
	ContentsColor uint32
	// PotionID ...
	ContentsType int32
	// FillLevel ...
	FillLevel int32
}

func (*CauldronUsed) tagEventData() uint32 { return 5 }

// Marshal reads or writes CauldronUsed using its canonical wire layout.
func (x *CauldronUsed) Marshal(io IO) {
	io.Varuint32(&x.ContentsColor)
	io.Varint32(&x.ContentsType)
	io.Varint32(&x.FillLevel)
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
	ClientCameraAimAssistActionSetfromcamerapreset ClientCameraAimAssistAction = 0
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
	ClientPlayModeLivingroom          ClientPlayMode = 6
	ClientPlayModeExitlevel           ClientPlayMode = 7
	ClientPlayModeExitlevellivingroom ClientPlayMode = 8
	ClientPlayModeNummodes            ClientPlayMode = 9
)

// Marshal reads or writes ClientPlayMode through its uint32 wire encoding.
func (x *ClientPlayMode) Marshal(io IO) { io.Varuint32((*uint32)(x)) }

type ClientboundTextureShiftAction uint8

const (
	ClientboundTextureShiftActionInvalid    ClientboundTextureShiftAction = 0
	ClientboundTextureShiftActionInitialize ClientboundTextureShiftAction = 1
	ClientboundTextureShiftActionStart      ClientboundTextureShiftAction = 2
	ClientboundTextureShiftActionSetenabled ClientboundTextureShiftAction = 3
	ClientboundTextureShiftActionSync       ClientboundTextureShiftAction = 4
)

// Marshal reads or writes ClientboundTextureShiftAction through its uint8 wire encoding.
func (x *ClientboundTextureShiftAction) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type CodeBuilderExecutionStateCodeStatus uint8

const (
	CodeBuilderExecutionStateCodeStatusNone       CodeBuilderExecutionStateCodeStatus = 0
	CodeBuilderExecutionStateCodeStatusNotstarted CodeBuilderExecutionStateCodeStatus = 1
	CodeBuilderExecutionStateCodeStatusInprogress CodeBuilderExecutionStateCodeStatus = 2
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
	CodeBuilderStorageQueryOptionsCategoryCodestatus    CodeBuilderStorageQueryOptionsCategory = 1
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

// ComposterInteractEvent is the event data sent when a composter is interacted with.
type ComposterUsed struct {
	// BlockInteractionType ...
	BlockInteractionType MinecraftEventingPOIBlockInteractionType
	// ItemID ...
	ItemID int32
}

func (*ComposterUsed) tagEventData() uint32 { return 11 }

// Marshal reads or writes ComposterUsed using its canonical wire layout.
func (x *ComposterUsed) Marshal(io IO) {
	x.BlockInteractionType.Marshal(io)
	io.Varint32(&x.ItemID)
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
	ConnectionDisconnectFailReasonCantconnectnointernet                         ConnectionDisconnectFailReason = 1
	ConnectionDisconnectFailReasonNopermissions                                 ConnectionDisconnectFailReason = 2
	ConnectionDisconnectFailReasonUnrecoverableerror                            ConnectionDisconnectFailReason = 3
	ConnectionDisconnectFailReasonThirdpartyblocked                             ConnectionDisconnectFailReason = 4
	ConnectionDisconnectFailReasonThirdpartynointernet                          ConnectionDisconnectFailReason = 5
	ConnectionDisconnectFailReasonThirdpartybadip                               ConnectionDisconnectFailReason = 6
	ConnectionDisconnectFailReasonThirdpartynoserverorserverlocked              ConnectionDisconnectFailReason = 7
	ConnectionDisconnectFailReasonVersionmismatch                               ConnectionDisconnectFailReason = 8
	ConnectionDisconnectFailReasonSkinissue                                     ConnectionDisconnectFailReason = 9
	ConnectionDisconnectFailReasonInvitesessionnotfound                         ConnectionDisconnectFailReason = 10
	ConnectionDisconnectFailReasonEdulevelsettingsmissing                       ConnectionDisconnectFailReason = 11
	ConnectionDisconnectFailReasonLocalservernotfound                           ConnectionDisconnectFailReason = 12
	ConnectionDisconnectFailReasonLegacydisconnect                              ConnectionDisconnectFailReason = 13
	ConnectionDisconnectFailReasonInternalUserleavegameattempted                ConnectionDisconnectFailReason = 14
	ConnectionDisconnectFailReasonPlatformlockedskinserror                      ConnectionDisconnectFailReason = 15
	ConnectionDisconnectFailReasonRealmsworldunassigned                         ConnectionDisconnectFailReason = 16
	ConnectionDisconnectFailReasonRealmsservercantconnect                       ConnectionDisconnectFailReason = 17
	ConnectionDisconnectFailReasonRealmsserverhidden                            ConnectionDisconnectFailReason = 18
	ConnectionDisconnectFailReasonRealmsserverdisabledbeta                      ConnectionDisconnectFailReason = 19
	ConnectionDisconnectFailReasonRealmsserverdisabled                          ConnectionDisconnectFailReason = 20
	ConnectionDisconnectFailReasonCrossplatformdisabled                         ConnectionDisconnectFailReason = 21
	ConnectionDisconnectFailReasonTestonlyCantconnect                           ConnectionDisconnectFailReason = 22
	ConnectionDisconnectFailReasonSessionnotfound                               ConnectionDisconnectFailReason = 23
	ConnectionDisconnectFailReasonClientsettingsincompatiblewithserver          ConnectionDisconnectFailReason = 24
	ConnectionDisconnectFailReasonServerfull                                    ConnectionDisconnectFailReason = 25
	ConnectionDisconnectFailReasonInvalidplatformskin                           ConnectionDisconnectFailReason = 26
	ConnectionDisconnectFailReasonEditionversionmismatch                        ConnectionDisconnectFailReason = 27
	ConnectionDisconnectFailReasonEditionmismatch                               ConnectionDisconnectFailReason = 28
	ConnectionDisconnectFailReasonLevelnewerthanexeversion                      ConnectionDisconnectFailReason = 29
	ConnectionDisconnectFailReasonInternalNofailoccurred                        ConnectionDisconnectFailReason = 30
	ConnectionDisconnectFailReasonBannedskin                                    ConnectionDisconnectFailReason = 31
	ConnectionDisconnectFailReasonTimeout                                       ConnectionDisconnectFailReason = 32
	ConnectionDisconnectFailReasonServernotfound                                ConnectionDisconnectFailReason = 33
	ConnectionDisconnectFailReasonOutdatedserver                                ConnectionDisconnectFailReason = 34
	ConnectionDisconnectFailReasonOutdatedclient                                ConnectionDisconnectFailReason = 35
	ConnectionDisconnectFailReasonNopremiumplatform                             ConnectionDisconnectFailReason = 36
	ConnectionDisconnectFailReasonMultiplayerdisabled                           ConnectionDisconnectFailReason = 37
	ConnectionDisconnectFailReasonNowifi                                        ConnectionDisconnectFailReason = 38
	ConnectionDisconnectFailReasonWorldcorruption                               ConnectionDisconnectFailReason = 39
	ConnectionDisconnectFailReasonNoreason                                      ConnectionDisconnectFailReason = 40
	ConnectionDisconnectFailReasonDisconnected                                  ConnectionDisconnectFailReason = 41
	ConnectionDisconnectFailReasonInvalidplayer                                 ConnectionDisconnectFailReason = 42
	ConnectionDisconnectFailReasonLoggedinotherlocation                         ConnectionDisconnectFailReason = 43
	ConnectionDisconnectFailReasonServeridconflict                              ConnectionDisconnectFailReason = 44
	ConnectionDisconnectFailReasonNotallowed                                    ConnectionDisconnectFailReason = 45
	ConnectionDisconnectFailReasonNotauthenticated                              ConnectionDisconnectFailReason = 46
	ConnectionDisconnectFailReasonInvalidtenant                                 ConnectionDisconnectFailReason = 47
	ConnectionDisconnectFailReasonUnknownpacket                                 ConnectionDisconnectFailReason = 48
	ConnectionDisconnectFailReasonUnexpectedpacket                              ConnectionDisconnectFailReason = 49
	ConnectionDisconnectFailReasonInvalidcommandrequestpacket                   ConnectionDisconnectFailReason = 50
	ConnectionDisconnectFailReasonHostsuspended                                 ConnectionDisconnectFailReason = 51
	ConnectionDisconnectFailReasonLoginpacketnorequest                          ConnectionDisconnectFailReason = 52
	ConnectionDisconnectFailReasonLoginpacketnocert                             ConnectionDisconnectFailReason = 53
	ConnectionDisconnectFailReasonMissingclient                                 ConnectionDisconnectFailReason = 54
	ConnectionDisconnectFailReasonKicked                                        ConnectionDisconnectFailReason = 55
	ConnectionDisconnectFailReasonKickedforexploit                              ConnectionDisconnectFailReason = 56
	ConnectionDisconnectFailReasonKickedforidle                                 ConnectionDisconnectFailReason = 57
	ConnectionDisconnectFailReasonResourcepackproblem                           ConnectionDisconnectFailReason = 58
	ConnectionDisconnectFailReasonIncompatiblepack                              ConnectionDisconnectFailReason = 59
	ConnectionDisconnectFailReasonOutofstorage                                  ConnectionDisconnectFailReason = 60
	ConnectionDisconnectFailReasonInvalidlevel                                  ConnectionDisconnectFailReason = 61
	ConnectionDisconnectFailReasonDisconnectpacket                              ConnectionDisconnectFailReason = 62
	ConnectionDisconnectFailReasonBlockmismatch                                 ConnectionDisconnectFailReason = 63
	ConnectionDisconnectFailReasonInvalidheights                                ConnectionDisconnectFailReason = 64
	ConnectionDisconnectFailReasonInvalidwidths                                 ConnectionDisconnectFailReason = 65
	ConnectionDisconnectFailReasonConnectionlost                                ConnectionDisconnectFailReason = 66
	ConnectionDisconnectFailReasonZombieconnection                              ConnectionDisconnectFailReason = 67
	ConnectionDisconnectFailReasonShutdown                                      ConnectionDisconnectFailReason = 68
	ConnectionDisconnectFailReasonReasonnotset                                  ConnectionDisconnectFailReason = 69
	ConnectionDisconnectFailReasonLoadingstatetimeout                           ConnectionDisconnectFailReason = 70
	ConnectionDisconnectFailReasonResourcepackloadingfailed                     ConnectionDisconnectFailReason = 71
	ConnectionDisconnectFailReasonSearchingforsessionloadingscreenfailed        ConnectionDisconnectFailReason = 72
	ConnectionDisconnectFailReasonNethernetprotocolversion                      ConnectionDisconnectFailReason = 73
	ConnectionDisconnectFailReasonSubsystemstatuserror                          ConnectionDisconnectFailReason = 74
	ConnectionDisconnectFailReasonEmptyauthfromdiscovery                        ConnectionDisconnectFailReason = 75
	ConnectionDisconnectFailReasonEmptyurlfromdiscovery                         ConnectionDisconnectFailReason = 76
	ConnectionDisconnectFailReasonExpiredauthfromdiscovery                      ConnectionDisconnectFailReason = 77
	ConnectionDisconnectFailReasonUnknownsignalservicesigninfailure             ConnectionDisconnectFailReason = 78
	ConnectionDisconnectFailReasonXbljoinlobbyfailure                           ConnectionDisconnectFailReason = 79
	ConnectionDisconnectFailReasonUnspecifiedclientinstancedisconnection        ConnectionDisconnectFailReason = 80
	ConnectionDisconnectFailReasonNethernetsessionnotfound                      ConnectionDisconnectFailReason = 81
	ConnectionDisconnectFailReasonNethernetcreatepeerconnection                 ConnectionDisconnectFailReason = 82
	ConnectionDisconnectFailReasonNethernetice                                  ConnectionDisconnectFailReason = 83
	ConnectionDisconnectFailReasonNethernetconnectrequest                       ConnectionDisconnectFailReason = 84
	ConnectionDisconnectFailReasonNethernetconnectresponse                      ConnectionDisconnectFailReason = 85
	ConnectionDisconnectFailReasonNethernetnegotiationtimeout                   ConnectionDisconnectFailReason = 86
	ConnectionDisconnectFailReasonNethernetinactivitytimeout                    ConnectionDisconnectFailReason = 87
	ConnectionDisconnectFailReasonStaleconnectionbeingreplaced                  ConnectionDisconnectFailReason = 88
	ConnectionDisconnectFailReasonRealmssessionnotfound                         ConnectionDisconnectFailReason = 89
	ConnectionDisconnectFailReasonBadpacket                                     ConnectionDisconnectFailReason = 90
	ConnectionDisconnectFailReasonNethernetfailedtocreateoffer                  ConnectionDisconnectFailReason = 91
	ConnectionDisconnectFailReasonNethernetfailedtocreateanswer                 ConnectionDisconnectFailReason = 92
	ConnectionDisconnectFailReasonNethernetfailedtosetlocaldescription          ConnectionDisconnectFailReason = 93
	ConnectionDisconnectFailReasonNethernetfailedtosetremotedescription         ConnectionDisconnectFailReason = 94
	ConnectionDisconnectFailReasonNethernetnegotiationtimeoutwaitingforresponse ConnectionDisconnectFailReason = 95
	ConnectionDisconnectFailReasonNethernetnegotiationtimeoutwaitingforaccept   ConnectionDisconnectFailReason = 96
	ConnectionDisconnectFailReasonNethernetincomingconnectionignored            ConnectionDisconnectFailReason = 97
	ConnectionDisconnectFailReasonNethernetsignalingparsingfailure              ConnectionDisconnectFailReason = 98
	ConnectionDisconnectFailReasonNethernetsignalingunknownerror                ConnectionDisconnectFailReason = 99
	ConnectionDisconnectFailReasonNethernetsignalingunicastdeliveryfailed       ConnectionDisconnectFailReason = 100
	ConnectionDisconnectFailReasonNethernetsignalingbroadcastdeliveryfailed     ConnectionDisconnectFailReason = 101
	ConnectionDisconnectFailReasonNethernetsignalinggenericdeliveryfailed       ConnectionDisconnectFailReason = 102
	ConnectionDisconnectFailReasonEditormismatcheditorworld                     ConnectionDisconnectFailReason = 103
	ConnectionDisconnectFailReasonEditormismatchvanillaworld                    ConnectionDisconnectFailReason = 104
	ConnectionDisconnectFailReasonWorldtransfernotprimaryclient                 ConnectionDisconnectFailReason = 105
	ConnectionDisconnectFailReasonInternalRequestservershutdown                 ConnectionDisconnectFailReason = 106
	ConnectionDisconnectFailReasonClientgamesetupcancelled                      ConnectionDisconnectFailReason = 107
	ConnectionDisconnectFailReasonClientgamesetupfailed                         ConnectionDisconnectFailReason = 108
	ConnectionDisconnectFailReasonNovenue                                       ConnectionDisconnectFailReason = 109
	ConnectionDisconnectFailReasonNethernetsignalingsigninfailed                ConnectionDisconnectFailReason = 110
	ConnectionDisconnectFailReasonSessionaccessdenied                           ConnectionDisconnectFailReason = 111
	ConnectionDisconnectFailReasonServicesigninissue                            ConnectionDisconnectFailReason = 112
	ConnectionDisconnectFailReasonNethernetnosignalingchannel                   ConnectionDisconnectFailReason = 113
	ConnectionDisconnectFailReasonNethernetnotloggedin                          ConnectionDisconnectFailReason = 114
	ConnectionDisconnectFailReasonNethernetclientsignalingerror                 ConnectionDisconnectFailReason = 115
	ConnectionDisconnectFailReasonSubclientlogindisabled                        ConnectionDisconnectFailReason = 116
	ConnectionDisconnectFailReasonDeeplinktryingtoopendemoworldwhilesignedin    ConnectionDisconnectFailReason = 117
	ConnectionDisconnectFailReasonAsyncjointaskdenied                           ConnectionDisconnectFailReason = 118
	ConnectionDisconnectFailReasonRealmstimelinerequired                        ConnectionDisconnectFailReason = 119
	ConnectionDisconnectFailReasonGuestwithouthost                              ConnectionDisconnectFailReason = 120
	ConnectionDisconnectFailReasonFailedtojoinexperience                        ConnectionDisconnectFailReason = 121
	ConnectionDisconnectFailReasonNethernetdatachannelclosed                    ConnectionDisconnectFailReason = 122
	ConnectionDisconnectFailReasonDiscoveryenvironmentmismatch                  ConnectionDisconnectFailReason = 123
	ConnectionDisconnectFailReasonHostwithoutkeys                               ConnectionDisconnectFailReason = 124
	ConnectionDisconnectFailReasonHostsignedout                                 ConnectionDisconnectFailReason = 125
	ConnectionDisconnectFailReasonScriptwatchdogexception                       ConnectionDisconnectFailReason = 126
	ConnectionDisconnectFailReasonScriptmemorylimitexceeded                     ConnectionDisconnectFailReason = 127
	ConnectionDisconnectFailReasonStoragelowduringgameplay                      ConnectionDisconnectFailReason = 128
	ConnectionDisconnectFailReasonStoragefullduringgameplay                     ConnectionDisconnectFailReason = 129
	ConnectionDisconnectFailReasonLevelstoragecorruption                        ConnectionDisconnectFailReason = 130
	ConnectionDisconnectFailReasonEditionmismatchvanillatoedu                   ConnectionDisconnectFailReason = 131
	ConnectionDisconnectFailReasonEditionmismatchedutovanilla                   ConnectionDisconnectFailReason = 132
	ConnectionDisconnectFailReasonEditormismatcheditortovanilla                 ConnectionDisconnectFailReason = 133
	ConnectionDisconnectFailReasonEditormismatchvanillatoeditor                 ConnectionDisconnectFailReason = 134
	ConnectionDisconnectFailReasonDenylisted                                    ConnectionDisconnectFailReason = 135
	ConnectionDisconnectFailReasonNoncemissing                                  ConnectionDisconnectFailReason = 136
	ConnectionDisconnectFailReasonNoncenotfound                                 ConnectionDisconnectFailReason = 137
	ConnectionDisconnectFailReasonNonceexpired                                  ConnectionDisconnectFailReason = 138
	ConnectionDisconnectFailReasonNoncenotvalid                                 ConnectionDisconnectFailReason = 139
	ConnectionDisconnectFailReasonHostdisconnected                              ConnectionDisconnectFailReason = 140
	ConnectionDisconnectFailReasonEditorjoinintentpolicyfailure                 ConnectionDisconnectFailReason = 141
	ConnectionDisconnectFailReasonNethernetidentitynotallowed                   ConnectionDisconnectFailReason = 142
	ConnectionDisconnectFailReasonInvalidname                                   ConnectionDisconnectFailReason = 143
	ConnectionDisconnectFailReasonExpiredtoken                                  ConnectionDisconnectFailReason = 144
	ConnectionDisconnectFailReasonHostacceptsnotypeofauth                       ConnectionDisconnectFailReason = 145
	ConnectionDisconnectFailReasonNotauthenticatedfastfail                      ConnectionDisconnectFailReason = 146
	ConnectionDisconnectFailReasonEditornotallowed                              ConnectionDisconnectFailReason = 147
	ConnectionDisconnectFailReasonMissingstructuredata                          ConnectionDisconnectFailReason = 148
	ConnectionDisconnectFailReasonUnsupportedtransport                          ConnectionDisconnectFailReason = 149
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
	DataItemTypeCompoundtag DataItemType = 5
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

type Downloading struct {
	ResponseType     string
	DownloadingPacks []string
}

func (*Downloading) tagResourcePackClientResponseData() uint32 { return 1 }

// Marshal reads or writes Downloading using its canonical wire layout.
func (x *Downloading) Marshal(io IO) {
	io.String(&x.ResponseType)
	FuncSliceLimits(io, &x.DownloadingPacks, io.Varuint32, 0, 65535, io.String)
}

type DownloadingFinished struct {
	ResponseType string
}

func (*DownloadingFinished) tagResourcePackClientResponseData() uint32 { return 2 }

// Marshal reads or writes DownloadingFinished using its canonical wire layout.
func (x *DownloadingFinished) Marshal(io IO) {
	io.String(&x.ResponseType)
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
	NoiseName  Optional[string]
	Dimension  DimensionType
	Settings   EASAttributeLayerSettings
	Attributes []EASEnvironmentAttributeData
}

// Marshal reads or writes EASAttributeLayerData using its canonical wire layout.
func (x *EASAttributeLayerData) Marshal(io IO) {
	io.StringLimits(&x.Name, 0, 128)
	OptionalFunc(io, &x.NoiseName, func(value *string) {
		io.StringLimits(value, 0, 128)
	})
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

type EASBoolAttributeData struct {
	Value     bool
	Operation string
}

func (*EASBoolAttributeData) tagEAS() uint32 { return 0 }

// Marshal reads or writes EASBoolAttributeData using its canonical wire layout.
func (x *EASBoolAttributeData) Marshal(io IO) {
	io.Bool(&x.Value)
	io.String(&x.Operation)
}

type EASColorAttributeData struct {
	Value     [4]int32
	Operation string
}

func (*EASColorAttributeData) tagEAS() uint32 { return 2 }

// Marshal reads or writes EASColorAttributeData using its canonical wire layout.
func (x *EASColorAttributeData) Marshal(io IO) {
	for index1 := range x.Value {
		io.Int32(&x.Value[index1])
	}
	io.String(&x.Operation)
}

type EASEnvironmentAttributeData struct {
	AttributeName          string
	FromAttribute          Optional[EAS]
	Attribute              EAS
	ToAttribute            Optional[EAS]
	CurrentTransitionTicks uint32
	TotalTransitionTicks   uint32
	Easing                 string
	LocalTransitionTicks   uint32
	NoiseTransition        bool
	NoiseAlignment         EASNoiseAlignment
}

// Marshal reads or writes EASEnvironmentAttributeData using its canonical wire layout.
func (x *EASEnvironmentAttributeData) Marshal(io IO) {
	io.StringLimits(&x.AttributeName, 0, 128)
	OptionalFunc(io, &x.FromAttribute, func(value *EAS) {
		MarshalEAS(io, value)
	})
	MarshalEAS(io, &x.Attribute)
	OptionalFunc(io, &x.ToAttribute, func(value *EAS) {
		MarshalEAS(io, value)
	})
	io.Uint32(&x.CurrentTransitionTicks)
	io.Uint32(&x.TotalTransitionTicks)
	io.String(&x.Easing)
	io.Uint32(&x.LocalTransitionTicks)
	io.Bool(&x.NoiseTransition)
	x.NoiseAlignment.Marshal(io)
}

type EASFloatAttributeData struct {
	Value         float32
	Operation     string
	ConstraintMin Optional[float32]
	ConstraintMax Optional[float32]
}

func (*EASFloatAttributeData) tagEAS() uint32 { return 1 }

// Marshal reads or writes EASFloatAttributeData using its canonical wire layout.
func (x *EASFloatAttributeData) Marshal(io IO) {
	io.Float32(&x.Value)
	io.String(&x.Operation)
	OptionalFunc(io, &x.ConstraintMin, io.Float32)
	OptionalFunc(io, &x.ConstraintMax, io.Float32)
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

type EditorWorldType int32

const (
	EditorWorldTypeNoneditor          EditorWorldType = 0
	EditorWorldTypeEditorproject      EditorWorldType = 1
	EditorWorldTypeEditortestlevel    EditorWorldType = 2
	EditorWorldTypeEditorrealmsupload EditorWorldType = 3
)

// Marshal reads or writes EditorWorldType through its int32 wire encoding.
func (x *EditorWorldType) Marshal(io IO) { io.Varint32((*int32)(x)) }

// EducationSharedResourceURI is an education edition feature that is used for transmitting education resource
// settings to clients. It contains a button name and a link URL.
type EduSharedURIResource struct {
	// ButtonName is the button name of the resource URI.
	ButtonName string
	// LinkURI is the link URI for the resource URI.
	LinkURI string
}

// Marshal reads or writes EduSharedURIResource using its canonical wire layout.
func (x *EduSharedURIResource) Marshal(io IO) {
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

type GameRulesChangedData struct {
	RulesList []GameRule
}

// Marshal reads or writes GameRulesChangedData using its canonical wire layout.
func (x *GameRulesChangedData) Marshal(io IO) {
	Slice(io, &x.RulesList)
}

type GameType int32

const (
	GameTypeUndefined GameType = -1
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
	GeneratorTypeTheend    GeneratorType = 4
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
	GraphicsModeRaytraced GraphicsMode = 3
)

// Marshal reads or writes GraphicsMode through its uint8 wire encoding.
func (x *GraphicsMode) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type GraphicsOverrideParameterType uint8

const (
	GraphicsOverrideParameterTypeSkyzenithcolor          GraphicsOverrideParameterType = 0
	GraphicsOverrideParameterTypeSkyhorizoncolor         GraphicsOverrideParameterType = 1
	GraphicsOverrideParameterTypeHorizonblendmin         GraphicsOverrideParameterType = 2
	GraphicsOverrideParameterTypeHorizonblendmax         GraphicsOverrideParameterType = 3
	GraphicsOverrideParameterTypeHorizonblendstart       GraphicsOverrideParameterType = 4
	GraphicsOverrideParameterTypeHorizonblendmiestart    GraphicsOverrideParameterType = 5
	GraphicsOverrideParameterTypeRayleighstrength        GraphicsOverrideParameterType = 6
	GraphicsOverrideParameterTypeSunmiestrength          GraphicsOverrideParameterType = 7
	GraphicsOverrideParameterTypeMoonmiestrength         GraphicsOverrideParameterType = 8
	GraphicsOverrideParameterTypeSunglareshape           GraphicsOverrideParameterType = 9
	GraphicsOverrideParameterTypeChlorophyll             GraphicsOverrideParameterType = 10
	GraphicsOverrideParameterTypeCdom                    GraphicsOverrideParameterType = 11
	GraphicsOverrideParameterTypeSuspendedsediment       GraphicsOverrideParameterType = 12
	GraphicsOverrideParameterTypeWavesdepth              GraphicsOverrideParameterType = 13
	GraphicsOverrideParameterTypeWavesfrequency          GraphicsOverrideParameterType = 14
	GraphicsOverrideParameterTypeWavesfrequencyscaling   GraphicsOverrideParameterType = 15
	GraphicsOverrideParameterTypeWavesspeed              GraphicsOverrideParameterType = 16
	GraphicsOverrideParameterTypeWavesspeedscaling       GraphicsOverrideParameterType = 17
	GraphicsOverrideParameterTypeWavesshape              GraphicsOverrideParameterType = 18
	GraphicsOverrideParameterTypeWavesoctaves            GraphicsOverrideParameterType = 19
	GraphicsOverrideParameterTypeWavesmix                GraphicsOverrideParameterType = 20
	GraphicsOverrideParameterTypeWavespull               GraphicsOverrideParameterType = 21
	GraphicsOverrideParameterTypeWavesdirectionincrement GraphicsOverrideParameterType = 22
	GraphicsOverrideParameterTypeMidtonescontrast        GraphicsOverrideParameterType = 23
	GraphicsOverrideParameterTypeHighlightscontrast      GraphicsOverrideParameterType = 24
	GraphicsOverrideParameterTypeShadowscontrast         GraphicsOverrideParameterType = 25
	GraphicsOverrideParameterTypeHighlightsgain          GraphicsOverrideParameterType = 26
	GraphicsOverrideParameterTypeHighlightsgamma         GraphicsOverrideParameterType = 27
	GraphicsOverrideParameterTypeHighlightsoffset        GraphicsOverrideParameterType = 28
	GraphicsOverrideParameterTypeHighlightssaturation    GraphicsOverrideParameterType = 29
	GraphicsOverrideParameterTypeMidtonesgain            GraphicsOverrideParameterType = 30
	GraphicsOverrideParameterTypeMidtonesgamma           GraphicsOverrideParameterType = 31
	GraphicsOverrideParameterTypeMidtonesoffset          GraphicsOverrideParameterType = 32
	GraphicsOverrideParameterTypeMidtonessaturation      GraphicsOverrideParameterType = 33
	GraphicsOverrideParameterTypeShadowsgain             GraphicsOverrideParameterType = 34
	GraphicsOverrideParameterTypeShadowsgamma            GraphicsOverrideParameterType = 35
	GraphicsOverrideParameterTypeShadowsoffset           GraphicsOverrideParameterType = 36
	GraphicsOverrideParameterTypeShadowssaturation       GraphicsOverrideParameterType = 37
	GraphicsOverrideParameterTypeHighlightsmin           GraphicsOverrideParameterType = 38
	GraphicsOverrideParameterTypeShadowsmax              GraphicsOverrideParameterType = 39
	GraphicsOverrideParameterTypeTemperature             GraphicsOverrideParameterType = 40
	GraphicsOverrideParameterTypeSuncolor                GraphicsOverrideParameterType = 41
	GraphicsOverrideParameterTypeSunilluminance          GraphicsOverrideParameterType = 42
	GraphicsOverrideParameterTypeMooncolor               GraphicsOverrideParameterType = 43
	GraphicsOverrideParameterTypeMoonilluminance         GraphicsOverrideParameterType = 44
	GraphicsOverrideParameterTypeFlashcolor              GraphicsOverrideParameterType = 45
	GraphicsOverrideParameterTypeFlashilluminance        GraphicsOverrideParameterType = 46
	GraphicsOverrideParameterTypeAmbientcolor            GraphicsOverrideParameterType = 47
	GraphicsOverrideParameterTypeAmbientilluminance      GraphicsOverrideParameterType = 48
	GraphicsOverrideParameterTypeEmissivedesaturation    GraphicsOverrideParameterType = 49
	GraphicsOverrideParameterTypeSkyintensity            GraphicsOverrideParameterType = 50
	GraphicsOverrideParameterTypeOrbitaloffsetdegrees    GraphicsOverrideParameterType = 51
)

// Marshal reads or writes GraphicsOverrideParameterType through its uint8 wire encoding.
func (x *GraphicsOverrideParameterType) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type HeightMapDataType uint8

const (
	HeightMapDataTypeNodata     HeightMapDataType = 0
	HeightMapDataTypeHasdata    HeightMapDataType = 1
	HeightMapDataTypeAlltoohigh HeightMapDataType = 2
	HeightMapDataTypeAlltoolow  HeightMapDataType = 3
	HeightMapDataTypeAllcopied  HeightMapDataType = 4
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
			FuncSlice(io, &(*value)[index1], io.Varuint32, io.Int8)
		}
	})
	x.RenderHeightMapType.Marshal(io)
	OptionalFunc(io, &x.SubchunkRenderHeightMap, func(value *[16][]int8) {
		for index2 := range *value {
			FuncSlice(io, &(*value)[index2], io.Varuint32, io.Int8)
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
	HudElementPaperdoll     HudElement = 0
	HudElementArmor         HudElement = 1
	HudElementTooltips      HudElement = 2
	HudElementTouchcontrols HudElement = 3
	HudElementCrosshair     HudElement = 4
	HudElementHotbar        HudElement = 5
	HudElementHealth        HudElement = 6
	HudElementProgressbar   HudElement = 7
	HudElementHunger        HudElement = 8
	HudElementAirbubbles    HudElement = 9
	HudElementHorsehealth   HudElement = 10
	HudElementStatuseffects HudElement = 11
	HudElementItemtext      HudElement = 12
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
	InputDataNorthjump                       InputData = 2
	InputDataJumpdown                        InputData = 3
	InputDataSprintdown                      InputData = 4
	InputDataChangeheight                    InputData = 5
	InputDataJumping                         InputData = 6
	InputDataAutojumpinginwater              InputData = 7
	InputDataSneaking                        InputData = 8
	InputDataSneakdown                       InputData = 9
	InputDataUp                              InputData = 10
	InputDataDown                            InputData = 11
	InputDataLeft                            InputData = 12
	InputDataRight                           InputData = 13
	InputDataUpleft                          InputData = 14
	InputDataUpright                         InputData = 15
	InputDataWantup                          InputData = 16
	InputDataWantdown                        InputData = 17
	InputDataWantdownslow                    InputData = 18
	InputDataWantupslow                      InputData = 19
	InputDataSprinting                       InputData = 20
	InputDataAscendblock                     InputData = 21
	InputDataDescendblock                    InputData = 22
	InputDataSneaktoggledown                 InputData = 23
	InputDataPersistsneak                    InputData = 24
	InputDataStartsprinting                  InputData = 25
	InputDataStopsprinting                   InputData = 26
	InputDataStartsneaking                   InputData = 27
	InputDataStopsneaking                    InputData = 28
	InputDataStartswimming                   InputData = 29
	InputDataStopswimming                    InputData = 30
	InputDataStartjumping                    InputData = 31
	InputDataStartgliding                    InputData = 32
	InputDataStopgliding                     InputData = 33
	InputDataPerformiteminteraction          InputData = 34
	InputDataPerformblockactions             InputData = 35
	InputDataPerformitemstackrequest         InputData = 36
	InputDataHandledteleport                 InputData = 37
	InputDataEmoting                         InputData = 38
	InputDataMissedswing                     InputData = 39
	InputDataStartcrawling                   InputData = 40
	InputDataStopcrawling                    InputData = 41
	InputDataStartflying                     InputData = 42
	InputDataStopflying                      InputData = 43
	InputDataClientackserverdata             InputData = 44
	InputDataIsinclientpredictedvehicle      InputData = 45
	InputDataPaddlingleft                    InputData = 46
	InputDataPaddlingright                   InputData = 47
	InputDataBlockbreakingdelayenabled       InputData = 48
	InputDataHorizontalcollision             InputData = 49
	InputDataVerticalcollision               InputData = 50
	InputDataDownleft                        InputData = 51
	InputDataDownright                       InputData = 52
	InputDataStartusingitem                  InputData = 53
	InputDataIscamerarelativemovementenabled InputData = 54
	InputDataIsrotcontrolledbymovedirection  InputData = 55
	InputDataStartspinattack                 InputData = 56
	InputDataStopspinattack                  InputData = 57
	InputDataIshotbaronlytouch               InputData = 58
	InputDataJumpreleasedraw                 InputData = 59
	InputDataJumppressedraw                  InputData = 60
	InputDataJumpcurrentraw                  InputData = 61
	InputDataSneakreleasedraw                InputData = 62
	InputDataSneakpressedraw                 InputData = 63
	InputDataSneakcurrentraw                 InputData = 64
	InputDataInternalupdate                  InputData = 65
)

// Marshal reads or writes InputData through its int32 wire encoding.
func (x *InputData) Marshal(io IO) { io.Varint32((*int32)(x)) }

type InputMode uint32

const (
	InputModeUndefined        InputMode = 0
	InputModeMouse            InputMode = 1
	InputModeTouch            InputMode = 2
	InputModeGamepad          InputMode = 3
	InputModeMotioncontroller InputMode = 4
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
	InteractActionStopriding     InteractAction = 3
	InteractActionInteractupdate InteractAction = 4
	InteractActionNpcopen        InteractAction = 5
	InteractActionOpeninventory  InteractAction = 6
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
	LabTableReactionTypeIcebomb            LabTableReactionType = 1
	LabTableReactionTypeBleach             LabTableReactionType = 2
	LabTableReactionTypeElephanttoothpaste LabTableReactionType = 3
	LabTableReactionTypeFertilizer         LabTableReactionType = 4
	LabTableReactionTypeHeatblock          LabTableReactionType = 5
	LabTableReactionTypeMagnesiumsalts     LabTableReactionType = 6
	LabTableReactionTypeMiscfire           LabTableReactionType = 7
	LabTableReactionTypeMiscexplosion      LabTableReactionType = 8
	LabTableReactionTypeMisclava           LabTableReactionType = 9
	LabTableReactionTypeMiscmystical       LabTableReactionType = 10
	LabTableReactionTypeMiscsmoke          LabTableReactionType = 11
	LabTableReactionTypeMisclargesmoke     LabTableReactionType = 12
)

// Marshal reads or writes LabTableReactionType through its uint8 wire encoding.
func (x *LabTableReactionType) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type LabTableType uint8

const (
	LabTableTypeStartcombine  LabTableType = 0
	LabTableTypeStartreaction LabTableType = 1
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

type LegacyDifficulty int32

const (
	LegacyDifficultyPeaceful LegacyDifficulty = 0
	LegacyDifficultyEasy     LegacyDifficulty = 1
	LegacyDifficultyNormal   LegacyDifficulty = 2
	LegacyDifficultyHard     LegacyDifficulty = 3
	LegacyDifficultyCount    LegacyDifficulty = 4
	LegacyDifficultyUnknown  LegacyDifficulty = 5
)

// Marshal reads or writes LegacyDifficulty through its int32 wire encoding.
func (x *LegacyDifficulty) Marshal(io IO) { io.Varint32((*int32)(x)) }

type LegacySetSlot struct {
	ContainerEnum ContainerEnumName
	Slots         []uint8
}

// Marshal reads or writes LegacySetSlot using its canonical wire layout.
func (x *LegacySetSlot) Marshal(io IO) {
	x.ContainerEnum.Marshal(io)
	io.ByteSlice(&x.Slots)
}

type LevelSettings struct {
	Seed                                   uint64
	SpawnSettings                          SpawnSettings
	GeneratorType                          GeneratorType
	GameType                               GameType
	IsHardcore                             bool
	GameDifficulty                         LegacyDifficulty
	DefaultSpawnBlockPosition              BlockPos
	AchievementsDisabled                   bool
	EditorWorldType                        EditorWorldType
	IsCreatedInEditor                      bool
	IsExportedFromEditor                   bool
	DayCycleStopTime                       int32
	EducationEditionOffer                  EducationEditionOffer
	EducationFeaturesEnabled               bool
	EducationProductID                     string
	RainLevel                              float32
	LightningLevel                         float32
	HasConfirmedPlatformLockedContent      bool
	MultiplayerGameIntent                  bool
	LANBroadcastIntent                     bool
	XboxLiveBroadcastSetting               SocialGamePublishSetting
	PlatformBroadcastSetting               SocialGamePublishSetting
	CommandsEnabled                        bool
	TexturePacksRequired                   bool
	RuleData                               GameRulesChangedData
	Experiments                            Experiments
	HasBonusChestEnabled                   bool
	StartWithMapEnabled                    bool
	PlayerPermissions                      PlayerPermissionLevel
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
	EduSharedURIResource                   EduSharedURIResource
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
	MinecraftEventingAchievementIdsChestfullofcobblestone          MinecraftEventingAchievementIds = 7
	MinecraftEventingAchievementIdsDiamondforyou                   MinecraftEventingAchievementIds = 10
	MinecraftEventingAchievementIdsIronbelly                       MinecraftEventingAchievementIds = 20
	MinecraftEventingAchievementIdsIronman                         MinecraftEventingAchievementIds = 21
	MinecraftEventingAchievementIdsOnarail                         MinecraftEventingAchievementIds = 29
	MinecraftEventingAchievementIdsOverkill                        MinecraftEventingAchievementIds = 30
	MinecraftEventingAchievementIdsReturntosender                  MinecraftEventingAchievementIds = 37
	MinecraftEventingAchievementIdsSniperduel                      MinecraftEventingAchievementIds = 38
	MinecraftEventingAchievementIdsStayinfrosty                    MinecraftEventingAchievementIds = 39
	MinecraftEventingAchievementIdsTakeinventory                   MinecraftEventingAchievementIds = 40
	MinecraftEventingAchievementIdsMaproom                         MinecraftEventingAchievementIds = 50
	MinecraftEventingAchievementIdsFreightstation                  MinecraftEventingAchievementIds = 52
	MinecraftEventingAchievementIdsSmelteverything                 MinecraftEventingAchievementIds = 53
	MinecraftEventingAchievementIdsTasteofyourownmedicine          MinecraftEventingAchievementIds = 54
	MinecraftEventingAchievementIdsWhenpigsfly                     MinecraftEventingAchievementIds = 56
	MinecraftEventingAchievementIdsInception                       MinecraftEventingAchievementIds = 58
	MinecraftEventingAchievementIdsArtificialselection             MinecraftEventingAchievementIds = 60
	MinecraftEventingAchievementIdsFreediver                       MinecraftEventingAchievementIds = 61
	MinecraftEventingAchievementIdsSpawnthewither                  MinecraftEventingAchievementIds = 62
	MinecraftEventingAchievementIdsBeaconator                      MinecraftEventingAchievementIds = 63
	MinecraftEventingAchievementIdsGreatview                       MinecraftEventingAchievementIds = 64
	MinecraftEventingAchievementIdsSupersonic                      MinecraftEventingAchievementIds = 65
	MinecraftEventingAchievementIdsTheendagain                     MinecraftEventingAchievementIds = 66
	MinecraftEventingAchievementIdsTreasurehunter                  MinecraftEventingAchievementIds = 67
	MinecraftEventingAchievementIdsShootingstar                    MinecraftEventingAchievementIds = 68
	MinecraftEventingAchievementIdsFashionshow                     MinecraftEventingAchievementIds = 69
	MinecraftEventingAchievementIdsSelfpublishedauthor             MinecraftEventingAchievementIds = 71
	MinecraftEventingAchievementIdsAlternativefuel                 MinecraftEventingAchievementIds = 72
	MinecraftEventingAchievementIdsSleepwiththefishes              MinecraftEventingAchievementIds = 73
	MinecraftEventingAchievementIdsCastaway                        MinecraftEventingAchievementIds = 74
	MinecraftEventingAchievementIdsImamarinebiologist              MinecraftEventingAchievementIds = 75
	MinecraftEventingAchievementIdsSailthe7seas                    MinecraftEventingAchievementIds = 76
	MinecraftEventingAchievementIdsMegold                          MinecraftEventingAchievementIds = 77
	MinecraftEventingAchievementIdsAhoy                            MinecraftEventingAchievementIds = 78
	MinecraftEventingAchievementIdsAtlantis                        MinecraftEventingAchievementIds = 79
	MinecraftEventingAchievementIdsOnepickletwopickleseapicklefour MinecraftEventingAchievementIds = 80
	MinecraftEventingAchievementIdsDoabarrelroll                   MinecraftEventingAchievementIds = 81
	MinecraftEventingAchievementIdsMoskstraumen                    MinecraftEventingAchievementIds = 82
	MinecraftEventingAchievementIdsEcholocation                    MinecraftEventingAchievementIds = 83
	MinecraftEventingAchievementIdsWherehaveyoubeen                MinecraftEventingAchievementIds = 84
	MinecraftEventingAchievementIdsTopoftheworld                   MinecraftEventingAchievementIds = 85
	MinecraftEventingAchievementIdsFruitontheloom                  MinecraftEventingAchievementIds = 86
	MinecraftEventingAchievementIdsSoundthealarm                   MinecraftEventingAchievementIds = 87
	MinecraftEventingAchievementIdsBuylowsellhigh                  MinecraftEventingAchievementIds = 88
	MinecraftEventingAchievementIdsDisenchanted                    MinecraftEventingAchievementIds = 89
	MinecraftEventingAchievementIdsTimeforstew                     MinecraftEventingAchievementIds = 90
	MinecraftEventingAchievementIdsBeeourguest                     MinecraftEventingAchievementIds = 91
	MinecraftEventingAchievementIdsTotalbeelocation                MinecraftEventingAchievementIds = 92
	MinecraftEventingAchievementIdsStickysituation                 MinecraftEventingAchievementIds = 93
	MinecraftEventingAchievementIdsCovermeindebris                 MinecraftEventingAchievementIds = 94
	MinecraftEventingAchievementIdsFloatyourgoat                   MinecraftEventingAchievementIds = 95
	MinecraftEventingAchievementIdsFriend                          MinecraftEventingAchievementIds = 96
	MinecraftEventingAchievementIdsWaxonwaxoff                     MinecraftEventingAchievementIds = 97
	MinecraftEventingAchievementIdsStriderriddeninlavainoverworld  MinecraftEventingAchievementIds = 98
	MinecraftEventingAchievementIdsGoathornacquired                MinecraftEventingAchievementIds = 99
	MinecraftEventingAchievementIdsJukeboxusedinmeadows            MinecraftEventingAchievementIds = 100
	MinecraftEventingAchievementIdsTradedatworldheight             MinecraftEventingAchievementIds = 101
	MinecraftEventingAchievementIdsSurvivedfallfromworldheight     MinecraftEventingAchievementIds = 102
	MinecraftEventingAchievementIdsSneakclosetosculksensor         MinecraftEventingAchievementIds = 103
	MinecraftEventingAchievementIdsItspreads                       MinecraftEventingAchievementIds = 104
	MinecraftEventingAchievementIdsBirthdaysong                    MinecraftEventingAchievementIds = 105
	MinecraftEventingAchievementIdsWithourpowerscombined           MinecraftEventingAchievementIds = 106
	MinecraftEventingAchievementIdsPlantingthepast                 MinecraftEventingAchievementIds = 107
	MinecraftEventingAchievementIdsCarefulrestoration              MinecraftEventingAchievementIds = 108
	MinecraftEventingAchievementIdsRevaulting                      MinecraftEventingAchievementIds = 109
	MinecraftEventingAchievementIdsCrafterscraftingcrafters        MinecraftEventingAchievementIds = 110
	MinecraftEventingAchievementIdsWhoneedsrockets                 MinecraftEventingAchievementIds = 111
	MinecraftEventingAchievementIdsOveroverkill                    MinecraftEventingAchievementIds = 112
	MinecraftEventingAchievementIdsHearttransplanter               MinecraftEventingAchievementIds = 113
	MinecraftEventingAchievementIdsStayhydrated                    MinecraftEventingAchievementIds = 114
	MinecraftEventingAchievementIdsMobkabob                        MinecraftEventingAchievementIds = 115
	MinecraftEventingAchievementIdsAdventuringtime                 MinecraftEventingAchievementIds = 116
	MinecraftEventingAchievementIdsUhoh                            MinecraftEventingAchievementIds = 117
	MinecraftEventingAchievementIdsGettingwood                     MinecraftEventingAchievementIds = 118
	MinecraftEventingAchievementIdsBenchmaking                     MinecraftEventingAchievementIds = 119
	MinecraftEventingAchievementIdsTimetomine                      MinecraftEventingAchievementIds = 120
	MinecraftEventingAchievementIdsHottopic                        MinecraftEventingAchievementIds = 121
	MinecraftEventingAchievementIdsAcquirehardware                 MinecraftEventingAchievementIds = 122
	MinecraftEventingAchievementIdsGettinganupgrade                MinecraftEventingAchievementIds = 123
	MinecraftEventingAchievementIdsMonsterhunter                   MinecraftEventingAchievementIds = 124
	MinecraftEventingAchievementIdsDiamonds                        MinecraftEventingAchievementIds = 125
	MinecraftEventingAchievementIdsPlethoraofcats                  MinecraftEventingAchievementIds = 126
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
	MinecraftEventingInteractionTypePetsleep   MinecraftEventingInteractionType = 14
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
	MinecraftEventingPOIBlockInteractionTypeCreatelocator       MinecraftEventingPOIBlockInteractionType = 5
	MinecraftEventingPOIBlockInteractionTypeRename              MinecraftEventingPOIBlockInteractionType = 6
	MinecraftEventingPOIBlockInteractionTypeItemplaced          MinecraftEventingPOIBlockInteractionType = 7
	MinecraftEventingPOIBlockInteractionTypeItemremoved         MinecraftEventingPOIBlockInteractionType = 8
	MinecraftEventingPOIBlockInteractionTypeCooking             MinecraftEventingPOIBlockInteractionType = 9
	MinecraftEventingPOIBlockInteractionTypeDousing             MinecraftEventingPOIBlockInteractionType = 10
	MinecraftEventingPOIBlockInteractionTypeLighting            MinecraftEventingPOIBlockInteractionType = 11
	MinecraftEventingPOIBlockInteractionTypeHaystack            MinecraftEventingPOIBlockInteractionType = 12
	MinecraftEventingPOIBlockInteractionTypeFilled              MinecraftEventingPOIBlockInteractionType = 13
	MinecraftEventingPOIBlockInteractionTypeEmptied             MinecraftEventingPOIBlockInteractionType = 14
	MinecraftEventingPOIBlockInteractionTypeAdddye              MinecraftEventingPOIBlockInteractionType = 15
	MinecraftEventingPOIBlockInteractionTypeDyeitem             MinecraftEventingPOIBlockInteractionType = 16
	MinecraftEventingPOIBlockInteractionTypeClearitem           MinecraftEventingPOIBlockInteractionType = 17
	MinecraftEventingPOIBlockInteractionTypeEnchantarrow        MinecraftEventingPOIBlockInteractionType = 18
	MinecraftEventingPOIBlockInteractionTypeCompostitemplaced   MinecraftEventingPOIBlockInteractionType = 19
	MinecraftEventingPOIBlockInteractionTypeRecoveredbonemeal   MinecraftEventingPOIBlockInteractionType = 20
	MinecraftEventingPOIBlockInteractionTypeBookplaced          MinecraftEventingPOIBlockInteractionType = 21
	MinecraftEventingPOIBlockInteractionTypeBookopened          MinecraftEventingPOIBlockInteractionType = 22
	MinecraftEventingPOIBlockInteractionTypeDisenchant          MinecraftEventingPOIBlockInteractionType = 23
	MinecraftEventingPOIBlockInteractionTypeRepair              MinecraftEventingPOIBlockInteractionType = 24
	MinecraftEventingPOIBlockInteractionTypeDisenchantandrepair MinecraftEventingPOIBlockInteractionType = 25
)

// Marshal reads or writes MinecraftEventingPOIBlockInteractionType through its uint8 wire encoding.
func (x *MinecraftEventingPOIBlockInteractionType) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type Mirror uint8

const (
	MirrorNone Mirror = 0
	MirrorX    Mirror = 1
	MirrorZ    Mirror = 2
	MirrorXZ   Mirror = 3
)

// Marshal reads or writes Mirror through its uint8 wire encoding.
func (x *Mirror) Marshal(io IO) { io.Uint8((*uint8)(x)) }

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
	MoLangVersionBeforeversioning                       MoLangVersion = 0
	MoLangVersionInitial                                MoLangVersion = 1
	MoLangVersionFixeditemremainingusedurationquery     MoLangVersion = 2
	MoLangVersionExpressionerrormessages                MoLangVersion = 3
	MoLangVersionUnexpectedoperatorerrors               MoLangVersion = 4
	MoLangVersionConditionaloperatorassociativity       MoLangVersion = 5
	MoLangVersionComparisonandlogicaloperatorprecedence MoLangVersion = 6
	MoLangVersionDividebynegativevalue                  MoLangVersion = 7
	MoLangVersionFixedcapeflapamountquery               MoLangVersion = 8
	MoLangVersionQueryblockpropertyrenamedtostate       MoLangVersion = 9
	MoLangVersionDeprecateoldblockquerynames            MoLangVersion = 10
	MoLangVersionDeprecatedsnifferandcamelqueries       MoLangVersion = 11
	MoLangVersionLeafsupportinginfirstsolidblockbelow   MoLangVersion = 12
	MoLangVersionLatest                                 MoLangVersion = 13
	MoLangVersionNumvalidversions                       MoLangVersion = 14
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
	io.ActorUniqueID(&x.InstigatorActorID)
	io.ActorUniqueID(&x.TargetActorID)
	x.InstigatorSChildActorType.Marshal(io)
	io.Varint32(&x.DamageSource)
	io.Varint32(&x.TradeTier)
	io.StringLimits(&x.TraderName, 0, 128)
}

type ModalFormCancelReason uint8

const (
	ModalFormCancelReasonUserclosed ModalFormCancelReason = 0
	ModalFormCancelReasonUserbusy   ModalFormCancelReason = 1
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
	MultiplayerSettingsTypeEnablemultiplayer  MultiplayerSettingsType = 0
	MultiplayerSettingsTypeDisablemultiplayer MultiplayerSettingsType = 1
	MultiplayerSettingsTypeRefreshjoincode    MultiplayerSettingsType = 2
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
	ItemID int32
}

func (*POICauldronUsed) tagEventData() uint32 { return 10 }

// Marshal reads or writes POICauldronUsed using its canonical wire layout.
func (x *POICauldronUsed) Marshal(io IO) {
	x.BlockInteractionType.Marshal(io)
	io.Varint32(&x.ItemID)
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
	PacketCompressionAlgorithmZlib   PacketCompressionAlgorithm = 0
	PacketCompressionAlgorithmSnappy PacketCompressionAlgorithm = 1
	PacketCompressionAlgorithmNone   PacketCompressionAlgorithm = 65535
)

// Marshal reads or writes PacketCompressionAlgorithm through its uint16 wire encoding.
func (x *PacketCompressionAlgorithm) Marshal(io IO) { io.Uint16((*uint16)(x)) }

type PacketType uint32

const (
	PacketTypeEmpty                    PacketType = 0
	PacketTypeInitiallyunlockedrecipes PacketType = 1
	PacketTypeNewlyunlockedrecipes     PacketType = 2
	PacketTypeRemoveunlockedrecipes    PacketType = 3
	PacketTypeRemoveallunlockedrecipes PacketType = 4
)

// Marshal reads or writes PacketType through its uint32 wire encoding.
func (x *PacketType) Marshal(io IO) { io.Uint32((*uint32)(x)) }

type PacketViolationSeverity int32

const (
	PacketViolationSeverityUnknown               PacketViolationSeverity = -1
	PacketViolationSeverityWarning               PacketViolationSeverity = 0
	PacketViolationSeverityFinalwarning          PacketViolationSeverity = 1
	PacketViolationSeverityTerminatingconnection PacketViolationSeverity = 2
)

// Marshal reads or writes PacketViolationSeverity through its int32 wire encoding.
func (x *PacketViolationSeverity) Marshal(io IO) { io.Varint32((*int32)(x)) }

type PacketViolationType int32

const (
	PacketViolationTypeUnknown         PacketViolationType = -1
	PacketViolationTypePacketmalformed PacketViolationType = 0
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
	PhotoTypePhotoitem PhotoType = 1
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
	PlayStatusTypeLoginsuccess                             PlayStatusType = 0
	PlayStatusTypeLoginfailedClientold                     PlayStatusType = 1
	PlayStatusTypeLoginfailedServerold                     PlayStatusType = 2
	PlayStatusTypePlayerspawn                              PlayStatusType = 3
	PlayStatusTypeLoginfailedInvalidtenant                 PlayStatusType = 4
	PlayStatusTypeLoginfailedEditionmismatchedutovanilla   PlayStatusType = 5
	PlayStatusTypeLoginfailedEditionmismatchvanillatoedu   PlayStatusType = 6
	PlayStatusTypeLoginfailedServerfullsubclient           PlayStatusType = 7
	PlayStatusTypeLoginfailedEditormismatcheditortovanilla PlayStatusType = 8
	PlayStatusTypeLoginfailedEditormismatchvanillatoeditor PlayStatusType = 9
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
	RandomDistributionTypeSinglevalued    RandomDistributionType = 0
	RandomDistributionTypeUniform         RandomDistributionType = 1
	RandomDistributionTypeGaussian        RandomDistributionType = 2
	RandomDistributionTypeInversegaussian RandomDistributionType = 3
	RandomDistributionTypeFixedgrid       RandomDistributionType = 4
	RandomDistributionTypeJitteredgrid    RandomDistributionType = 5
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
	RequestTypeSetactions             RequestType = 0
	RequestTypeExecuteaction          RequestType = 1
	RequestTypeExecuteclosingcommands RequestType = 2
	RequestTypeSetname                RequestType = 3
	RequestTypeSetskin                RequestType = 4
	RequestTypeSetinteracttext        RequestType = 5
	RequestTypeExecuteopeningcommands RequestType = 6
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

type Rotation uint8

const (
	RotationNone      Rotation = 0
	RotationRotate90  Rotation = 1
	RotationRotate180 Rotation = 2
	RotationRotate270 Rotation = 3
)

// Marshal reads or writes Rotation through its uint8 wire encoding.
func (x *Rotation) Marshal(io IO) { io.Uint8((*uint8)(x)) }

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
	io.ActorUniqueIDInt64(&x.TargetPlayerRawID)
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
	PlayFabID                    string
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
	io.String(&x.PlayFabID)
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
	ServerEditorConnectionPolicyMatchworldtype ServerEditorConnectionPolicy = 0
	ServerEditorConnectionPolicyEditoronly     ServerEditorConnectionPolicy = 1
	ServerEditorConnectionPolicyVanillaonly    ServerEditorConnectionPolicy = 2
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
	ServerboundLoadingScreenTypeStartloadingscreen ServerboundLoadingScreenType = 1
	ServerboundLoadingScreenTypeEndloadingscreen   ServerboundLoadingScreenType = 2
)

// Marshal reads or writes ServerboundLoadingScreenType through its int32 wire encoding.
func (x *ServerboundLoadingScreenType) Marshal(io IO) { io.Varint32((*int32)(x)) }

type ShowStoreOfferRedirectType uint8

const (
	ShowStoreOfferRedirectTypeMarketplaceoffer     ShowStoreOfferRedirectType = 0
	ShowStoreOfferRedirectTypeDressingroomoffer    ShowStoreOfferRedirectType = 1
	ShowStoreOfferRedirectTypeThirdpartyserverpage ShowStoreOfferRedirectType = 2
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

type SocialGamePublishSetting int32

const (
	SocialGamePublishSettingNomultiplay      SocialGamePublishSetting = 0
	SocialGamePublishSettingInviteonly       SocialGamePublishSetting = 1
	SocialGamePublishSettingFriendsonly      SocialGamePublishSetting = 2
	SocialGamePublishSettingFriendsoffriends SocialGamePublishSetting = 3
	SocialGamePublishSettingPublic           SocialGamePublishSetting = 4
)

// Marshal reads or writes SocialGamePublishSetting through its int32 wire encoding.
func (x *SocialGamePublishSetting) Marshal(io IO) { io.Varint32((*int32)(x)) }

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
	SpawnBiomeTypeUserdefined SpawnBiomeType = 1
)

// Marshal reads or writes SpawnBiomeType through its int16 wire encoding.
func (x *SpawnBiomeType) Marshal(io IO) { io.Int16((*int16)(x)) }

type SpawnPositionType int32

const (
	SpawnPositionTypePlayerrespawn SpawnPositionType = 0
	SpawnPositionTypeWorldspawn    SpawnPositionType = 1
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

type Subtype uint16

const (
	SubtypeUninitializedsubtype        Subtype = 0
	SubtypeEnablecommands              Subtype = 1
	SubtypeDisablecommands             Subtype = 2
	SubtypeUnlockworldtemplatesettings Subtype = 3
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
	TitleTypeTitletextobject     TitleType = 6
	TitleTypeSubtitletextobject  TitleType = 7
	TitleTypeActionbartextobject TitleType = 8
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
	io.ActorUniqueIDVaruint64(&x.SyncMessageEntityUniqueID)
	io.Varuint32(&x.SyncMessageMessage)
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
