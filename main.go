package main

import (
	"flag"
	"fmt"
	"os"
	"path"
)

var debug = false

const (
	SnapFull = "snap-full"
)

type AppFlags struct {
	takeS bool
	loadS bool
	showS bool

	loadFull bool

	increaseBootTimes bool

	showDB      bool
	showCurrent bool

	cacheDir string

	scanMountPoints stringSlice

	takeApp bool
}

func normalizeFlags(af AppFlags, args []string) ([]string, error) {
	var err error
	switch {
	case af.takeApp:
		if len(args) != 1 {
			return nil, fmt.Errorf("Please specify the applicaton's identify file path.")
		}
	case af.takeS, af.loadS, af.showS:
		if len(args) != 1 {
			return nil, fmt.Errorf("Please specify the snapshot name to handle.")
		}
	default:
		args = af.scanMountPoints
	}
	return args, err
}

func doActions(af AppFlags, args []string) error {
	m, err := NewManager(af.scanMountPoints, af.cacheDir)
	if err != nil {
		return err
	}
	if af.increaseBootTimes {
		err = m.IncreaseBootTimes()
	}

	switch {
	case af.showDB:
		err = m.ShowHistory()
	case af.loadFull:
		err = LoadFull(af.cacheDir)
	case af.takeApp:
		err = TryMkdir(af.cacheDir)
		if err != nil {
			return err
		}
		err = TakeApplicationSnapshot(af.cacheDir, af.scanMountPoints, args[0])
	case af.takeS:
		err = m.TakeSnapshot(args[0])
	case af.loadS:
		err = m.LoadSnapshot(args[0])
	case af.showS:
		err = m.ShowSnapshot(args[0])
	case af.showCurrent:
		err = DumpCurrentPageCache(af.scanMountPoints)
	default:
		err = m.ShowHistory()
	}
	return err
}

func LoadFull(baseDir string) error {
	err := LoadSnapshot(path.Join(baseDir, SnapFull), false)
	for _, app := range EnumerateAllApps(baseDir) {
		err = LoadSnapshot(app, true)
	}
	return err
}

func main() {
	var af AppFlags
	af.scanMountPoints = ListMountPoints()

	flag.StringVar(&af.cacheDir, "cacheDir", "./warm-sched-cache", "base cache directory")
	flag.Var(&af.scanMountPoints, "scanMountPoints", "The mount points to scan.")
	flag.BoolVar(&af.showDB, "info", false, "shwo the snapshot informations")
	flag.BoolVar(&af.showCurrent, "c", false, "shwo current page cache info")
	flag.BoolVar(&af.takeS, "take", false, "take a snapshot")
	flag.BoolVar(&af.loadS, "load", false, "load the snapshot")
	flag.BoolVar(&af.showS, "show", false, "show content of the snapshot")

	flag.BoolVar(&af.increaseBootTimes, "increase-boottimes", false, "DON'T do this.")

	flag.BoolVar(&af.loadFull, "load-full", false, "load full snapshot then load all applications snapshot")
	flag.BoolVar(&af.takeApp, "take-app", false, "take snapshot of the snapshot")

	flag.Var(&BlackDirectory, "black", "List of blacklist directory")
	flag.BoolVar(&debug, "debug", false, "debug mode")

	flag.Parse()
	args, err := normalizeFlags(af, flag.Args())
	if err == nil {
		err = doActions(af, args)
	}
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
