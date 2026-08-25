package workflow

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/audit"
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/store"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	repo, err := store.NewDiskRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 25, 10, 0, 0, 0, time.UTC)
	return NewService(repo, audit.NewEngine()).WithClock(func() time.Time { now = now.Add(time.Minute); return now })
}

func TestFullWorkflowAndIdempotency(t *testing.T) {
	ctx := context.Background()
	s := newTestService(t)
	create := CreateCaseCommand{Title: "办事指南", Organization: "服务中心", Owner: "编辑", TargetPublishDate: "2026-10-01", Actor: "编辑", RequestID: "create-1"}
	created, err := s.CreateCase(ctx, create)
	if err != nil {
		t.Fatal(err)
	}
	repeated, err := s.CreateCase(ctx, create)
	if err != nil || repeated != created {
		t.Fatalf("重复建档没有返回原结果：%#v %v", repeated, err)
	}
	items, _ := s.ListCases(ctx)
	if len(items) != 1 {
		t.Fatalf("重复建档产生了 %d 个验收案", len(items))
	}

	saved, err := s.SaveContent(ctx, SaveContentCommand{CaseID: created.CaseID, Revision: created.Revision, Actor: "编辑", RequestID: "save-1", Blocks: []domain.ContentBlock{{ID: "image-1", Type: domain.BlockImage, Text: "流程图"}}})
	if err != nil {
		t.Fatal(err)
	}
	audited, err := s.Audit(ctx, AuditCommand{CaseID: created.CaseID, Revision: saved.Revision, Actor: "编辑", RequestID: "audit-1"})
	if err != nil {
		t.Fatal(err)
	}
	detail, _ := s.GetCase(ctx, created.CaseID)
	if len(detail.Case.Issues) != 1 || audited.Status != domain.StatusRemediating {
		t.Fatalf("审查结果不正确：%#v", detail.Case)
	}
	issueID := detail.Case.Issues[0].ID
	_, err = s.SubmitEvidence(ctx, SubmitEvidenceCommand{CaseID: created.CaseID, IssueID: issueID, Revision: audited.Revision, ContentRevision: 0, Description: "补充", Text: "已有替代文本", Actor: "编辑", RequestID: "evidence-stale"})
	if !errors.Is(err, domain.ErrStaleEvidence) {
		t.Fatalf("预期旧证据错误，得到 %v", err)
	}
	submitted, err := s.SubmitEvidence(ctx, SubmitEvidenceCommand{CaseID: created.CaseID, IssueID: issueID, Revision: audited.Revision, ContentRevision: audited.ContentRevision, Description: "补充", Text: "已有替代文本", Actor: "编辑", RequestID: "evidence-1"})
	if err != nil || submitted.Status != domain.StatusReview {
		t.Fatalf("证据提交失败：%#v %v", submitted, err)
	}
	accepted, err := s.ReviewIssue(ctx, ReviewIssueCommand{CaseID: created.CaseID, IssueID: issueID, Revision: submitted.Revision, Accept: true, Reviewer: "审核员", Comment: "通过", RequestID: "review-1"})
	if err != nil {
		t.Fatal(err)
	}
	approved, err := s.Approve(ctx, ApproveCommand{CaseID: created.CaseID, Revision: accepted.Revision, Approver: "审核员", RequestID: "approve-1"})
	if err != nil || approved.DeclarationID == "" {
		t.Fatalf("批准失败：%#v %v", approved, err)
	}
	published, err := s.Publish(ctx, PublishCommand{CaseID: created.CaseID, Revision: approved.Revision, Actor: "审核员", RequestID: "publish-1"})
	if err != nil || published.Status != domain.StatusPublished {
		t.Fatalf("发布失败：%#v %v", published, err)
	}
	again, err := s.Publish(ctx, PublishCommand{CaseID: created.CaseID, Revision: 1, Actor: "审核员", RequestID: "publish-1"})
	if err != nil || again != published {
		t.Fatalf("重复发布没有返回原结果：%#v %v", again, err)
	}
	declaration, err := s.Declaration(ctx, created.CaseID)
	if err != nil || declaration.RuleVersion != audit.RuleVersion || len(declaration.Dispositions) != 1 {
		t.Fatalf("声明不完整：%#v %v", declaration, err)
	}
}

func TestOptimisticConcurrency(t *testing.T) {
	s := newTestService(t)
	ctx := context.Background()
	created, _ := s.CreateCase(ctx, CreateCaseCommand{Title: "指南", Organization: "机构", Owner: "编辑", TargetPublishDate: "2026-10-01", Actor: "编辑", RequestID: "create"})
	_, err := s.UpdateCase(ctx, UpdateCaseCommand{CaseID: created.CaseID, Revision: created.Revision, Title: "新指南", Organization: "机构", Owner: "编辑", TargetPublishDate: "2026-10-01", Actor: "编辑", RequestID: "update-1"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.UpdateCase(ctx, UpdateCaseCommand{CaseID: created.CaseID, Revision: created.Revision, Title: "又一标题", Organization: "机构", Owner: "编辑", TargetPublishDate: "2026-10-01", Actor: "编辑", RequestID: "update-2"})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("预期修订冲突，得到 %v", err)
	}
}

func TestConcurrentDuplicateCreateReturnsOneCase(t *testing.T) {
	repo, err := store.NewDiskRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	s := NewService(repo, audit.NewEngine())
	cmd := CreateCaseCommand{Title: "并发指南", Organization: "机构", Owner: "编辑", TargetPublishDate: "2026-10-01", Actor: "编辑", RequestID: "same-request"}
	results := make([]MutationResult, 12)
	errs := make([]error, len(results))
	var wg sync.WaitGroup
	for i := range results {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			results[index], errs[index] = s.CreateCase(context.Background(), cmd)
		}(i)
	}
	wg.Wait()
	for i := range results {
		if errs[i] != nil {
			t.Fatalf("并发请求 %d 失败：%v", i, errs[i])
		}
		if results[i] != results[0] {
			t.Fatalf("并发请求结果不一致：%#v / %#v", results[0], results[i])
		}
	}
	cases, err := s.ListCases(context.Background())
	if err != nil || len(cases) != 1 {
		t.Fatalf("并发重复建档产生 %d 个验收案：%v", len(cases), err)
	}
}
