package core

import (
	"fmt"
)

type FileInfoHandleFunc func(FileInfo) error

type PageRange struct {
	Offset int
	Count  int
}

type FileInfo struct {
	Name     string
	Mapping  []PageRange
	FileSize uint64

	dev    uint64
	sector uint64
}

func (info FileInfo) String() string {
	return fmt.Sprintf("%s\t%d%%\t%s",
		HumanSize(info.RAMSize()),
		info.Percentage(),
		info.Name,
	)
}

func (info FileInfo) Percentage() int {
	if info.FileSize == 0 {
		return 0
	}
	return 100 * info.RAMSize() / int(roundPageCount(int64(info.FileSize))*pageSize)
}

func (info FileInfo) RAMSize() int {
	c := 0
	for _, r := range info.Mapping {
		c += int(r.Count)
	}
	return c * pageSize
}
