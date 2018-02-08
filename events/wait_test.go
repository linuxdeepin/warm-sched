package events

import (
	"fmt"
	"testing"
	"time"
)

type timeoutScope struct {
	start time.Time
}

func init() {
	s := timeoutScope{start: time.Now()}
	Register(&s)
}

func (timeoutScope) Scope() string { return "timeout" }

func (s timeoutScope) Check(points []string) []string {
	elapse := time.Now().Sub(s.start).Seconds()
	var ret []string
	for _, p := range points {
		var t int
		fmt.Sscanf(p, "%ds\n", &t)
		if elapse >= float64(t) {
			ret = append(ret, p)
		}
	}
	return ret
}

func TestSystemd(t *testing.T) {
	err := Connect([]string{"systemd:local-fs.target"}, nil)
	if err != nil {
		t.Error(err)
	}
	err = Run()
	if err != nil {
		t.Error(err)
	}
}

func TestWait(t *testing.T) {
	err := Connect(nil, nil)
	if err == nil {
		t.Error("Shouldn't support illegal format events")
	}

	err = Connect([]string{"non-exists:abc"}, nil)
	if err == nil {
		t.Error("Shouldn't support empty events")
	}

	err = Connect([]string{"timeout:1s", "timeout:5s"}, nil)
	if err != nil {
		t.Error("Should support timeout events")
	}

	err = Run()
	if err != nil {
		t.Error(err)
	}
}
