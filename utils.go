package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

const KB = 1024
const MB = 1024 * KB
const GB = 1024 * MB

func FileExists(f string) bool {
	info, err := os.Stat(f)
	if err != nil || info.IsDir() {
		return false
	}
	return true
}

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
	return int64(RoundPageCount(s)) * PageSize64
}

func RoundPageCount(s int64) int {
	return int((s + PageSize64 - 1) / PageSize64)
}

var PageSize64 = int64(PageSize)
var PageSizeKB = PageSize / KB

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

func splitPageRange(r PageRange, mc int) []PageRange {
	var ret []PageRange
	if r.Count <= mc {
		ret = append(ret, r)
	} else {
		i := PageRange{
			Offset: r.Offset,
			Count:  mc,
		}
		j := PageRange{
			Offset: r.Offset + mc,
			Count:  r.Count - mc,
		}
		ret = append(ret, i)
		ret = append(ret, splitPageRange(j, mc)...)
	}
	return ret
}

func PageRangeToSizeRange(pageSize int, maxPageCount int, rs ...PageRange) [][2]int {
	var ret [][2]int
	for _, r := range rs {
		for _, rr := range splitPageRange(r, maxPageCount) {
			ret = append(ret, [2]int{rr.Offset * pageSize, rr.Count * pageSize})
		}
	}
	return ret
}

func ToRanges(vec []bool) []PageRange {
	var ret []PageRange
	var i PageRange
	var pos = -1
	for {
		i, vec = toRange(vec)
		if i.Offset == -1 {
			break
		}
		if pos != -1 {
			i.Offset += pos
		}
		pos = i.Offset + i.Count

		ret = append(ret, i)
		if len(vec) == 0 {
			break
		}
	}
	return ret
}

func toRange(vec []bool) (PageRange, []bool) {
	var s int
	var offset = -1
	var found = false
	for i, v := range vec {
		if !found && v {
			offset = i
		}
		if v {
			found = true
		}
		if !v && found {
			break
		}
		s++
	}
	return PageRange{offset, s - offset}, vec[s:]
}

func ListMountPoints() []string {
	cmd := exec.Command("/bin/df",
		"-t", "ext2",
		"-t", "ext3",
		"-t", "ext4",
		"-t", "fat",
		"-t", "ntfs",
		"--output=target")
	cmd.Env = []string{"LC_ALL=C"}
	buf := bytes.NewBuffer(nil)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return nil
	}

	line, err := buf.ReadString('\n')
	if line != "Mounted on\n" || err != nil {
		return nil
	}
	var ret []string
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			break
		}
		ret = append(ret, strings.TrimSpace(line))
	}
	sort.Sort(sort.Reverse(sort.StringSlice(ret)))
	return ret
}
