package audit

import (
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"
	"sort"
)

type Engine struct{ rules []Rule }

func NewEngine() *Engine {
	return &Engine{rules: []Rule{HeadingRule{}, LinkTextRule{}, ImageAltRule{}, TableHeaderRule{}}}
}

func (e *Engine) Version() string { return RuleVersion }

func (e *Engine) Run(content domain.DocumentContent) []domain.AuditIssue {
	var findings []Finding
	for _, rule := range e.rules {
		findings = append(findings, rule.Evaluate(content)...)
	}
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].BlockID == findings[j].BlockID {
			return findings[i].RuleCode < findings[j].RuleCode
		}
		return findings[i].BlockID < findings[j].BlockID
	})
	issues := make([]domain.AuditIssue, 0, len(findings))
	for _, f := range findings {
		issues = append(issues, domain.AuditIssue{ID: domain.StableID("issue", content.CaseID, f.RuleCode, f.BlockID), CaseID: content.CaseID, RuleCode: f.RuleCode, RuleVersion: RuleVersion, Severity: f.Severity, BlockID: f.BlockID, Message: f.Message, DetectedContentRevision: content.Revision, Status: domain.IssueOpen})
	}
	return issues
}
