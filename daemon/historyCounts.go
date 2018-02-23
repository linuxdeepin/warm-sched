package main

import (
	"../core"
	"sync"
)

type HistoryCounts struct {
	Apply    map[string]int
	Capture  map[string]int
	Lifetime map[string]int
	backpath string
	lock     sync.RWMutex
}

func NewHistoryCounts(p string) *HistoryCounts {
	init := &HistoryCounts{
		Apply:    make(map[string]int),
		Capture:  make(map[string]int),
		Lifetime: make(map[string]int),
		backpath: p,
	}
	core.LoadFrom(p, init)
	return init
}
func (hc *HistoryCounts) store() error {
	hc.lock.Lock()
	err := core.StoreTo(hc.backpath, hc)
	hc.lock.Unlock()
	return err
}

const (
	lifetimeC int = iota
	applyC
	captureC
)

func (hc *HistoryCounts) add(t int, id string, v int) {
	hc.lock.Lock()
	switch t {
	case lifetimeC:
		hc.Lifetime[id] = hc.Lifetime[id] + v
	case applyC:
		hc.Apply[id] = hc.Apply[id] + v
	case captureC:
		hc.Capture[id] = hc.Capture[id] + v
	default:
		panic("Histroy count doesn't support")
	}
	hc.lock.Unlock()

	err := hc.store()
	if err != nil {
		Log("Store failed when add %d : %v\n", t, err)
	}
}
func (hc *HistoryCounts) get(t int, id string) int {
	hc.lock.RLock()
	defer hc.lock.RUnlock()
	switch t {
	case lifetimeC:
		return hc.Lifetime[id]
	case applyC:
		return hc.Apply[id]
	case captureC:
		return hc.Capture[id]
	default:
		panic("Histroy count doesn't support")
	}
}
func (hc *HistoryCounts) AddCapture(id string)      { hc.add(captureC, id, 1) }
func (hc *HistoryCounts) GetCapture(id string) int  { return hc.get(captureC, id) }
func (hc *HistoryCounts) AddApply(id string)        { hc.add(applyC, id, 1) }
func (hc *HistoryCounts) GetApply(id string) int    { return hc.get(applyC, id) }
func (hc *HistoryCounts) GetLifetime(id string) int { return hc.get(lifetimeC, id) }

func (hc *HistoryCounts) SetLifetime(id string, v int) {
	if v == 0 {
		return
	}
	hc.lock.Lock()
	hc.Lifetime[id] = v
	hc.lock.Unlock()
	hc.store()
}
func (hc *HistoryCounts) IsDirty(id string) bool {
	lc := hc.get(lifetimeC, id)
	if lc == 0 {
		return false
	}
	ac := hc.GetApply(id)
	if ac == 0 {
		return false
	}
	return ac%lc == 0
}
