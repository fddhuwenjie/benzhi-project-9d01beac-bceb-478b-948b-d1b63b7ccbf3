package eventlogrotation_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/store"
)

func TestRepositoryReopensEventLogAfterRotation(t *testing.T) {
	dir := t.TempDir()
	repo, err := store.NewDiskRepository(dir)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 25, 10, 0, 0, 0, time.UTC)
	c, err := domain.NewCase("办事指南", "公共服务中心", "编辑", "2026-09-01", now)
	if err != nil {
		t.Fatal(err)
	}
	created := domain.NewEvent(c.ID, "case.created", "编辑", "create-rotation", c.Revision, now, nil)
	if err := repo.Create(context.Background(), c, []domain.CaseEvent{created}, "create-rotation", &struct{}{}); err != nil {
		t.Fatal(err)
	}

	activeLog := filepath.Join(dir, "events.jsonl")
	if err := os.Rename(activeLog, filepath.Join(dir, "events.rotated.jsonl")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(activeLog, nil, 0600); err != nil {
		t.Fatal(err)
	}

	updated := *c
	updated.Title = "办事指南（修订）"
	updated.Touch(now.Add(time.Minute))
	saved := domain.NewEvent(updated.ID, "case.metadata_updated", "编辑", "save-after-rotation", updated.Revision, now.Add(time.Minute), map[string]any{"title": updated.Title})
	result := map[string]any{"case_id": updated.ID, "revision": updated.Revision}
	if err := repo.Save(context.Background(), store.SaveRequest{Case: &updated, ExpectedRevision: c.Revision, Events: []domain.CaseEvent{saved}, RequestID: "save-after-rotation", Result: &result}); err != nil {
		t.Fatalf("轮换后的保存意外失败：%v", err)
	}
	loaded, _, err := repo.Get(context.Background(), c.ID)
	if err != nil || loaded.Title != updated.Title || loaded.Revision != updated.Revision {
		t.Fatalf("保存成功后快照未更新：case=%#v err=%v", loaded, err)
	}
	var prior map[string]any
	ok, err := repo.IdempotentResult(context.Background(), c.ID, "save-after-rotation", &prior)
	if err != nil || !ok || prior["revision"] != float64(updated.Revision) {
		t.Fatalf("保存成功后幂等结果未记录：result=%#v ok=%v err=%v", prior, ok, err)
	}

	events, err := repo.Timeline(context.Background(), c.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, event := range events {
		if event.RequestID == "save-after-rotation" {
			return
		}
	}
	t.Fatalf("活动事件日志缺少已成功保存的事件，events=%#v", events)
}
