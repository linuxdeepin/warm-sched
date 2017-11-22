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
		totalFileSize += len(info.InCache) * PageSize
		if info.InN > 0 {
			usedFile++
			totalRAMSize += info.InN * PageSize
			fmt.Println(info)
		}
	}
	if usedFile > 0 {
		fmt.Fprintf(os.Stderr, "---------------Total--------------\n")
		fmt.Fprintf(os.Stderr, "Size: %0.2fMB/%0.2fMB\t Number: %d/%d\n",
			float32(totalRAMSize)/float32(MB),
			float32(totalFileSize)/float32(MB),
			usedFile,
			totalFile,
		)
	}
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
