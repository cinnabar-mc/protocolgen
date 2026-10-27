// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

import (
	"image/color"
)

// MapDecoration is a fixed decoration on a map: Its position or other properties do not change automatically
// client-side.
type MapDecoration struct {
	// Type is the type of the map decoration. The type specifies the shape (and sometimes the colour) that the
	// map decoration gets. It is one of the MapDecorationType constants above.
	ImageType MapDecorationType
	// Rotation is the rotation of the map decoration. It is byte due to the 16 fixed directions that the map
	// decoration may face.
	Rotation uint8
	// X is the offset on the X axis in pixels of the decoration.
	X uint8
	// Y is the offset on the Y axis in pixels of the decoration.
	Y uint8
	// Label is the name of the map decoration. This name may be of any value.
	Label string
	// Colour is the colour of the map decoration. Some map decoration types have a specific colour set
	// automatically, whereas others may be changed.
	Color color.RGBA
}

// Marshal reads or writes MapDecoration using its canonical wire layout.
func (x *MapDecoration) Marshal(io IO) {
	x.ImageType.Marshal(io)
	io.Uint8(&x.Rotation)
	io.Uint8(&x.X)
	io.Uint8(&x.Y)
	io.String(&x.Label)
	io.RGBA(&x.Color)
}

// MapDecoration is a fixed decoration on a map: Its position or other properties do not change automatically
// client-side.
type MapDecorationType int8

const (
	MapDecorationTypeMarkerWhite       MapDecorationType = 0
	MapDecorationTypeMarkerGreen       MapDecorationType = 1
	MapDecorationTypeMarkerRed         MapDecorationType = 2
	MapDecorationTypeMarkerBlue        MapDecorationType = 3
	MapDecorationTypeXWhite            MapDecorationType = 4
	MapDecorationTypeTriangleRed       MapDecorationType = 5
	MapDecorationTypeSquareWhite       MapDecorationType = 6
	MapDecorationTypeMarkerSign        MapDecorationType = 7
	MapDecorationTypeMarkerPink        MapDecorationType = 8
	MapDecorationTypeMarkerOrange      MapDecorationType = 9
	MapDecorationTypeMarkerYellow      MapDecorationType = 10
	MapDecorationTypeMarkerTeal        MapDecorationType = 11
	MapDecorationTypeTriangleGreen     MapDecorationType = 12
	MapDecorationTypeSmallSquareWhite  MapDecorationType = 13
	MapDecorationTypeMansion           MapDecorationType = 14
	MapDecorationTypeMonument          MapDecorationType = 15
	MapDecorationTypeNoDraw            MapDecorationType = 16
	MapDecorationTypeVillageDesert     MapDecorationType = 17
	MapDecorationTypeVillagePlains     MapDecorationType = 18
	MapDecorationTypeVillageSavanna    MapDecorationType = 19
	MapDecorationTypeVillageSnowy      MapDecorationType = 20
	MapDecorationTypeVillageTaiga      MapDecorationType = 21
	MapDecorationTypeJungleTemple      MapDecorationType = 22
	MapDecorationTypeWitchHut          MapDecorationType = 23
	MapDecorationTypeTrialChambers     MapDecorationType = 24
	MapDecorationTypeAbandonedCamp     MapDecorationType = 25
	MapDecorationTypeBuriedAncientCity MapDecorationType = 26
	MapDecorationTypeBuriedMineshaft   MapDecorationType = 27
	MapDecorationTypeDesertPyramid     MapDecorationType = 28
	MapDecorationTypeWarmOceanRuins    MapDecorationType = 29
	MapDecorationTypeCount             MapDecorationType = 30
)

// Marshal reads or writes MapDecorationType through its int8 wire encoding.
func (x *MapDecorationType) Marshal(io IO) { io.Int8((*int8)(x)) }

type MapItemTrackedActorType int32

const (
	MapItemTrackedActorTypeEntity      MapItemTrackedActorType = 0
	MapItemTrackedActorTypeBlockEntity MapItemTrackedActorType = 1
	MapItemTrackedActorTypeOther       MapItemTrackedActorType = 2
)

// Marshal reads or writes MapItemTrackedActorType through its int32 wire encoding.
func (x *MapItemTrackedActorType) Marshal(io IO) { io.Int32((*int32)(x)) }

// MapTrackedObject is an object on a map that is 'tracked' by the client, such as an entity or a block. This
// object may move, which is handled client-side.
type MapItemTrackedActorUniqueID struct {
	// Type is the type of the tracked object. It is either MapObjectTypeEntity or MapObjectTypeBlock.
	Type MapItemTrackedActorType
	// EntityUniqueID is the optional unique ID of the tracked entity.
	EntityID Optional[int64]
	// BlockPosition is the optional position of the tracked block.
	BlockPosition Optional[BlockPos]
}

// Marshal reads or writes MapItemTrackedActorUniqueID using its canonical wire layout.
func (x *MapItemTrackedActorUniqueID) Marshal(io IO) {
	x.Type.Marshal(io)
	OptionalFunc(io, &x.EntityID, io.ActorUniqueID)
	OptionalMarshaler(io, &x.BlockPosition)
}

// PixelRequest is the request for the colour of a pixel in a MapInfoRequest packet.
type PixelRequest struct {
	Pixel uint32
	Index uint16
}

// Marshal reads or writes PixelRequest using its canonical wire layout.
func (x *PixelRequest) Marshal(io IO) {
	io.Uint32(&x.Pixel)
	io.Uint16(&x.Index)
}
