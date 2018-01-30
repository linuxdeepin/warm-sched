package main

import (
	"fmt"
	"strings"
)

func DumpCurrentPageCache(dirs []string) error {
	ch := make(chan Inode)
	err := Produce(ch, dirs)
	if err != nil {
		return err
	}
	return consumePrint(ch, dirs)
}

func consumePrint(ch <-chan Inode, dirs []string) error {
	var totalRAMSize, totalFileSize, totalUsedFileSize int
	var totalFile, usedFile int

	for info := range ch {
		totalFile++
		totalFileSize += int(info.Size)
		if len(info.Mapping) > 0 {
			usedFile++
			totalUsedFileSize += int(info.Size)
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
