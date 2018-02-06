package events

import (
	"fmt"
	"github.com/coreos/go-systemd/dbus"
)

func init() {
	Register(&_SystemdEventSources{})
}

type _SystemdEventSources struct {
	stop chan struct{}
}

func (s *_SystemdEventSources) Scope() string { return "systemd" }

func (s _SystemdEventSources) Check(names []string) []string {
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
