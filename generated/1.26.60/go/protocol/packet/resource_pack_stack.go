// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// ResourcePackStack is sent by the server to send the order in which resource packs and behaviour packs
// should be applied (and downloaded) by the client.
type ResourcePackStack struct {
	// TexturePackRequired specifies if the client must accept the texture packs the server has in order to join
	// the server. If set to true, the client gets the option to either download the resource packs and join, or
	// quit entirely. Behaviour packs never have to be downloaded.
	TexturePackRequired bool
	// TexturePacks is a list of texture packs that the client needs to download before joining the server. The
	// order of these texture packs specifies the order that they are applied in on the client side. The first in
	// the list will be applied first.
	TexturePackList []protocol.PackInstanceID
	// BaseGameVersion is the vanilla version that the client should set its resource pack stack to.
	BaseGameVersion string
	// Experiments holds a list of experiments that are either enabled or disabled in the world that the player
	// spawns in. It is not clear why experiments are sent both here and in the StartGame packet.
	Experiments protocol.Experiments
	// IncludeEditorPacks specifies if vanilla editor packs should be included in the resource pack stack when
	// connecting to an editor world.
	IncludeEditorPacks bool
}

// ID ...
func (*ResourcePackStack) ID() uint32 {
	return IDResourcePackStack
}

func (pk *ResourcePackStack) Marshal(io protocol.IO) {
	io.Bool(&pk.TexturePackRequired)
	protocol.SliceLimits(io, &pk.TexturePackList, 0, 65535)
	io.String(&pk.BaseGameVersion)
	pk.Experiments.Marshal(io)
	io.Bool(&pk.IncludeEditorPacks)
}
