package main

import (
	"sort"
)

func sortApplyItems(items []_ApplyItem) {
	sort.Sort(applyItems(items))
}

type applyItems []_ApplyItem

func (q applyItems) Less(i, j int) bool {
	return q[i].Priority > q[j].Priority
}
func (q applyItems) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}
func (q applyItems) Len() int { return len(q) }
