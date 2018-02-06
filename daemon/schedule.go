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
		// for _, before := range cfg.Capture.Before {
		// 	if !events.IsSupport(before) {
		// 		return nil, nil, fmt.Errorf("Doesn't support event %q in %q snapshot configure.", e, cfg.Id)
		// 	}

		// 	befores = append(befores, before)
		// }

	}
	return befores, afters, nil
}

func (d *Daemon) Schedule() error {
	// 1. ensure configures is valid
	// 2. schedule apply chains
	// 3. schedule capture chains

	before, after, err := d.CaptureEvents()
	if err != nil {
		return err
	}
	events.WaitAll(after...)
	events.WaitAll(before...)
	return nil
	//panic("Not Implement")
}
