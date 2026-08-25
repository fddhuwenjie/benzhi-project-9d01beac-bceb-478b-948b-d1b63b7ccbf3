package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type config struct {
	Addr     string
	DataDir  string
	Selftest bool
}

func parseConfig(args []string) (config, error) {
	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	addr := fs.String("addr", "127.0.0.1:19081", "HTTP 监听地址")
	data := fs.String("data", "data", "持久化数据目录")
	selftest := fs.Bool("selftest", false, "执行真实 HTTP 自检后退出")
	if err := fs.Parse(args); err != nil {
		return config{}, err
	}
	if fs.NArg() != 0 {
		return config{}, errors.New("存在未识别的位置参数")
	}
	explicit := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "addr" {
			explicit = true
		}
	})
	if !explicit {
		if raw := strings.TrimSpace(os.Getenv("PORT")); raw != "" {
			port, err := strconv.Atoi(raw)
			if err != nil || port < 1 || port > 65535 {
				return config{}, fmt.Errorf("PORT 必须为 1 到 65535 的端口号")
			}
			*addr = net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
		}
	}
	if err := validateAddr(*addr); err != nil {
		return config{}, err
	}
	if strings.TrimSpace(*data) == "" {
		return config{}, errors.New("数据目录不能为空")
	}
	return config{Addr: *addr, DataDir: *data, Selftest: *selftest}, nil
}

func validateAddr(addr string) error {
	host, portText, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("无效的 -addr：%w", err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil || port < 1 || port > 65535 {
		return errors.New("-addr 端口必须为 1 到 65535")
	}
	if host != "127.0.0.1" && host != "localhost" && host != "::1" {
		return errors.New("-addr 必须使用回环地址 127.0.0.1、localhost 或 ::1")
	}
	return nil
}
