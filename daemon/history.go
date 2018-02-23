package main

import (
	"../core"
	"../events"
	"context"
	"fmt"
	"os"
	"path"
	"sort"
	"sync"
	"time"
)

type History struct {
	cacheDir string
	ss       *snapshotSource

	counts *HistoryCounts

	applyLock  sync.Mutex
	applyQueue []_ApplyItem
}

type HistoryStatus struct {
	CaptureCount     int
	ApplyCount       int
	LifetimeCount    int
	SnapSize         int
	ContentSize      int
	LoadedPercentage float32
}

func NewHistory(ctx context.Context, cache string) *History {
	ss := &snapshotSource{
		loaded: make(map[string]bool),
	}
	events.Register(ss)

	h := &History{
		cacheDir: cache,
		ss:       ss,
	}
	h.counts = NewHistoryCounts(h.hpath())

	go h.poll(ctx)
	return h
}

func (h *History) Status(id string, current *core.Snapshot) (HistoryStatus, *core.Snapshot) {
	snap := h.FindSnapshot(id)
	if snap == nil {
		return HistoryStatus{}, nil
	}

	cs, p := snap.AnalyzeSnapshotLoad(current)

	return HistoryStatus{
		CaptureCount:  h.counts.GetCapture(id),
		ApplyCount:    h.counts.GetApply(id),
		LifetimeCount: h.counts.GetLifetime(id),
		SnapSize:      h.getSnapSize(id),
		ContentSize:   cs,

		LoadedPercentage: p,
	}, snap
}

func (h *History) DoCapture(id string, c CaptureConfig) (err error) {
	h.counts.AddCapture(id)
	h.counts.SetLifetime(id, c.Lifetime)

	if h.has(id) {
		if !h.counts.IsDirty(id) {
			Log("Ignore capture %q because already has one sample.\n", id)
			return nil
		}
		Log("The snapshot of %q is dirty. so recapture one.\n", id)
	}

	Log("Begin DoCapture %q\n", id)
	defer func() {
		if err != nil {
			Log("End DoCapture %q failed: %v\n", id, err)
		} else {
			Log("End DoCapture %q\n", id)
		}
	}()

	if c.WaitSecond > 0 {
		Log("Wait %d Second.\n", c.WaitSecond)
		time.Sleep(time.Duration(c.WaitSecond) * time.Second)
	}

	snap, err := core.CaptureSnapshot(c.Method...)
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
		Priority: initUsage + h.counts.GetCapture(id),
	})
	h.applyLock.Unlock()
	return nil
}

func (h *History) FindSnapshot(id string) *core.Snapshot {
	var snap core.Snapshot
	err := core.LoadFrom(h.path(id), &snap)
	if err != nil {
		return nil
	}
	return &snap
}

type _ApplyItem struct {
	Id       string
	Priority int
}

func (h *History) has(id string) bool    { return FileExist(h.path(id)) }
func (h *History) path(id string) string { return path.Join(h.cacheDir, "snap", id) }
func (h *History) hpath() string         { return path.Join(h.cacheDir, "history") }

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

func (h *History) getSnapSize(id string) int {
	info, err := os.Stat(h.path(id))
	if err != nil {
		return 0
	}
	return int(info.Size())
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
	h.counts.AddApply(id)
	h.ss.markLoaded(id)
	return nil
}
