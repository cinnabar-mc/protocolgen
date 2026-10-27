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
	tagSoundDataEvent() uint8
}

// MarshalSoundDataEvent reads or writes the SoundDataEvent union using its canonical wire layout.
func MarshalSoundDataEvent(io IO, x *SoundDataEvent) {
	Union(io, x, io.Uint8, SoundDataEvent.tagSoundDataEvent, func(tag uint8) SoundDataEvent {
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
	Duration     float32
	TargetVolume float32
}

func (*SoundDataEventFade) tagSoundDataEvent() uint8 { return 3 }

// Marshal reads or writes SoundDataEventFade using its canonical wire layout.
func (x *SoundDataEventFade) Marshal(io IO) {
	io.Float32(&x.Duration)
	io.Float32(&x.TargetVolume)
}

type SoundDataEventPause struct {
}

func (*SoundDataEventPause) tagSoundDataEvent() uint8 { return 5 }

// Marshal reads or writes SoundDataEventPause using its canonical wire layout.
func (x *SoundDataEventPause) Marshal(io IO) {
}

type SoundDataEventResume struct {
}

func (*SoundDataEventResume) tagSoundDataEvent() uint8 { return 6 }

// Marshal reads or writes SoundDataEventResume using its canonical wire layout.
func (x *SoundDataEventResume) Marshal(io IO) {
}

type SoundDataEventSeekTo struct {
	Seconds float32
}

func (*SoundDataEventSeekTo) tagSoundDataEvent() uint8 { return 4 }

// Marshal reads or writes SoundDataEventSeekTo using its canonical wire layout.
func (x *SoundDataEventSeekTo) Marshal(io IO) {
	io.Float32(&x.Seconds)
}

type SoundDataEventSetPitch struct {
	Pitch float32
}

func (*SoundDataEventSetPitch) tagSoundDataEvent() uint8 { return 2 }

// Marshal reads or writes SoundDataEventSetPitch using its canonical wire layout.
func (x *SoundDataEventSetPitch) Marshal(io IO) {
	io.Float32(&x.Pitch)
}

type SoundDataEventSetVolume struct {
	Volume float32
}

func (*SoundDataEventSetVolume) tagSoundDataEvent() uint8 { return 1 }

// Marshal reads or writes SoundDataEventSetVolume using its canonical wire layout.
func (x *SoundDataEventSetVolume) Marshal(io IO) {
	io.Float32(&x.Volume)
}

type SoundDataEventStop struct {
}

func (*SoundDataEventStop) tagSoundDataEvent() uint8 { return 0 }

// Marshal reads or writes SoundDataEventStop using its canonical wire layout.
func (x *SoundDataEventStop) Marshal(io IO) {
}
