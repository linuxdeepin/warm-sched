package main

import (
	"fmt"
	"os"
	"path"
)

func ShowFileCacheInfo(fname string) error {
	info, err := FileMincore(fname)
	if err != nil {
		return err
	}
	if info.Percentage() > 0 {
		fmt.Println(info)
	}
	return nil
}
func ShowDirCacheInfos(root string) error {
	f, err := os.Open(root)
	if err != nil {
		return err
	}
	infos, err := f.Readdir(0)
	if err != nil {
		return err
	}
	for _, info := range infos {
		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			continue
		}
		name := path.Join(root, info.Name())
		if info.IsDir() {
			ShowDirCacheInfos(name)
		} else {
			ShowFileCacheInfo(name)
		}
	}
	return nil
}

func main() {
	for _, f := range os.Args[1:] {
		e := ShowDirCacheInfos(f)
		if e != nil {
			fmt.Fprintln(os.Stderr, "E:", e)
		}
	}
}
