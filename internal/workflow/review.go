package workflow

import (
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/store"
	"context"
	"fmt"
	"strings"
)

func (s *Service) ReviewIssue(ctx context.Context, cmd ReviewIssueCommand) (MutationResult, error) {
	if err := validateMutation(cmd.CaseID, cmd.Reviewer, cmd.RequestID, cmd.Revision); err != nil {
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
	if c.Status != domain.StatusReview {
		return MutationResult{}, domain.ErrNotReady
	}
	issue, err := c.FindIssue(cmd.IssueID)
	if err != nil {
		return MutationResult{}, err
	}
	now := s.now()
	expected := c.Revision
	if err = issue.Review(cmd.Accept, cmd.Reviewer, cmd.Comment, c.ContentRevision, now); err != nil {
		return MutationResult{}, err
	}
	c.Touch(now)
	c.RefreshReviewStatus(now)
	decision := "returned"
	if cmd.Accept {
		decision = "accepted"
	}
	result := resultFor(c)
	result.IssueID = issue.ID
	event := domain.NewEvent(c.ID, "issue.reviewed", cmd.Reviewer, cmd.RequestID, c.Revision, now, map[string]any{"issue_id": issue.ID, "decision": decision, "comment": cmd.Comment})
	if err = s.repo.Save(ctx, store.SaveRequest{Case: c, ExpectedRevision: expected, Events: []domain.CaseEvent{event}, RequestID: cmd.RequestID, Result: &result}); err != nil {
		return MutationResult{}, err
	}
	return result, nil
}

func (s *Service) ReviewBatch(ctx context.Context, cmd ReviewBatchCommand) (MutationResult, error) {
	if err := validateMutation(cmd.CaseID, cmd.Reviewer, cmd.RequestID, cmd.Revision); err != nil {
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
	if c.Status != domain.StatusReview {
		return MutationResult{}, domain.ErrNotReady
	}
	var fields domain.ValidationErrors
	if len(cmd.Decisions) == 0 {
		fields = append(fields, domain.FieldError{Field: "decisions", Message: "至少选择一项待复核证据"})
	}
	seen := make(map[string]bool)
	issues := make([]*domain.AuditIssue, len(cmd.Decisions))
	for index, decision := range cmd.Decisions {
		prefix := "decisions[" + decision.IssueID + "]"
		if strings.TrimSpace(decision.IssueID) == "" {
			prefix = fmt.Sprintf("decisions[%d]", index)
			fields = append(fields, domain.FieldError{Field: prefix + ".issue_id", Message: "问题标识不能为空"})
			continue
		}
		if seen[decision.IssueID] {
			fields = append(fields, domain.FieldError{Field: prefix + ".issue_id", Message: "同一问题不能重复复核"})
			continue
		}
		seen[decision.IssueID] = true
		issue, findErr := c.FindIssue(decision.IssueID)
		if findErr != nil || issue.CaseID != c.ID {
			fields = append(fields, domain.FieldError{Field: prefix + ".issue_id", Message: "问题不属于当前验收案"})
			continue
		}
		issues[index] = issue
		if issue.Status != domain.IssueSubmitted || issue.Evidence == nil {
			fields = append(fields, domain.FieldError{Field: prefix + ".status", Message: "证据已被处理或当前不可复核"})
			continue
		}
		if issue.Evidence.ContentRevision != c.ContentRevision {
			fields = append(fields, domain.FieldError{Field: prefix + ".content_revision", Message: domain.ErrStaleEvidence.Error()})
		}
		if !decision.Accept && strings.TrimSpace(decision.Comment) == "" {
			fields = append(fields, domain.FieldError{Field: prefix + ".comment", Message: "退回原因不能为空"})
		}
	}
	if !fields.Empty() {
		return MutationResult{}, fields
	}
	now := s.now()
	payload := make([]map[string]any, len(cmd.Decisions))
	for index, issue := range issues {
		decision := cmd.Decisions[index]
		_ = issue.Review(decision.Accept, cmd.Reviewer, decision.Comment, c.ContentRevision, now)
		value := "returned"
		if decision.Accept {
			value = "accepted"
		}
		payload[index] = map[string]any{"issue_id": decision.IssueID, "decision": value, "comment": decision.Comment}
	}
	expected := c.Revision
	c.Touch(now)
	c.RefreshReviewStatus(now)
	result := resultFor(c)
	event := domain.NewEvent(c.ID, "issue.batch_reviewed", cmd.Reviewer, cmd.RequestID, c.Revision, now, map[string]any{"decisions": payload, "count": len(payload)})
	if err = s.repo.Save(ctx, store.SaveRequest{Case: c, ExpectedRevision: expected, Events: []domain.CaseEvent{event}, RequestID: cmd.RequestID, Result: &result}); err != nil {
		return MutationResult{}, err
	}
	return result, nil
}
