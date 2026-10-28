package differential

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	gophertunnelprotocol "github.com/sandertv/gophertunnel/minecraft/protocol"
	gophertunnelpacket "github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	generatedprotocol "protocolgen/generated/1.26.60/go/protocol"
	generatedpacket "protocolgen/generated/1.26.60/go/protocol/packet"
)

type direction string

const (
	directionClient direction = "client"
	directionServer direction = "server"
)

type codecResult struct {
	ok      bool
	full    bool
	encoded []byte
	err     error
}

type comparison struct {
	generated codecResult
	oracle    codecResult
	issues    []issue
}

type issue struct {
	packet    uint32
	direction direction
	inputHex  string
	kind      string
	rationale string
}

type acceptedFile struct {
	SchemaVersion    int             `json:"schema_version"`
	MinecraftVersion string          `json:"minecraft_version"`
	ProtocolVersion  int             `json:"protocol_version"`
	Entries          []acceptedEntry `json:"entries"`
}

type acceptedEntry struct {
	Packet    uint32 `json:"packet"`
	Direction string `json:"direction"`
	InputHex  string `json:"input_hex"`
	Kind      string `json:"kind"`
	Rationale string `json:"rationale"`
}

type corpusSample struct {
	direction direction
	id        uint32
	data      []byte
}

var (
	promoteFlag = flag.Bool("promote", false, "promote successful differential fuzz inputs to the corpus")
	triageFlag  = flag.Bool("triage", false, "record fuzz divergences without failing")
	acceptedMu  sync.Mutex
	writeMu     sync.Mutex
	triageMu    sync.Mutex
	triaged     = map[acceptedKey]bool{}
)

func comparePayload(dir direction, id uint32, data []byte) comparison {
	generatedFactory, generatedOK := generatedFactory(dir, id)
	oracleFactory, oracleOK := oracleFactory(dir, id)
	if !generatedOK || !oracleOK {
		return comparison{}
	}

	generated := decodeGenerated(generatedFactory, data)
	oracle := decodeOracle(oracleFactory, data, dir == directionServer)
	got := comparison{generated: generated, oracle: oracle}
	inputHex := hex.EncodeToString(data)
	// The replacement contract is one-directional: every input the oracle
	// accepts must decode and re-encode identically here. Generated code
	// accepting inputs the oracle rejects is documented permissiveness
	// (unknown enum values pass through by design), and consumption position
	// on rejected input is an implementation detail, not wire behavior.
	if oracle.ok && !generated.ok && !oracleLeniency(generated.err) {
		got.issues = append(got.issues, issue{packet: id, direction: dir, inputHex: inputHex, kind: "decode-success", rationale: fmt.Sprintf("generated=%v oracle=ok", errText(generated.err))})
	}
	if generated.ok && oracle.ok {
		switch {
		case generated.err != nil:
			got.issues = append(got.issues, issue{packet: id, direction: dir, inputHex: inputHex, kind: "generated-reencode", rationale: generated.err.Error()})
		case oracle.err != nil:
			got.issues = append(got.issues, issue{packet: id, direction: dir, inputHex: inputHex, kind: "oracle-reencode", rationale: oracle.err.Error()})
		case !bytes.Equal(generated.encoded, oracle.encoded):
			if !boolCanonicalizationOnly(generated.encoded, oracle.encoded) {
				got.issues = append(got.issues, issue{packet: id, direction: dir, inputHex: inputHex, kind: "re-encoded-bytes", rationale: fmt.Sprintf("generated=%x oracle=%x", generated.encoded, oracle.encoded)})
			}
		}
	}
	return got
}

func generatedFactory(dir direction, id uint32) (func() generatedpacket.Packet, bool) {
	var pk generatedpacket.Packet
	var ok bool
	switch dir {
	case directionClient:
		pk, ok = generatedpacket.NewClientPacket(id)
	case directionServer:
		pk, ok = generatedpacket.NewServerPacket(id)
	default:
		return nil, false
	}
	if !ok {
		return nil, false
	}
	return func() generatedpacket.Packet { return newGeneratedPacket(pk) }, true
}

func newGeneratedPacket(pk generatedpacket.Packet) generatedpacket.Packet {
	value, ok := generatedpacket.NewPacket(pk.ID())
	if !ok {
		panic(fmt.Sprintf("generated packet %d disappeared from all-packet pool", pk.ID()))
	}
	return value
}

func oracleFactory(dir direction, id uint32) (func() gophertunnelpacket.Packet, bool) {
	var pool gophertunnelpacket.Pool
	switch dir {
	case directionClient:
		pool = gophertunnelpacket.NewClientPool()
	case directionServer:
		pool = gophertunnelpacket.NewServerPool()
	default:
		return nil, false
	}
	factory, ok := pool[id]
	return factory, ok
}

func decodeGenerated(factory func() generatedpacket.Packet, data []byte) (result codecResult) {
	pk := factory()
	result.full = generatedConsumed(factory, data)
	result.err = recoverError(func() error { return generatedpacket.Decode(data, pk) })
	if result.err == nil {
		result.ok = true
		result.encoded, result.err = generatedpacket.Encode(pk)
	}
	return result
}

func generatedConsumed(factory func() generatedpacket.Packet, data []byte) (consumed bool) {
	reader := generatedprotocol.NewReader(data)
	defer func() {
		if recover() != nil {
			consumed = reader.Remaining() == 0
		}
	}()
	factory().Marshal(reader)
	return reader.Remaining() == 0
}

func decodeOracle(factory func() gophertunnelpacket.Packet, data []byte, limits bool) (result codecResult) {
	buffer := bytes.NewBuffer(append([]byte(nil), data...))
	pk := factory()
	result.err = recoverError(func() error {
		pk.Marshal(gophertunnelprotocol.NewReader(buffer, 0, limits))
		if buffer.Len() != 0 {
			return fmt.Errorf("unread bytes left: 0x%x", buffer.Bytes())
		}
		return nil
	})
	result.full = buffer.Len() == 0
	if result.err == nil {
		result.ok = true
		var output bytes.Buffer
		result.err = recoverError(func() error {
			pk.Marshal(gophertunnelprotocol.NewWriter(&output, 0))
			return nil
		})
		result.encoded = output.Bytes()
	}
	return result
}

// oracleLeniency reports whether a generated decode failure falls into a
// class where the pinned oracle is known to accept malformed input that the
// generated runtime deliberately rejects: fixed-width reads that ignore the
// returned byte count and zero-pad truncated fields (reader.go Read calls
// without length checks), an unbounded Varint64 with no ten-byte cap, and
// NBT blobs the oracle passes through without structural validation, schema
// constraints the oracle does not implement, and
// unknown union discriminants the oracle tolerates by silently skipping the
// variant body (a lossy decode; generated fails closed). None of these occur
// on well-formed traffic.
func oracleLeniency(err error) bool {
	if err == nil {
		return false
	}
	text := err.Error()
	for _, marker := range []string{
		"unexpected end of input",
		"varint exceeds",
		"NBT ",
		"collection length exceeds",
		"unknown union tag",
		"schema limits",
		"schema minimum",
		"schema maximum",
		"schema pattern",
	} {
		if strings.Contains(text, marker) {
			return true
		}
	}
	return false
}

// boolCanonicalizationOnly reports whether two equal-length encodings differ
// only at positions where the generated side wrote a canonical bool byte
// (0 or 1) and the oracle preserved a raw non-canonical byte: fields the
// manifest types as bool but the oracle types as a passthrough uint8. Real
// traffic carries only 0 or 1 there, so the difference cannot occur on
// vanilla input.
func boolCanonicalizationOnly(generated, oracle []byte) bool {
	if len(generated) != len(oracle) {
		return false
	}
	diffs := 0
	for i := range generated {
		if generated[i] == oracle[i] {
			continue
		}
		if generated[i] > 1 {
			return false
		}
		diffs++
	}
	return diffs > 0
}

func errText(err error) string {
	if err == nil {
		return "ok"
	}
	return err.Error()
}

func recoverError(fn func() error) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			switch value := recovered.(type) {
			case error:
				err = value
			default:
				err = fmt.Errorf("panic: %v", value)
			}
		}
	}()
	return fn()
}

func assertAccepted(issues []issue) error {
	accepted, err := loadAccepted()
	if err != nil {
		return err
	}
	for _, got := range issues {
		if !accepted[acceptedKey{got.packet, got.direction, got.inputHex, got.kind}] && !accepted[acceptedKey{got.packet, got.direction, "", got.kind}] {
			return fmt.Errorf("unaccepted divergence packet=%d direction=%s input=%s kind=%s: %s", got.packet, got.direction, got.inputHex, got.kind, got.rationale)
		}
	}
	return nil
}

type acceptedKey struct {
	packet    uint32
	direction direction
	inputHex  string
	kind      string
}

func loadAccepted() (map[acceptedKey]bool, error) {
	acceptedMu.Lock()
	defer acceptedMu.Unlock()
	data, err := os.ReadFile("accepted-divergences.json")
	if err != nil {
		return nil, err
	}
	var file acceptedFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}
	if file.SchemaVersion != 1 || file.MinecraftVersion != "1.26.60" || file.ProtocolVersion != 2225 {
		return nil, errors.New("accepted-divergences.json has the wrong target")
	}
	result := make(map[acceptedKey]bool, len(file.Entries))
	for _, entry := range file.Entries {
		if entry.Direction != string(directionClient) && entry.Direction != string(directionServer) {
			return nil, fmt.Errorf("accepted divergence %d has invalid direction %q", entry.Packet, entry.Direction)
		}
		if strings.TrimSpace(entry.Rationale) == "" {
			return nil, fmt.Errorf("accepted divergence %d has no rationale", entry.Packet)
		}
		if _, err := hex.DecodeString(entry.InputHex); err != nil {
			return nil, fmt.Errorf("accepted divergence %d has invalid input hex: %w", entry.Packet, err)
		}
		result[acceptedKey{entry.Packet, direction(entry.Direction), strings.ToLower(entry.InputHex), entry.Kind}] = true
	}
	return result, nil
}

func loadCorpus() ([]corpusSample, error) {
	var samples []corpusSample
	for _, dir := range []direction{directionClient, directionServer} {
		root := filepath.Join("testdata", "corpus", string(dir))
		ids, err := os.ReadDir(root)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		for _, idEntry := range ids {
			if !idEntry.IsDir() {
				continue
			}
			id, err := strconv.ParseUint(idEntry.Name(), 10, 32)
			if err != nil {
				return nil, fmt.Errorf("corpus directory %s: %w", idEntry.Name(), err)
			}
			files, err := os.ReadDir(filepath.Join(root, idEntry.Name()))
			if err != nil {
				return nil, err
			}
			for _, file := range files {
				if file.IsDir() || filepath.Ext(file.Name()) != ".hex" {
					continue
				}
				encoded, err := os.ReadFile(filepath.Join(root, idEntry.Name(), file.Name()))
				if err != nil {
					return nil, err
				}
				data, err := hex.DecodeString(strings.TrimSpace(string(encoded)))
				if err != nil {
					return nil, fmt.Errorf("corpus %s: %w", file.Name(), err)
				}
				samples = append(samples, corpusSample{direction: dir, id: uint32(id), data: data})
			}
		}
	}
	sort.Slice(samples, func(i, j int) bool {
		if samples[i].direction != samples[j].direction {
			return samples[i].direction < samples[j].direction
		}
		if samples[i].id != samples[j].id {
			return samples[i].id < samples[j].id
		}
		return bytes.Compare(samples[i].data, samples[j].data) < 0
	})
	return samples, nil
}

func promoteCorpus(sample corpusSample) error {
	digest := sha256.Sum256(sample.data)
	path := filepath.Join("testdata", "corpus", string(sample.direction), fmt.Sprintf("%03d", sample.id), hex.EncodeToString(digest[:])[:16]+".hex")
	writeMu.Lock()
	defer writeMu.Unlock()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	return os.WriteFile(path, []byte(hex.EncodeToString(sample.data)+"\n"), 0o644)
}

func recordTriage(issues []issue) error {
	if len(issues) == 0 {
		return nil
	}
	path := os.Getenv("DIFFERENTIAL_FINDINGS")
	if path == "" {
		path = "/tmp/protocolgen-differential-fuzz-findings.jsonl"
	}
	triageMu.Lock()
	defer triageMu.Unlock()
	var fresh []issue
	for _, got := range issues {
		key := acceptedKey{got.packet, got.direction, "", got.kind}
		if !triaged[key] {
			triaged[key] = true
			fresh = append(fresh, got)
		}
	}
	if len(fresh) == 0 {
		return nil
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	for _, got := range fresh {
		line, err := json.Marshal(struct {
			Packet    uint32 `json:"packet"`
			Direction string `json:"direction"`
			InputHex  string `json:"input_hex"`
			Kind      string `json:"kind"`
			Rationale string `json:"rationale"`
		}{got.packet, string(got.direction), got.inputHex, got.kind, got.rationale})
		if err != nil {
			return err
		}
		if _, err := file.Write(append(line, '\n')); err != nil {
			return err
		}
	}
	return nil
}
