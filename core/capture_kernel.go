package core

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Read PageCache from /proc/mincores
const mincoresPath = "/proc/mincores"

var SystemMountPoints = ListMountPoints(sysProcMountsPath)

func calcRealTargets(dirs []string, mountPoints []string,
	bindMap map[string]string) []string {

	targets := make(map[string]struct{})

	// 根据 dir 获取挂载点，并记录到 targets 中
	getDirMountPoint := func(targets map[string]struct{}, dir string) {
		// 要求 mountPoints 按字符串长度（长->短）排序
		for _, mp := range mountPoints {
			if strings.HasPrefix(dir, mp) {
				_, ok := targets[mp]
				if !ok {
					targets[mp] = struct{}{}
				}
				break
			}
		}
	}

	for _, dir := range dirs {
		getDirMountPoint(targets, dir)
	}
	// 获取无重复的挂载点列表
	ret := mapStrEmptyStructKeys(targets)

	// 解析 mount bind 信息，获取更加下一级的挂载点
	targets = make(map[string]struct{})
	for _, t := range ret {
		// 把 /home 转换为 /data/home
		resolved := resolveMountBind(bindMap, t)
		if resolved != "" {
			// 获取 /data/home 的更下一级的挂载点，即 /data
			getDirMountPoint(targets, resolved)
		} else {
			// 并非 mount bind 的结果
			targets[t] = struct{}{}
		}
	}

	ret = mapStrEmptyStructKeys(targets)
	return ret
}

func mapStrEmptyStructKeys(dict map[string]struct{}) []string {
	keys := make([]string, 0, len(dict))
	for key := range dict {
		keys = append(keys, key)
	}
	return keys
}

func resolveMountBind(bindMap map[string]string, mountPoint string) string {
	for src, dst := range bindMap {
		if dst == mountPoint {
			return src
		}
	}
	return ""
}

func generateFileInfoByKernel(ch chan<- FileInfo, mountpoints []string, bindMap map[string]string) {
	defer close(ch)

	for _, t := range mountpoints {
		collectMincores(ch, t, bindMap)
	}
}

func supportProduceByKernel() error {
	if _, err := os.Stat(mincoresPath); err != nil {
		return fmt.Errorf("Please insmod mincores")
	}
	return nil
}

func collectMincores(ch chan<- FileInfo, mntPoint string, bindMap map[string]string) {
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

		fname := filepath.Join(mntPoint, strings.TrimSpace(fields[3]))

		// 根据 mount bind 信息转换一下路径
		// 比如 /data/home/uos/file1 变为 /home/uos/file1, 更加正常
		for k, v := range bindMap {
			if strings.HasPrefix(fname, k) {
				fname = strings.TrimPrefix(fname, k)
				fname = filepath.Join(v, fname)
				break
			}
		}

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
		Name:     fname,
		Mapping:  bm,
		FileSize: uint64(filePages) * uint64(pageSize),
		dev:      0,
		sector:   uint64(bn),
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
