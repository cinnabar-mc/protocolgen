// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// NetworkSettings is sent by the server to update a variety of network settings. These settings modify the
// way packets are sent over the network stack.
type NetworkSettings struct {
	// CompressionThreshold is the minimum size of a packet that is compressed when sent. If the size of a packet
	// is under this value, it is not compressed. When set to 0, all packets will be left uncompressed.
	CompressionThreshold uint16
	// CompressionAlgorithm is the algorithm that is used to compress packets.
	CompressionAlgorithm protocol.PacketCompressionAlgorithm
	// ClientThrottle regulates whether the client should throttle players when exceeding of the threshold.
	// Players outside threshold will not be ticked, improving performance on low-end devices.
	ClientThrottleEnabled bool
	// ClientThrottleThreshold is the threshold for client throttling. If the number of players exceeds this
	// value, the client will throttle players.
	ClientThrottleThreshold uint8
	// ClientThrottleScalar is the scalar for client throttling. The scalar is the amount of players that are
	// ticked when throttling is enabled.
	ClientThrottleScalar float32
}

// ID ...
func (*NetworkSettings) ID() uint32 {
	return IDNetworkSettings
}

func (pk *NetworkSettings) Marshal(io protocol.IO) {
	io.Uint16(&pk.CompressionThreshold)
	pk.CompressionAlgorithm.Marshal(io)
	io.Bool(&pk.ClientThrottleEnabled)
	io.Uint8(&pk.ClientThrottleThreshold)
	io.Float32(&pk.ClientThrottleScalar)
}
