package storage

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
)

// JSONStore persists sessions as individual JSON files in a directory.
type JSONStore struct {
	Directory string
}

func (s *JSONStore) Save(_ context.Context, clientID string, snapshot map[string]any) error {
	if err := os.MkdirAll(s.Directory, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(s.Directory, clientID+".json")
	return os.WriteFile(path, b, 0o644)
}

func (s *JSONStore) Load(_ context.Context, clientID string) (map[string]any, error) {
	path := filepath.Join(s.Directory, clientID+".json")
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *JSONStore) Delete(_ context.Context, clientID string) error {
	path := filepath.Join(s.Directory, clientID+".json")
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (s *JSONStore) List(_ context.Context) ([]string, error) {
	entries, err := os.ReadDir(s.Directory)
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
			ids = append(ids, e.Name()[:len(e.Name())-5])
		}
	}
	return ids, nil
}

// MultiJSONStore stores all accounts in a single JSON file.
type MultiJSONStore struct {
	Path string
}

type multiJSONData struct {
	Sessions map[string]map[string]any `json:"sessions"`
}

func (s *MultiJSONStore) load() (*multiJSONData, error) {
	b, err := os.ReadFile(s.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return &multiJSONData{Sessions: make(map[string]map[string]any)}, nil
		}
		return nil, err
	}
	var data multiJSONData
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, err
	}
	if data.Sessions == nil {
		data.Sessions = make(map[string]map[string]any)
	}
	return &data, nil
}

func (s *MultiJSONStore) save(data *multiJSONData) error {
	if dir := filepath.Dir(s.Path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.Path, b, 0o644)
}

func (s *MultiJSONStore) Save(_ context.Context, clientID string, snapshot map[string]any) error {
	data, err := s.load()
	if err != nil {
		return err
	}
	data.Sessions[clientID] = snapshot
	return s.save(data)
}

func (s *MultiJSONStore) Load(_ context.Context, clientID string) (map[string]any, error) {
	data, err := s.load()
	if err != nil {
		return nil, err
	}
	snap, ok := data.Sessions[clientID]
	if !ok {
		return nil, nil
	}
	return snap, nil
}

func (s *MultiJSONStore) Delete(_ context.Context, clientID string) error {
	data, err := s.load()
	if err != nil {
		return err
	}
	delete(data.Sessions, clientID)
	return s.save(data)
}

func (s *MultiJSONStore) List(_ context.Context) ([]string, error) {
	data, err := s.load()
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(data.Sessions))
	for id := range data.Sessions {
		ids = append(ids, id)
	}
	return ids, nil
}

func init() {
	var _ Store = (*JSONStore)(nil)
	var _ Lister = (*JSONStore)(nil)
	var _ Store = (*MultiJSONStore)(nil)
	var _ Lister = (*MultiJSONStore)(nil)
}
