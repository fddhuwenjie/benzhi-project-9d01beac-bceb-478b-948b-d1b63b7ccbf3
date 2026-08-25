package audit

import (
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"
	"strings"
)

type LinkTextRule struct{}

func (LinkTextRule) Code() string { return "A11Y-LINK-TEXT" }
func (r LinkTextRule) Evaluate(content domain.DocumentContent) []Finding {
	var out []Finding
	for _, block := range content.Blocks {
		if block.Type == domain.BlockLink && strings.TrimSpace(block.Text) == "" {
			out = append(out, Finding{RuleCode: r.Code(), Severity: "error", BlockID: block.ID, Message: "链接文本不能为空，需说明链接目的"})
		}
	}
	return out
}
