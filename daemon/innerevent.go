package main

import (
	"../events"
)

type innerSource struct {
}

func init() {
	events.Register(&innerSource{})
}

func (innerSource) Scope() string              { return "inner" }
func (innerSource) Prepare(ids []string) error { return nil }
func (innerSource) Stop()                      {}
func (s innerSource) Run() error               { return nil }
