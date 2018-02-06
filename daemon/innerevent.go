package main

import (
	"../events"
)

// implement idle and full events
type innerSource struct {
}

func init() {
	events.Register(&innerSource{})
}

func (innerSource) Scope() string               { return "inner" }
func (innerSource) Check(ids []string) []string { return []string{"idle"} }
