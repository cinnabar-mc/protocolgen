package source

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"path"
	"strings"
)

// BehaviorEntities reads official entity definitions from the already locked
// sample archive. The complete archive is authenticated before any JSON is used;
// no extracted cache tree or additional source revision is trusted.
func (s *Sources) BehaviorEntities(ctx context.Context, cache string) (map[string][]byte, error) {
	if _, ok := s.Lock.Inputs["resource_pack"]; !ok {
		return nil, fmt.Errorf("source lock has no sample archive")
	}
	archive, err := s.Read(ctx, "resource_pack", cache)
	if err != nil {
		return nil, err
	}
	return readBehaviorEntities(bytes.NewReader(archive))
}

// readBehaviorEntities selects regular, immediate entity JSON entries without
// extracting source materials into the repository or following archive links.
func readBehaviorEntities(input io.Reader) (map[string][]byte, error) {
	gz, err := gzip.NewReader(input)
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	reader := tar.NewReader(gz)
	files := make(map[string][]byte)
	var total int64
	for {
		header, err := reader.Next()
		if err == io.EOF {
			if len(files) == 0 {
				return nil, fmt.Errorf("sample archive has no behavior_pack/entities JSON")
			}
			return files, nil
		}
		if err != nil {
			return nil, err
		}
		_, relative, _ := strings.Cut(header.Name, "/")
		if !strings.HasPrefix(relative, "behavior_pack/entities/") || header.Typeflag == tar.TypeDir {
			continue
		}
		if path.Clean(relative) != relative || strings.Contains(relative, "\\") || path.Dir(relative) != "behavior_pack/entities" || !strings.HasSuffix(relative, ".json") || header.Typeflag != tar.TypeReg {
			return nil, fmt.Errorf("unsupported behavior entity entry %q", header.Name)
		}
		if _, exists := files[relative]; exists {
			return nil, fmt.Errorf("duplicate behavior entity %q", relative)
		}
		const limit = 32 << 20
		if header.Size < 0 || header.Size > limit-total {
			return nil, fmt.Errorf("behavior entities exceed the expanded size limit")
		}
		total += header.Size
		content, err := io.ReadAll(reader)
		if err != nil {
			return nil, err
		}
		files[relative] = content
	}
}
