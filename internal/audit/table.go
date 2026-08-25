package audit

import (
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"
	"strings"
)

type TableHeaderRule struct{}

func (TableHeaderRule) Code() string { return "A11Y-TABLE-HEADER" }
func (r TableHeaderRule) Evaluate(content domain.DocumentContent) []Finding {
	var out []Finding
	for _, block := range content.Blocks {
		if block.Type != domain.BlockTable {
			continue
		}
		missing := len(block.TableHeaders) == 0
		for _, header := range block.TableHeaders {
			if strings.TrimSpace(header) == "" {
				missing = true
			}
		}
		if missing {
			out = append(out, Finding{RuleCode: r.Code(), Severity: "error", BlockID: block.ID, Message: "数据表格缺少完整的列标题"})
		}
	}
	return out
}
