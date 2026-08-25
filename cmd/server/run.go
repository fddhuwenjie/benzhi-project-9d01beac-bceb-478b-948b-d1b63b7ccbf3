package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{Addr: addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 20 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 1 << 20}
}

func serve(cfg config) error {
	app, err := buildHandler(cfg.DataDir)
	if err != nil {
		return err
	}
	server := newHTTPServer(cfg.Addr, app.Handler())
	listener, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return err
	}
	log.Printf("无障碍文档整改验收台已监听 http://%s", listener.Addr())
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)
	errCh := make(chan error, 1)
	go func() { errCh <- server.Serve(listener) }()
	select {
	case err = <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-stop:
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		return server.Shutdown(ctx)
	}
}
