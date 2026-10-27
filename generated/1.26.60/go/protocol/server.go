// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

type MatchmakingState uint8

const (
	MatchmakingStateIdle             MatchmakingState = 0
	MatchmakingStateMatchmaking      MatchmakingState = 1
	MatchmakingStateMatchFound       MatchmakingState = 2
	MatchmakingStateCanceled         MatchmakingState = 3
	MatchmakingStatePlayerLeftParty  MatchmakingState = 4
	MatchmakingStatePlayerLeftServer MatchmakingState = 5
	MatchmakingStateServerShutdown   MatchmakingState = 6
	MatchmakingStateTimedOut         MatchmakingState = 7
	MatchmakingStateRequeueAsParty   MatchmakingState = 8
)

// Marshal reads or writes MatchmakingState through its uint8 wire encoding.
func (x *MatchmakingState) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type MatchmakingStateOptions struct {
	TriggeringPlayerName   Optional[string]
	TriggeredByLocalPlayer Optional[bool]
}

// Marshal reads or writes MatchmakingStateOptions using its canonical wire layout.
func (x *MatchmakingStateOptions) Marshal(io IO) {
	OptionalFunc(io, &x.TriggeringPlayerName, func(value *string) {
		io.StringLimits(value, 0, 40)
	})
	OptionalFunc(io, &x.TriggeredByLocalPlayer, io.Bool)
}
