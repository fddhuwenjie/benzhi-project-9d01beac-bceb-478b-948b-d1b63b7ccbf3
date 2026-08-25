package web

import (
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/workflow"
	"embed"
	"html/template"
	"net/http"
	"strings"
	"time"
)

//go:embed ui/*
var uiFiles embed.FS

type Server struct {
	service     *workflow.Service
	mux         *http.ServeMux
	declaration *template.Template
	activity    *requestActivity
}

func NewServer(service *workflow.Service) (*Server, error) {
	t, err := template.ParseFS(uiFiles, "ui/declaration.html")
	if err != nil {
		return nil, err
	}
	s := &Server{service: service, mux: http.NewServeMux(), declaration: t, activity: newRequestActivity()}
	s.routes()
	return s, nil
}

func (s *Server) Handler() http.Handler { return requestLog(s.activity, securityHeaders(s.mux)) }

func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", s.HealthHandler)
	s.mux.HandleFunc("GET /readyz", s.ReadyHandler)
	s.mux.HandleFunc("GET /", s.WorkbenchHandler)
	s.mux.HandleFunc("GET /assets/app.css", s.CSSHandler)
	s.mux.HandleFunc("GET /assets/app.js", s.JSHandler)
	s.mux.HandleFunc("GET /api/cases", s.ListCasesHandler)
	s.mux.HandleFunc("GET /api/dashboard", s.DashboardHandler)
	s.mux.HandleFunc("GET /api/rules", s.RuleCatalogHandler)
	s.mux.HandleFunc("POST /api/cases", s.CreateCaseHandler)
	s.mux.HandleFunc("GET /api/cases/{id}", s.GetCaseHandler)
	s.mux.HandleFunc("PATCH /api/cases/{id}", s.UpdateCaseHandler)
	s.mux.HandleFunc("PUT /api/cases/{id}/content", s.SaveContentHandler)
	s.mux.HandleFunc("GET /api/cases/{id}/content/revisions", s.ContentHistoryHandler)
	s.mux.HandleFunc("GET /api/cases/{id}/content/diff", s.ContentDifferenceHandler)
	s.mux.HandleFunc("POST /api/cases/{id}/content/restore", s.RestoreContentHandler)
	s.mux.HandleFunc("POST /api/cases/{id}/audit", s.AuditCaseHandler)
	s.mux.HandleFunc("POST /api/cases/{id}/evidence/batch", s.SubmitEvidenceBatchHandler)
	s.mux.HandleFunc("POST /api/cases/{id}/issues/{issueID}/evidence", s.SubmitEvidenceHandler)
	s.mux.HandleFunc("POST /api/cases/{id}/reviews/batch", s.ReviewBatchHandler)
	s.mux.HandleFunc("POST /api/cases/{id}/issues/{issueID}/review", s.ReviewIssueHandler)
	s.mux.HandleFunc("GET /api/cases/{id}/timeline", s.TimelineHandler)
	s.mux.HandleFunc("GET /api/cases/{id}/declaration/preview", s.DeclarationPreviewHandler)
	s.mux.HandleFunc("POST /api/cases/{id}/approve", s.ApproveCaseHandler)
	s.mux.HandleFunc("POST /api/cases/{id}/publish", s.PublishCaseHandler)
	s.mux.HandleFunc("GET /api/cases/{id}/declaration", s.DeclarationJSONHandler)
	s.mux.HandleFunc("GET /cases/{id}/declaration", s.DeclarationPageHandler)
}

func serveEmbedded(w http.ResponseWriter, name, contentType string) {
	b, err := uiFiles.ReadFile(name)
	if err != nil {
		http.NotFound(w, nil)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write(b)
}

func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "time": time.Now().UTC()})
}
func (s *Server) ReadyHandler(w http.ResponseWriter, r *http.Request) {
	_, err := s.service.ListCases(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}
func (s *Server) WorkbenchHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	serveEmbedded(w, "ui/index.html", "text/html; charset=utf-8")
}
func (s *Server) CSSHandler(w http.ResponseWriter, r *http.Request) {
	serveEmbedded(w, "ui/app.css", "text/css; charset=utf-8")
}
func (s *Server) JSHandler(w http.ResponseWriter, r *http.Request) {
	serveEmbedded(w, "ui/app.js", "text/javascript; charset=utf-8")
}

func requestID(r *http.Request, body string) string {
	if v := strings.TrimSpace(r.Header.Get("Idempotency-Key")); v != "" {
		return v
	}
	return strings.TrimSpace(body)
}
