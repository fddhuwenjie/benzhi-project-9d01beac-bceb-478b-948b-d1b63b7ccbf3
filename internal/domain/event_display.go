package domain

import "fmt"

var eventPayloadFields = map[string]map[string]bool{
	"case.created":             {"title": true, "organization": true},
	"case.metadata_updated":    {"title": true, "organization": true, "owner": true, "target_publish_date": true},
	"content.saved":            {"content_revision": true, "block_count": true},
	"content.restored":         {"source_revision": true, "content_revision": true, "block_count": true},
	"audit.completed":          {"rule_version": true, "content_revision": true, "issue_count": true},
	"audit.recompleted":        {"rule_version": true, "content_revision": true, "issue_count": true, "resolved": true, "persistent": true, "added": true},
	"evidence.submitted":       {"issue_id": true, "content_revision": true},
	"evidence.batch_submitted": {"issue_ids": true, "content_revision": true, "count": true},
	"issue.reviewed":           {"issue_id": true, "decision": true, "comment": true},
	"issue.batch_reviewed":     {"decisions": true, "count": true},
	"case.approved":            {"declaration_id": true, "rule_version": true, "content_revision": true, "digest": true},
	"declaration.published":    {"declaration_id": true, "digest": true},
}

func (e CaseEvent) DisplayPayload() map[string]any {
	allowed, known := eventPayloadFields[e.EventType]
	if !known {
		return map[string]any{"summary": "该事件类型的业务载荷不对外展示"}
	}
	out := make(map[string]any)
	for key, value := range e.Payload {
		if allowed[key] {
			out[key] = displayValue(value)
		}
	}
	return out
}

func displayValue(value any) any {
	switch typed := value.(type) {
	case nil, string, bool, float64, int, int64, []string, []any:
		return typed
	default:
		return fmt.Sprint(typed)
	}
}
