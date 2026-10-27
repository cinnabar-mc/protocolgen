// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// GameTestResults is a packet sent in response to the GameTestRequest packet, with a boolean indicating
// whether the test was successful or not, and an error string if the test failed.
type GameTestResults struct {
	// Succeeded indicates whether the test succeeded or not.
	Succeeded bool
	// Error is the error that occurred. If Succeeded is true, this field is empty.
	Error string
	// Name represents the name of the test.
	TestName string
}

// ID ...
func (*GameTestResults) ID() uint32 {
	return IDGameTestResults
}

func (pk *GameTestResults) Marshal(io protocol.IO) {
	io.Bool(&pk.Succeeded)
	io.String(&pk.Error)
	io.String(&pk.TestName)
}
