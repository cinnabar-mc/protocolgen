// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

import (
	"github.com/google/uuid"
)

type PackIDVersion struct {
	PackUUID    uuid.UUID
	PackVersion SemVersion
}

// Marshal reads or writes PackIDVersion using its canonical wire layout.
func (x *PackIDVersion) Marshal(io IO) {
	io.UUID(&x.PackUUID)
	x.PackVersion.Marshal(io)
}

type PackIDVersionData struct {
	PackUUID    uuid.UUID
	PackVersion SemVersionData
}

// Marshal reads or writes PackIDVersionData using its canonical wire layout.
func (x *PackIDVersionData) Marshal(io IO) {
	io.UUID(&x.PackUUID)
	x.PackVersion.Marshal(io)
}

type PackInfoData struct {
	PackIDVersion       PackIDVersionData
	PackSize            uint64
	ContentKey          string
	SubpackName         string
	ContentIdentity     ContentIdentity
	HasScripts          bool
	IsAddonPack         bool
	IsRayTracingCapable bool
	CDNURL              string
}

// Marshal reads or writes PackInfoData using its canonical wire layout.
func (x *PackInfoData) Marshal(io IO) {
	x.PackIDVersion.Marshal(io)
	io.Uint64(&x.PackSize)
	io.String(&x.ContentKey)
	io.String(&x.SubpackName)
	x.ContentIdentity.Marshal(io)
	io.Bool(&x.HasScripts)
	io.Bool(&x.IsAddonPack)
	io.Bool(&x.IsRayTracingCapable)
	io.String(&x.CDNURL)
}

// StackResourcePack represents a resource pack sent on the stack of the client. When sent, the client will
// apply them in the order of the stack sent.
type PackInstanceID struct {
	// UUID is the UUID of the resource pack. Each resource pack downloaded must have a different UUID in order
	// for the client to be able to handle them properly.
	PackID string
	// Version is the version of the resource pack. The client will cache resource packs sent by the server as
	// long as they carry the same version. Sending a resource pack with a different version than previously will
	// force the client to re-download it.
	Version string
	// SubPackName ...
	SubPackName string
}

// Marshal reads or writes PackInstanceID using its canonical wire layout.
func (x *PackInstanceID) Marshal(io IO) {
	io.String(&x.PackID)
	io.String(&x.Version)
	io.String(&x.SubPackName)
}
