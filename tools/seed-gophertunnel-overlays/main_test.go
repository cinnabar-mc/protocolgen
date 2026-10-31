package main

import (
	"testing"

	"protocolgen/internal/docs"
	"protocolgen/internal/manifest"
	"protocolgen/internal/naming"
)

func TestPortFieldDocsDoesNotReuseReorderedSource(t *testing.T) {
	field := manifest.Primitive("u32le")
	fields := []manifest.Field{{Name: "b", Encode: field}, {Name: "renamed_a", Encode: field}}
	source := sourceType{Fields: []sourceField{{Name: "OldA", Type: "uint32", Doc: "OldA describes A."}, {Name: "B", Type: "uint32", Doc: "B describes B."}}}
	overlay := docs.Overlay{Fields: map[string]string{}}
	ported, unmatched, skipped := portFieldDocs("Owner", fields, source, overlay)
	if ported != 1 || unmatched != 1 || skipped != 1 || overlay.Fields[docs.FieldKey("Owner", "renamed_a")] != "" {
		t.Fatalf("reused documentation: %+v; counts %d %d %d", overlay, ported, unmatched, skipped)
	}
}

func TestPortDocsUsesSharedGoCasing(t *testing.T) {
	m := manifest.Manifest{Packets: []manifest.Packet{{Name: "JsonRequestPacket", Fields: []manifest.Field{{Name: "Data", Encode: manifest.Node{Kind: manifest.KindStruct, TypeID: "UuidData", Fields: []manifest.Field{{Name: "Value", Encode: manifest.Node{Kind: manifest.KindPrimitive}}}}}}}}}
	index, err := buildIdentityIndex(m, naming.Overlay{})
	if err != nil {
		t.Fatal(err)
	}
	sources := map[string]sourceType{"UUIDData": {Name: "UUIDData", File: "identity", Doc: "UUIDData contains identity data."}}
	overlay, report := portDocs(m, index, sources, map[string]sourceType{"JSONRequest": {Name: "JSONRequest", Doc: "JSONRequest requests data."}})
	if report.typeDocs != 2 || overlay.Types["UuidData"] == "" || overlay.Types["JsonRequestPacket"] == "" {
		t.Fatalf("casing lost docs: %+v; %+v; %+v", index, overlay, report)
	}
	if domain, direct := domainForName(index.Names["UuidData"], sources); !direct || domain != "identity" {
		t.Fatalf("domain = %q, direct = %v", domain, direct)
	}
}

func TestSourceResolutionRejectsAmbiguousSuffixes(t *testing.T) {
	sources := map[string]sourceType{"Entry": {Name: "Entry", File: "one"}, "EntryInfoData": {Name: "EntryInfoData", File: "two"}}
	for range 100 {
		if _, ok := sourceTypeForName("EntryInfo", sources); ok {
			t.Fatal("ambiguous source resolved")
		}
		if _, direct := domainForName("EntryInfo", sources); direct {
			t.Fatal("ambiguous domain resolved")
		}
	}
	sources["EntryInfo"] = sourceType{Name: "EntryInfo", File: "exact"}
	if source, ok := sourceTypeForName("EntryInfo", sources); !ok || source.File != "exact" {
		t.Fatal("exact match lost")
	}
}

func TestUndocumentedFieldsStillDisableReorderedFallback(t *testing.T) {
	fields := []manifest.Field{{Name: "b", Encode: manifest.Primitive("u32le")}, {Name: "renamed_a", Encode: manifest.Primitive("u32le")}}
	source := sourceType{Fields: []sourceField{{Name: "OldA", Type: "uint32", Doc: "OldA describes A."}, {Name: "B", Type: "uint32"}}}
	overlay := docs.Overlay{Fields: map[string]string{}}
	portFieldDocs("Owner", fields, source, overlay)
	if len(overlay.Fields) != 0 {
		t.Fatalf("copied reordered documentation: %+v", overlay)
	}
}
