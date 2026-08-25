package store

import (
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/domain"
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
)

func (r *DiskRepository) appendEventsLocked(events []domain.CaseEvent) error {
	if len(events) == 0 {
		return nil
	}
	if r.eventsFile == nil {
		f, err := os.OpenFile(r.eventsPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			return err
		}
		r.eventsFile = f
	}
	w := bufio.NewWriter(r.eventsFile)
	for _, event := range events {
		b, err := json.Marshal(event)
		if err != nil {
			return err
		}
		if _, err = w.Write(append(b, '\n')); err != nil {
			return err
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}
	return r.eventsFile.Sync()
}

func readEvents(path, caseID string) ([]domain.CaseEvent, error) {
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return []domain.CaseEvent{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var events []domain.CaseEvent
	r := bufio.NewReader(f)
	for {
		line, readErr := r.ReadBytes('\n')
		if len(line) > 0 {
			var event domain.CaseEvent
			if err := json.Unmarshal(line, &event); err != nil {
				return nil, err
			}
			if event.CaseID == caseID {
				events = append(events, event)
			}
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return nil, readErr
		}
	}
	return events, nil
}
