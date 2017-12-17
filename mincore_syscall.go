package main

import (
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

var BlackDirectory = []string{"/sys", "/proc", "/dev", "/run", "/boot"}

func shouldSkipDirectory(root string) bool {
	r, err := filepath.Abs(root)
	if err != nil {
		return true
	}
	for _, i := range BlackDirectory {
		if strings.HasPrefix(r, i) {
			return true
		}
	}
	return false
}

func ProduceBySyscall(ch chan<- FileCacheInfo, dirs []string) {
	defer close(ch)
	for _, f := range dirs {
		info, err := os.Lstat(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid Args %v\n", err)
			return
		}
		if info.IsDir() {
			err = showDirCacheInfos(f, ch)
		} else {
			err = showFileCacheInfo(f, ch)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "E:", err)
		}
	}
}

func showDirCacheInfos(root string, ch chan<- FileCacheInfo) error {
	if shouldSkipDirectory(root) {
		return nil
	}
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
			showDirCacheInfos(name, ch)
		} else {
			showFileCacheInfo(name, ch)
		}
	}
	return nil
}

func showFileCacheInfo(fname string, ch chan<- FileCacheInfo) error {
	info, err := fileMincore(fname)
	if err != nil {
		return err
	}
	ch <- info
	return nil
}

func fileMincore(fname string) (FileCacheInfo, error) {
	fname, err := filepath.Abs(fname)
	if err != nil {
		return ZeroFileInfo, err
	}

	info, err := os.Lstat(fname)
	if err != nil {
		return ZeroFileInfo, err
	}
	if !info.Mode().IsRegular() {
		return ZeroFileInfo, err
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
	var sector = uint64(0)
	if inCache > 0 && RunByRoot {
		sector = GetSectorNumber(f.Fd())
	}

	sinfo := info.Sys().(*syscall.Stat_t)

	return FileCacheInfo{
		FName:   fname,
		InCache: mc,
		InN:     inCache,

		dev:    sinfo.Dev,
		inode:  sinfo.Ino,
		sector: sector,
	}, nil
}

var RunByRoot = os.Geteuid() == 0

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

func GetSectorNumber(fd uintptr) uint64 {
	b := 0
	const FIBMAP = 1
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), FIBMAP, uintptr(unsafe.Pointer(&b)))
	if err != 0 {
		fmt.Println("E:", err)
	}
	return uint64(b)
}
