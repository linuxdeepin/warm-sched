package main

import (
	"../events"
	"context"
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
			err = events.Connect(afters, func() { d.history.RequestApply(name, initUsage) })
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
			err = events.Connect(afters, func() {
				d.history.DoCapture(id, capture, mtime)
			})
		}
		if err != nil {
			return err
		}
	}
	return nil
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
