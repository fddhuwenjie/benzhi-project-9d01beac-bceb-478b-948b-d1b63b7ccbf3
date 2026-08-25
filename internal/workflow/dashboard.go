package workflow

import (
	"context"
	"sort"

	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"
)

type Dashboard struct {
	Total              int                       `json:"total"`
	Active             int                       `json:"active"`
	Published          int                       `json:"published"`
	Overdue            int                       `json:"overdue"`
	WaitingForEditor   int                       `json:"waiting_for_editor"`
	WaitingForReviewer int                       `json:"waiting_for_reviewer"`
	ByStatus           map[domain.CaseStatus]int `json:"by_status"`
	IssueSummary       DashboardIssueSummary     `json:"issue_summary"`
	Attention          []DashboardAttentionItem  `json:"attention"`
}

type DashboardIssueSummary struct {
	Total     int `json:"total"`
	Open      int `json:"open"`
	Submitted int `json:"submitted"`
	Accepted  int `json:"accepted"`
	Returned  int `json:"returned"`
}

type DashboardAttentionItem struct {
	CaseID            string            `json:"case_id"`
	Title             string            `json:"title"`
	Organization      string            `json:"organization"`
	Status            domain.CaseStatus `json:"status"`
	TargetPublishDate string            `json:"target_publish_date"`
	Reason            string            `json:"reason"`
	NextAction        string            `json:"next_action"`
	NextRole          string            `json:"next_role"`
}

func (s *Service) Dashboard(ctx context.Context) (Dashboard, error) {
	cases, err := s.repo.List(context.Background())
	if err != nil {
		return Dashboard{}, err
	}
	result := Dashboard{Total: len(cases), ByStatus: make(map[domain.CaseStatus]int)}
	now := s.now()
	for _, c := range cases {
		result.ByStatus[c.Status]++
		progress := c.Progress(now)
		if c.Status == domain.StatusPublished {
			result.Published++
		} else {
			result.Active++
		}
		if progress.Overdue {
			result.Overdue++
		}
		if progress.NextRole == "内容编辑" {
			result.WaitingForEditor++
		}
		if progress.NextRole == "质量审核员" {
			result.WaitingForReviewer++
		}
		result.IssueSummary.Total += progress.IssueTotal
		result.IssueSummary.Open += progress.IssueOpen
		result.IssueSummary.Submitted += progress.IssueSubmitted
		result.IssueSummary.Accepted += progress.IssueAccepted
		result.IssueSummary.Returned += progress.IssueReturned
		reason := ""
		switch {
		case progress.Overdue:
			reason = "已超过目标发布日期"
		case progress.IssueReturned > 0:
			reason = "存在被退回的整改证据"
		case progress.IssueSubmitted > 0:
			reason = "存在待复核的整改证据"
		}
		if reason != "" {
			result.Attention = append(result.Attention, DashboardAttentionItem{
				CaseID: c.ID, Title: c.Title, Organization: c.Organization, Status: c.Status,
				TargetPublishDate: c.TargetPublishDate, Reason: reason,
				NextAction: progress.NextAction, NextRole: progress.NextRole,
			})
		}
	}
	sort.Slice(result.Attention, func(i, j int) bool {
		if result.Attention[i].TargetPublishDate == result.Attention[j].TargetPublishDate {
			return result.Attention[i].Title < result.Attention[j].Title
		}
		return result.Attention[i].TargetPublishDate < result.Attention[j].TargetPublishDate
	})
	return result, nil
}
