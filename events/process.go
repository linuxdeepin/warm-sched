package events

import (
	"os"
	"path"
	"path/filepath"
)

type processSource struct {
}

func init() {
	Register(&processSource{})
}

const ProcessScope = "process"

func (processSource) Scope() string { return ProcessScope }

func (s processSource) Check(names []string) []string {
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
