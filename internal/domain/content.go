package domain

import (
	"strings"
	"time"
)

type BlockType string

const (
	BlockHeading   BlockType = "heading"
	BlockParagraph BlockType = "paragraph"
	BlockLink      BlockType = "link"
	BlockImage     BlockType = "image"
	BlockTable     BlockType = "table"
)

type ContentBlock struct {
	ID           string    `json:"id"`
	Type         BlockType `json:"type"`
	HeadingLevel int       `json:"heading_level,omitempty"`
	Text         string    `json:"text,omitempty"`
	LinkTarget   string    `json:"link_target,omitempty"`
	AltText      string    `json:"alt_text,omitempty"`
	TableHeaders []string  `json:"table_headers,omitempty"`
}

type DocumentContent struct {
	CaseID     string         `json:"case_id"`
	Revision   int64          `json:"revision"`
	Blocks     []ContentBlock `json:"blocks"`
	SavedBy    string         `json:"saved_by"`
	BlockCount int            `json:"block_count"`
	SavedAt    time.Time      `json:"saved_at"`
}

func (c DocumentContent) Validate() ValidationErrors {
	var errs ValidationErrors
	if len(c.Blocks) == 0 {
		errs = append(errs, FieldError{Field: "blocks", Message: "至少需要一个内容块"})
	}
	seen := make(map[string]bool)
	for i, block := range c.Blocks {
		field := "blocks[" + itoa(i) + "]"
		if strings.TrimSpace(block.ID) == "" {
			errs = append(errs, FieldError{Field: field + ".id", Message: "内容块标识不能为空"})
		}
		if seen[block.ID] {
			errs = append(errs, FieldError{Field: field + ".id", Message: "内容块标识不能重复"})
		}
		seen[block.ID] = true
		switch block.Type {
		case BlockHeading:
			if block.HeadingLevel < 1 || block.HeadingLevel > 6 {
				errs = append(errs, FieldError{Field: field + ".heading_level", Message: "标题层级必须在 1 到 6 之间"})
			}
			if strings.TrimSpace(block.Text) == "" {
				errs = append(errs, FieldError{Field: field + ".text", Message: "标题文本不能为空"})
			}
		case BlockParagraph:
			if strings.TrimSpace(block.Text) == "" {
				errs = append(errs, FieldError{Field: field + ".text", Message: "段落文本不能为空"})
			}
		case BlockLink:
			if strings.TrimSpace(block.LinkTarget) == "" {
				errs = append(errs, FieldError{Field: field + ".link_target", Message: "链接目标不能为空"})
			}
		case BlockImage, BlockTable:
		default:
			errs = append(errs, FieldError{Field: field + ".type", Message: "不支持的内容块类型"})
		}
	}
	return errs
}

func NewDocumentContent(caseID, actor string, revision int64, blocks []ContentBlock, now time.Time) DocumentContent {
	return DocumentContent{CaseID: caseID, Revision: revision, Blocks: blocks, SavedBy: strings.TrimSpace(actor), BlockCount: len(blocks), SavedAt: now.UTC()}
}

func (c DocumentContent) Clone() DocumentContent {
	out := c
	out.Blocks = make([]ContentBlock, len(c.Blocks))
	for i, b := range c.Blocks {
		out.Blocks[i] = b
		out.Blocks[i].TableHeaders = append([]string(nil), b.TableHeaders...)
	}
	return out
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for v > 0 {
		i--
		b[i] = byte('0' + v%10)
		v /= 10
	}
	return string(b[i:])
}
