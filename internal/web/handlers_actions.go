package web

import (
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/workflow"
	"net/http"
)

func (s *Server) AuditCaseHandler(w http.ResponseWriter, r *http.Request) {
	var req mutationRequest
	if !decodeOrWrite(w, r, &req) {
		return
	}
	req.RequestID = requestID(r, req.RequestID)
	res, err := s.service.Audit(r.Context(), workflow.AuditCommand{CaseID: r.PathValue("id"), Revision: req.Revision, Actor: req.Actor, RequestID: req.RequestID})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 200, res)
}
func (s *Server) SubmitEvidenceHandler(w http.ResponseWriter, r *http.Request) {
	var req evidenceRequest
	if !decodeOrWrite(w, r, &req) {
		return
	}
	req.RequestID = requestID(r, req.RequestID)
	res, err := s.service.SubmitEvidence(r.Context(), workflow.SubmitEvidenceCommand{CaseID: r.PathValue("id"), IssueID: r.PathValue("issueID"), Revision: req.Revision, ContentRevision: req.ContentRevision, Description: req.Description, Text: req.Text, Actor: req.Actor, RequestID: req.RequestID})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 200, res)
}
func (s *Server) SubmitEvidenceBatchHandler(w http.ResponseWriter, r *http.Request) {
	var req evidenceBatchRequest
	if !decodeOrWrite(w, r, &req) {
		return
	}
	req.RequestID = requestID(r, req.RequestID)
	items := make([]workflow.EvidenceInput, len(req.Items))
	for i, item := range req.Items {
		items[i] = workflow.EvidenceInput{IssueID: item.IssueID, Description: item.Description, Text: item.Text, ContentRevision: item.ContentRevision}
	}
	res, err := s.service.SubmitEvidenceBatch(r.Context(), workflow.SubmitEvidenceBatchCommand{CaseID: r.PathValue("id"), Revision: req.Revision, ContentRevision: req.ContentRevision, Actor: req.Actor, RequestID: req.RequestID, Items: items})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}
func (s *Server) ReviewIssueHandler(w http.ResponseWriter, r *http.Request) {
	var req reviewRequest
	if !decodeOrWrite(w, r, &req) {
		return
	}
	req.RequestID = requestID(r, req.RequestID)
	res, err := s.service.ReviewIssue(r.Context(), workflow.ReviewIssueCommand{CaseID: r.PathValue("id"), IssueID: r.PathValue("issueID"), Revision: req.Revision, Accept: req.Accept, Reviewer: req.Reviewer, Comment: req.Comment, RequestID: req.RequestID})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 200, res)
}
func (s *Server) ReviewBatchHandler(w http.ResponseWriter, r *http.Request) {
	var req reviewBatchRequest
	if !decodeOrWrite(w, r, &req) {
		return
	}
	req.RequestID = requestID(r, req.RequestID)
	decisions := make([]workflow.ReviewDecision, len(req.Decisions))
	for i, item := range req.Decisions {
		decisions[i] = workflow.ReviewDecision{IssueID: item.IssueID, Accept: item.Accept, Comment: item.Comment}
	}
	res, err := s.service.ReviewBatch(r.Context(), workflow.ReviewBatchCommand{CaseID: r.PathValue("id"), Revision: req.Revision, Reviewer: req.Reviewer, RequestID: req.RequestID, Decisions: decisions})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (s *Server) TimelineHandler(w http.ResponseWriter, r *http.Request) {
	from, err := intQuery(r, "revision_from")
	if err != nil {
		writeError(w, err)
		return
	}
	to, err := intQuery(r, "revision_to")
	if err != nil {
		writeError(w, err)
		return
	}
	limit, err := intQuery(r, "limit")
	if err != nil {
		writeError(w, err)
		return
	}
	q := r.URL.Query()
	page, err := s.service.Timeline(r.Context(), r.PathValue("id"), workflow.TimelineQuery{EventType: q.Get("event_type"), Actor: q.Get("actor"), RevisionFrom: int64(from), RevisionTo: int64(to), Cursor: q.Get("cursor"), Limit: limit})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}
func (s *Server) ApproveCaseHandler(w http.ResponseWriter, r *http.Request) {
	var req mutationRequest
	if !decodeOrWrite(w, r, &req) {
		return
	}
	req.RequestID = requestID(r, req.RequestID)
	res, err := s.service.Approve(r.Context(), workflow.ApproveCommand{CaseID: r.PathValue("id"), Revision: req.Revision, Approver: req.Actor, RequestID: req.RequestID})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 200, res)
}
func (s *Server) PublishCaseHandler(w http.ResponseWriter, r *http.Request) {
	var req mutationRequest
	if !decodeOrWrite(w, r, &req) {
		return
	}
	req.RequestID = requestID(r, req.RequestID)
	res, err := s.service.Publish(r.Context(), workflow.PublishCommand{CaseID: r.PathValue("id"), Revision: req.Revision, Actor: req.Actor, RequestID: req.RequestID})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 200, res)
}
