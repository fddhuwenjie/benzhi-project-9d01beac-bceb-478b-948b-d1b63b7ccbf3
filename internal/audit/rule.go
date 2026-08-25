package audit

import "benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"

const RuleVersion = "WCAG-CN-1.0"

type Finding struct {
	RuleCode string
	Severity string
	BlockID  string
	Message  string
}

type Rule interface {
	Code() string
	Evaluate(domain.DocumentContent) []Finding
}
