package retail

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// Every comparison stage checks the rows it reads against a key set derived
// independently of those rows, so a dropped or repeated row fails the run
// instead of shrinking it.

// CheckKeys fails unless got holds exactly the keys of want, each once.
func CheckKeys(stage string, want, got []string) error {
	wantN, gotN := map[string]int{}, map[string]int{}
	for _, key := range want {
		wantN[key]++
	}
	for _, key := range got {
		gotN[key]++
	}
	keys := unique(append(append([]string(nil), want...), got...))
	sort.Strings(keys)
	var problems []string
	for _, key := range keys {
		switch {
		case wantN[key] > 1:
			problems = append(problems, "expected twice: "+key)
		case gotN[key] == 0:
			problems = append(problems, "missing: "+key)
		case wantN[key] == 0:
			problems = append(problems, "unexpected: "+key)
		case gotN[key] > 1:
			problems = append(problems, fmt.Sprintf("repeated %d times: %s", gotN[key], key))
		}
	}
	if len(problems) == 0 {
		return nil
	}
	if len(problems) > 20 {
		problems = append(problems[:20], fmt.Sprintf("and %d more", len(problems)-20))
	}
	return fmt.Errorf("%s: %s", stage, strings.Join(problems, "; "))
}

// InputKeys returns the case/fixture/digest key of every input.
func InputKeys(inputs []Input) []string {
	keys := make([]string, 0, len(inputs))
	for _, in := range inputs {
		keys = append(keys, in.Case+"/"+in.Fixture+"@"+Digest(in.Batch))
	}
	return keys
}

// ExpectedKeys returns the fixture/case/entry key of every expected entry.
func ExpectedKeys(expected []Expected) []string {
	keys := make([]string, 0, len(expected))
	for _, e := range expected {
		keys = append(keys, fmt.Sprintf("%s/%s/%d", e.Fixture, e.Case, e.Entry))
	}
	return keys
}

// FieldRowKeys returns the keys of the Rust verdicts whose decoded fields the
// runner must have rendered: every known packet decoded in a field case.
func FieldRowKeys(rust []Verdict) []string {
	var keys []string
	for _, v := range rust {
		if v.Outcome == OutcomeOK && slices.Contains(fieldCases, v.Case) {
			keys = append(keys, fmt.Sprintf("%s/%s/%d", v.Fixture, v.Case, v.Entry))
		}
	}
	return keys
}

// fieldCases are the cases whose decoded fields the Rust runner renders; it
// names the same three in verdict.rs.
var fieldCases = []string{CaseRetail, CaseCrafting, CaseLarge}

// CheckEncodeDir fails unless dir holds exactly one <name>.bin per case.
func CheckEncodeDir(dir string, cases []EncodeCase) error {
	var want []string
	for _, c := range cases {
		want = append(want, c.Name+".bin")
	}
	got, err := filepath.Glob(filepath.Join(dir, "*.bin"))
	if err != nil {
		return err
	}
	for i := range got {
		got[i] = filepath.Base(got[i])
	}
	return CheckKeys("encode-from-struct cases", want, got)
}

// CheckSyntheticDir fails unless the manifest names each committed file once
// and dir holds no other batch.
func CheckSyntheticDir(dir string, files []SyntheticFile) error {
	var names, listed []string
	for _, f := range files {
		names = append(names, f.Name)
		listed = append(listed, f.File)
	}
	if err := CheckKeys("synthetic case names", unique(names), names); err != nil {
		return err
	}
	onDisk, err := filepath.Glob(filepath.Join(dir, "*.bin"))
	if err != nil {
		return err
	}
	for i := range onDisk {
		onDisk[i] = filepath.Base(onDisk[i])
	}
	return CheckKeys("synthetic files", onDisk, listed)
}

func unique(keys []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, key := range keys {
		if !seen[key] {
			seen[key] = true
			out = append(out, key)
		}
	}
	return out
}

// Bench groups, as both benches name them.
const (
	groupDecode      = "Decode"
	groupEncode      = "Encode"
	groupDecodeBatch = "DecodeBatch"
	groupEncodeBatch = "EncodeBatch"
)

// BenchKeys returns every Group/Case path the paired benches must report for
// the synthetic corpus.
func BenchKeys(files []SyntheticFile) []string {
	var keys []string
	for _, f := range files {
		groups := []string{groupDecode, groupEncode}
		if f.Set == SetBatch {
			groups = []string{groupDecodeBatch, groupEncodeBatch}
		}
		for _, g := range groups {
			keys = append(keys, g+"/"+f.Name)
		}
	}
	return keys
}

var goBenchLine = regexp.MustCompile(`^Benchmark(\S+?)(-\d+)?\s+\d+\s`)

// GoBenchKeys reads `go test -bench` output, returning each benchmark path
// once per repeat it ran.
func GoBenchKeys(path string, count int) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	runs := map[string]int{}
	var order []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if m := goBenchLine.FindStringSubmatch(scanner.Text()); m != nil {
			if runs[m[1]] == 0 {
				order = append(order, m[1])
			}
			runs[m[1]]++
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	var keys []string
	for _, name := range order {
		if runs[name] != count {
			return nil, fmt.Errorf("go benches: %s ran %d times, want %s", name, runs[name], strconv.Itoa(count))
		}
		keys = append(keys, name)
	}
	return keys, nil
}

// CriterionKeys returns the full id of every benchmark under a criterion
// output directory.
func CriterionKeys(dir string) ([]string, error) {
	var keys []string
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || d.Name() != "benchmark.json" || filepath.Base(filepath.Dir(path)) != "new" {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var bench struct {
			FullID string `json:"full_id"`
		}
		if err := json.Unmarshal(data, &bench); err != nil || bench.FullID == "" {
			return fmt.Errorf("%s: no full_id (%v)", path, err)
		}
		keys = append(keys, bench.FullID)
		return nil
	})
	return keys, err
}
