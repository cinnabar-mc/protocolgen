// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"

	"github.com/go-gl/mathgl/mgl32"
)

// GraphicsOverrideParameter is sent by the server to override graphics parameters.
type GraphicsOverrideParameter struct {
	ParameterKeyframeValues []protocol.OrderedEntry[float32, mgl32.Vec3]
	// FloatValue is an optional single float graphics parameter to be overridden.
	FloatValue protocol.Optional[float32]
	// Vec3Value is an optional single Vec3 graphics parameter to be overridden.
	Vec3Value protocol.Optional[mgl32.Vec3]
	// BiomeIdentifier is the identifier of the biome for which the parameters apply.
	BiomeIdentifier string
	// PlayerIdentifier is the optional identifier of the player for which the override parameter applies.
	PlayerIdentifier       protocol.Optional[string]
	IdentifierForParameter protocol.GraphicsOverrideParameterType
	ResetParameter         bool
}

// ID ...
func (*GraphicsOverrideParameter) ID() uint32 {
	return IDGraphicsOverrideParameter
}

func (pk *GraphicsOverrideParameter) Marshal(io protocol.IO) {
	protocol.OrderedMapLimits(io, &pk.ParameterKeyframeValues, io.Varuint32, 0, 255, io.Float32, io.Vec3)
	protocol.OptionalFunc(io, &pk.FloatValue, io.Float32)
	protocol.OptionalFunc(io, &pk.Vec3Value, io.Vec3)
	io.StringLimits(&pk.BiomeIdentifier, 0, 255)
	protocol.OptionalFunc(io, &pk.PlayerIdentifier, func(value *string) {
		io.StringLimits(value, 0, 255)
	})
	pk.IdentifierForParameter.Marshal(io)
	io.Bool(&pk.ResetParameter)
}
