package domain

import "testing"

func TestCompareContentClassifiesChanges(t *testing.T) {
	from := DocumentContent{Revision: 1, Blocks: []ContentBlock{
		{ID: "title", Type: BlockHeading, HeadingLevel: 2, Text: "标题"},
		{ID: "link", Type: BlockLink, Text: "查看", LinkTarget: "/old"},
		{ID: "text", Type: BlockParagraph, Text: "说明"},
	}}
	to := DocumentContent{Revision: 3, Blocks: []ContentBlock{
		{ID: "text", Type: BlockParagraph, Text: "说明"},
		{ID: "title", Type: BlockHeading, HeadingLevel: 3, Text: "标题"},
		{ID: "image", Type: BlockImage, AltText: "流程"},
	}}
	diff := CompareContent(from, to)
	if len(diff.Added) != 1 || diff.Added[0].BlockID != "image" {
		t.Fatalf("新增分类不正确：%#v", diff.Added)
	}
	if len(diff.Deleted) != 1 || diff.Deleted[0].BlockID != "link" {
		t.Fatalf("删除分类不正确：%#v", diff.Deleted)
	}
	if len(diff.Moved) != 2 || len(diff.Changed) != 1 || diff.Changed[0].Changes[0].Field != "heading_level" {
		t.Fatalf("移动或字段变化分类不正确：%#v", diff)
	}
}
