package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"
)

func TestDiskRepositoryPersistsSnapshotTimelineAndRequestResult(t *testing.T) {
	dir := t.TempDir()
	repo, err := NewDiskRepository(dir)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 25, 10, 0, 0, 0, time.UTC)
	c, err := domain.NewCase("办事指南", "公共服务中心", "编辑", "2026-09-01", now)
	if err != nil {
		t.Fatal(err)
	}
	result := map[string]any{"case_id": c.ID, "revision": float64(c.Revision)}
	event := domain.NewEvent(c.ID, "case.created", "编辑", "create-1", c.Revision, now, nil)
	if err := repo.Create(context.Background(), c, []domain.CaseEvent{event}, "create-1", result); err != nil {
		t.Fatal(err)
	}

	reopened, err := NewDiskRepository(dir)
	if err != nil {
		t.Fatal(err)
	}
	loaded, _, err := reopened.Get(context.Background(), c.ID)
	if err != nil || loaded.Title != c.Title {
		t.Fatalf("快照恢复失败：%#v %v", loaded, err)
	}
	var prior map[string]any
	ok, err := reopened.IdempotentResult(context.Background(), "", "create-1", &prior)
	if err != nil || !ok || prior["case_id"] != c.ID {
		t.Fatalf("幂等结果恢复失败：%#v %v", prior, err)
	}
	events, err := reopened.Timeline(context.Background(), c.ID)
	if err != nil || len(events) != 1 || events[0].EventType != "case.created" {
		t.Fatalf("事件恢复失败：%#v %v", events, err)
	}
}

func TestDiskRepositoryRejectsStaleRevision(t *testing.T) {
	repo, _ := NewDiskRepository(t.TempDir())
	now := time.Now().UTC()
	c, _ := domain.NewCase("指南", "机构", "编辑", "2026-09-01", now)
	if err := repo.Create(context.Background(), c, nil, "create", map[string]any{}); err != nil {
		t.Fatal(err)
	}
	changed := *c
	changed.Touch(now.Add(time.Minute))
	err := repo.Save(context.Background(), SaveRequest{Case: &changed, ExpectedRevision: c.Revision - 1})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("预期修订冲突，得到 %v", err)
	}
}
