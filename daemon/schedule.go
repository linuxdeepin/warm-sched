package main

import (
	"../events"
	"fmt"
	"time"
)

func (d *Daemon) CaptureEvents() ([]string, []string, error) {
	var befores, afters []string
	for _, cfg := range d.cfgs {
		for _, after := range cfg.Capture.After {
			if !events.IsSupport(after) {
				return nil, nil, fmt.Errorf("Doesn't support event %q in %q snapshot configure.",
					after, cfg.Id)
			}
			afters = append(afters, after)
		}
	}
	return befores, afters, nil
}

func (d *Daemon) scheduleApplys() error {
	for _, cfg := range d.cfgs {
		apply := cfg.Apply
		name := cfg.Id

		if !d.history.Has(name) {
			Log("Ignore apply %q because hasn't any samples.\n", name)
		}

		afters := apply.After
		befores := apply.Before

		if len(befores) != 0 && len(events.Check(befores)) > 0 {
			Log("Ignore apply %q because some of %v is already done.\n", name, befores)
			continue
		}

		if len(afters) == 0 {
			err := d.history.DoApply(name)
			if err != nil {
				return err
			}
		} else {
			err := events.Connect(afters, func() {
				err := d.history.DoApply(name)
				if err != nil {
					Log("Capture %q failed:%v", name, err)
				}
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *Daemon) scheduleCaptures() error {
	for _, cfg := range d.cfgs {
		capture := cfg.Capture
		name := cfg.Id

		if d.history.Has(name) && !capture.AlwaysLoad {
			Log("Ignore capture %q because already has one sample.\n", name)
			continue
		}

		afters := capture.After
		befores := capture.Before
		methods := capture.Method

		if len(befores) != 0 && len(events.Check(befores)) > 0 {
			Log("Ignore capture %q because some of %v is already done.\n", name, befores)
			continue
		}

		if len(afters) == 0 {
			err := d.history.DoCapture(name, methods)
			if err != nil {
				return err
			}
		} else {
			err := events.Connect(afters, func() {
				err := d.history.DoCapture(name, methods)
				if err != nil {
					Log("Capture %q failed:%v", name, err)
				}
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *Daemon) Schedule() error {
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

	go dumpWaitings()

	// 3. wait all events
	return events.Run()
}

func dumpWaitings() {
	for {
		time.Sleep(time.Second * 5)
		anything := false
		for _, s := range events.Scopes() {
			p := events.Pendings(s)
			if len(p) != 0 {
				Log("Waiting %v@%s\n", p, s)
				anything = true
			}
		}
		if !anything {
			return
		}
	}
}
