package audit

import (
	"reflect"
	"testing"

	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"
)

func TestEngineFindsAllRequiredRulesDeterministically(t *testing.T) {
	content := domain.DocumentContent{CaseID: "case-test", Revision: 1, Blocks: []domain.ContentBlock{
		{ID: "h1", Type: domain.BlockHeading, HeadingLevel: 1, Text: "主标题"},
		{ID: "h3", Type: domain.BlockHeading, HeadingLevel: 3, Text: "跳级标题"},
		{ID: "link", Type: domain.BlockLink, LinkTarget: "/target"},
		{ID: "image", Type: domain.BlockImage},
		{ID: "table", Type: domain.BlockTable},
	}}
	engine := NewEngine()
	first, second := engine.Run(content), engine.Run(content)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("相同内容的审查结果不稳定：%#v != %#v", first, second)
	}
	if len(first) != 4 {
		t.Fatalf("预期 4 个问题，得到 %d", len(first))
	}
	codes := map[string]bool{}
	for _, issue := range first {
		codes[issue.RuleCode] = true
		if issue.ID == "" || issue.BlockID == "" || issue.RuleVersion != RuleVersion {
			t.Fatalf("问题缺少稳定定位信息：%#v", issue)
		}
	}
	for _, code := range []string{"A11Y-HEADING-LEVEL", "A11Y-LINK-TEXT", "A11Y-IMAGE-ALT", "A11Y-TABLE-HEADER"} {
		if !codes[code] {
			t.Errorf("缺少规则 %s", code)
		}
	}
}

func TestEngineAcceptsCompliantContent(t *testing.T) {
	content := domain.DocumentContent{CaseID: "case-ok", Revision: 1, Blocks: []domain.ContentBlock{
		{ID: "h1", Type: domain.BlockHeading, HeadingLevel: 1, Text: "主标题"},
		{ID: "h2", Type: domain.BlockHeading, HeadingLevel: 2, Text: "次标题"},
		{ID: "link", Type: domain.BlockLink, Text: "在线办理", LinkTarget: "/target"},
		{ID: "image", Type: domain.BlockImage, AltText: "办事窗口位置图"},
		{ID: "table", Type: domain.BlockTable, TableHeaders: []string{"事项", "时限"}},
	}}
	if issues := NewEngine().Run(content); len(issues) != 0 {
		t.Fatalf("合规内容不应产生问题：%#v", issues)
	}
}
