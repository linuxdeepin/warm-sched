package events

import (
	"context"
	"fmt"
	"os/exec"
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
	if exec.Command("systemctl", "is-system-running").Run() != nil {
		t.Skip("Currently not in Systemd environment")
		return
	}
	err := Connect([]string{"systemd:local-fs.target"}, nil)
	if err != nil {
		t.Error(err)
	}
	err = Run(context.TODO())
	if err != nil {
		t.Error(err)
	}
}

func TestDBUSClose(t *testing.T) {
	ss := systemdEventSources{}
	for i := 0; i < 500; i++ {
		ss.Check(nil)
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

	err = Run(context.TODO())
	if err != nil {
		t.Error(err)
	}
}

func TestProcess(t *testing.T) {
	found := false
	Connect([]string{"process:go"}, func() {
		found = true
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err := Run(ctx)
	if err != nil {
		t.Error(err)
	}
	if !found {
		t.Fatal("Should found process of go")
	}
}

func TestFile(t *testing.T) {
	Connect([]string{"file:/shouldn't have this file"}, func() {
		t.Fatal("The file shouldn't exists.")
	})

	found := false
	Connect([]string{"file:/etc"}, func() {
		found = true
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	err := Run(ctx)
	if err != nil {
		t.Error(err)
	}
	if !found {
		t.Fatal("Should found /etc")
	}
}
