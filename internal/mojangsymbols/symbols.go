// Package mojangsymbols imports reference constants, not packet wire shapes.
package mojangsymbols

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"protocolgen/internal/manifest"
	"protocolgen/internal/sourcelock"
)

const catalogRevision = "0e00fe80f4f3c71572ff6429de40146d1f4412fc"
const catalogLocator = "https://github.com/Mojang/bedrock-protocol-docs/blob/" + catalogRevision + "/html/enums.html"

// AddMetadataSymbols reads the existing Mojang documentation submodule. Its
// historical catalog has a distinct versioned source pin: it is not promoted
// into the output target's wire sources or relabelled as current native data.
func AddMetadataSymbols(m *manifest.Manifest, lockPath string) error {
	lock, err := sourcelock.Load(lockPath)
	if err != nil {
		return err
	}
	if lock.AppliesTo == nil || *lock.AppliesTo != m.Target {
		return fmt.Errorf("metadata catalog lock does not explicitly apply to manifest target")
	}
	if len(lock.Sources) != 1 || lock.Catalog == "" || filepath.IsAbs(lock.Catalog) {
		return fmt.Errorf("metadata catalog requires one source and a relative catalog path")
	}
	pin := lock.Sources[0]
	if pin.Kind != "mojang-enum-catalog" || pin.Revision != catalogRevision || pin.Locator != catalogLocator {
		return fmt.Errorf("unsupported metadata catalog: requires pinned Mojang enum export %s", catalogRevision)
	}
	if pin.MinecraftVersion == "" || pin.ProtocolVersion <= 0 {
		return fmt.Errorf("metadata catalog reference source version is required")
	}
	data, err := os.ReadFile(filepath.Join(filepath.Dir(lockPath), lock.Catalog))
	if err != nil {
		return fmt.Errorf("read pinned Mojang enum catalog (initialize bedrock-protocol-docs submodule): %w", err)
	}
	digest := sha256.Sum256(data)
	if got := "sha256:" + hex.EncodeToString(digest[:]); got != pin.Digest {
		return fmt.Errorf("Mojang enum catalog digest mismatch: lock has %s, file has %s", pin.Digest, got)
	}
	symbols, err := extractMetadataSymbols(data, pin)
	if err != nil {
		return err
	}
	result := *m
	result.Symbols = append([]manifest.Symbol(nil), m.Symbols...)
	for _, symbol := range symbols {
		exists := false
		for _, existing := range m.Symbols {
			if existing.ID == symbol.ID || existing.Name == symbol.Name {
				if !reflect.DeepEqual(existing, symbol) {
					return fmt.Errorf("metadata symbol %q conflicts with pinned Mojang catalog", symbol.ID)
				}
				exists = true
			}
		}
		if !exists {
			result.Symbols = append(result.Symbols, symbol)
		}
	}
	if err := manifest.Validate(result); err != nil {
		return err
	}
	*m = result
	return nil
}

type catalogCell struct {
	Text string `xml:",chardata"`
}

type catalogTableCell struct {
	Table struct {
		Rows []struct {
			Cells []catalogCell `xml:"td"`
		} `xml:"tr"`
	} `xml:"table"`
}

// The pinned export uses explicit table cells, decimal integer values and HTML
// void elements. Decode that structure; never infer values from row positions,
// scrape with regular expressions, or replace Mojang's RESERVED names.
func extractMetadataSymbols(data []byte, pin manifest.SourcePin) ([]manifest.Symbol, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.Strict = false
	decoder.AutoClose = xml.HTMLAutoClose
	decoder.Entity = xml.HTMLEntity
	names := map[string]int{"ActorDataIDs": 0, "ActorFlags": 1}
	symbols := make([]manifest.Symbol, len(names))
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("parse Mojang enum catalog: %w", err)
		}
		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "td" {
			continue
		}
		var cell catalogCell
		if err := decoder.DecodeElement(&cell, &start); err != nil {
			return nil, fmt.Errorf("parse Mojang enum name: %w", err)
		}
		name := strings.TrimSpace(cell.Text)
		index, wanted := names[name]
		if !wanted {
			continue
		}
		if symbols[index].Name != "" {
			return nil, fmt.Errorf("duplicate Mojang enum table %s", name)
		}
		var table catalogTableCell
		if err := decoder.Decode(&table); err != nil {
			return nil, fmt.Errorf("parse Mojang enum %s table: %w", name, err)
		}
		symbol := manifest.Symbol{
			ID: "symbols/" + name, Name: name, Representation: "u32", Source: pin,
			Provenance: manifest.Provenance{Pins: []string{pin.ID}},
		}
		count, maximum := int64(-1), int64(-1)
		for _, row := range table.Table.Rows {
			if len(row.Cells) < 2 {
				return nil, fmt.Errorf("Mojang enum %s has a row without name and value", name)
			}
			key := strings.TrimSpace(row.Cells[0].Text)
			value, err := strconv.ParseUint(strings.TrimSpace(row.Cells[1].Text), 0, 32)
			if err != nil {
				return nil, fmt.Errorf("Mojang enum %s value %s is not u32: %w", name, key, err)
			}
			if key == "Count" {
				if count != -1 {
					return nil, fmt.Errorf("duplicate Mojang enum %s Count sentinel", name)
				}
				count = int64(value)
				continue
			}
			symbol.Values = append(symbol.Values, manifest.EnumValue{Name: key, Value: int64(value)})
			if int64(value) > maximum {
				maximum = int64(value)
			}
		}
		if maximum < 0 || count != maximum+1 {
			return nil, fmt.Errorf("Mojang enum %s has no values or inconsistent Count sentinel", name)
		}
		symbols[index] = symbol
	}
	for name, index := range names {
		if symbols[index].Name == "" {
			return nil, fmt.Errorf("Mojang enum catalog omits %s", name)
		}
	}
	return symbols, nil
}
