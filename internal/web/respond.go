package web

import (
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

const maxBody = 1 << 20

type errorEnvelope struct {
	Error apiError `json:"error"`
}
type apiError struct {
	Code    string                  `json:"code"`
	Message string                  `json:"message"`
	Fields  domain.ValidationErrors `json:"fields,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func writeError(w http.ResponseWriter, err error) {
	status, code, message := http.StatusInternalServerError, "internal_error", "服务暂时无法处理请求"
	switch {
	case errors.Is(err, domain.ErrNotFound):
		status, code, message = 404, "not_found", err.Error()
	case errors.Is(err, domain.ErrConflict):
		status, code, message = 409, "revision_conflict", err.Error()
	case errors.Is(err, domain.ErrStaleEvidence):
		status, code, message = 409, "stale_evidence", err.Error()
	case errors.Is(err, domain.ErrIntegrity):
		status, code, message = 409, "integrity_error", err.Error()
	case errors.Is(err, domain.ErrNotReady), errors.Is(err, domain.ErrInvalidTransition):
		status, code, message = 422, "not_ready", err.Error()
	case errors.Is(err, domain.ErrValidation):
		status, code, message = 422, "validation_error", err.Error()
	}
	env := errorEnvelope{Error: apiError{Code: code, Message: message}}
	var fields domain.ValidationErrors
	if errors.As(err, &fields) {
		env.Error.Fields = fields
	}
	writeJSON(w, status, env)
}
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	if !strings.HasPrefix(strings.ToLower(r.Header.Get("Content-Type")), "application/json") {
		return &protocolError{status: 415, message: "请求 Content-Type 必须为 application/json"}
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBody)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return &protocolError{status: 400, message: "JSON 请求体无效：" + err.Error()}
	}
	var extra any
	if err := dec.Decode(&extra); err == nil {
		return &protocolError{status: 400, message: "请求体只能包含一个 JSON 对象"}
	}
	return nil
}

type protocolError struct {
	status  int
	message string
}

func (e *protocolError) Error() string { return e.message }
func decodeOrWrite(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := decodeJSON(w, r, dst); err != nil {
		if p, ok := err.(*protocolError); ok {
			writeJSON(w, p.status, errorEnvelope{Error: apiError{Code: "bad_request", Message: p.message}})
		} else {
			writeError(w, err)
		}
		return false
	}
	return true
}
