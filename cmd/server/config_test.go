package main

import "testing"

func TestParseConfigDefaultsAndExplicitAddress(t *testing.T) {
	t.Setenv("PORT", "")
	cfg, err := parseConfig(nil)
	if err != nil || cfg.Addr != "127.0.0.1:19081" {
		t.Fatalf("默认配置错误：%#v %v", cfg, err)
	}
	t.Setenv("PORT", "19444")
	cfg, err = parseConfig(nil)
	if err != nil || cfg.Addr != "127.0.0.1:19444" {
		t.Fatalf("PORT 配置错误：%#v %v", cfg, err)
	}
	cfg, err = parseConfig([]string{"-addr=127.0.0.1:19555"})
	if err != nil || cfg.Addr != "127.0.0.1:19555" {
		t.Fatalf("显式地址未优先：%#v %v", cfg, err)
	}
}

func TestParseConfigRejectsUnsafeAddress(t *testing.T) {
	t.Setenv("PORT", "")
	if _, err := parseConfig([]string{"-addr=0.0.0.0:19081"}); err == nil {
		t.Fatal("应拒绝非回环监听地址")
	}
	if _, err := parseConfig([]string{"-addr=127.0.0.1:0"}); err == nil {
		t.Fatal("应拒绝零端口")
	}
}
