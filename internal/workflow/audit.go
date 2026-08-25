package workflow

import (
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/store"
	"context"
)

func (s *Service) Audit(ctx context.Context, cmd AuditCommand) (MutationResult, error) {
	if err := validateMutation(cmd.CaseID, cmd.Actor, cmd.RequestID, cmd.Revision); err != nil {
		return MutationResult{}, err
	}
	if prior, ok, err := s.prior(ctx, cmd.CaseID, cmd.RequestID); err != nil || ok {
		return prior, err
	}
	c, content, err := s.repo.Get(ctx, cmd.CaseID)
	if err != nil {
		return MutationResult{}, err
	}
	if err = checkRevision(c, cmd.Revision); err != nil {
		return MutationResult{}, err
	}
	if content == nil {
		return MutationResult{}, domain.ErrNotReady
	}
	now := s.now()
	issues := s.auditor.Run(*content)
	expected := c.Revision
	eventType := "audit.completed"
	payload := map[string]any{"rule_version": s.auditor.Version(), "content_revision": content.Revision, "issue_count": len(issues)}
	if c.Status == domain.StatusDraft {
		if err = c.SetAudit(issues, s.auditor.Version(), now); err != nil {
			return MutationResult{}, err
		}
		c.RefreshReviewStatus(now)
	} else {
		difference, reauditErr := c.Reaudit(issues, s.auditor.Version(), now)
		if reauditErr != nil {
			return MutationResult{}, reauditErr
		}
		eventType = "audit.recompleted"
		payload["resolved"], payload["persistent"], payload["added"] = difference.Resolved, difference.Persistent, difference.Added
	}
	result := resultFor(c)
	event := domain.NewEvent(c.ID, eventType, cmd.Actor, cmd.RequestID, c.Revision, now, payload)
	if err = s.repo.Save(ctx, store.SaveRequest{Case: c, ExpectedRevision: expected, Events: []domain.CaseEvent{event}, RequestID: cmd.RequestID, Result: &result}); err != nil {
		return MutationResult{}, err
	}
	return result, nil
}
