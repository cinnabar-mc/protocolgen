package differential

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"testing"

	gophertunnelpacket "github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	generatedpacket "protocolgen/generated/1.26.60/go/protocol/packet"
)

func TestDifferentialCorpus(t *testing.T) {
	samples, err := loadCorpus()
	if err != nil {
		t.Fatal(err)
	}
	for _, sample := range samples {
		if _, ok := generatedFactory(sample.direction, sample.id); !ok {
			continue
		}
		if _, ok := oracleFactory(sample.direction, sample.id); !ok {
			continue
		}
		got := comparePayload(sample.direction, sample.id, sample.data)
		if err := assertAccepted(got.issues); err != nil {
			t.Fatalf("%s packet %d input %x: %v", sample.direction, sample.id, sample.data, err)
		}
	}
}

func TestCorpusCoversSharedPacketPools(t *testing.T) {
	if *promoteFlag {
		t.Skip("promotion fills missing corpus seeds")
	}
	samples, err := loadCorpus()
	if err != nil {
		t.Fatal(err)
	}
	covered := make(map[direction]map[uint32]bool)
	for _, sample := range samples {
		if covered[sample.direction] == nil {
			covered[sample.direction] = make(map[uint32]bool)
		}
		covered[sample.direction][sample.id] = true
	}
	for _, dir := range []direction{directionClient, directionServer} {
		generated := generatedPacketPool(dir)
		oracle := oraclePacketPool(dir)
		for id := range generated {
			if _, ok := oracle[id]; ok && !covered[dir][id] {
				t.Errorf("%s packet %d has no committed corpus seed", dir, id)
			}
		}
	}
}

func TestPromoteZeroCorpus(t *testing.T) {
	if !*promoteFlag {
		t.Skip("pass -promote to write zero-value seeds")
	}
	samples, err := loadCorpus()
	if err != nil {
		t.Fatal(err)
	}
	covered := make(map[direction]map[uint32]bool)
	for _, sample := range samples {
		if covered[sample.direction] == nil {
			covered[sample.direction] = make(map[uint32]bool)
		}
		covered[sample.direction][sample.id] = true
	}
	for _, dir := range []direction{directionClient, directionServer} {
		var pool generatedpacket.Pool
		if dir == directionClient {
			pool = generatedpacket.NewClientPool()
		} else {
			pool = generatedpacket.NewServerPool()
		}
		ids := make([]uint32, 0, len(pool))
		for id := range pool {
			ids = append(ids, id)
		}
		sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
		for _, id := range ids {
			if _, ok := oracleFactory(dir, id); !ok {
				continue
			}
			if covered[dir][id] {
				continue
			}
			pk, ok := pool[id]
			if !ok {
				continue
			}
			candidates := make([][]byte, 0, 130)
			if data, err := generatedpacket.Encode(pk()); err == nil {
				candidates = append(candidates, data)
			}
			candidates = append(candidates, nil)
			for length := 1; length <= 128; length++ {
				candidates = append(candidates, make([]byte, length))
			}
			var fallback []byte
			var fallbackSet bool
			var selected []byte
			var selectedSet bool
			for _, data := range candidates {
				got := comparePayload(dir, id, data)
				if assertAccepted(got.issues) != nil {
					continue
				}
				if !fallbackSet {
					fallback = data
					fallbackSet = true
				}
				if got.generated.ok && got.oracle.ok {
					selected = data
					selectedSet = true
					break
				}
			}
			if !selectedSet && fallbackSet {
				selected = fallback
				selectedSet = true
			}
			if !selectedSet {
				t.Errorf("no non-divergent seed candidate for %s packet %d", dir, id)
				continue
			}
			if err := promoteCorpus(corpusSample{direction: dir, id: id, data: selected}); err != nil {
				t.Fatalf("promote packet %d: %v", id, err)
			}
		}
	}
}

func generatedPacketPool(dir direction) generatedpacket.Pool {
	if dir == directionClient {
		return generatedpacket.NewClientPool()
	}
	return generatedpacket.NewServerPool()
}

func oraclePacketPool(dir direction) gophertunnelpacket.Pool {
	if dir == directionClient {
		return gophertunnelpacket.NewClientPool()
	}
	return gophertunnelpacket.NewServerPool()
}

func FuzzDifferential(f *testing.F) {
	samples, err := loadCorpus()
	if err != nil {
		f.Fatalf("load corpus: %v", err)
	}
	for _, sample := range samples {
		f.Add(uint8(directionCode(sample.direction)), sample.id, sample.data)
	}
	f.Add(uint8(directionCode(directionServer)), uint32(2), mustDecodeHex(f, "00000000"))
	f.Fuzz(func(t *testing.T, directionValue uint8, id uint32, data []byte) {
		if len(data) > 64<<10 {
			return
		}
		dir, ok := decodeDirection(directionValue)
		if !ok {
			return
		}
		if _, ok := generatedFactory(dir, id); !ok {
			return
		}
		if _, ok := oracleFactory(dir, id); !ok {
			return
		}
		got := comparePayload(dir, id, data)
		if *promoteFlag && got.generated.ok && got.oracle.ok {
			if err := promoteCorpus(corpusSample{direction: dir, id: id, data: data}); err != nil {
				t.Fatalf("promote corpus: %v", err)
			}
		}
		if *triageFlag {
			if err := recordTriage(got.issues); err != nil {
				t.Fatalf("record triage: %v", err)
			}
			return
		}
		if err := assertAccepted(got.issues); err != nil {
			t.Fatalf("%v", err)
		}
	})
}

func directionCode(dir direction) byte {
	if dir == directionServer {
		return 1
	}
	return 0
}

func decodeDirection(value uint8) (direction, bool) {
	switch value {
	case 0:
		return directionClient, true
	case 1:
		return directionServer, true
	default:
		return "", false
	}
}

func TestDecodePairAgreesOnValidPayload(t *testing.T) {
	got := comparePayload(directionServer, 2, mustDecodeHex(t, "00000000"))
	if !got.generated.ok || !got.oracle.ok {
		t.Fatalf("valid payload did not decode: generated=%v oracle=%v", got.generated.err, got.oracle.err)
	}
	if !got.generated.full || !got.oracle.full {
		t.Fatalf("valid payload was not fully consumed: generated=%v oracle=%v", got.generated.full, got.oracle.full)
	}
	if !bytes.Equal(got.generated.encoded, got.oracle.encoded) {
		t.Fatalf("re-encoded payloads differ: generated=%x oracle=%x", got.generated.encoded, got.oracle.encoded)
	}
}

func TestDecodePairAgreesOnTrailingPayload(t *testing.T) {
	got := comparePayload(directionServer, 2, mustDecodeHex(t, "0000000000"))
	if got.generated.ok != got.oracle.ok {
		t.Fatalf("decode success disagreement: generated=%v oracle=%v", got.generated.err, got.oracle.err)
	}
	if got.generated.full != got.oracle.full {
		t.Fatalf("full-consumption disagreement: generated=%v oracle=%v", got.generated.full, got.oracle.full)
	}
}

func TestOracleLeniencyIncludesPublishedSchemaConstraints(t *testing.T) {
	for _, message := range []string{
		"string length outside schema limits",
		"value is below schema minimum",
		"value exceeds schema maximum",
		"string does not match schema pattern",
	} {
		if !oracleLeniency(fmt.Errorf("invalid value: %s", message)) {
			t.Errorf("oracleLeniency(%q) = false, want true", message)
		}
	}
	if oracleLeniency(errors.New("invalid UTF-8")) {
		t.Fatal("oracleLeniency accepted a non-schema decode failure")
	}
}

func mustDecodeHex(t testing.TB, value string) []byte {
	t.Helper()
	data, err := hex.DecodeString(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
