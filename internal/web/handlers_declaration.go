package web

import "net/http"

func (s *Server) DeclarationPreviewHandler(w http.ResponseWriter, r *http.Request) {
	d, err := s.service.DeclarationPreview(r.Context(), r.PathValue("id"), r.URL.Query().Get("approver"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, d)
}

func (s *Server) DeclarationJSONHandler(w http.ResponseWriter, r *http.Request) {
	d, err := s.service.Declaration(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Content-Disposition", "attachment; filename=accessibility-declaration.json")
	writeJSON(w, 200, d)
}
func (s *Server) DeclarationPageHandler(w http.ResponseWriter, r *http.Request) {
	d, err := s.service.Declaration(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err = s.declaration.ExecuteTemplate(w, "declaration.html", d); err != nil {
		http.Error(w, "声明页面生成失败", 500)
	}
}
