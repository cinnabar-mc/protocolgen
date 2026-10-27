// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// CommandBlockUpdate is sent by the client to update a command block at a specific position. The command
// block may be either a physical block or an entity.
type CommandBlockUpdate struct {
	Target protocol.CommandBlockUpdateData
	// Command is the command currently entered in the command block. This is the command that is executed when
	// the command block is activated.
	Command string
	// LastOutput is the output of the last command executed by the command block. It may be left empty to show
	// simply no output at all, in combination with setting ShouldTrackOutput to false.
	LastOutput string
	// Name is the name of the command block updated. If not empty, it will show this name hovering above the
	// command block when hovering over the block with the cursor.
	Name string
	// FilteredName is a filtered version of Name with all the profanity removed. The client will use this over
	// Name if this field is not empty and they have the "Filter Profanity" setting enabled.
	FilteredName string
	// ShouldTrackOutput specifies if the command block tracks output. If set to false, the output box won't be
	// shown within the command block.
	TrackOutput bool
	// TickDelay is the delay in ticks between executions of a command block, if it is a repeating command block.
	TickDelay int32
	// ExecuteOnFirstTick specifies if the command block should execute on the first tick, AKA as soon as the
	// command block is enabled.
	ExecuteOnFirstTick bool
}

// ID ...
func (*CommandBlockUpdate) ID() uint32 {
	return IDCommandBlockUpdate
}

func (pk *CommandBlockUpdate) Marshal(io protocol.IO) {
	protocol.MarshalCommandBlockUpdateData(io, &pk.Target)
	io.String(&pk.Command)
	io.String(&pk.LastOutput)
	io.String(&pk.Name)
	io.String(&pk.FilteredName)
	io.Bool(&pk.TrackOutput)
	io.Int32(&pk.TickDelay)
	io.Bool(&pk.ExecuteOnFirstTick)
}
