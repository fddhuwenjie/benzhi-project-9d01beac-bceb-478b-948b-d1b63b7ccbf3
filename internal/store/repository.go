package store

import (
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"
	"context"
)

type SaveRequest struct {
	Case             *domain.AcceptanceCase
	Content          *domain.DocumentContent
	ExpectedRevision int64
	Events           []domain.CaseEvent
	RequestID        string
	Result           any
}

type Repository interface {
	Create(context.Context, *domain.AcceptanceCase, []domain.CaseEvent, string, any) error
	Save(context.Context, SaveRequest) error
	Get(context.Context, string) (*domain.AcceptanceCase, *domain.DocumentContent, error)
	List(context.Context) ([]domain.AcceptanceCase, error)
	SearchCases(context.Context, CaseQuery) (CasePage, error)
	ContentHistory(context.Context, string) ([]domain.DocumentContent, error)
	Timeline(context.Context, string) ([]domain.CaseEvent, error)
	IdempotentResult(context.Context, string, string, any) (bool, error)
}
