package domain

import "time"

type CaseProgress struct {
	StageIndex        int    `json:"stage_index"`
	StageCount        int    `json:"stage_count"`
	StageLabel        string `json:"stage_label"`
	IssueTotal        int    `json:"issue_total"`
	IssueOpen         int    `json:"issue_open"`
	IssueSubmitted    int    `json:"issue_submitted"`
	IssueAccepted     int    `json:"issue_accepted"`
	IssueReturned     int    `json:"issue_returned"`
	CompletionPercent int    `json:"completion_percent"`
	NextAction        string `json:"next_action"`
	NextRole          string `json:"next_role"`
	Overdue           bool   `json:"overdue"`
}

func (c AcceptanceCase) Progress(now time.Time) CaseProgress {
	progress := CaseProgress{StageCount: 6, StageLabel: c.Status.Label(), IssueTotal: len(c.Issues)}
	switch c.Status {
	case StatusDraft:
		progress.StageIndex = 1
		if c.ContentRevision == 0 {
			progress.NextAction, progress.NextRole = "录入并保存结构化文档内容", "内容编辑"
		} else {
			progress.NextAction, progress.NextRole = "执行无障碍规则审查", "内容编辑"
		}
	case StatusAudited:
		progress.StageIndex = 2
		progress.NextAction, progress.NextRole = "开始问题整改", "内容编辑"
	case StatusRemediating:
		progress.StageIndex = 3
		if c.ContentRevision > c.AuditContentRevision {
			progress.NextAction, progress.NextRole = "对最新内容修订再次执行规则审查", "内容编辑"
		} else {
			progress.NextAction, progress.NextRole = "为全部问题提交最新内容版本的证据", "内容编辑"
		}
	case StatusReview:
		progress.StageIndex = 4
		if c.AllAcceptedLatest() {
			progress.NextAction, progress.NextRole = "批准验收案", "质量审核员"
		} else {
			progress.NextAction, progress.NextRole = "逐项复核整改证据", "质量审核员"
		}
	case StatusApproved:
		progress.StageIndex = 5
		progress.NextAction, progress.NextRole = "发布合格声明", "质量审核员"
	case StatusPublished:
		progress.StageIndex = 6
		progress.NextAction, progress.NextRole = "验收流程已关闭", ""
	}
	for _, issue := range c.Issues {
		switch issue.Status {
		case IssueOpen:
			progress.IssueOpen++
		case IssueSubmitted:
			progress.IssueSubmitted++
		case IssueAccepted:
			progress.IssueAccepted++
		case IssueReturned:
			progress.IssueReturned++
		}
	}
	if progress.IssueTotal > 0 {
		progress.CompletionPercent = progress.IssueAccepted * 100 / progress.IssueTotal
	} else if c.Status != StatusDraft {
		progress.CompletionPercent = 100
	}
	if deadline, err := time.Parse("2006-01-02", c.TargetPublishDate); err == nil {
		today := time.Date(now.UTC().Year(), now.UTC().Month(), now.UTC().Day(), 0, 0, 0, 0, time.UTC)
		progress.Overdue = c.Status != StatusPublished && deadline.Before(today)
	}
	return progress
}
