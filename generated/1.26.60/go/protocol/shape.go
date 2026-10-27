// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

import (
	"image/color"

	"github.com/go-gl/mathgl/mgl32"
)

// PrimitiveShape defines a single shape to be rendered on the client. Each shape has a unique NetworkID and a
// set of optional parameters depending on its type.
type PrimitiveShape struct {
	// NetworkID is the network ID of the shape.
	NetworkID uint64
	// ShapeType is the optional dimension ID where the shape is rendered.
	ShapeType Optional[ScriptModuleMinecraftScriptPrimitiveShapeType]
	// Location is the location of the shape.
	Location Optional[mgl32.Vec3]
	// Scale is the scale of the shape.
	Scale Optional[float32]
	// Rotation is the rotation of the shape.
	Rotation Optional[mgl32.Vec3]
	// TotalTimeLeft is the total time left of the shape.
	TotalTimeLeft Optional[float32]
	// MaximumRenderDistance is the rotation of the shape.
	MaximumRenderDistance Optional[float32]
	// Color is the total time left of the shape.
	Color Optional[color.RGBA]
	// DimensionID is the optional dimension ID where the shape is rendered.
	DimensionID Optional[DimensionType]
	// AttachedToEntityID is the optional runtime ID of the entity the shape is attached to.
	AttachedToEntityID Optional[int64]
	// ExtraShapeData holding data specific to the type of shape (such as text string for the text shape).
	ExtraShapeData PrimitiveShapeExtraShapeData
}

// Marshal reads or writes PrimitiveShape using its canonical wire layout.
func (x *PrimitiveShape) Marshal(io IO) {
	io.Varuint64(&x.NetworkID)
	OptionalMarshaler(io, &x.ShapeType)
	OptionalFunc(io, &x.Location, io.Vec3)
	OptionalFunc(io, &x.Scale, io.Float32)
	OptionalFunc(io, &x.Rotation, io.Vec3)
	OptionalFunc(io, &x.TotalTimeLeft, io.Float32)
	OptionalFunc(io, &x.MaximumRenderDistance, io.Float32)
	OptionalFunc(io, &x.Color, io.RGBA)
	OptionalMarshaler(io, &x.DimensionID)
	OptionalFunc(io, &x.AttachedToEntityID, io.ActorUniqueID)
	MarshalPrimitiveShapeExtraShapeData(io, &x.ExtraShapeData)
}

// TextShape represents a text debug shape.
type TextShape struct {
	// Text is the text of the debug text shape.
	Text string
	// UseRotation is if the text should use the provided rotation, meaning it will be static and does not follow
	// the camera. Use false for default behaviour.
	UseRotation bool
	// BackgroundColor is the RGBA colour to use for the text background. This is a translucent black colour by
	// default.
	BackgroundColor Optional[color.RGBA]
	// LineGapHeight is the gap to leave between each line of multiline text. If not set, the client uses its
	// default gap.
	LineGapHeight float32
	// DepthTest is whether the text should show through walls. Use true for default behaviour.
	DepthTest bool
	// ShowBackface is if the background should render on the back side of the shape. This only has a visible
	// effect when UseRotation is true since you cannot see the back side of the text otherwise. Use true for
	// default behaviour.
	ShowBackface bool
	// ShowTextBackface is if the text should render on the back side of the shape. This only has a visible effect
	// when UseRotation is true since you cannot see the back side of the text otherwise. Use true for default
	// behaviour.
	ShowTextBackface bool
}

func (*TextShape) tagPrimitiveShapeExtraShapeData() uint32 { return 2 }

// Marshal reads or writes TextShape using its canonical wire layout.
func (x *TextShape) Marshal(io IO) {
	io.String(&x.Text)
	io.Bool(&x.UseRotation)
	OptionalFunc(io, &x.BackgroundColor, io.RGBA)
	io.Float32(&x.LineGapHeight)
	io.Bool(&x.DepthTest)
	io.Bool(&x.ShowBackface)
	io.Bool(&x.ShowTextBackface)
}
