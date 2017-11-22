package main

import (
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
	"unsafe"
)

type FileCacheInfo struct {
	FName   string
	InCache []bool
}

var ZeroFileInfo = FileCacheInfo{}

func (info FileCacheInfo) String() string {
	per := info.Percentage()
	return fmt.Sprintf("%.0fKB\t%0.1f%%\t%s",
		per*float32(len(info.InCache)*os.Getpagesize())/1024/100,
		per,
		info.FName,
	)
}

func (info FileCacheInfo) Percentage() float32 {
	var c int
	for _, i := range info.InCache {
		if i {
			c++
		}
	}
	return 100 * float32(c) / float32(len(info.InCache))
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
	info, err := os.Stat(fname)
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

	for i, b := range vec {
		if b%2 == 1 {
			mc[i] = true
		} else {
			mc[i] = false
		}
	}

	return FileCacheInfo{
		FName:   fname,
		InCache: mc,
	}, nil
}
