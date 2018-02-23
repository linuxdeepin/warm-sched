package events

import (
	"os"
)

type fileSource struct {
}

func init() {
	Register(&fileSource{})
}

const FileScope = "file"

func (fileSource) Scope() string { return FileScope }

func (s fileSource) Check(fs []string) []string {
	var ret []string
	for _, f := range fs {
		_, err := os.Stat(f)
		if err != nil {
			continue
		}
		ret = append(ret, f)
	}
	return ret
}
