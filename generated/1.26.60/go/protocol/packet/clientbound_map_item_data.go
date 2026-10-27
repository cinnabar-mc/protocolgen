// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// ClientBoundMapItemData is sent by the server to the client to update the data of a map shown to the client.
// It is sent with a combination of flags that specify what data is updated. The ClientBoundMapItemData packet
// may be used to update specific parts of the map only. It is not required to send the entire map each time
// when updating one part.
type ClientboundMapItemData struct {
	// MapID is the unique identifier that represents the map that is updated over network. It remains consistent
	// across sessions.
	MapID int64
	// Dimension is the dimension of the map that should be updated, for example the overworld (0), the nether (1)
	// or the end (2).
	Dimension uint8
	// LockedMap specifies if the map that was updated was a locked map, which may be done using a cartography
	// table.
	IsLocked bool
	// Origin is the center position of the map being updated.
	MapOrigin protocol.BlockPos
	// MapsIncludedIn holds an array of map IDs that the map updated is included in. This has to do with the scale
	// of the map: Each map holds its own map ID and all map IDs of maps that include this map and have a bigger
	// scale. This means that a scale 0 map will have 5 map IDs in this slice, whereas a scale 4 map will have
	// only 1 (its own). The actual use of this field remains unknown.
	CreationMapIDs protocol.Optional[[]int64]
	// Scale is the scale of the map as it is shown in-game.
	Scale protocol.Optional[int8]
	// TrackedObjects is a list of tracked objects on the map, which may either be entities or blocks. The client
	// makes sure these tracked objects are actually tracked. (position updated etc.)
	TrackedActorIDs protocol.Optional[[]protocol.MapItemTrackedActorUniqueID]
	// Decorations is a list of fixed decorations located on the map. The decorations will not change client-side,
	// unless the server updates them.
	Decorations protocol.Optional[[]protocol.MapDecoration]
	// Width is the width of the texture area that was updated. The width may be a subset of the total width of
	// the map.
	Width protocol.Optional[int32]
	// Height is the height of the texture area that was updated. The height may be a subset of the total height
	// of the map.
	Height protocol.Optional[int32]
	// XOffset is the X offset in pixels at which the updated texture area starts. From this X, the updated
	// texture will extend exactly Width pixels to the right.
	StartX protocol.Optional[int32]
	// YOffset is the Y offset in pixels at which the updated texture area starts. From this Y, the updated
	// texture will extend exactly Height pixels up.
	StartY protocol.Optional[int32]
	// Pixels is a list of pixel colours for the new texture of the map. It is indexed as Pixels[y*height + x].
	Pixels protocol.Optional[[]uint32]
}

// ID ...
func (*ClientboundMapItemData) ID() uint32 {
	return IDClientboundMapItemData
}

func (pk *ClientboundMapItemData) Marshal(io protocol.IO) {
	io.ActorUniqueID(&pk.MapID)
	io.Uint8(&pk.Dimension)
	io.Bool(&pk.IsLocked)
	pk.MapOrigin.Marshal(io)
	protocol.OptionalFunc(io, &pk.CreationMapIDs, func(value *[]int64) {
		protocol.FuncSliceLimits(io, value, io.Varuint32, 0, 65535, io.ActorUniqueID)
	})
	protocol.OptionalFunc(io, &pk.Scale, io.Int8)
	protocol.OptionalFunc(io, &pk.TrackedActorIDs, func(value *[]protocol.MapItemTrackedActorUniqueID) {
		protocol.SliceLimits(io, value, 0, 65535)
	})
	protocol.OptionalFunc(io, &pk.Decorations, func(value *[]protocol.MapDecoration) {
		protocol.SliceLimits(io, value, 0, 65535)
	})
	protocol.OptionalFunc(io, &pk.Width, io.Varint32)
	protocol.OptionalFunc(io, &pk.Height, io.Varint32)
	protocol.OptionalFunc(io, &pk.StartX, io.Varint32)
	protocol.OptionalFunc(io, &pk.StartY, io.Varint32)
	protocol.OptionalFunc(io, &pk.Pixels, func(value *[]uint32) {
		protocol.FuncSliceLimits(io, value, io.Varuint32, 0, 16384, io.Uint32)
	})
}
