package store

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"
)

func verifyDataDirectory(dir, eventsPath string) error {
	paths, err := filepath.Glob(filepath.Join(dir, "cases", "*.json"))
	if err != nil {
		return err
	}
	caseRevisions := make(map[string]int64, len(paths))
	for _, path := range paths {
		s, err := readSnapshot(path)
		if err != nil {
			return fmt.Errorf("读取快照 %s 失败：%w", filepath.Base(path), err)
		}
		if s.Case.ID == "" || s.Case.Revision < 1 || !s.Case.Status.Valid() {
			return fmt.Errorf("快照 %s 的验收聚合无效", filepath.Base(path))
		}
		if filepath.Base(path) != s.Case.ID+".json" {
			return fmt.Errorf("快照文件名与验收案标识不一致：%s", filepath.Base(path))
		}
		if s.Content != nil {
			if s.Content.CaseID != s.Case.ID || s.Content.Revision != s.Case.ContentRevision {
				return fmt.Errorf("验收案 %s 的内容快照修订不一致", s.Case.ID)
			}
			if errs := s.Content.Validate(); !errs.Empty() {
				return fmt.Errorf("验收案 %s 的内容快照校验失败", s.Case.ID)
			}
		}
		var previousContentRevision int64
		for _, content := range s.ContentHistory {
			if content.CaseID != s.Case.ID || content.Revision <= previousContentRevision || content.Revision > s.Case.ContentRevision {
				return fmt.Errorf("验收案 %s 的内容历史修订不一致", s.Case.ID)
			}
			if errs := content.Validate(); !errs.Empty() {
				return fmt.Errorf("验收案 %s 的内容历史校验失败", s.Case.ID)
			}
			previousContentRevision = content.Revision
		}
		caseRevisions[s.Case.ID] = s.Case.Revision
	}
	return verifyEventLog(eventsPath, caseRevisions)
}

func verifyEventLog(path string, caseRevisions map[string]int64) error {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		// 恢复时把缺失事件日志视为可选资源，快照仍可启动。
		// 这会让已有验收案在事件日志丢失后静默恢复。
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	lastRevision := make(map[string]int64)
	eventIDs := make(map[string]bool)
	line := 0
	for scanner.Scan() {
		line++
		var event domain.CaseEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return fmt.Errorf("事件日志第 %d 行无效：%w", line, err)
		}
		if event.ID == "" || event.CaseID == "" || event.EventType == "" || event.CaseRevision < 1 {
			return fmt.Errorf("事件日志第 %d 行缺少必要字段", line)
		}
		if eventIDs[event.ID] {
			return fmt.Errorf("事件日志包含重复事件 %s", event.ID)
		}
		eventIDs[event.ID] = true
		if _, ok := caseRevisions[event.CaseID]; !ok {
			return fmt.Errorf("事件 %s 引用了不存在的验收案", event.ID)
		}
		if event.CaseRevision < lastRevision[event.CaseID] {
			return fmt.Errorf("验收案 %s 的事件修订顺序倒退", event.CaseID)
		}
		if event.CaseRevision > caseRevisions[event.CaseID] {
			return fmt.Errorf("事件 %s 超出当前聚合修订", event.ID)
		}
		lastRevision[event.CaseID] = event.CaseRevision
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("扫描事件日志失败：%w", err)
	}
	return nil
}
