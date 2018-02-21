package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func FileExist(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func Log(fmtStr string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, fmtStr, args...)
}

func MemAvailable() uint64 {
	bs, err := ioutil.ReadFile("/proc/meminfo")
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(string(bs), "\n") {
		if !strings.HasPrefix(line, "MemAvailable:") {
			continue
		}
		var d uint64
		fmt.Sscanf(line, "MemAvailable: %d kB\n", &d)
		return d * 1024
	}
	return 0
}
