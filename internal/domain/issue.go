package domain

import (
	"strings"
	"time"
)

type IssueStatus string

const (
	IssueOpen      IssueStatus = "open"
	IssueSubmitted IssueStatus = "submitted"
	IssueAccepted  IssueStatus = "accepted"
	IssueReturned  IssueStatus = "returned"
)

type Evidence struct {
	Description     string    `json:"description"`
	Text            string    `json:"text"`
	ContentRevision int64     `json:"content_revision"`
	SubmittedBy     string    `json:"submitted_by"`
	SubmittedAt     time.Time `json:"submitted_at"`
}

type AuditIssue struct {
	ID                      string      `json:"id"`
	CaseID                  string      `json:"case_id"`
	RuleCode                string      `json:"rule_code"`
	RuleVersion             string      `json:"rule_version"`
	Severity                string      `json:"severity"`
	BlockID                 string      `json:"block_id"`
	Message                 string      `json:"message"`
	DetectedContentRevision int64       `json:"detected_content_revision"`
	Status                  IssueStatus `json:"status"`
	Evidence                *Evidence   `json:"evidence,omitempty"`
	Reviewer                string      `json:"reviewer,omitempty"`
	ReviewComment           string      `json:"review_comment,omitempty"`
	ReviewedAt              *time.Time  `json:"reviewed_at,omitempty"`
}

func (i *AuditIssue) Submit(description, text, actor string, contentRevision int64, now time.Time) error {
	if i.Status != IssueOpen && i.Status != IssueReturned {
		return ErrNotReady
	}
	if strings.TrimSpace(description) == "" || strings.TrimSpace(text) == "" || strings.TrimSpace(actor) == "" {
		return ErrValidation
	}
	i.Evidence = &Evidence{Description: strings.TrimSpace(description), Text: strings.TrimSpace(text), ContentRevision: contentRevision, SubmittedBy: strings.TrimSpace(actor), SubmittedAt: now.UTC()}
	i.Status = IssueSubmitted
	i.Reviewer, i.ReviewComment, i.ReviewedAt = "", "", nil
	return nil
}

func (i *AuditIssue) Review(accept bool, reviewer, comment string, latestContentRevision int64, now time.Time) error {
	if i.Status != IssueSubmitted || i.Evidence == nil {
		return ErrNotReady
	}
	if i.Evidence.ContentRevision != latestContentRevision {
		return ErrStaleEvidence
	}
	if strings.TrimSpace(reviewer) == "" {
		return ErrValidation
	}
	if !accept && strings.TrimSpace(comment) == "" {
		return ErrValidation
	}
	i.Reviewer = strings.TrimSpace(reviewer)
	i.ReviewComment = strings.TrimSpace(comment)
	t := now.UTC()
	i.ReviewedAt = &t
	if accept {
		i.Status = IssueAccepted
	} else {
		i.Status = IssueReturned
	}
	return nil
}
