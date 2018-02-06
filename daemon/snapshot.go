package main

import (
	"../core"
	"fmt"
	"sync"
)

type History struct {
}

// implement snapshot apply events
type snapshotSource struct {
	lock   sync.Mutex
	loaded map[string]bool
}

func (*snapshotSource) Scope() string              { return "snapshot" }
func (*snapshotSource) Prepare(ids []string) error { return nil }

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
func (*snapshotSource) Stop()        {}
func (s *snapshotSource) Run() error { return nil }

func (*Daemon) DoCapture(id string, methods []*core.CaptureMethod) error {
	snap, err := core.CaptureSnapshot(methods...)
	if err != nil {
		return fmt.Errorf("DoCapture %q failed: %v", id, err)
	}
	Log("Capture %q....%v\n", id, snap)
	return nil
}

func (*Daemon) DoApply(id string) error {
	panic("Not Implement")
}
