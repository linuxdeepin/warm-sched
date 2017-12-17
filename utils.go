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

func RoundPageSize(s int64) int64 {
	n := (s + PageSize64 - 1) / PageSize64
	return n * PageSize64
}

var ZeroFileInfo = FileCacheInfo{}
var PageSize = os.Getpagesize()
var PageSize64 = int64(PageSize)
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

func FullRanges(fileSize int64) []MemRange {
	n := (fileSize + MaxAdviseSize - 1) / MaxAdviseSize
	var ret []MemRange
	for i := 0; i < int(n); i++ {
		ret = append(ret, MemRange{
			Offset: int64(i) * MaxAdviseSize,
			Length: MaxAdviseSize,
		})
	}
	return ret
}

func ToRanges(vec []bool, pageSize int64) []MemRange {
	var ret []MemRange
	var i MemRange
	var pos int64 = -1
	for {
		i, vec = toRange(vec, pageSize)
		if i.Offset == -1 {
			break
		}
		if pos != -1 {
			i.Offset += pos
		}
		pos = i.Offset + i.Length

		ret = append(ret, i)
		if len(vec) == 0 {
			break
		}
	}
	return ret
}

func toRange(vec []bool, pageSize int64) (MemRange, []bool) {
	var s int64
	var offset int64 = -1
	for i, v := range vec {
		if v && offset < 0 {
			offset = int64(i) * pageSize
		}
		if !v && offset > 0 {
			return MemRange{offset, s - offset}, vec[i:]
		}

		length := s - offset
		s += pageSize
		if length >= MaxAdviseSize {
			return MemRange{offset, length}, vec[i:]
		}

	}
	return MemRange{offset, s - offset}, nil
}
