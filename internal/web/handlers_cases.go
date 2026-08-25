package web

import (
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/workflow"
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) ListCasesHandler(w http.ResponseWriter, r *http.Request) {
	limit, err := intQuery(r, "limit")
	if err != nil {
		writeError(w, err)
		return
	}
	q := r.URL.Query()
	dateFrom := firstQuery(q.Get("date_from"), q.Get("target_publish_date_from"))
	dateTo := firstQuery(q.Get("date_to"), q.Get("target_publish_date_to"))
	items, err := s.service.SearchCases(r.Context(), workflow.CaseListQuery{Keyword: q.Get("q"), Owner: q.Get("owner"), Status: q.Get("status"), DateFrom: dateFrom, DateTo: dateTo, Sort: q.Get("sort"), Order: q.Get("order"), Cursor: q.Get("cursor"), Limit: limit})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 200, items)
}
func (s *Server) GetCaseHandler(w http.ResponseWriter, r *http.Request) {
	detail, err := s.service.GetCase(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 200, detail)
}

func (s *Server) ContentHistoryHandler(w http.ResponseWriter, r *http.Request) {
	items, err := s.service.ContentHistory(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"revisions": items})
}

func (s *Server) ContentDifferenceHandler(w http.ResponseWriter, r *http.Request) {
	from, err := requiredIntQuery(r, "from")
	if err != nil {
		writeError(w, err)
		return
	}
	to, err := requiredIntQuery(r, "to")
	if err != nil {
		writeError(w, err)
		return
	}
	difference, err := s.service.CompareContent(r.Context(), r.PathValue("id"), int64(from), int64(to))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, difference)
}

func (s *Server) RestoreContentHandler(w http.ResponseWriter, r *http.Request) {
	var req restoreContentRequest
	if !decodeOrWrite(w, r, &req) {
		return
	}
	req.RequestID = requestID(r, req.RequestID)
	result, err := s.service.RestoreContent(r.Context(), workflow.RestoreContentCommand{CaseID: r.PathValue("id"), Revision: req.Revision, SourceRevision: req.SourceRevision, Actor: req.Actor, RequestID: req.RequestID})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func intQuery(r *http.Request, name string) (int, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(name))
	if raw == "" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, domain.ValidationErrors{{Field: name, Message: name + " 必须为整数"}}
	}
	return value, nil
}

func requiredIntQuery(r *http.Request, name string) (int, error) {
	value, err := intQuery(r, name)
	if err != nil {
		return 0, err
	}
	if value < 1 {
		return 0, domain.ValidationErrors{{Field: name, Message: name + " 必须为正整数"}}
	}
	return value, nil
}

func firstQuery(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
func (s *Server) CreateCaseHandler(w http.ResponseWriter, r *http.Request) {
	var req createCaseRequest
	if !decodeOrWrite(w, r, &req) {
		return
	}
	req.RequestID = requestID(r, req.RequestID)
	result, err := s.service.CreateCase(r.Context(), workflow.CreateCaseCommand{Title: req.Title, Organization: req.Organization, Owner: req.Owner, TargetPublishDate: req.TargetPublishDate, Actor: req.Actor, RequestID: req.RequestID})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 201, result)
}
func (s *Server) UpdateCaseHandler(w http.ResponseWriter, r *http.Request) {
	var req updateCaseRequest
	if !decodeOrWrite(w, r, &req) {
		return
	}
	req.RequestID = requestID(r, req.RequestID)
	result, err := s.service.UpdateCase(r.Context(), workflow.UpdateCaseCommand{
		CaseID: r.PathValue("id"), Revision: req.Revision, Title: req.Title,
		Organization: req.Organization, Owner: req.Owner, TargetPublishDate: req.TargetPublishDate,
		Actor: req.Actor, RequestID: req.RequestID,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}
func (s *Server) SaveContentHandler(w http.ResponseWriter, r *http.Request) {
	var req contentRequest
	if !decodeOrWrite(w, r, &req) {
		return
	}
	req.RequestID = requestID(r, req.RequestID)
	result, err := s.service.SaveContent(r.Context(), workflow.SaveContentCommand{CaseID: r.PathValue("id"), Revision: req.Revision, Actor: req.Actor, RequestID: req.RequestID, Blocks: req.Blocks})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, 200, result)
}
