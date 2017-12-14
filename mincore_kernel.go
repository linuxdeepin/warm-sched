package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Read PageCache from /proc/mincores
const MincoresPath = "/proc/mincores"

func SupportProduceByKernel() bool {
	return false
}

func ProduceByKernel(ch chan<- FileCacheInfo, dirs []string) {
	defer close(ch)
	collectMincores(ch, "")
}

func collectMincores(ch chan<- FileCacheInfo, mntPoint string) {
	if mntPoint != "" && mntPoint != "." {
		os.Chdir(mntPoint)
	}
	f, err := os.Open(MincoresPath)
	if err != nil {
		return
	}
	buf := bufio.NewReader(f)

	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			break
		}
		fields := strings.SplitN(line, "\t", 4)
		if len(fields) != 4 {
			break
		}
		bn, err := strconv.ParseInt(fields[0], 10, 64)
		if err != nil {
			break
		}
		s, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			break
		}

		info, err := buildFileCacheInfoFromKernel(
			strings.TrimSpace(fields[3]),
			bn,
			s,
			fields[2],
		)
		if err != nil {
			fmt.Printf("E:%q %v\n", line, err)
			break
		}
		ch <- info
	}
}

func buildFileCacheInfoFromKernel(fname string, bn int64, filePages int64, mapping string) (FileCacheInfo, error) {
	inN, bm, err := parseMapRange(filePages, mapping)
	if err != nil {
		return ZeroFileInfo, err
	}
	return FileCacheInfo{
		FName:   fname,
		sector:  uint64(bn),
		InCache: bm,
		InN:     int(inN),
	}, nil
}

func parseMapRange(filePages int64, raw string) (int64, []bool, error) {
	mc := make([]bool, filePages+1)
	for i := range mc {
		mc[i] = false
	}
	var start, end int64
	var total int64
	for _, r := range strings.Split(raw, ",") {
		_, err := fmt.Sscanf(r, "[%d:%d]", &start, &end)
		if err != nil {
			break
		}
		total += end - start + 1
		for i := start; i <= end; i++ {
			if i > filePages {
				return 0, nil, fmt.Errorf("WTF: %d %d\n", i, filePages)
			}
			mc[i] = true
		}
	}
	return total, mc, nil
}
