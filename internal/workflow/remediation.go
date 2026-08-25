package workflow

import (
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/store"
	"context"
	"fmt"
	"strings"
)

func (s *Service) SubmitEvidence(ctx context.Context, cmd SubmitEvidenceCommand) (MutationResult, error) {
	if err := validateMutation(cmd.CaseID, cmd.Actor, cmd.RequestID, cmd.Revision); err != nil {
		return MutationResult{}, err
	}
	if prior, ok, err := s.prior(ctx, cmd.CaseID, cmd.RequestID); err != nil || ok {
		return prior, err
	}
	c, _, err := s.repo.Get(ctx, cmd.CaseID)
	if err != nil {
		return MutationResult{}, err
	}
	if err = checkRevision(c, cmd.Revision); err != nil {
		return MutationResult{}, err
	}
	if cmd.ContentRevision != c.ContentRevision {
		return MutationResult{}, domain.ErrStaleEvidence
	}
	if c.Status != domain.StatusRemediating && c.Status != domain.StatusReview {
		return MutationResult{}, domain.ErrNotReady
	}
	issue, err := c.FindIssue(cmd.IssueID)
	if err != nil {
		return MutationResult{}, err
	}
	now := s.now()
	expected := c.Revision
	if err = issue.Submit(cmd.Description, cmd.Text, cmd.Actor, cmd.ContentRevision, now); err != nil {
		return MutationResult{}, err
	}
	c.Touch(now)
	c.RefreshReviewStatus(now)
	result := resultFor(c)
	result.IssueID = issue.ID
	event := domain.NewEvent(c.ID, "evidence.submitted", cmd.Actor, cmd.RequestID, c.Revision, now, map[string]any{"issue_id": issue.ID, "content_revision": cmd.ContentRevision})
	if err = s.repo.Save(context.WithoutCancel(ctx), store.SaveRequest{Case: c, ExpectedRevision: expected, Events: []domain.CaseEvent{event}, RequestID: cmd.RequestID, Result: &result}); err != nil {
		return MutationResult{}, err
	}
	return result, nil
}

func (s *Service) SubmitEvidenceBatch(ctx context.Context, cmd SubmitEvidenceBatchCommand) (MutationResult, error) {
	if err := validateMutation(cmd.CaseID, cmd.Actor, cmd.RequestID, cmd.Revision); err != nil {
		return MutationResult{}, err
	}
	if prior, ok, err := s.prior(ctx, cmd.CaseID, cmd.RequestID); err != nil || ok {
		return prior, err
	}
	c, _, err := s.repo.Get(ctx, cmd.CaseID)
	if err != nil {
		return MutationResult{}, err
	}
	if err = checkRevision(c, cmd.Revision); err != nil {
		return MutationResult{}, err
	}
	if c.Status != domain.StatusRemediating && c.Status != domain.StatusReview {
		return MutationResult{}, domain.ErrNotReady
	}
	var fields domain.ValidationErrors
	if len(cmd.Items) == 0 {
		fields = append(fields, domain.FieldError{Field: "items", Message: "至少选择一个待整改问题"})
	}
	if cmd.ContentRevision != c.ContentRevision {
		fields = append(fields, domain.FieldError{Field: "content_revision", Message: domain.ErrStaleEvidence.Error()})
	}
	seen := make(map[string]bool)
	issues := make([]*domain.AuditIssue, len(cmd.Items))
	for index, item := range cmd.Items {
		prefix := "items[" + item.IssueID + "]"
		if strings.TrimSpace(item.IssueID) == "" {
			prefix = fmt.Sprintf("items[%d]", index)
			fields = append(fields, domain.FieldError{Field: prefix + ".issue_id", Message: "问题标识不能为空"})
			continue
		}
		if seen[item.IssueID] {
			fields = append(fields, domain.FieldError{Field: prefix + ".issue_id", Message: "同一问题不能重复提交"})
			continue
		}
		seen[item.IssueID] = true
		issue, findErr := c.FindIssue(item.IssueID)
		if findErr != nil || issue.CaseID != c.ID {
			fields = append(fields, domain.FieldError{Field: prefix + ".issue_id", Message: "问题不属于当前验收案"})
			continue
		}
		issues[index] = issue
		if issue.Status != domain.IssueOpen && issue.Status != domain.IssueReturned {
			fields = append(fields, domain.FieldError{Field: prefix + ".status", Message: "问题当前状态不可提交证据"})
		}
		if strings.TrimSpace(item.Description) == "" {
			fields = append(fields, domain.FieldError{Field: prefix + ".description", Message: "整改说明不能为空"})
		}
		if strings.TrimSpace(item.Text) == "" {
			fields = append(fields, domain.FieldError{Field: prefix + ".text", Message: "文本证据不能为空"})
		}
		if item.ContentRevision != 0 && item.ContentRevision != c.ContentRevision {
			fields = append(fields, domain.FieldError{Field: prefix + ".content_revision", Message: "该问题证据已过期"})
		}
	}
	if !fields.Empty() {
		return MutationResult{}, fields
	}
	now := s.now()
	for index, issue := range issues {
		_ = issue.Submit(cmd.Items[index].Description, cmd.Items[index].Text, cmd.Actor, c.ContentRevision, now)
	}
	expected := c.Revision
	c.Touch(now)
	c.RefreshReviewStatus(now)
	result := resultFor(c)
	issueIDs := make([]string, len(cmd.Items))
	for i := range cmd.Items {
		issueIDs[i] = cmd.Items[i].IssueID
	}
	event := domain.NewEvent(c.ID, "evidence.batch_submitted", cmd.Actor, cmd.RequestID, c.Revision, now, map[string]any{"issue_ids": issueIDs, "content_revision": c.ContentRevision, "count": len(issueIDs)})
	if err = s.repo.Save(context.WithoutCancel(ctx), store.SaveRequest{Case: c, ExpectedRevision: expected, Events: []domain.CaseEvent{event}, RequestID: cmd.RequestID, Result: &result}); err != nil {
		return MutationResult{}, err
	}
	return result, nil
}
