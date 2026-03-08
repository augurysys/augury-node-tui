package developerdownloads

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

const indexFilename = "developer-downloads/index.json"

type SourceState string

const (
	SourceBuilt    SourceState = "built"
	SourceHydrated SourceState = "hydrated"
	SourceMissing  SourceState = "missing"
)

type platformEntry struct {
	Name        string `json:"name"`
	Enabled     bool   `json:"enabled"`
	Source      string `json:"source"`
	SourceSHA   string `json:"source_sha"`
	SourceBranch string `json:"source_branch"`
}

type indexJSON struct {
	Platforms []platformEntry `json:"platforms"`
}

type Index struct {
	sources map[string]SourceState
}

func ReadAt(root string) (*Index, error) {
	path := filepath.Join(root, indexFilename)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var raw indexJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	idx := &Index{sources: make(map[string]SourceState)}
	for _, p := range raw.Platforms {
		if strings.TrimSpace(p.Name) == "" {
			continue
		}
		switch p.Source {
		case "built":
			idx.sources[p.Name] = SourceBuilt
		case "hydrated":
			idx.sources[p.Name] = SourceHydrated
		case "missing":
			idx.sources[p.Name] = SourceMissing
		default:
			idx.sources[p.Name] = SourceMissing
		}
	}
	return idx, nil
}

func (idx *Index) SourceState(platformID string) SourceState {
	if idx == nil {
		return ""
	}
	if s, ok := idx.sources[platformID]; ok {
		return s
	}
	return SourceMissing
}
