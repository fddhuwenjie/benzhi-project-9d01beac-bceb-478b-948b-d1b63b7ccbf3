package audit

import (
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"
	"strings"
)

type ImageAltRule struct{}

func (ImageAltRule) Code() string { return "A11Y-IMAGE-ALT" }
func (r ImageAltRule) Evaluate(content domain.DocumentContent) []Finding {
	var out []Finding
	for _, block := range content.Blocks {
		if block.Type == domain.BlockImage && strings.TrimSpace(block.AltText) == "" {
			out = append(out, Finding{RuleCode: r.Code(), Severity: "error", BlockID: block.ID, Message: "图片缺少可理解的替代文本"})
		}
	}
	return out
}
