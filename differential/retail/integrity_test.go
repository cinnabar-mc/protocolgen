package retail

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	genpacket "protocolgen/generated/1.26.51/go/protocol/packet"
)

func wantFailure(t *testing.T, stage string, err error, mention string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), mention) {
		t.Errorf("%s: error %v, want one mentioning %q", stage, err, mention)
	}
}

func TestCheckKeysRequiresExactlyOnceEach(t *testing.T) {
	if err := CheckKeys("s", []string{"a", "b"}, []string{"b", "a"}); err != nil {
		t.Fatal(err)
	}
	wantFailure(t, "missing", CheckKeys("s", []string{"a", "b"}, []string{"a"}), "missing: b")
	wantFailure(t, "repeated", CheckKeys("s", []string{"a"}, []string{"a", "a"}), "repeated 2 times: a")
	wantFailure(t, "unexpected", CheckKeys("s", []string{"a"}, []string{"a", "c"}), "unexpected: c")
	wantFailure(t, "expected twice", CheckKeys("s", []string{"a", "a"}, []string{"a"}), "expected twice: a")
}

func dropFirst[T any](rows []T) []T { return append([]T(nil), rows[1:]...) }

func repeatFirst[T any](rows []T) []T { return append(append([]T(nil), rows...), rows[0]) }

// Inputs and expected entries: compare rebuilds both, so a dropped, repeated
// or altered input fails.
func TestInputAndExpectedStagesRejectMissingAndRepeatedRows(t *testing.T) {
	inputs, err := CraftingInputs()
	if err != nil {
		t.Fatal(err)
	}
	want := InputKeys(inputs)
	wantFailure(t, "inputs dropped", CheckKeys("inputs", want, InputKeys(dropFirst(inputs))), "missing")
	wantFailure(t, "inputs repeated", CheckKeys("inputs", want, InputKeys(repeatFirst(inputs))), "repeated")
	altered := append([]Input(nil), inputs...)
	altered[0].Batch = append(append([]byte(nil), altered[0].Batch...), 0)
	wantFailure(t, "inputs altered", CheckKeys("inputs", want, InputKeys(altered)), "missing")
	expected, err := ExpectedEntries(inputs)
	if err != nil {
		t.Fatal(err)
	}
	wantKeys := ExpectedKeys(expected)
	wantFailure(t, "expected dropped", CheckKeys("expected", wantKeys, ExpectedKeys(dropFirst(expected))), "missing")
	wantFailure(t, "expected repeated", CheckKeys("expected", wantKeys, ExpectedKeys(repeatFirst(expected))), "repeated")
}

// Verdicts, including the oversized decisions, must cover every framed entry
// once per decoder.
func TestVerdictStageRejectsRepeatedAndMissingOversizedRows(t *testing.T) {
	expected := append(testExpected(), Expected{"o.bin", CaseOversized, 0, "9:0:0", "o"})
	verdicts := agreeing(expected)
	repeated := verdicts[DecoderGo][0]
	verdicts[DecoderGo] = append(verdicts[DecoderGo], repeated)
	remove(verdicts, DecoderGophertunnel, CaseOversized, "o.bin", 0)
	found := kindsOf(Compare(expected, verdicts))
	for _, want := range []string{"go " + repeated.Case + " " + repeated.Fixture + " extra", "gophertunnel oversized o.bin missing"} {
		if !strings.Contains(strings.Join(found, "\n"), want) {
			t.Errorf("disagreements %q lack %q", found, want)
		}
	}
}

// Field rows must be exactly the field-case entries Rust decoded.
func TestFieldRowStageRejectsMissingAndRepeatedRows(t *testing.T) {
	body, err := genpacket.Encode(generatedMovePlayer())
	if err != nil {
		t.Fatal(err)
	}
	in := Input{Fixture: "m.bin", Case: CaseRetail, Sender: SenderServer, Batch: JoinBatch([]Entry{NewEntry(genpacket.IDMovePlayer, body)})}
	rust := []Verdict{{Fixture: "m.bin", Case: CaseRetail, Entry: 0, Outcome: OutcomeOK}, {Fixture: "m.bin", Case: ProbeElementLimit, Entry: 0, Outcome: OutcomeOK}}
	row := "m.bin\tretail\t0\t" + movePlayerDebug + "\n"
	for name, tc := range map[string]struct{ rows, mention string }{
		"agrees":   {row, ""},
		"missing":  {"", "missing: m.bin/retail/0"},
		"repeated": {row + row, "repeated 2 times"},
		"stray":    {row + strings.Replace(row, "\t0\t", "\t1\t", 1), "unexpected: m.bin/retail/1"},
	} {
		out := t.TempDir()
		if err := os.MkdirAll(filepath.Join(out, "fields"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(out, "fields", "rust.tsv"), []byte(tc.rows), 0o644); err != nil {
			t.Fatal(err)
		}
		checks, err := CompareFields(out, []Input{in}, rust, testSchema(t))
		if tc.mention == "" {
			if err != nil || len(checks) != 1 || checks[0].Diff != "" {
				t.Errorf("%s: %v %+v", name, err, checks)
			}
			continue
		}
		wantFailure(t, name, err, tc.mention)
	}
}

// The Rust runner renders exactly the cases fieldCases names.
func TestRunnerRendersTheFieldCases(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("runner", "src", "verdict.rs"))
	if err != nil {
		t.Fatal(err)
	}
	want := `const FIELD_CASES: [&str; 3] = ["` + strings.Join(fieldCases, `", "`) + `"];`
	if !strings.Contains(string(source), want) {
		t.Fatalf("runner/src/verdict.rs does not declare %s", want)
	}
}

func copyDir(t *testing.T, from string) string {
	t.Helper()
	dir := t.TempDir()
	entries, err := os.ReadDir(from)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		data, err := os.ReadFile(filepath.Join(from, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, e.Name()), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestSyntheticStageRejectsMissingRepeatedAndStrayFiles(t *testing.T) {
	files, err := ReadSyntheticCorpus(SyntheticDir)
	if err != nil {
		t.Fatal(err)
	}
	dir := copyDir(t, SyntheticDir)
	// The files on disk are the independent key set the manifest must match.
	wantFailure(t, "manifest entry dropped", CheckSyntheticDir(dir, dropFirst(files)), "missing: "+files[0].File)
	wantFailure(t, "manifest entry repeated", CheckSyntheticDir(dir, repeatFirst(files)), "repeated")
	if err := os.Remove(filepath.Join(dir, files[0].File)); err != nil {
		t.Fatal(err)
	}
	wantFailure(t, "file deleted", CheckSyntheticDir(dir, files), "unexpected: "+files[0].File)
}

func TestEncodeStageRejectsMissingRepeatedAndStrayCases(t *testing.T) {
	cases := EncodeCases()
	if err := CheckEncodeDir(EncodeDir, cases); err != nil {
		t.Fatal(err)
	}
	wantFailure(t, "case dropped", CheckEncodeDir(EncodeDir, dropFirst(cases)), "unexpected")
	wantFailure(t, "case repeated", CheckEncodeDir(EncodeDir, repeatFirst(cases)), "expected twice")
	dir := copyDir(t, EncodeDir)
	if err := os.Remove(filepath.Join(dir, cases[0].Name+".bin")); err != nil {
		t.Fatal(err)
	}
	wantFailure(t, "file deleted", CheckEncodeDir(dir, cases), "missing")
}

func TestBenchStageRejectsMissingAndRepeatedPaths(t *testing.T) {
	files, err := ReadSyntheticCorpus(SyntheticDir)
	if err != nil {
		t.Fatal(err)
	}
	want := BenchKeys(files)
	var lines []string
	for _, key := range want {
		for range 2 {
			lines = append(lines, "Benchmark"+key+"-4 \t 100\t 10.0 ns/op")
		}
	}
	write := func(lines []string) string {
		path := filepath.Join(t.TempDir(), "go.txt")
		if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		return path
	}
	got, err := GoBenchKeys(write(lines), 2)
	if err != nil || CheckKeys("go", want, got) != nil {
		t.Fatalf("complete output: %v %v", err, CheckKeys("go", want, got))
	}
	got, err = GoBenchKeys(write(lines[2:]), 2)
	if err != nil {
		t.Fatal(err)
	}
	wantFailure(t, "go path dropped", CheckKeys("go", want, got), "missing: "+want[0])
	_, err = GoBenchKeys(write(lines[1:]), 2)
	wantFailure(t, "go repeat dropped", err, "ran 1 times")
	_, err = GoBenchKeys(write(append(lines, lines[0])), 2)
	wantFailure(t, "go repeat added", err, "ran 3 times")

	criterion := t.TempDir()
	add := func(dir, id string) {
		path := filepath.Join(criterion, dir, "new")
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(path, "benchmark.json"), []byte(`{"full_id":"`+id+`"}`), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, key := range want[1:] {
		add(key, key)
	}
	add("copy", want[1])
	got, err = CriterionKeys(criterion)
	if err != nil {
		t.Fatal(err)
	}
	err = CheckKeys("criterion", want, got)
	wantFailure(t, "criterion path dropped", err, "missing: "+want[0])
	wantFailure(t, "criterion path repeated", err, "repeated 2 times: "+want[1])
}
