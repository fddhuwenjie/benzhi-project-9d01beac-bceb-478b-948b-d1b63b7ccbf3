package main

import (
	"log"
	"os"
)

func main() {
	cfg, err := parseConfig(os.Args[1:])
	if err != nil {
		log.Printf("配置错误：%v", err)
		os.Exit(2)
	}
	if cfg.Selftest {
		err = runSelftest(cfg)
	} else {
		err = serve(cfg)
	}
	if err != nil {
		log.Printf("服务退出：%v", err)
		os.Exit(1)
	}
}
