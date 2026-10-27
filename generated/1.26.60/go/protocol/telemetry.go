// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

type LegacyTelemetryType int32

const (
	LegacyTelemetryTypeAchievement                     LegacyTelemetryType = 0
	LegacyTelemetryTypeInteraction                     LegacyTelemetryType = 1
	LegacyTelemetryTypePortalCreated                   LegacyTelemetryType = 2
	LegacyTelemetryTypePortalUsed                      LegacyTelemetryType = 3
	LegacyTelemetryTypeMobKilled                       LegacyTelemetryType = 4
	LegacyTelemetryTypeCauldronUsed                    LegacyTelemetryType = 5
	LegacyTelemetryTypePlayerDied                      LegacyTelemetryType = 6
	LegacyTelemetryTypeBossKilled                      LegacyTelemetryType = 7
	LegacyTelemetryTypeAgentCommandOBSOLETE            LegacyTelemetryType = 8
	LegacyTelemetryTypeAgentCreated                    LegacyTelemetryType = 9
	LegacyTelemetryTypePatternRemovedOBSOLETE          LegacyTelemetryType = 10
	LegacyTelemetryTypeSlashCommand                    LegacyTelemetryType = 11
	LegacyTelemetryTypeFishBucketedOBSOLETE            LegacyTelemetryType = 12
	LegacyTelemetryTypeMobBorn                         LegacyTelemetryType = 13
	LegacyTelemetryTypePetDiedOBSOLETE                 LegacyTelemetryType = 14
	LegacyTelemetryTypePOICauldronUsed                 LegacyTelemetryType = 15
	LegacyTelemetryTypeComposterUsed                   LegacyTelemetryType = 16
	LegacyTelemetryTypeBellUsed                        LegacyTelemetryType = 17
	LegacyTelemetryTypeActorDefinition                 LegacyTelemetryType = 18
	LegacyTelemetryTypeRaidUpdate                      LegacyTelemetryType = 19
	LegacyTelemetryTypePlayerMovementAnomalyOBSOLETE   LegacyTelemetryType = 20
	LegacyTelemetryTypePlayerMovementCorrectedOBSOLETE LegacyTelemetryType = 21
	LegacyTelemetryTypeHoneyHarvested                  LegacyTelemetryType = 22
	LegacyTelemetryTypeTargetBlockHit                  LegacyTelemetryType = 23
	LegacyTelemetryTypePiglinBarter                    LegacyTelemetryType = 24
	LegacyTelemetryTypePlayerWaxedOrUnwaxedCopper      LegacyTelemetryType = 25
	LegacyTelemetryTypeCodeBuilderRuntimeAction        LegacyTelemetryType = 26
	LegacyTelemetryTypeCodeBuilderScoreboard           LegacyTelemetryType = 27
	LegacyTelemetryTypeStriderRiddenInLavaInOverworld  LegacyTelemetryType = 28
	LegacyTelemetryTypeSneakCloseToSculkSensor         LegacyTelemetryType = 29
	LegacyTelemetryTypeCarefulRestoration              LegacyTelemetryType = 30
	LegacyTelemetryTypeItemUsed                        LegacyTelemetryType = 31
)

// Marshal reads or writes LegacyTelemetryType through its int32 wire encoding.
func (x *LegacyTelemetryType) Marshal(io IO) { io.Varint32((*int32)(x)) }
