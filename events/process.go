package events

import (
	"os"
	"path"
	"path/filepath"
)

type ProcessSource struct {
}

func init() {
	Register(&ProcessSource{})
}

const processScope = "process"

func (ProcessSource) Scope() string { return processScope }

func (s ProcessSource) Check(names []string) []string {
	es, err := filepath.Glob("/proc/*/exe")
	if err != nil {
		return nil
	}
	var ret []string
	for _, i := range es {
		e, err := os.Readlink(i)
		if err != nil {
			continue
		}
		base := path.Base(e)
		for _, j := range names {
			if j == base {
				ret = append(ret, j)
			}
		}
	}
	return ret
}
