package store

import (
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

type DiskRepository struct {
	mu              sync.RWMutex
	dir             string
	eventsPath      string
	idempotencyPath string
	records         map[string]idempotencyRecord
}

func NewDiskRepository(dir string) (*DiskRepository, error) {
	if err := os.MkdirAll(filepath.Join(dir, "cases"), 0700); err != nil {
		return nil, err
	}
	r := &DiskRepository{dir: dir, eventsPath: filepath.Join(dir, "events.jsonl"), idempotencyPath: filepath.Join(dir, "idempotency.json")}
	records, err := loadIdempotency(r.idempotencyPath)
	if err != nil {
		return nil, err
	}
	r.records = records
	if err := verifyDataDirectory(dir, r.eventsPath); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *DiskRepository) casePath(id string) string { return filepath.Join(r.dir, "cases", id+".json") }

func (r *DiskRepository) Create(ctx context.Context, c *domain.AcceptanceCase, events []domain.CaseEvent, requestID string, result any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if requestID != "" {
		if record, ok := r.records["\x00"+requestID]; ok {
			return json.Unmarshal(record.Result, result)
		}
	}
	if _, err := os.Stat(r.casePath(c.ID)); err == nil {
		return domain.ErrConflict
	}
	if err := writeSnapshot(r.casePath(c.ID), snapshot{Case: *c, ContentHistory: []domain.DocumentContent{}}); err != nil {
		return err
	}
	if err := appendEvents(r.eventsPath, events); err != nil {
		return err
	}
	return r.rememberLocked("", requestID, result)
}

func (r *DiskRepository) Save(ctx context.Context, req SaveRequest) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if req.RequestID != "" {
		if record, ok := r.records[req.Case.ID+"\x00"+req.RequestID]; ok {
			return json.Unmarshal(record.Result, req.Result)
		}
	}
	s, err := readSnapshot(r.casePath(req.Case.ID))
	if errors.Is(err, os.ErrNotExist) {
		return domain.ErrNotFound
	}
	if err != nil {
		return err
	}
	if s.Case.Revision != req.ExpectedRevision {
		return domain.ErrConflict
	}
	var content *domain.DocumentContent
	history := cloneContentHistory(s.ContentHistory)
	if len(history) == 0 && s.Content != nil {
		history = append(history, s.Content.Clone())
	}
	if req.Content != nil {
		cloned := req.Content.Clone()
		content = &cloned
		if len(history) == 0 || history[len(history)-1].Revision != cloned.Revision {
			history = append(history, cloned.Clone())
		}
	} else {
		content = s.Content
	}
	if err := writeSnapshot(r.casePath(req.Case.ID), snapshot{Case: *req.Case, Content: content, ContentHistory: history}); err != nil {
		return err
	}
	if err := appendEvents(r.eventsPath, req.Events); err != nil {
		return err
	}
	return r.rememberLocked(req.Case.ID, req.RequestID, req.Result)
}

func cloneContentHistory(items []domain.DocumentContent) []domain.DocumentContent {
	out := make([]domain.DocumentContent, len(items))
	for i := range items {
		out[i] = items[i].Clone()
	}
	return out
}

func (r *DiskRepository) ContentHistory(ctx context.Context, id string) ([]domain.DocumentContent, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, err := readSnapshot(r.casePath(id))
	if errors.Is(err, os.ErrNotExist) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	history := cloneContentHistory(s.ContentHistory)
	if len(history) == 0 && s.Content != nil {
		history = append(history, s.Content.Clone())
	}
	return history, nil
}

func (r *DiskRepository) rememberLocked(caseID, requestID string, result any) error {
	if requestID == "" {
		return nil
	}
	b, err := json.Marshal(result)
	if err != nil {
		return err
	}
	r.records[caseID+"\x00"+requestID] = idempotencyRecord{CaseID: caseID, RequestID: requestID, Result: b}
	return writeIdempotency(r.idempotencyPath, r.records)
}

func (r *DiskRepository) Get(ctx context.Context, id string) (*domain.AcceptanceCase, *domain.DocumentContent, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, err := readSnapshot(r.casePath(id))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, nil, err
	}
	c := s.Case
	var content *domain.DocumentContent
	if s.Content != nil {
		cloned := s.Content.Clone()
		content = &cloned
	}
	return &c, content, nil
}

func (r *DiskRepository) List(ctx context.Context) ([]domain.AcceptanceCase, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	paths, err := filepath.Glob(filepath.Join(r.dir, "cases", "*.json"))
	if err != nil {
		return nil, err
	}
	items := make([]domain.AcceptanceCase, 0, len(paths))
	for _, path := range paths {
		s, err := readSnapshot(path)
		if err != nil {
			return nil, err
		}
		items = append(items, s.Case)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].UpdatedAt.After(items[j].UpdatedAt) })
	return items, nil
}

func (r *DiskRepository) Timeline(ctx context.Context, caseID string) ([]domain.CaseEvent, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return readEvents(r.eventsPath, caseID)
}

func (r *DiskRepository) IdempotentResult(ctx context.Context, caseID, requestID string, dst any) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if requestID == "" {
		return false, nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	record, ok := r.records[caseID+"\x00"+requestID]
	if !ok {
		return false, nil
	}
	return true, json.Unmarshal(record.Result, dst)
}
