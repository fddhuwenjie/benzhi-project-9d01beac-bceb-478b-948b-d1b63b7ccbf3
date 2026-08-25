package domain

import "time"

type CaseEvent struct {
	ID           string         `json:"id"`
	CaseID       string         `json:"case_id"`
	EventType    string         `json:"event_type"`
	Actor        string         `json:"actor"`
	RequestID    string         `json:"request_id,omitempty"`
	CaseRevision int64          `json:"case_revision"`
	OccurredAt   time.Time      `json:"occurred_at"`
	Payload      map[string]any `json:"payload,omitempty"`
}

func NewEvent(caseID, eventType, actor, requestID string, revision int64, now time.Time, payload map[string]any) CaseEvent {
	return CaseEvent{ID: NewID("evt"), CaseID: caseID, EventType: eventType, Actor: actor, RequestID: requestID, CaseRevision: revision, OccurredAt: now.UTC(), Payload: payload}
}
