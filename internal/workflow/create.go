package workflow

import (
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"
	"context"
	"strings"
)

func (s *Service) CreateCase(ctx context.Context, cmd CreateCaseCommand) (MutationResult, error) {
	if strings.TrimSpace(cmd.Actor) == "" || strings.TrimSpace(cmd.RequestID) == "" {
		return MutationResult{}, domain.ErrValidation
	}
	fp := fingerprint("case.created", cmd.Title, cmd.Organization, cmd.Owner, cmd.TargetPublishDate, cmd.Actor)
	if prior, ok, err := s.prior(ctx, "", cmd.RequestID, fp); err != nil || ok {
		return prior, err
	}
	now := s.now()
	c, err := domain.NewCase(cmd.Title, cmd.Organization, cmd.Owner, cmd.TargetPublishDate, now)
	if err != nil {
		return MutationResult{}, err
	}
	result := resultFor(c)
	event := domain.NewEvent(c.ID, "case.created", cmd.Actor, cmd.RequestID, c.Revision, now, map[string]any{"title": c.Title, "organization": c.Organization})
	if err := s.repo.CreateWithFingerprint(ctx, c, []domain.CaseEvent{event}, cmd.RequestID, fp, &result); err != nil {
		return MutationResult{}, err
	}
	return result, nil
}
