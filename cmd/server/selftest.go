package main

import (
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/workflow"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type smokeClient struct {
	base string
	http *http.Client
	seq  int
}

func (c *smokeClient) post(path string, body any, dst any) error {
	return c.request(http.MethodPost, path, body, dst)
}
func (c *smokeClient) request(method, path string, body any, dst any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(method, c.base+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%s %s 返回 %d：%s", method, path, resp.StatusCode, string(payload))
	}
	if dst != nil {
		return json.Unmarshal(payload, dst)
	}
	return nil
}
func (c *smokeClient) get(path string, dst any) error {
	resp, err := c.http.Get(c.base + path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GET %s 返回 %d：%s", path, resp.StatusCode, string(b))
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}
func (c *smokeClient) checkPage(path, marker string) error {
	resp, err := c.http.Get(c.base + path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s 返回 %d：%s", path, resp.StatusCode, string(payload))
	}
	if !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html") || !bytes.Contains(payload, []byte(marker)) {
		return fmt.Errorf("GET %s 未返回预期 HTML 声明页面", path)
	}
	return nil
}
func (c *smokeClient) rid() string { c.seq++; return fmt.Sprintf("selftest-%02d", c.seq) }

func runSelftest(cfg config) error {
	tmp, err := os.MkdirTemp("", "benzhi-a11y-selftest-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	app, err := buildHandler(tmp)
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return fmt.Errorf("自检监听失败：%w", err)
	}
	server := newHTTPServer(cfg.Addr, app.Handler())
	serveErr := make(chan error, 1)
	go func() { serveErr <- server.Serve(listener) }()
	baseURL := "http://" + listener.Addr().String()
	if u, parseErr := url.Parse(baseURL); parseErr != nil {
		return parseErr
	} else if u.Hostname() == "::1" {
		baseURL = "http://[::1]:" + u.Port()
	}
	client := &smokeClient{base: baseURL, http: &http.Client{Timeout: 5 * time.Second}}
	runErr := executeSmoke(client)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	shutdownErr := server.Shutdown(ctx)
	cancel()
	serverErr := <-serveErr
	if runErr != nil {
		return runErr
	}
	if shutdownErr != nil {
		return shutdownErr
	}
	if !errors.Is(serverErr, http.ErrServerClosed) {
		return serverErr
	}
	fmt.Println("自检通过：建档、内容审查、四项整改、逐项复核、批准、声明发布与导出均成功")
	return nil
}

func executeSmoke(c *smokeClient) error {
	var created workflow.MutationResult
	if err := c.post("/api/cases", map[string]any{"title": "公共服务办事指南", "organization": "市公共服务中心", "owner": "内容编辑", "target_publish_date": "2026-09-30", "actor": "内容编辑", "request_id": c.rid()}, &created); err != nil {
		return err
	}
	blocks := []domain.ContentBlock{{ID: "h1", Type: domain.BlockHeading, HeadingLevel: 1, Text: "办事指南"}, {ID: "h3", Type: domain.BlockHeading, HeadingLevel: 3, Text: "办理材料"}, {ID: "link", Type: domain.BlockLink, LinkTarget: "/apply"}, {ID: "image", Type: domain.BlockImage, Text: "办理流程图"}, {ID: "table", Type: domain.BlockTable, Text: "办理时限"}}
	var saved workflow.MutationResult
	if err := c.request(http.MethodPut, "/api/cases/"+created.CaseID+"/content", map[string]any{"revision": created.Revision, "actor": "内容编辑", "request_id": c.rid(), "blocks": blocks}, &saved); err != nil {
		return err
	}
	var audited workflow.MutationResult
	if err := c.post("/api/cases/"+created.CaseID+"/audit", map[string]any{"revision": saved.Revision, "actor": "内容编辑", "request_id": c.rid()}, &audited); err != nil {
		return err
	}
	var detail workflow.CaseDetail
	if err := c.get("/api/cases/"+created.CaseID, &detail); err != nil {
		return err
	}
	if len(detail.Case.Issues) != 4 {
		return fmt.Errorf("预期 4 个问题，实际得到 %d", len(detail.Case.Issues))
	}
	current := audited
	for _, issue := range detail.Case.Issues {
		var result workflow.MutationResult
		body := map[string]any{"revision": current.Revision, "content_revision": current.ContentRevision, "description": "已修正 " + issue.RuleCode, "text": "复查确认内容块 " + issue.BlockID + " 已符合规则要求", "actor": "内容编辑", "request_id": c.rid()}
		if err := c.post("/api/cases/"+created.CaseID+"/issues/"+issue.ID+"/evidence", body, &result); err != nil {
			return err
		}
		current = result
	}
	for _, issue := range detail.Case.Issues {
		var result workflow.MutationResult
		body := map[string]any{"revision": current.Revision, "accept": true, "reviewer": "质量审核员", "comment": "证据充分", "request_id": c.rid()}
		if err := c.post("/api/cases/"+created.CaseID+"/issues/"+issue.ID+"/review", body, &result); err != nil {
			return err
		}
		current = result
	}
	var approved workflow.MutationResult
	if err := c.post("/api/cases/"+created.CaseID+"/approve", map[string]any{"revision": current.Revision, "actor": "质量审核员", "request_id": c.rid()}, &approved); err != nil {
		return err
	}
	var published workflow.MutationResult
	if err := c.post("/api/cases/"+created.CaseID+"/publish", map[string]any{"revision": approved.Revision, "actor": "质量审核员", "request_id": c.rid()}, &published); err != nil {
		return err
	}
	if published.Status != domain.StatusPublished {
		return errors.New("发布后状态不正确")
	}
	var declaration domain.Declaration
	if err := c.get("/api/cases/"+created.CaseID+"/declaration", &declaration); err != nil {
		return err
	}
	if declaration.RuleVersion == "" || len(declaration.Dispositions) != 4 {
		return errors.New("声明数据不完整")
	}
	if err := c.checkPage("/cases/"+created.CaseID+"/declaration", "合格声明"); err != nil {
		return err
	}
	return nil
}
