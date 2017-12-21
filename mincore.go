package main

import (
	"fmt"
	"os"
	"strings"
)

type MemRange struct {
	Offset int
	Count  int
}

var PageSize = os.Getpagesize()

type Inode struct {
	Name    string
	Mapping []MemRange
	Size    uint64

	dev    uint64
	sector uint64
}

var BlackDirectory = stringSlice{
	"/sys/",
	"/proc/",
	"/dev/",
	"/run/",
	"/boot/",
	"/tmp/",
	"/var/log/",
	"/var/lib/dpkg/",
	"/var/lib/apt/",
	"/var/lib/lastore/",
	"/var/lib/docker/",
}

func (ss stringSlice) ShouldSkip(d string) bool {
	for _, i := range ss {
		if strings.HasPrefix(d, i) {
			return true
		}
	}
	return false
}

func Produce(ch chan<- Inode, mps []string) error {
	if err := SupportProduceByKernel(); err != nil {
		return err
	}
	go ProduceByKernel(ch, mps)
	return nil
}

func (info Inode) String() string {
	return fmt.Sprintf("%s\t%d%%\t%s",
		humanSize(info.RAMSize()),
		info.Percentage(),
		info.Name,
	)
}

func (info Inode) Percentage() int {
	if info.Size == 0 {
		return 0
	}
	return 100 * info.RAMSize() / int(info.Size)
}

func (info Inode) RAMSize() int {
	c := 0
	for _, r := range info.Mapping {
		c += int(r.Count)
	}
	return c * PageSize
}
