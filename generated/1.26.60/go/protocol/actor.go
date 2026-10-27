// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

import (
	"github.com/go-gl/mathgl/mgl32"
)

type ActorDataBoundingBoxComponent struct {
	ActorDataBoundingBox [3]float32
}

// Marshal reads or writes ActorDataBoundingBoxComponent using its canonical wire layout.
func (x *ActorDataBoundingBoxComponent) Marshal(io IO) {
	for index1 := range x.ActorDataBoundingBox {
		io.Float32(&x.ActorDataBoundingBox[index1])
	}
}

type ActorDataFlagComponent struct {
	ActorFlagBitsetData Bitset131
}

// Marshal reads or writes ActorDataFlagComponent using its canonical wire layout.
func (x *ActorDataFlagComponent) Marshal(io IO) {
	io.Bitset(x.ActorFlagBitsetData[:], 131)
}

type ActorDefinition struct {
	EventName string
}

func (*ActorDefinition) tagEventData() uint32 { return 13 }

// Marshal reads or writes ActorDefinition using its canonical wire layout.
func (x *ActorDefinition) Marshal(io IO) {
	io.StringLimits(&x.EventName, 0, 256)
}

type ActorEventType uint8

const (
	ActorEventTypeNone                             ActorEventType = 0
	ActorEventTypeJump                             ActorEventType = 1
	ActorEventTypeHurt                             ActorEventType = 2
	ActorEventTypeDeath                            ActorEventType = 3
	ActorEventTypeStartAttacking                   ActorEventType = 4
	ActorEventTypeStopAttacking                    ActorEventType = 5
	ActorEventTypeTamingFailed                     ActorEventType = 6
	ActorEventTypeTamingSucceeded                  ActorEventType = 7
	ActorEventTypeShakeWetness                     ActorEventType = 8
	ActorEventTypeEatGrass                         ActorEventType = 10
	ActorEventTypeFishhookBubble                   ActorEventType = 11
	ActorEventTypeFishhookFishPosition             ActorEventType = 12
	ActorEventTypeFishhookHookTime                 ActorEventType = 13
	ActorEventTypeFishhookTease                    ActorEventType = 14
	ActorEventTypeSquidFleeing                     ActorEventType = 15
	ActorEventTypeZombieConverting                 ActorEventType = 16
	ActorEventTypePlayAmbient                      ActorEventType = 17
	ActorEventTypeSpawnAlive                       ActorEventType = 18
	ActorEventTypeStartOfferFlower                 ActorEventType = 19
	ActorEventTypeStopOfferFlower                  ActorEventType = 20
	ActorEventTypeLoveHearts                       ActorEventType = 21
	ActorEventTypeVillagerAngry                    ActorEventType = 22
	ActorEventTypeVillagerHappy                    ActorEventType = 23
	ActorEventTypeWitchHatMagic                    ActorEventType = 24
	ActorEventTypeFireworksExplode                 ActorEventType = 25
	ActorEventTypeInLoveHearts                     ActorEventType = 26
	ActorEventTypeSilverfishMergeAnimation         ActorEventType = 27
	ActorEventTypeGuardianAttackSound              ActorEventType = 28
	ActorEventTypeDrinkPotion                      ActorEventType = 29
	ActorEventTypeThrowPotion                      ActorEventType = 30
	ActorEventTypePrimeTNTCart                     ActorEventType = 31
	ActorEventTypePrimeCreeper                     ActorEventType = 32
	ActorEventTypeAirSupply                        ActorEventType = 33
	ActorEventTypeDeprecatedAddPlayerLevels        ActorEventType = 34
	ActorEventTypeGuardianMiningFatigue            ActorEventType = 35
	ActorEventTypeAgentSwingArm                    ActorEventType = 36
	ActorEventTypeDragonStartDeathAnimation        ActorEventType = 37
	ActorEventTypeGroundDust                       ActorEventType = 38
	ActorEventTypeShake                            ActorEventType = 39
	ActorEventTypeFeed                             ActorEventType = 57
	ActorEventTypeBabyAge                          ActorEventType = 60
	ActorEventTypeInstantDeath                     ActorEventType = 61
	ActorEventTypeNotifyTrade                      ActorEventType = 62
	ActorEventTypeLeashDestroyed                   ActorEventType = 63
	ActorEventTypeCaravanUpdated                   ActorEventType = 64
	ActorEventTypeTalismanActivate                 ActorEventType = 65
	ActorEventTypeDeprecatedUpdateStructureFeature ActorEventType = 66
	ActorEventTypePlayerSpawnedMob                 ActorEventType = 67
	ActorEventTypePuke                             ActorEventType = 68
	ActorEventTypeUpdateStackSize                  ActorEventType = 69
	ActorEventTypeStartSwimming                    ActorEventType = 70
	ActorEventTypeBalloonPop                       ActorEventType = 71
	ActorEventTypeTreasureHunt                     ActorEventType = 72
	ActorEventTypeSummonAgent                      ActorEventType = 73
	ActorEventTypeFinishedChargingItem             ActorEventType = 74
	ActorEventTypeActorGrowUp                      ActorEventType = 76
	ActorEventTypeVibrationDetected                ActorEventType = 77
	ActorEventTypeDrinkMilk                        ActorEventType = 78
	ActorEventTypeShakeWetnessStop                 ActorEventType = 79
	ActorEventTypeKineticDamageDealt               ActorEventType = 80
	ActorEventTypeHurtWithoutReceivingDamage       ActorEventType = 81
)

// Marshal reads or writes ActorEventType through its uint8 wire encoding.
func (x *ActorEventType) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type ActorLinkType uint8

const (
	ActorLinkTypeNone      ActorLinkType = 0
	ActorLinkTypeRiding    ActorLinkType = 1
	ActorLinkTypePassenger ActorLinkType = 2
)

// Marshal reads or writes ActorLinkType through its uint8 wire encoding.
func (x *ActorLinkType) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type ActorType int32

const (
	ActorTypeUndefined                  ActorType = 1
	ActorTypeItemEntity                 ActorType = 64
	ActorTypePrimedTNT                  ActorType = 65
	ActorTypeFallingBlock               ActorType = 66
	ActorTypeMovingBlock                ActorType = 67
	ActorTypeExperience                 ActorType = 69
	ActorTypeEyeOfEnder                 ActorType = 70
	ActorTypeEnderCrystal               ActorType = 71
	ActorTypeFireworksRocket            ActorType = 72
	ActorTypeFishingHook                ActorType = 77
	ActorTypeChalkboard                 ActorType = 78
	ActorTypePainting                   ActorType = 83
	ActorTypeLeashKnot                  ActorType = 88
	ActorTypeBoatRideable               ActorType = 90
	ActorTypeLightningBolt              ActorType = 93
	ActorTypeAreaEffectCloud            ActorType = 95
	ActorTypeBalloon                    ActorType = 107
	ActorTypeShield                     ActorType = 117
	ActorTypeLectern                    ActorType = 119
	ActorTypeOminousItemSpawner         ActorType = 145
	ActorTypeCushion                    ActorType = 154
	ActorTypeChestBoatRideable          ActorType = 218
	ActorTypeMob                        ActorType = 256
	ActorTypeNpc                        ActorType = 307
	ActorTypeAgent                      ActorType = 312
	ActorTypeArmorStand                 ActorType = 317
	ActorTypeTripodCamera               ActorType = 318
	ActorTypePlayer                     ActorType = 319
	ActorTypeBee                        ActorType = 378
	ActorTypePiglin                     ActorType = 379
	ActorTypePiglinBrute                ActorType = 383
	ActorTypeAllay                      ActorType = 390
	ActorTypePathfinderMob              ActorType = 768
	ActorTypeIronGolem                  ActorType = 788
	ActorTypeSnowGolem                  ActorType = 789
	ActorTypeWanderingTrader            ActorType = 886
	ActorTypeCopperGolem                ActorType = 916
	ActorTypeSulfurCube                 ActorType = 921
	ActorTypeMonster                    ActorType = 2816
	ActorTypeCreeper                    ActorType = 2849
	ActorTypeSlime                      ActorType = 2853
	ActorTypeEnderMan                   ActorType = 2854
	ActorTypeGhast                      ActorType = 2857
	ActorTypeLavaSlime                  ActorType = 2858
	ActorTypeBlaze                      ActorType = 2859
	ActorTypeWitch                      ActorType = 2861
	ActorTypeGuardian                   ActorType = 2865
	ActorTypeElderGuardian              ActorType = 2866
	ActorTypeDragon                     ActorType = 2869
	ActorTypeShulker                    ActorType = 2870
	ActorTypeVindicator                 ActorType = 2873
	ActorTypeIllagerBeast               ActorType = 2875
	ActorTypeEvocationIllager           ActorType = 2920
	ActorTypeVex                        ActorType = 2921
	ActorTypePillager                   ActorType = 2930
	ActorTypeElderGuardianGhost         ActorType = 2936
	ActorTypeWarden                     ActorType = 2947
	ActorTypeBreeze                     ActorType = 2956
	ActorTypeCreaking                   ActorType = 2962
	ActorTypeAnimal                     ActorType = 4864
	ActorTypeChicken                    ActorType = 4874
	ActorTypeCow                        ActorType = 4875
	ActorTypePig                        ActorType = 4876
	ActorTypeSheep                      ActorType = 4877
	ActorTypeMushroomCow                ActorType = 4880
	ActorTypeRabbit                     ActorType = 4882
	ActorTypePolarBear                  ActorType = 4892
	ActorTypeLlama                      ActorType = 4893
	ActorTypeTurtle                     ActorType = 4938
	ActorTypePanda                      ActorType = 4977
	ActorTypeFox                        ActorType = 4985
	ActorTypeHoglin                     ActorType = 4988
	ActorTypeStrider                    ActorType = 4989
	ActorTypeGoat                       ActorType = 4992
	ActorTypeAxolotl                    ActorType = 4994
	ActorTypeFrog                       ActorType = 4996
	ActorTypeCamel                      ActorType = 5002
	ActorTypeSniffer                    ActorType = 5003
	ActorTypeArmadillo                  ActorType = 5006
	ActorTypeHappyGhast                 ActorType = 5011
	ActorTypeTraderLlama                ActorType = 5021
	ActorTypeWaterAnimal                ActorType = 8960
	ActorTypeSquid                      ActorType = 8977
	ActorTypeDolphin                    ActorType = 8991
	ActorTypePufferfish                 ActorType = 9068
	ActorTypeSalmon                     ActorType = 9069
	ActorTypeTropicalfish               ActorType = 9071
	ActorTypeFish                       ActorType = 9072
	ActorTypeGlowSquid                  ActorType = 9089
	ActorTypeTadpole                    ActorType = 9093
	ActorTypeNautilus                   ActorType = 9109
	ActorTypeTamableAnimal              ActorType = 21248
	ActorTypeWolf                       ActorType = 21262
	ActorTypeOcelot                     ActorType = 21270
	ActorTypeParrot                     ActorType = 21278
	ActorTypeCat                        ActorType = 21323
	ActorTypeAmbient                    ActorType = 33024
	ActorTypeBat                        ActorType = 33043
	ActorTypeUndeadMonster              ActorType = 68352
	ActorTypePigZombie                  ActorType = 68388
	ActorTypeWitherBoss                 ActorType = 68404
	ActorTypePhantom                    ActorType = 68410
	ActorTypeZoglin                     ActorType = 68478
	ActorTypeCamelHusk                  ActorType = 70552
	ActorTypeZombieNautilus             ActorType = 74646
	ActorTypeZombieMonster              ActorType = 199424
	ActorTypeZombie                     ActorType = 199456
	ActorTypeZombieVillager             ActorType = 199468
	ActorTypeHusk                       ActorType = 199471
	ActorTypeDrowned                    ActorType = 199534
	ActorTypeZombieVillagerV2           ActorType = 199540
	ActorTypeFrostbite                  ActorType = 199579
	ActorTypeArthropod                  ActorType = 264960
	ActorTypeSpider                     ActorType = 264995
	ActorTypeSilverfish                 ActorType = 264999
	ActorTypeCaveSpider                 ActorType = 265000
	ActorTypeEndermite                  ActorType = 265015
	ActorTypeMinecart                   ActorType = 524288
	ActorTypeMinecartRideable           ActorType = 524372
	ActorTypeMinecartHopper             ActorType = 524384
	ActorTypeMinecartTNT                ActorType = 524385
	ActorTypeMinecartChest              ActorType = 524386
	ActorTypeMinecartFurnace            ActorType = 524387
	ActorTypeMinecartCommandBlock       ActorType = 524388
	ActorTypeSkeletonMonster            ActorType = 1116928
	ActorTypeSkeleton                   ActorType = 1116962
	ActorTypeStray                      ActorType = 1116974
	ActorTypeWitherSkeleton             ActorType = 1116976
	ActorTypeBogged                     ActorType = 1117072
	ActorTypeParched                    ActorType = 1117079
	ActorTypeEquineAnimal               ActorType = 2118400
	ActorTypeHorse                      ActorType = 2118423
	ActorTypeDonkey                     ActorType = 2118424
	ActorTypeMule                       ActorType = 2118425
	ActorTypeSkeletonHorse              ActorType = 2183962
	ActorTypeZombieHorse                ActorType = 2183963
	ActorTypeProjectile                 ActorType = 4194304
	ActorTypeExperiencePotion           ActorType = 4194372
	ActorTypeShulkerBullet              ActorType = 4194380
	ActorTypeDragonFireball             ActorType = 4194383
	ActorTypeSnowball                   ActorType = 4194385
	ActorTypeThrownEgg                  ActorType = 4194386
	ActorTypeLargeFireball              ActorType = 4194389
	ActorTypeThrownPotion               ActorType = 4194390
	ActorTypeEnderpearl                 ActorType = 4194391
	ActorTypeWitherSkull                ActorType = 4194393
	ActorTypeWitherSkullDangerous       ActorType = 4194395
	ActorTypeSmallFireball              ActorType = 4194398
	ActorTypeLingeringPotion            ActorType = 4194405
	ActorTypeLlamaSpit                  ActorType = 4194406
	ActorTypeEvocationFang              ActorType = 4194407
	ActorTypeIceBomb                    ActorType = 4194410
	ActorTypeBreezeWindChargeProjectile ActorType = 4194445
	ActorTypeWindChargeProjectile       ActorType = 4194447
	ActorTypeIceBall                    ActorType = 4194460
	ActorTypeAbstractArrow              ActorType = 8388608
	ActorTypeTrident                    ActorType = 12582985
	ActorTypeArrow                      ActorType = 12582992
	ActorTypeVillagerBase               ActorType = 16777984
	ActorTypeVillager                   ActorType = 16777999
	ActorTypeVillagerV2                 ActorType = 16778099
)

// Marshal reads or writes ActorType through its int32 wire encoding.
func (x *ActorType) Marshal(io IO) { io.Varint32((*int32)(x)) }

type ActorUniqueIdentifier struct {
	ActorUniqueID int64
}

// Marshal reads or writes ActorUniqueIdentifier using its canonical wire layout.
func (x *ActorUniqueIdentifier) Marshal(io IO) {
	io.Varint64(&x.ActorUniqueID)
}

type PassengerOfBlockArguments struct {
	BlockPos      BlockPos
	Offset        mgl32.Vec3
	Rotation      float32
	RotationLimit float32
	EmoteType     PassengerOfBlockArgumentsEmoteType
}

// Marshal reads or writes PassengerOfBlockArguments using its canonical wire layout.
func (x *PassengerOfBlockArguments) Marshal(io IO) {
	x.BlockPos.Marshal(io)
	io.Vec3(&x.Offset)
	io.Float32(&x.Rotation)
	Minimum(io, &x.Rotation, 0)
	Maximum(io, &x.Rotation, 360)
	io.Float32(&x.RotationLimit)
	Minimum(io, &x.RotationLimit, 0)
	Maximum(io, &x.RotationLimit, 360)
	x.EmoteType.Marshal(io)
}

type PassengerOfBlockArgumentsEmoteType uint8

const (
	PassengerOfBlockArgumentsEmoteTypeStanding PassengerOfBlockArgumentsEmoteType = 0
	PassengerOfBlockArgumentsEmoteTypeRiding   PassengerOfBlockArgumentsEmoteType = 1
	PassengerOfBlockArgumentsEmoteTypeLaying   PassengerOfBlockArgumentsEmoteType = 2
)

// Marshal reads or writes PassengerOfBlockArgumentsEmoteType through its uint8 wire encoding.
func (x *PassengerOfBlockArgumentsEmoteType) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type PlayerUpdateEntityOverridesData interface {
	Marshaler
	tagPlayerUpdateEntityOverridesData() uint32
}

// MarshalPlayerUpdateEntityOverridesData reads or writes the PlayerUpdateEntityOverridesData union using its canonical wire layout.
func MarshalPlayerUpdateEntityOverridesData(io IO, x *PlayerUpdateEntityOverridesData) {
	Union(io, x, io.Varuint32, PlayerUpdateEntityOverridesData.tagPlayerUpdateEntityOverridesData, func(tag uint32) PlayerUpdateEntityOverridesData {
		switch tag {
		case 0:
			return new(ClearOverride)
		case 1:
			return new(RemoveOverride)
		case 2:
			return new(IntOverride)
		case 3:
			return new(FloatOverride)
		}
		return nil
	})
}
