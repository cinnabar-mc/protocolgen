package changelog

import (
	"strings"
	"testing"
)

func TestGenerateDoesNotRenameSurvivingPackets(t *testing.T) {
	for _, tc := range []struct {
		name     string
		from, to map[string]int
	}{
		{"reuse", map[string]int{"A": 7}, map[string]int{"A": 8, "B": 7}},
		{"swap", map[string]int{"A": 7, "B": 8}, map[string]int{"A": 8, "B": 7}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			from, to := t.TempDir(), t.TempDir()
			for dir, packets := range map[string]map[string]int{from: tc.from, to: tc.to} {
				for name, id := range packets {
					writeSchema(t, dir, name, packetSchema("1.0", 1, id, "Payload"))
				}
			}
			got, err := Generate(Config{FromDir: from, ToDir: to})
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(got), "## Renames") || !strings.Contains(string(got), "packet id changed from `7` to `8`") {
				t.Fatalf("packet identity was lost: %s", got)
			}
			if tc.name == "reuse" && !strings.Contains(string(got), "**1** new") {
				t.Fatalf("missing new packet: %s", got)
			}
		})
	}
}
