package store

import (
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"
	"encoding/json"
	"os"
	"path/filepath"
)

type snapshot struct {
	Case           domain.AcceptanceCase    `json:"case"`
	Content        *domain.DocumentContent  `json:"content,omitempty"`
	ContentHistory []domain.DocumentContent `json:"content_history,omitempty"`
}

func readSnapshot(path string) (*snapshot, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s snapshot
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func writeSnapshot(path string, s snapshot) error {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".snapshot-*.tmp")
	if err != nil {
		return err
	}
	name := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(name)
		}
	}()
	if _, err = tmp.Write(b); err == nil {
		err = tmp.Sync()
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err = os.Rename(name, path); err != nil {
		return err
	}
	cleanup = false
	if d, openErr := os.Open(dir); openErr == nil {
		_ = d.Sync()
		_ = d.Close()
	}
	return nil
}

func restoreSnapshot(path string, s *snapshot) {
	if s == nil {
		_ = os.Remove(path)
		return
	}
	_ = writeSnapshot(path, *s)
}
