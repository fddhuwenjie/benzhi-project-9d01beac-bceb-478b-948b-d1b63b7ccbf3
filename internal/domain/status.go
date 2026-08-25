package domain

type CaseStatus string

const (
	StatusDraft       CaseStatus = "draft"
	StatusAudited     CaseStatus = "audited"
	StatusRemediating CaseStatus = "remediating"
	StatusReview      CaseStatus = "review"
	StatusApproved    CaseStatus = "approved"
	StatusPublished   CaseStatus = "published"
)

var statusLabels = map[CaseStatus]string{
	StatusDraft: "草稿", StatusAudited: "已审查", StatusRemediating: "整改中",
	StatusReview: "待复核", StatusApproved: "已批准", StatusPublished: "已发布",
}

func (s CaseStatus) Label() string { return statusLabels[s] }

func (s CaseStatus) Valid() bool { _, ok := statusLabels[s]; return ok }

func CanTransition(from, to CaseStatus) bool {
	allowed := map[CaseStatus][]CaseStatus{
		StatusDraft: {StatusAudited}, StatusAudited: {StatusRemediating},
		StatusRemediating: {StatusReview}, StatusReview: {StatusRemediating, StatusApproved},
		StatusApproved: {StatusPublished},
	}
	for _, candidate := range allowed[from] {
		if candidate == to {
			return true
		}
	}
	return false
}
