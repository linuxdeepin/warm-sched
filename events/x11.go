package events

import (
	"fmt"
	"os"

	"github.com/linuxdeepin/warm-sched/core"
)

type x11Source struct {
}

func init() {
	Register(&x11Source{})
}

const X11Scope = "x11"

func (x11Source) Scope() string { return X11Scope }

func (s x11Source) Check(names []string) []string {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Fatal:%v\n", r)
		}
	}()
	var ret []string
	core.X11ClientIterate(func(_ int, wmName string) bool {
		for _, n := range names {
			if n == wmName {
				ret = append(ret, n)
			}
		}
		return false
	})
	return ret
}
