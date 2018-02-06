package main

import (
	"../events"
)

// implement idle and full events
type innerSource struct {
}

func init() {
	events.Register(&innerSource{})
	events.Register(&snapshotSource{})
}

func (innerSource) Scope() string               { return "inner" }
func (innerSource) Prepare(ids []string) error  { return nil }
func (innerSource) Check(ids []string) []string { return nil }
func (innerSource) Stop()                       {}
func (s innerSource) Run() error                { events.Emit(s.Scope(), "idle"); return nil }
