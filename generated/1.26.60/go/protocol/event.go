// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

type EventData interface {
	Marshaler
	tagEventData() uint32
}

// MarshalEventData reads or writes the EventData union using its canonical wire layout.
func MarshalEventData(io IO, x *EventData) {
	Union(io, x, io.Varuint32, EventData.tagEventData, func(tag uint32) EventData {
		switch tag {
		case 0:
			return new(Achievement)
		case 1:
			return new(Interaction)
		case 2:
			return new(PortalCreated)
		case 3:
			return new(PortalUsed)
		case 4:
			return new(MobKilled)
		case 5:
			return new(CauldronUsed)
		case 6:
			return new(PlayerDied)
		case 7:
			return new(BossKilled)
		case 8:
			return new(SlashCommand)
		case 9:
			return new(MobBorn)
		case 10:
			return new(POICauldronUsed)
		case 11:
			return new(ComposterUsed)
		case 12:
			return new(BellUsed)
		case 13:
			return new(ActorDefinition)
		case 14:
			return new(RaidUpdate)
		case 15:
			return new(TargetBlockHit)
		case 16:
			return new(PiglinBarter)
		case 17:
			return new(PlayerWaxedOrUnwaxedCopper)
		case 18:
			return new(CodeBuilderRuntimeAction)
		case 19:
			return new(CodeBuilderScoreboard)
		case 20:
			return new(ItemUsed)
		case 21:
			return new(Empty)
		}
		return nil
	})
}
