package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/audit"
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/store"
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/workflow"
)

func testHandler(t *testing.T) http.Handler {
	t.Helper()
	repo, err := store.NewDiskRepository(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	server, err := NewServer(workflow.NewService(repo, audit.NewEngine()))
	if err != nil {
		t.Fatal(err)
	}
	return server.Handler()
}

func TestWorkbenchAndRuleCatalogAreServed(t *testing.T) {
	handler := testHandler(t)
	for _, path := range []string{"/", "/assets/app.css", "/assets/app.js", "/api/rules", "/api/dashboard", "/healthz", "/readyz"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		if res.Code != http.StatusOK || res.Body.Len() == 0 {
			t.Errorf("GET %s 返回 %d，长度 %d", path, res.Code, res.Body.Len())
		}
	}
}

func TestCreateCaseProtocolAndConflictResponse(t *testing.T) {
	handler := testHandler(t)
	body := map[string]any{"title": "指南", "organization": "机构", "owner": "编辑", "target_publish_date": "2026-10-01", "actor": "编辑", "request_id": "create-http"}
	b, _ := json.Marshal(body)
	request := func(contentType string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/cases", bytes.NewReader(b))
		req.Header.Set("Content-Type", contentType)
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		return res
	}
	if res := request("text/plain"); res.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("错误 Content-Type 返回 %d", res.Code)
	}
	first := request("application/json")
	if first.Code != http.StatusCreated {
		t.Fatalf("创建返回 %d：%s", first.Code, first.Body.String())
	}
	second := request("application/json; charset=utf-8")
	if second.Code != http.StatusCreated || second.Body.String() != first.Body.String() {
		t.Fatalf("重复建档没有返回相同结果：%s / %s", first.Body.String(), second.Body.String())
	}
}
