package main

import (
	"flag"
	"fmt"
	"os"
)

var debug bool

type AppFlags struct {
	load  bool // action of applying a snapshot
	store bool // action of taking a application's snapshot. TODO: use -base to implement this logical
	show  bool // action of show current page cache

	cacheDir string // the base cache directory.
}

func InitFlags() AppFlags {
	var af AppFlags

	flag.StringVar(&af.cacheDir, "cacheDir", "./warm-sched-cache", "base cache directory")

	flag.BoolVar(&debug, "debug", false, "debug mode")
	flag.BoolVar(&af.load, "load", false, "apply the snapshot")
	flag.BoolVar(&af.store, "store", false, "take application of the snapshot")

	flag.Parse()
	return af
}

func doActions(c RPCClient, af AppFlags, args []string) error {
	var err error
	switch {
	case af.store:
	case af.load:
		//		err = daemon.LoadAndApply("abc.snap", true)
	case af.show:
		snap, err := c.Capture()
		if err != nil {
			return err
		}
		DumpSnapshot(snap)
	default:
		cfgs, err := c.ListConfig()
		if err != nil {
			return err
		}
		for _, cfg := range cfgs {
			fmt.Println(cfg.Id, cfg.Description, cfg.TryFile)
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
