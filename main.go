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

	wait bool
	ply  bool

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
	var err error
	switch {
	case af.loadFull:
		err = LoadFull(af.cacheDir)
	case af.takeApp:
		err = TakeApplicationSnapshot(af.cacheDir, af.scanMountPoints, args[0])
	case af.takeS:
		err = TakeSnapshot(af.scanMountPoints, args[0])
	case af.loadS:
		err = LoadSnapshot(args[0], af.wait, af.ply)
	case af.showS:
		err = ShowSnapshot(args[0])
	default:
		err = DumpCurrentPageCache(af.scanMountPoints)
	}
	return err
}

func LoadFull(baseDir string) error {
	err := LoadSnapshot(path.Join(baseDir, SnapFull), true, false)
	for _, app := range EnumerateAllApps(baseDir) {
		err = LoadSnapshot(app, true, false)
	}
	return err
}

func main() {
	var af AppFlags
	af.scanMountPoints = ListMountPoints()

	flag.StringVar(&af.cacheDir, "cacheDir", "/var/cache/warm-sched", "base cache directory")
	flag.Var(&af.scanMountPoints, "scanMountPoints", "The mount points to scan.")

	flag.BoolVar(&af.takeS, "take", false, "take a snapshot")
	flag.BoolVar(&af.loadS, "load", false, "load the snapshot")
	flag.BoolVar(&af.showS, "show", false, "show content of the snapshot")

	flag.BoolVar(&af.loadFull, "load-full", false, "load full snapshot then load all applications snapshot")
	flag.BoolVar(&af.takeApp, "take-app", false, "take snapshot of the snapshot")

	flag.BoolVar(&af.wait, "wait", true, "wait load completed")
	flag.BoolVar(&af.ply, "plymouth", false, "report progress to plymouth")

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
