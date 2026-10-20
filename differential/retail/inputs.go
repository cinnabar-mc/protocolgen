package retail

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Input cases. Every decoder runs every input.
const (
	CaseRetail      = "retail"       // the corpus batches unchanged
	CaseCrafting    = "crafting"     // gophertunnel-encoded recipe and descriptor variants
	CaseUnknownID   = "unknown-id"   // retail entries under an id no codec defines
	CaseWrongSender = "wrong-sender" // one-direction retail batches from the other peer
	CaseTruncated   = "truncated"    // retail entries missing their last byte
	CaseTrailing    = "trailing"     // retail entries with one extra byte
	CaseLarge       = "large"        // a generated-Go-encoded packet with a large collection
	CaseOversized   = "oversized"    // entries at and past decoder ceilings, and lengths past the input
)

// Probe cases, which runners derive from one input case.
const (
	ProbeElementLimit = "element-limit" // retail entries with a collection limit of zero
	ProbeAllocFailure = "alloc-failure" // large entries with allocations above a ceiling refused
	ProbeZeroCopy     = "zero-copy"     // unknown-id entries checked for shared buffers
)

// Input is one batch every decoder runs.
type Input struct {
	Fixture string
	Case    string
	Sender  string
	Batch   []byte
}

// Expected is one entry as framed by the harness.
type Expected struct {
	Fixture string
	Case    string
	Entry   int
	Header  string
	Payload string
}

const (
	inputsFile   = "inputs.tsv"
	expectedFile = "expected.tsv"
	verdictDir   = "verdicts"
)

// WriteInputs writes every batch under out/inputs, the inputs.tsv index the
// runners read, and expected.tsv for the comparator.
func WriteInputs(out string, inputs []Input) error {
	var index, expected strings.Builder
	seen := map[string]bool{}
	for _, input := range inputs {
		key := input.Case + "/" + input.Fixture
		if seen[key] {
			return fmt.Errorf("input %s is repeated", key)
		}
		seen[key] = true
		rel := filepath.Join("inputs", input.Case, input.Fixture)
		if err := os.MkdirAll(filepath.Join(out, "inputs", input.Case), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(out, rel), input.Batch, 0o644); err != nil {
			return err
		}
		sum := sha256.Sum256(input.Batch)
		fmt.Fprintf(&index, "%s\t%s\t%s\t%s\t%s\n", input.Fixture, input.Case, input.Sender, hex.EncodeToString(sum[:]), filepath.ToSlash(rel))
		entries, err := SplitBatch(input.Batch)
		if err != nil {
			return fmt.Errorf("input %s: %w", key, err)
		}
		for i, entry := range entries {
			fmt.Fprintf(&expected, "%s\t%s\t%d\t%s\t%s\n", input.Fixture, input.Case, i, HeaderText(entry.Header), Digest(entry.Payload))
		}
	}
	if err := os.WriteFile(filepath.Join(out, inputsFile), []byte(index.String()), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(out, expectedFile), []byte(expected.String()), 0o644)
}

// ReadInputs reads the inputs WriteInputs wrote, checking each digest.
func ReadInputs(out string) ([]Input, error) {
	var inputs []Input
	err := readRows(filepath.Join(out, inputsFile), 5, func(cols []string) error {
		data, err := os.ReadFile(filepath.Join(out, filepath.FromSlash(cols[4])))
		if err != nil {
			return err
		}
		if sum := sha256.Sum256(data); hex.EncodeToString(sum[:]) != cols[3] {
			return fmt.Errorf("%s does not match its digest", cols[4])
		}
		inputs = append(inputs, Input{Fixture: cols[0], Case: cols[1], Sender: cols[2], Batch: data})
		return nil
	})
	return inputs, err
}

// ReadExpected reads expected.tsv.
func ReadExpected(out string) ([]Expected, error) {
	var expected []Expected
	err := readRows(filepath.Join(out, expectedFile), 5, func(cols []string) error {
		entry, err := strconv.Atoi(cols[2])
		if err != nil {
			return err
		}
		expected = append(expected, Expected{Fixture: cols[0], Case: cols[1], Entry: entry, Header: cols[3], Payload: cols[4]})
		return nil
	})
	return expected, err
}

// WriteVerdicts writes one decoder's verdict file.
func WriteVerdicts(out, decoder string, verdicts []Verdict) error {
	if err := os.MkdirAll(filepath.Join(out, verdictDir), 0o755); err != nil {
		return err
	}
	var b strings.Builder
	for _, verdict := range verdicts {
		b.WriteString(verdict.Line())
		b.WriteByte('\n')
	}
	return os.WriteFile(filepath.Join(out, verdictDir, decoder+".tsv"), []byte(b.String()), 0o644)
}

// ReadVerdicts reads one decoder's verdict file.
func ReadVerdicts(out, decoder string) ([]Verdict, error) {
	path := filepath.Join(out, verdictDir, decoder+".tsv")
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var verdicts []Verdict
	scanner := bufio.NewScanner(file)
	scanner.Buffer(nil, 1<<20)
	for line := 1; scanner.Scan(); line++ {
		verdict, err := ParseVerdict(scanner.Text())
		if err != nil {
			return nil, fmt.Errorf("%s:%d: %w", path, line, err)
		}
		verdicts = append(verdicts, verdict)
	}
	return verdicts, scanner.Err()
}

func readRows(path string, cols int, row func([]string) error) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for line := 1; scanner.Scan(); line++ {
		fields := strings.Split(scanner.Text(), "\t")
		if len(fields) != cols {
			return fmt.Errorf("%s:%d: %d columns, want %d", path, line, len(fields), cols)
		}
		if err := row(fields); err != nil {
			return fmt.Errorf("%s:%d: %w", path, line, err)
		}
	}
	return scanner.Err()
}
