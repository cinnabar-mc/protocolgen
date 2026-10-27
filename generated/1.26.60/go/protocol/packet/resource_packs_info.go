// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// ResourcePacksInfo is sent by the server to inform the client on what resource packs the server has. It
// sends a list of the resource packs it has and basic information on them like the version and description.
type ResourcePacksInfo struct {
	ResourcePackRequired bool
	HasAddonPacks        bool
	// HasScripts specifies if any of the resource packs contain scripts in them. If set to true, only clients
	// that support scripts will be able to download them.
	HasScripts bool
	// ForceDisableVibrantVisuals specifies if the vibrant visuals feature should be forcibly disabled on the
	// server. If set to true, the server will ensure that vibrant visuals are not enabled, regardless of the
	// client's settings.
	ForceDisableVibrantVisuals bool
	WorldTemplateIDAndVersion  protocol.PackIDVersion
	ResourcePacks              []protocol.PackInfoData
}

// ID ...
func (*ResourcePacksInfo) ID() uint32 {
	return IDResourcePacksInfo
}

func (pk *ResourcePacksInfo) Marshal(io protocol.IO) {
	io.Bool(&pk.ResourcePackRequired)
	io.Bool(&pk.HasAddonPacks)
	io.Bool(&pk.HasScripts)
	io.Bool(&pk.ForceDisableVibrantVisuals)
	pk.WorldTemplateIDAndVersion.Marshal(io)
	protocol.SliceLimits(io, &pk.ResourcePacks, 0, 65535)
}
