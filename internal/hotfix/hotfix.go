// Package hotfix derives a narrowly fingerprinted same-protocol hotfix from a
// fully reconciled canonical manifest.
package hotfix

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"protocolgen/internal/manifest"
)

const SchemaVersion uint32 = 1

type Spec struct {
	SchemaVersion      uint32               `json:"schema_version"`
	Target             manifest.Target      `json:"target"`
	BaseManifestSHA256 string               `json:"base_manifest_sha256"`
	Sources            []manifest.SourcePin `json:"sources"`
	Operations         []Operation          `json:"operations"`
}

type Operation struct {
	ID                 string              `json:"id"`
	PacketID           uint32              `json:"packet_id"`
	FieldOrdinal       int                 `json:"field_ordinal"`
	Path               string              `json:"path"`
	Operation          string              `json:"operation"`
	PrePatchNodeSHA256 string              `json:"pre_patch_node_sha256"`
	Evidence           []manifest.Evidence `json:"evidence"`
	Reason             string              `json:"reason"`
}

func LoadSpec(path string) (Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Spec{}, fmt.Errorf("read hotfix spec: %w", err)
	}
	var spec Spec
	if err := json.Unmarshal(data, &spec); err != nil {
		return Spec{}, fmt.Errorf("parse hotfix spec: %w", err)
	}
	return spec, nil
}

func Apply(base manifest.Manifest, baseBytes []byte, spec Spec) (manifest.Manifest, error) {
	if spec.SchemaVersion != SchemaVersion {
		return manifest.Manifest{}, fmt.Errorf("hotfix schema_version %d is not v%d", spec.SchemaVersion, SchemaVersion)
	}
	if got := bytesDigest(baseBytes); got != spec.BaseManifestSHA256 {
		return manifest.Manifest{}, fmt.Errorf("base manifest fingerprint is %s, spec requires %s", got, spec.BaseManifestSHA256)
	}
	var authoritative manifest.Manifest
	if err := json.Unmarshal(baseBytes, &authoritative); err != nil {
		return manifest.Manifest{}, fmt.Errorf("parse base manifest: %w", err)
	}
	base = authoritative
	if base.Derivation != nil {
		return manifest.Manifest{}, fmt.Errorf("cannot derive a hotfix from an already derived manifest")
	}
	if spec.Target.MinecraftVersion == "" || spec.Target.ProtocolVersion != base.Target.ProtocolVersion || spec.Target.MinecraftVersion == base.Target.MinecraftVersion {
		return manifest.Manifest{}, fmt.Errorf("hotfix target must change Minecraft version while retaining protocol %d", base.Target.ProtocolVersion)
	}
	if len(spec.Sources) == 0 || len(spec.Operations) == 0 {
		return manifest.Manifest{}, fmt.Errorf("hotfix must pin target evidence sources and at least one operation")
	}
	for _, source := range spec.Sources {
		if source.MinecraftVersion != spec.Target.MinecraftVersion || source.ProtocolVersion != spec.Target.ProtocolVersion {
			return manifest.Manifest{}, fmt.Errorf("hotfix evidence source %q does not pin target Minecraft %s / protocol %d", source.ID, spec.Target.MinecraftVersion, spec.Target.ProtocolVersion)
		}
	}

	result := authoritative
	if err := manifest.Validate(result); err != nil {
		return manifest.Manifest{}, err
	}
	result.Target = spec.Target
	result.Sources = append(result.Sources, spec.Sources...)
	proof := &manifest.DerivationProof{BaseTarget: base.Target, BaseManifestSHA256: spec.BaseManifestSHA256}
	seen := map[string]bool{}
	for _, operation := range spec.Operations {
		if operation.ID == "" || seen[operation.ID] || operation.PacketID == 0 || operation.FieldOrdinal < 0 || operation.Path == "" || operation.Reason == "" || len(operation.Evidence) == 0 {
			return manifest.Manifest{}, fmt.Errorf("hotfix operation %q is incomplete or duplicated", operation.ID)
		}
		seen[operation.ID] = true
		node, err := findNode(&result, operation.PacketID, operation.FieldOrdinal, operation.Path)
		if err != nil {
			return manifest.Manifest{}, fmt.Errorf("hotfix operation %q: %w", operation.ID, err)
		}
		before, err := nodeDigest(*node)
		if err != nil {
			return manifest.Manifest{}, err
		}
		if before != operation.PrePatchNodeSHA256 {
			return manifest.Manifest{}, fmt.Errorf("hotfix operation %q is stale: node is %s, spec requires %s", operation.ID, before, operation.PrePatchNodeSHA256)
		}
		switch operation.Operation {
		case "wrap_optional":
			if node.Kind != manifest.KindOptional {
				return manifest.Manifest{}, fmt.Errorf("hotfix operation %q can only wrap an optional node, got %s", operation.ID, node.Kind)
			}
			*node = manifest.Optional(*node)
		default:
			return manifest.Manifest{}, fmt.Errorf("hotfix operation %q uses unsupported operation %q", operation.ID, operation.Operation)
		}
		after, err := nodeDigest(*node)
		if err != nil {
			return manifest.Manifest{}, err
		}
		proof.Operations = append(proof.Operations, manifest.DerivationOperationProof{
			ID: operation.ID, PacketID: operation.PacketID, FieldOrdinal: operation.FieldOrdinal,
			Path: operation.Path, Operation: operation.Operation, PrePatchNodeSHA256: before,
			PostPatchNodeSHA256: after, Evidence: operation.Evidence, Reason: operation.Reason,
		})
	}
	result.Derivation = proof
	if err := manifest.Validate(result); err != nil {
		return manifest.Manifest{}, fmt.Errorf("validate derived manifest: %w", err)
	}
	return result, nil
}

func findNode(value *manifest.Manifest, packetID uint32, fieldOrdinal int, path string) (*manifest.Node, error) {
	var node *manifest.Node
	for packetIndex := range value.Packets {
		if value.Packets[packetIndex].ID != packetID {
			continue
		}
		for fieldIndex := range value.Packets[packetIndex].Fields {
			field := &value.Packets[packetIndex].Fields[fieldIndex]
			if field.Ordinal == fieldOrdinal {
				node = &field.Encode
				break
			}
		}
	}
	if node == nil {
		return nil, fmt.Errorf("packet %d field ordinal %d is missing", packetID, fieldOrdinal)
	}
	parts := strings.Split(path, ".")
	if len(parts) == 0 || parts[0] != "encode" {
		return nil, fmt.Errorf("path %q must start with encode", path)
	}
	for _, part := range parts[1:] {
		switch {
		case part == "encode":
			continue
		case part == "value":
			node = node.Value
		case part == "element":
			node = node.Element
		case part == "control":
			node = node.Control
		case strings.HasPrefix(part, "fields["):
			index, err := indexed(part, "fields")
			if err != nil || index >= len(node.Fields) {
				return nil, fmt.Errorf("path %q has invalid field segment %q", path, part)
			}
			node = &node.Fields[index].Encode
		case strings.HasPrefix(part, "variants["):
			index, err := indexed(part, "variants")
			if err != nil || index >= len(node.Variants) {
				return nil, fmt.Errorf("path %q has invalid variant segment %q", path, part)
			}
			node = &node.Variants[index].Encode
		default:
			return nil, fmt.Errorf("path %q has unsupported segment %q", path, part)
		}
		if node == nil {
			return nil, fmt.Errorf("path %q traverses a missing node", path)
		}
	}
	return node, nil
}

func indexed(part, prefix string) (int, error) {
	value := strings.TrimSuffix(strings.TrimPrefix(part, prefix+"["), "]")
	index, err := strconv.Atoi(value)
	if err != nil || index < 0 {
		return 0, fmt.Errorf("invalid index")
	}
	return index, nil
}

func nodeDigest(node manifest.Node) (string, error) {
	data, err := json.Marshal(node)
	if err != nil {
		return "", err
	}
	return bytesDigest(data), nil
}

func bytesDigest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
