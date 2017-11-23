package main

import (
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"path"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"
)

type MemRange struct {
	Offset int64
	Length int64
}

func toRange(vec []bool, pageSize int64) (MemRange, []bool) {
	var s int64
	var offset int64 = -1
	for i, v := range vec {
		if v && offset < 0 {
			offset = int64(i) * pageSize
		}
		if !v && offset > 0 {
			return MemRange{offset, s - offset}, vec[i+1:]
		}
		s += pageSize
	}
	return MemRange{offset, 0}, nil
}

func ToRanges(vec []bool, pageSize int64) []MemRange {
	panic("Not Implement.")
}

type FileCacheInfo struct {
	FName   string
	InCache []bool
	InN     int
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

func isNormalFile(info os.FileInfo) bool {
	mode := info.Mode()
	if mode&os.ModeType != 0 {
		return false
	}
	if info.Size() == 0 {
		return false
	}
	return true
}

func FileMincore(fname string) (FileCacheInfo, error) {
	fname, err := filepath.Abs(fname)
	if err != nil {
		return ZeroFileInfo, err
	}
	info, err := os.Lstat(fname)
	if err != nil {
		return ZeroFileInfo, err
	}
	if !isNormalFile(info) {
		return ZeroFileInfo, nil
	}
	size := info.Size()

	f, err := os.Open(fname)
	if err != nil {
		return ZeroFileInfo, err
	}
	defer f.Close()

	mmap, err := unix.Mmap(int(f.Fd()), 0, int(size), unix.PROT_NONE, unix.MAP_SHARED)
	if err != nil {
		return ZeroFileInfo, fmt.Errorf("could not mmap %s: %v", fname, err)
	}

	vecsz := (size + PageSize64 - 1) / PageSize64
	vec := make([]byte, vecsz)

	// get all of the arguments to the mincore syscall converted to uintptr
	mmap_ptr := uintptr(unsafe.Pointer(&mmap[0]))
	size_ptr := uintptr(size)
	vec_ptr := uintptr(unsafe.Pointer(&vec[0]))

	ret, _, err := unix.Syscall(unix.SYS_MINCORE, mmap_ptr, size_ptr, vec_ptr)
	if ret != 0 {
		return ZeroFileInfo, fmt.Errorf("syscall SYS_MINCORE failed: %v", err)
	}
	defer unix.Munmap(mmap)

	mc := make([]bool, vecsz)

	inCache := 0
	for i, b := range vec {
		if b%2 == 1 {
			inCache++
			mc[i] = true
		} else {
			mc[i] = false
		}
	}

	return FileCacheInfo{
		FName:   fname,
		InCache: mc,
		InN:     inCache,
	}, nil
}

func ShowFileCacheInfo(fname string, ch chan<- FileCacheInfo) error {
	info, err := FileMincore(fname)
	if err != nil {
		return err
	}
	ch <- info
	return nil
}

func ShowDirCacheInfos(root string, ch chan<- FileCacheInfo) error {
	f, err := os.Open(root)
	if err != nil {
		return err
	}
	defer f.Close()
	infos, err := f.Readdir(0)
	if err != nil {
		return err
	}
	for _, info := range infos {
		name := path.Join(root, info.Name())
		if info.IsDir() {
			ShowDirCacheInfos(name, ch)
		} else {
			ShowFileCacheInfo(name, ch)
		}
	}
	return nil
}

func Produce(ch chan<- FileCacheInfo, dirs []string) {
	defer close(ch)
	for _, f := range dirs {
		info, err := os.Lstat(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid Args %v\n", err)
			return
		}
		if info.IsDir() {
			err = ShowDirCacheInfos(f, ch)
		} else {
			err = ShowFileCacheInfo(f, ch)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "E:", err)
		}
	}
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

func LoadFiles(files []string) {
	wrapDoFiles(files, loadFiles)
}
func DropFiles(files []string) {
	wrapDoFiles(files, dropFiles)
}

func loadFiles(files []string) {
	for _, file := range files {
		FAdvise(file, AdviseLoad)
	}
}
func dropFiles(files []string) {
	for _, file := range files {
		FAdvise(file, AdviseDrop)
	}
}

func FAdvise(fname string, action int) error {
	var finfo syscall.Stat_t
	syscall.Stat(fname, &finfo)

	fd, err := syscall.Open(fname, syscall.O_RDONLY, 0755)
	if err != nil {
		return err
	}
	defer syscall.Close(fd)

	size := RoundPageSize(finfo.Size)

	err = unix.Fadvise(fd, 0, size, action)
	if err != nil {
		fmt.Println("E:", err)
	}
	return nil
}
