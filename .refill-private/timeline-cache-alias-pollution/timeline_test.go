package timelinecache

import (
	"context"
	"testing"
	"time"

	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/store"
)

func TestTimelineCacheDoesNotExposeMutablePayload(t *testing.T) {
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
	event := domain.NewEvent(c.ID, "case.created", c.Owner, "create-1", c.Revision, now, map[string]any{
		"title":        c.Title,
		"organization": c.Organization,
	})
	if err := repo.Create(context.Background(), c, []domain.CaseEvent{event}, "create-1", map[string]any{"case_id": c.ID}); err != nil {
		t.Fatal(err)
	}

	first, err := repo.Timeline(context.Background(), c.ID)
	if err != nil || len(first) != 1 {
		t.Fatalf("首次读取时间线失败：%v，事件数 %d", err, len(first))
	}
	first[0].Actor = "被污染的调用方"
	first[0].Payload["title"] = "被污染的标题"

	second, err := repo.Timeline(context.Background(), c.ID)
	if err != nil || len(second) != 1 {
		t.Fatalf("再次读取时间线失败：%v，事件数 %d", err, len(second))
	}
	if second[0].Actor != c.Owner || second[0].Payload["title"] != c.Title {
		t.Fatalf("时间线缓存泄漏了调用方修改：actor=%q title=%q", second[0].Actor, second[0].Payload["title"])
	}

	reopened, err := store.NewDiskRepository(dir)
	if err != nil {
		t.Fatal(err)
	}
	persisted, err := reopened.Timeline(context.Background(), c.ID)
	if err != nil || len(persisted) != 1 {
		t.Fatalf("重启后读取时间线失败：%v，事件数 %d", err, len(persisted))
	}
	if persisted[0].Actor != c.Owner || persisted[0].Payload["title"] != c.Title {
		t.Fatalf("磁盘事件本应保持原值：actor=%q title=%q", persisted[0].Actor, persisted[0].Payload["title"])
	}
}
