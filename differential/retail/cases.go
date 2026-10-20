package retail

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strings"

	gtprotocol "github.com/sandertv/gophertunnel/minecraft/protocol"
	gtpacket "github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	genpacket "protocolgen/generated/1.26.51/go/protocol/packet"
)

// BuildInputs returns the retail batches and the cases derived from them.
// unknownID must be an id that no decoder defines.
func BuildInputs(corpus Corpus, batches map[string][]byte, unknownID uint32) ([]Input, error) {
	var inputs []Input
	for _, fixture := range corpus.Fixtures {
		batch, ok := batches[fixture.File]
		if !ok {
			return nil, fmt.Errorf("fixture %s was not read", fixture.File)
		}
		entries, err := SplitBatch(batch)
		if err != nil {
			return nil, fmt.Errorf("fixture %s: %w", fixture.File, err)
		}
		add := func(caseName, sender string, batch []byte) {
			inputs = append(inputs, Input{Fixture: fixture.File, Case: caseName, Sender: sender, Batch: batch})
		}
		add(CaseRetail, fixture.Sender, batch)
		add(CaseUnknownID, fixture.Sender, JoinBatch(mapEntries(entries, func(e Entry) Entry {
			return NewEntry(e.Header&^idMask|unknownID, e.Body)
		})))
		if fixture.Direction != "bidirectional" {
			add(CaseWrongSender, otherSender(fixture.Sender), batch)
		}
		if allBodies(entries) {
			add(CaseTruncated, fixture.Sender, JoinBatch(mapEntries(entries, func(e Entry) Entry {
				return NewEntry(e.Header, e.Body[:len(e.Body)-1])
			})))
		}
		add(CaseTrailing, fixture.Sender, JoinBatch(mapEntries(entries, func(e Entry) Entry {
			return NewEntry(e.Header, append(append([]byte(nil), e.Body...), 0))
		})))
	}
	return inputs, nil
}

// LargeInputs encodes, with the generated Go codec, a server packet whose
// collection needs a reservation well above the allocation probe's ceiling:
// 3000 strings of at least 32 bytes each in the contract's Rust model.
func LargeInputs() ([]Input, error) {
	body, err := genpacket.Encode(&genpacket.UpdateSoftEnum{EnumName: "harness", Values: make([]string, 3000)})
	if err != nil {
		return nil, err
	}
	return []Input{{
		Fixture: "update_soft_enum_3000_values.bin",
		Case:    CaseLarge,
		Sender:  SenderServer,
		Batch:   JoinBatch([]Entry{NewEntry(genpacket.IDUpdateSoftEnum, body)}),
	}}, nil
}

// UnusedPacketID returns the highest ten-bit packet id for which known is
// false.
func UnusedPacketID(known func(uint32) bool) (uint32, error) {
	for id := uint32(idMask); id > 0; id-- {
		if !known(id) {
			return id, nil
		}
	}
	return 0, errors.New("every packet id is defined")
}

func mapEntries(entries []Entry, f func(Entry) Entry) []Entry {
	out := make([]Entry, len(entries))
	for i, entry := range entries {
		out[i] = f(entry)
	}
	return out
}

func allBodies(entries []Entry) bool {
	for _, entry := range entries {
		if len(entry.Body) == 0 {
			return false
		}
	}
	return len(entries) != 0
}

func otherSender(sender string) string {
	if sender == SenderServer {
		return SenderClient
	}
	return SenderServer
}

// OversizedInputs returns entries at and just past the client collection and
// byte-slice ceilings, server packets far larger than any limit, and entries
// whose declared lengths exceed the input. Each decoder's accept or reject
// decision must match gophertunnel's.
func OversizedInputs() ([]Input, error) {
	actions := func(n int) []gtprotocol.PlayerBlockAction { return make([]gtprotocol.PlayerBlockAction, n) }
	blockActions := func(n int) *gtpacket.PlayerAuthInput {
		pk := &gtpacket.PlayerAuthInput{InputData: gtprotocol.NewInputFlags(gtpacket.InputFlagCount), BlockActions: gtprotocol.Option(actions(n))}
		pk.InputData.Set(gtpacket.InputFlagPerformBlockActions)
		return pk
	}
	requests := func(n int) *gtpacket.ItemStackRequest {
		// One action each: the manifest also requires at least one.
		pk := &gtpacket.ItemStackRequest{Requests: make([]gtprotocol.ItemStackRequest, n)}
		for i := range pk.Requests {
			pk.Requests[i].Actions = []gtprotocol.StackRequestAction{&gtprotocol.DropStackRequestAction{Count: 1}}
		}
		return pk
	}
	// gophertunnel's maxByteSliceLength, which is also its batch ceiling, so a
	// slice past it is rejected with the whole batch.
	const byteSliceCeiling = 16 << 20
	cases := []struct {
		name   string
		sender string
		pk     gtpacket.Packet
	}{
		{"player_auth_input_block_actions_100", SenderClient, blockActions(100)},
		{"player_auth_input_block_actions_101", SenderClient, blockActions(101)},
		{"player_auth_input_block_actions_4096", SenderClient, blockActions(4096)},
		{"player_auth_input_block_actions_4097", SenderClient, blockActions(4097)},
		{"item_stack_request_100", SenderClient, requests(100)},
		{"item_stack_request_101", SenderClient, requests(101)},
		{"item_stack_request_4096", SenderClient, requests(4096)},
		{"item_stack_request_4097", SenderClient, requests(4097)},
		{"login_request_16mib_minus_64", SenderClient, &gtpacket.Login{ClientProtocol: 2193, ConnectionRequest: make([]byte, byteSliceCeiling-64)}},
		{"login_request_16mib_plus_1", SenderClient, &gtpacket.Login{ClientProtocol: 2193, ConnectionRequest: make([]byte, byteSliceCeiling+1)}},
		{"level_chunk_payload_8mib", SenderServer, &gtpacket.LevelChunk{SubChunkCount: 24, RawPayload: make([]byte, 8<<20)}},
	}
	var inputs []Input
	for _, c := range cases {
		payload, err := encodeGophertunnel(gtpacket.Header{PacketID: c.pk.ID()}, c.pk)
		if err != nil {
			return nil, fmt.Errorf("encode %s: %w", c.name, err)
		}
		inputs = append(inputs, Input{Fixture: "oversized_" + c.name + ".bin", Case: CaseOversized, Sender: c.sender, Batch: JoinBatch([]Entry{{Payload: payload}})})
	}
	// gophertunnel refuses to write more than four translation parameters or
	// a message over 65536 bytes, so those are spliced into its encodings.
	for _, n := range []int{4, 5, 4097} {
		entry, err := splicedText(&gtpacket.Text{TextType: gtpacket.TextTypeTranslation, NeedsTranslation: true, Message: "chat.type.text", Parameters: make([]string, 4)},
			[]byte("chat.type.text\x04\x00\x00\x00\x00"), append(append([]byte("chat.type.text"), appendVaruint32(nil, uint32(n))...), make([]byte, n)...))
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, Input{Fixture: fmt.Sprintf("oversized_text_parameters_%d.bin", n), Case: CaseOversized, Sender: SenderClient, Batch: JoinBatch([]Entry{entry})})
	}
	for _, n := range []int{65536, 65537} {
		entry, err := splicedText(&gtpacket.Text{TextType: gtpacket.TextTypeRaw, Message: "x"},
			[]byte("\x01x"), append(appendVaruint32(nil, uint32(n)), strings.Repeat("m", n)...))
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, Input{Fixture: fmt.Sprintf("oversized_text_message_%d.bin", n), Case: CaseOversized, Sender: SenderServer, Batch: JoinBatch([]Entry{entry})})
	}
	// Declared lengths far beyond the entry, with nothing after them: a string
	// and a byte slice claiming 2^32-1 bytes, and collections claiming 2^20
	// elements. The generated Go reserves a server collection's declared
	// count before reading it, so a 2^32-1 count would exhaust the runner.
	huge := appendVaruint32(nil, math.MaxUint32)
	million := appendVaruint32(nil, 1<<20)
	declared := []struct {
		name   string
		sender string
		id     uint32
		body   []byte
	}{
		{"text_message_length_4gib", SenderServer, gtpacket.IDText, append([]byte{0, 0, gtpacket.TextTypeRaw}, huge...)},
		{"available_commands_enum_values_1m", SenderServer, gtpacket.IDAvailableCommands, million},
		{"login_request_length_4gib", SenderClient, gtpacket.IDLogin, append(binary.BigEndian.AppendUint32(nil, 2193), huge...)},
		{"crafting_data_recipes_1m", SenderServer, gtpacket.IDCraftingData, million},
	}
	for _, d := range declared {
		inputs = append(inputs, Input{Fixture: "oversized_" + d.name + ".bin", Case: CaseOversized, Sender: d.sender, Batch: JoinBatch([]Entry{NewEntry(d.id, d.body)})})
	}
	return inputs, nil
}

// splicedText encodes pk with gophertunnel and replaces the one occurrence
// of old in the entry with replacement.
func splicedText(pk *gtpacket.Text, old, replacement []byte) (Entry, error) {
	payload, err := encodeGophertunnel(gtpacket.Header{PacketID: pk.ID()}, pk)
	if err != nil {
		return Entry{}, err
	}
	if bytes.Count(payload, old) != 1 {
		return Entry{}, fmt.Errorf("text entry %x does not hold %x once", payload, old)
	}
	spliced := bytes.Replace(payload, old, replacement, 1)
	header, n, err := readVaruint32(spliced)
	if err != nil {
		return Entry{}, err
	}
	return Entry{Header: header, Payload: spliced, Body: spliced[n:]}, nil
}
