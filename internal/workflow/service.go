package workflow

import (
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/audit"
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/store"
	"context"
	"errors"
	"strings"
	"time"
)

type Service struct {
	repo    store.Repository
	auditor *audit.Engine
	now     func() time.Time
}

func NewService(repo store.Repository, auditor *audit.Engine) *Service {
	return &Service{repo: repo, auditor: auditor, now: time.Now}
}
func (s *Service) WithClock(clock func() time.Time) *Service { s.now = clock; return s }

func validateMutation(caseID, actor, requestID string, revision int64) error {
	if strings.TrimSpace(caseID) == "" || strings.TrimSpace(actor) == "" || strings.TrimSpace(requestID) == "" || revision < 1 {
		return domain.ErrValidation
	}
	return nil
}

func resultFor(c *domain.AcceptanceCase) MutationResult {
	r := MutationResult{CaseID: c.ID, Revision: c.Revision, ContentRevision: c.ContentRevision, Status: c.Status}
	if c.Declaration != nil {
		r.DeclarationID = c.Declaration.ID
	}
	return r
}

func (s *Service) prior(ctx context.Context, caseID, requestID, fingerprint string) (MutationResult, bool, error) {
	var result MutationResult
	ok, err := s.repo.IdempotentCheck(ctx, caseID, requestID, fingerprint, &result)
	return result, ok, err
}

func checkRevision(c *domain.AcceptanceCase, expected int64) error {
	if c.Revision != expected {
		return domain.ErrConflict
	}
	return nil
}
func is(err, target error) bool { return errors.Is(err, target) }
