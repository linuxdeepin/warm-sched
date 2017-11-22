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
	if usedFile > 0 {
		fmt.Fprintf(os.Stderr, "---------------Total--------------\n")
		fmt.Fprintf(os.Stderr, "Size: %s/%s\t Number: %d/%d\n",
			humanSize(totalRAMSize),
			humanSize(totalFileSize),
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
