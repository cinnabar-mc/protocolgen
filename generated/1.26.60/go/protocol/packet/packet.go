// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"fmt"

	"protocolgen/generated/1.26.60/go/protocol"
)

// Packet is the common runtime contract for every generated Bedrock packet.
// Marshal reads from or writes to the supplied protocol IO implementation.
type Packet interface {
	ID() uint32
	Marshal(protocol.IO)
}

// Decode unmarshals one packet and rejects malformed or trailing input.
func Decode(data []byte, pk Packet) error {
	if pk == nil {
		return fmt.Errorf("decode packet <nil>")
	}
	reader := protocol.NewReader(data)
	pk.Marshal(reader)
	if err := reader.Err(); err != nil {
		return fmt.Errorf("decode packet %T: %w", pk, err)
	}
	if remaining := reader.Remaining(); remaining != 0 {
		return fmt.Errorf("decode packet %T: %d trailing bytes", pk, remaining)
	}
	return nil
}

// Encode marshals one packet and reports codec errors.
func Encode(pk Packet) ([]byte, error) {
	if pk == nil {
		return nil, fmt.Errorf("encode packet <nil>")
	}
	writer := protocol.NewWriter()
	pk.Marshal(writer)
	if err := writer.Err(); err != nil {
		return nil, fmt.Errorf("encode packet %T: %w", pk, err)
	}
	return writer.Data(), nil
}
