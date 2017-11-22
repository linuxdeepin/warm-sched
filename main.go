package main

import (
	"fmt"
	"os"
)

func ConsumePrint(ch <-chan FileCacheInfo) {
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

	fmt.Printf("%s\t%d%%\t%s",
		humanSize(totalRAMSize),
		totalRAMSize*100/totalFileSize,
		fmt.Sprintf("%d/%d (used files/total files) [THIS LINE IS FOR SUMMARY]\n", usedFile, totalFile),
	)

}

func main() {
	ch := make(chan FileCacheInfo)

	if len(os.Args) == 1 {
		go Produce(ch, []string{"."})
	} else {
		go Produce(ch, os.Args[1:])
	}

	ConsumePrint(ch)
}
