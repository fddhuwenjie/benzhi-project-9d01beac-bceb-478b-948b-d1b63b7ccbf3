package store

import (
	"encoding/json"
	"errors"
	"os"
)

type idempotencyRecord struct {
	CaseID      string          `json:"case_id"`
	RequestID   string          `json:"request_id"`
	Fingerprint string          `json:"fingerprint"`
	Result      json.RawMessage `json:"result"`
}

func loadIdempotency(path string) (map[string]idempotencyRecord, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return make(map[string]idempotencyRecord), nil
	}
	if err != nil {
		return nil, err
	}
	var records []idempotencyRecord
	if err := json.Unmarshal(b, &records); err != nil {
		return nil, err
	}
	out := make(map[string]idempotencyRecord, len(records))
	for _, record := range records {
		out[record.CaseID+"\x00"+record.RequestID] = record
	}
	return out, nil
}

func writeIdempotency(path string, records map[string]idempotencyRecord) error {
	all := make([]idempotencyRecord, 0, len(records))
	for _, record := range records {
		all = append(all, record)
	}
	b, err := json.MarshalIndent(all, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
