package eventfailurepartialcommit_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/store"
)

func TestSaveFailureDoesNotCommitSnapshot(t *testing.T) {
	dir := t.TempDir()
	repo, err := store.NewDiskRepository(dir)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 25, 10, 0, 0, 0, time.UTC)
	c, err := domain.NewCase("原始标题", "机构", "编辑", "2026-10-01", now)
	if err != nil {
		t.Fatal(err)
	}
	created := domain.NewEvent(c.ID, "case.created", "编辑", "create", c.Revision, now, nil)
	if err := repo.Create(context.Background(), c, []domain.CaseEvent{created}, "create", &struct{}{}); err != nil {
		t.Fatal(err)
	}

	eventsPath := filepath.Join(dir, "events.jsonl")
	if err := os.Remove(eventsPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(eventsPath, 0700); err != nil {
		t.Fatal(err)
	}

	changed := *c
	if err := changed.UpdateMetadata("不应提交的标题", changed.Organization, changed.Owner, changed.TargetPublishDate, now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	event := domain.NewEvent(changed.ID, "case.metadata_updated", "编辑", "update", changed.Revision, now.Add(time.Minute), nil)
	err = repo.Save(context.Background(), store.SaveRequest{Case: &changed, ExpectedRevision: c.Revision, Events: []domain.CaseEvent{event}, RequestID: "update", Result: &struct{}{}})
	if err == nil {
		t.Fatal("事件日志不可写时 Save 意外成功")
	}
	loaded, _, getErr := repo.Get(context.Background(), c.ID)
	if getErr != nil {
		t.Fatal(getErr)
	}
	if loaded.Title != c.Title || loaded.Revision != c.Revision {
		t.Fatalf("Save 返回失败后快照仍被提交：title=%q revision=%d", loaded.Title, loaded.Revision)
	}
}
