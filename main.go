package main

import (
	"./module/show"
	"./module/store"
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

func doActions(af AppFlags, args []string) error {
	var err error
	switch {
	case af.store:
		err = store.CaptureAndStore([]string{"/"}, "abc", "abc.snap")
	case af.load:
		err = store.LoadAndApply("abc.snap", true)
	case af.show:
		err = show.DumpPageCache()
	default:
		err = show.DumpPageCache()
	}
	return err
}

func main() {
	af := InitFlags()
	err := doActions(af, flag.Args())
	if err != nil {
		fmt.Fprintln(os.Stderr, "E:", err)
	}
}
