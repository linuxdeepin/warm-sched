package main

import (
	"fmt"
	"golang.org/x/sys/unix"
	"syscall"
	"time"
)

type FileCacheInfo struct {
	FName   string
	InCache []bool
	InN     int

	inode  uint64
	dev    uint64
	sector uint64
}

func Produce(ch chan<- FileCacheInfo, mps []string) {
	if SupportProduceByKernel() {
		go ProduceByKernel(ch, mps)
	} else {
		go ProduceBySyscall(ch, mps)
	}
}

func (info FileCacheInfo) String() string {
	return fmt.Sprintf("%s\t%d%%\t%s",
		humanSize(info.RAMSize()),
		info.Percentage(),
		info.FName,
	)
}

func (info FileCacheInfo) Percentage() int {
	n := len(info.InCache)
	if n == 0 {
		return 0
	}
	return info.InN * 100 / n
}
func (info FileCacheInfo) RAMSize() int {
	return info.InN * PageSize
}
func (info FileCacheInfo) FileSize() int {
	return len(info.InCache) * PageSize
}

const (
	AdviseLoad = unix.FADV_WILLNEED
	AdviseDrop = unix.FADV_DONTNEED
)

func wrapDoFiles(files []string, fn func(files []string)) {
	if debug {
		fmt.Println("------------BEFORE PRELOAD-----------")
		ShowRAMUsage(files)

		fn(files)

		time.Sleep(time.Millisecond * 400)
		fmt.Println("------------AFTER PRELOAD------------")
		ShowRAMUsage(files)
	} else {
		fn(files)
	}
}

func LoadFiles(files []string) error {
	wrapDoFiles(files, loadFiles)
	return nil
}
func DropFiles(files []string) error {
	wrapDoFiles(files, dropFiles)
	return nil
}

func loadFiles(files []string) {
	for _, file := range files {
		FAdvise(file, nil, AdviseLoad)
	}
}
func dropFiles(files []string) {
	for _, file := range files {
		FAdvise(file, nil, AdviseDrop)
	}
}

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
