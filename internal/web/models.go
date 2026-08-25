package web

import "benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"

type createCaseRequest struct {
	Title             string `json:"title"`
	Organization      string `json:"organization"`
	Owner             string `json:"owner"`
	TargetPublishDate string `json:"target_publish_date"`
	Actor             string `json:"actor"`
	RequestID         string `json:"request_id"`
}
type updateCaseRequest struct {
	Revision          int64  `json:"revision"`
	Title             string `json:"title"`
	Organization      string `json:"organization"`
	Owner             string `json:"owner"`
	TargetPublishDate string `json:"target_publish_date"`
	Actor             string `json:"actor"`
	RequestID         string `json:"request_id"`
}
type mutationRequest struct {
	Revision  int64  `json:"revision"`
	Actor     string `json:"actor"`
	RequestID string `json:"request_id"`
}
type contentRequest struct {
	Revision  int64                 `json:"revision"`
	Actor     string                `json:"actor"`
	RequestID string                `json:"request_id"`
	Blocks    []domain.ContentBlock `json:"blocks"`
}
type evidenceRequest struct {
	Revision        int64  `json:"revision"`
	ContentRevision int64  `json:"content_revision"`
	Description     string `json:"description"`
	Text            string `json:"text"`
	Actor           string `json:"actor"`
	RequestID       string `json:"request_id"`
}
type reviewRequest struct {
	Revision  int64  `json:"revision"`
	Accept    bool   `json:"accept"`
	Reviewer  string `json:"reviewer"`
	Comment   string `json:"comment"`
	RequestID string `json:"request_id"`
}
type restoreContentRequest struct {
	Revision       int64  `json:"revision"`
	SourceRevision int64  `json:"source_revision"`
	Actor          string `json:"actor"`
	RequestID      string `json:"request_id"`
}
type evidenceBatchItemRequest struct {
	IssueID         string `json:"issue_id"`
	Description     string `json:"description"`
	Text            string `json:"text"`
	ContentRevision int64  `json:"content_revision,omitempty"`
}
type evidenceBatchRequest struct {
	Revision        int64                      `json:"revision"`
	ContentRevision int64                      `json:"content_revision"`
	Actor           string                     `json:"actor"`
	RequestID       string                     `json:"request_id"`
	Items           []evidenceBatchItemRequest `json:"items"`
}
type reviewDecisionRequest struct {
	IssueID string `json:"issue_id"`
	Accept  bool   `json:"accept"`
	Comment string `json:"comment"`
}
type reviewBatchRequest struct {
	Revision  int64                   `json:"revision"`
	Reviewer  string                  `json:"reviewer"`
	RequestID string                  `json:"request_id"`
	Decisions []reviewDecisionRequest `json:"decisions"`
}
