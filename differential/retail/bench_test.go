package retail

import (
	"bytes"
	"testing"

	gtprotocol "github.com/sandertv/gophertunnel/minecraft/protocol"
	gtpacket "github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// The gophertunnel half of the codec benches. Each case reads the same
// committed testdata/bench file, under the same Group/Name, as the criterion
// benches in bench/benches/codec.rs, so bench-compare can join the two.
//
// Both sides time the same boundaries. Decode reads one length-prefixed
// entry: the length prefix, the header varint and the body, rejecting unread
// bytes. Encode writes one length-prefixed entry into a reused buffer from a
// packet decoded before timing. The batch benches go through each library's
// batch API: the 0xfe marker, every length prefix and every entry.

type benchEntry struct {
	framed  []byte // length prefix, header varint and body
	payload []byte // header varint and body
	header  gtpacket.Header
	body    []byte
	pk      gtpacket.Packet // decoded once, for the encode benches
}

func benchFiles(b *testing.B) []SyntheticFile {
	files, err := ReadSyntheticCorpus(SyntheticDir)
	if err != nil {
		b.Fatal(err)
	}
	return files
}

func benchEntries(b *testing.B, f SyntheticFile) []benchEntry {
	entries, err := SplitBatch(f.Batch)
	if err != nil {
		b.Fatal(err)
	}
	out := make([]benchEntry, 0, len(entries))
	for _, e := range entries {
		buf := bytes.NewBuffer(e.Payload)
		var header gtpacket.Header
		if err := header.Read(buf); err != nil {
			b.Fatal(err)
		}
		framed := appendVaruint32(nil, uint32(len(e.Payload)))
		be := benchEntry{framed: append(framed, e.Payload...), payload: e.Payload, header: header, body: buf.Bytes()}
		be.pk, _ = newGophertunnelPacket(f.Sender, header.PacketID)
		if err := decodeGophertunnel(be.pk, be.body, f.Sender == SenderClient); err != nil {
			b.Fatalf("%s: %v", f.File, err)
		}
		out = append(out, be)
	}
	return out
}

// decodeFramed decodes one length-prefixed entry.
func decodeFramed(pool gtpacket.Pool, framed []byte, limits bool) {
	buf := bytes.NewBuffer(framed)
	var length uint32
	if err := gtprotocol.Varuint32(buf, &length); err != nil || int(length) != buf.Len() {
		panic("bad length prefix")
	}
	decodeBuffer(pool, buf, limits)
}

// decodeEntry decodes one entry, header varint and body, as a gophertunnel
// connection does with each entry of a batch.
func decodeEntry(pool gtpacket.Pool, payload []byte, limits bool) {
	decodeBuffer(pool, bytes.NewBuffer(payload), limits)
}

func decodeBuffer(pool gtpacket.Pool, buf *bytes.Buffer, limits bool) {
	var header gtpacket.Header
	if err := header.Read(buf); err != nil {
		panic(err)
	}
	pk := pool[header.PacketID]()
	pk.Marshal(gtprotocol.NewReader(buf, 0, limits))
	if buf.Len() != 0 {
		panic("unread bytes")
	}
}

func encodeEntry(out *bytes.Buffer, e benchEntry) {
	if err := e.header.Write(out); err != nil {
		panic(err)
	}
	e.pk.Marshal(gtprotocol.NewWriter(out, 0))
}

// encodeFramed writes one length-prefixed entry into out, through scratch as
// gophertunnel's encoder does: the entry is marshalled, then framed.
func encodeFramed(out, scratch *bytes.Buffer, e benchEntry) {
	scratch.Reset()
	encodeEntry(scratch, e)
	out.Reset()
	if err := gtprotocol.WriteVaruint32(out, uint32(scratch.Len())); err != nil {
		panic(err)
	}
	out.Write(scratch.Bytes())
}

func pool(sender string) gtpacket.Pool {
	if sender == SenderClient {
		return gophertunnelClientPool
	}
	return gophertunnelServerPool
}

func BenchmarkDecode(b *testing.B) {
	for _, f := range benchFiles(b) {
		if f.Set == SetBatch {
			continue
		}
		e := benchEntries(b, f)[0]
		p, limits := pool(f.Sender), f.Sender == SenderClient
		b.Run(f.Name, func(b *testing.B) {
			b.SetBytes(int64(len(e.payload)))
			b.ReportAllocs()
			for b.Loop() {
				decodeFramed(p, e.framed, limits)
			}
		})
	}
}

func BenchmarkEncode(b *testing.B) {
	for _, f := range benchFiles(b) {
		if f.Set == SetBatch {
			continue
		}
		e := benchEntries(b, f)[0]
		b.Run(f.Name, func(b *testing.B) {
			var out, scratch bytes.Buffer
			encodeFramed(&out, &scratch, e)
			if !bytes.Equal(out.Bytes(), e.framed) {
				b.Fatalf("%s does not re-encode byte for byte", f.File)
			}
			b.SetBytes(int64(len(e.payload)))
			b.ReportAllocs()
			for b.Loop() {
				encodeFramed(&out, &scratch, e)
			}
		})
	}
}

func BenchmarkDecodeBatch(b *testing.B) {
	for _, f := range benchFiles(b) {
		if f.Set != SetBatch {
			continue
		}
		p, limits := pool(f.Sender), f.Sender == SenderClient
		b.Run(f.Name, func(b *testing.B) {
			decoder := gtpacket.NewDecoder(&batchSource{f.Batch})
			decode := func(payload []byte) error {
				decodeEntry(p, payload, limits)
				return nil
			}
			b.SetBytes(int64(len(f.Batch)))
			b.ReportAllocs()
			for b.Loop() {
				if err := decoder.DecodeFunc(decode); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkEncodeBatch(b *testing.B) {
	for _, f := range benchFiles(b) {
		if f.Set != SetBatch {
			continue
		}
		entries := benchEntries(b, f)
		b.Run(f.Name, func(b *testing.B) {
			var out bytes.Buffer
			encoder := gtpacket.NewEncoder(&out)
			bufs := make([]bytes.Buffer, len(entries))
			payloads := make([][]byte, len(entries))
			encode := func() {
				out.Reset()
				for i, e := range entries {
					bufs[i].Reset()
					encodeEntry(&bufs[i], e)
					payloads[i] = bufs[i].Bytes()
				}
				if err := encoder.Encode(payloads); err != nil {
					b.Fatal(err)
				}
			}
			encode()
			if !bytes.Equal(out.Bytes(), f.Batch) {
				b.Fatalf("%s does not re-encode byte for byte", f.File)
			}
			b.SetBytes(int64(len(f.Batch)))
			b.ReportAllocs()
			for b.Loop() {
				encode()
			}
		})
	}
}
