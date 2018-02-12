package main

import (
	"../core"
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
	doCapture := func(name string, ws int, force bool, methods []*core.CaptureMethod) error {
		Log("Begin DoCapture %q\n", name)
		err := d.history.DoCapture(name, ws, force, methods)
		if err != nil {
			Log("End DoCapture %q failed: %v\n", name, err)
			return err
		} else {
			Log("End DoCapture %q\n", name)
			return nil
		}
	}

	for _, cfg := range d.cfgs {
		capture := cfg.Capture
		name := cfg.Id

		afters := capture.After
		befores := capture.Before
		methods := capture.Method

		if len(befores) != 0 && len(events.Check(befores)) > 0 {
			Log("Ignore capture %q because some of %v is already done.\n", name, befores)
			continue
		}

		var err error
		if len(afters) == 0 {
			err = doCapture(name, capture.WaitSeond, capture.AlwaysLoad, methods)
		} else {
			err = events.Connect(afters, func() { doCapture(name, capture.WaitSeond, capture.AlwaysLoad, methods) })
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

func EventWaits() map[string][]string {
	var ret = make(map[string][]string)
	for _, s := range events.Scopes() {
		p := events.Pendings(s)
		if len(p) != 0 {
			ret[s] = p
		}
	}
	return ret
}
