package main

import (
	"path/filepath"
	"testing"
)

func TestLoading(t *testing.T) {
	cfgDir := "../../etc"

	cfg, err := LoadConfig(filepath.Join(cfgDir, "basic.json"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Id != "BASIC" {
		t.Fatal("Result error")
	}

	cfg, err = LoadConfig(filepath.Join(cfgDir, "chrome.json"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Id != "google-chrome" {
		t.Fatal("Result error")
	}
	if cfg.Capture.Method[0].WMClass == "" {
		t.Fatal("Result error")
	}

	_, err = ScanConfigs(cfgDir)
	if err != nil {
		t.Fatal(err)
	}
}
