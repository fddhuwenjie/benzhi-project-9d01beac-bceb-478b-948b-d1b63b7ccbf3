package auditcachecrosscase

import (
	"context"
	"testing"

	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/audit"
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/store"
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/workflow"
)

func TestAuditCacheIsolatedAcrossCases(t *testing.T) {
	t.Parallel()

	repo, err := store.NewDiskRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	service := workflow.NewService(repo, audit.NewEngine())
	ctx := context.Background()

	first := createAndSave(t, ctx, service, "first", []domain.ContentBlock{{
		ID: "image-1", Type: domain.BlockImage, Text: "办事流程图",
	}})
	if _, err := service.Audit(ctx, workflow.AuditCommand{
		CaseID: first.CaseID, Revision: first.Revision, Actor: "编辑甲", RequestID: "first-audit",
	}); err != nil {
		t.Fatalf("审查第一个验收案失败：%v", err)
	}

	second := createAndSave(t, ctx, service, "second", []domain.ContentBlock{{
		ID: "heading-1", Type: domain.BlockHeading, HeadingLevel: 1, Text: "办理说明",
	}})
	if _, err := service.Audit(ctx, workflow.AuditCommand{
		CaseID: second.CaseID, Revision: second.Revision, Actor: "编辑乙", RequestID: "second-audit",
	}); err != nil {
		t.Fatalf("审查第二个验收案失败：%v", err)
	}

	detail, err := service.GetCase(ctx, second.CaseID)
	if err != nil {
		t.Fatal(err)
	}
	if len(detail.Case.Issues) != 0 {
		t.Fatalf("第二个验收案复用了其他验收案的审查问题：%#v", detail.Case.Issues)
	}
	if detail.Case.Status != domain.StatusReview {
		t.Fatalf("无问题验收案状态错误：%s", detail.Case.Status)
	}
}

func createAndSave(t *testing.T, ctx context.Context, service *workflow.Service, key string, blocks []domain.ContentBlock) workflow.MutationResult {
	t.Helper()
	created, err := service.CreateCase(ctx, workflow.CreateCaseCommand{
		Title: "无障碍指南-" + key, Organization: "公共服务中心", Owner: "编辑", TargetPublishDate: "2026-10-01",
		Actor: "编辑", RequestID: key + "-create",
	})
	if err != nil {
		t.Fatalf("创建验收案失败：%v", err)
	}
	saved, err := service.SaveContent(ctx, workflow.SaveContentCommand{
		CaseID: created.CaseID, Revision: created.Revision, Actor: "编辑", RequestID: key + "-save", Blocks: blocks,
	})
	if err != nil {
		t.Fatalf("保存内容失败：%v", err)
	}
	return saved
}
