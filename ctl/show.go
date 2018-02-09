package main

import "../core"
import "fmt"

func DumpSnapshot(snap *core.Snapshot) error {
	var totalRAMSize, totalFileSize, totalUsedFileSize int
	var totalFile, usedFile int

	for _, info := range snap.Infos {
		totalFile++
		totalFileSize += int(info.FileSize)
		if len(info.Mapping) > 0 {
			usedFile++
			totalUsedFileSize += int(info.FileSize)
			totalRAMSize += info.RAMSize()
			fmt.Println(info)
		}
	}

	if totalUsedFileSize > 0 {
		fmt.Printf("%s\t%d%%\t%s",
			core.HumanSize(totalRAMSize),
			totalRAMSize*100/totalUsedFileSize,
			fmt.Sprintf("[FOR %q FILES USED/TOTAL: %d/%d]\n",
				"/",
				usedFile, totalFile,
			),
		)
	}
	return nil
}
