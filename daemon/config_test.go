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

	_, err = ScanConfigs("../etc")
	if err != nil {
		t.Fatal(err)
	}
}
