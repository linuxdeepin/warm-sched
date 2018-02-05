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
			Sink(s.Scope(), fmt.Sprintf("%ds", id))
			g.Done()
		})
	}
	g.Wait()
	close(ch)
	return nil
}
func (s timeoutScope) Stop()       {}
func (timeoutScope) Scope() string { return "timeout" }

func TestSystemd(t *testing.T) {
	err := WaitAll("systemd:ssh.service")
	if err != nil {
		t.Error(err)
	}
}

func TestWait(t *testing.T) {
	err := WaitAll("")
	if err == nil {
		t.Error("Shouldn't support illegal format events")
	}

	err = WaitAll("non-exists:abc")
	if err == nil {
		t.Error("Shouldn't support empty events")
	}

	err = WaitAll("timeout:1s", "timeout:2s")
	if err != nil {
		t.Error("Should support timeout events")
	}
}
