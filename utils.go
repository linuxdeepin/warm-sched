package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

const KB = 1024
const MB = 1024 * KB
const GB = 1024 * MB

func humanSize(s int) string {
	if s > GB {
		return fmt.Sprintf("%0.2fG", float32(s)/float32(GB))
	} else if s > MB {
		return fmt.Sprintf("%0.1fM", float32(s)/float32(MB))
	} else if s > KB {
		return fmt.Sprintf("%0.0fK", float32(s)/float32(KB))
	} else {
		return fmt.Sprintf("%dB", s)
	}
}

var ZeroFileInfo = FileCacheInfo{}
var PageSize = os.Getpagesize()
var PageSizeKB = os.Getpagesize() / KB

func SystemMemoryInfo() int64 {
	bs, err := ioutil.ReadFile("/proc/meminfo")
	if err != nil {
		return 0
	}
	const t = "MemAvailable:"
	for _, line := range strings.Split(string(bs), "\n") {
		if !strings.HasPrefix(line, t) {
			continue
		}
		fields := strings.Split(line, ":")
		if len(fields) != 2 {
			return 0
		}
		t, _ := strconv.ParseInt(strings.Trim(fields[1], " kB"), 10, 64)
		return t * int64(KB)
	}
	return 0
}
