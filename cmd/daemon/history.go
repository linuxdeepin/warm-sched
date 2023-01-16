package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"github.com/linuxdeepin/warm-sched/core"
	"github.com/linuxdeepin/warm-sched/events"
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

func (h *History) DoCapture(id string, c CaptureConfig, cfgmtime time.Time) (err error) {
	h.counts.AddCapture(id)
	h.counts.SetLifetime(id, c.Lifetime)

	if mt, err := h.mtime(id); err == nil {
		if cfgmtime.After(mt) {
			Log("The snapshot of %q's config is after current data, so recapure one.\n", id)
			if c.HasMincores() {
				Log("  %q use mincores directly, so delete it and delay real capture when next boot.\n", id)
				return os.Remove(h.path(id))
			}
			goto CONTINUE
		}
		if h.counts.IsDirty(id) {
			Log("The snapshot of %q's lifetime is end. so recapture one.\n", id)
			goto CONTINUE
		}

		Log("Ignore capture %q because already has one sample.\n", id)
		return nil
	}
CONTINUE:

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

func (h *History) has(id string) bool { return FileExist(h.path(id)) }
func (h *History) mtime(id string) (time.Time, error) {
	s, err := os.Stat(h.path(id))
	if err != nil {
		return time.Time{}, err
	}
	return s.ModTime(), nil
}

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

	sortApplyItems(h.applyQueue)

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
	snapPath := h.path(id)
	var snap core.Snapshot
	err := core.LoadFrom(snapPath, &snap)
	if err != nil {
		// 如果加载失败，表示 snap 文件已损坏，应该删除
		Log("remove the corrupted snapshot file %s", snapPath)
		rmErr := os.Remove(snapPath)
		if rmErr != nil {
			Log("WARN: remove the corrupted snapshot file failed: %v\n", err)
		}

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
