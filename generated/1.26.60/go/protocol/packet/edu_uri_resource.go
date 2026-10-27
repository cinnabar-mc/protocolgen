// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

type EduURIResource struct {
	// ButtonName is the button name of the resource URI.
	ButtonName string
	// LinkURI is the link URI for the resource URI.
	LinkURI string
}

// ID ...
func (*EduURIResource) ID() uint32 {
	return IDEduURIResource
}

func (pk *EduURIResource) Marshal(io protocol.IO) {
	io.String(&pk.ButtonName)
	io.String(&pk.LinkURI)
}
