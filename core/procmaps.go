package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func ReferencedFilesByPID(pids ...int) ([]string, error) {
	cache := make(map[string]struct{})
	for _, pid := range pids {
		mrs, err := parseMaps(fmt.Sprintf("/proc/%d/maps", pid))
		if err != nil {
			fmt.Fprintf(os.Stderr, "ReferecedFilesByPID: %v\n", err)
			continue
		}
		for _, name := range mrs {
			cache[name] = struct{}{}
		}
	}
	var ret []string
	for k := range cache {
		ret = append(ret, k)
	}
	return ret, nil
}

func parseMapsLine(d string) (string, error) {
	fs := strings.Fields(d)
	if len(fs) < 5 {
		return "", fmt.Errorf("Invalid input %q\n", d)
	}
	name := fs[len(fs)-1]
	if len(name) == 0 {
		return "", nil
	} else if name[0] != '/' || strings.HasSuffix(name, " (deleted)") {
		return "", nil
	} else {
		return name, nil
	}

}

func parseMaps(fpath string) ([]string, error) {
	var ret []string
	maps, err := ioutil.ReadFile(fpath)
	for _, line := range strings.Split(string(maps), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		r, err := parseMapsLine(line)
		if err != nil {
			return nil, err
		}
		if r == "" {
			continue
		}
		ret = append(ret, r)
	}
	return ret, err
}
