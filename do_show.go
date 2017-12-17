package main

import (
	"fmt"
	"strings"
)

func DumpCurrentPageCache(dirs []string) error {
	ch := make(chan FileCacheInfo)
	err := Produce(ch, dirs)
	if err != nil {
		return err
	}
	return consumePrint(ch, dirs)
}

func consumePrint(ch <-chan FileCacheInfo, dirs []string) error {
	var totalRAMSize, totalFileSize, totalUsedFileSize int
	var totalFile, usedFile int

	for info := range ch {
		totalFile++
		s := info.FileSize()
		totalFileSize += s
		if info.InN > 0 {
			usedFile++
			totalUsedFileSize += s
			totalRAMSize += info.RAMSize()
			fmt.Println(info)
		}
	}

	if totalUsedFileSize > 0 {
		fmt.Printf("%s\t%d%%\t%s",
			humanSize(totalRAMSize),
			totalRAMSize*100/totalUsedFileSize,
			fmt.Sprintf("[FOR %q FILES USED/TOTAL: %d/%d]\n",
				strings.Join(dirs, ","),
				usedFile, totalFile,
			),
		)
	}
	return nil
}
