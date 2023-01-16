package main

import (
	"testing"
)

func TestSortApplyItem(t *testing.T) {
	data := []_ApplyItem{
		{"A", 1},
		{"B", 4},
		{"C", 2},
	}

	sortApplyItems(data)
	if data[0].Id != "B" || data[2].Id != "A" || len(data) != 3 {
		t.Fatal(data)
	}
}
