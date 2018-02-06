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

func (innerSource) Scope() string              { return "inner" }
func (innerSource) Prepare(ids []string) error { return nil }
func (innerSource) Stop()                      {}
func (s innerSource) Run() error               { events.Emit(s.Scope(), "idle"); return nil }

// implement snapshot apply events
type snapshotSource struct {
}

func (snapshotSource) Scope() string              { return "snapshot" }
func (snapshotSource) Prepare(ids []string) error { panic("not implement") }
func (snapshotSource) Stop()                      {}
func (s snapshotSource) Run() error               { return nil }
