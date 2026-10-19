package source

import (
	"encoding/json"
	"fmt"
	"slices"
)

// Capture describes the build arguments for an available capture adapter.
type Capture struct {
	GoArgs []string `json:"go_args"`
}

// Release connects a catalog label to its channel, wire protocol and codec snapshot.
type Release struct {
	MinecraftVersion string   `json:"minecraft_version"`
	ProtocolVersion  int      `json:"protocol_version"`
	Channel          string   `json:"channel"`
	Aliases          []string `json:"aliases"`
	Snapshot         string   `json:"snapshot"`
	Capture          *Capture `json:"capture,omitempty"`
}

// ReleaseIndex declares supported snapshots and the defaults used by tooling.
type ReleaseIndex struct {
	SchemaVersion   int                `json:"schema_version"`
	DefaultProtocol string             `json:"default_protocol"`
	DefaultCapture  string             `json:"default_capture"`
	Releases        map[string]Release `json:"releases"`
}

// Releases loads the shared release registry, including explicit preview aliases.
func Releases() (ReleaseIndex, error) {
	var index ReleaseIndex
	contents, err := manifests.ReadFile("releases.json")
	if err != nil {
		return index, err
	}
	if err := json.Unmarshal(contents, &index); err != nil {
		return index, err
	}
	if index.SchemaVersion != 1 {
		return index, fmt.Errorf("unsupported release index schema")
	}
	for id, release := range index.Releases {
		if id != release.Snapshot || release.ProtocolVersion <= 0 || !slices.Contains(release.Aliases, release.MinecraftVersion) || (release.Channel != "retail" && release.Channel != "preview") {
			return index, fmt.Errorf("incomplete release %q", id)
		}
	}
	if _, ok := index.Releases[index.DefaultProtocol]; !ok {
		return index, fmt.Errorf("unknown default protocol snapshot")
	}
	if index.Releases[index.DefaultCapture].Capture == nil {
		return index, fmt.Errorf("default capture has no adapter")
	}
	return index, nil
}
