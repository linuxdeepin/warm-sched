package main

import (
	"flag"
	"fmt"
	"os"
)

type AppFlags struct {
	capture  bool
	apply    bool
	schedule bool
}

func InitFlags() AppFlags {
	var af AppFlags

	flag.BoolVar(&af.capture, "c", false, "capture a snapshot")
	flag.BoolVar(&af.apply, "a", false, "apply the snapshot")
	flag.BoolVar(&af.schedule, "s", false, "schedule handle snapshot by configures")

	flag.Parse()
	return af
}

func doActions(c RPCClient, af AppFlags, args []string) error {
	cfgs, err := c.ListConfig()
	if err != nil {
		return err
	}

	switch {
	case af.capture:
		if len(args) == 0 {
			return fmt.Errorf("Please specify configure name")
		}
		snap, err := c.Capture(args[0])
		if err != nil {
			return err
		}
		DumpSnapshot(snap)

	case af.apply:
		panic("not impement apply operatin")
	case af.schedule:
		err = c.Schedule()
	default:
		for _, cfg := range cfgs {
			fmt.Println(cfg)
		}
	}
	return err
}

func main() {
	c, err := NewRPCClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cant Init RPC:", err)
		return
	}

	af := InitFlags()
	err = doActions(c, af, flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, "E:", err)
	}
}
