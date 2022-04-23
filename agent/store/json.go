package store

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/menmos/menmos-agent/fs"
)

type JSONStateStore struct {
	path string
}

func NewJSONStore(path string) (*JSONStateStore, error) {
	if err := fs.EnsureDirExists(path); err != nil {
		return nil, err
	}
	return &JSONStateStore{path: path}, nil
}

func (s *JSONStateStore) getFilePath(id string) string {
	return filepath.Join(s.path, id) + ".json"
}

func (s *JSONStateStore) Get(id string) (*NodeInfo, error) {
	path := s.getFilePath(id)

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var info NodeInfo
	if err := json.Unmarshal(b, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

func (s *JSONStateStore) Save(info NodeInfo) error {
	path := s.getFilePath(info.ID)

	b, err := json.Marshal(info)
	if err != nil {
		return err
	}

	return os.WriteFile(path, b, 0666)
}

func (s *JSONStateStore) ListNodes() ([]NodeInfo, error) {
	entries, err := os.ReadDir(s.path)
	if err != nil {
		return nil, err
	}

	var nodes []NodeInfo
	for i := 0; i < len(entries); i++ {
		entry := entries[i]
		if entry.IsDir() {
			continue
		}

		path := filepath.Join(s.path, entry.Name())

		b, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		var info NodeInfo
		if err := json.Unmarshal(b, &info); err != nil {
			return nil, err
		}

		nodes = append(nodes, info)
	}

	return nodes, nil
}

func (s *JSONStateStore) DeleteNode(id string) error {
	path := s.getFilePath(id)
	return os.Remove(path)
}
