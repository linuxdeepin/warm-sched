package main

import (
	"fmt"
	"strings"
)

func ShowRAMUsage(dirs []string) {
	ch := make(chan FileCacheInfo)
	go Produce(ch, dirs)
	consumePrint(ch, dirs)
}

func consumePrint(ch <-chan FileCacheInfo, dirs []string) {
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

	if totalFileSize > 0 {
		fmt.Printf("%s\t%d%%\t%s",
			humanSize(totalRAMSize),
			totalRAMSize*100/totalUsedFileSize,
			fmt.Sprintf("[FOR %q FILES USED/TOTAL: %d/%d]\n",
				strings.Join(dirs, ","),
				usedFile, totalFile,
			),
		)
	}
}
