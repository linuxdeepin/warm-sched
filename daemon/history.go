package main

import (
	"../core"
	"../events"
	"context"
	"fmt"
	"path"
	"sort"
	"sync"
	"time"
)

type SnapStatus struct {
	Id            string
	Description   string
	Captured      bool
	Summary       string
	LoadCount     int //实际被加载次数
	OccurCount    int //满足被加载条件次数
	ConfigurePath string
}

func (h *History) SnapStatus() []SnapStatus {
	var ret []SnapStatus
	for _, v := range h.status {
		ret = append(ret, v)
	}
	return ret
}

type _ApplyItem struct {
	Id       string
	Priority int
}

type History struct {
	status   map[string]SnapStatus
	cacheDir string
	ss       *snapshotSource

	usage     map[string]int
	usageLock sync.Mutex

	applyLock  sync.Mutex
	applyQueue []_ApplyItem
}

func NewHistory(cache string) *History {
	ss := &snapshotSource{
		loaded: make(map[string]bool),
	}
	events.Register(ss)

	h := &History{
		cacheDir: cache,
		status:   make(map[string]SnapStatus),
		ss:       ss,
		usage:    make(map[string]int),
	}
	h.loadUsage()

	go h.poll(context.TODO())
	return h
}

func (h *History) has(id string) bool    { return FileExist(h.path(id)) }
func (h *History) path(id string) string { return path.Join(h.cacheDir, "snap", id) }
func (h *History) hpath() string         { return path.Join(h.cacheDir, "history") }

func (h *History) loadUsage() {
	h.usageLock.Lock()
	core.LoadFrom(h.hpath(), &h.usage)
	h.usageLock.Unlock()
}
func (h *History) storeUsage() {
	h.usageLock.Lock()
	core.StoreTo(h.hpath(), h.usage)
	h.usageLock.Unlock()
}
func (h *History) addUsage(id string) {
	h.usageLock.Lock()
	h.usage[id]++
	h.usageLock.Unlock()
	h.storeUsage()
}
func (h *History) Usage(id string) int {
	h.usageLock.Lock()
	v := h.usage[id]
	h.usageLock.Unlock()
	return v
}

func (h *History) DoCapture(id string, force bool, methods []*core.CaptureMethod) error {
	h.addUsage(id)
	if h.has(id) && !force {
		Log("Ignore capture %q because already has one sample.\n", id)
		return nil
	}

	snap, err := core.CaptureSnapshot(methods...)
	if err != nil {
		return fmt.Errorf("DoCapture %q failed: %v", id, err)
	}
	return core.StoreTo(h.path(id), snap)
}

func (h *History) RequestApply(id string, initUsage int) error {
	if !h.has(id) {
		Log("Ignore apply %q because hasn't any samples.\n", id)
		return nil
	}

	h.applyLock.Lock()
	h.applyQueue = append(h.applyQueue, _ApplyItem{
		Id:       id,
		Priority: initUsage + h.Usage(id),
	})
	h.applyLock.Unlock()
	return nil
}

func (h *History) poll(ctx context.Context) {
	for {
		select {
		case <-time.After(time.Second):
			h.handleApply()
		case <-ctx.Done():
			return
		}
	}
}

func (h *History) handleApply() error {
	h.applyLock.Lock()
	n := len(h.applyQueue)
	h.applyLock.Unlock()
	if n == 0 {
		return nil
	}

	h.applyLock.Lock()
	sort.Slice(h.applyQueue, func(i, j int) bool {
		return h.applyQueue[i].Priority > h.applyQueue[j].Priority
	})
	id := h.applyQueue[0].Id
	h.applyQueue = h.applyQueue[1:]
	h.applyLock.Unlock()

	Log("Begin DoApply %q\n", id)
	err := h.doApply(id)
	if err != nil {
		Log("End DoApply %q failed: %v\n", id, err)
	} else {
		Log("End DoApply %q\n", id)
	}
	return nil
}

func (h *History) doApply(id string) error {
	var snap core.Snapshot
	err := core.LoadFrom(h.path(id), &snap)
	if err != nil {
		return err
	}
	err = core.ApplySnapshot(&snap, true)
	if err != nil {
		return err
	}
	h.ss.markLoaded(id)
	return nil
}
