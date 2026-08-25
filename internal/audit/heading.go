package audit

import "benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"

type HeadingRule struct{}

func (HeadingRule) Code() string { return "A11Y-HEADING-LEVEL" }
func (r HeadingRule) Evaluate(content domain.DocumentContent) []Finding {
	previous := 0
	var out []Finding
	for _, block := range content.Blocks {
		if block.Type != domain.BlockHeading {
			continue
		}
		if previous > 0 && block.HeadingLevel > previous+1 {
			out = append(out, Finding{RuleCode: r.Code(), Severity: "error", BlockID: block.ID, Message: "标题层级从 H" + digit(previous) + " 跳至 H" + digit(block.HeadingLevel)})
		}
		previous = block.HeadingLevel
	}
	return out
}
func digit(i int) string { return string(rune('0' + i)) }
