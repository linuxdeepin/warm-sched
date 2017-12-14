package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
)

// Read PageCache from /proc/mincores
const MincoresPath = "/proc/mincores"

func SupportProduceByKernel() bool {
	return true
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
		//		verifyBySyscall(info)

		ch <- info
	}
}

func verifyBySyscall(info FileCacheInfo) {
	info2, err := fileMincore(path.Join("/home", info.FName))
	if err != nil {
		fmt.Println("SysscallFail...", info.FName)
	}
	r1, r2 := ToRanges(info.InCache, 1), ToRanges(info2.InCache, 1)
	if !reflect.DeepEqual(r1, r2) {
		fmt.Printf("WTF: %s \n\t%v\n\t%v\n", info2.FName, info.InCache, info2.InCache)
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
	mc := make([]bool, filePages)
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
