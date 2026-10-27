// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// NetworkStackLatency is sent by the server (and the client, on development builds) to measure the latency
// over the entire Minecraft stack, rather than the RakNet latency. It has other usages too, such as the
// ability to be used as some kind of acknowledgement packet, to know when the client has received a certain
// other packet.
type NetworkStackLatency struct {
	// Timestamp is the timestamp of the network stack latency packet. The client will, if NeedsResponse is set to
	// true, send a NetworkStackLatency packet with this same timestamp packet in response.
	CreationTime uint64
	// NeedsResponse specifies if the sending side of this packet wants a response to the packet, meaning that the
	// other side should send a NetworkStackLatency packet back.
	IsFromServer bool
}

// ID ...
func (*NetworkStackLatency) ID() uint32 {
	return IDNetworkStackLatency
}

func (pk *NetworkStackLatency) Marshal(io protocol.IO) {
	io.Uint64(&pk.CreationTime)
	io.Bool(&pk.IsFromServer)
}
