package main

import (
	"fmt"
	"syscall"
)

const (
	AdviseLoad = syscall.MADV_WILLNEED
	AdviseDrop = syscall.MADV_DONTNEED
)

func Readahead(fname string, rs []PageRange) error {
	fd, err := syscall.Open(fname, syscall.O_RDONLY, 0755)
	if err != nil {
		return err
	}
	defer syscall.Close(fd)

	if len(rs) == 0 {
		var finfo syscall.Stat_t
		syscall.Stat(fname, &finfo)
		rs = []PageRange{PageRange{0, RoundPageCount(finfo.Size)}}
	}
	for _, r := range PageRangeToSizeRange(PageSize, 32, rs...) {
		_, _, e := syscall.Syscall(syscall.SYS_READAHEAD,
			uintptr(fd),
			uintptr(r[0]),
			uintptr(r[1]),
		)
		if e != 0 {
			fmt.Println("E:", e)
		}
	}
	return nil
}

func FAdvise(fname string, rs []PageRange, action int) error {
	fd, err := syscall.Open(fname, syscall.O_RDONLY, 0755)
	if err != nil {
		return err
	}
	defer syscall.Close(fd)

	if len(rs) == 0 {
		var finfo syscall.Stat_t
		syscall.Stat(fname, &finfo)
		rs = append(rs, PageRange{0, RoundPageCount(finfo.Size)})
	}

	var maxUnit int
	switch action {
	case AdviseLoad:
		maxUnit = 32
	case AdviseDrop:
		maxUnit = 1000000
	default:
		panic("Unknown Action")
	}

	for _, r := range PageRangeToSizeRange(PageSize, maxUnit, rs...) {
		_, _, errno := syscall.Syscall6(syscall.SYS_FADVISE64,
			uintptr(fd),
			uintptr(r[0]),
			uintptr(r[1]),
			uintptr(action),
			0, 0)
		if errno != 0 {
			return err
		}
	}
	return nil
}
