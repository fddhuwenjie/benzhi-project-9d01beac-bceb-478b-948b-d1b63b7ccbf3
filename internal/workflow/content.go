package workflow

import (
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/store"
	"context"
)

func (s *Service) SaveContent(ctx context.Context, cmd SaveContentCommand) (MutationResult, error) {
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
	newContent := domain.NewDocumentContent(c.ID, cmd.Actor, c.ContentRevision+1, cmd.Blocks, now)
	content := &newContent
	if errs := content.Validate(); !errs.Empty() {
		return MutationResult{}, errs
	}
	expected := c.Revision
	if err = c.SetContentRevision(content.Revision, now); err != nil {
		return MutationResult{}, err
	}
	result := resultFor(c)
	event := domain.NewEvent(c.ID, "content.saved", cmd.Actor, cmd.RequestID, c.Revision, now, map[string]any{"content_revision": content.Revision, "block_count": len(content.Blocks)})
	if err = s.repo.Save(context.WithoutCancel(ctx), store.SaveRequest{Case: c, Content: content, ExpectedRevision: expected, Events: []domain.CaseEvent{event}, RequestID: cmd.RequestID, Result: &result}); err != nil {
		return MutationResult{}, err
	}
	return result, nil
}

func (s *Service) RestoreContent(ctx context.Context, cmd RestoreContentCommand) (MutationResult, error) {
	if err := validateMutation(cmd.CaseID, cmd.Actor, cmd.RequestID, cmd.Revision); err != nil {
		return MutationResult{}, err
	}
	if cmd.SourceRevision < 1 {
		return MutationResult{}, domain.ValidationErrors{{Field: "source_revision", Message: "回退来源修订无效"}}
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
	if c.Status != domain.StatusDraft && c.Status != domain.StatusRemediating {
		return MutationResult{}, domain.ErrNotReady
	}
	history, err := s.repo.ContentHistory(ctx, cmd.CaseID)
	if err != nil {
		return MutationResult{}, err
	}
	var source *domain.DocumentContent
	for i := range history {
		if history[i].Revision == cmd.SourceRevision {
			cloned := history[i].Clone()
			source = &cloned
			break
		}
	}
	if source == nil {
		return MutationResult{}, domain.ErrNotFound
	}
	now := s.now()
	restored := domain.NewDocumentContent(c.ID, cmd.Actor, c.ContentRevision+1, source.Blocks, now)
	if errs := restored.Validate(); !errs.Empty() {
		return MutationResult{}, errs
	}
	expected := c.Revision
	if err = c.SetContentRevision(restored.Revision, now); err != nil {
		return MutationResult{}, err
	}
	result := resultFor(c)
	event := domain.NewEvent(c.ID, "content.restored", cmd.Actor, cmd.RequestID, c.Revision, now, map[string]any{"source_revision": cmd.SourceRevision, "content_revision": restored.Revision, "block_count": restored.BlockCount})
	if err = s.repo.Save(context.WithoutCancel(ctx), store.SaveRequest{Case: c, Content: &restored, ExpectedRevision: expected, Events: []domain.CaseEvent{event}, RequestID: cmd.RequestID, Result: &result}); err != nil {
		return MutationResult{}, err
	}
	return result, nil
}
