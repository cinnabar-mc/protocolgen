// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

import (
	"github.com/google/uuid"
)

// ChainedSubcommand represents a subcommand that can have chained commands, such as /execute which allows you
// to run another command as another entity or at a different position etc.
type ChainedSubcommand struct {
	// Name is the name of the chained subcommand and shows up in the list as a regular subcommand enum.
	Name string
	// SubCommandValues contains the index and parameter type of the chained subcommand.
	SubCommandValues []ChainedSubcommandValue
}

// Marshal reads or writes ChainedSubcommand using its canonical wire layout.
func (x *ChainedSubcommand) Marshal(io IO) {
	io.StringLimits(&x.Name, 0, 512)
	SliceLimits(io, &x.SubCommandValues, 0, 32)
}

// ChainedSubcommandValue represents the value for a chained subcommand argument.
type ChainedSubcommandValue struct {
	// SubCommandFirstValue is the index of the argument in the ChainedSubcommandValues slice from the
	// AvailableCommands packet. This is then used to set the type specified by the Value field below.
	SubCommandFirstValue uint32
	// SubCommandSecondValue is a combination of the flags above and specified the type of argument. Unlike
	// regular parameter types, this should NOT contain any of the special flags (valid, enum, suffixed or soft
	// enum) but only the basic types.
	SubCommandSecondValue uint32
}

// Marshal reads or writes ChainedSubcommandValue using its canonical wire layout.
func (x *ChainedSubcommandValue) Marshal(io IO) {
	io.Varuint32(&x.SubCommandFirstValue)
	io.Varuint32(&x.SubCommandSecondValue)
}

type CommandBlockUpdateData interface {
	Marshaler
	tagCommandBlockUpdateData() uint32
}

// MarshalCommandBlockUpdateData reads or writes the CommandBlockUpdateData union using its canonical wire layout.
func MarshalCommandBlockUpdateData(io IO, x *CommandBlockUpdateData) {
	Union(io, x, io.Varuint32, CommandBlockUpdateData.tagCommandBlockUpdateData, func(tag uint32) CommandBlockUpdateData {
		switch tag {
		case 0:
			return new(EntityCommandTarget)
		case 1:
			return new(BlockCommandData)
		}
		return nil
	})
}

// CommandEnum represents an enum in a command usage. The enum typically has a type and a set of options that
// are valid. A value that is not one of the options results in a failure during execution.
type CommandEnum struct {
	// Name is the type of the command enum. The type will show up in the command usage as the type of the
	// argument if it has a certain amount of arguments, or when Options is set to true in the command holding the
	// enum.
	Name string
	// Values holds a list of indices that point to the EnumValues slice in the AvailableCommandsPacket. These
	// represent the options of the enum.
	Values []uint32
}

// Marshal reads or writes CommandEnum using its canonical wire layout.
func (x *CommandEnum) Marshal(io IO) {
	io.String(&x.Name)
	FuncSlice(io, &x.Values, io.Varuint32, io.Uint32)
}

// CommandEnumConstraint is sent in the AvailableCommands packet to limit what values of an enum may be used
// taking in account things such as whether cheats are enabled.
type CommandEnumConstraint struct {
	// EnumValueSymbol points to an enum value in the AvailableCommands packet that this constraint should apply
	// to.
	EnumValueSymbol uint32
	// EnumSymbol points to an enum in the AvailableCommands packet to which this constraint should apply to.
	EnumSymbol uint32
	// ConstraintIndices holds a slice of constraints as present in the constants above.
	ConstraintIndices []uint8
}

// Marshal reads or writes CommandEnumConstraint using its canonical wire layout.
func (x *CommandEnumConstraint) Marshal(io IO) {
	io.Uint32(&x.EnumValueSymbol)
	io.Uint32(&x.EnumSymbol)
	io.ByteSliceLimits(&x.ConstraintIndices, 0, 250)
}

// CommandOrigin holds data that identifies the origin of the requesting of a command. It holds several fields
// that may be used to get specific information. When sent in a CommandRequest packet, the same CommandOrigin
// should be sent in a CommandOutput packet.
type CommandOriginData struct {
	Type string
	// UUID is a unique identifier for every instantiation of a command.
	UUID uuid.UUID
	// RequestID is an ID that identifies the request of the client. The server should send a CommandOrigin with
	// the same request ID to ensure it can be matched with the request by the caller of the command. This is
	// especially important for websocket servers and it seems that this field is only non-empty for these
	// websocket servers.
	RequestID string
	PlayerID  int64
}

// Marshal reads or writes CommandOriginData using its canonical wire layout.
func (x *CommandOriginData) Marshal(io IO) {
	io.String(&x.Type)
	io.UUID(&x.UUID)
	io.StringLimits(&x.RequestID, 0, 39)
	io.Int64(&x.PlayerID)
}

type CommandOutputData struct {
	OutputType     string
	SuccessCount   uint32
	OutputMessages []CommandOutputMessage
	DataSet        Optional[string]
}

// Marshal reads or writes CommandOutputData using its canonical wire layout.
func (x *CommandOutputData) Marshal(io IO) {
	io.String(&x.OutputType)
	io.Uint32(&x.SuccessCount)
	Slice(io, &x.OutputMessages)
	OptionalFunc(io, &x.DataSet, io.String)
}

// CommandOutputMessage represents a message sent by a command that holds the output of one of the commands
// executed.
type CommandOutputMessage struct {
	// Message is the message that is sent to the client in the chat window. It may either be simply a message or
	// a translated built-in string like 'commands.tp.success.coordinates', combined with specific parameters
	// below.
	MessageID string
	// Success indicates if the output message was one of a successful command execution. If set to true, the
	// output message is by default coloured white, whereas if set to false, the message is by default coloured
	// red.
	Successful bool
	// Parameters is a list of parameters that serve to supply the message sent with additional information, such
	// as the position that a player was teleported to or the effect that was applied to an entity. These
	// parameters only apply for the Minecraft built-in command output.
	Parameters []string
}

// Marshal reads or writes CommandOutputMessage using its canonical wire layout.
func (x *CommandOutputMessage) Marshal(io IO) {
	io.String(&x.MessageID)
	io.Bool(&x.Successful)
	FuncSlice(io, &x.Parameters, io.Varuint32, io.String)
}

// CommandOverload represents an overload of a command. This overload can be compared to function overloading
// in languages such as java. It represents a single usage of the command. A command may have multiple
// different overloads, which are handled differently.
type CommandOverload struct {
	// IsChaining determines if the parameters use chained subcommands or not.
	IsChaining bool
	// ParameterData is a list of command parameters that are part of the overload. These parameters specify the
	// usage of the command when this overload is applied.
	ParameterData []CommandParameter
}

// Marshal reads or writes CommandOverload using its canonical wire layout.
func (x *CommandOverload) Marshal(io IO) {
	io.Bool(&x.IsChaining)
	Slice(io, &x.ParameterData)
}

// CommandParameter represents a single parameter of a command overload, which accepts a certain type of input
// values. It has a name and a type which show up client-side when a player is entering the command.
type CommandParameter struct {
	// Name is the name of the command parameter. It shows up in the usage like <$Name: $Type>, with the exception
	// of enum types, which show up simply as a list of options if the list is short enough and Options is set to
	// false.
	Name string
	// Type is a rather odd combination of type(flag)s that result in a certain parameter type to show up
	// client-side. It is a combination of the flags above. The basic types must be combined with the
	// ArgumentTypeFlagBasic flag (and integers with a suffix ArgumentTypeFlagSuffixed), whereas enums are
	// combined with the ArgumentTypeFlagEnum flag.
	ParseSymbol uint32
	// Optional specifies if the command parameter is optional to enter. Note that no non-optional parameter
	// should ever be present in a command overload after an optional parameter. When optional, the parameter
	// shows up like so: [$Name: $Type], whereas when mandatory, it shows up like so: <$Name: $Type>.
	IsOptional bool
	// Options holds a combinations of options that additionally apply to the command parameter. The list of
	// options can be found above.
	Options uint8
}

// Marshal reads or writes CommandParameter using its canonical wire layout.
func (x *CommandParameter) Marshal(io IO) {
	io.String(&x.Name)
	io.Uint32(&x.ParseSymbol)
	io.Bool(&x.IsOptional)
	io.Uint8(&x.Options)
}

type CommandPermissionLevel uint8

const (
	CommandPermissionLevelAny           CommandPermissionLevel = 0
	CommandPermissionLevelGameDirectors CommandPermissionLevel = 1
	CommandPermissionLevelAdmin         CommandPermissionLevel = 2
	CommandPermissionLevelHost          CommandPermissionLevel = 3
	CommandPermissionLevelOwner         CommandPermissionLevel = 4
	CommandPermissionLevelInternal      CommandPermissionLevel = 5
)

// Marshal reads or writes CommandPermissionLevel through its uint8 wire encoding.
func (x *CommandPermissionLevel) Marshal(io IO) { io.Uint8((*uint8)(x)) }

// DynamicEnum is an enum variant that can have its options changed during runtime, without sending a new
// AvailableCommands packet.
type DynamicEnum struct {
	// EnumName is the type of the command enum. The type will show up in the command usage as the type of the
	// argument if it has a certain amount of arguments, or when Options is set to true in the command holding the
	// enum.
	EnumName string
	// EnumOptions is a slice of possible options for the enum.
	EnumOptions []string
}

// Marshal reads or writes DynamicEnum using its canonical wire layout.
func (x *DynamicEnum) Marshal(io IO) {
	io.StringLimits(&x.EnumName, 0, 512)
	FuncSlice(io, &x.EnumOptions, io.Varuint32, io.String)
}
