package events

import (
	"fmt"
	"github.com/coreos/go-systemd/dbus"
)

func init() {
	Register(&systemdEventSources{})
}

type systemdEventSources struct {
	stop chan struct{}
}

const SystemdScope = "systemd"

func (s *systemdEventSources) Scope() string { return SystemdScope }

func (s systemdEventSources) Check(names []string) []string {
	conn, err := dbus.New()
	if err != nil {
		fmt.Println("E:", err)
		return nil
	}
	defer conn.Close()
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
