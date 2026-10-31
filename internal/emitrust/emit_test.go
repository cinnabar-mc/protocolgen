package emitrust

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"protocolgen/internal/docs"
	"protocolgen/internal/domains"
	"protocolgen/internal/licensing"
	"protocolgen/internal/manifest"
)

func generatedRustSource(m manifest.Manifest) (string, error) {
	files, err := GenerateFiles(m)
	if err != nil {
		return "", err
	}
	var source strings.Builder
	for name, contents := range files {
		if !strings.HasSuffix(name, ".rs") {
			continue
		}
		source.WriteString(contents)
	}
	return source.String(), nil
}

func TestGenerateRustConsumesCanonicalManifest(t *testing.T) {
	m := manifest.Manifest{
		SchemaVersion: 2,
		Target:        manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168},
		Sources:       []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:rust", MinecraftVersion: "fixture", ProtocolVersion: 2168}},
		Packets:       []manifest.Packet{{ID: 1, Name: "FixturePacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{{Ordinal: 0, Name: "Value", Encode: manifest.Optional(manifest.Primitive("u8")), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}},
	}
	source, err := generatedRustSource(m)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(source, "pub struct Fixture") || !strings.Contains(source, "pub const ID: u32 = 1;") || !strings.Contains(source, `pub const GAME_VERSION: &str = "fixture";`) || !strings.Contains(source, "pub const PROTOCOL_VERSION: i32 = 2168;") {
		t.Fatalf("generated Rust omitted packet definition or ID:\n%s", source)
	}
	if strings.Contains(source, "SHAPE") || strings.Contains(source, "WireEncoder") {
		t.Fatalf("generated Rust exposed runtime schema or placeholder codecs:\n%s", source)
	}
	if strings.Contains(source, "gophertunnel") || strings.Contains(source, "bedrock-protocol-docs") {
		t.Fatalf("generated Rust contains source/profile lookup text")
	}
}

func TestGenerateRustDistributesContributionNotices(t *testing.T) {
	for _, contract := range []int{0, CodecContractVersion} {
		files, err := GenerateFilesWithOptions(contractFixture(), Options{CodecContract: contract})
		if err != nil {
			t.Fatal(err)
		}
		for name, want := range map[string]string{"LICENSE-MIT": licensing.MIT, "NOTICE": licensing.Notice} {
			if files[name] != want {
				t.Errorf("contract %d: %s differs from the canonical notice", contract, name)
			}
			root, err := os.ReadFile(filepath.Join("..", "..", name))
			if err != nil {
				t.Fatal(err)
			}
			if string(root) != want {
				t.Errorf("root %s differs from internal/licensing/%s", name, name)
			}
		}
		if !strings.Contains(files["Cargo.toml"], `license = "`+licensing.SPDX+`"`) {
			t.Errorf("contract %d: Cargo metadata omits the authored implementation license", contract)
		}
	}
}

func TestGenerateRustEmitsSchemaConstraintChecks(t *testing.T) {
	text := manifest.String(manifest.Primitive("var_u32"))
	text.Constraints = &manifest.Constraints{MinLength: pointerTo(uint64(1)), MaxLength: pointerTo(uint64(65536)), Pattern: "^[a-z]+$"}
	count := manifest.Array(manifest.Primitive("var_u32"), manifest.Primitive("u16le"))
	count.Constraints = &manifest.Constraints{MaxItems: pointerTo(uint64(4))}
	raw := manifest.Array(manifest.Primitive("var_u32"), manifest.Primitive("u8"))
	raw.Constraints = &manifest.Constraints{MaxItems: pointerTo(uint64(4))}
	number := manifest.Primitive("zigzag_i32")
	number.Constraints = &manifest.Constraints{Minimum: pointerTo(json.Number("-1")), Maximum: pointerTo(json.Number("64"))}
	decimal := manifest.Primitive("f32le")
	decimal.Constraints = &manifest.Constraints{Minimum: pointerTo(json.Number("0")), Maximum: pointerTo(json.Number("1"))}
	largeInteger := manifest.Primitive("u32le")
	largeInteger.Constraints = &manifest.Constraints{Maximum: pointerTo(json.Number("4294967294"))}
	m := manifest.Manifest{SchemaVersion: 2, Target: manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 1}, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "1", Digest: "fixture", MinecraftVersion: "fixture", ProtocolVersion: 1}}, Packets: []manifest.Packet{{ID: 1, Name: "LimitsPacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{
		{Ordinal: 0, Name: "Text", Encode: text, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		{Ordinal: 1, Name: "Values", Encode: count, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		{Ordinal: 2, Name: "Number", Encode: number, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		{Ordinal: 3, Name: "Decimal", Encode: decimal, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		{Ordinal: 4, Name: "LargeInteger", Encode: largeInteger, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		{Ordinal: 5, Name: "MaybeText", Encode: manifest.Optional(text), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		{Ordinal: 6, Name: "Raw", Encode: raw, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
	}}}}
	files, err := GenerateFiles(m)
	if err != nil {
		t.Fatal(err)
	}
	packet := files["src/packets.rs"]
	for _, want := range []string{
		"wire::encode_string_limits(writer, &self.text, 1, 65536);",
		"wire::decode_string_limits(reader, 1, 65536)?",
		`wire::assert_pattern(&self.text, "^[a-z]+$");`,
		`wire::validate_pattern(&value, "^[a-z]+$")?;`,
		"wire::encode_collection_limits(writer, self.values.as_slice(), 0, 4);",
		"wire::decode_collection_limits::<wire::U16LE>(reader, 2, 0, 4)?",
		"pub raw: bytes::Bytes,",
		"wire::encode_bytes_limits(writer, self.raw.as_ref(), 0, 4);",
		"wire::decode_bytes_limits(reader, 0, 4)?",
		"wire::assert_number_limits(self.number.0, Some(-1), Some(64));",
		"wire::validate_number_limits(value.0, Some(-1), Some(64))?;",
		"wire::assert_number_limits(self.decimal.0, Some(0.0), Some(1.0));",
		"wire::assert_number_limits(self.large_integer.0, None, Some(4294967294));",
		"wire::encode_string_limits(writer, value, 1, 65536);",
		`wire::assert_pattern(value, "^[a-z]+$");`,
	} {
		if !strings.Contains(packet, want) {
			t.Fatalf("generated Rust packet omits %q:\n%s", want, packet)
		}
	}
	if roundtrip := files["tests/roundtrip.rs"]; !strings.Contains(roundtrip, "schema_constraint_panic") || !strings.Contains(roundtrip, "every_schema_valid_packet_default_round_trips") {
		t.Fatalf("generated Rust round-trip test does not distinguish invalid schema defaults:\n%s", roundtrip)
	}
}

func pointerTo[T any](value T) *T { return &value }

func TestGenerateRustOrdersDefinitionsByReviewedDomain(t *testing.T) {
	alpha := manifest.Node{Kind: manifest.KindStruct, TypeID: "Alpha", Fields: []manifest.Field{{Ordinal: 0, Name: "Value", Encode: manifest.Primitive("u8"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}
	beta := manifest.Node{Kind: manifest.KindStruct, TypeID: "Beta", Fields: []manifest.Field{{Ordinal: 0, Name: "Value", Encode: manifest.Primitive("u8"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}
	m := manifest.Manifest{
		SchemaVersion: 2,
		Target:        manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 1},
		Sources:       []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "1", Digest: "fixture", MinecraftVersion: "fixture", ProtocolVersion: 1}},
		Packets: []manifest.Packet{{ID: 1, Name: "FixturePacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{
			{Ordinal: 0, Name: "Beta", Encode: beta, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
			{Ordinal: 1, Name: "Alpha", Encode: alpha, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		}}},
	}
	files, err := GenerateFilesWithOptions(m, Options{Domains: domains.Overlay{Domains: map[string]string{"Alpha": "shared", "Beta": "shared"}}})
	if err != nil {
		t.Fatalf("GenerateFilesWithOptions: %v", err)
	}
	types := files["src/types.rs"]
	if !strings.Contains(types, "// Domain: shared") || strings.Index(types, "pub struct Alpha") > strings.Index(types, "pub struct Beta") {
		t.Fatalf("Rust definitions were not grouped and sorted:\n%s", types)
	}
}

func TestGenerateRustEmitsReviewedDocs(t *testing.T) {
	shared := manifest.Node{Kind: manifest.KindStruct, TypeID: "Shared", Fields: []manifest.Field{
		{Ordinal: 0, Name: "Value", Encode: manifest.Primitive("u8"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		{Ordinal: 1, Name: "Other", Encode: manifest.Primitive("u8"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
	}}
	m := manifest.Manifest{
		SchemaVersion: 2,
		Target:        manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 1},
		Sources:       []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "1", Digest: "fixture", MinecraftVersion: "fixture", ProtocolVersion: 1}},
		Packets:       []manifest.Packet{{ID: 1, Name: "FixturePacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{{Ordinal: 0, Name: "Shared Value", Encode: shared, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}},
	}
	files, err := GenerateFilesWithOptions(m, Options{Docs: docs.Overlay{
		Types:  map[string]string{"Shared": "Shared docs.", "FixturePacket": "Packet docs."},
		Fields: map[string]string{docs.FieldKey("Shared", "Value"): "Value docs.", docs.FieldKey("FixturePacket", "Shared Value"): "Packet value docs."},
	}})
	if err != nil {
		t.Fatalf("GenerateFilesWithOptions: %v", err)
	}
	if !strings.Contains(files["src/types.rs"], "/// Shared docs.") || !strings.Contains(files["src/types.rs"], "/// `value` docs.") {
		t.Fatalf("shared docs were not emitted:\n%s", files["src/types.rs"])
	}
	if !strings.Contains(files["src/packets.rs"], "/// Packet docs.") || !strings.Contains(files["src/packets.rs"], "/// Packet value docs.") {
		t.Fatalf("packet docs were not emitted:\n%s", files["src/packets.rs"])
	}
}

func TestGenerateRustFilesUseNativeEnumsAndPacketModules(t *testing.T) {
	packetType := manifest.Enum("zigzag_i32",
		manifest.EnumValue{Name: "EnableMultiplayer", Value: 0},
		manifest.EnumValue{Name: "DisableMultiplayer", Value: 1},
		manifest.EnumValue{Name: "RefreshJoincode", Value: 2},
	)
	packetType.Semantic = "MultiplayerSettingsPacketType"
	packetType.TypeID = "enums/MultiplayerSettingsPacketType"
	m := manifest.Manifest{SchemaVersion: 2, Target: manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:rust", MinecraftVersion: "fixture", ProtocolVersion: 2168}}, Packets: []manifest.Packet{{ID: 139, Name: "MultiplayerSettingsPacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{{Ordinal: 0, Name: "Packet Type", Encode: packetType, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}}}
	files, err := GenerateFiles(m)
	if err != nil {
		t.Fatal(err)
	}
	enums := files["src/enums.rs"]
	for _, text := range []string{"pub enum MultiplayerSettingsPacketType", "Enable,", "Disable,", "RefreshJoinCode,", "Unknown(i32)", "impl From<i32> for MultiplayerSettingsPacketType", "fn to_raw(self) -> i32"} {
		if !strings.Contains(enums, text) {
			t.Fatalf("native enum output omits %q:\n%s", text, enums)
		}
	}
	if strings.Contains(enums, "TryFrom<i32>") || strings.Contains(enums, "type Error = i32") {
		t.Fatalf("open enum retained closed decoding:\n%s", enums)
	}
	if strings.Contains(enums, "ENUMS") || strings.Contains(enums, "repr(transparent)") {
		t.Fatalf("enum output retained wrapper constants:\n%s", enums)
	}
	if _, ok := files["src/packets.rs"]; !ok {
		t.Fatalf("packet module was not emitted: %v", files)
	}
}

func TestGenerateRustFailsClosedForUnresolved(t *testing.T) {
	m := manifest.Manifest{
		SchemaVersion: 2,
		Target:        manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168},
		Sources:       []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:rust", MinecraftVersion: "fixture", ProtocolVersion: 2168}},
		Packets:       []manifest.Packet{{ID: 1, Name: "Blocked", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{{Ordinal: 0, Name: "Value", Encode: manifest.Unresolved("dynamic", true), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}},
	}
	if _, err := GenerateFiles(m); err == nil || !strings.Contains(err.Error(), "unresolved") {
		t.Fatalf("Generate error = %v, want unresolved failure", err)
	}
}

func TestGenerateRustEscapesKeywordFieldNames(t *testing.T) {
	m := manifest.Manifest{
		SchemaVersion: 2,
		Target:        manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168},
		Sources:       []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:rust", MinecraftVersion: "fixture", ProtocolVersion: 2168}},
		Packets:       []manifest.Packet{{ID: 1, Name: "KeywordPacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{{Ordinal: 0, Name: "Type", Encode: manifest.Primitive("u8"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}},
	}
	source, err := generatedRustSource(m)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(source, "pub type_: wire::U8") {
		t.Fatalf("generated Rust did not escape keyword field:\n%s", source)
	}
}

func TestRustFieldNamesUseSnakeCase(t *testing.T) {
	tests := map[string]string{
		"Actor Unique ID":   "actor_unique_id",
		"FrameRate":         "frame_rate",
		"Pack UUID":         "pack_uuid",
		"Sender's XUID":     "sender_xuid",
		"Tracked Actor IDs": "tracked_actor_ids",
		"No PvM":            "no_pvm",
		"No MvP":            "no_mvp",
		"Type":              "type_",
	}
	for input, want := range tests {
		if got := fieldName(input); got != want {
			t.Fatalf("fieldName(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestGenerateRustUsesCanonicalNamedTypesAndOrderedMapTuples(t *testing.T) {
	biome := manifest.Node{Kind: manifest.KindStruct, Semantic: "BiomeDefinitionData", TypeID: "BiomeDefinitionData", Fields: []manifest.Field{{Ordinal: 0, Name: "id", Encode: manifest.Primitive("u16le"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}
	mapping := manifest.Map(manifest.Primitive("var_u32"), manifest.Primitive("u16le"), biome)
	m := manifest.Manifest{SchemaVersion: 2, Target: manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:rust", MinecraftVersion: "fixture", ProtocolVersion: 2168}}, Packets: []manifest.Packet{{ID: 122, Name: "BiomeDefinitionListPacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{{Ordinal: 0, Name: "Map of Biome names to data", Encode: mapping, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}}}

	source, err := generatedRustSource(m)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(source, "pub struct BiomeDefinitionData") {
		t.Fatalf("generated Rust did not use canonical type identity:\n%s", source)
	}
	if !strings.Contains(source, "pub map_of_biome_names_to_data: Vec<(wire::U16LE, BiomeDefinitionData)>") {
		t.Fatalf("generated Rust did not use ordered map tuples:\n%s", source)
	}
	if strings.Contains(source, "BiomeDefinitionListPacketMapOfBiomeNamesToDataValueStruct") {
		t.Fatalf("generated Rust retained a path-derived type name:\n%s", source)
	}
}

func TestGenerateRustUsesTypedUnionEnum(t *testing.T) {
	union := manifest.Union(manifest.Primitive("var_u32"),
		manifest.Variant{Value: 0, Name: "SoundDataEvent::Stop", Encode: manifest.Void()},
		manifest.Variant{Value: 1, Name: "SoundDataEvent::SetVolume", Encode: manifest.Node{Kind: manifest.KindStruct, Semantic: "SoundDataEvent::SetVolume", TypeID: "SoundDataEvent::SetVolume", Fields: []manifest.Field{{Ordinal: 0, Name: "Volume", Encode: manifest.Primitive("f32le"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}},
	)
	m := manifest.Manifest{SchemaVersion: 2, Target: manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:rust", MinecraftVersion: "fixture", ProtocolVersion: 2168}}, Packets: []manifest.Packet{{ID: 348, Name: "SoundPacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{{Ordinal: 0, Name: "Event", Encode: union, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}}}

	source, err := generatedRustSource(m)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(source, "pub enum SoundDataEvent") || !strings.Contains(source, "SetVolume {") || !strings.Contains(source, "volume: wire::F32LE") || strings.Contains(source, "pub struct SoundDataEventSetVolume") || strings.Contains(source, "pub tag: i64") {
		t.Fatalf("generated Rust did not emit a typed union enum:\n%s", source)
	}
}

func TestGenerateRustDropsUnionDiscriminantPayloadField(t *testing.T) {
	discriminant := manifest.Enum("u8", manifest.EnumValue{Name: "Byte", Value: 0})
	discriminant.TypeID = "enums/DataItemType"
	variant := manifest.Node{Kind: manifest.KindStruct, TypeID: "DataItemByte", Fields: []manifest.Field{
		{Ordinal: 0, Name: "Type", Encode: discriminant, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		{Ordinal: 1, Name: "Value", Encode: manifest.Primitive("i8"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
	}}
	union := manifest.Union(manifest.Primitive("u8"), manifest.Variant{Value: 0, Name: "DataItemByte", Encode: variant})
	m := manifest.Manifest{SchemaVersion: 2, Target: manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:union-tag", MinecraftVersion: "fixture", ProtocolVersion: 2168}}, Packets: []manifest.Packet{{ID: 1, Name: "DataItemPacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{{Ordinal: 0, Name: "Value", Encode: union, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}}}
	source, err := generatedRustSource(m)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(source, "DataItemByte {\n        value: wire::I8,") || strings.Contains(source, "r#type: DataItemType") {
		t.Fatalf("union emitted a redundant discriminant payload field:\n%s", source)
	}
}

func TestGenerateRustPreservesWirePrimitiveTypesAndOptionalPresence(t *testing.T) {
	m := manifest.Manifest{SchemaVersion: 2, Target: manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:wire", MinecraftVersion: "fixture", ProtocolVersion: 2168}}, Packets: []manifest.Packet{{ID: 1, Name: "WirePacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{
		{Ordinal: 0, Name: "Count", Encode: manifest.Primitive("var_u32"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		{Ordinal: 1, Name: "Delta", Encode: manifest.Primitive("zigzag_i32"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		{Ordinal: 2, Name: "Value", Encode: manifest.Primitive("u32le"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		{Ordinal: 3, Name: "Ratio", Encode: manifest.Primitive("f32le"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		{Ordinal: 4, Name: "Maybe", Encode: manifest.Optional(manifest.Primitive("u8")), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
	}}}}
	files, err := GenerateFiles(m)
	if err != nil {
		t.Fatal(err)
	}
	packet := files["src/packets.rs"]
	for _, want := range []string{"pub count: wire::VarUInt", "pub delta: wire::ZigZag32", "pub value: wire::U32LE", "pub ratio: wire::F32LE", "/// Wire presence: optional value is preceded by a presence marker."} {
		if !strings.Contains(packet, want) {
			t.Fatalf("generated packet omitted %q:\n%s", want, packet)
		}
	}
	wire := files["src/wire.rs"]
	for _, want := range []string{"var_codec!(VarUInt, u32, 32, write_var_u32, read_var_u32)", "var_codec!(ZigZag32, i32, 32, write_zigzag_i32, read_zigzag_i32)", "fixed_codec!(U32LE, u32", "fixed_float_codec!(F32LE, f32", "fn encode"} {
		if !strings.Contains(wire, want) {
			t.Fatalf("wire module omitted %q:\n%s", want, wire)
		}
	}
}

func TestGenerateRustEmitsUnionDiscriminantMapping(t *testing.T) {
	union := manifest.Union(manifest.Primitive("u8"),
		manifest.Variant{Value: 1, Name: "First", Encode: manifest.Void()},
		manifest.Variant{Value: 7, Name: "Second", Encode: manifest.Void()},
	)
	m := manifest.Manifest{SchemaVersion: 2, Target: manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:union-wire", MinecraftVersion: "fixture", ProtocolVersion: 2168}}, Packets: []manifest.Packet{{ID: 1, Name: "UnionPacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{{Ordinal: 0, Name: "Value", Encode: union, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}}}
	source, err := generatedRustSource(m)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"pub fn discriminant(&self) -> u8", "Self::First => 1", "Self::Second => 7"} {
		if !strings.Contains(source, want) {
			t.Fatalf("union omitted %q:\n%s", want, source)
		}
	}
}

func TestGenerateRustEmitsPacketRegistryAndSum(t *testing.T) {
	fields := make([]manifest.Field, 8)
	for index := range fields {
		fields[index] = manifest.Field{Ordinal: index, Name: "Value" + string(rune('A'+index)), Encode: manifest.Primitive("u8"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}
	}
	m := manifest.Manifest{SchemaVersion: 2, Target: manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:packet-registry", MinecraftVersion: "fixture", ProtocolVersion: 2168}}, Packets: []manifest.Packet{
		{ID: 1, Name: "SmallPacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{{Ordinal: 0, Name: "Value", Encode: manifest.Primitive("u8"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}},
		{ID: 9, Name: "LargePacket", Direction: manifest.DirectionServerbound, Fields: fields},
	}}
	files, err := GenerateFiles(m)
	if err != nil {
		t.Fatal(err)
	}
	packets := files["src/packets.rs"]
	for _, want := range []string{"#[repr(u32)]\npub enum PacketId", "Small = 1", "Large = 9", "pub fn from_raw(raw: u32) -> Option<Self>", "pub enum Packet", "Small(Small)", "Large(Box<Large>)", "impl From<Small> for Packet", "impl From<Large> for Packet"} {
		if !strings.Contains(packets, want) {
			t.Fatalf("packet registry omitted %q:\n%s", want, packets)
		}
	}
}

func TestGenerateRustUsesDefaultsAndTupleWrappers(t *testing.T) {
	actor := manifest.Node{Kind: manifest.KindStruct, Semantic: "ActorRuntimeID", TypeID: "ActorRuntimeID", Fields: []manifest.Field{{Ordinal: 0, Name: "Actor Runtime ID", Encode: manifest.Primitive("var_u64"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}
	packet := manifest.Packet{ID: 1, Name: "DefaultPacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{{Ordinal: 0, Name: "Runtime", Encode: actor, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}
	m := manifest.Manifest{SchemaVersion: 2, Target: manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:defaults", MinecraftVersion: "fixture", ProtocolVersion: 2168}}, Packets: []manifest.Packet{packet}}
	files, err := GenerateFiles(m)
	if err != nil {
		t.Fatal(err)
	}
	types := files["src/types.rs"]
	packets := files["src/packets.rs"]
	for _, want := range []string{"pub struct ActorRuntimeID(pub u64);", "impl wire::Encode for ActorRuntimeID", "impl wire::Decode for ActorRuntimeID", "Default, PartialEq, Eq, Hash"} {
		if !strings.Contains(types, want) {
			t.Fatalf("generated types omitted %q:\n%s", want, types)
		}
	}
	if !strings.Contains(packets, "#[derive(Clone, Debug, Default, PartialEq)]") {
		t.Fatalf("packet did not derive Default:\n%s", packets)
	}
}

func TestGenerateRustBoxesLargeUnionFields(t *testing.T) {
	large := manifest.Node{Kind: manifest.KindStruct, Semantic: "LargeRecord", TypeID: "LargeRecord"}
	for index := 0; index < 8; index++ {
		large.Fields = append(large.Fields, manifest.Field{Ordinal: index, Name: "Text" + string(rune('A'+index)), Encode: manifest.String(manifest.Primitive("var_u32")), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}})
	}
	variant := manifest.Node{Kind: manifest.KindStruct, TypeID: "LargeUnionVariant", Fields: []manifest.Field{
		{Ordinal: 0, Name: "Payload", Encode: large, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
	}}
	union := manifest.Union(manifest.Primitive("u8"), manifest.Variant{Value: 0, Name: "Add", Encode: variant}, manifest.Variant{Value: 1, Name: "Remove", Encode: manifest.Void()})
	m := manifest.Manifest{SchemaVersion: 2, Target: manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:large-union", MinecraftVersion: "fixture", ProtocolVersion: 2168}}, Packets: []manifest.Packet{{ID: 1, Name: "LargeUnionPacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{{Ordinal: 0, Name: "Value", Encode: union, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}}}
	source, err := generatedRustSource(m)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(source, "payload: Box<LargeRecord>") {
		t.Fatalf("large union field was not boxed:\n%s", source)
	}
}

func TestGenerateRustUsesAddressableModulesAndCrateIdentity(t *testing.T) {
	m := manifest.Manifest{SchemaVersion: 2, Target: manifest.Target{MinecraftVersion: "1.26.40", ProtocolVersion: 2168}, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:modules", MinecraftVersion: "1.26.40", ProtocolVersion: 2168}}, Packets: []manifest.Packet{{ID: 1, Name: "FixturePacket", Direction: manifest.DirectionClientbound}}}
	files, err := GenerateFiles(m)
	if err != nil {
		t.Fatal(err)
	}
	lib := files["src/lib.rs"]
	for _, want := range []string{"pub mod enums;", "pub mod types;", "pub mod packets;", "pub mod wire;", "pub const PROTOCOL_VERSION: i32 = 2168;"} {
		if !strings.Contains(lib, want) {
			t.Fatalf("lib.rs omitted %q:\n%s", want, lib)
		}
	}
	if strings.Contains(lib, "pub use") || strings.Contains(lib, "allow(dead_code)") {
		t.Fatalf("lib.rs retained flat/glob API scaffolding:\n%s", lib)
	}
	for _, want := range []string{"name = \"bedrock-protocol-1-26-40\"", "version = \"0.1.0\"", "edition = \"2024\""} {
		if !strings.Contains(files["Cargo.toml"], want) {
			t.Fatalf("Cargo.toml omitted %q:\n%s", want, files["Cargo.toml"])
		}
	}
	if _, ok := files["src/packets.rs"]; !ok {
		t.Fatalf("collapsed packets module was not emitted: %v", files)
	}
	for path := range files {
		if strings.HasPrefix(path, "src/packets/") {
			t.Fatalf("flat packet module retained per-packet file %q", path)
		}
	}
}

func TestGenerateRustKeepsSharedUnionPayloadNamed(t *testing.T) {
	shared := manifest.Node{Kind: manifest.KindStruct, Semantic: "SharedRecord", TypeID: "SharedRecord", Fields: []manifest.Field{{Ordinal: 0, Name: "Value", Encode: manifest.Primitive("u8"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}
	union := manifest.Union(manifest.Primitive("u8"), manifest.Variant{Value: 0, Name: "Choice::Shared", Encode: shared})
	m := manifest.Manifest{SchemaVersion: 2, Target: manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:shared-rust", MinecraftVersion: "fixture", ProtocolVersion: 2168}}, Packets: []manifest.Packet{{ID: 1, Name: "EnvelopePacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{{Ordinal: 0, Name: "Choice", Encode: union, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}, {Ordinal: 1, Name: "Shared", Encode: shared, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}}}
	source, err := generatedRustSource(m)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(source, "Shared(SharedRecord)") || !strings.Contains(source, "pub struct SharedRecord") {
		t.Fatalf("shared union payload was incorrectly inlined:\n%s", source)
	}
}

func TestGenerateRustEmptyPacketOnlyEmitsDefinitionAndID(t *testing.T) {
	m := manifest.Manifest{SchemaVersion: 2, Target: manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:rust", MinecraftVersion: "fixture", ProtocolVersion: 2168}}, Packets: []manifest.Packet{{ID: 4, Name: "EmptyPacket", Direction: manifest.DirectionServerbound, Fields: []manifest.Field{}}}}
	files, err := GenerateFiles(m)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	packet := files["src/packets.rs"]
	if !strings.Contains(packet, "pub struct Empty") || !strings.Contains(packet, "pub const ID: u32 = 4;") {
		t.Fatalf("empty packet omitted its definition or ID:\n%s", packet)
	}
	// A fieldless packet still round trips, reading and writing nothing.
	for _, want := range []string{
		"impl wire::Encode for Empty {\n    fn encode(&self, writer: &mut wire::Writer) {\n        let _ = writer;\n    }\n}",
		"        Ok(Self {\n        })",
	} {
		if !strings.Contains(packet, want) {
			t.Fatalf("empty packet codec omitted %q:\n%s", want, packet)
		}
	}
}

func ptr(node manifest.Node) *manifest.Node { return &node }

func TestGenerateRustKeepsUnrelatedAnonymousUnionsDistinct(t *testing.T) {
	first := manifest.Union(manifest.Primitive("u8"), manifest.Variant{Value: 0, Name: "First", Encode: manifest.Void()})
	second := manifest.Union(manifest.Primitive("u8"), manifest.Variant{Value: 0, Name: "Second", Encode: manifest.Void()})
	m := manifest.Manifest{SchemaVersion: 2, Target: manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:rust", MinecraftVersion: "fixture", ProtocolVersion: 2168}}, Packets: []manifest.Packet{{ID: 1, Name: "Packet", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{{Ordinal: 0, Name: "Alpha", Encode: first, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}, {Ordinal: 1, Name: "Beta", Encode: second, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}}}
	source, err := generatedRustSource(m)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(source, "pub alpha: PacketAlpha") || !strings.Contains(source, "pub beta: PacketBeta") {
		t.Fatalf("anonymous unions were incorrectly deduplicated:\n%s", source)
	}
}

func TestRustPublicNamesDropSchemaScaffolding(t *testing.T) {
	tests := map[string]string{
		"enums/MoLangVersion":                           "MoLangVersion",
		"enums/MolangVersion":                           "MoLangVersion",
		"PlayerVideoCapturePacketPayload::Action":       "Action",
		"DataItemEntryPayloadUnion":                     "DataItemEntryValue",
		"DimensionDefinitionGroup::DimensionDefinition": "DimensionDefinition",
		"ServerWaypointGroup::Action":                   "ServerWaypointGroupAction",
		"SharedTypes::v1_26_0::CameraSplineDefinition":  "CameraSplineDefinition",
	}
	for input, want := range tests {
		if got := publicTypeName(input); got != want {
			t.Fatalf("publicTypeName(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestRustEnumVariantsUseIdiomaticCase(t *testing.T) {
	tests := map[string]string{
		"START_ATTACKING":   "StartAttacking",
		"RefreshJoincode":   "RefreshJoinCode",
		"EnableMultiplayer": "Enable",
	}
	for input, want := range tests {
		if got := enumVariantName("MultiplayerSettingsPacketType", input); got != want {
			t.Fatalf("enumVariantName(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestGenerateRustMarksPlaceholderUnionVariants(t *testing.T) {
	union := manifest.Union(manifest.Primitive("u8"), manifest.Variant{Value: 1, Name: "Empty1", Encode: manifest.Void()})
	m := manifest.Manifest{SchemaVersion: 2, Target: manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:placeholder", MinecraftVersion: "fixture", ProtocolVersion: 2168}}, Packets: []manifest.Packet{{ID: 1, Name: "PlaceholderPacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{{Ordinal: 0, Name: "Value", Encode: union, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}}}
	source, err := generatedRustSource(m)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(source, "Naming overlay required: source placeholder `Empty1`.") {
		t.Fatalf("placeholder variant was not marked for naming overlay:\n%s", source)
	}
}

func TestGenerateRustMapsCanonicalSemanticsToNativeTypes(t *testing.T) {
	vec3 := manifest.Node{Kind: manifest.KindStruct, Semantic: "Vec3", TypeID: "Vec3", Fields: []manifest.Field{
		{Ordinal: 0, Name: "X", Encode: manifest.Primitive("f32le"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		{Ordinal: 1, Name: "Y", Encode: manifest.Primitive("f32le"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		{Ordinal: 2, Name: "Z", Encode: manifest.Primitive("f32le"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
	}}
	m := manifest.Manifest{SchemaVersion: 2, Target: manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:native-rust", MinecraftVersion: "fixture", ProtocolVersion: 2168}}, Packets: []manifest.Packet{{ID: 1, Name: "NativePacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{
		{Ordinal: 0, Name: "ID", Encode: manifest.Primitive("uuid"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		{Ordinal: 1, Name: "Position", Encode: vec3, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		{Ordinal: 2, Name: "Data", Encode: manifest.NBT(manifest.NBTNetwork), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		{Ordinal: 3, Name: "Payload", Encode: manifest.Bytes(manifest.Primitive("var_u32")), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
	}}}}
	files, err := GenerateFiles(m)
	if err != nil {
		t.Fatal(err)
	}
	packet := files["src/packets.rs"]
	for _, want := range []string{"uuid::Uuid", "glam::Vec3", "wire::NetworkNbt", "bytes::Bytes"} {
		if !strings.Contains(packet, want) {
			t.Fatalf("native Rust output omits %q:\n%s", want, packet)
		}
	}
	if !strings.Contains(files["src/wire.rs"], "nbt_codec!(\n    NetworkNbt,\n    NbtVariant::Network,") || strings.Contains(files["src/types.rs"], "pub struct Vec3") {
		t.Fatalf("Rust shared types do not reflect native mapping:\n%s", files["src/types.rs"])
	}
	for _, dependency := range []string{`bytes = "1"`, `glam = "0.33"`, `uuid = "1"`} {
		if !strings.Contains(files["Cargo.toml"], dependency) {
			t.Fatalf("generated Cargo.toml omits %q:\n%s", dependency, files["Cargo.toml"])
		}
	}
}

func TestNativeRustTypeRejectsMislabelledVector(t *testing.T) {
	g := &generator{}
	node := manifest.Node{Kind: manifest.KindStruct, TypeID: "Vec2", Fields: []manifest.Field{{Encode: manifest.Primitive("f64le")}}}
	if _, matched, err := g.nativeRustType(node); !matched || err == nil {
		t.Fatalf("nativeRustType matched=%v error=%v, want a closed failure", matched, err)
	}
}

func TestGenerateRustCollapsesDoubleOptionalAndUsesNamedRecursiveType(t *testing.T) {
	dynamic := manifest.Union(manifest.Primitive("u8"),
		manifest.Variant{Value: 0, Name: "Empty", Encode: manifest.Void()},
		manifest.Variant{Value: 1, Name: "List", Encode: manifest.Array(manifest.Primitive("var_u32"), manifest.Recursive("cereal::DynamicValue"))},
	)
	dynamic.TypeID = "cereal::DynamicValue"
	m := manifest.Manifest{
		SchemaVersion: 2,
		Target:        manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168},
		Sources:       []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:rust", MinecraftVersion: "fixture", ProtocolVersion: 2168}},
		Packets: []manifest.Packet{{ID: 1, Name: "FixturePacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{
			{Ordinal: 0, Name: "Maybe", Encode: manifest.Optional(manifest.Optional(manifest.Primitive("i32le"))), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
			{Ordinal: 1, Name: "Dynamic", Encode: dynamic, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
			{Ordinal: 2, Name: "Flags", Encode: manifest.Bitset(131), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		}}},
	}
	files, err := GenerateFiles(m)
	if err != nil {
		t.Fatal(err)
	}
	var all strings.Builder
	for _, source := range files {
		all.WriteString(source)
	}
	for _, want := range []string{"pub maybe: Option<wire::I32LE>", "Vec<CerealDynamicValue>", "pub struct Bitset131(pub [u64; 3])", "pub flags: Bitset131"} {
		if !strings.Contains(all.String(), want) {
			t.Fatalf("generated output missing %q:\n%s", want, all.String())
		}
	}
	if strings.Contains(all.String(), "Option<Option<i32>>") {
		t.Fatalf("generated Rust retained public double optional:\n%s", all.String())
	}
}

func TestGenerateRustFailsClosedForAsymmetricField(t *testing.T) {
	decode := manifest.Primitive("u16le")
	field := manifest.Field{Ordinal: 0, Name: "Value", Encode: manifest.Primitive("u8"), Decode: &decode, Symmetry: manifest.Asymmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}
	m := manifest.Manifest{SchemaVersion: 2, Target: manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 1}, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "1", Digest: "fixture", MinecraftVersion: "fixture", ProtocolVersion: 1}}, Packets: []manifest.Packet{{ID: 1, Name: "AsymmetricPacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{field}}}}
	if _, err := GenerateFiles(m); err == nil || !strings.Contains(err.Error(), "asymmetric") {
		t.Fatalf("GenerateFiles error = %v, want asymmetric failure", err)
	}
}

func TestGenerateRustPreservesNBTEncodingTypes(t *testing.T) {
	m := manifest.Manifest{SchemaVersion: manifest.SchemaVersion, Target: manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:nbt", MinecraftVersion: "fixture", ProtocolVersion: 2168}}, Packets: []manifest.Packet{{ID: 1, Name: "NBTPacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{
		{Ordinal: 0, Name: "Network", Encode: manifest.NBT(manifest.NBTNetwork), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		{Ordinal: 1, Name: "Persistent", Encode: manifest.NBT(manifest.NBTPersistent), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
	}}}}
	files, err := GenerateFiles(m)
	if err != nil {
		t.Fatal(err)
	}
	packet := files["src/packets.rs"]
	for _, want := range []string{"pub network: wire::NetworkNbt", "pub persistent: wire::PersistentNbt"} {
		if !strings.Contains(packet, want) {
			t.Fatalf("packet omitted %q:\n%s", want, packet)
		}
	}
	wire := files["src/wire.rs"]
	for _, want := range []string{"NetworkNbt,\n    NbtVariant::Network,", "PersistentNbt,\n    NbtVariant::Persistent,"} {
		if !strings.Contains(wire, want) {
			t.Fatalf("wire module omitted %q:\n%s", want, wire)
		}
	}
}

func TestGenerateRustEmitsPacketCodecs(t *testing.T) {
	m := manifest.Manifest{SchemaVersion: 2, Target: manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:codec", MinecraftVersion: "fixture", ProtocolVersion: 2168}}, Packets: []manifest.Packet{{ID: 1, Name: "CodecPacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{
		{Ordinal: 0, Name: "Count", Encode: manifest.Primitive("var_u32"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
	}}}}
	files, err := GenerateFiles(m)
	if err != nil {
		t.Fatal(err)
	}
	packets := files["src/packets.rs"]
	for _, want := range []string{
		"impl wire::Encode for Codec {",
		"impl wire::Decode for Codec {",
		"self.count.encode(writer);",
		// The decode site names its type, so a disagreement with the type
		// emitter is a compile error rather than silent inference.
		"let count = <wire::VarUInt as wire::Decode>::decode(reader)?;",
	} {
		if !strings.Contains(packets, want) {
			t.Fatalf("packet codec omitted %q:\n%s", want, packets)
		}
	}
}

// A declared collection length must be bounded before it is used to reserve.
func TestGenerateRustBoundsCollectionsBeforeAllocating(t *testing.T) {
	array := manifest.Node{Kind: manifest.KindArray, Prefix: ptr(manifest.Primitive("var_u32")), Element: ptr(manifest.Primitive("u64le"))}
	m := manifest.Manifest{SchemaVersion: 2, Target: manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:bounds", MinecraftVersion: "fixture", ProtocolVersion: 2168}}, Packets: []manifest.Packet{{ID: 1, Name: "BoundsPacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{
		{Ordinal: 0, Name: "Items", Encode: array, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
	}}}}
	files, err := GenerateFiles(m)
	if err != nil {
		t.Fatal(err)
	}
	packets := files["src/packets.rs"]
	// Element minimum is 8 bytes, so the count is checked against both the
	// element limit and the bytes actually remaining.
	if !strings.Contains(packets, "wire::decode_collection::<wire::U64LE>(reader, 8)?") {
		t.Fatalf("collection decode is not bounded:\n%s", packets)
	}
	wire := files["src/wire.rs"]
	for _, want := range []string{
		"pub fn checked_count(",
		"LengthLimitExceeded",
		"LengthNotRepresentable",
		// The bound is a reader setting, matching the Go backend, so a peer
		// sending a legitimately large collection is recoverable.
		"pub fn set_collection_limit(",
	} {
		if !strings.Contains(wire, want) {
			t.Fatalf("wire runtime omitted %q", want)
		}
	}
}

func TestGenerateRustEmitsDirectionRegistry(t *testing.T) {
	m := manifest.Manifest{SchemaVersion: 2, Target: manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:direction", MinecraftVersion: "fixture", ProtocolVersion: 2168}}, Packets: []manifest.Packet{
		{ID: 1, Name: "DownPacket", Direction: manifest.DirectionClientbound},
		{ID: 2, Name: "UpPacket", Direction: manifest.DirectionServerbound},
	}}
	files, err := GenerateFiles(m)
	if err != nil {
		t.Fatal(err)
	}
	packets := files["src/packets.rs"]
	for _, want := range []string{
		"Self::Down => Direction::Clientbound,",
		"Self::Up => Direction::Serverbound,",
		"pub fn decode_from(",
		"if !packet.direction().permits(sender) {",
		// A direction violation is distinct from an unrecognised id: one is a
		// protocol violation, the other may just be a newer peer.
		"return Err(wire::DecodeError::UnexpectedDirection(id));",
		"pub fn decode_exact_from(",
		"pub const ALL: &'static [PacketId] = &[",
	} {
		if !strings.Contains(packets, want) {
			t.Fatalf("direction registry omitted %q:\n%s", want, packets)
		}
	}
}

// Cereal's always-present outer optional keeps both markers on the wire even
// though the Rust type exposes only the inner state.
func TestGenerateRustWritesBothNestedOptionalMarkers(t *testing.T) {
	nested := manifest.Optional(manifest.Optional(manifest.Primitive("u8")))
	m := manifest.Manifest{SchemaVersion: 2, Target: manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:nested", MinecraftVersion: "fixture", ProtocolVersion: 2168}}, Packets: []manifest.Packet{{ID: 1, Name: "NestedPacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{
		{Ordinal: 0, Name: "Value", Encode: nested, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
	}}}}
	files, err := GenerateFiles(m)
	if err != nil {
		t.Fatal(err)
	}
	packets := files["src/packets.rs"]
	if !strings.Contains(packets, "writer.write_u8(1);\n        match &self.value {") {
		t.Fatalf("nested optional did not write the outer marker:\n%s", packets)
	}
	if got := strings.Count(packets, "reader.read_u8()? != 0"); got != 2 {
		t.Fatalf("nested optional read %d presence markers, want 2:\n%s", got, packets)
	}
}

func TestGenerateRustRejectsUnknownUnionDiscriminant(t *testing.T) {
	union := manifest.Union(manifest.Primitive("u8"),
		manifest.Variant{Value: 1, Name: "First", Encode: manifest.Void()},
	)
	m := manifest.Manifest{SchemaVersion: 2, Target: manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:union-unknown", MinecraftVersion: "fixture", ProtocolVersion: 2168}}, Packets: []manifest.Packet{{ID: 1, Name: "UnionUnknownPacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{{Ordinal: 0, Name: "Value", Encode: union, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}}}
	source, err := generatedRustSource(m)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(source, "wire::DecodeError::UnknownVariant {") {
		t.Fatalf("union decode does not reject unknown discriminants:\n%s", source)
	}
}

func TestRustEnumAliasesEncodeAndDecode(t *testing.T) {
	var source strings.Builder
	emitRustEnum(&source, definition{Name: "GameType", Underlying: "i32", Variants: manifest.Enum("zigzag_i32",
		manifest.EnumValue{Name: "Undefined", Value: -1},
		manifest.EnumValue{Name: "Survival", Value: 0},
		manifest.EnumValue{Name: "Creative", Value: 1},
		manifest.EnumValue{Name: "Adventure", Value: 2},
		manifest.EnumValue{Name: "Default", Value: 5},
		manifest.EnumValue{Name: "Spectator", Value: 6},
		manifest.EnumValue{Name: "WorldDefault", Value: 0},
	).Variants})
	if strings.Count(source.String(), "0 => Self::") != 1 || !strings.Contains(source.String(), "0 => Self::Survival") || !strings.Contains(source.String(), "Self::WorldDefault => 0") {
		t.Fatalf("enum aliases must encode independently and decode canonically:\n%s", source.String())
	}
	rustc, err := exec.LookPath("rustc")
	if err != nil {
		t.Skip("rustc unavailable; verified emitted alias mappings above")
	}
	source.WriteString(`
fn main() {
    assert_eq!(GameType::Survival.to_raw(), 0);
    assert_eq!(GameType::WorldDefault.to_raw(), 0);
    assert_eq!(GameType::from(0), GameType::Survival);
    assert_eq!(GameType::from(-1), GameType::Undefined);
    assert_eq!(GameType::from(5), GameType::Default);
    assert_eq!(GameType::from(6), GameType::Spectator);
    assert_eq!(GameType::from(3), GameType::Unknown(3));
    assert_eq!(GameType::Unknown(3).to_raw(), 3);
    assert_eq!(GameType::default(), GameType::Undefined);
    assert_eq!(GameType::from(GameType::default().to_raw()), GameType::default());
}
`)
	dir := t.TempDir()
	path, binary := filepath.Join(dir, "aliases.rs"), filepath.Join(dir, "aliases")
	if err := os.WriteFile(path, []byte(source.String()), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(rustc, "--edition=2021", "-D", "unreachable_patterns", path, "-o", binary).CombinedOutput(); err != nil {
		t.Fatalf("compile Rust aliases: %v\n%s", err, out)
	}
	if out, err := exec.Command(binary).CombinedOutput(); err != nil {
		t.Fatalf("run Rust aliases: %v\n%s", err, out)
	}
}

func TestRustEnumAliasCanonicalNameSurvivesManifestRoundTrip(t *testing.T) {
	value := manifest.Enum("zigzag_i32",
		manifest.EnumValue{Name: "Zulu", Value: 0},
		manifest.EnumValue{Name: "Alpha", Value: 0},
		manifest.EnumValue{Name: "One", Value: 1},
	)
	value.Semantic, value.TypeID = "AliasType", "enums/AliasType"
	m := manifest.Manifest{
		SchemaVersion: 2,
		Target:        manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168},
		Sources:       []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:aliases", MinecraftVersion: "fixture", ProtocolVersion: 2168}},
		Packets:       []manifest.Packet{{ID: 1, Name: "AliasPacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{{Ordinal: 0, Name: "Alias", Encode: value, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}},
	}
	before, err := GenerateFiles(m)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "manifest.json")
	if err := manifest.Write(path, m); err != nil {
		t.Fatal(err)
	}
	loaded, err := manifest.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	after, err := GenerateFiles(loaded)
	if err != nil {
		t.Fatal(err)
	}
	if before["src/enums.rs"] != after["src/enums.rs"] {
		t.Fatalf("manifest round trip changed alias decoding or default:\nbefore:\n%s\nafter:\n%s", before["src/enums.rs"], after["src/enums.rs"])
	}
	if !strings.Contains(after["src/enums.rs"], "0 => Self::Zulu") {
		t.Fatalf("canonical alias lost declared order:\n%s", after["src/enums.rs"])
	}
	first, err := manifest.MarshalStable(loaded)
	if err != nil {
		t.Fatal(err)
	}
	if err := manifest.Write(path, loaded); err != nil {
		t.Fatal(err)
	}
	second, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Fatal("alias manifest serialization is not stable")
	}
}

// An enum type used at sites that each allow a subset keeps every site's values.
func TestGenerateRustMergesSharedEnumValuesAcrossSites(t *testing.T) {
	first := manifest.Enum("u8", manifest.EnumValue{Name: "raw", Value: 0}, manifest.EnumValue{Name: "tip", Value: 5})
	second := manifest.Enum("u8", manifest.EnumValue{Name: "chat", Value: 1})
	first.Semantic, first.TypeID = "TextPacketType", "enums/TextPacketType"
	second.Semantic, second.TypeID = first.Semantic, first.TypeID
	pins := manifest.Provenance{Pins: []string{"fixture"}}
	m := manifest.Manifest{
		SchemaVersion: 2,
		Target:        manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168},
		Sources:       []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:shared-enum", MinecraftVersion: "fixture", ProtocolVersion: 2168}},
		Packets: []manifest.Packet{{ID: 1, Name: "TextPacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{
			{Ordinal: 0, Name: "First", Encode: first, Symmetry: manifest.Symmetric, Provenance: pins},
			{Ordinal: 1, Name: "Second", Encode: second, Symmetry: manifest.Symmetric, Provenance: pins},
		}}},
	}
	for _, contract := range []int{0, CodecContractVersion} {
		files, err := GenerateFilesWithOptions(m, Options{CodecContract: contract})
		if err != nil {
			t.Fatal(err)
		}
		enums := files["src/enums.rs"]
		for _, want := range []string{"0 => Self::Raw", "1 => Self::Chat", "5 => Self::Tip"} {
			if !strings.Contains(enums, want) {
				t.Fatalf("contract %d omits shared enum value %q:\n%s", contract, want, enums)
			}
		}
	}
}
