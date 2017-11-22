package main

import (
	"fmt"
)

func ShowRAMUsage(dirs []string) {
	ch := make(chan FileCacheInfo)
	go Produce(ch, dirs)
	consumePrint(ch)
}

func consumePrint(ch <-chan FileCacheInfo) {
	var totalRAMSize, totalFileSize int
	var totalFile, usedFile int

	for info := range ch {
		totalFile++
		totalFileSize += info.FileSize()
		if info.InN > 0 {
			usedFile++
			totalRAMSize += info.RAMSize()
			fmt.Println(info)
		}
	}

	if totalFileSize > 0 {
		fmt.Printf("%s\t%d%%\t%s",
			humanSize(totalRAMSize),
			totalRAMSize*100/totalFileSize,
			fmt.Sprintf("%d/%d (used files/total files) [THIS LINE IS FOR SUMMARY]\n", usedFile, totalFile),
		)
	}
}
