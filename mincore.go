package main

import (
	"fmt"
	"strings"
)

type MemRange struct {
	Offset int64
	Length int64
}

var MaxAdviseSize = int64(128 * KB)

type FileCacheInfo struct {
	FName   string
	InCache []bool
	InN     int

	inode  uint64
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

func Produce(ch chan<- FileCacheInfo, mps []string) error {
	if err := SupportProduceByKernel(); err != nil {
		return err
	}
	go ProduceByKernel(ch, mps)
	return nil
}

func (info FileCacheInfo) String() string {
	return fmt.Sprintf("%s\t%d%%\t%s",
		humanSize(info.RAMSize()),
		info.Percentage(),
		info.FName,
	)
}

func (info FileCacheInfo) Percentage() int {
	n := len(info.InCache)
	if n == 0 {
		return 0
	}
	return info.InN * 100 / n
}
func (info FileCacheInfo) RAMSize() int {
	return info.InN * PageSize
}
func (info FileCacheInfo) FileSize() int {
	return len(info.InCache) * PageSize
}
