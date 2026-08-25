package domain

import (
	"errors"
	"testing"
	"time"
)

func TestIssueRejectsStaleEvidence(t *testing.T) {
	now := time.Now().UTC()
	issue := AuditIssue{ID: "issue-1", Status: IssueOpen}
	if err := issue.Submit("已修正", "替代文本已补充", "编辑", 1, now); err != nil {
		t.Fatal(err)
	}
	if err := issue.Review(true, "审核员", "", 2, now); !errors.Is(err, ErrStaleEvidence) {
		t.Fatalf("预期旧证据错误，得到 %v", err)
	}
}

func TestCaseCannotApproveUntilEveryIssueAccepted(t *testing.T) {
	now := time.Now().UTC()
	c, _ := NewCase("指南", "机构", "编辑", "2026-10-01", now)
	c.ContentRevision = 1
	c.Status = StatusReview
	c.Issues = []AuditIssue{{ID: "one", Status: IssueAccepted, Evidence: &Evidence{ContentRevision: 1}}, {ID: "two", Status: IssueSubmitted, Evidence: &Evidence{ContentRevision: 1}}}
	if err := c.Approve("审核员", now); !errors.Is(err, ErrNotReady) {
		t.Fatalf("未全部接受时应拒绝批准：%v", err)
	}
	c.Issues[1].Status = IssueAccepted
	if err := c.Approve("审核员", now); err != nil {
		t.Fatalf("全部接受后应允许批准：%v", err)
	}
	if c.Status != StatusApproved || c.Declaration == nil || len(c.Declaration.Dispositions) != 2 {
		t.Fatalf("批准结果不完整：%#v", c)
	}
}

func TestPublishedMetadataIsImmutable(t *testing.T) {
	now := time.Now().UTC()
	c, _ := NewCase("指南", "机构", "编辑", "2026-10-01", now)
	c.Status = StatusPublished
	err := c.UpdateMetadata("新标题", "机构", "编辑", "2026-10-01", now)
	if !errors.Is(err, ErrNotReady) || c.Title != "指南" {
		t.Fatalf("已发布信息不应变更：%v %#v", err, c)
	}
}

func TestDeclarationRejectsDamagedDigestWithoutMutation(t *testing.T) {
	now := time.Date(2026, 8, 25, 10, 0, 0, 0, time.UTC)
	c, err := NewCase("指南", "机构", "编辑", "2026-10-01", now)
	if err != nil {
		t.Fatal(err)
	}
	c.Status = StatusReview
	c.ContentRevision = 1
	c.RuleVersion = "rules-v1"
	if err = c.Approve("审核员", now); err != nil {
		t.Fatal(err)
	}
	revision := c.Revision
	c.Declaration.Digest = "sha256:damaged"
	if err = c.Publish("审核员", now.Add(time.Hour)); !errors.Is(err, ErrIntegrity) {
		t.Fatalf("预期声明完整性错误，得到 %v", err)
	}
	if c.Status != StatusApproved || c.Revision != revision || c.Declaration.PublishedAt != nil {
		t.Fatal("完整性失败后声明或聚合发生变化")
	}
}
