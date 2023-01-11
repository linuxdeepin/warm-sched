package core

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
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

const sysFstabPath = "/etc/fstab"

var _bindMapCache map[string]string
var _getMountBindMapOnce sync.Once

func getMountBindMap(fstabPath string) map[string]string {
	_getMountBindMapOnce.Do(func() {
		var err error
		_bindMapCache, err = getMountBindMapRaw(fstabPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "WARN: getMountBindMap failed:", err)
		}
	})
	return _bindMapCache
}

func getMountBindMapRaw(fstabPath string) (map[string]string, error) {
	fstabData, err := ioutil.ReadFile(fstabPath)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	// result 的 key 是源路径，value 是挂载点
	lines := strings.Split(string(fstabData), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line[0] == '#' {
			// ignore empty and comment line
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		src := fields[0]
		dst := fields[1]
		fsType := fields[2]
		if fsType == "none" && strings.HasPrefix(src, "/") &&
			strings.HasPrefix(dst, "/") {
			result[src] = dst
		}
	}

	return result, nil
}

const sysProcMountsPath = "/proc/mounts"

func ListMountPoints(mountsPath string) []string {
	mountData, err := ioutil.ReadFile(mountsPath)
	if err != nil {
		return []string{"/"}
	}

	supportedFsTypes := []string{"ext2", "ext3", "ext4",
		"fat", "ntfs", "btrfs", "xfs"}

	var result []string

	lines := strings.Split(string(mountData), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		mountPoint := fields[1]
		fsType := fields[2]
		if strSliceContains(supportedFsTypes, fsType) {
			// is supported fs
			result = append(result, mountPoint)
		}
	}

	// 按字符串长度（长->短）排序 result
	// 比如 /home/test, /home, /
	sort.Slice(result, func(i, j int) bool {
		// less func
		return len(result[i]) > len(result[j])
	})

	return result
}

func strSliceContains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
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

func LoadJSONFrom(fname string, o interface{}) error {
	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(o)
	if err != nil {
		return fmt.Errorf("LoadFrom(%q, %T) -> %q", fname, o, err.Error())
	}
	return nil
}
func LoadFrom(fname string, o interface{}) error {
	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer f.Close()
	err = gob.NewDecoder(f).Decode(o)
	if err != nil {
		return fmt.Errorf("LoadFrom(%q, %T) -> %q", fname, o, err.Error())
	}
	return nil
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
	defer w.Close()
	return gob.NewEncoder(w).Encode(o)
}

func ReadFileInclude(f string) []string {
	bs, err := ioutil.ReadFile(f)
	if err != nil {
		return nil
	}
	var ret []string
	for _, l := range strings.Split(string(bs), "\n") {
		l = strings.TrimSpace(l)
		if l != "" {
			fileInfo, err := os.Stat(l)
			if err == nil && !fileInfo.IsDir() {
				ret = append(ret, l)
			}
		}
	}
	return ret
}
