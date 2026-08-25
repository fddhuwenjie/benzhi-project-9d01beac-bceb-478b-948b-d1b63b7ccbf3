package workflow

import (
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"
	"context"
	"encoding/base64"
	"encoding/json"
	"sort"
	"strings"
	"time"
)

type TimelineQuery struct {
	EventType, Actor, Cursor string
	RevisionFrom, RevisionTo int64
	Limit                    int
}
type TimelineIntegrity struct {
	Valid  bool     `json:"valid"`
	Status string   `json:"status"`
	Errors []string `json:"errors,omitempty"`
}
type TimelineEvent struct {
	ID           string         `json:"id"`
	CaseID       string         `json:"case_id"`
	EventType    string         `json:"event_type"`
	Actor        string         `json:"actor"`
	RequestID    string         `json:"request_id"`
	CaseRevision int64          `json:"case_revision"`
	OccurredAt   time.Time      `json:"occurred_at"`
	Payload      map[string]any `json:"payload"`
}
type TimelinePage struct {
	Events     []TimelineEvent   `json:"events"`
	NextCursor string            `json:"next_cursor,omitempty"`
	Integrity  TimelineIntegrity `json:"integrity"`
}
type timelineCursor struct {
	OccurredAt time.Time `json:"occurred_at"`
	ID         string    `json:"id"`
}

func (s *Service) Timeline(ctx context.Context, caseID string, query TimelineQuery) (TimelinePage, error) {
	if strings.TrimSpace(caseID) == "" {
		return TimelinePage{}, domain.ErrValidation
	}
	var fields domain.ValidationErrors
	if query.RevisionFrom < 0 {
		fields = append(fields, domain.FieldError{Field: "revision_from", Message: "起始修订不能为负数"})
	}
	if query.RevisionTo < 0 || query.RevisionTo > 0 && query.RevisionFrom > query.RevisionTo {
		fields = append(fields, domain.FieldError{Field: "revision_to", Message: "结束修订不能早于起始修订"})
	}
	if query.Limit < 0 || query.Limit > 100 {
		fields = append(fields, domain.FieldError{Field: "limit", Message: "每页数量必须在 1 到 100 之间"})
	}
	if !fields.Empty() {
		return TimelinePage{}, fields
	}
	c, _, err := s.repo.Get(ctx, caseID)
	if err != nil {
		return TimelinePage{}, err
	}
	events, err := s.repo.Timeline(ctx, caseID)
	if err != nil {
		return TimelinePage{}, err
	}
	integrity := validateTimeline(events, c.Revision)
	filtered := make([]domain.CaseEvent, 0, len(events))
	for _, event := range events {
		if query.EventType != "" && event.EventType != query.EventType {
			continue
		}
		if query.Actor != "" && !strings.EqualFold(strings.TrimSpace(event.Actor), strings.TrimSpace(query.Actor)) {
			continue
		}
		if query.RevisionFrom > 0 && event.CaseRevision < query.RevisionFrom {
			continue
		}
		if query.RevisionTo > 0 && event.CaseRevision > query.RevisionTo {
			continue
		}
		filtered = append(filtered, event)
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		if filtered[i].OccurredAt.Equal(filtered[j].OccurredAt) {
			return filtered[i].ID < filtered[j].ID
		}
		return filtered[i].OccurredAt.Before(filtered[j].OccurredAt)
	})
	start := 0
	if query.Cursor != "" {
		var cursor timelineCursor
		b, decodeErr := base64.RawURLEncoding.DecodeString(query.Cursor)
		if decodeErr == nil {
			decodeErr = json.Unmarshal(b, &cursor)
		}
		found := false
		if decodeErr == nil {
			for i := range filtered {
				if filtered[i].ID == cursor.ID && filtered[i].OccurredAt.Equal(cursor.OccurredAt) {
					start = i + 1
					found = true
					break
				}
			}
		}
		if !found {
			return TimelinePage{}, domain.ValidationErrors{{Field: "cursor", Message: "时间线分页游标无效或已失效"}}
		}
	}
	limit := query.Limit
	if limit == 0 {
		limit = 30
	}
	end := start + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	page := TimelinePage{Events: make([]TimelineEvent, 0, end-start), Integrity: integrity}
	for _, event := range filtered[start:end] {
		page.Events = append(page.Events, TimelineEvent{ID: event.ID, CaseID: event.CaseID, EventType: event.EventType, Actor: event.Actor, RequestID: event.RequestID, CaseRevision: event.CaseRevision, OccurredAt: event.OccurredAt, Payload: event.DisplayPayload()})
	}
	if end < len(filtered) && end > 0 {
		cursor, _ := json.Marshal(timelineCursor{filtered[end-1].OccurredAt, filtered[end-1].ID})
		page.NextCursor = base64.RawURLEncoding.EncodeToString(cursor)
	}
	return page, nil
}

func validateTimeline(events []domain.CaseEvent, caseRevision int64) TimelineIntegrity {
	result := TimelineIntegrity{Valid: true, Status: "ok"}
	var previous int64
	requests := make(map[string]string)
	for _, event := range events {
		if previous > 0 && (event.CaseRevision < previous || event.CaseRevision > previous+1) {
			result.Errors = append(result.Errors, "事件 "+event.ID+" 的聚合修订关系不连续")
		}
		if event.CaseRevision > previous {
			previous = event.CaseRevision
		}
		if event.RequestID != "" {
			value, _ := json.Marshal(struct {
				Type     string
				Revision int64
				Payload  map[string]any
			}{event.EventType, event.CaseRevision, event.Payload})
			if prior, exists := requests[event.RequestID]; exists && prior != string(value) {
				result.Errors = append(result.Errors, "request_id "+event.RequestID+" 对应了冲突结果")
			} else {
				requests[event.RequestID] = string(value)
			}
		}
	}
	if len(events) > 0 && previous != caseRevision {
		result.Errors = append(result.Errors, "事件时间线末尾修订与验收案快照不一致")
	}
	if len(result.Errors) > 0 {
		result.Valid = false
		result.Status = "integrity_error"
	}
	return result
}
