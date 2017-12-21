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

func ShowPlymouthMessage(msg string) {
	cmd := exec.Command("plymouth", "display-message", msg)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("PLY:%v %v\n", err, cmd.Args)
	}
}

func AdjustToMaxAdviseRange(s int64, maxUnit int) []MemRange {
	bs := int(RoundPageSize(s))
	n := (bs + maxUnit - 1) / maxUnit
	var ret []MemRange
	for i := 0; i < int(n); i++ {
		ret = append(ret, MemRange{
			Offset: i * maxUnit,
			Count:  maxUnit,
		})
	}
	return ret
}

func ToRanges(vec []bool, maxUnit int) []MemRange {
	var ret []MemRange
	var i MemRange
	var pos = -1
	for {
		i, vec = toRange(vec, maxUnit)
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

func toRange(vec []bool, maxUnit int) (MemRange, []bool) {
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
		if (s-offset) >= maxUnit || (!v && found) {
			break
		}
		s++
	}
	return MemRange{offset, s - offset}, vec[s:]
}

func IsInPlymouthEnv() bool {
	if _, err := exec.LookPath("plymouth"); err != nil {
		return false
	}
	return exec.Command("pgrep", "plymouthd").Run() == nil
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
