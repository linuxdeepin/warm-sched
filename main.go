package main

import (
	"flag"
	"fmt"
	"os"
)

var debug = false

func main() {
	var load bool
	var drop bool
	var takeS, loadS, wait, ply bool
	var out string

	flag.BoolVar(&load, "l", false, "preload files")
	flag.BoolVar(&drop, "d", false, "drop files")
	flag.BoolVar(&debug, "debug", false, "debug mode")
	flag.BoolVar(&takeS, "take", false, "take a snapshot")
	flag.BoolVar(&loadS, "load", false, "load the snapshot")
	flag.StringVar(&out, "out", "/dev/shm/hh", "the file name for snapshot")
	flag.BoolVar(&wait, "wait", true, "wait load completed")
	flag.BoolVar(&ply, "plymouth", false, "report progress to plymouth")

	flag.Parse()

	var files []string

	if flag.NArg() == 0 {
		files = []string{os.Getenv("PWD")}
	} else {
		files = flag.Args()
	}

	var err error
	switch {
	case takeS:
		err = TakeSnapshot(files, out)
	case loadS:
		err = LoadSnapshot(out, wait, ply)
	case load:
		err = LoadFiles(files)
	case drop:
		err = DropFiles(files)
	default:
		err = ShowRAMUsage(files)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "E:", err)
	}
}
