package mojangsymbols

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"protocolgen/internal/manifest"
	"protocolgen/internal/sourcelock"
)

const catalogFixture = `<!DOCTYPE html><html><body><table>
<tr><td>Unrelated</td><td><table><tr><td>ActorFlags</td><td>999</td></tr></table></td></tr>
<tr><td>ActorDataIDs</td><td><table><colgroup><col width="40%"><col width="10%"></colgroup>
<tr><td>FIRST</td><td>4</td><td>not row zero</td></tr>
<tr><td>RESERVED_038</td><td>38</td><td></td></tr>
<tr><td>ALIAS</td><td>38</td><td></td></tr>
<tr><td>Count</td><td>39</td><td></td></tr>
</table></td></tr>
<tr><td>ActorFlags</td><td><table>
<tr><td>LINGERING</td><td>47</td><td></td></tr>
<tr><td>HAS_GRAVITY</td><td>49</td><td></td></tr>
<tr><td>Count</td><td>50</td><td></td></tr>
</table></td></tr></table></body></html>`

func TestCatalogPreservesExplicitValuesReservedNamesAndAliases(t *testing.T) {
	pin := manifest.SourcePin{ID: "official-reference"}
	symbols, err := extractMetadataSymbols([]byte(catalogFixture), pin)
	if err != nil {
		t.Fatal(err)
	}
	want := []manifest.EnumValue{{Name: "FIRST", Value: 4}, {Name: "RESERVED_038", Value: 38}, {Name: "ALIAS", Value: 38}}
	if !reflect.DeepEqual(symbols[0].Values, want) {
		t.Fatalf("field IDs must preserve source names, gaps and aliases: %+v", symbols[0].Values)
	}
	wantFlags := []manifest.EnumValue{{Name: "LINGERING", Value: 47}, {Name: "HAS_GRAVITY", Value: 49}}
	if !reflect.DeepEqual(symbols[1].Values, wantFlags) {
		t.Fatalf("flags must be indices, not masks or row positions: %+v", symbols[1].Values)
	}
}

func TestCatalogRejectsIncompleteOrNonnumericTables(t *testing.T) {
	for name, data := range map[string]string{
		"missing table":    strings.ReplaceAll(catalogFixture, "ActorDataIDs", "OtherIDs"),
		"invalid number":   strings.Replace(catalogFixture, "<td>38</td>", "<td>unknown</td>", 1),
		"negative":         strings.Replace(catalogFixture, "<td>38</td>", "<td>-1</td>", 1),
		"overflow":         strings.Replace(catalogFixture, "<td>38</td>", "<td>4294967296</td>", 1),
		"missing sentinel": strings.ReplaceAll(catalogFixture, "<td>Count</td>", "<td>Other</td>"),
		"wrong sentinel":   strings.Replace(catalogFixture, "<td>39</td>", "<td>40</td>", 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := extractMetadataSymbols([]byte(data), manifest.SourcePin{ID: "fixture"}); err == nil {
				t.Fatal("accepted an incomplete or invalid source catalog")
			}
		})
	}
}

func TestCatalogRejectsSourceDriftWithoutMutatingManifest(t *testing.T) {
	for _, mode := range []string{"missing source", "changed bytes", "wrong output target", "conflicting symbols"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			wireTarget := manifest.Target{MinecraftVersion: "1.26.51", ProtocolVersion: 2193}
			pin := manifest.SourcePin{
				ID: "official-reference", Kind: "mojang-enum-catalog", Revision: catalogRevision,
				Locator: catalogLocator, Digest: fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(catalogFixture))),
				MinecraftVersion: "1.26.40-beta.0", ProtocolVersion: 2168,
			}
			lock := sourcelock.Lock{
				SchemaVersion: sourcelock.SchemaVersion,
				Target:        manifest.Target{MinecraftVersion: pin.MinecraftVersion, ProtocolVersion: pin.ProtocolVersion},
				AppliesTo:     &wireTarget, Catalog: "catalog.html", Sources: []manifest.SourcePin{pin},
			}
			if mode != "missing source" {
				data := []byte(catalogFixture)
				if mode == "changed bytes" {
					data = append(data, '\n')
				}
				if err := os.WriteFile(filepath.Join(root, lock.Catalog), data, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			lockPath := filepath.Join(root, "symbols-lock.json")
			if err := sourcelock.Write(lockPath, lock); err != nil {
				t.Fatal(err)
			}
			m := manifest.Manifest{
				SchemaVersion: manifest.SchemaVersion, Target: wireTarget,
				Sources: []manifest.SourcePin{{ID: "wire", Kind: "synthetic", Revision: "fixture", Digest: "fixture:wire"}},
				Packets: []manifest.Packet{{ID: 1, Name: "Fixture", Direction: manifest.DirectionClientbound}},
			}
			if mode == "wrong output target" {
				m.Target.ProtocolVersion++
			}
			if mode == "conflicting symbols" {
				symbols, err := extractMetadataSymbols([]byte(catalogFixture), pin)
				if err != nil {
					t.Fatal(err)
				}
				symbols[0].Values[0].Value++
				m.Symbols = symbols
			}
			before, err := manifest.MarshalStable(m)
			if err != nil {
				t.Fatal(err)
			}
			if err := AddMetadataSymbols(&m, lockPath); err == nil {
				t.Fatal("accepted source drift or conflicting reference data")
			}
			after, err := manifest.MarshalStable(m)
			if err != nil || !bytes.Equal(before, after) {
				t.Fatalf("failed import partially mutated caller's manifest: %v", err)
			}
		})
	}
}
