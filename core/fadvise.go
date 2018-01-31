package core

import (
	"fmt"
	"syscall"
)

const (
	_AdviseLoad = syscall.MADV_WILLNEED
	_AdviseDrop = syscall.MADV_DONTNEED
)

func Apply(info FileInfo) error {
	err := _FAdvise(info.Name, info.Mapping, _AdviseLoad)
	if err != nil {
		return fmt.Errorf("Apply failed for %s : %v", info.Name, err)
	}
	return nil
}

func _Readahead(fname string, pageCluster int, rs []PageRange) error {
	fd, err := syscall.Open(fname, syscall.O_RDONLY, 0755)
	if err != nil {
		return err
	}
	defer syscall.Close(fd)

	if len(rs) == 0 {
		var finfo syscall.Stat_t
		syscall.Stat(fname, &finfo)
		rs = []PageRange{PageRange{0, roundPageCount(finfo.Size)}}
	}
	for _, r := range pageRangeToSizeRange(pageSize, pageCluster, rs...) {
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

func _FAdvise(fname string, rs []PageRange, action int) error {
	fd, err := syscall.Open(fname, syscall.O_RDONLY, 0755)
	if err != nil {
		return err
	}
	defer syscall.Close(fd)

	if len(rs) == 0 {
		var finfo syscall.Stat_t
		syscall.Stat(fname, &finfo)
		rs = append(rs, PageRange{0, roundPageCount(finfo.Size)})
	}

	var maxUnit int
	switch action {
	case _AdviseLoad:
		maxUnit = 32
	case _AdviseDrop:
		maxUnit = 1000000
	default:
		panic("Unknown Action")
	}

	for _, r := range pageRangeToSizeRange(pageSize, maxUnit, rs...) {
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
