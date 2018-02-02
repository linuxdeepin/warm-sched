package main

import (
	"testing"
)

func TestLoading(t *testing.T) {
	cfg, err := LoadConfig("../etc/basic.json")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Id != "BASIC" {
		t.Fatal("Result error")
	}

	cfg, err = LoadConfig("../etc/chrome.json")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Id != "google-chrome" {
		t.Fatal("Result error")
	}
	if cfg.Capture.Method[0].WMClass == "" {
		t.Fatal("Result error")
	}

	_, err = ScanConfigs("../etc")
	if err != nil {
		t.Fatal(err)
	}
}
