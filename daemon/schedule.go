package main

import (
	"context"

	"github.com/linuxdeepin/warm-sched/events"
)

func (d *Daemon) scheduleApplys() error {
	for _, cfg := range d.cfgs {
		apply := cfg.Apply
		name := cfg.Id
		initUsage := apply.InitUsage

		afters := apply.After
		befores := apply.Before

		if len(befores) != 0 && len(events.Check(befores)) > 0 {
			Log("Ignore apply %q because some of %v is already done.\n", name, befores)
			continue
		}

		var err error
		if len(afters) == 0 {
			err = d.history.RequestApply(name, initUsage)
		} else {
			err = events.Connect(afters, func() bool {
				d.history.RequestApply(name, initUsage)
				return false
			})
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Daemon) scheduleCaptures() error {
	for _, cfg := range d.cfgs {
		capture := cfg.Capture
		id := cfg.Id
		afters := capture.After
		befores := capture.Before
		mtime := cfg.mtime

		if len(befores) != 0 && len(events.Check(befores)) > 0 {
			Log("Ignore capture %q because some of %v is already done.\n", id, befores)
			continue
		}

		var err error
		if len(afters) == 0 {
			err = d.history.DoCapture(id, capture, mtime)
		} else {
			err = events.Connect(afters, func() bool {
				capErr := d.history.DoCapture(id, capture, mtime)
				if capErr != nil {
					return d.retryCapture(id, &capture)
				}
				return false
			})
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Daemon) recordCaptureRetry(id string) (can bool) {
	const retryLimit = 30

	d.retryCaptureMu.Lock()
	defer d.retryCaptureMu.Unlock()

	d.retryCaptureCount[id]++
	count := d.retryCaptureCount[id]
	return count < retryLimit
}

func (d *Daemon) retryCapture(id string, captureCfg *CaptureConfig) (noDel bool) {
	// 仅能重试 After 含有 x11 scope 事件的捕获
	hasX11 := false
	for _, event := range captureCfg.After {
		scope, _, ok := events.SplitEvent(event)
		if ok && scope == "x11" {
			hasX11 = true
		}
	}
	if !hasX11 {
		return false
	}

	if !d.recordCaptureRetry(id) {
		Log("retries to capture %v have reached the limit\n", id)
		return false
	}

	// 可以重试
	Log("retry capture %v\n", id)
	return true
}

func (d *Daemon) Schedule(ctx context.Context) error {
	if d.scheduling {
		return nil
	}
	d.scheduling = true
	defer func() { d.scheduling = false }()
	// 1. schedule apply chains
	err := d.scheduleCaptures()
	if err != nil {
		Log("Schedule Capture failed: %v", err)
	}

	// 2. schedule capture chains
	err = d.scheduleApplys()
	if err != nil {
		Log("Schedule Capture failed: %v", err)
	}

	// 3. wait all events
	return events.Run(ctx)
}
