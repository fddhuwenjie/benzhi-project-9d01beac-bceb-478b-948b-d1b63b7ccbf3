package workflow

import (
	"context"

	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/store"
)

func (s *Service) UpdateCase(ctx context.Context, cmd UpdateCaseCommand) (MutationResult, error) {
	if err := validateMutation(cmd.CaseID, cmd.Actor, cmd.RequestID, cmd.Revision); err != nil {
		return MutationResult{}, err
	}
	fp := fingerprint("case.metadata_updated", cmd.CaseID, cmd.Title, cmd.Organization, cmd.Owner, cmd.TargetPublishDate, cmd.Actor)
	if prior, ok, err := s.prior(ctx, cmd.CaseID, cmd.RequestID, fp); err != nil || ok {
		return prior, err
	}
	c, _, err := s.repo.Get(ctx, cmd.CaseID)
	if err != nil {
		return MutationResult{}, err
	}
	if err = checkRevision(c, cmd.Revision); err != nil {
		return MutationResult{}, err
	}
	expected := c.Revision
	now := s.now()
	if err = c.UpdateMetadata(cmd.Title, cmd.Organization, cmd.Owner, cmd.TargetPublishDate, now); err != nil {
		return MutationResult{}, err
	}
	result := resultFor(c)
	event := domain.NewEvent(c.ID, "case.metadata_updated", cmd.Actor, cmd.RequestID, c.Revision, now, map[string]any{
		"title": c.Title, "organization": c.Organization, "owner": c.Owner,
		"target_publish_date": c.TargetPublishDate,
	})
	err = s.repo.Save(ctx, store.SaveRequest{Case: c, ExpectedRevision: expected,
		Events: []domain.CaseEvent{event}, RequestID: cmd.RequestID, Fingerprint: fp, Result: &result})
	if err != nil {
		return MutationResult{}, err
	}
	return result, nil
}
