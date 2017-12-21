package main

import (
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"syscall"
	"unsafe"
)

func ProduceBySyscall(ch chan<- Inode, dirs []string) {
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
			err = showInode(f, ch)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "E:", err)
		}
	}
}

func showDirCacheInfos(root string, ch chan<- Inode) error {
	if BlackDirectory.ShouldSkip(root) {
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
			showInode(name, ch)
		}
	}
	return nil
}

func showInode(fname string, ch chan<- Inode) error {
	info, err := fileMincore(fname)
	if err != nil {
		return err
	}
	ch <- info
	return nil
}

func fileMincore(fname string) (Inode, error) {
	fname, err := filepath.Abs(fname)
	if err != nil {
		return Inode{}, err
	}

	info, err := os.Lstat(fname)
	if err != nil {
		return Inode{}, err
	}
	if !info.Mode().IsRegular() {
		return Inode{}, err
	}

	size := info.Size()

	f, err := os.Open(fname)
	if err != nil {
		return Inode{}, err
	}
	defer f.Close()

	mmap, err := unix.Mmap(int(f.Fd()), 0, int(size), unix.PROT_NONE, unix.MAP_SHARED)
	if err != nil {
		return Inode{}, fmt.Errorf("could not mmap %s: %v", fname, err)
	}

	vecsz := (size + PageSize64 - 1) / PageSize64
	vec := make([]byte, vecsz)

	// get all of the arguments to the mincore syscall converted to uintptr
	mmap_ptr := uintptr(unsafe.Pointer(&mmap[0]))
	size_ptr := uintptr(size)
	vec_ptr := uintptr(unsafe.Pointer(&vec[0]))

	ret, _, err := unix.Syscall(unix.SYS_MINCORE, mmap_ptr, size_ptr, vec_ptr)
	if ret != 0 {
		return Inode{}, fmt.Errorf("syscall SYS_MINCORE failed: %v", err)
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

	return Inode{
		Name:    fname,
		Size:    uint64(sinfo.Size),
		Mapping: ToRanges(mc, inCache),

		dev:    sinfo.Dev,
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

func VerifyBySyscall(info Inode) {
	info2, err := fileMincore(info.Name)
	if err != nil {
		fmt.Println("The mincore failed:", err)
		return
	}
	r1, r2 := info.Mapping, info2.Mapping
	if !reflect.DeepEqual(r1, r2) {
		fmt.Printf("WTF: %s \n\tKern:%v(%d)\n\tSys:%v(%d)\n", info2.Name,
			r1, int(info.Size)/PageSize, r2, int(info2.Size)/PageSize)
	}
}
