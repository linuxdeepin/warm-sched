package core

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
)

// Read PageCache from /proc/mincores
const mincoresPath = "/proc/mincores"

func generateFileInfoByKernel(ch chan<- FileInfo, mps []string) {
	defer close(ch)
	for _, t := range mps {
		collectMincores(ch, t)
	}
}

func supportProduceByKernel() error {
	if _, err := os.Stat(mincoresPath); err != nil {
		return fmt.Errorf("Please insmod mincores")
	}
	return nil
}

func collectMincores(ch chan<- FileInfo, mntPoint string) {
	if mntPoint != "" && mntPoint != "." {
		wd, _ := os.Getwd()
		defer os.Chdir(wd)
		os.Chdir(mntPoint)
	}
	f, err := os.Open(mincoresPath)
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

		fname := path.Join(mntPoint, strings.TrimSpace(fields[3]))

		info, err := buildFileInfoFromKernel(
			fname,
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

func buildFileInfoFromKernel(fname string, bn int64, filePages int64, mapping string) (FileInfo, error) {
	bm, err := parseMapRange(mapping)
	if err != nil {
		return FileInfo{}, err
	}
	return FileInfo{
		Name:    fname,
		Mapping: bm,
		Size:    uint64(filePages) * uint64(pageSize),
		dev:     0,
		sector:  uint64(bn),
	}, nil
}

func parseMapRange(raw string) ([]PageRange, error) {
	mc := make([]PageRange, 0)
	var start, end int
	for _, r := range strings.Split(raw, ",") {
		_, err := fmt.Sscanf(r, "[%d:%d]", &start, &end)
		if err != nil {
			break
		}
		mc = append(mc, PageRange{
			Offset: start,
			Count:  end - start + 1,
		})
	}
	return mc, nil
}
