package web

import (
	"net/http"

	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/audit"
)

func (s *Server) DashboardHandler(w http.ResponseWriter, r *http.Request) {
	dashboard, err := s.service.Dashboard(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dashboard)
}

func (s *Server) RuleCatalogHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"version": audit.RuleVersion,
		"rules":   audit.Catalog(),
	})
}
