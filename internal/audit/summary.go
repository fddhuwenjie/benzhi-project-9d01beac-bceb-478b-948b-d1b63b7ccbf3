package audit

import "benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"

type Summary struct {
	Total    int            `json:"total"`
	Errors   int            `json:"errors"`
	Warnings int            `json:"warnings"`
	ByRule   map[string]int `json:"by_rule"`
}

func Summarize(issues []domain.AuditIssue) Summary {
	s := Summary{Total: len(issues), ByRule: make(map[string]int)}
	for _, issue := range issues {
		s.ByRule[issue.RuleCode]++
		if issue.Severity == "error" {
			s.Errors++
		} else {
			s.Warnings++
		}
	}
	return s
}
