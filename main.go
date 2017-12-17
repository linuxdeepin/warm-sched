package main

import (
	"flag"
	"fmt"
	"os"
)

var debug = false

func main() {
	var takeS, loadS, showS bool
	var wait, ply bool
	var out string

	flag.BoolVar(&takeS, "take", false, "take a snapshot")
	flag.BoolVar(&loadS, "load", false, "load the snapshot")
	flag.BoolVar(&showS, "show", false, "show content of the snapshot")

	flag.Var(&BlackDirectory, "black", "List of blacklist directory")
	flag.StringVar(&out, "out", "/dev/shm/hh", "the file name for snapshot")
	flag.BoolVar(&debug, "debug", false, "debug mode")
	flag.BoolVar(&wait, "wait", true, "wait load completed")
	flag.BoolVar(&ply, "plymouth", false, "report progress to plymouth")

	flag.Parse()

	var files []string

	if flag.NArg() == 0 {
		files = ListMountPoints()
	} else {
		files = CalcRealTargets(flag.Args(), ListMountPoints())
	}

	var err error
	switch {
	case takeS:
		err = TakeSnapshot(files, out)
	case loadS:
		err = LoadSnapshot(out, wait, ply)
	case showS:
		err = ShowSnapshot(out)
	default:
		err = DumpCurrentPageCache(files)

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
