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

type FileCacheInfo struct {
	FName   string
	InCache []bool
	InN     int

	inode  uint64
	dev    uint64
	sector uint64
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

func isNormalFile(info syscall.Stat_t) bool {
	mode := info.Mode
	if os.FileMode(mode)&os.ModeType != 0 {
		return false
	}
	if info.Size == 0 {
		return false
	}
	return true
}

func FileMincore(fname string) (FileCacheInfo, error) {
	fname, err := filepath.Abs(fname)
	if err != nil {
		return ZeroFileInfo, err
	}
	var info syscall.Stat_t
	if err := syscall.Lstat(fname, &info); err != nil {
		return ZeroFileInfo, err
	}
	if !isNormalFile(info) {
		return ZeroFileInfo, nil
	}
	size := info.Size

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
	var sector = uint64(0)
	if inCache > 0 {
		sector = GetSectorNumber(f.Fd())
	}

	return FileCacheInfo{
		FName:   fname,
		InCache: mc,
		InN:     inCache,

		dev:    info.Dev,
		inode:  info.Ino,
		sector: sector,
	}, nil
}

func GetSectorNumber(fd uintptr) uint64 {
	b := 0
	const FIBMAP = 1
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), FIBMAP, uintptr(unsafe.Pointer(&b)))
	if err != 0 {
		fmt.Println("E:", err)
	}
	return uint64(b)
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
