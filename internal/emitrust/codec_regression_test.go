package emitrust

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"protocolgen/internal/emitter"
	"protocolgen/internal/manifest"
)

// regressionField gives generated compile fixtures valid field metadata.
func regressionField(ordinal int, name string, node manifest.Node) manifest.Field {
	return manifest.Field{Ordinal: ordinal, Name: name, Encode: node, Symmetry: manifest.Symmetric, Provenance: manifest.Provenance{Pins: []string{"fixture"}}}
}

func TestGeneratedCodecRegressions(t *testing.T) {
	cargo, err := exec.LookPath("cargo")
	if err != nil {
		t.Skip("cargo is not installed")
	}
	directory := t.TempDir()
	for _, contract := range []int{0, CodecContractVersion} {
		t.Run(fmt.Sprintf("profile-%d", contract), func(t *testing.T) {
			m := collectionFixture(manifest.Primitive("u8"))
			m.Packets[0].Name = "FixturePacket"
			shared := manifest.Struct(regressionField(0, "Reader", manifest.Primitive("u8")), regressionField(1, "Value", manifest.Primitive("u8")))
			shared.TypeID = "Shared"
			payload := manifest.Struct(regressionField(0, "Writer", manifest.Primitive("u8")), regressionField(1, "Reader", manifest.Primitive("u8")), regressionField(2, "Value", manifest.Primitive("u8")))
			choice := manifest.Union(manifest.Primitive("u8"), manifest.Variant{Value: 0, Name: "Data", Encode: payload})
			m.Packets[0].Fields = []manifest.Field{
				regressionField(0, "Reader", manifest.Primitive("u8")), regressionField(1, "Value", manifest.Primitive("u8")),
				regressionField(2, "Shared", shared), regressionField(3, "Choice", choice),
				regressionField(4, "SmallBits", manifest.Bitset(2048)), regressionField(5, "LargeBits", manifest.Bitset(2049)),
			}
			var tests strings.Builder
			tests.WriteString("\n#[cfg(test)]\nmod codec_regressions {\nuse super::{packets::*, types::*, wire::{Decode, Encode}};\n#[test]\nfn bindings_and_large_bitsets_round_trip() {\nlet value = Fixture::default();\nassert!(Bitset2049::default().0.iter().all(|word| *word == 0));\nassert!(value.small_bits.0.iter().all(|word| *word == 0));\nassert!(value.large_bits.0.iter().all(|word| *word == 0));\n")
			encoded := "value.encode_to_vec()"
			if contract > 0 {
				encoded += ".unwrap()"
			}
			fmt.Fprintf(&tests, "let data = %s;\nassert_eq!(Fixture::decode_exact(&data).unwrap(), value);\n}\n", encoded)
			for _, code := range []string{"var_u32", "u32le"} {
				prefix := manifest.Primitive(code)
				prefix.Constraints = &manifest.Constraints{Minimum: pointerTo(json.Number("2e0")), Maximum: pointerTo(json.Number("6.4e1"))}
				for _, kind := range []string{"String", "Bytes", "ByteArray", "Array", "Map", "NestedString"} {
					name := typeName(code) + kind
					node := manifest.String(prefix)
					value, width := `"x".repeat(length).into()`, 1
					switch kind {
					case "Bytes":
						node = manifest.Bytes(prefix)
						value = "vec![1u8; length].into()"
					case "ByteArray":
						node = manifest.Array(prefix, manifest.Primitive("u8"))
						value = "vec![1u8; length].into()"
						if code == "u32le" {
							value = "vec![super::wire::U8(1); length]"
						}
					case "NestedString":
						node = manifest.Array(manifest.Primitive("var_u32"), manifest.String(prefix))
						value = `vec!["x".repeat(length).into()]`
					case "Array":
						node = manifest.Array(prefix, manifest.Primitive("u16le"))
						value, width = "vec![super::wire::U16LE(1); length]", 2
					case "Map":
						node = manifest.Map(prefix, manifest.Primitive("u8"), manifest.Primitive("u16le"))
						value, width = "vec![(super::wire::U8(1), super::wire::U16LE(1)); length]", 3
					}
					m.Packets = append(m.Packets, manifest.Packet{ID: uint32(len(m.Packets) + 10), Name: name + "Packet", Direction: manifest.DirectionClientbound, Fields: []manifest.Field{regressionField(0, "Value", node)}})
					fmt.Fprintf(&tests, "#[test]\nfn prefix_%s_%s() {\nfor length in [0usize, 1, 2, 64, 65] {\nlet value = %s { value: %s };\nlet encoded = std::panic::catch_unwind(|| value.encode_to_vec());\nlet valid = (2..=64).contains(&length);\n", code, strings.ToLower(kind), name, value)
					if contract > 0 {
						tests.WriteString("assert_eq!(encoded.as_ref().is_ok_and(|result| result.is_ok()), valid);\n")
					} else {
						tests.WriteString("assert_eq!(encoded.is_ok(), valid);\n")
					}
					if code == "var_u32" {
						tests.WriteString("let mut data = vec![length as u8];\n")
					} else {
						tests.WriteString("let mut data = (length as u32).to_le_bytes().to_vec();\n")
					}
					if kind == "NestedString" {
						tests.WriteString("data.insert(0, 1);\n")
					}
					fmt.Fprintf(&tests, "data.resize(data.len()+length*%d, 1);\nassert_eq!(%s::decode_exact(&data).is_ok(), valid);\n}\n}\n", width, name)
				}
			}
			tests.WriteString("}\n")
			files, err := GenerateFilesWithOptions(m, Options{CodecContract: contract})
			if err != nil {
				t.Fatal(err)
			}
			files["src/lib.rs"] += tests.String()
			path := filepath.Join(directory, fmt.Sprintf("profile-%d", contract))
			if err := emitter.WriteFiles(path, files); err != nil {
				t.Fatal(err)
			}
			command := exec.Command(cargo, "test", "--manifest-path", filepath.Join(path, "Cargo.toml"), "--all-features", "--quiet")
			command.Env = append(os.Environ(), "CARGO_TARGET_DIR="+filepath.Join(directory, "target"))
			if output, err := command.CombinedOutput(); err != nil {
				t.Fatalf("generated codec tests: %v\n%s", err, output)
			}
		})
	}
}
