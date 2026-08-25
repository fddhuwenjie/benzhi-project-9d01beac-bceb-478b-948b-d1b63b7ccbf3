package domain

import (
	"strings"
	"time"
)

type AcceptanceCase struct {
	ID                   string           `json:"id"`
	Title                string           `json:"title"`
	Organization         string           `json:"organization"`
	Owner                string           `json:"owner"`
	TargetPublishDate    string           `json:"target_publish_date"`
	Status               CaseStatus       `json:"status"`
	ContentRevision      int64            `json:"content_revision"`
	Revision             int64            `json:"revision"`
	RuleVersion          string           `json:"rule_version,omitempty"`
	AuditContentRevision int64            `json:"audit_content_revision,omitempty"`
	Issues               []AuditIssue     `json:"issues,omitempty"`
	IssueHistory         []AuditIssue     `json:"issue_history,omitempty"`
	LastAuditDifference  *AuditDifference `json:"last_audit_difference,omitempty"`
	Declaration          *Declaration     `json:"declaration,omitempty"`
	CreatedAt            time.Time        `json:"created_at"`
	UpdatedAt            time.Time        `json:"updated_at"`
}

func NewCase(title, organization, owner, targetDate string, now time.Time) (*AcceptanceCase, error) {
	c := &AcceptanceCase{ID: NewID("case"), Title: strings.TrimSpace(title), Organization: strings.TrimSpace(organization), Owner: strings.TrimSpace(owner), TargetPublishDate: strings.TrimSpace(targetDate), Status: StatusDraft, Revision: 1, CreatedAt: now.UTC(), UpdatedAt: now.UTC()}
	if errs := c.ValidateMetadata(); !errs.Empty() {
		return nil, errs
	}
	return c, nil
}

func (c *AcceptanceCase) ValidateMetadata() ValidationErrors {
	var errs ValidationErrors
	if c.Title == "" {
		errs = append(errs, FieldError{Field: "title", Message: "文档标题不能为空"})
	}
	if c.Organization == "" {
		errs = append(errs, FieldError{Field: "organization", Message: "所属机构不能为空"})
	}
	if c.Owner == "" {
		errs = append(errs, FieldError{Field: "owner", Message: "负责人不能为空"})
	}
	if _, err := time.Parse("2006-01-02", c.TargetPublishDate); err != nil {
		errs = append(errs, FieldError{Field: "target_publish_date", Message: "目标发布日期必须为 YYYY-MM-DD"})
	}
	return errs
}

func (c *AcceptanceCase) UpdateMetadata(title, organization, owner, targetDate string, now time.Time) error {
	if c.Status == StatusPublished {
		return ErrNotReady
	}
	previousTitle, previousOrganization := c.Title, c.Organization
	previousOwner, previousDate := c.Owner, c.TargetPublishDate
	c.Title = strings.TrimSpace(title)
	c.Organization = strings.TrimSpace(organization)
	c.Owner = strings.TrimSpace(owner)
	c.TargetPublishDate = strings.TrimSpace(targetDate)
	if errs := c.ValidateMetadata(); !errs.Empty() {
		c.Title, c.Organization = previousTitle, previousOrganization
		c.Owner, c.TargetPublishDate = previousOwner, previousDate
		return errs
	}
	c.Touch(now)
	return nil
}

func (c *AcceptanceCase) Transition(to CaseStatus, now time.Time) error {
	if !CanTransition(c.Status, to) {
		return ErrInvalidTransition
	}
	c.Status = to
	c.Touch(now)
	return nil
}

func (c *AcceptanceCase) Touch(now time.Time) { c.Revision++; c.UpdatedAt = now.UTC() }

func (c *AcceptanceCase) SetContentRevision(revision int64, now time.Time) error {
	if c.Status != StatusDraft && c.Status != StatusRemediating {
		return ErrNotReady
	}
	if revision != c.ContentRevision+1 {
		return ErrConflict
	}
	c.ContentRevision = revision
	c.Touch(now)
	return nil
}

func (c *AcceptanceCase) SetAudit(issues []AuditIssue, version string, now time.Time) error {
	if c.Status != StatusDraft || c.ContentRevision == 0 {
		return ErrNotReady
	}
	c.Issues = append([]AuditIssue(nil), issues...)
	c.RuleVersion = version
	c.AuditContentRevision = c.ContentRevision
	c.Status = StatusRemediating
	c.Touch(now)
	return nil
}

func (c *AcceptanceCase) FindIssue(id string) (*AuditIssue, error) {
	for i := range c.Issues {
		if c.Issues[i].ID == id {
			return &c.Issues[i], nil
		}
	}
	return nil, ErrNotFound
}

func (c *AcceptanceCase) AllSubmitted() bool {
	if len(c.Issues) == 0 {
		return true
	}
	for _, issue := range c.Issues {
		if issue.Status != IssueSubmitted && issue.Status != IssueAccepted {
			return false
		}
	}
	return true
}

func (c *AcceptanceCase) AllAcceptedLatest() bool {
	for _, issue := range c.Issues {
		if issue.Status != IssueAccepted || issue.Evidence == nil || issue.Evidence.ContentRevision != c.ContentRevision {
			return false
		}
	}
	return true
}

func (c *AcceptanceCase) RefreshReviewStatus(now time.Time) {
	if c.Status == StatusRemediating && c.AllSubmitted() {
		c.Status = StatusReview
	}
	if c.Status == StatusReview && !c.AllSubmitted() {
		c.Status = StatusRemediating
	}
}
