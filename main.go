package main

import (
	"flag"
	"fmt"
	"os"
)

var debug = false

type AppFlags struct {
	apply      bool // action of applying a snapshot
	applyAll   bool // action of applying DE snapshot and properly application's snapshots
	applyBasic bool // action of applying boot snapshot

	showSnapshot bool // action of showing a snapshot
	showInfo     bool // action of dump history database
	showCurrent  bool // action of taking a snapshot with current Page Cache

	takeDesktop bool
	takeBasic   bool
	takeApp     bool // action of taking a application's snapshot. TODO: use -base to implement this logical

	cacheDir        string      // the base cache directory.
	scanMountPoints stringSlice // the which mount points will be used when taking a snapshot
}

func InitFlags() AppFlags {
	var af AppFlags
	af.scanMountPoints = ListMountPoints()

	flag.StringVar(&af.cacheDir, "cacheDir", "./warm-sched-cache", "base cache directory")
	//	flag.Var(&af.scanMountPoints, "scanMountPoints", "The mount points to scan.")
	//	flag.Var(&BlackDirectory, "black", "List of blacklist directory")
	flag.BoolVar(&debug, "debug", false, "debug mode")

	flag.BoolVar(&af.showInfo, "info", false, "dump history database")
	flag.BoolVar(&af.showCurrent, "c", false, "shwo current page cache info")
	flag.BoolVar(&af.showSnapshot, "show", false, "show content of the snapshot")

	flag.BoolVar(&af.apply, "apply", false, "apply the snapshot")
	flag.BoolVar(&af.applyBasic, "apply-basic", false, "apply boot snapshot and increasing boo-times")
	flag.BoolVar(&af.applyAll, "apply-all", false, "apply full snapshot then load all applications snapshot")

	flag.BoolVar(&af.takeBasic, "take-basic", false, "taking boot snapshot and increasing boo-times")
	flag.BoolVar(&af.takeDesktop, "take-desktop", false, "take the desktop environment snapshot")
	flag.BoolVar(&af.takeApp, "take", false, "take application of the snapshot")

	flag.Parse()
	return af
}

func doActions(af AppFlags, args []string) error {
	m, err := NewManager(af.scanMountPoints, af.cacheDir)
	if err != nil {
		return err
	}
	switch {
	case af.takeBasic:
		err = m.TakeBasic()
	case af.takeDesktop:
		err = m.TakeDesktop()
	case af.takeApp:
		if len(args) < 1 {
			flag.Usage()
			os.Exit(1)
		}
		err = m.TakeApplication(args[0])
	case af.applyAll:
		err = m.ApplyAll()
	case af.applyBasic:
		err = m.ApplyBasic()
	case af.showSnapshot:
		if len(args) < 1 {
			flag.Usage()
			os.Exit(1)
		}
		err = m.ShowSnapshot(args[0])
	case af.showCurrent:
		err = DumpCurrentPageCache(af.scanMountPoints)
	default:
		err = m.ShowHistory()
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

type stringSlice []string

func (s *stringSlice) String() string {
	return fmt.Sprintf("%s", *s)
}

func (i *stringSlice) Set(v string) error {
	*i = append(*i, v)
	return nil
}
