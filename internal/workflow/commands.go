package workflow

import (
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"
	"encoding/json"
	"strconv"
	"time"
)

type CreateCaseCommand struct{ Title, Organization, Owner, TargetPublishDate, Actor, RequestID string }
type UpdateCaseCommand struct {
	CaseID            string
	Revision          int64
	Title             string
	Organization      string
	Owner             string
	TargetPublishDate string
	Actor             string
	RequestID         string
}
type SaveContentCommand struct {
	CaseID           string
	Revision         int64
	Actor, RequestID string
	Blocks           []domain.ContentBlock
}
type RestoreContentCommand struct {
	CaseID           string
	Revision         int64
	SourceRevision   int64
	Actor, RequestID string
}
type AuditCommand struct {
	CaseID           string
	Revision         int64
	Actor, RequestID string
}
type SubmitEvidenceCommand struct {
	CaseID, IssueID                     string
	Revision                            int64
	Description, Text, Actor, RequestID string
	ContentRevision                     int64
}
type EvidenceInput struct {
	IssueID, Description, Text string
	ContentRevision            int64
}
type SubmitEvidenceBatchCommand struct {
	CaseID                    string
	Revision, ContentRevision int64
	Actor, RequestID          string
	Items                     []EvidenceInput
}
type ReviewIssueCommand struct {
	CaseID, IssueID              string
	Revision                     int64
	Accept                       bool
	Reviewer, Comment, RequestID string
}
type ReviewDecision struct {
	IssueID string
	Accept  bool
	Comment string
}
type ReviewBatchCommand struct {
	CaseID              string
	Revision            int64
	Reviewer, RequestID string
	Decisions           []ReviewDecision
}
type ApproveCommand struct {
	CaseID              string
	Revision            int64
	Approver, RequestID string
}
type PublishCommand struct {
	CaseID           string
	Revision         int64
	Actor, RequestID string
}

type MutationResult struct {
	CaseID          string            `json:"case_id"`
	Revision        int64             `json:"revision"`
	ContentRevision int64             `json:"content_revision"`
	Status          domain.CaseStatus `json:"status"`
	IssueID         string            `json:"issue_id,omitempty"`
	DeclarationID   string            `json:"declaration_id,omitempty"`
}

type CaseDetail struct {
	Case              *domain.AcceptanceCase  `json:"case"`
	Content           *domain.DocumentContent `json:"content,omitempty"`
	Timeline          []TimelineEvent         `json:"timeline"`
	Progress          domain.CaseProgress     `json:"progress"`
	ContentRevisions  []ContentRevisionInfo   `json:"content_revisions"`
	TimelineIntegrity TimelineIntegrity       `json:"timeline_integrity"`
}

type ContentRevisionInfo struct {
	Revision   int64     `json:"revision"`
	SavedBy    string    `json:"saved_by"`
	SavedAt    time.Time `json:"saved_at"`
	BlockCount int       `json:"block_count"`
}

// fingerprint computes a stable hash of an idempotency request's semantic
// content (operation type plus request payload, excluding the optimistic
// concurrency revision). The same request_id reused for a replay of the same
// operation yields the same fingerprint and reuses the cached result; using
// the same request_id for a different operation or payload yields a different
// fingerprint and is rejected as a conflict rather than silently returning the
// unrelated cached result.
func fingerprint(operation string, parts ...any) string {
	h := uint64(1469598103934665603)
	mix := func(s string) {
		for _, c := range []byte(s) {
			h ^= uint64(c)
			h *= 1099511628211
		}
		h ^= 0xff
		h *= 1099511628211
	}
	mix(operation)
	for _, part := range parts {
		switch v := part.(type) {
		case string:
			mix(v)
		case int64:
			mix(strconv.FormatInt(v, 10))
		case int:
			mix(strconv.Itoa(v))
		case bool:
			if v {
				mix("true")
			} else {
				mix("false")
			}
		default:
			b, _ := json.Marshal(v)
			mix(string(b))
		}
	}
	const digits = "0123456789abcdef"
	buf := make([]byte, 16)
	for i := 15; i >= 0; i-- {
		buf[i] = digits[h&15]
		h >>= 4
	}
	return operation + ":" + string(buf)
}
