package workflow

import (
	"context"
	"errors"
	"testing"
	"time"

	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"
)

func TestSearchValidationStableOrderAndReadOnlyProgress(t *testing.T) {
	s := newTestService(t)
	ctx := context.Background()
	first, _ := s.CreateCase(ctx, CreateCaseCommand{Title: "  Alpha 指南 ", Organization: "公共服务中心", Owner: "编辑甲", TargetPublishDate: "2026-09-02", Actor: "编辑甲", RequestID: "search-create-1"})
	second, _ := s.CreateCase(ctx, CreateCaseCommand{Title: "beta", Organization: "PUBLIC 服务中心", Owner: "编辑甲", TargetPublishDate: "2026-09-01", Actor: "编辑甲", RequestID: "search-create-2"})
	_, _ = s.CreateCase(ctx, CreateCaseCommand{Title: "其他", Organization: "别处", Owner: "编辑乙", TargetPublishDate: "2026-09-01", Actor: "编辑乙", RequestID: "search-create-3"})
	page, err := s.SearchCases(ctx, CaseListQuery{Keyword: " public ", Owner: " 编辑甲 ", Sort: "target_publish_date", Order: "asc", Limit: 1})
	if err != nil || page.Total != 1 || page.Cases[0].ID != second.CaseID {
		t.Fatalf("组合检索结果不正确：%#v %v", page, err)
	}
	page, err = s.SearchCases(ctx, CaseListQuery{Owner: "编辑甲", Sort: "target_publish_date", Order: "asc", Limit: 1})
	if err != nil || page.Total != 2 || page.Cases[0].ID != second.CaseID || page.NextCursor == "" {
		t.Fatalf("稳定排序首页不正确：%#v %v", page, err)
	}
	next, err := s.SearchCases(ctx, CaseListQuery{Owner: "编辑甲", Sort: "target_publish_date", Order: "asc", Limit: 1, Cursor: page.NextCursor})
	if err != nil || len(next.Cases) != 1 || next.Cases[0].ID != first.CaseID {
		t.Fatalf("稳定游标下一页不正确：%#v %v", next, err)
	}
	detail, _ := s.GetCase(ctx, first.CaseID)
	if detail.Case.Revision != first.Revision {
		t.Fatal("只读检索改变了聚合修订")
	}
	_, err = s.SearchCases(ctx, CaseListQuery{DateFrom: "2026-09-03", DateTo: "2026-09-01"})
	var fields domain.ValidationErrors
	if !errors.As(err, &fields) || fields[0].Field != "date_to" {
		t.Fatalf("非法日期区间未返回字段错误：%v", err)
	}
}

func TestContentHistoryRestoreReauditAndAtomicBatches(t *testing.T) {
	s := newTestService(t)
	ctx := context.Background()
	created, _ := s.CreateCase(ctx, CreateCaseCommand{Title: "指南", Organization: "机构", Owner: "编辑", TargetPublishDate: "2026-10-01", Actor: "编辑", RequestID: "ext-create"})
	v1Blocks := []domain.ContentBlock{{ID: "image", Type: domain.BlockImage, Text: "图"}, {ID: "link", Type: domain.BlockLink, LinkTarget: "/go"}}
	v1, err := s.SaveContent(ctx, SaveContentCommand{CaseID: created.CaseID, Revision: created.Revision, Actor: "编辑", RequestID: "ext-save-1", Blocks: v1Blocks})
	if err != nil {
		t.Fatal(err)
	}
	v2Blocks := []domain.ContentBlock{{ID: "image", Type: domain.BlockImage, Text: "图", AltText: "流程图"}, {ID: "link", Type: domain.BlockLink, Text: "办理", LinkTarget: "/go"}, {ID: "heading", Type: domain.BlockHeading, HeadingLevel: 1, Text: "办理"}}
	v2, _ := s.SaveContent(ctx, SaveContentCommand{CaseID: created.CaseID, Revision: v1.Revision, Actor: "编辑乙", RequestID: "ext-save-2", Blocks: v2Blocks})
	diff, err := s.CompareContent(ctx, created.CaseID, 1, 2)
	if err != nil || len(diff.Added) != 1 || len(diff.Changed) != 2 {
		t.Fatalf("内容差异不正确：%#v %v", diff, err)
	}
	restored, err := s.RestoreContent(ctx, RestoreContentCommand{CaseID: created.CaseID, Revision: v2.Revision, SourceRevision: 1, Actor: "编辑", RequestID: "ext-restore"})
	if err != nil || restored.ContentRevision != 3 {
		t.Fatalf("回退未生成新版本：%#v %v", restored, err)
	}
	repeated, err := s.RestoreContent(ctx, RestoreContentCommand{CaseID: created.CaseID, Revision: 1, SourceRevision: 1, Actor: "编辑", RequestID: "ext-restore"})
	if err != nil || repeated != restored {
		t.Fatalf("回退幂等结果不一致：%#v %v", repeated, err)
	}
	_, err = s.RestoreContent(ctx, RestoreContentCommand{CaseID: created.CaseID, Revision: v2.Revision, SourceRevision: 2, Actor: "编辑", RequestID: "ext-restore-conflict"})
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("并发回退未冲突：%v", err)
	}
	audited, err := s.Audit(ctx, AuditCommand{CaseID: created.CaseID, Revision: restored.Revision, Actor: "编辑", RequestID: "ext-audit"})
	if err != nil {
		t.Fatal(err)
	}
	detail, _ := s.GetCase(ctx, created.CaseID)
	if len(detail.Case.Issues) != 2 {
		t.Fatalf("预期两个问题：%#v", detail.Case.Issues)
	}
	invalidItems := []EvidenceInput{{IssueID: detail.Case.Issues[0].ID, Description: "已修复", Text: "证据"}, {IssueID: detail.Case.Issues[1].ID, Description: "", Text: "证据"}}
	_, err = s.SubmitEvidenceBatch(ctx, SubmitEvidenceBatchCommand{CaseID: created.CaseID, Revision: audited.Revision, ContentRevision: audited.ContentRevision, Actor: "编辑", RequestID: "ext-batch-invalid", Items: invalidItems})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("无效批次未失败：%v", err)
	}
	afterInvalid, _ := s.GetCase(ctx, created.CaseID)
	if afterInvalid.Case.Revision != audited.Revision || afterInvalid.Case.Issues[0].Evidence != nil {
		t.Fatal("无效证据批次发生了部分保存")
	}
	validItems := []EvidenceInput{{IssueID: detail.Case.Issues[0].ID, Description: "已修复", Text: "证据"}, {IssueID: detail.Case.Issues[1].ID, Description: "已修复", Text: "证据"}}
	submitted, err := s.SubmitEvidenceBatch(ctx, SubmitEvidenceBatchCommand{CaseID: created.CaseID, Revision: audited.Revision, ContentRevision: audited.ContentRevision, Actor: "编辑", RequestID: "ext-batch-valid", Items: validItems})
	if err != nil || submitted.Revision != audited.Revision+1 || submitted.Status != domain.StatusReview {
		t.Fatalf("证据批次结果不正确：%#v %v", submitted, err)
	}
	decisions := []ReviewDecision{{IssueID: detail.Case.Issues[0].ID, Accept: true, Comment: "通过"}, {IssueID: detail.Case.Issues[1].ID, Accept: false}}
	_, err = s.ReviewBatch(ctx, ReviewBatchCommand{CaseID: created.CaseID, Revision: submitted.Revision, Reviewer: "审核员", RequestID: "ext-review-invalid", Decisions: decisions})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("无效复核批次未失败：%v", err)
	}
	afterReviewInvalid, _ := s.GetCase(ctx, created.CaseID)
	if afterReviewInvalid.Case.Issues[0].Status != domain.IssueSubmitted {
		t.Fatal("无效复核批次发生了部分保存")
	}
	for i := range decisions {
		decisions[i].Accept = true
		decisions[i].Comment = "通过"
	}
	accepted, err := s.ReviewBatch(ctx, ReviewBatchCommand{CaseID: created.CaseID, Revision: submitted.Revision, Reviewer: "审核员", RequestID: "ext-review-valid", Decisions: decisions})
	if err != nil || accepted.Revision != submitted.Revision+1 || accepted.Status != domain.StatusReview {
		t.Fatalf("复核批次结果不正确：%#v %v", accepted, err)
	}
	preview, err := s.DeclarationPreview(ctx, created.CaseID, "审核员")
	if err != nil || !preview.ValidateDigest() {
		t.Fatalf("声明预览不完整：%#v %v", preview, err)
	}
	approved, err := s.Approve(ctx, ApproveCommand{CaseID: created.CaseID, Revision: accepted.Revision, Approver: "审核员", RequestID: "ext-approve"})
	if err != nil {
		t.Fatal(err)
	}
	published, err := s.Publish(ctx, PublishCommand{CaseID: created.CaseID, Revision: approved.Revision, Actor: "审核员", RequestID: "ext-publish"})
	if err != nil || published.Status != domain.StatusPublished {
		t.Fatalf("锁定声明发布失败：%#v %v", published, err)
	}
}

func TestReauditTracksResolvedAndAddedIssues(t *testing.T) {
	s := newTestService(t)
	ctx := context.Background()
	created, _ := s.CreateCase(ctx, CreateCaseCommand{Title: "指南", Organization: "机构", Owner: "编辑", TargetPublishDate: "2026-10-01", Actor: "编辑", RequestID: "reaudit-create"})
	saved, _ := s.SaveContent(ctx, SaveContentCommand{CaseID: created.CaseID, Revision: created.Revision, Actor: "编辑", RequestID: "reaudit-save-1", Blocks: []domain.ContentBlock{{ID: "image", Type: domain.BlockImage}}})
	audited, _ := s.Audit(ctx, AuditCommand{CaseID: created.CaseID, Revision: saved.Revision, Actor: "编辑", RequestID: "reaudit-1"})
	changed, err := s.SaveContent(ctx, SaveContentCommand{CaseID: created.CaseID, Revision: audited.Revision, Actor: "编辑", RequestID: "reaudit-save-2", Blocks: []domain.ContentBlock{{ID: "image", Type: domain.BlockImage, AltText: "流程"}, {ID: "link", Type: domain.BlockLink, LinkTarget: "/go"}}})
	if err != nil {
		t.Fatal(err)
	}
	result, err := s.Audit(ctx, AuditCommand{CaseID: created.CaseID, Revision: changed.Revision, Actor: "编辑", RequestID: "reaudit-2"})
	if err != nil {
		t.Fatal(err)
	}
	detail, _ := s.GetCase(ctx, created.CaseID)
	diff := detail.Case.LastAuditDifference
	if result.Status != domain.StatusRemediating || diff == nil || len(diff.Resolved) != 1 || len(diff.Added) != 1 || len(diff.Persistent) != 0 {
		t.Fatalf("复审差异不正确：%#v", diff)
	}
	if detail.Case.Issues[0].RuleCode != "A11Y-LINK-TEXT" || detail.Case.Issues[0].Evidence != nil {
		t.Fatalf("复审当前问题不正确：%#v", detail.Case.Issues)
	}
}

func TestTimelineIntegrityDetectsRevisionAndRequestConflict(t *testing.T) {
	now := time.Date(2026, 8, 25, 10, 0, 0, 0, time.UTC)
	events := []domain.CaseEvent{
		{ID: "event-1", CaseID: "case-1", EventType: "case.created", Actor: "编辑", RequestID: "request-1", CaseRevision: 1, OccurredAt: now, Payload: map[string]any{"title": "指南"}},
		{ID: "event-2", CaseID: "case-1", EventType: "content.saved", Actor: "编辑", RequestID: "request-2", CaseRevision: 2, OccurredAt: now.Add(time.Minute), Payload: map[string]any{"content_revision": 1}},
		{ID: "event-3", CaseID: "case-1", EventType: "issue.reviewed", Actor: "审核员", RequestID: "request-3", CaseRevision: 3, OccurredAt: now.Add(2 * time.Minute), Payload: map[string]any{"issue_id": "issue-1"}},
	}
	integrity := validateTimeline(events, 3)
	if !integrity.Valid {
		t.Fatalf("正常时间线被判为异常：%#v", integrity)
	}
	events[2].CaseRevision = 1
	events[2].RequestID = "request-2"
	integrity = validateTimeline(events, 3)
	if integrity.Valid || integrity.Status != "integrity_error" || len(integrity.Errors) < 2 {
		t.Fatalf("时间线异常未被识别：%#v", integrity)
	}
}
