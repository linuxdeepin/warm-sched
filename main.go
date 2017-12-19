package main

import (
	"flag"
	"fmt"
	"os"
)

var debug = false

type AppFlags struct {
	takeS bool
	loadS bool
	showS bool

	wait bool
	ply  bool
	out  string

	takeApp bool
}

func normalizeFlags(af AppFlags, args []string) ([]string, error) {
	var err error
	switch {
	case af.takeApp:
		if len(args) == 0 {
			return nil, fmt.Errorf("Please specify the applicaton's path")
		}
	case af.takeS, af.loadS, af.showS:
		if flag.NArg() == 0 {
			args = ListMountPoints()
		} else {
			args = CalcRealTargets(args, ListMountPoints())
		}
	default:
		args = ListMountPoints()
	}
	return args, err
}

func doActions(af AppFlags, files []string) error {
	var err error
	switch {
	case af.takeApp:
		err = TakeApplicationSnapshot(files)
	case af.takeS:
		err = TakeSnapshot(files, af.out)
	case af.loadS:
		err = LoadSnapshot(af.out, af.wait, af.ply)
	case af.showS:
		err = ShowSnapshot(af.out)
	default:
		err = DumpCurrentPageCache(files)
	}
	return err
}

func main() {
	var af AppFlags

	flag.BoolVar(&af.takeS, "take", false, "take a snapshot")
	flag.BoolVar(&af.loadS, "load", false, "load the snapshot")
	flag.BoolVar(&af.showS, "show", false, "show content of the snapshot")

	flag.BoolVar(&af.takeApp, "take-app", false, "take snapshot of the snapshot")

	flag.StringVar(&af.out, "out", "/dev/shm/hh", "the file name for snapshot")
	flag.BoolVar(&af.wait, "wait", true, "wait load completed")
	flag.BoolVar(&af.ply, "plymouth", false, "report progress to plymouth")

	flag.Var(&BlackDirectory, "black", "List of blacklist directory")
	flag.BoolVar(&debug, "debug", false, "debug mode")

	flag.Parse()
	files, err := normalizeFlags(af, flag.Args())
	if err == nil {
		err = doActions(af, files)
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
