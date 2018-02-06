package events

import (
	"fmt"
	"github.com/coreos/go-systemd/dbus"
	"time"
)

func init() {
	Register(&_SystemdEventSources{})
}

type _SystemdEventSources struct {
	stop chan struct{}
}

func (s *_SystemdEventSources) Scope() string { return "systemd" }

func (s *_SystemdEventSources) Prepare(ids []string) error {
	return nil
}

func (s *_SystemdEventSources) Stop() {
	if s.stop != nil {
		s.stop <- struct{}{}
	}
}

func (s *_SystemdEventSources) Run() error {
	if s.stop != nil {
		return fmt.Errorf("BUG ON RUN")
	}

	scope := s.Scope()

	s.stop = make(chan struct{})

	for {
		select {
		case <-time.After(time.Second):
			pending := Pendings(scope)
			for _, id := range checkSystemd(pending) {
				Sink(scope, id)
			}
		case <-s.stop:
			s.stop = nil
			return nil
		}
	}
}

func checkSystemd(names []string) []string {
	conn, err := dbus.New()
	if err != nil {
		fmt.Println("E:", err)
		return nil
	}
	us, err := conn.ListUnitsByNames(names)
	if err != nil {
		fmt.Println("E:", err)
		return nil
	}
	var actives []string
	for _, u := range us {
		if u.ActiveState != "active" {
			continue
		}
		actives = append(actives, u.Name)
	}
	return actives
}
