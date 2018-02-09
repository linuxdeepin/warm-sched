package core

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

func FileMincore(fname string) (FileInfo, error) {
	info, err := os.Lstat(fname)
	if err != nil {
		return FileInfo{}, err
	}

	size := info.Size()
	if !info.Mode().IsRegular() || size == 0 {
		return FileInfo{}, fmt.Errorf("%q is not a normal file", fname)
	}

	f, err := os.Open(fname)
	if err != nil {
		return FileInfo{}, err
	}
	defer f.Close()

	mappings, err := fileMincore(int(f.Fd()), size)
	if err != nil {
		return FileInfo{}, err
	}

	var sector = uint64(0)
	if len(mappings) > 0 && RunByRoot {
		sector, err = getSectorNumber(f.Fd())
		if err != nil {
			return FileInfo{}, err
		}
	}
	sinfo := info.Sys().(*syscall.Stat_t)

	return FileInfo{
		Name:     fname,
		FileSize: uint64(size),
		Mapping:  mappings,

		dev:    sinfo.Dev,
		sector: sector,
	}, nil
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

func toRanges(vec []bool) []PageRange {
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

func fileMincore(fd int, size int64) ([]PageRange, error) {
	mmap, _, err := syscall.Syscall6(syscall.SYS_MMAP,
		uintptr(0),
		uintptr(size),
		syscall.PROT_NONE,
		syscall.MAP_SHARED,
		uintptr(fd),
		0)
	if err != 0 {
		return nil, err
	}

	defer syscall.Syscall(syscall.SYS_MUNMAP, mmap, uintptr(size), 0)

	vec := make([]bool, roundPageCount(size))
	_, _, err = syscall.Syscall(syscall.SYS_MINCORE,
		mmap,
		uintptr(size),
		uintptr(unsafe.Pointer(&vec[0])),
	)
	if err != 0 {
		return nil, err
	}
	return toRanges(vec), nil
}

var RunByRoot = os.Geteuid() == 0

func getSectorNumber(fd uintptr) (uint64, error) {
	b := 0
	const FIBMAP = 1
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), FIBMAP, uintptr(unsafe.Pointer(&b)))
	if err != 0 {
		return 0, err
	}
	return uint64(b), nil
}
