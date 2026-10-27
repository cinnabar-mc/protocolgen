// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

// SyncWorldClockStateData represents the state data for synchronising a world clock.
type SyncWorldClockStateData struct {
	// ClockID is the unique identifier for the clock.
	ClockID uint64
	// Time is the current time of the clock.
	Time int32
	// IsPaused indicates if the clock is paused.
	IsPaused bool
}

// Marshal reads or writes SyncWorldClockStateData using its canonical wire layout.
func (x *SyncWorldClockStateData) Marshal(io IO) {
	io.Varuint64(&x.ClockID)
	io.Varint32(&x.Time)
	io.Bool(&x.IsPaused)
}

type SyncWorldClocksData interface {
	Marshaler
	tagSyncWorldClocksData() uint32
}

// MarshalSyncWorldClocksData reads or writes the SyncWorldClocksData union using its canonical wire layout.
func MarshalSyncWorldClocksData(io IO, x *SyncWorldClocksData) {
	Union(io, x, io.Varuint32, SyncWorldClocksData.tagSyncWorldClocksData, func(tag uint32) SyncWorldClocksData {
		switch tag {
		case 0:
			return new(SyncStateData)
		case 1:
			return new(InitializeRegistryData)
		case 2:
			return new(AddTimeMarkerData)
		case 3:
			return new(RemoveTimeMarkerData)
		}
		return nil
	})
}

// TimeMarkerData represents a time marker within a world clock.
type TimeMarkerData struct {
	// ID is the unique identifier for the time marker.
	ID uint64
	// Name is the name of the time marker.
	Name string
	// Time is the time at which the marker is set.
	Time int32
	// Period is the optional period for the time marker.
	Period Optional[int32]
}

// Marshal reads or writes TimeMarkerData using its canonical wire layout.
func (x *TimeMarkerData) Marshal(io IO) {
	io.Varuint64(&x.ID)
	io.StringLimits(&x.Name, 0, 128)
	io.Varint32(&x.Time)
	OptionalFunc(io, &x.Period, io.Int32)
}

// WorldClockData represents a complete world clock with its time markers.
type WorldClockData struct {
	// ID is the unique identifier for the clock.
	ID uint64
	// Name is the name of the clock.
	Name string
	// Time is the current time of the clock.
	Time int32
	// IsPaused indicates if the clock is paused.
	IsPaused bool
	// TimeMarkers is a list of time markers for this clock.
	TimeMarkers []TimeMarkerData
}

// Marshal reads or writes WorldClockData using its canonical wire layout.
func (x *WorldClockData) Marshal(io IO) {
	io.Varuint64(&x.ID)
	io.StringLimits(&x.Name, 0, 128)
	io.Varint32(&x.Time)
	io.Bool(&x.IsPaused)
	SliceLimits(io, &x.TimeMarkers, 0, 256)
}
