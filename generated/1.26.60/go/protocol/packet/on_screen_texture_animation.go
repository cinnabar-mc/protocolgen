// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// OnScreenTextureAnimation is sent by the server to show a certain animation on the screen of the player. The
// packet is used, as an example, for when a raid is triggered and when a raid is defeated.
type OnScreenTextureAnimation struct {
	// EffectID is the type of the animation to show. The packet provides no further extra data to allow modifying
	// the duration or other properties of the animation.
	EffectID uint32
}

// ID ...
func (*OnScreenTextureAnimation) ID() uint32 {
	return IDOnScreenTextureAnimation
}

func (pk *OnScreenTextureAnimation) Marshal(io protocol.IO) {
	io.Uint32(&pk.EffectID)
}
