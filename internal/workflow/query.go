package workflow

import (
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/store"
	"context"
	"strings"
	"time"
)

func (s *Service) ListCases(ctx context.Context) ([]domain.AcceptanceCase, error) {
	return s.repo.List(ctx)
}

type caseViews struct {
	timeline TimelinePage
	history  []domain.DocumentContent
}

func (s *Service) loadCaseViews(id string) (caseViews, error) {
	timeline, err := s.Timeline(context.Background(), id, TimelineQuery{Limit: 100})
	if err != nil {
		return caseViews{}, err
	}
	history, err := s.repo.ContentHistory(context.Background(), id)
	if err != nil {
		return caseViews{}, err
	}
	return caseViews{timeline: timeline, history: history}, nil
}

func (s *Service) GetCase(ctx context.Context, id string) (CaseDetail, error) {
	c, content, err := s.repo.Get(ctx, id)
	if err != nil {
		return CaseDetail{}, err
	}
	views, err := s.loadCaseViews(id)
	if err != nil {
		return CaseDetail{}, err
	}
	revisions := make([]ContentRevisionInfo, len(views.history))
	for index, item := range views.history {
		revisions[index] = ContentRevisionInfo{Revision: item.Revision, SavedBy: item.SavedBy, SavedAt: item.SavedAt, BlockCount: len(item.Blocks)}
	}
	return CaseDetail{Case: c, Content: content, Timeline: views.timeline.Events, Progress: c.Progress(s.now()), ContentRevisions: revisions, TimelineIntegrity: views.timeline.Integrity}, nil
}
func (s *Service) Declaration(ctx context.Context, id string) (*domain.Declaration, error) {
	c, _, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if c.Status != domain.StatusPublished || c.Declaration == nil {
		return nil, domain.ErrNotReady
	}
	d := *c.Declaration
	d.Dispositions = append([]domain.IssueDisposition(nil), c.Declaration.Dispositions...)
	return &d, nil
}

type CaseListQuery struct {
	Keyword, Owner, Status, DateFrom, DateTo, Sort, Order, Cursor string
	Limit                                                         int
}

type CaseListItem struct {
	domain.AcceptanceCase
	Overdue    bool   `json:"overdue"`
	NextRole   string `json:"next_role"`
	NextAction string `json:"next_action"`
}

type CaseListPage struct {
	Cases      []CaseListItem `json:"cases"`
	Total      int            `json:"total"`
	NextCursor string         `json:"next_cursor,omitempty"`
}

func (s *Service) SearchCases(ctx context.Context, query CaseListQuery) (CaseListPage, error) {
	query.Keyword, query.Owner = strings.TrimSpace(query.Keyword), strings.TrimSpace(query.Owner)
	var fields domain.ValidationErrors
	parseDate := func(field, value string) {
		if value != "" {
			if _, err := time.Parse("2006-01-02", value); err != nil {
				fields = append(fields, domain.FieldError{Field: field, Message: "日期必须为 YYYY-MM-DD"})
			}
		}
	}
	parseDate("date_from", query.DateFrom)
	parseDate("date_to", query.DateTo)
	if query.DateFrom != "" && query.DateTo != "" && query.DateTo < query.DateFrom {
		fields = append(fields, domain.FieldError{Field: "date_to", Message: "结束日期不能早于开始日期"})
	}
	status := domain.CaseStatus(query.Status)
	if status != "" && !status.Valid() {
		fields = append(fields, domain.FieldError{Field: "status", Message: "验收状态无效"})
	}
	if query.Sort == "" {
		query.Sort = "updated_at"
	}
	if query.Sort != "updated_at" && query.Sort != "target_publish_date" {
		fields = append(fields, domain.FieldError{Field: "sort", Message: "排序字段仅支持 target_publish_date 或 updated_at"})
	}
	if query.Order == "" {
		if query.Sort == "target_publish_date" {
			query.Order = "asc"
		} else {
			query.Order = "desc"
		}
	}
	if query.Order != "asc" && query.Order != "desc" {
		fields = append(fields, domain.FieldError{Field: "order", Message: "排序方向仅支持 asc 或 desc"})
	}
	if query.Limit < 0 || query.Limit > 100 {
		fields = append(fields, domain.FieldError{Field: "limit", Message: "每页数量必须在 1 到 100 之间"})
	}
	if !fields.Empty() {
		return CaseListPage{}, fields
	}
	page, err := s.repo.SearchCases(ctx, store.CaseQuery{Keyword: query.Keyword, Owner: query.Owner, Status: status, DateFrom: query.DateFrom, DateTo: query.DateTo, Sort: query.Sort, Order: query.Order, Cursor: query.Cursor, Limit: query.Limit})
	if err != nil {
		return CaseListPage{}, err
	}
	result := CaseListPage{Cases: make([]CaseListItem, len(page.Items)), Total: page.Total, NextCursor: page.NextCursor}
	for index, item := range page.Items {
		progress := item.Progress(s.now())
		result.Cases[index] = CaseListItem{AcceptanceCase: item, Overdue: progress.Overdue, NextRole: progress.NextRole, NextAction: progress.NextAction}
	}
	return result, nil
}

func (s *Service) ContentHistory(ctx context.Context, caseID string) ([]ContentRevisionInfo, error) {
	history, err := s.repo.ContentHistory(ctx, caseID)
	if err != nil {
		return nil, err
	}
	items := make([]ContentRevisionInfo, len(history))
	for i, item := range history {
		items[i] = ContentRevisionInfo{Revision: item.Revision, SavedBy: item.SavedBy, SavedAt: item.SavedAt, BlockCount: len(item.Blocks)}
	}
	return items, nil
}

func (s *Service) CompareContent(ctx context.Context, caseID string, from, to int64) (domain.ContentDifference, error) {
	history, err := s.repo.ContentHistory(ctx, caseID)
	if err != nil {
		return domain.ContentDifference{}, err
	}
	var left, right *domain.DocumentContent
	for i := range history {
		if history[i].Revision == from {
			item := history[i].Clone()
			left = &item
		}
		if history[i].Revision == to {
			item := history[i].Clone()
			right = &item
		}
	}
	if left == nil || right == nil {
		return domain.ContentDifference{}, domain.ErrNotFound
	}
	return domain.CompareContent(*left, *right), nil
}

func (s *Service) DeclarationPreview(ctx context.Context, id, approver string) (*domain.Declaration, error) {
	c, _, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if c.Declaration != nil {
		d := *c.Declaration
		d.Dispositions = append([]domain.IssueDisposition(nil), c.Declaration.Dispositions...)
		return &d, nil
	}
	return c.BuildDeclaration(strings.TrimSpace(approver), s.now())
}
