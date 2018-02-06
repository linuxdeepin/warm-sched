package events

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

type timeoutScope struct {
	points []int
}

func init() {
	s := timeoutScope{}
	Register(&s)
}

func (s *timeoutScope) Prepare(ids []string) error {
	for _, id := range ids {
		var t int
		_, err := fmt.Sscanf(id, "%ds\n", &t)
		if err != nil {
			return err
		}
		s.points = append(s.points, t)
	}
	return nil
}
func (s *timeoutScope) Run() error {
	ch := make(chan string)
	var g sync.WaitGroup
	for _, d := range s.points {
		g.Add(1)
		id := d

		time.AfterFunc(time.Duration(d)*time.Second, func() {
			Emit(s.Scope(), fmt.Sprintf("%ds", id))
			g.Done()
		})
	}
	g.Wait()
	close(ch)
	return nil
}
func (s timeoutScope) Stop()                 {}
func (timeoutScope) Scope() string           { return "timeout" }
func (timeoutScope) Check([]string) []string { return nil }

func TestSystemd(t *testing.T) {
	err := Connect([]string{"systemd:ssh.service"}, nil)
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

	err = Connect([]string{"timeout:1s", "timeout:2s"}, nil)
	if err != nil {
		t.Error("Should support timeout events")
	}

	err = Run()
	if err != nil {
		t.Error(err)
	}
}
