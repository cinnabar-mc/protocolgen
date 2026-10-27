// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

type AudioContentRegistrationEntry struct {
	AudioContentID  string
	SharedMetadata  AudioSignedAudioContent
	ServerContent   AudioSignedAudioContent
	PlaybackContent AudioSignedAudioContent
}

// Marshal reads or writes AudioContentRegistrationEntry using its canonical wire layout.
func (x *AudioContentRegistrationEntry) Marshal(io IO) {
	io.StringLimits(&x.AudioContentID, 1, 36)
	x.SharedMetadata.Marshal(io)
	x.ServerContent.Marshal(io)
	x.PlaybackContent.Marshal(io)
}

type AudioSignedAudioContent struct {
	CompactJWT string
}

// Marshal reads or writes AudioSignedAudioContent using its canonical wire layout.
func (x *AudioSignedAudioContent) Marshal(io IO) {
	io.StringLimits(&x.CompactJWT, 0, 32768)
}

type AudioSound struct {
	Name                     string
	Position                 BlockPosition
	Volume                   float32
	Pitch                    float32
	LoopCount                int32
	BypassListenerRangeCheck bool
	ServerSoundHandle        Optional[AudioSoundHandle]
	PlaybackPositionSeconds  Optional[float32]
}

// Marshal reads or writes AudioSound using its canonical wire layout.
func (x *AudioSound) Marshal(io IO) {
	io.String(&x.Name)
	x.Position.Marshal(io)
	io.Float32(&x.Volume)
	io.Float32(&x.Pitch)
	io.Varint32(&x.LoopCount)
	io.Bool(&x.BypassListenerRangeCheck)
	OptionalMarshaler(io, &x.ServerSoundHandle)
	OptionalFunc(io, &x.PlaybackPositionSeconds, io.Float32)
}

type SoundDataEvent interface {
	Marshaler
	tagSoundDataEvent() uint32
}

// MarshalSoundDataEvent reads or writes the SoundDataEvent union using its canonical wire layout.
func MarshalSoundDataEvent(io IO, x *SoundDataEvent) {
	Union(io, x, io.Varuint32, SoundDataEvent.tagSoundDataEvent, func(tag uint32) SoundDataEvent {
		switch tag {
		case 0:
			return new(SoundDataEventStop)
		case 1:
			return new(SoundDataEventSetVolume)
		case 2:
			return new(SoundDataEventSetPitch)
		case 3:
			return new(SoundDataEventFade)
		case 4:
			return new(SoundDataEventSeekTo)
		case 5:
			return new(SoundDataEventPause)
		case 6:
			return new(SoundDataEventResume)
		}
		return nil
	})
}

type SoundDataEventFade struct {
	Type         SoundDataEventType
	Duration     float32
	TargetVolume float32
}

func (*SoundDataEventFade) tagSoundDataEvent() uint32 { return 3 }

// Marshal reads or writes SoundDataEventFade using its canonical wire layout.
func (x *SoundDataEventFade) Marshal(io IO) {
	x.Type.Marshal(io)
	io.Float32(&x.Duration)
	io.Float32(&x.TargetVolume)
}

type SoundDataEventPause struct {
	Type SoundDataEventType
}

func (*SoundDataEventPause) tagSoundDataEvent() uint32 { return 5 }

// Marshal reads or writes SoundDataEventPause using its canonical wire layout.
func (x *SoundDataEventPause) Marshal(io IO) {
	x.Type.Marshal(io)
}

type SoundDataEventResume struct {
	Type SoundDataEventType
}

func (*SoundDataEventResume) tagSoundDataEvent() uint32 { return 6 }

// Marshal reads or writes SoundDataEventResume using its canonical wire layout.
func (x *SoundDataEventResume) Marshal(io IO) {
	x.Type.Marshal(io)
}

type SoundDataEventSeekTo struct {
	Type    SoundDataEventType
	Seconds float32
}

func (*SoundDataEventSeekTo) tagSoundDataEvent() uint32 { return 4 }

// Marshal reads or writes SoundDataEventSeekTo using its canonical wire layout.
func (x *SoundDataEventSeekTo) Marshal(io IO) {
	x.Type.Marshal(io)
	io.Float32(&x.Seconds)
}

type SoundDataEventSetPitch struct {
	Type  SoundDataEventType
	Pitch float32
}

func (*SoundDataEventSetPitch) tagSoundDataEvent() uint32 { return 2 }

// Marshal reads or writes SoundDataEventSetPitch using its canonical wire layout.
func (x *SoundDataEventSetPitch) Marshal(io IO) {
	x.Type.Marshal(io)
	io.Float32(&x.Pitch)
}

type SoundDataEventSetVolume struct {
	Type   SoundDataEventType
	Volume float32
}

func (*SoundDataEventSetVolume) tagSoundDataEvent() uint32 { return 1 }

// Marshal reads or writes SoundDataEventSetVolume using its canonical wire layout.
func (x *SoundDataEventSetVolume) Marshal(io IO) {
	x.Type.Marshal(io)
	io.Float32(&x.Volume)
}

type SoundDataEventStop struct {
	Type SoundDataEventType
}

func (*SoundDataEventStop) tagSoundDataEvent() uint32 { return 0 }

// Marshal reads or writes SoundDataEventStop using its canonical wire layout.
func (x *SoundDataEventStop) Marshal(io IO) {
	x.Type.Marshal(io)
}

type SoundDataEventType uint8

const (
	SoundDataEventTypeStop      SoundDataEventType = 0
	SoundDataEventTypeSetVolume SoundDataEventType = 1
	SoundDataEventTypeSetPitch  SoundDataEventType = 2
	SoundDataEventTypeFade      SoundDataEventType = 3
	SoundDataEventTypeSeekTo    SoundDataEventType = 4
	SoundDataEventTypePause     SoundDataEventType = 5
	SoundDataEventTypeResume    SoundDataEventType = 6
)

// Marshal reads or writes SoundDataEventType through its uint8 wire encoding.
func (x *SoundDataEventType) Marshal(io IO) { io.Uint8((*uint8)(x)) }
