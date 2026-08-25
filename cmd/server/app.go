package main

import (
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/audit"
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/store"
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/web"
	"benzhi-project-9d01beac-bceb-478b-948b-d1b63b7ccbf3/internal/workflow"
)

func buildHandler(dataDir string) (*web.Server, error) {
	repo, err := store.NewDiskRepository(dataDir)
	if err != nil {
		return nil, err
	}
	service := workflow.NewService(repo, audit.NewEngine())
	return web.NewServer(service)
}
