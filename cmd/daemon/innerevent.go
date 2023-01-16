package main

import (
	"sync"
)

// implement idle and full events
type innerSource struct {
	IsInUser bool
}

func (innerSource) Scope() string { return "inner" }
func (s innerSource) Check(ids []string) []string {
	var ret []string
	for _, id := range ids {
		switch id {
		case "user":
			if s.IsInUser {
				ret = append(ret, id)
			}
		}
	}
	return ret
}

func (s *innerSource) MarkUser() {
	s.IsInUser = true
}

// implement snapshot apply events
type snapshotSource struct {
	lock   sync.Mutex
	loaded map[string]bool
}

func (s *snapshotSource) markLoaded(id string) {
	s.lock.Lock()
	s.loaded[id] = true
	s.lock.Unlock()
}

func (*snapshotSource) Scope() string { return "snapshot" }
func (s *snapshotSource) Check(ids []string) []string {
	var ret []string
	s.lock.Lock()
	for _, id := range ids {
		if s.loaded[id] {
			ret = append(ret, id)
		}
	}
	s.lock.Unlock()
	return ret
}
