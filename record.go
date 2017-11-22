package main

import (
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"path"
	"path/filepath"
	"unsafe"
)

type FileCacheInfo struct {
	FName   string
	InCache []bool
	InN     int
}

const KB = 1024
const MB = 1024 * KB
const GB = 1024 * MB

func humanSize(s int) string {
	if s > GB {
		return fmt.Sprintf("%0.2fG", float32(s)/float32(GB))
	} else if s > MB {
		return fmt.Sprintf("%0.1fM", float32(s)/float32(MB))
	} else if s > KB {
		return fmt.Sprintf("%0.0fK", float32(s)/float32(KB))
	} else {
		return fmt.Sprintf("%dB", s)
	}
}

var ZeroFileInfo = FileCacheInfo{}
var PageSize = os.Getpagesize()
var PageSizeKB = os.Getpagesize() / KB

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

	// mmap is a []byte
	mmap, err := unix.Mmap(int(f.Fd()), 0, int(size), unix.PROT_NONE, unix.MAP_SHARED)
	if err != nil {
		return ZeroFileInfo, fmt.Errorf("could not mmap %s: %v", fname, err)
	}
	// TODO: check for MAP_FAILED which is ((void *) -1)
	// but maybe unnecessary since it looks like errno is always set when MAP_FAILED

	// one byte per page, only LSB is used, remainder is reserved and clear
	vecsz := (size + int64(os.Getpagesize()) - 1) / int64(os.Getpagesize())
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
			fmt.Fprintf(os.Stderr, "Invalid Args %v", err)
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
