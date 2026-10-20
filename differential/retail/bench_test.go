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

type benchEntry struct {
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
		be := benchEntry{payload: e.Payload, header: header, body: buf.Bytes()}
		be.pk, _ = newGophertunnelPacket(f.Sender, header.PacketID)
		if err := decodeGophertunnel(be.pk, be.body, f.Sender == SenderClient); err != nil {
			b.Fatalf("%s: %v", f.File, err)
		}
		out = append(out, be)
	}
	return out
}

// decodeEntry decodes one entry as a gophertunnel connection does.
func decodeEntry(pool gtpacket.Pool, payload []byte, limits bool) {
	buf := bytes.NewBuffer(payload)
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
				decodeEntry(p, e.payload, limits)
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
			var out bytes.Buffer
			encodeEntry(&out, e)
			if !bytes.Equal(out.Bytes(), e.payload) {
				b.Fatalf("%s does not re-encode byte for byte", f.File)
			}
			b.SetBytes(int64(len(e.payload)))
			b.ReportAllocs()
			for b.Loop() {
				out.Reset()
				encodeEntry(&out, e)
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
