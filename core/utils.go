package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
)

var pageSize = os.Getpagesize()
var pageSize64 = int64(pageSize)

func HumanSize(s int) string {
	if s > _GB {
		return fmt.Sprintf("%0.2fG", float32(s)/float32(_GB))
	} else if s > _MB {
		return fmt.Sprintf("%0.1fM", float32(s)/float32(_MB))
	} else if s > _KB {
		return fmt.Sprintf("%0.0fK", float32(s)/float32(_KB))
	} else {
		return fmt.Sprintf("%dB", s)
	}
}

const _KB = 1024
const _MB = 1024 * _KB
const _GB = 1024 * _MB

func roundPageCount(s int64) int {
	return int((s + pageSize64 - 1) / pageSize64)
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

func pageRangeToSizeRange(pageSize int, maxPageCount int, rs ...PageRange) [][2]int {
	var ret [][2]int
	for _, r := range rs {
		for _, rr := range splitPageRange(r, maxPageCount) {
			ret = append(ret, [2]int{rr.Offset * pageSize, rr.Count * pageSize})
		}
	}
	return ret
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

func _ReduceFilePath(expandFn func(string) string, fs ...string) []string {
	cache := make(map[string]bool)
	for _, f := range fs {
		ff := os.Expand(f, expandFn)
		cache[ff] = true
	}
	var ret []string
	for k, v := range cache {
		if v {
			ret = append(ret, k)
		}
	}
	return ret
}

func EnsureDir(d string) error {
	info, err := os.Stat(d)
	if err == nil && !info.IsDir() {
		return fmt.Errorf("%q is not a directory", d)
	}
	return os.MkdirAll(d, 0755)
}

func StoreTo(fname string, o interface{}) error {
	err := EnsureDir(path.Dir(fname))
	if err != nil {
		return err
	}
	w, err := os.Create(fname)
	if err != nil {
		return err
	}
	return storeTo(w, o)
}

func LoadFrom(fname string, o interface{}) error {
	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	err = loadFrom(f, o)
	if err != nil {
		return fmt.Errorf("LoadFrom(%q, %T) -> %q", fname, o, err.Error())
	}
	return nil
}

func storeTo(w io.Writer, o interface{}) error {
	return json.NewEncoder(w).Encode(o)
}
func loadFrom(r io.Reader, o interface{}) error {
	return json.NewDecoder(r).Decode(o)
}
