package retail

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
	"strings"

	gtprotocol "github.com/sandertv/gophertunnel/minecraft/protocol"
	gtpacket "github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	genprotocol "protocolgen/generated/1.26.51/go/protocol"
	genpacket "protocolgen/generated/1.26.51/go/protocol/packet"
)

var (
	gophertunnelServerPool = gtpacket.NewServerPool()
	gophertunnelClientPool = gtpacket.NewClientPool()
)

// KnownPacket reports whether the generated Go codec or gophertunnel defines
// id for either sender.
func KnownPacket(id uint32) bool {
	_, generated := genpacket.NewPacket(id)
	_, server := gophertunnelServerPool[id]
	_, client := gophertunnelClientPool[id]
	return generated || server || client
}

// DecodeGenerated runs the generated Go codec over in, with the generated
// reader's default limits for client packets and none for server packets.
// For retail inputs it also runs the element-limit probe.
func DecodeGenerated(in Input) []Verdict {
	verdicts := splitAndDecode(in, in.Case, func(e Entry, v *Verdict) {
		decodeGeneratedEntry(in.Sender, e, v, generatedReader(in.Sender))
	})
	if in.Case == CaseRetail {
		verdicts = append(verdicts, splitAndDecode(in, ProbeElementLimit, func(e Entry, v *Verdict) {
			decodeGeneratedEntry(in.Sender, e, v, func(data []byte) *genprotocol.Reader {
				return genprotocol.NewReaderWithLimit(data, 0)
			})
		})...)
	}
	return verdicts
}

func generatedReader(sender string) func([]byte) *genprotocol.Reader {
	if sender == SenderClient {
		return genprotocol.NewReader
	}
	return genprotocol.NewReaderWithoutLimit
}

// splitAndDecode frames in with the harness splitter, for codecs without a
// batch reader of their own.
func splitAndDecode(in Input, caseName string, decode func(Entry, *Verdict)) []Verdict {
	entries, err := SplitBatch(in.Batch)
	verdicts := make([]Verdict, 0, len(entries)+1)
	for i, entry := range entries {
		v := Verdict{Fixture: in.Fixture, Case: caseName, Entry: i, Header: HeaderText(entry.Header), Payload: Digest(entry.Payload)}
		decode(entry, &v)
		verdicts = append(verdicts, v)
	}
	if err != nil {
		verdicts = append(verdicts, Verdict{Fixture: in.Fixture, Case: caseName, Entry: len(entries), Outcome: OutcomeFrameError, Detail: err.Error()})
	}
	return verdicts
}

// decodeGeneratedEntry treats an id outside the manifest as an unknown packet
// forwarded unchanged, and a manifest id the sender may not use as a direction
// error.
func decodeGeneratedEntry(sender string, e Entry, v *Verdict, newReader func([]byte) *genprotocol.Reader) {
	id := e.Header & idMask
	if _, known := genpacket.NewPacket(id); !known {
		v.Outcome, v.Encoded = OutcomeUnknown, Digest(e.Payload)
		return
	}
	pk, ok := newGeneratedPacket(sender, id)
	if !ok {
		v.fail(ClassDirection, fmt.Errorf("packet %d is not sent by the %s", id, sender))
		return
	}
	if err := decodeGenerated(pk, e.Body, newReader); err != nil {
		v.fail(generatedClass(err), err)
		return
	}
	v.Outcome, v.Fields = OutcomeOK, renderDigest(pk)
	body, err := genpacket.Encode(pk)
	if err != nil {
		v.encodeFailed(err)
		return
	}
	v.Encoded = Digest(NewEntry(e.Header, body).Payload)
	v.RoundTrip = RoundTripDiffers
	again, _ := newGeneratedPacket(sender, id)
	if decodeGenerated(again, body, newReader) == nil && renderDigest(again) == v.Fields {
		v.RoundTrip = RoundTripSame
	}
}

func newGeneratedPacket(sender string, id uint32) (genpacket.Packet, bool) {
	if sender == SenderClient {
		return genpacket.NewClientPacket(id)
	}
	return genpacket.NewServerPacket(id)
}

func decodeGenerated(pk genpacket.Packet, body []byte, newReader func([]byte) *genprotocol.Reader) error {
	return recoverError(func() error {
		reader := newReader(body)
		pk.Marshal(reader)
		if err := reader.Err(); err != nil {
			return err
		}
		if remaining := reader.Remaining(); remaining != 0 {
			return fmt.Errorf("%d trailing bytes", remaining)
		}
		return nil
	})
}

func generatedClass(err error) string {
	switch text := err.Error(); {
	case strings.Contains(text, "exceeds decoder limit"):
		return ClassLimit
	case strings.Contains(text, "trailing bytes"):
		return ClassTrailing
	default:
		return ClassDecode
	}
}

// batchSource hands packet.Decoder one whole batch per read, as a RakNet or
// NetherNet connection does; a plain io.Reader would cap batches at the
// decoder's 3 MiB read buffer.
type batchSource struct{ batch []byte }

func (s *batchSource) Read(p []byte) (int, error)  { return copy(p, s.batch), nil }
func (s *batchSource) ReadPacket() ([]byte, error) { return s.batch, nil }

// DecodeGophertunnel runs the pinned gophertunnel as its connections do:
// batch framing by packet.Decoder, the sender's pool with packet.Unknown for
// absent ids, and reader limits only for client packets.
func DecodeGophertunnel(in Input) []Verdict {
	payloads, err := gtpacket.NewDecoder(&batchSource{in.Batch}).Decode()
	if err != nil {
		return []Verdict{{Fixture: in.Fixture, Case: in.Case, Outcome: OutcomeFrameError, Detail: err.Error()}}
	}
	verdicts := make([]Verdict, 0, len(payloads))
	for i, payload := range payloads {
		v := Verdict{Fixture: in.Fixture, Case: in.Case, Entry: i, Payload: Digest(payload)}
		decodeGophertunnelEntry(in.Sender, payload, &v)
		verdicts = append(verdicts, v)
	}
	return verdicts
}

func decodeGophertunnelEntry(sender string, payload []byte, v *Verdict) {
	buf := bytes.NewBuffer(append([]byte(nil), payload...))
	var header gtpacket.Header
	if err := header.Read(buf); err != nil {
		v.Outcome, v.Detail = OutcomeFrameError, err.Error()
		return
	}
	v.Header = fmt.Sprintf("%d:%d:%d", header.PacketID, header.SenderSubClient, header.TargetSubClient)
	body := buf.Bytes()
	pk, known := newGophertunnelPacket(sender, header.PacketID)
	limits := sender == SenderClient
	if err := decodeGophertunnel(pk, body, limits); err != nil {
		v.fail(gophertunnelClass(err), err)
		return
	}
	v.Outcome, v.Fields = OutcomeOK, renderDigest(pk)
	if !known {
		v.Outcome = OutcomeUnknown
	}
	encoded, err := encodeGophertunnel(header, pk)
	if err != nil {
		v.encodeFailed(err)
		return
	}
	v.Encoded = Digest(encoded)
	v.RoundTrip = RoundTripDiffers
	again, _ := newGophertunnelPacket(sender, header.PacketID)
	var reread gtpacket.Header
	rest := bytes.NewBuffer(encoded)
	if reread.Read(rest) == nil && decodeGophertunnel(again, rest.Bytes(), limits) == nil && renderDigest(again) == v.Fields {
		v.RoundTrip = RoundTripSame
	}
}

func newGophertunnelPacket(sender string, id uint32) (gtpacket.Packet, bool) {
	pool := gophertunnelServerPool
	if sender == SenderClient {
		pool = gophertunnelClientPool
	}
	if factory, ok := pool[id]; ok {
		return factory(), true
	}
	return &gtpacket.Unknown{PacketID: id}, false
}

func decodeGophertunnel(pk gtpacket.Packet, body []byte, limits bool) error {
	buf := bytes.NewBuffer(append([]byte(nil), body...))
	return recoverError(func() error {
		pk.Marshal(gtprotocol.NewReader(buf, 0, limits))
		if buf.Len() != 0 {
			return fmt.Errorf("%d unread bytes left", buf.Len())
		}
		return nil
	})
}

func encodeGophertunnel(header gtpacket.Header, pk gtpacket.Packet) ([]byte, error) {
	var out bytes.Buffer
	if err := header.Write(&out); err != nil {
		return nil, err
	}
	err := recoverError(func() error {
		pk.Marshal(gtprotocol.NewWriter(&out, 0))
		return nil
	})
	return out.Bytes(), err
}

func gophertunnelClass(err error) string {
	switch text := err.Error(); {
	case strings.Contains(text, "slice length was too long"):
		return ClassLimit
	case strings.Contains(text, "unread bytes left"):
		return ClassTrailing
	default:
		return ClassDecode
	}
}

// panicError is a panic recovered from a decoder that is not one of
// gophertunnel's error-carrying reader panics.
type panicError struct{ value any }

func (p *panicError) Error() string { return fmt.Sprintf("panicked: %v", p.value) }

// recoverError runs fn, returning gophertunnel's reader panics, which carry
// decode errors, as errors and any other panic as a *panicError.
func recoverError(fn func() error) (err error) {
	defer func() {
		recovered := recover()
		if recovered == nil {
			return
		}
		if value, ok := recovered.(error); ok {
			var runtimeErr runtime.Error
			if !errors.As(value, &runtimeErr) {
				err = value
				return
			}
		}
		err = &panicError{recovered}
	}()
	return fn()
}
