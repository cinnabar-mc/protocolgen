// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// CodeBuilderSource is an Education Edition packet sent by the client to the server to run an operation with
// a code builder.
type CodeBuilderSource struct {
	// Operation is used to distinguish the operation performed. It is always one of the constants listed above.
	Operation protocol.CodeBuilderStorageQueryOptionsOperation
	// Category is used to distinguish the category of the operation performed. It is always one of the constants
	// listed above.
	Category protocol.CodeBuilderStorageQueryOptionsCategory
	// CodeStatus is the status of the code builder. It is always one of the constants listed above.
	CodeStatus protocol.CodeBuilderExecutionStateCodeStatus
}

// ID ...
func (*CodeBuilderSource) ID() uint32 {
	return IDCodeBuilderSource
}

func (pk *CodeBuilderSource) Marshal(io protocol.IO) {
	pk.Operation.Marshal(io)
	pk.Category.Marshal(io)
	pk.CodeStatus.Marshal(io)
}
