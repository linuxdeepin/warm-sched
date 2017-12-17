package main

import (
	"fmt"
	"golang.org/x/sys/unix"
	"syscall"
)

const (
	AdviseLoad = unix.FADV_WILLNEED
	AdviseDrop = unix.FADV_DONTNEED
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
		rs = FullRanges(finfo.Size)
	}
	for _, r := range rs {
		_, _, e := syscall.Syscall(syscall.SYS_READAHEAD,
			uintptr(fd),
			uintptr(r.Offset),
			uintptr(r.Length))
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
		if action == AdviseLoad {
			rs = FullRanges(finfo.Size)
		} else {
			rs = append(rs, MemRange{0, RoundPageSize(finfo.Size)})
		}
	}
	return fadvise(fd, rs, action)
}

func fadvise(fd int, rs []MemRange, action int) error {
	for _, r := range rs {
		err := unix.Fadvise(fd, r.Offset, r.Length, action)
		if err != nil {
			return err
		}
	}
	return nil
}
