package emitgo

import (
	"go/format"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"protocolgen/internal/docs"
	"protocolgen/internal/domains"
	"protocolgen/internal/layout"
	"protocolgen/internal/licensing"
	"protocolgen/internal/manifest"
	"protocolgen/internal/naming"
	"protocolgen/internal/semantics"
)

func TestMarshalEmitterUsesOneAddressParameterizedWalk(t *testing.T) {
	source, err := os.ReadFile("emit.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	if strings.Count(text, "func (e *marshalEmitter) node(") != 1 {
		t.Fatalf("marshal emitter has more than one node walk")
	}
	for _, duplicate := range []string{"nodePointer", "optionalCallPointer", "collectionPointer", "mapEntriesPointer"} {
		if strings.Contains(text, duplicate) {
			t.Fatalf("marshal emitter retains duplicate helper %q", duplicate)
		}
	}
}

func TestGenerateConsumesOnlyCanonicalManifest(t *testing.T) {
	m := manifest.Manifest{
		SchemaVersion: 2,
		Target:        manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168},
		Sources:       []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture-2168", Digest: "fixture:golden", MinecraftVersion: "fixture", ProtocolVersion: 2168}},
		Packets: []manifest.Packet{{ID: 1, Name: "VocabularyPacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{
			{Ordinal: 0, Name: "Value", Encode: manifest.Primitive("u8"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
			{Ordinal: 1, Name: "Maybe", Encode: manifest.Optional(manifest.String(manifest.Primitive("var_u32"))), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		}}},
	}
	files, err := Generate(m, "wiregen")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	for name, source := range files {
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		if _, err := format.Source([]byte(source)); err != nil {
			t.Fatalf("%s is not valid Go: %v\n%s", name, err, source)
		}
		if strings.Contains(source, "gophertunnel") || strings.Contains(source, "bedrock-protocol-docs") {
			t.Errorf("%s contains source/profile lookup text", name)
		}
	}
	for name, want := range map[string]string{"LICENSE-MIT": licensing.MIT, "NOTICE": licensing.Notice} {
		if files[name] != want {
			t.Errorf("%s differs from the canonical contribution notice", name)
		}
	}
	packet := files["protocol/packet/vocabulary.go"]
	version := files["protocol/version.go"]
	if !strings.Contains(version, `GAME_VERSION     = "fixture"`) || !strings.Contains(version, "PROTOCOL_VERSION = 2168") {
		t.Fatalf("generated version constants are incomplete:\n%s", version)
	}
	if !strings.Contains(packet, "type Vocabulary struct") || !strings.Contains(packet, "Maybe protocol.Optional[string]") || !strings.Contains(files["protocol/packet/ids.go"], "IDVocabulary uint32 = 1") {
		t.Fatalf("generated output omitted packet definition or ID:\n%s\n%s", packet, files["protocol/packet/ids.go"])
	}
	for _, want := range []string{"func (pk *Vocabulary) Marshal(io protocol.IO)", "io.Uint8(&pk.Value)", "protocol.OptionalFunc(io, &pk.Maybe, io.String)"} {
		if !strings.Contains(packet, want) {
			t.Fatalf("generated packet marshal omits %q:\n%s", want, packet)
		}
	}
	if strings.Contains(packet, "Shape{") || strings.Contains(packet, "func (p *Vocabulary) Encode") {
		t.Fatalf("generated packet exposed runtime schema or placeholder codecs:\n%s", packet)
	}
}

func TestGenerateEmitsSchemaConstraintChecks(t *testing.T) {
	text := manifest.String(manifest.Primitive("var_u32"))
	text.Constraints = &manifest.Constraints{MinLength: pointerTo(uint64(1)), MaxLength: pointerTo(uint64(65536)), Pattern: "^[a-z]+$"}
	count := manifest.Array(manifest.Primitive("var_u32"), manifest.Primitive("u16le"))
	count.Constraints = &manifest.Constraints{MaxItems: pointerTo(uint64(4))}
	raw := manifest.Array(manifest.Primitive("var_u32"), manifest.Primitive("u8"))
	raw.Constraints = &manifest.Constraints{MaxItems: pointerTo(uint64(4))}
	number := manifest.Primitive("zigzag_i32")
	number.Constraints = &manifest.Constraints{Minimum: pointerTo(-1.0), Maximum: pointerTo(64.0)}
	m := manifest.Manifest{SchemaVersion: 2, Target: manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 1}, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "1", Digest: "fixture", MinecraftVersion: "fixture", ProtocolVersion: 1}}, Packets: []manifest.Packet{{ID: 1, Name: "LimitsPacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{
		{Ordinal: 0, Name: "Text", Encode: text, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		{Ordinal: 1, Name: "Values", Encode: count, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		{Ordinal: 2, Name: "Number", Encode: number, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		{Ordinal: 3, Name: "Raw", Encode: raw, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
	}}}}
	files, err := Generate(m, "wiregen")
	if err != nil {
		t.Fatal(err)
	}
	packet := files["protocol/packet/limits.go"]
	for _, want := range []string{
		"io.StringLimits(&pk.Text, 1, 65536)",
		`protocol.Pattern(io, &pk.Text, "^[a-z]+$")`,
		"protocol.FuncSliceLimits(io, &pk.Values, io.Varuint32, 0, 4, io.Uint16)",
		"Raw    []uint8",
		"io.ByteSliceLimits(&pk.Raw, 0, 4)",
		"protocol.Minimum(io, &pk.Number, -1)",
		"protocol.Maximum(io, &pk.Number, 64)",
	} {
		if !strings.Contains(packet, want) {
			t.Fatalf("generated packet omits %q:\n%s", want, packet)
		}
	}
}

func pointerTo[T any](value T) *T { return &value }

func TestGenerateGroupsSharedDefinitionsByReviewedDomain(t *testing.T) {
	alpha := manifest.Node{Kind: manifest.KindStruct, TypeID: "Alpha", Fields: []manifest.Field{{Name: "Value", Encode: manifest.Primitive("u8"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}
	beta := manifest.Node{Kind: manifest.KindStruct, TypeID: "Beta", Fields: []manifest.Field{{Name: "Value", Encode: manifest.Primitive("u8"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}
	m := manifest.Manifest{
		SchemaVersion: 2,
		Target:        manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 1},
		Sources:       []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "1", Digest: "fixture", MinecraftVersion: "fixture", ProtocolVersion: 1}},
		Packets: []manifest.Packet{{ID: 1, Name: "FixturePacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{
			{Ordinal: 0, Name: "Beta", Encode: beta, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
			{Ordinal: 1, Name: "Alpha", Encode: alpha, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		}}, {ID: 2, Name: "OtherPacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{
			{Ordinal: 0, Name: "Beta", Encode: beta, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
			{Ordinal: 1, Name: "Alpha", Encode: alpha, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		}}},
	}
	files, err := GenerateWithOptions(m, Options{
		ProtocolImportPath: "fixture",
		NativeTypes:        false,
		EmitPacketRuntime:  true,
		EmitPacketPools:    true,
		Domains: domains.Overlay{Domains: map[string]string{
			"Alpha": "shared",
			"Beta":  "shared",
		}},
	})
	if err != nil {
		t.Fatalf("GenerateWithOptions: %v", err)
	}
	source := files["protocol/shared.go"]
	if !strings.Contains(source, "type Alpha struct") || !strings.Contains(source, "type Beta struct") {
		t.Fatalf("shared domain file omitted definitions:\n%s", source)
	}
	if strings.Index(source, "type Alpha struct") > strings.Index(source, "type Beta struct") {
		t.Fatalf("definitions are not sorted by name within domain:\n%s", source)
	}
	if _, ok := files["protocol/alpha.go"]; ok {
		t.Fatalf("domain grouping retained per-type file")
	}
}

func TestGenerateEmitsReviewedGoDocs(t *testing.T) {
	shared := manifest.Node{Kind: manifest.KindStruct, TypeID: "Shared", Fields: []manifest.Field{{Name: "Value", Encode: manifest.Primitive("u8"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}
	m := manifest.Manifest{
		SchemaVersion: 2,
		Target:        manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 1},
		Sources:       []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "1", Digest: "fixture", MinecraftVersion: "fixture", ProtocolVersion: 1}},
		Packets: []manifest.Packet{
			{ID: 1, Name: "FixturePacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{{Name: "Shared Value", Encode: shared, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}},
			{ID: 2, Name: "OtherPacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{{Name: "Shared Value", Encode: shared, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}},
		},
	}
	files, err := GenerateWithOptions(m, Options{ProtocolImportPath: "fixture", NativeTypes: false, Docs: docs.Overlay{
		Types:  map[string]string{"Shared": "Shared docs.", "FixturePacket": "Packet docs."},
		Fields: map[string]string{docs.FieldKey("Shared", "Value"): "Value docs.", docs.FieldKey("FixturePacket", "Shared Value"): "Packet value docs."},
	}})
	if err != nil {
		t.Fatalf("GenerateWithOptions: %v", err)
	}
	if !strings.Contains(files["protocol/shared.go"], "// Shared docs.") || !strings.Contains(files["protocol/shared.go"], "// Value docs.") {
		t.Fatalf("shared docs were not emitted:\n%s", files["protocol/shared.go"])
	}
	if !strings.Contains(files["protocol/packet/fixture.go"], "// Packet docs.") || !strings.Contains(files["protocol/packet/fixture.go"], "// Packet value docs.") {
		t.Fatalf("packet docs were not emitted:\n%s", files["protocol/packet/fixture.go"])
	}
}

func TestGenerateUsesRuntimeHelpersForEnumsAndOptionals(t *testing.T) {
	value := manifest.Enum("u8", manifest.EnumValue{Name: "Zero", Value: 0}, manifest.EnumValue{Name: "One", Value: 1})
	entry := manifest.Struct(manifest.Field{
		Ordinal:    0,
		Name:       "Value",
		Encode:     manifest.Primitive("u8"),
		Symmetry:   manifest.Symmetric,
		Provenance: manifest.Provenance{Pins: []string{"fixture"}},
	})
	m := manifest.Manifest{
		SchemaVersion: 2,
		Target:        manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168},
		Sources:       []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:helpers", MinecraftVersion: "fixture", ProtocolVersion: 2168}},
		Packets: []manifest.Packet{{ID: 1, Name: "HelperPacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{
			{Ordinal: 0, Name: "Kind", Encode: value, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
			{Ordinal: 1, Name: "Maybe", Encode: manifest.Optional(manifest.Primitive("i32le")), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
			{Ordinal: 2, Name: "Kinds", Encode: manifest.Array(manifest.Primitive("var_u32"), value), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
			{Ordinal: 3, Name: "Entries", Encode: manifest.Array(manifest.Primitive("var_u32"), entry), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
			{Ordinal: 4, Name: "MaybeKinds", Encode: manifest.Optional(manifest.Array(manifest.Primitive("var_u32"), value)), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		}}},
	}
	files, err := Generate(m, "wiregen")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	source := files["protocol/packet/helper.go"]
	for _, want := range []string{
		"pk.Kind.Marshal(io)",
		"protocol.OptionalFunc(io, &pk.Maybe, io.Int32)",
		"protocol.Slice(io, &pk.Kinds)",
		"protocol.Slice(io, &pk.Entries)",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated packet omitted runtime helper %q:\n%s", want, source)
		}
	}
	if strings.Contains(source, "io.Reading()") || strings.Contains(source, "unknown enum value") {
		t.Fatalf("ordinary enum/optional mechanics leaked into packet code:\n%s", source)
	}
	if strings.Contains(source, "item := *value") || strings.Contains(source, "*value = item") {
		t.Fatalf("callback codec copied values instead of using the supplied pointer:\n%s", source)
	}
	if !strings.Contains(source, "protocol.Slice(io, value)") {
		t.Fatalf("nested collection callback did not retain the supplied slice pointer:\n%s", source)
	}
	if strings.Contains(source, "FuncSlice(io, &pk.") || strings.Contains(source, "IntegerFunc") {
		t.Fatalf("self-marshaling values still emit an escaping callback:\n%s", source)
	}
	if !strings.Contains(generatedSource(files), "Marshal(io IO) { io.Uint8((*uint8)(x)) }") {
		t.Fatalf("enum does not marshal itself:\n%s", generatedSource(files))
	}
}

func TestGenerateKeepsUnionValidation(t *testing.T) {
	union := manifest.Union(manifest.Primitive("u8"),
		manifest.Variant{Value: 0, Name: "First", Encode: manifest.Void()},
		manifest.Variant{Value: 1, Name: "Second", Encode: manifest.Void()},
	)
	m := manifest.Manifest{
		SchemaVersion: 2,
		Target:        manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168},
		Sources:       []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:union", MinecraftVersion: "fixture", ProtocolVersion: 2168}},
		Packets:       []manifest.Packet{{ID: 1, Name: "UnionPacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{{Ordinal: 0, Name: "Value", Encode: union, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}},
	}
	files, err := Generate(m, "wiregen")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	marshal := generatedSource(files)
	if !strings.Contains(marshal, "unknown union tag") {
		t.Fatalf("union discriminator validation was removed:\n%s", marshal)
	}
	for name, source := range files {
		if name != "protocol/helpers.go" && strings.Contains(source, "io.Reading()") {
			t.Fatalf("union direction mechanics leaked into %s:\n%s", name, source)
		}
	}
}

func TestGenerateIncludesConcreteCodecRuntime(t *testing.T) {
	m := manifest.Manifest{
		SchemaVersion: 2,
		Target:        manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168},
		Sources:       []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:runtime", MinecraftVersion: "fixture", ProtocolVersion: 2168}},
		Packets:       []manifest.Packet{{ID: 1, Name: "RuntimePacket", Direction: manifest.DirectionClientbound}},
	}
	files, err := Generate(m, "wiregen")
	if err != nil {
		t.Fatal(err)
	}
	for name, wants := range map[string][]string{
		"protocol/codec.go":   {"type IO interface"},
		"protocol/helpers.go": {"func OptionalMarshaler", "func OptionalFunc", "func Union", "func FuncSlice"},
		"protocol/reader.go":  {"type Reader struct", "func NewReader", "func (r *Reader) NBT", "func (r *Reader) SliceLength"},
		"protocol/writer.go":  {"type Writer struct", "func NewWriter", "func (w *Writer) NBT", "func (w *Writer) Data"},
		"protocol/types.go":   {"type Optional", "type OrderedEntry"},
	} {
		source, ok := files[name]
		if !ok {
			t.Fatalf("generated output omits %s", name)
		}
		for _, want := range wants {
			if !strings.Contains(source, want) {
				t.Fatalf("%s omits %q:\n%s", name, want, source)
			}
		}
	}
	packetRuntime := files["protocol/packet/packet.go"]
	for _, want := range []string{"func Decode(data []byte, pk Packet) error", "func Encode(pk Packet) ([]byte, error)"} {
		if !strings.Contains(packetRuntime, want) {
			t.Fatalf("packet runtime omits %q:\n%s", want, packetRuntime)
		}
	}
}

func TestGenerateWrapsRepeatedUnionPayloadTypes(t *testing.T) {
	message := manifest.Struct(manifest.Field{Ordinal: 0, Name: "Text", Encode: manifest.String(manifest.Primitive("var_u32")), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}})
	message.TypeID = "Message"
	union := manifest.Union(manifest.Primitive("u8"),
		manifest.Variant{Value: 0, Name: "First", Encode: message},
		manifest.Variant{Value: 1, Name: "Second", Encode: message},
	)
	union.TypeID = "Choice"
	m := manifest.Manifest{SchemaVersion: 2, Target: manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 1}, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "1", Digest: "fixture", MinecraftVersion: "fixture", ProtocolVersion: 1}}, Packets: []manifest.Packet{{ID: 1, Name: "EnvelopePacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{{Ordinal: 0, Name: "Choice", Encode: union, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}}}
	files, err := Generate(m, "fixture")
	if err != nil {
		t.Fatal(err)
	}
	marshal := generatedSource(files)
	if !strings.Contains(marshal, "func (*Message) tagChoice() uint8 { return 0 }") || !strings.Contains(marshal, "func (*ChoiceSecond) tagChoice() uint8 { return 1 }") {
		t.Fatalf("repeated union payloads were not made tag-distinct:\n%s", marshal)
	}
	if !strings.Contains(marshal, "return new(Message)") || !strings.Contains(marshal, "return new(ChoiceSecond)") {
		t.Fatalf("union decode does not allocate pointer payloads:\n%s", marshal)
	}
}

func TestGenerateCodecFailsClosedForSequence(t *testing.T) {
	m := manifest.Manifest{SchemaVersion: 2, Target: manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 1}, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "1", Digest: "fixture", MinecraftVersion: "fixture", ProtocolVersion: 1}}, Packets: []manifest.Packet{{ID: 1, Name: "SequencePacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{{Ordinal: 0, Name: "Value", Encode: manifest.Sequence(manifest.Primitive("u8")), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}}}
	if _, err := Generate(m, "fixture"); err == nil || !strings.Contains(err.Error(), "sequence") {
		t.Fatalf("Generate error = %v, want sequence failure", err)
	}
}

func TestGenerateCodecFailsClosedForAsymmetricField(t *testing.T) {
	decode := manifest.Primitive("u16le")
	field := manifest.Field{Ordinal: 0, Name: "Value", Encode: manifest.Primitive("u8"), Decode: &decode, Symmetry: manifest.Asymmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}
	m := manifest.Manifest{SchemaVersion: 2, Target: manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 1}, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "1", Digest: "fixture", MinecraftVersion: "fixture", ProtocolVersion: 1}}, Packets: []manifest.Packet{{ID: 1, Name: "AsymmetricPacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{field}}}}
	if _, err := Generate(m, "fixture"); err == nil || !strings.Contains(err.Error(), "asymmetric") {
		t.Fatalf("Generate error = %v, want asymmetric failure", err)
	}
}

func TestGenerateCollapsesCerealDoubleOptionalToOnePublicOptional(t *testing.T) {
	doubleOptional := manifest.Optional(manifest.Optional(manifest.Primitive("i32le")))
	m := manifest.Manifest{
		SchemaVersion: 2,
		Target:        manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168},
		Sources:       []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:optional", MinecraftVersion: "fixture", ProtocolVersion: 2168}},
		Packets:       []manifest.Packet{{ID: 1, Name: "OptionalPacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{{Ordinal: 0, Name: "Value", Encode: doubleOptional, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}},
	}
	files, err := Generate(m, "wiregen")
	if err != nil {
		t.Fatal(err)
	}
	packet := files["protocol/packet/optional.go"]
	if !strings.Contains(packet, "Value protocol.Optional[int32]") || strings.Contains(packet, "**int32") || strings.Contains(packet, "Optional[Optional[") {
		t.Fatalf("double optional did not use one public Optional:\n%s", packet)
	}
	types := files["protocol/types.go"]
	for _, want := range []string{"type Optional[T any] struct", "func Option[T any]", "func (o Optional[T]) Value() (T, bool)"} {
		if !strings.Contains(types, want) {
			t.Fatalf("generated optional support omits %q:\n%s", want, types)
		}
	}
}

func TestGenerateFailsClosedForReachableUnresolvedNode(t *testing.T) {
	m := manifest.Manifest{
		SchemaVersion: 2,
		Target:        manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168},
		Sources:       []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture-2168", Digest: "fixture:golden", MinecraftVersion: "fixture", ProtocolVersion: 2168}},
		Packets:       []manifest.Packet{{ID: 1, Name: "Blocked", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{{Ordinal: 0, Name: "Value", Encode: manifest.Unresolved("dynamic", true), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}},
	}
	if _, err := Generate(m, "wiregen"); err == nil || !strings.Contains(err.Error(), "unresolved") {
		t.Fatalf("Generate error = %v, want unresolved failure", err)
	}
}

func TestGenerateUsesCanonicalNamedTypesAndOrderedMapEntries(t *testing.T) {
	biome := manifest.Node{Kind: manifest.KindStruct, Semantic: "BiomeDefinitionData", TypeID: "BiomeDefinitionData", Fields: []manifest.Field{{Ordinal: 0, Name: "id", Encode: manifest.Primitive("u16le"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}
	mapping := manifest.Map(manifest.Primitive("var_u32"), manifest.Primitive("u16le"), biome)
	m := manifest.Manifest{SchemaVersion: 2, Target: manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:go", MinecraftVersion: "fixture", ProtocolVersion: 2168}}, Packets: []manifest.Packet{{ID: 122, Name: "BiomeDefinitionListPacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{{Ordinal: 0, Name: "Map of Biome names to data", Encode: mapping, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}}}

	files, err := Generate(m, "wiregen")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	source := generatedSource(files)
	if !strings.Contains(source, "type BiomeDefinitionData struct") || !strings.Contains(source, "[]protocol.OrderedEntry[uint16, protocol.BiomeDefinitionData]") {
		t.Fatalf("generated Go did not use canonical ordered map definitions:\n%s", source)
	}
}

func TestGenerateUsesTypedUnionInterface(t *testing.T) {
	union := manifest.Union(manifest.Primitive("var_u32"),
		manifest.Variant{Value: 0, Name: "SoundDataEvent::Stop", Encode: manifest.Node{Kind: manifest.KindStruct, Semantic: "SoundDataEvent::Stop", TypeID: "SoundDataEvent::Stop"}},
		manifest.Variant{Value: 1, Name: "SoundDataEvent::SetVolume", Encode: manifest.Node{Kind: manifest.KindStruct, Semantic: "SoundDataEvent::SetVolume", TypeID: "SoundDataEvent::SetVolume", Fields: []manifest.Field{{Ordinal: 0, Name: "Volume", Encode: manifest.Primitive("f32le"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}},
	)
	m := manifest.Manifest{SchemaVersion: 2, Target: manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:go", MinecraftVersion: "fixture", ProtocolVersion: 2168}}, Packets: []manifest.Packet{{ID: 348, Name: "SoundPacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{{Ordinal: 0, Name: "Event", Encode: union, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}}}
	files, err := Generate(m, "wiregen")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	source := generatedSource(files)
	if !strings.Contains(source, "type SoundDataEvent interface") || !strings.Contains(source, "func (*SoundDataEventSetVolume) tagSoundDataEvent() uint32 { return 1 }") || !strings.Contains(source, "Union(io, x, io.Varuint32, SoundDataEvent.tagSoundDataEvent, func(tag uint32) SoundDataEvent {") || strings.Contains(source, "Tag int64") {
		t.Fatalf("generated Go did not emit a typed union interface:\n%s", source)
	}
}

func TestGenerateSplitsPacketsAndSharedDefinitions(t *testing.T) {
	m := manifest.Manifest{SchemaVersion: 2, Target: manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:go", MinecraftVersion: "fixture", ProtocolVersion: 2168}}, Packets: []manifest.Packet{{ID: 1, Name: "LoginPacket", Direction: manifest.DirectionServerbound}}}
	files, err := Generate(m, "wiregen")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"protocol/packet/ids.go", "protocol/types.go", "protocol/packet/login.go"} {
		if _, ok := files[name]; !ok {
			t.Fatalf("generated files omit %s: %v", name, files)
		}
	}
	if !strings.Contains(files["protocol/types.go"], "package protocol") || !strings.Contains(files["protocol/packet/login.go"], "package packet") || !strings.Contains(files["protocol/packet/login.go"], "import (\n\t\"wiregen\"\n)") {
		t.Fatalf("generated packages were not separated or wired together:\n%s", files["protocol/packet/login.go"])
	}
	if _, ok := files["protocol/marshal.go"]; ok {
		t.Fatalf("generated output unexpectedly contains monolithic marshal.go")
	}
	if _, ok := files["protocol/enums.go"]; ok {
		t.Fatalf("generated output unexpectedly contains monolithic enums.go")
	}
	if strings.Contains(files["protocol/packet/login.go"], "LoginPacket") || !strings.Contains(files["protocol/packet/login.go"], "type Login struct") {
		t.Fatalf("packet name was not cleaned:\n%s", files["protocol/packet/login.go"])
	}
	if _, ok := files["wire.go"]; ok {
		t.Fatalf("definition-only output unexpectedly contains wire.go")
	}
}

func generatedSource(files map[string]string) string {
	var b strings.Builder
	for name, source := range files {
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		b.WriteString(source)
	}
	return b.String()
}

func TestPublicNamesDropSchemaScaffolding(t *testing.T) {
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

func TestGoNamesUseCommonInitialisms(t *testing.T) {
	for input, want := range map[string]string{
		"Container Id":   "ContainerID",
		"Json Uri Xz Ui": "JSONURIXZUI",
		"Identifier":     "Identifier",
	} {
		if got := exportName(input); got != want {
			t.Errorf("exportName(%q) = %q, want %q", input, got, want)
		}
	}
	if got := naming.EnumVariantName("VALUE_ID"); got != "ValueID" {
		t.Fatalf("enumVariantName initialism = %q, want ValueID", got)
	}
}

func TestEnumVariantNamesAreIdiomaticGo(t *testing.T) {
	for input, want := range map[string]string{
		"NONE":                    "None",
		"START_ATTACKING":         "StartAttacking",
		"FISHHOOK_FISHPOS":        "FishhookFishPosition",
		"FISHHOOK_HOOKTIME":       "FishhookHookTime",
		"DRAGON_START_DEATH_ANIM": "DragonStartDeathAnimation",
		"SILVERFISH_MERGE_ANIM":   "SilverfishMergeAnimation",
		"PRIMED_TNT":              "PrimedTNT",
		"PRIME_TNTCART":           "PrimeTNTCart",
		"Primed_Tnt":              "PrimedTNT",
		"OSX":                     "OSX",
		"UWP":                     "UWP",
	} {
		if got := naming.EnumVariantName(input); got != want {
			t.Fatalf("naming.EnumVariantName(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestGenerateMapsCanonicalSemanticsToNativeGoTypes(t *testing.T) {
	vec3 := manifest.Node{Kind: manifest.KindStruct, Semantic: "Vec3", TypeID: "Vec3", Fields: []manifest.Field{
		{Ordinal: 0, Name: "X", Encode: manifest.Primitive("f32le"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		{Ordinal: 1, Name: "Y", Encode: manifest.Primitive("f32le"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		{Ordinal: 2, Name: "Z", Encode: manifest.Primitive("f32le"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
	}}
	colour := manifest.Node{Kind: manifest.KindStruct, Semantic: "mce::Color", TypeID: "mce::Color", Fields: []manifest.Field{
		{Ordinal: 0, Name: "Color", Encode: manifest.Primitive("i32le"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
	}}
	runtimeID := manifest.Node{Kind: manifest.KindStruct, Semantic: "ActorRuntimeID", TypeID: "ActorRuntimeID", Fields: []manifest.Field{{Ordinal: 0, Name: "Value", Encode: manifest.Primitive("var_u64"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}
	uniqueID := manifest.Node{Kind: manifest.KindStruct, Semantic: "ActorUniqueID", TypeID: "ActorUniqueID", Fields: []manifest.Field{{Ordinal: 0, Name: "Value", Encode: manifest.Primitive("zigzag_i64"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}
	m := manifest.Manifest{SchemaVersion: 2, Target: manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168}, Sources: []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:native-go", MinecraftVersion: "fixture", ProtocolVersion: 2168}}, Packets: []manifest.Packet{{ID: 1, Name: "NativePacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{
		{Ordinal: 0, Name: "ID", Encode: manifest.Primitive("uuid"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		{Ordinal: 1, Name: "Position", Encode: vec3, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		{Ordinal: 2, Name: "Colour", Encode: colour, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		{Ordinal: 3, Name: "Data", Encode: manifest.NBT(manifest.NBTNetwork), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		{Ordinal: 4, Name: "Runtime", Encode: runtimeID, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		{Ordinal: 5, Name: "Unique", Encode: uniqueID, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
	}}}}
	files, err := Generate(m, "wiregen")
	if err != nil {
		t.Fatal(err)
	}
	source := files["protocol/packet/native.go"]
	for _, want := range []string{"uuid.UUID", "mgl32.Vec3", "color.RGBA", "[]byte", "Runtime  uint64", "Unique   int64", `"github.com/google/uuid"`, `"github.com/go-gl/mathgl/mgl32"`, `"image/color"`} {
		if !strings.Contains(source, want) {
			t.Fatalf("native Go output omits %q:\n%s", want, source)
		}
	}
	if strings.Contains(files["protocol/types.go"], "type Vec3 struct") || strings.Contains(files["protocol/types.go"], "type MceColor struct") {
		t.Fatalf("native Go types were redundantly regenerated:\n%s", files["protocol/types.go"])
	}
	if !strings.Contains(source, "io.ActorRuntimeID(&pk.Runtime)") || !strings.Contains(source, "io.ActorUniqueID(&pk.Unique)") {
		t.Fatalf("semantic ID IO methods were not used:\n%s", source)
	}
}

func TestNativeGoTypeRejectsMislabelledVector(t *testing.T) {
	node := manifest.Node{Kind: manifest.KindStruct, TypeID: "Vec3", Fields: []manifest.Field{{Encode: manifest.Primitive("f64le")}}}
	if _, matched, err := nativeGoType(node); !matched || err == nil {
		t.Fatalf("nativeGoType matched=%v error=%v, want a closed failure", matched, err)
	}
}

func TestSemanticIOCallUsesExactIdentifierCodecs(t *testing.T) {
	tests := []struct {
		typeID string
		code   string
		want   string
	}{
		{typeID: "ActorRuntimeID", code: "var_u64", want: "ActorRuntimeID"},
		{typeID: "ActorRuntimeID", code: "zigzag_i64", want: "ActorRuntimeIDVarint64"},
		{typeID: "ActorRuntimeID", code: "var_u32", want: "ActorRuntimeIDVaruint32"},
		{typeID: "ActorRuntimeID", code: "var_i64", want: "SignedVarint64"},
		{typeID: "ActorUniqueID", code: "zigzag_i64", want: "ActorUniqueID"},
		{typeID: "ActorUniqueID", code: "i64le", want: "ActorUniqueIDInt64"},
		{typeID: "ActorUniqueID", code: "u64le", want: "ActorUniqueIDUint64"},
		{typeID: "ActorUniqueID", code: "var_u64", want: "ActorUniqueIDVaruint64"},
		{typeID: "ActorUniqueID", code: "var_i64", want: "SignedVarint64"},
	}
	for _, test := range tests {
		t.Run(test.typeID+"/"+test.code, func(t *testing.T) {
			node := manifest.Node{Kind: manifest.KindStruct, TypeID: test.typeID, Fields: []manifest.Field{{Encode: manifest.Primitive(test.code)}}}
			if method, ok := semanticIOCall(node); !ok || method != test.want {
				t.Fatalf("semanticIOCall = (%q, %v), want (%q, true)", method, ok, test.want)
			}
		})
	}

	bad := manifest.Node{Kind: manifest.KindStruct, TypeID: "ActorUniqueID", Fields: []manifest.Field{{Encode: manifest.Primitive("f32le")}}}
	if method, ok := semanticIOCall(bad); ok {
		t.Fatalf("semanticIOCall returned %q for unsupported identifier layout", method)
	}
}

func TestGenerateUsesNamedRecursiveTypeAndBoundedBitset(t *testing.T) {
	dynamic := manifest.Union(manifest.Primitive("u8"),
		manifest.Variant{Value: 0, Name: "Empty", Encode: manifest.Void()},
		manifest.Variant{Value: 1, Name: "List", Encode: manifest.Array(manifest.Primitive("var_u32"), manifest.Recursive("cereal::DynamicValue"))},
	)
	dynamic.TypeID = "cereal::DynamicValue"
	m := manifest.Manifest{
		SchemaVersion: 2,
		Target:        manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 1},
		Sources:       []manifest.SourcePin{{ID: "fixture", Kind: "fixture", Revision: "1", Digest: "sha256:fixture", MinecraftVersion: "fixture", ProtocolVersion: 1}},
		Packets: []manifest.Packet{{ID: 1, Name: "RecursivePacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{
			{Ordinal: 0, Name: "Value", Encode: dynamic, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
			{Ordinal: 1, Name: "Flags", Encode: manifest.Bitset(131), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		}}},
	}
	files, err := Generate(m, "fixture")
	if err != nil {
		t.Fatal(err)
	}
	var all strings.Builder
	for _, source := range files {
		all.WriteString(source)
	}
	for _, want := range []string{"[]CerealDynamicValue", "type Bitset131 [3]uint64", "Flags protocol.Bitset131"} {
		if !strings.Contains(all.String(), want) {
			t.Fatalf("generated output missing %q:\n%s", want, all.String())
		}
	}
}

func TestGenerateEmitsPacketRuntimePoolsAndOptionalNativeProfile(t *testing.T) {
	packets := []manifest.Packet{
		{ID: 1, Name: "ServerEnvelope", Direction: manifest.DirectionClientbound},
		{ID: 2, Name: "ClientEnvelope", Direction: manifest.DirectionServerbound},
		{ID: 3, Name: "SharedEnvelope", Direction: manifest.DirectionBidirectional},
	}
	m := manifest.Manifest{
		SchemaVersion: 2,
		Target:        manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 1},
		Sources:       []manifest.SourcePin{{ID: "fixture", Kind: "fixture", Revision: "1", Digest: "sha256:fixture", MinecraftVersion: "fixture", ProtocolVersion: 1}},
		Packets:       packets,
	}
	files, err := Generate(m, "fixture")
	if err != nil {
		t.Fatal(err)
	}
	serverSource := ""
	for _, source := range files {
		if strings.Contains(source, "type ServerEnvelope struct") {
			serverSource = source
			break
		}
	}
	if !strings.Contains(files["protocol/packet/packet.go"], "type Packet interface") || !strings.Contains(serverSource, "func (*ServerEnvelope) ID() uint32 {\n\treturn IDServerEnvelope\n}") {
		t.Fatalf("packet runtime contract was not emitted:\n%s\n%s", files["protocol/packet/packet.go"], serverSource)
	}
	pool := files["protocol/packet/pool.go"]
	for _, want := range []string{"IDServerEnvelope: func() Packet", "IDClientEnvelope: func() Packet", "IDSharedEnvelope: func() Packet", "func NewClientPacket", "func NewServerPacket"} {
		if !strings.Contains(pool, want) {
			t.Fatalf("packet pools omit %q:\n%s", want, pool)
		}
	}
	withoutNative, err := GenerateWithOptions(m, Options{ProtocolImportPath: "fixture", EmitPacketRuntime: true, EmitPacketPools: true})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(withoutNative["protocol/packet/server_packet.go"], "uuid.UUID") {
		t.Fatal("native profile unexpectedly leaked into disabled output")
	}
}

func TestGenerateWithoutNativeTypesUsesFixedUUIDBytes(t *testing.T) {
	m := manifest.Manifest{
		SchemaVersion: 2,
		Target:        manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 1},
		Sources:       []manifest.SourcePin{{ID: "fixture", Kind: "fixture", Revision: "1", Digest: "sha256:fixture", MinecraftVersion: "fixture", ProtocolVersion: 1}},
		Packets: []manifest.Packet{{ID: 1, Name: "UUIDPacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{{
			Ordinal: 0, Name: "UUID", Encode: manifest.Primitive("uuid"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}},
		}}}},
	}
	files, err := GenerateWithOptions(m, Options{ProtocolImportPath: "fixture", EmitPacketRuntime: true, EmitPacketPools: true})
	if err != nil {
		t.Fatal(err)
	}
	packet := ""
	for _, source := range files {
		if strings.Contains(source, "type UUID struct") {
			packet = source
			break
		}
	}
	if !strings.Contains(packet, "UUID [16]byte") || !strings.Contains(packet, "io.UUIDBytes(&pk.UUID)") {
		t.Fatalf("disabled native profile did not emit fixed UUID bytes:\n%s", packet)
	}
}

func TestGenerateGoPreservesNBTEncodingOnIOCalls(t *testing.T) {
	m := manifest.Manifest{
		SchemaVersion: manifest.SchemaVersion,
		Target:        manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168},
		Sources:       []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "sha256:fixture"}},
		Packets: []manifest.Packet{{ID: 1, Name: "NBTPacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{
			{Ordinal: 0, Name: "Network", Encode: manifest.NBT(manifest.NBTNetwork), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
			{Ordinal: 1, Name: "Persistent", Encode: manifest.NBT(manifest.NBTPersistent), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		}}},
	}
	files, err := Generate(m, "example.com/protocol")
	if err != nil {
		t.Fatal(err)
	}
	packet := files["protocol/packet/nbt.go"]
	for _, want := range []string{"io.NBT(&pk.Network, protocol.NBTNetwork)", "io.NBT(&pk.Persistent, protocol.NBTPersistent)"} {
		if !strings.Contains(packet, want) {
			t.Fatalf("packet omitted %q:\n%s", want, packet)
		}
	}
	if !strings.Contains(files["protocol/codec.go"], "NBT(*[]byte, NBTEncoding)") {
		t.Fatalf("runtime IO did not expose format-aware NBT:\n%s", files["protocol/codec.go"])
	}
}

func TestGeneratePreservesEnumAliases(t *testing.T) {
	value := manifest.Enum("zigzag_i32",
		manifest.EnumValue{Name: "Undefined", Value: -1},
		manifest.EnumValue{Name: "Survival", Value: 0},
		manifest.EnumValue{Name: "WorldDefault", Value: 0},
	)
	value.Semantic, value.TypeID = "GameType", "enums/GameType"
	m := manifest.Manifest{
		SchemaVersion: 2,
		Target:        manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168},
		Sources:       []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:aliases", MinecraftVersion: "fixture", ProtocolVersion: 2168}},
		Packets:       []manifest.Packet{{ID: 1, Name: "AliasPacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{{Ordinal: 0, Name: "GameType", Encode: value, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}},
	}
	files, err := Generate(m, "wiregen")
	if err != nil {
		t.Fatal(err)
	}
	var source strings.Builder
	for _, contents := range files {
		source.WriteString(contents)
	}
	// Whitespace is formatting-dependent; check each declaration after tokenizing it.
	normalized := strings.Join(strings.Fields(source.String()), " ")
	for _, want := range []string{"GameTypeUndefined GameType = -1", "GameTypeSurvival GameType = 0", "GameTypeWorldDefault GameType = 0"} {
		if !strings.Contains(normalized, want) {
			t.Fatalf("generated Go omits alias declaration %q", want)
		}
	}
}

// A layout overlay moves an enum's constants next to the packet that uses them
// under the fork's names and renames fields, without changing wire code or the
// names of nested anonymous types.
func TestGenerateAppliesLayoutOverlay(t *testing.T) {
	kind := manifest.Enum("u8", manifest.EnumValue{Value: 0, Name: "Survival"}, manifest.EnumValue{Value: 1, Name: "Creative"})
	kind.TypeID = "enums/GameType"
	entry := manifest.Struct(manifest.Field{Ordinal: 0, Name: "Value", Encode: manifest.Primitive("u8"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}})
	m := manifest.Manifest{
		SchemaVersion: 2,
		Target:        manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 2168},
		Sources:       []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "fixture", Digest: "fixture:layout", MinecraftVersion: "fixture", ProtocolVersion: 2168}},
		Packets: []manifest.Packet{{ID: 1, Name: "SetPlayerGameTypePacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{
			{Ordinal: 0, Name: "Game Type", Encode: kind, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
			{Ordinal: 1, Name: "Entries", Encode: manifest.Array(manifest.Primitive("var_u32"), entry), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		}}},
	}
	overlay := layout.Overlay{
		Constants: map[string]layout.Placement{"enums/GameType": {Package: "packet", File: "set_player_game_type", Names: map[string]string{"Creative": "GameTypeCreative"}}},
		Fields:    map[string]string{layout.FieldKey("SetPlayerGameTypePacket", "Game Type"): "PlayerGameMode", layout.FieldKey("SetPlayerGameTypePacket", "Entries"): "Rows"},
		Files:     map[string]string{"SetPlayerGameTypePacket": "set_player_game_type", "enums/GameType": "game_mode"},
		Types:     map[string]string{"SetPlayerGameTypePacket": "SetPlayerGameMode"},
	}
	files, err := GenerateWithOptions(m, Options{ProtocolImportPath: "wiregen", NativeTypes: true, EmitPacketRuntime: true, EmitPacketPools: true, Layout: overlay})
	if err != nil {
		t.Fatal(err)
	}
	packet := files["protocol/packet/set_player_game_type.go"]
	for _, want := range []string{
		"type SetPlayerGameMode struct",
		"func (*SetPlayerGameMode) ID() uint32",
		"IDSetPlayerGameMode",
		"PlayerGameMode protocol.GameType",
		"pk.PlayerGameMode.Marshal(io)",
		"[]protocol.SetPlayerGameModeEntriesItemStruct", // anonymous types are named from the wire field, not the layout name
		"GameTypeCreative protocol.GameType = 1",
		"GameTypeSurvival protocol.GameType = 0",
	} {
		if !strings.Contains(packet, want) {
			t.Fatalf("layout output omits %q:\n%s", want, packet)
		}
	}
	if strings.Contains(generatedSource(files), "GameTypeCreative GameType = 1") {
		t.Fatal("relocated constants were also emitted in the protocol package")
	}
	if strings.Contains(packet, "GameTypeGameTypeCreative") {
		t.Fatal("reviewed constant name was prefixed with the enum name")
	}
	if strings.Index(packet, "const (") > strings.Index(packet, "type SetPlayerGameMode struct") {
		t.Fatalf("relocated constants are not above the packet:\n%s", packet)
	}
	if !strings.Contains(files["protocol/game_mode.go"], "type GameType uint8") {
		t.Fatalf("enum was not placed in its reviewed file: %v", sortedKeys(files))
	}
}

func sortedKeys(files map[string]string) []string {
	keys := make([]string, 0, len(files))
	for key := range files {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// A packet whose only field is a single-use payload struct takes that struct's
// fields, with docs and reviewed names keyed by the declaring struct; a
// sub-struct of a multi-field packet keeps its name.
func TestGenerateInlinesSingleUsePayloadStructs(t *testing.T) {
	payload := manifest.Node{Kind: manifest.KindStruct, TypeID: "MoveActorAbsoluteData", Fields: []manifest.Field{
		{Ordinal: 0, Name: "Actor Runtime ID", Encode: manifest.Primitive("var_u64"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		{Ordinal: 1, Name: "Flags", Encode: manifest.Primitive("u8"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
	}}
	shared := manifest.Node{Kind: manifest.KindStruct, TypeID: "Shared", Fields: []manifest.Field{{Ordinal: 0, Name: "Value", Encode: manifest.Primitive("u8"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}}}
	m := manifest.Manifest{
		SchemaVersion: 2,
		Target:        manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 1},
		Sources:       []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "1", Digest: "fixture", MinecraftVersion: "fixture", ProtocolVersion: 1}},
		Packets: []manifest.Packet{
			{ID: 1, Name: "MoveActorAbsolutePacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{
				{Ordinal: 0, Name: "Move Data", Encode: payload, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
			}},
			{ID: 2, Name: "OtherPacket", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{
				{Ordinal: 0, Name: "Extra", Encode: shared, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
				{Ordinal: 1, Name: "Count", Encode: manifest.Primitive("u8"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
			}},
		},
	}
	files, err := GenerateWithOptions(m, Options{ProtocolImportPath: "wiregen", NativeTypes: false, EmitPacketRuntime: true, EmitPacketPools: true,
		Docs:   docs.Overlay{Fields: map[string]string{docs.FieldKey("MoveActorAbsoluteData", "Flags"): "Flags docs."}},
		Layout: layout.Overlay{Fields: map[string]string{layout.FieldKey("MoveActorAbsoluteData", "Actor Runtime ID"): "EntityRuntimeID"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	packet := files["protocol/packet/move_actor_absolute.go"]
	for _, want := range []string{"EntityRuntimeID uint64", "// Flags docs.", "Flags uint8", "io.Varuint64(&pk.EntityRuntimeID)"} {
		if !strings.Contains(packet, want) {
			t.Fatalf("inlined packet omits %q:\n%s", want, packet)
		}
	}
	if strings.Contains(generatedSource(files), "type MoveActorAbsoluteData struct") || strings.Contains(packet, "MoveData") {
		t.Fatalf("single-use payload struct was still emitted:\n%s", packet)
	}
	other := files["protocol/packet/other.go"]
	if !strings.Contains(other, "Extra protocol.Shared") || !strings.Contains(generatedSource(files), "type Shared struct") {
		t.Fatalf("sub-struct of a multi-field packet was inlined:\n%s", other)
	}
}

// A plain integer marked as an actor identifier uses the identifier IO
// operation of its wire width and keeps its Go type.
func TestGenerateAppliesSemanticsOverlay(t *testing.T) {
	m := manifest.Manifest{
		SchemaVersion: 2,
		Target:        manifest.Target{MinecraftVersion: "fixture", ProtocolVersion: 1},
		Sources:       []manifest.SourcePin{{ID: "fixture", Kind: "synthetic", Revision: "1", Digest: "fixture", MinecraftVersion: "fixture", ProtocolVersion: 1}},
		Packets: []manifest.Packet{{ID: 1, Name: "ActorPickRequestPacket", Direction: manifest.DirectionServerbound, Fields: []manifest.Field{
			{Ordinal: 0, Name: "Actor ID", Encode: manifest.Primitive("i64le"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
			{Ordinal: 1, Name: "Target", Encode: manifest.Primitive("var_u64"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
			{Ordinal: 2, Name: "Slot", Encode: manifest.Primitive("u8"), Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}},
		}}},
	}
	overlay := semantics.Overlay{Fields: map[string]string{
		semantics.FieldKey("ActorPickRequestPacket", "Actor ID"): semantics.ActorUniqueID,
		semantics.FieldKey("ActorPickRequestPacket", "Target"):   semantics.ActorRuntimeID,
	}}
	files, err := GenerateWithOptions(m, Options{ProtocolImportPath: "wiregen", NativeTypes: true, EmitPacketRuntime: true, EmitPacketPools: true, Semantics: overlay})
	if err != nil {
		t.Fatal(err)
	}
	packet := files["protocol/packet/actor_pick_request.go"]
	for _, want := range []string{`ActorID\s+int64`, `Target\s+uint64`, `io\.ActorUniqueIDInt64\(&pk\.ActorID\)`, `io\.ActorRuntimeID\(&pk\.Target\)`, `io\.Uint8\(&pk\.Slot\)`} {
		if !regexp.MustCompile(want).MatchString(packet) {
			t.Fatalf("semantics output omits %q:\n%s", want, packet)
		}
	}
}

// An enum type used at sites that each allow a subset keeps every site's values.
func TestGenerateMergesSharedEnumValuesAcrossSites(t *testing.T) {
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
	files, err := Generate(m, "wiregen")
	if err != nil {
		t.Fatal(err)
	}
	normalized := strings.Join(strings.Fields(generatedSource(files)), " ")
	for _, want := range []string{"TextPacketType = 0", "TextPacketType = 1", "TextPacketType = 5"} {
		if !strings.Contains(normalized, want) {
			t.Fatalf("generated Go omits shared enum value %q", want)
		}
	}
}
