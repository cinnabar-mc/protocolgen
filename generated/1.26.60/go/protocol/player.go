// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

type PlayerActionType int32

const (
	PlayerActionTypeUnknown               PlayerActionType = -1
	PlayerActionTypeStartDestroyBlock     PlayerActionType = 0
	PlayerActionTypeAbortDestroyBlock     PlayerActionType = 1
	PlayerActionTypeStopDestroyBlock      PlayerActionType = 2
	PlayerActionTypeGetUpdatedBlock       PlayerActionType = 3
	PlayerActionTypeDropItem              PlayerActionType = 4
	PlayerActionTypeStartSleeping         PlayerActionType = 5
	PlayerActionTypeStopSleeping          PlayerActionType = 6
	PlayerActionTypeRespawn               PlayerActionType = 7
	PlayerActionTypeStartJump             PlayerActionType = 8
	PlayerActionTypeStartSprinting        PlayerActionType = 9
	PlayerActionTypeStopSprinting         PlayerActionType = 10
	PlayerActionTypeStartSneaking         PlayerActionType = 11
	PlayerActionTypeStopSneaking          PlayerActionType = 12
	PlayerActionTypeCreativeDestroyBlock  PlayerActionType = 13
	PlayerActionTypeChangeDimensionAck    PlayerActionType = 14
	PlayerActionTypeStartGliding          PlayerActionType = 15
	PlayerActionTypeStopGliding           PlayerActionType = 16
	PlayerActionTypeDenyDestroyBlock      PlayerActionType = 17
	PlayerActionTypeCrackBlock            PlayerActionType = 18
	PlayerActionTypeChangeSkin            PlayerActionType = 19
	PlayerActionTypeUpdatedEnchantingSeed PlayerActionType = 20
	PlayerActionTypeStartSwimming         PlayerActionType = 21
	PlayerActionTypeStopSwimming          PlayerActionType = 22
	PlayerActionTypeStartSpinAttack       PlayerActionType = 23
	PlayerActionTypeStopSpinAttack        PlayerActionType = 24
	PlayerActionTypeInteractWithBlock     PlayerActionType = 25
	PlayerActionTypePredictDestroyBlock   PlayerActionType = 26
	PlayerActionTypeContinueDestroyBlock  PlayerActionType = 27
	PlayerActionTypeStartItemUseOn        PlayerActionType = 28
	PlayerActionTypeStopItemUseOn         PlayerActionType = 29
	PlayerActionTypeHandledTeleport       PlayerActionType = 30
	PlayerActionTypeMissedSwing           PlayerActionType = 31
	PlayerActionTypeStartCrawling         PlayerActionType = 32
	PlayerActionTypeStopCrawling          PlayerActionType = 33
	PlayerActionTypeStartFlying           PlayerActionType = 34
	PlayerActionTypeStopFlying            PlayerActionType = 35
	PlayerActionTypeClientAckServerData   PlayerActionType = 36
	PlayerActionTypeStartUsingItem        PlayerActionType = 37
	PlayerActionTypeInternalUpdate        PlayerActionType = 38
	PlayerActionTypeCount                 PlayerActionType = 39
)

// Marshal reads or writes PlayerActionType through its int32 wire encoding.
func (x *PlayerActionType) Marshal(io IO) { io.Varint32((*int32)(x)) }

// PlayerBlockAction ...
type PlayerBlockActionData struct {
	// Action is the action to be performed, and is one of the constants listed above.
	PlayerActionType PlayerActionType
	// BlockPos is the position of the block that was interacted with.
	Position BlockPos
	// Face is the face of the block that was interacted with.
	Facing int32
}

// Marshal reads or writes PlayerBlockActionData using its canonical wire layout.
func (x *PlayerBlockActionData) Marshal(io IO) {
	x.PlayerActionType.Marshal(io)
	x.Position.Marshal(io)
	io.Varint32(&x.Facing)
}

// PlayerDiedEvent is the event data sent when a player dies.
type PlayerDied struct {
	// AttackerEntityID ...
	InstigatorActorID int32
	// AttackerVariant ...
	InstigatorMobVariant int32
	// EntityDamageCause ...
	DamageSource int32
	// InRaid ...
	DiedInRaid bool
}

func (*PlayerDied) tagEventData() uint32 { return 6 }

// Marshal reads or writes PlayerDied using its canonical wire layout.
func (x *PlayerDied) Marshal(io IO) {
	io.Varint32(&x.InstigatorActorID)
	io.Varint32(&x.InstigatorMobVariant)
	io.Varint32(&x.DamageSource)
	io.Bool(&x.DiedInRaid)
}

type PlayerListData interface {
	Marshaler
	tagPlayerListData() uint32
}

// MarshalPlayerListData reads or writes the PlayerListData union using its canonical wire layout.
func MarshalPlayerListData(io IO, x *PlayerListData) {
	Union(io, x, io.Varuint32, PlayerListData.tagPlayerListData, func(tag uint32) PlayerListData {
		switch tag {
		case 0:
			return new(RemoveEntry)
		case 1:
			return new(AddEntry)
		}
		return nil
	})
}

type PlayerListPacketType uint8

const (
	PlayerListPacketTypeAdd    PlayerListPacketType = 0
	PlayerListPacketTypeRemove PlayerListPacketType = 1
)

// Marshal reads or writes PlayerListPacketType through its uint8 wire encoding.
func (x *PlayerListPacketType) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type PlayerLocationData interface {
	Marshaler
	tagPlayerLocationData() uint32
}

// MarshalPlayerLocationData reads or writes the PlayerLocationData union using its canonical wire layout.
func MarshalPlayerLocationData(io IO, x *PlayerLocationData) {
	Union(io, x, io.Varuint32, PlayerLocationData.tagPlayerLocationData, func(tag uint32) PlayerLocationData {
		switch tag {
		case 0:
			return new(CoordinatesLocation)
		case 1:
			return new(HiddenLocation)
		}
		return nil
	})
}

type PlayerLocationType int32

const (
	PlayerLocationTypePlayerLocationCoordinates PlayerLocationType = 0
	PlayerLocationTypePlayerLocationHide        PlayerLocationType = 1
)

// Marshal reads or writes PlayerLocationType through its int32 wire encoding.
func (x *PlayerLocationType) Marshal(io IO) { io.Varint32((*int32)(x)) }

type PlayerPartyInfo struct {
	PartyID       string
	IsPartyLeader bool
}

// Marshal reads or writes PlayerPartyInfo using its canonical wire layout.
func (x *PlayerPartyInfo) Marshal(io IO) {
	io.StringLimits(&x.PartyID, 0, 49)
	io.Bool(&x.IsPartyLeader)
}

type PlayerPermission int8

const (
	PlayerPermissionVisitor  PlayerPermission = 0
	PlayerPermissionMember   PlayerPermission = 1
	PlayerPermissionOperator PlayerPermission = 2
	PlayerPermissionCustom   PlayerPermission = 3
)

// Marshal reads or writes PlayerPermission through its int8 wire encoding.
func (x *PlayerPermission) Marshal(io IO) { io.Int8((*int8)(x)) }

type PlayerPermissionLevel int8

const (
	PlayerPermissionLevelVisitor  PlayerPermissionLevel = 0
	PlayerPermissionLevelMember   PlayerPermissionLevel = 1
	PlayerPermissionLevelOperator PlayerPermissionLevel = 2
	PlayerPermissionLevelCustom   PlayerPermissionLevel = 3
)

// Marshal reads or writes PlayerPermissionLevel through its int8 wire encoding.
func (x *PlayerPermissionLevel) Marshal(io IO) { io.Int8((*int8)(x)) }

type PlayerPositionModeComponentPositionMode uint8

const (
	PlayerPositionModeComponentPositionModeNormal      PlayerPositionModeComponentPositionMode = 0
	PlayerPositionModeComponentPositionModeRespawn     PlayerPositionModeComponentPositionMode = 1
	PlayerPositionModeComponentPositionModeTeleport    PlayerPositionModeComponentPositionMode = 2
	PlayerPositionModeComponentPositionModeOnlyHeadRot PlayerPositionModeComponentPositionMode = 3
)

// Marshal reads or writes PlayerPositionModeComponentPositionMode through its uint8 wire encoding.
func (x *PlayerPositionModeComponentPositionMode) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type PlayerRespawnState uint8

const (
	PlayerRespawnStateSearchingForSpawn  PlayerRespawnState = 0
	PlayerRespawnStateReadyToSpawn       PlayerRespawnState = 1
	PlayerRespawnStateClientReadyToSpawn PlayerRespawnState = 2
)

// Marshal reads or writes PlayerRespawnState through its uint8 wire encoding.
func (x *PlayerRespawnState) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type PlayerScoreboardID struct {
	PlayerUniqueID int64
}

// Marshal reads or writes PlayerScoreboardID using its canonical wire layout.
func (x *PlayerScoreboardID) Marshal(io IO) {
	io.Varint64(&x.PlayerUniqueID)
}

type PlayerVideoCaptureData interface {
	Marshaler
	tagPlayerVideoCaptureData() uint8
}

// MarshalPlayerVideoCaptureData reads or writes the PlayerVideoCaptureData union using its canonical wire layout.
func MarshalPlayerVideoCaptureData(io IO, x *PlayerVideoCaptureData) {
	Union(io, x, io.Uint8, PlayerVideoCaptureData.tagPlayerVideoCaptureData, func(tag uint8) PlayerVideoCaptureData {
		switch tag {
		case 0:
			return new(StopVideoCapture)
		case 1:
			return new(StartVideoCapture)
		}
		return nil
	})
}

type PlayerWaxedOrUnwaxedCopper struct {
	PlayerWaxedOrUnwaxedCopperBlockID int32
}

func (*PlayerWaxedOrUnwaxedCopper) tagEventData() uint32 { return 17 }

// Marshal reads or writes PlayerWaxedOrUnwaxedCopper using its canonical wire layout.
func (x *PlayerWaxedOrUnwaxedCopper) Marshal(io IO) {
	io.Varint32(&x.PlayerWaxedOrUnwaxedCopperBlockID)
}
