package workflow

import (
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/store"
	"context"
)

func (s *Service) Approve(ctx context.Context, cmd ApproveCommand) (MutationResult, error) {
	if err := validateMutation(cmd.CaseID, cmd.Approver, cmd.RequestID, cmd.Revision); err != nil {
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
	now := s.now()
	expected := c.Revision
	if err = c.Approve(cmd.Approver, now); err != nil {
		return MutationResult{}, err
	}
	result := resultFor(c)
	event := domain.NewEvent(c.ID, "case.approved", cmd.Approver, cmd.RequestID, c.Revision, now, map[string]any{"declaration_id": c.Declaration.ID, "rule_version": c.RuleVersion, "content_revision": c.ContentRevision, "digest": c.Declaration.Digest})
	if err = s.repo.Save(ctx, store.SaveRequest{Case: c, ExpectedRevision: expected, Events: []domain.CaseEvent{event}, RequestID: cmd.RequestID, Result: &result}); err != nil {
		return MutationResult{}, err
	}
	return result, nil
}

func (s *Service) Publish(ctx context.Context, cmd PublishCommand) (MutationResult, error) {
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
	now := s.now()
	expected := c.Revision
	if err = c.Publish(cmd.Actor, now); err != nil {
		return MutationResult{}, err
	}
	result := resultFor(c)
	event := domain.NewEvent(c.ID, "declaration.published", cmd.Actor, cmd.RequestID, c.Revision, now, map[string]any{"declaration_id": c.Declaration.ID, "digest": c.Declaration.Digest})
	if err = s.repo.Save(ctx, store.SaveRequest{Case: c, ExpectedRevision: expected, Events: []domain.CaseEvent{event}, RequestID: cmd.RequestID, Result: &result}); err != nil {
		return MutationResult{}, err
	}
	return result, nil
}
