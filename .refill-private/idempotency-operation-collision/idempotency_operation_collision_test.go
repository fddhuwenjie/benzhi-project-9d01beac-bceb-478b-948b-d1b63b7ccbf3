package idempotencyoperationcollision_test

import (
	"context"
	"errors"
	"testing"

	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/audit"
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/store"
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/workflow"
)

func TestRequestIDCannotReplayDifferentOperation(t *testing.T) {
	repo, err := store.NewDiskRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	service := workflow.NewService(repo, audit.NewEngine())
	ctx := context.Background()
	created, err := service.CreateCase(ctx, workflow.CreateCaseCommand{Title: "原始标题", Organization: "机构", Owner: "编辑", TargetPublishDate: "2026-10-01", Actor: "编辑", RequestID: "create"})
	if err != nil {
		t.Fatal(err)
	}
	saved, err := service.SaveContent(ctx, workflow.SaveContentCommand{CaseID: created.CaseID, Revision: created.Revision, Actor: "编辑", RequestID: "shared-key", Blocks: []domain.ContentBlock{{ID: "p1", Type: domain.BlockParagraph, Text: "正文"}}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.UpdateCase(ctx, workflow.UpdateCaseCommand{CaseID: created.CaseID, Revision: saved.Revision, Title: "新标题", Organization: "机构", Owner: "编辑", TargetPublishDate: "2026-10-01", Actor: "编辑", RequestID: "shared-key"})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("不同写操作复用 request_id 时应返回 ErrConflict，得到 %v", err)
	}
}
