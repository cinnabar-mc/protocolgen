// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// AnimateEntity is sent by the server to animate an entity client-side. It may be used to play a single
// animation, or to activate a controller which can start a sequence of animations based on different
// conditions specified in an animation controller. Much of the documentation of this packet can be found at
// https://learn.microsoft.com/en-us/minecraft/creator/reference/content/animationsreference
type AnimateEntity struct {
	// MAnimation is the name of a single animation to start playing.
	MAnimation string
	// MNextState is the first state to start with. These states are declared in animation controllers (which, in
	// themselves, are animations too). These states in turn may have animations and transitions to move to a next
	// state.
	MNextState string
	// MStopExpression is a MoLang expression that specifies when the animation should be stopped.
	MStopExpression string
	// MStopExpressionVersion is the MoLang stop condition version.
	MStopExpressionVersion int32
	// MController is the animation controller that is used to manage animations. These controllers decide when to
	// play which animation.
	MController string
	// MBlendOutTime does not currently seem to be used.
	MBlendOutTime float32
	// MRuntimeIds is list of runtime IDs of entities that the animation should be applied to.
	MRuntimeIds []uint64
}

// ID ...
func (*AnimateEntity) ID() uint32 {
	return IDAnimateEntity
}

func (pk *AnimateEntity) Marshal(io protocol.IO) {
	io.String(&pk.MAnimation)
	io.String(&pk.MNextState)
	io.String(&pk.MStopExpression)
	io.Int32(&pk.MStopExpressionVersion)
	io.String(&pk.MController)
	io.Float32(&pk.MBlendOutTime)
	protocol.FuncSlice(io, &pk.MRuntimeIds, io.Varuint32, io.ActorRuntimeID)
}
