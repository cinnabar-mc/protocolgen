// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package packet

import (
	"protocolgen/generated/1.26.60/go/protocol"
)

// InventoryTransaction is a packet sent by the client. It essentially exists out of multiple sub-packets,
// each of which have something to do with the inventory in one way or another. Some of these sub-packets
// directly relate to the inventory, others relate to interaction with the world, that could potentially
// result in a change in the inventory.
type InventoryTransaction struct {
	// LegacyRequestID is an ID that is only non-zero at times when sent by the client. The server should always
	// send 0 for this. When this field is not 0, the LegacySetItemSlots slice below will have values in it.
	// LegacyRequestID ties in with the ItemStackResponse packet. If this field is non-0, the server should
	// respond with an ItemStackResponse packet. Some inventory actions such as dropping an item out of the hotbar
	// are still one using this packet, and the ItemStackResponse packet needs to tie in with it.
	LegacyRequestID protocol.ItemStackLegacyRequestID
	// LegacySetItemSlots are only present if the LegacyRequestID is non-zero. These item slots inform the server
	// of the slots that were changed during the inventory transaction, and the server should send back an
	// ItemStackResponse packet with these slots present in it. (Or false with no slots, if rejected.)
	LegacySetItemSlots protocol.Optional[[]protocol.LegacySetSlot]
	// TransactionData is a data object that holds data specific to the type of transaction that the
	// TransactionPacket held. Its concrete type must be one of NormalTransactionData, MismatchTransactionData
	// UseItemTransactionData, UseItemOnEntityTransactionData or ReleaseItemTransactionData. If nil is set, the
	// transaction will be assumed to of type InventoryTransactionTypeNormal.
	Transaction protocol.InventoryTransactionPacketData
}

// ID ...
func (*InventoryTransaction) ID() uint32 {
	return IDInventoryTransaction
}

func (pk *InventoryTransaction) Marshal(io protocol.IO) {
	pk.LegacyRequestID.Marshal(io)
	protocol.OptionalFunc(io, &pk.LegacySetItemSlots, func(value *[]protocol.LegacySetSlot) {
		protocol.Slice(io, value)
	})
	protocol.MarshalInventoryTransactionPacketData(io, &pk.Transaction)
}
