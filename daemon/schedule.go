package main

import (
	"../events"
	"fmt"
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

func (d *Daemon) Schedule() error {
	// 1. ensure configures is valid
	// 2. schedule apply chains
	// 3. schedule capture chains

	// _, after, err := d.CaptureEvents()
	// if err != nil {
	// 	return err
	// }

	for _, cfg := range d.cfgs {
		capture := cfg.Capture
		name := cfg.Id
		afters := capture.After
		befores := capture.Before
		methods := capture.Method

		if len(befores) != 0 && len(events.Check(befores)) > 0 {
			Log("Ignore snapshot %q because some of %v is already done.\n", name, befores)
			continue
		}

		if len(afters) == 0 {
			err := d.DoCapture(name, methods)
			if err != nil {
				return err
			}
		} else {
			err := events.Connect(afters, func() {
				err := d.DoCapture(name, methods)
				if err != nil {
					Log("Capture %q failed:%v", name, err)
				}
			})
			if err != nil {
				return err
			}
		}
	}
	return events.Run()
}
