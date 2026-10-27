// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

// AttributeData represents a polymorphic attribute value.
type AttributeData struct {
	MinValue        float32
	MaxValue        float32
	CurrentValue    float32
	DefaultMinValue float32
	DefaultMaxValue float32
	DefaultValue    float32
	Name            string
	Modifiers       []AttributeModifier
}

// Marshal reads or writes AttributeData using its canonical wire layout.
func (x *AttributeData) Marshal(io IO) {
	io.Float32(&x.MinValue)
	io.Float32(&x.MaxValue)
	io.Float32(&x.CurrentValue)
	io.Float32(&x.DefaultMinValue)
	io.Float32(&x.DefaultMaxValue)
	io.Float32(&x.DefaultValue)
	io.String(&x.Name)
	Slice(io, &x.Modifiers)
}

// AttributeLayerData represents a complete attribute layer.
type AttributeLayerData struct {
	AttributeLayers []EASAttributeLayerData
}

func (*AttributeLayerData) tagAttributeLayerSyncData() uint32 { return 0 }

// Marshal reads or writes AttributeLayerData using its canonical wire layout.
func (x *AttributeLayerData) Marshal(io IO) {
	SliceLimits(io, &x.AttributeLayers, 0, 512)
}

// AttributeLayerSettings represents settings for an attribute layer.
type AttributeLayerSettings struct {
	AttributeLayerName      string
	AttributeLayerDimension DimensionType
	AttributesLayerSettings EASAttributeLayerSettings
}

func (*AttributeLayerSettings) tagAttributeLayerSyncData() uint32 { return 1 }

// Marshal reads or writes AttributeLayerSettings using its canonical wire layout.
func (x *AttributeLayerSettings) Marshal(io IO) {
	io.StringLimits(&x.AttributeLayerName, 0, 128)
	x.AttributeLayerDimension.Marshal(io)
	x.AttributesLayerSettings.Marshal(io)
}

type AttributeLayerSyncData interface {
	Marshaler
	tagAttributeLayerSyncData() uint32
}

// MarshalAttributeLayerSyncData reads or writes the AttributeLayerSyncData union using its canonical wire layout.
func MarshalAttributeLayerSyncData(io IO, x *AttributeLayerSyncData) {
	Union(io, x, io.Varuint32, AttributeLayerSyncData.tagAttributeLayerSyncData, func(tag uint32) AttributeLayerSyncData {
		switch tag {
		case 0:
			return new(AttributeLayerData)
		case 1:
			return new(AttributeLayerSettings)
		case 2:
			return new(EnvironmentAttributeData)
		case 3:
			return new(RemoveEnvironmentAttributes)
		}
		return nil
	})
}

type AttributeValue interface {
	Marshaler
	tagAttributeValue() uint32
}

// MarshalAttributeValue reads or writes the AttributeValue union using its canonical wire layout.
func MarshalAttributeValue(io IO, x *AttributeValue) {
	Union(io, x, io.Varuint32, AttributeValue.tagAttributeValue, func(tag uint32) AttributeValue {
		switch tag {
		case 0:
			return new(BooleanAttributeValue)
		case 1:
			return new(FloatAttributeValue)
		case 2:
			return new(ColourAttributeValue)
		}
		return nil
	})
}

type ConstantEnvironmentAttribute struct {
	Attribute AttributeValue
}

func (*ConstantEnvironmentAttribute) tagEnvironmentAttribute() uint32 { return 0 }

// Marshal reads or writes ConstantEnvironmentAttribute using its canonical wire layout.
func (x *ConstantEnvironmentAttribute) Marshal(io IO) {
	MarshalAttributeValue(io, &x.Attribute)
}

// NoiseAlignment represents the way the noise of an environment attribute transition is aligned.
type EASNoiseAlignment struct {
	Type  EASNoiseAlignmentType
	Value uint32
}

// Marshal reads or writes EASNoiseAlignment using its canonical wire layout.
func (x *EASNoiseAlignment) Marshal(io IO) {
	x.Type.Marshal(io)
	io.Varuint32(&x.Value)
}

type EASNoiseAlignmentType uint8

const (
	EASNoiseAlignmentTypeMinLocalTransitionEnd EASNoiseAlignmentType = 0
)

// Marshal reads or writes EASNoiseAlignmentType through its uint8 wire encoding.
func (x *EASNoiseAlignmentType) Marshal(io IO) { io.Uint8((*uint8)(x)) }

type EASNoiseTransitionSettingsData struct {
	TotalTransitionTicks   uint32
	CurrentTransitionTicks uint32
	Easing                 EasingType
	ClockName              string
	LocalTransitionTicks   uint32
	NoiseName              string
	NoiseAlignment         EASNoiseAlignment
}

// Marshal reads or writes EASNoiseTransitionSettingsData using its canonical wire layout.
func (x *EASNoiseTransitionSettingsData) Marshal(io IO) {
	io.Varuint32(&x.TotalTransitionTicks)
	io.Varuint32(&x.CurrentTransitionTicks)
	x.Easing.Marshal(io)
	io.StringLimits(&x.ClockName, 0, 128)
	io.Varuint32(&x.LocalTransitionTicks)
	io.StringLimits(&x.NoiseName, 0, 128)
	x.NoiseAlignment.Marshal(io)
}

type EASTransitionSettingsData struct {
	TotalTransitionTicks   uint32
	CurrentTransitionTicks uint32
	Easing                 EasingType
	ClockName              string
}

// Marshal reads or writes EASTransitionSettingsData using its canonical wire layout.
func (x *EASTransitionSettingsData) Marshal(io IO) {
	io.Varuint32(&x.TotalTransitionTicks)
	io.Varuint32(&x.CurrentTransitionTicks)
	x.Easing.Marshal(io)
	io.StringLimits(&x.ClockName, 0, 128)
}

type EasingType int32

const (
	EasingTypeLinear       EasingType = 0
	EasingTypeSpring       EasingType = 1
	EasingTypeInQuad       EasingType = 2
	EasingTypeOutQuad      EasingType = 3
	EasingTypeInOutQuad    EasingType = 4
	EasingTypeInCubic      EasingType = 5
	EasingTypeOutCubic     EasingType = 6
	EasingTypeInOutCubic   EasingType = 7
	EasingTypeInQuart      EasingType = 8
	EasingTypeOutQuart     EasingType = 9
	EasingTypeInOutQuart   EasingType = 10
	EasingTypeInQuint      EasingType = 11
	EasingTypeOutQuint     EasingType = 12
	EasingTypeInOutQuint   EasingType = 13
	EasingTypeInSine       EasingType = 14
	EasingTypeOutSine      EasingType = 15
	EasingTypeInOutSine    EasingType = 16
	EasingTypeInExpo       EasingType = 17
	EasingTypeOutExpo      EasingType = 18
	EasingTypeInOutExpo    EasingType = 19
	EasingTypeInCirc       EasingType = 20
	EasingTypeOutCirc      EasingType = 21
	EasingTypeInOutCirc    EasingType = 22
	EasingTypeInBounce     EasingType = 23
	EasingTypeOutBounce    EasingType = 24
	EasingTypeInOutBounce  EasingType = 25
	EasingTypeInBack       EasingType = 26
	EasingTypeOutBack      EasingType = 27
	EasingTypeInOutBack    EasingType = 28
	EasingTypeInElastic    EasingType = 29
	EasingTypeOutElastic   EasingType = 30
	EasingTypeInOutElastic EasingType = 31
)

// Marshal reads or writes EasingType through its int32 wire encoding.
func (x *EasingType) Marshal(io IO) { io.Varint32((*int32)(x)) }

type EnvironmentAttribute interface {
	Marshaler
	tagEnvironmentAttribute() uint32
}

// MarshalEnvironmentAttribute reads or writes the EnvironmentAttribute union using its canonical wire layout.
func MarshalEnvironmentAttribute(io IO, x *EnvironmentAttribute) {
	Union(io, x, io.Varuint32, EnvironmentAttribute.tagEnvironmentAttribute, func(tag uint32) EnvironmentAttribute {
		switch tag {
		case 0:
			return new(ConstantEnvironmentAttribute)
		case 1:
			return new(TransitionEnvironmentAttribute)
		case 2:
			return new(NoiseTransitionEnvironmentAttribute)
		}
		return nil
	})
}

// EnvironmentAttributeData represents an environment attribute with optional transition data.
type EnvironmentAttributeData struct {
	AttributeLayerName      string
	AttributeLayerDimension DimensionType
	Attributes              []EASEnvironmentAttributeData
}

func (*EnvironmentAttributeData) tagAttributeLayerSyncData() uint32 { return 2 }

// Marshal reads or writes EnvironmentAttributeData using its canonical wire layout.
func (x *EnvironmentAttributeData) Marshal(io IO) {
	io.StringLimits(&x.AttributeLayerName, 0, 128)
	x.AttributeLayerDimension.Marshal(io)
	SliceLimits(io, &x.Attributes, 0, 1024)
}

type NoiseTransitionEnvironmentAttribute struct {
	FromAttribute AttributeValue
	ToAttribute   AttributeValue
	Settings      EASNoiseTransitionSettingsData
}

func (*NoiseTransitionEnvironmentAttribute) tagEnvironmentAttribute() uint32 { return 2 }

// Marshal reads or writes NoiseTransitionEnvironmentAttribute using its canonical wire layout.
func (x *NoiseTransitionEnvironmentAttribute) Marshal(io IO) {
	MarshalAttributeValue(io, &x.FromAttribute)
	MarshalAttributeValue(io, &x.ToAttribute)
	x.Settings.Marshal(io)
}

type TransitionEnvironmentAttribute struct {
	FromAttribute AttributeValue
	ToAttribute   AttributeValue
	Settings      EASTransitionSettingsData
}

func (*TransitionEnvironmentAttribute) tagEnvironmentAttribute() uint32 { return 1 }

// Marshal reads or writes TransitionEnvironmentAttribute using its canonical wire layout.
func (x *TransitionEnvironmentAttribute) Marshal(io IO) {
	MarshalAttributeValue(io, &x.FromAttribute)
	MarshalAttributeValue(io, &x.ToAttribute)
	x.Settings.Marshal(io)
}
