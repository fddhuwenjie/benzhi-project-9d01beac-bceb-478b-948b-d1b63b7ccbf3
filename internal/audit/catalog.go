package audit

type RuleDescriptor struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	AppliesTo   string `json:"applies_to"`
	Remediation string `json:"remediation"`
}

func Catalog() []RuleDescriptor {
	return []RuleDescriptor{
		{
			Code: "A11Y-HEADING-LEVEL", Name: "标题层级连续",
			Description: "相邻标题不得跳过层级，以保持辅助技术可理解的文档提纲。",
			Severity:    "error", AppliesTo: "heading",
			Remediation: "调整当前标题层级，或补充必要的中间层级标题。",
		},
		{
			Code: "A11Y-LINK-TEXT", Name: "链接文本可理解",
			Description: "链接必须包含能够独立说明目的的可见文本。",
			Severity:    "error", AppliesTo: "link",
			Remediation: "填写清晰的链接文本，避免使用空文本或仅依赖链接地址。",
		},
		{
			Code: "A11Y-IMAGE-ALT", Name: "图片替代文本",
			Description: "图片必须具有能够传达同等信息的替代文本。",
			Severity:    "error", AppliesTo: "image",
			Remediation: "为信息图片补充简洁、准确且符合上下文的替代文本。",
		},
		{
			Code: "A11Y-TABLE-HEADER", Name: "表格表头完整",
			Description: "数据表格的每一列必须具有非空表头，便于建立单元格关联。",
			Severity:    "error", AppliesTo: "table",
			Remediation: "补充所有列标题，并确保标题能够说明该列数据含义。",
		},
	}
}

func Descriptor(code string) (RuleDescriptor, bool) {
	for _, descriptor := range Catalog() {
		if descriptor.Code == code {
			return descriptor, true
		}
	}
	return RuleDescriptor{}, false
}
