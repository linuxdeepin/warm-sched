package main

import (
	"path/filepath"
	"strings"
)

var BlackDirectory = []string{"/sys", "/proc", "/dev", "/run", "/boot"}

func ShouldSkipDirectory(root string) bool {
	r, err := filepath.Abs(root)
	if err != nil {
		return true
	}
	for _, i := range BlackDirectory {
		if strings.HasPrefix(r, i) {
			return true
		}
	}
	return false
}
