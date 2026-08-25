package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
	"time"
)

type IssueDisposition struct {
	RuleCode string `json:"rule_code"`
	BlockID  string `json:"block_id"`
	Result   string `json:"result"`
}

type Declaration struct {
	ID                string             `json:"id"`
	CaseID            string             `json:"case_id"`
	DocumentTitle     string             `json:"document_title"`
	Organization      string             `json:"organization"`
	Owner             string             `json:"owner"`
	TargetPublishDate string             `json:"target_publish_date"`
	RuleVersion       string             `json:"rule_version"`
	ContentRevision   int64              `json:"content_revision"`
	Approver          string             `json:"approver"`
	ApprovedAt        time.Time          `json:"approved_at"`
	PublishedAt       *time.Time         `json:"published_at,omitempty"`
	Dispositions      []IssueDisposition `json:"issue_dispositions"`
	Digest            string             `json:"digest"`
}

type declarationCanonical struct {
	CaseID, DocumentTitle, Organization, Owner, TargetPublishDate string
	RuleVersion                                                   string
	ContentRevision                                               int64
	Approver                                                      string
	ApprovedAt                                                    time.Time
	Dispositions                                                  []IssueDisposition
}

func (d Declaration) CalculateDigest() string {
	value := declarationCanonical{d.CaseID, d.DocumentTitle, d.Organization, d.Owner, d.TargetPublishDate, d.RuleVersion, d.ContentRevision, d.Approver, d.ApprovedAt.UTC(), d.Dispositions}
	b, _ := json.Marshal(value)
	sum := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func (d Declaration) ValidateDigest() bool {
	return d.Digest != "" && d.Digest == d.CalculateDigest()
}

func (c *AcceptanceCase) BuildDeclaration(approver string, now time.Time) (*Declaration, error) {
	if c.Status != StatusReview || !c.AllAcceptedLatest() || strings.TrimSpace(approver) == "" {
		return nil, ErrNotReady
	}
	d := &Declaration{ID: NewID("decl"), CaseID: c.ID, DocumentTitle: c.Title, Organization: c.Organization, Owner: c.Owner, TargetPublishDate: c.TargetPublishDate, RuleVersion: c.RuleVersion, ContentRevision: c.ContentRevision, Approver: strings.TrimSpace(approver), ApprovedAt: now.UTC()}
	for _, issue := range c.Issues {
		d.Dispositions = append(d.Dispositions, IssueDisposition{RuleCode: issue.RuleCode, BlockID: issue.BlockID, Result: "accepted"})
	}
	current := make(map[string]bool, len(c.Issues))
	for _, issue := range c.Issues {
		current[issue.RuleCode+"\x00"+issue.BlockID] = true
	}
	seenHistory := make(map[string]bool)
	for _, issue := range c.IssueHistory {
		key := issue.RuleCode + "\x00" + issue.BlockID
		if !current[key] && !seenHistory[key] {
			d.Dispositions = append(d.Dispositions, IssueDisposition{RuleCode: issue.RuleCode, BlockID: issue.BlockID, Result: "resolved"})
			seenHistory[key] = true
		}
	}
	sort.Slice(d.Dispositions, func(i, j int) bool {
		if d.Dispositions[i].RuleCode == d.Dispositions[j].RuleCode {
			return d.Dispositions[i].BlockID < d.Dispositions[j].BlockID
		}
		return d.Dispositions[i].RuleCode < d.Dispositions[j].RuleCode
	})
	d.Digest = d.CalculateDigest()
	return d, nil
}

func (c *AcceptanceCase) Approve(approver string, now time.Time) error {
	d, err := c.BuildDeclaration(approver, now)
	if err != nil {
		return err
	}
	c.Status = StatusApproved
	c.Touch(now)
	c.Declaration = d
	return nil
}

func (c *AcceptanceCase) Publish(actor string, now time.Time) error {
	if c.Status != StatusApproved || c.Declaration == nil {
		return ErrNotReady
	}
	if strings.TrimSpace(actor) != c.Declaration.Approver || c.Declaration.ContentRevision != c.ContentRevision || c.Declaration.RuleVersion != c.RuleVersion || !c.Declaration.ValidateDigest() {
		return ErrIntegrity
	}
	c.Status = StatusPublished
	c.Touch(now)
	published := now.UTC()
	c.Declaration.PublishedAt = &published
	return nil
}
