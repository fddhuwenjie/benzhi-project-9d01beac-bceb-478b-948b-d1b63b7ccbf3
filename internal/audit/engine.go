package audit

import (
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"
	"sort"
	"strconv"
	"sync"
)

type Engine struct {
	rules []Rule
	mu    sync.RWMutex
	cache map[string][]domain.AuditIssue
}

func NewEngine() *Engine {
	return &Engine{
		rules: []Rule{HeadingRule{}, LinkTextRule{}, ImageAltRule{}, TableHeaderRule{}},
		cache: make(map[string][]domain.AuditIssue),
	}
}

func (e *Engine) Version() string { return RuleVersion }

func (e *Engine) cacheKey(content domain.DocumentContent) string {
	// Revision is per-case, so the cache must be keyed by both CaseID and
	// Revision to prevent issues from one acceptance case leaking into
	// another case that happens to share the same revision number.
	return content.CaseID + "@" + strconv.FormatInt(content.Revision, 10)
}

func (e *Engine) Run(content domain.DocumentContent) []domain.AuditIssue {
	key := e.cacheKey(content)
	e.mu.RLock()
	cached, ok := e.cache[key]
	e.mu.RUnlock()
	if ok {
		return append([]domain.AuditIssue(nil), cached...)
	}

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
	e.mu.Lock()
	e.cache[key] = append([]domain.AuditIssue(nil), issues...)
	e.mu.Unlock()
	return issues
}
