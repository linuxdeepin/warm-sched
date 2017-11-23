package main

import (
	"flag"
)

var debug = false

func main() {
	var load bool
	var drop bool
	flag.BoolVar(&load, "l", false, "preload files")
	flag.BoolVar(&drop, "d", false, "drop files")
	flag.BoolVar(&debug, "debug", false, "debug mode")
	flag.Parse()

	var files []string

	if flag.NArg() == 0 {
		files = []string{"."}
	} else {
		files = flag.Args()
	}

	switch {
	case load:
		LoadFiles(files)
	case drop:
		DropFiles(files)
	default:
		ShowRAMUsage(files)
	}
}
