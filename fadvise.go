package main

import (
	"fmt"
	"golang.org/x/sys/unix"
	"syscall"
)

const (
	AdviseLoad = unix.FADV_WILLNEED
	AdviseDrop = unix.FADV_DONTNEED

	MaxAdvisePageCount = 32
)

func Readahead(fname string, rs []MemRange) error {
	fd, err := syscall.Open(fname, syscall.O_RDONLY, 0755)
	if err != nil {
		return err
	}
	defer syscall.Close(fd)

	if len(rs) == 0 {
		var finfo syscall.Stat_t
		syscall.Stat(fname, &finfo)
		rs = AdjustToMaxAdviseRange(finfo.Size, MaxAdvisePageCount)
	}
	for _, r := range rs {
		_, _, e := syscall.Syscall(syscall.SYS_READAHEAD,
			uintptr(fd),
			uintptr(r.Offset*PageSize),
			uintptr(r.Count*PageSize))
		if e != 0 {
			fmt.Println("E:", e)
		}
	}
	return nil
}

func FAdvise(fname string, rs []MemRange, action int) error {
	fd, err := syscall.Open(fname, syscall.O_RDONLY, 0755)
	if err != nil {
		return err
	}
	defer syscall.Close(fd)

	if len(rs) == 0 {
		var finfo syscall.Stat_t
		syscall.Stat(fname, &finfo)
		switch action {
		case AdviseLoad:
			// fadvise(2) decline the request if size is
			// greater than MaxAdviseSize.
			rs = AdjustToMaxAdviseRange(finfo.Size, MaxAdvisePageCount)
		case AdviseDrop:
			rs = append(rs, MemRange{0, RoundPageCount(finfo.Size)})
		default:
			panic("Unknown Action")
		}
	}
	return fadvise(fd, rs, action)
}

func fadvise(fd int, rs []MemRange, action int) error {
	for _, r := range rs {
		err := unix.Fadvise(fd, int64(r.Offset*PageSize), int64(r.Count*PageSize), action)
		if err != nil {
			return err
		}
	}
	return nil
}
