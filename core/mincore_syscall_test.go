package core

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

func toRange(vec []bool) (PageRange, []bool) {
	var s int
	var offset = -1
	var found = false
	for i, v := range vec {
		if !found && v {
			offset = i
		}
		if v {
			found = true
		}
		if !v && found {
			break
		}
		s++
	}
	return PageRange{offset, s - offset}, vec[s:]
}

func ToRanges(vec []bool) []PageRange {
	var ret []PageRange
	var i PageRange
	var pos = -1
	for {
		i, vec = toRange(vec)
		if i.Offset == -1 {
			break
		}
		if pos != -1 {
			i.Offset += pos
		}
		pos = i.Offset + i.Count

		ret = append(ret, i)
		if len(vec) == 0 {
			break
		}
	}
	return ret
}

func ProduceBySyscall(ch chan<- FileInfo, dirs []string) {
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
			err = showFileInfo(f, ch)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "E:", err)
		}
	}
}

func showDirCacheInfos(root string, ch chan<- FileInfo) error {
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
			showFileInfo(name, ch)
		}
	}
	return nil
}

func showFileInfo(fname string, ch chan<- FileInfo) error {
	info, err := fileMincore(fname)
	if err != nil {
		return err
	}
	ch <- info
	return nil
}

func fileMincore(fname string) (FileInfo, error) {
	fname, err := filepath.Abs(fname)
	if err != nil {
		return FileInfo{}, err
	}

	info, err := os.Lstat(fname)
	if err != nil {
		return FileInfo{}, err
	}
	if !info.Mode().IsRegular() {
		return FileInfo{}, err
	}

	size := info.Size()

	f, err := os.Open(fname)
	if err != nil {
		return FileInfo{}, err
	}
	defer f.Close()

	mmap, err := unix.Mmap(int(f.Fd()), 0, int(size), unix.PROT_NONE, unix.MAP_SHARED)
	if err != nil {
		return FileInfo{}, fmt.Errorf("could not mmap %s: %v", fname, err)
	}

	vecsz := (size + pageSize64 - 1) / pageSize64
	vec := make([]byte, vecsz)

	// get all of the arguments to the mincore syscall converted to uintptr
	mmap_ptr := uintptr(unsafe.Pointer(&mmap[0]))
	size_ptr := uintptr(size)
	vec_ptr := uintptr(unsafe.Pointer(&vec[0]))

	ret, _, err := unix.Syscall(unix.SYS_MINCORE, mmap_ptr, size_ptr, vec_ptr)
	if ret != 0 {
		return FileInfo{}, fmt.Errorf("syscall SYS_MINCORE failed: %v", err)
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

	return FileInfo{
		Name:    fname,
		Size:    uint64(sinfo.Size),
		Mapping: ToRanges(mc),

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

func VerifyBySyscall(info FileInfo) error {
	if err := syscall.Access(info.Name, unix.R_OK); err != nil {
		return nil
	}
	info2, err := fileMincore(info.Name)
	if err != nil {
		return err
	}
	r1, r2 := info.Mapping, info2.Mapping
	if !reflect.DeepEqual(r1, r2) {
		return fmt.Errorf("WTF: %s \n\tKern:%v(%d)\n\tSys:%v(%d)\n", info2.Name,
			r1, int(info.Size)/pageSize, r2, int(info2.Size)/pageSize)
	}
	return nil
}
